# Go Architecture Recipes

Reference implementations from [SKILL.md](SKILL.md). Currently: a `go.work` multi-module workspace with two bounded contexts — `order` (a normal context) and `checkout` (an orchestrator that coordinates `order`, inventory, and payment). See §16 for the rules; this file is the worked code.

## Multi-module workspace layout

```
myworkspace/
├── go.work
│
├── domain/                              # ONLY for value objects genuinely shared across contexts
│   └── money/
│       └── money.go
│
├── pkg/                                  # cross-cutting, shared by 2+ services — each its own module
│   ├── problem/                          # RFC 7807 mapper
│   ├── validation/
│   └── obs/
│
├── internal/
│   ├── order/                            # bounded context — own module
│   │   ├── domain/
│   │   │   └── order/
│   │   │       ├── order.go
│   │   │       ├── repository.go         # secondary port — lives in domain (DIP)
│   │   │       └── errors.go
│   │   ├── app/
│   │   │   └── order/
│   │   │       └── place_order.go
│   │   └── adapters/
│   │       ├── primary/
│   │       │   └── http/
│   │       │       └── handler.go
│   │       └── secondary/
│   │           └── postgres/
│   │               └── order_repo.go
│   │
│   └── checkout/                          # orchestrator context — own module
│       ├── app/
│       │   ├── ports.go                   # its OWN ports — not order's
│       │   └── checkout.go
│       └── adapters/
│           ├── primary/
│           │   └── http/
│           │       └── handler.go
│           └── secondary/
│               ├── orderclient/            # implements app.OrderCreator via HTTP
│               ├── inventoryclient/
│               └── paymentclient/
│
├── app/                                    # composition roots — one per deployable
│   ├── order/wire.go
│   └── checkout/wire.go
│
└── cmd/
    ├── order/main.go                        # thin entrypoint
    └── checkout/main.go
```

Every arrow points inward: `domain` imports nothing; `app` imports only `domain`; `adapters/*` import `app`+`domain` and satisfy their interfaces with compile-time assertions; `checkout` never imports `order` — only HTTP, decoded through the same `pkg/problem` mapper both services share.

## `domain/money/money.go` — shared value object (deliberate Shared Kernel)

```go
package money

type Money struct {
    Amount   int64 // minor units (cents)
    Currency string
}

func (m Money) IsZero() bool { return m.Amount == 0 }
```

## `internal/order/domain/order/` — pure business logic + the port

```go
// order.go
package order

func New(customer CustomerID, total money.Money) (Order, error) {
    if total.IsZero() {
        return Order{}, ErrEmptyOrder
    }
    return Order{ID: newID(), Customer: customer, Total: total, Status: StatusPending}, nil
}

type Order struct {
    ID       ID
    Customer CustomerID
    Total    money.Money
    Status   Status
}
```

```go
// repository.go — secondary port; DIP requires this live in domain, not app
package order

import "context"

type Repository interface {
    Save(ctx context.Context, o Order) error
    FindByID(ctx context.Context, id ID) (Order, error)
}
```

```go
// errors.go
package order

import "errors"

var (
    ErrEmptyOrder = errors.New("order: total cannot be zero")
    ErrNotFound   = errors.New("order: not found")
)
```

## `internal/order/app/order/place_order.go` — thin orchestration

