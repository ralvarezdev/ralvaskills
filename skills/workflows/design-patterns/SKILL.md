---
name: design-patterns
version: 1.0.0
description: Pragmatic catalog of design patterns for modern Go and Python — keeps the few that earn their place (repository, adapter, strategy, decorator, observer/pub-sub, builder for complex construction), flags the rest as anti-patterns. Use when deciding whether to introduce a pattern, refactoring toward one, or pushing back on cargo-cult patterns.
---

# Design Patterns — Skeptical, Pragmatic Catalog

Most "design patterns" exist to work around missing language features. **Go and Python in 2026 don't need most of them.** This skill names what's still useful, what's covered by other skills, and what to actively avoid. Companion to [ddd-architect](../../design/ddd-architect/SKILL.md), [hexagonal-arch](../../design/hexagonal-arch/SKILL.md), [logic-cleaner](../logic-cleaner/SKILL.md), and [improve-codebase-architecture](../improve-codebase-architecture/SKILL.md) — those skills carry more domain weight; this one is the catalog you reach for when someone says "should we apply pattern X here?".

## 1. The framing

Patterns aren't free. Each one:

- Adds a name a future reader has to learn.
- Adds at least one type or function whose only purpose is the pattern.
- Promises a benefit (testability, swappability, polymorphism) that may not pay back if the abstraction has one implementation forever.

