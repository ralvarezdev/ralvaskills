# identity

`github.com/ralvarezdev/identity` — one bounded context: who a user is and how they prove it. User accounts, authentication, sessions, TOTP/2FA, WebAuthn, RBAC and Personal Access Tokens. Do not write a users table, password hashing, JWT issuing or a session store in an app; wire this.

## Slices

| Package | Owns |
|---|---|
| `identity/user` | `User`, the `Email` value object (`NewEmail`), lifecycle `pending → active → suspended → deleted`, `Service`, `UserStore` |
| `identity/auth` | token lifecycle, sessions, TOTP/2FA, RBAC, the driven ports, `Claims`, `Config` |
| `identity/auth/{jwt,password,totp,webauthn,postgres,cache/valkey}` | adapters: HS256 JWTs with opaque refresh tokens, bcrypt, `pquerna/otp`, `go-webauthn`, Postgres stores, Valkey caches |
| `identity/pat` | Personal Access Tokens: long-lived Bearer credentials resolving to the same `auth.Claims` as a session (`pat/postgres` is the store) |
| `identity/app` | the use cases an app calls: register, login, reset password, `CreatePAT`, `AuthenticatePAT`, `EnsureAdmin`, and the `EmailComposer` for mail content |
| `identity/migrate` | `Up`/`Down`/`Version`/`Status` for identity's schema, version table `identity_schema_version` |

## Wiring

`*app.App` is the only thing handlers and middleware talk to. Build it in a bootstrap package: `user.NewService(userpg.NewStore(db))`, `auth.NewService(...)` with the Postgres stores, the Valkey caches, `jwt.NewIssuer`, `password.New(bcrypt.DefaultCost)`, `totp`, and an `auth.Config` of TTLs, then `identityapp.New(userSvc, authSvc, stores..., mailer, identityapp.Config{...})`.

Reference consumer: finance's `backend/internal/bootstrap/bootstrap.go` (identity app, WebAuthn adapter, `LinkEmailComposer`, PAT store, `EnsureAdmin` seeding, Valkey adapters with a no-op fallback). Read it before wiring a second app. The package README also carries a full in-process quick start.

```go
identityApp := identityapp.New(/* userSvc, authSvc, stores, mailer */, identityapp.Config{
    PATs:      patpg.NewStore(identityDB), // omit and the PAT use cases return ErrPATsNotConfigured
    PATPrefix: "myapp_",                   // defaults to "pat_"
})

created, _ := identityApp.CreatePAT(ctx, userID, "ci", nil) // created.Raw is shown once
claims, err := identityApp.AuthenticatePAT(ctx, rawBearer)  // same auth.Claims as a cookie session
```

## Rules

- **HTTP stays in the app.** identity ships no handlers or middleware. Use [IDENTITYGIN.md](IDENTITYGIN.md) for gin; otherwise write the thin adapter over `app.App` yourself.
- **Email is a dependency.** identity mails verification, reset and change links through `github.com/ralvarezdev/email` (see [EMAIL.md](EMAIL.md)); pass a `Mailer`. The default `EmailComposer` mails the raw token as plain text, so set `LinkEmailComposer` to point at the frontend route.
- **Migrations**: run `identity/migrate.Up(ctx, db)` at startup or deploy; it has its own version table, so it coexists with the app's migrations. Migrations are forward-only: after a release that changes identity's schema, the previous app image may not run against it, so rehearse the upgrade on a copy of live data first.
- **Two pools, two Postgres roles.** Apps reach identity data only through identity's interface, enforced at the database level by separate roles and connection pools.
- **Authorization is separate from authentication.** identity answers who the caller is and their RBAC roles; an app derives its own permissions from that (finance's `authz` resolves roles through identity).
- **Login throttling is not rate limiting.** `auth.AttemptLimiter` counts failed logins per account. API throughput is [RATELIMIT.md](RATELIMIT.md).
- **The domain imports no infrastructure**; adapters point inward (hexagonal). New storage or cache backends implement identity's ports in an adapter subpackage.
- **Never log or return credentials.** A PAT's raw value exists once, at creation; only its hash is stored.

## Not here

Response shapes, HTTP middleware, cookies (identitygin), throughput limiting (ratelimit), email delivery (email).
