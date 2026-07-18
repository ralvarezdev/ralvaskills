# Go Library Builder — Skeletons & Reference

Load on demand when scaffolding. Tooling files (`Taskfile.yml`, `mise.toml`, `.editorconfig`, `renovate.json`, `.golangci.doccheck.yml`, `.gitignore`) are ready to copy from [assets/](assets). This file holds the Go-source skeletons and the two folder trees.

## 1. Folder trees

### Flat single-port library (`ratelimit`, `email` scale)

```
<name>/
├── <name>.go              # domain: value objects, Config, Result, sentinel errors, package doc
├── <port>.go              # the port interface (may be merged into <name>.go for a single port)
├── <name>_test.go         # black-box tests (package <name>_test)
├── memory/                # in-memory adapter — the reusable test double
│   ├── <port>.go
│   └── <port>_test.go
├── <tech>/                # production adapter (smtp, valkey, postgres, …)
│   ├── <port>.go
│   ├── <port>_test.go     # integration (testcontainers)
│   └── main_test.go       # TestMain: one container per package
├── logger/                # optional dev/observability adapter
│   └── <port>.go
├── internal/testhelper/   # container bootstrap (only if an adapter needs a container)
├── docs/YYYY-MM-DD-spec.md
├── go.mod / go.sum
├── mise.toml / Taskfile.yml
├── .golangci.yml / .golangci.doccheck.yml
├── .editorconfig / .gitignore / renovate.json / README.md / LICENSE / CLAUDE.md
└── .rsk/{rsk.mod,rsk.lock,CLAUDE.md}
```

### Vertical-slice library (`identity` scale)

```
<name>/
├── <context>/                     # one bounded context (user, auth, …)
│   ├── <context>.go               #   domain: entities, VOs, sentinel errors
│   ├── store.go                   #   driven port(s) — segregated + composed
│   ├── service.go                 #   domain service (concrete *Service) + driving port
│   ├── <concept>.go               #   one file per port group (rbac.go, throttle.go, …)
│   └── <tech>/                    #   adapter subpackage (postgres/, jwt/, cache/, …)
│       ├── <port>.go
│       ├── queries/*.sql          #   sqlc input (postgres adapters)
│       └── db/                    #   sqlc-GENERATED (do not edit)
├── app/                           # application layer: cross-context use-cases on *App
│   ├── service.go                 #   App struct + New() wiring + Config.setDefaults()
│   └── <flow>.go                  #   registration.go, recovery.go, …
├── internal/
│   ├── acl/<x>lookup.go           #   anticorruption layer between contexts
│   ├── pg/errors.go               #   shared driver-error helpers
│   └── testhelper/{db.go,...}
├── migrate/migrate.go             # public migration runner (goose + //go:embed)
├── migrations/{NNN_verb_noun.sql, embed.go}
├── cmd/gen/main.go                # migrations → schema.sql → sqlc generate
├── sqlc.yaml / tools.go / generate.go
└── (same root tooling set as the flat tree)
```

## 2. Config & constructors

Plain struct, `setDefaults()` over `Default*` constants, validate in `New`. **No functional options** — the house style favors explicit, exhaustively-initialized structs (`exhaustruct` scoped to `Config` forces every field at every literal, which is the point). The only concession to "options" is `setDefaults()` filling zero values.

```go
const DefaultBurst = 10

type Config struct {
	Rate  float64 // tokens per second; required
	Burst int     // bucket capacity; defaults to DefaultBurst
}

func (c *Config) setDefaults() {
	if c.Burst == 0 {
		c.Burst = DefaultBurst
	}
}

func (c Config) Validate() error {
	var errs []error
	if c.Rate <= 0 {
		errs = append(errs, ErrInvalidRate)
	}
	if c.Burst < 0 {
		errs = append(errs, ErrInvalidBurst)
	}
	return errors.Join(errs...)
}
```

The library never resolves config from env/files — the consuming app passes plain values. For a service constructor with real dependencies: `func New(secret []byte, accessTTL, refreshTTL time.Duration) *Issuer`.

## 3. Domain package skeleton (port + value object + errors)

Root package, stdlib + uuid only. Ordering per `decorder`: const → var → type → func.