No explicit primary port here: `PlaceOrderService` has exactly one production caller (the HTTP handler below), so per [hexagonal-arch §7](../../design/hexagonal-arch/SKILL.md#7-when-not-to-use-it) an interface would be ceremony without payoff. Introduce one (`PlaceOrderer`, declared here since this is the layer offering the capability) only when a second caller genuinely needs to swap it — e.g. a handler test injecting a fake.

```go
package order

type PlaceOrderCommand struct {
    Customer domain.CustomerID
    Total    money.Money
}

type PlaceOrderService struct {
    repo domain.Repository // depends on the port, not a concrete adapter
}

func NewPlaceOrderService(repo domain.Repository) *PlaceOrderService {
    return &PlaceOrderService{repo: repo}
}

func (s *PlaceOrderService) PlaceOrder(ctx context.Context, cmd PlaceOrderCommand) (domain.ID, error) {
    o, err := domain.New(cmd.Customer, cmd.Total)
    if err != nil {
        return domain.ID{}, fmt.Errorf("place order: %w", err)
    }
    if err := s.repo.Save(ctx, o); err != nil {
        return domain.ID{}, fmt.Errorf("place order: %w", err)
    }
    return o.ID, nil
}
```

## `internal/order/adapters/secondary/postgres/order_repo.go` — driven adapter

```go
package postgres

var _ domain.Repository = (*OrderRepo)(nil) // makes the contract checkable at compile time

type OrderRepo struct{ db *sqlx.DB }

//go:embed queries/save_order.sql
var saveOrderSQL string

func (r *OrderRepo) Save(ctx context.Context, o domain.Order) error {
    _, err := r.db.ExecContext(ctx, saveOrderSQL, o.ID, o.Customer, o.Total.Amount, o.Status)
    return err
}

func (r *OrderRepo) FindByID(ctx context.Context, id domain.ID) (domain.Order, error) {
    var o domain.Order
    err := r.db.GetContext(ctx, &o, findOrderByIDSQL, id)
    return o, err
}
```

## `internal/order/adapters/primary/http/handler.go` — driving adapter

RFC 7807 is applied exactly at this boundary — the domain and app layers never format a wire response.

```go
package http

type Handler struct{ svc *orderapp.PlaceOrderService }

func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
    var cmd orderapp.PlaceOrderCommand
    if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
        problem.Write(w, problem.BadRequest("invalid request body"))
        return
    }
    id, err := h.svc.PlaceOrder(r.Context(), cmd)
    if err != nil {
        problem.WriteFromError(w, err) // maps domain/app errors → RFC 7807
        return
    }
    json.NewEncoder(w).Encode(struct {
        ID string `json:"id"`
    }{id.String()})
}
```

## `internal/checkout/app/ports.go` — the orchestrator's own ports

Unlike `order`'s optional primary port, these are load-bearing from day one: `checkout` has no other way to reach `order`/inventory/payment, and each port already has two real implementations (the HTTP client below, and a fake used to unit-test the orchestration logic without booting three services).

```go
package app

type OrderCreator interface {
    CreateOrder(ctx context.Context, customer string, total money.Money) (string, error)
}
type StockReserver interface {
    Reserve(ctx context.Context, sku string, qty int) error
}
type PaymentCharger interface {
    Charge(ctx context.Context, customer string, amount money.Money) error
}
```

```go
// checkout.go
package app

type CheckoutService struct {
    orders    OrderCreator
    inventory StockReserver
    payments  PaymentCharger
}

func NewCheckoutService(orders OrderCreator, inventory StockReserver, payments PaymentCharger) *CheckoutService {
    return &CheckoutService{orders: orders, inventory: inventory, payments: payments}
}

func (s *CheckoutService) Checkout(ctx context.Context, cmd CheckoutCommand) error {
    if err := s.inventory.Reserve(ctx, cmd.SKU, cmd.Qty); err != nil {
        return fmt.Errorf("checkout: reserve stock: %w", err)
    }
    if err := s.payments.Charge(ctx, cmd.Customer, cmd.Total); err != nil {
        return fmt.Errorf("checkout: charge payment: %w", err) // caller compensates: release reservation
    }
    if _, err := s.orders.CreateOrder(ctx, cmd.Customer, cmd.Total); err != nil {
        return fmt.Errorf("checkout: create order: %w", err)
    }
    return nil
}
```

## `internal/checkout/adapters/secondary/orderclient/client.go` — crosses the real service boundary

An anticorruption layer per [ddd-architect §8](../../design/ddd-architect/SKILL.md#8-anticorruption-layers-acl): translates at the boundary, never lets `order`'s wire model leak into `checkout`.

```go
package orderclient

var _ app.OrderCreator = (*Client)(nil)

type Client struct {
    base *url.URL
    hc   *http.Client
}

func (c *Client) CreateOrder(ctx context.Context, customer string, total money.Money) (string, error) {
    resp, err := c.hc.Post(c.base.String()+"/orders", "application/json", encode(customer, total))
    if err != nil {
        return "", fmt.Errorf("orderclient: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 {
        return "", problem.DecodeAsError(resp.Body) // decodes the RFC 7807 body back into a Go error
    }
    return decodeID(resp.Body)
}
```

## `app/order/wire.go` — composition root

```go
package app

func Run(ctx context.Context, cfg Config) error {
    db := mustConnect(cfg.DB)
    repo := &postgres.OrderRepo{DB: db}
    svc := orderapp.NewPlaceOrderService(repo)
    handler := &orderhttp.Handler{Svc: svc}
    return startServer(ctx, cfg.Addr, handler)
}
```

## `cmd/order/main.go` — thin entrypoint

```go
func main() {
    cfg := loadConfig()
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()
    if err := app.Run(ctx, cfg); err != nil {
        slog.Error("order service exited", "err", err)
        os.Exit(1)
    }
}
```