**Never introduce a pattern speculatively.** Wait for the *third* place that wants the same behavior (rule of three, per [logic-cleaner §6](../logic-cleaner/SKILL.md#6-duplication-rule-of-three)). One repetition is fine; two might be coincidence; three is a pattern asking to be extracted.

The "Gang of Four" book is a museum piece. Most of its patterns were workarounds for Java's lack of first-class functions, sum types, and structural typing. Modern Go has interfaces-where-used; modern Python has Protocols, dataclasses, and `match`. Neither needs Visitor, Abstract Factory, or Singleton-as-language-feature.

## 2. Patterns still worth knowing

### Repository

**One per aggregate root**, abstracts persistence behind collection semantics (`find_by_id`, `add`, `remove`). Depth lives in [ddd-architect §5](../../design/ddd-architect/SKILL.md#5-repositories). Implementation lives in [sql-architect](../../databases/sql-architect/SKILL.md) — raw SQL via `psycopg`/`sqlx`, no ORM.

**Use when:** the domain has aggregates with identity that need persistence.
**Skip when:** there's no domain layer — just functions that read and write rows.

### Adapter (ports & adapters)

Wrap an external dependency in an interface that the consumer defines. Full pattern lives in [hexagonal-arch](../../design/hexagonal-arch/SKILL.md). Rule: *one adapter is a hypothetical seam; two adapters are a real one*.

**Use when:** the dependency has at least two real implementations (typically prod + test fake) and the surface is small enough to express as an interface.
**Skip when:** there's only one implementation and no plausible second one.

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

**Use when:** behavior genuinely varies per caller / per context / per environment.
**Skip when:** there's one strategy and a config flag would have done it.

### Decorator (middleware-style wrapping)

Wrap a function or handler to add cross-cutting concerns (logging, timing, retry, auth). Idiomatic in both languages — Python `@decorator`; Go `func(http.Handler) http.Handler` middleware chain (per [nethttp-architect §7](../../frameworks/nethttp-architect/SKILL.md#7-middleware--function-wrapping)).

**Use when:** the concern is genuinely cross-cutting and orthogonal to the wrapped logic.
**Skip when:** there's only one wrapped function — just inline the concern.

### Observer / pub-sub

Notify many subscribers when something happens. Modern shape: **domain events** (see [ddd-architect §6](../../design/ddd-architect/SKILL.md#6-domain-events)) for in-process or via a message bus for cross-service.

**Use when:** multiple unknown consumers care about a state change, or you want to decouple producers from consumers.
**Skip when:** there's exactly one consumer — call it directly.

### Builder (only when construction is complex)

For objects with many optional parameters where the constructor would have 12 arguments. In Go: **functional options** are the idiomatic equivalent (`WithTimeout(d)`, `WithRetries(n)`). In Python: dataclass defaults + `__init__` keyword arguments cover almost every case.

**Use when:** construction has genuinely many optional parameters, validation rules, or multi-step assembly.
**Skip when:** 4 args fit on one line.

### Factory method (only when construction depends on runtime data)

A function that decides which concrete type to return based on input. Useful for parsing (parse this config blob into the right handler type), plugin systems, and discriminated unions.

**Use when:** the caller can't know the concrete type at compile time.
**Skip when:** the caller knows — just call the constructor directly.

### Iterator (already a language feature)

Go has `iter.Seq[T]` (1.23+). Python has generators. **Don't reach for an Iterator pattern as a separate class** — use the language feature. Detailed iterator guidance lives in [go-architect §6](../../languages/go-architect/SKILL.md#6-iterators-go-123).

## 3. Patterns to actively avoid (or replace)

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

Centralizes communication between objects "to reduce coupling." Usually grows into the largest class in the codebase. Prefer **domain events** for cross-aggregate coordination, or just direct calls when objects naturally depend.

### Template Method (inheritance-based)

`AbstractClass.run()` calls `step1()` / `step2()` / `step3()`, subclasses override the steps. Replaced by **composition + functions/strategies passed in**. Inheritance for behavior reuse is a code smell — prefer interface satisfaction.

### Prototype / Memento

Mostly historical. Modern languages have native cloning (`copy.deepcopy`, Go's `*new(T)` or explicit copy constructors) and serialization (`json.dumps`, `encoding/json`). No pattern needed.

## 4. Patterns deferred to other skills

These are real patterns; this skill names them once and points to their owner:

| Pattern | Owner skill |
|---|---|
| Repository | [ddd-architect](../../design/ddd-architect/SKILL.md) §5 |
| Aggregate | [ddd-architect](../../design/ddd-architect/SKILL.md) §3–4 |
| Value Object | [ddd-architect](../../design/ddd-architect/SKILL.md) §3 |
| Domain Event | [ddd-architect](../../design/ddd-architect/SKILL.md) §6 |
| Anticorruption Layer | [ddd-architect](../../design/ddd-architect/SKILL.md) §8 |
| Ports & Adapters | [hexagonal-arch](../../design/hexagonal-arch/SKILL.md) |
| Middleware chain | [nethttp-architect](../../frameworks/nethttp-architect/SKILL.md) §7, [gin-architect](../../frameworks/gin-architect/SKILL.md) §9, [fastapi-architect](../../frameworks/fastapi-architect/SKILL.md) §8 |
| Functional options (Go) | [go-architect](../../languages/go-architect/SKILL.md) §3 |
| Dependency injection | [go-architect](../../languages/go-architect/SKILL.md) §11, [python-architect](../../languages/python-architect/SKILL.md) §3 |
| Iterator / generator | [go-architect](../../languages/go-architect/SKILL.md) §6 |

Don't restate them here — load the owner skill when the topic comes up.

## 5. Decision rules

When asked "should we apply pattern X?":

1. **Is the problem actually there yet?** Has it appeared at least three times (rule of three)? If not — wait. Premature patterns are the most expensive kind of debt because they make the existing code more confusing without solving a real problem.
2. **Would a simpler language feature do?** A function value usually beats Strategy. `match` beats Visitor. A goroutine + channel beats Producer-Consumer-with-buffer-class. Reach for patterns *after* you've exhausted what the language gives you for free.
3. **Will there be a second implementation?** Patterns that introduce interfaces (Adapter, Strategy, Repository) only pay back if at least two implementations exist. One adapter is a hypothetical seam — and hypothetical seams are pure cost.
4. **Does it make the code easier to delete?** Good patterns concentrate change — touch one place when the rule evolves. Bad patterns spread change — every variation needs N synchronized files updated.
5. **Does the name carry weight?** "It's a Repository" tells a future reader the persistence shape immediately. "It's a Manager" / "Helper" / "Service" tells them nothing. Use named patterns when the name is genuinely informative; invent ad-hoc structure when no standard name fits.

If the answer to (1) is "no" or to (3) is "no", **skip the pattern** and revisit when reality has caught up.

## 6. Code-smell → pattern map

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

The pattern is the *option*, not the answer. Always check the decision rules in §5 first.

## 7. Anti-patterns specific to Go and Python

**Go:**

- **`interface{}` (now `any`) used liberally** — kills type safety; usually a sign of premature abstraction. Stay typed unless you genuinely need dynamic dispatch.
- **Empty struct receivers** (`func (Receiver) Method() {}`) used to "namespace" functions — just use package-level functions.
- **Mutex + atomic mixed in the same struct** — pick one consistency model per field.
- **Returning interface types from constructors** — go-architect §4 says return concrete types; let callers narrow. Returning interfaces forces every adapter to satisfy them.

**Python:**

- **`@staticmethod` everywhere on a class with no state** — it's a module, not a class. Use module-level functions.
- **Inheritance for behavior reuse** — prefer composition + Protocol satisfaction.
- **Metaclasses for "magic" registration** — fragile, surprising. Explicit registration is clearer.
- **`*args, **kwargs` passthrough that hides the actual signature** — make the contract explicit.

## 8. When a pattern hardens into a convention

If a pattern appears across many skills and projects in this repo (e.g. Repository everywhere, Middleware in every HTTP framework), it stops being a "design pattern" decision and becomes a **convention** encoded in the architect skill for that area. Read the architect skill first; this catalog is for when you're in genuinely new territory.

The architect skills already enforce the patterns we use universally. This skill exists for the edge cases where the architect skills don't pre-decide, and for pushing back when someone wants to add a pattern without a real reason.