```go
// Package email sends transactional mail. The domain owns the Mailer port;
// only adapter subpackages may import a mail library. Nothing here imports
// anything beyond the standard library.
package email

var (
	ErrNoRecipients = errors.New("email: message has no recipients")
	ErrNoBody       = errors.New("email: message has no body")
)

// Address is a validated RFC 5322 address; the zero value is invalid.
type Address struct {
	name  string
	email string
}

func NewAddress(addr, name string) (Address, error) {
	if _, err := mail.ParseAddress(addr); err != nil {
		return Address{}, fmt.Errorf("email: parse address: %w", err)
	}
	return Address{name: name, email: addr}, nil
}

// Mailer is the port. One method — grow by composition, never past ~3.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}
```

## 4. Adapter skeleton (production + in-memory)

Adapter imports the domain, contains its third-party dep, translates errors inward, returns concrete `*T`.

```go
package smtp

type Mailer struct {
	dialer *gomail.Dialer
	from   string
}

func NewMailer(cfg Config) *Mailer { /* build dialer; SSL when Port==465 */ }

func (m *Mailer) Send(ctx context.Context, msg email.Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("smtp: validate: %w", err) // domain sentinel travels via %w
	}
	// ... build MIME, DialAndSend, wrap transport errors ...
}
```

The in-memory adapter is a real implementation that doubles as the test double:

```go
package memory

type Mailer struct {
	mu   sync.Mutex
	sent []email.Message
}

func NewMailer() *Mailer { return &Mailer{} }

func (m *Mailer) Send(_ context.Context, msg email.Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("memory: validate: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, msg)
	return nil
}

// Sent returns a defensive copy; Reset truncates. Assertion helpers for tests.
func (m *Mailer) Sent() []email.Message { /* copy under lock */ }
```

## 5. Interface segregation + composition (slice scale)

Grow a fat port from small ones so callers depend only on what they use:

```go
type (
	UserRecordReader interface {
		FindByID(ctx context.Context, id uuid.UUID) (User, error)
	}
	UserRecordWriter interface {
		Create(ctx context.Context, u User) error
	}
	// UserStore is the full port; one *postgres.Store satisfies it and the parts.
	UserStore interface {
		UserRecordReader
		UserRecordWriter
	}
)
```

## 6. Anticorruption layer (cross-context seam)

`auth` needs user data but must not import `user`. It declares `UserLookup`; the ACL adapts `user.Service` and translates errors + projects a narrow VO:

```go
package acl

// UserLookup adapts the user context to auth.UserLookup.
type UserLookup struct{ users user.Service }

func (l UserLookup) Find(ctx context.Context, id uuid.UUID) (auth.UserIdentity, error) {
	u, err := l.users.FindByID(ctx, id)
	if errors.Is(err, user.ErrNotFound) {
		return auth.UserIdentity{}, auth.ErrNotFound // translate neighbor's error
	}
	if err != nil {
		return auth.UserIdentity{}, fmt.Errorf("acl: user lookup: %w", err)
	}
	return auth.UserIdentity{ID: u.ID(), Email: u.Email().String()}, nil // narrow projection
}
```

## 7. Test harness (testcontainers, one container per package)

```go
// main_test.go
func TestMain(m *testing.M) {
	client, cleanup, err := testhelper.StartValkey()
	if err != nil { /* log + os.Exit(1) */ }
	sharedClient = client
	code := m.Run()
	cleanup()
	os.Exit(code)
}

// per test: t.Parallel(); isolate via uuid keys (no rollback) or BeginTx + t.Cleanup (SQL).
```

Gate integration tests behind `testing.Short()` or a build tag so `go test -short ./...` skips containers.

## 8. Spec-first doc

Every repo carries `docs/YYYY-MM-DD-spec.md` written so an agent can rebuild the module from it alone: module path, package/port list, an explicit **numbered invariants** list (e.g. "1. Domain imports stdlib + uuid only. 2. Ports never expose driver types. 3. No global state or init()."), the build order, and the adapter matrix (port × technology). Design rationale that would bloat the spec goes in sibling dated docs (`YYYY-MM-DD-design.md`, `YYYY-MM-DD-algorithms.md`).
