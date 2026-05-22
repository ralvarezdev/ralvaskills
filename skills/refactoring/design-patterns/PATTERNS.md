# Design Patterns — Detailed Catalog

Detail beyond the SKILL.md summary. Loaded on demand when you need code snippets or the full anti-pattern rationale.

## 1. Patterns worth knowing — detail

### Repository

**One per aggregate root.** Abstracts persistence behind collection semantics (`find_by_id`, `add`, `remove`). Depth lives in [ddd-architect §5](../../design/ddd-architect/SKILL.md#5-repositories). Implementation lives in [sql-architect](../../databases/sql-architect/SKILL.md) — raw SQL via `psycopg`/`sqlx`, no ORM.

- **Use when:** the domain has aggregates with identity that need persistence.
- **Skip when:** there's no domain layer — just functions that read and write rows.

### Adapter (ports & adapters)

Wrap an external dependency in an interface that the consumer defines. Full pattern in [hexagonal-arch](../../design/hexagonal-arch/SKILL.md). **One adapter is a hypothetical seam; two adapters are a real one.**

- **Use when:** the dependency has at least two real implementations (typically prod + test fake) and the surface is small enough to express as an interface.
- **Skip when:** there's only one implementation and no plausible second one.

### Strategy

A behavior parameter that varies at runtime. In Go and Python, **a function value** is usually enough — no need for a `Strategy` class hierarchy.

```python
def sort_users(users: list[User], compare: Callable[[User, User], int]) -> list[User]:
    return sorted(users, key=cmp_to_key(compare))
```

```go
func SortUsers(users []User, less func(a, b User) bool) {
    sort.Slice(users, func(i, j int) bool { return less(users[i], users[j]) })
}
```

- **Use when:** behavior genuinely varies per caller / per context / per environment.
- **Skip when:** there's one strategy and a config flag would have done it.

### Decorator (middleware-style wrapping)

Wrap a function or handler to add cross-cutting concerns (logging, timing, retry, auth). Idiomatic in both languages — Python `@decorator`; Go `func(http.Handler) http.Handler` middleware chain (per [nethttp-architect §7](../../frameworks/nethttp-architect/SKILL.md#7-middleware--function-wrapping)).

- **Use when:** the concern is genuinely cross-cutting and orthogonal to the wrapped logic.
- **Skip when:** there's only one wrapped function — just inline the concern.

### Observer / pub-sub

Notify many subscribers when something happens. Modern shape: **domain events** ([ddd-architect §6](../../design/ddd-architect/SKILL.md#6-domain-events)) for in-process or via a message bus for cross-service.

- **Use when:** multiple unknown consumers care about a state change, or you want to decouple producers from consumers.
- **Skip when:** there's exactly one consumer — call it directly.

### Builder (only when construction is complex)

For objects with many optional parameters where the constructor would have 12 arguments. In Go: **functional options** are the idiomatic equivalent (`WithTimeout(d)`, `WithRetries(n)`). In Python: dataclass defaults + `__init__` keyword arguments cover almost every case.

- **Use when:** construction has genuinely many optional parameters, validation rules, or multi-step assembly.
- **Skip when:** 4 args fit on one line.

### Factory method (only when construction depends on runtime data)

A function that decides which concrete type to return based on input. Useful for parsing (parse this config blob into the right handler type), plugin systems, and discriminated unions.

- **Use when:** the caller can't know the concrete type at compile time.
- **Skip when:** the caller knows — just call the constructor directly.

### Iterator (already a language feature)

Go has `iter.Seq[T]` (1.23+). Python has generators. **Don't reach for an Iterator pattern as a separate class** — use the language feature. Detail in [go-architect §6](../../languages/go-architect/SKILL.md#6-iterators-go-123).

## 2. Patterns to avoid (or replace) — detail

### Singleton

**Never.** Globals are anti-patterns per [go-architect §11](../../languages/go-architect/SKILL.md#11-dependencies--logging) and [python-architect §3](../../languages/python-architect/SKILL.md#3-interfaces--di). Use dependency injection — pass the instance to constructors. The "convenience" of `Singleton.getInstance()` makes tests painful and hides the dependency graph.

**Exceptions:** language-runtime singletons you don't control (`os.Stdout`, `logging.getLogger()`). Even those are usually accessed via a small DI-friendly wrapper.

### Abstract Factory

A factory of factories. Over-engineering in modern code; you almost never need it. If you find yourself reaching for it, you probably want a discriminated union + a parse function instead.

### Visitor

Workaround for languages without sum types or pattern matching. **Python `match` and Go type switches kill it.**

```python
match shape:
    case Circle(r): area = math.pi * r * r
    case Square(s): area = s * s
    case Triangle(b, h): area = 0.5 * b * h
```

No `accept(visitor)`, no `VisitCircle`, no `ConcreteVisitor`. Direct, exhaustive, readable.

### Chain of Responsibility

Replaced by middleware chains (`func(handler) handler`), per [nethttp-architect §7](../../frameworks/nethttp-architect/SKILL.md#7-middleware--function-wrapping) and Gin/FastAPI equivalents. Same idea, less ceremony.

### Mediator (god object in disguise)

Centralizes communication between objects "to reduce coupling." Usually grows into the largest class in the codebase. Prefer **domain events** for cross-aggregate coordination, or direct calls when objects naturally depend.

### Template Method (inheritance-based)

`AbstractClass.run()` calls `step1()` / `step2()` / `step3()`, subclasses override the steps. Replaced by **composition + functions/strategies passed in**. Inheritance for behavior reuse is a code smell — prefer interface satisfaction.

### Prototype / Memento

Mostly historical. Modern languages have native cloning (`copy.deepcopy`, Go's `*new(T)` or explicit copy constructors) and serialization (`json.dumps`, `encoding/json`). No pattern needed.

## 3. Code-smell → pattern map

When the smell appears, the named pattern is often (not always) the right escape:

| Smell | Pattern to consider |
|---|---|
| `if isinstance(x, A): ... elif isinstance(x, B): ...` repeated in multiple places | Sum type + `match` (Python) or interface (Go); don't build a Visitor |
| Constructor with 8+ arguments, many optional | Builder (Python keyword args usually enough) / functional options (Go) |
| Five copies of "do thing X, log it, time it, retry it, do thing X again" | Decorator / middleware |
| Multiple places do the same DB query | Repository method |
| Cross-cutting "every service needs to handle authentication" | Decorator / middleware on the entry handlers — not in every service |
| One class with 30 methods of unrelated concerns | Decompose into per-concern modules — not a Facade |
| Several similar callbacks fired on the same event | Observer / domain event |
| Global state that "needs to be accessed everywhere" | Inject it; the urge to make it global is the bug |
| Subclass hierarchy with one method overridden per subclass | Strategy via function value, not inheritance |
| Manual switch on type to dispatch behavior | Interface method + polymorphism (Go) / `match` (Python) — not Factory |

The pattern is the *option*, not the answer. Always check the decision rules in [SKILL §5](SKILL.md#5-decision-rules) first.

## 4. Language-specific anti-patterns

### Go

- **`interface{}` (now `any`) used liberally** — kills type safety; usually a sign of premature abstraction. Stay typed unless you genuinely need dynamic dispatch.
- **Empty struct receivers** (`func (Receiver) Method() {}`) used to "namespace" functions — just use package-level functions.
- **Mutex + atomic mixed in the same struct** — pick one consistency model per field.
- **Returning interface types from constructors** — go-architect §4 says return concrete types; let callers narrow. Returning interfaces forces every adapter to satisfy them.

### Python

- **`@staticmethod` everywhere on a class with no state** — it's a module, not a class. Use module-level functions.
- **Inheritance for behavior reuse** — prefer composition + Protocol satisfaction.
- **Metaclasses for "magic" registration** — fragile, surprising. Explicit registration is clearer.
- **`*args, **kwargs` passthrough that hides the actual signature** — make the contract explicit.
