# identity

`github.com/ralvarezdev/identity` — one bounded context: who a user is and how they prove it. User accounts, authentication, sessions, TOTP/2FA, WebAuthn, RBAC and Personal Access Tokens. Do not write a users table, password hashing, JWT issuing or a session store in an app; wire this.

## Slices

| Package | Owns |
|---|---|
| `identity/user` | `User`, the `Email` value object (`NewEmail`), lifecycle `pending → active → suspended → deleted`, `Service`, `UserStore`; optional store capabilities `FieldUpdater`, `PasswordSwapper`, `DeleteFieldUpdater` (the Postgres store implements all, the service falls back to read-modify-write without) |
| `identity/auth` | token lifecycle, sessions, TOTP/2FA, RBAC, the driven ports, `Claims`, `Config` |
| `identity/auth/{jwt,password,totp,webauthn,postgres,cache/valkey}` | adapters: HS256 JWTs with opaque refresh tokens (`jwt.NewChecked`), bcrypt, `pquerna/otp` plus `totp.NewSecretBox` (AES-GCM), `go-webauthn`, Postgres stores, Valkey caches (`NewPermCache`, `NewPreAuth`, `NewThrottle`, `NewCeremonies`, `NewTOTPReplay`). The old go-redis `auth/cache` is gone |
| `identity/pat` | Personal Access Tokens: long-lived Bearer credentials resolving to the same `auth.Claims` as a session (`pat/postgres` is the store) |
| `identity/app` | the use cases an app calls: register, login, reset password, `CreatePAT`, `AuthenticatePAT`, `EnsureAdmin`, and the `EmailComposer` for mail content. Build it with `NewFromDeps(Deps)` (named fields, nil-checked, runs validation); `New` takes eighteen positional ports and checks nothing |
| `identity/migrate` | `Up`/`Down`/`Version`/`Status` for identity's schema (migrations 001-018), version table `identity_schema_version` |

## Wiring

`*app.App` is the only thing handlers and middleware talk to. Build it in a bootstrap package: `user.NewService(userpg.NewStore(db))`, `auth.NewService(...)` with the Postgres stores, the Valkey caches, `jwt.NewIssuer`, `password.New(bcrypt.DefaultCost)`, `totp`, and an `auth.Config` of TTLs, then `identityapp.New(userSvc, authSvc, stores..., mailer, identityapp.Config{...})`.

Prefer `identityapp.NewFromDeps(identityapp.Deps{Users, Store, PermCache, PreAuth, Issuer, Hasher, Limiter, TOTP, WebAuthn, Mailer, Config})`: one `*auth/postgres.Store` satisfies `Store` (all eight auth store ports), a missing dependency returns `ErrMissingDependency` naming the field, and it runs `LinkEmailComposer.Validate` and `auth.Service.Validate` (hardening misconfiguration fails at startup). It nil-checks interfaces only, so a typed nil pointer passes. Build the JWT issuer with `jwt.NewChecked` (secret >= 32 bytes, optional `Issuer`/`Audience`/`NotBefore`/`Leeway`); `jwt.New` is deprecated.

A full wiring covers the identity app, WebAuthn adapter, `LinkEmailComposer`, PAT store, `EnsureAdmin` seeding, Valkey adapters with a no-op fallback, and since v1.10 the ceremony store, TOTP replay guard and backup-code pepper. The package README also carries a full in-process quick start.

```go
identityApp := identityapp.New(/* userSvc, authSvc, stores, mailer */, identityapp.Config{
    PATs:      patpg.NewStore(identityDB), // omit and the PAT use cases return ErrPATsNotConfigured
    PATPrefix: "myapp_",                   // defaults to "pat_"
})

created, _ := identityApp.CreatePAT(ctx, userID, "ci", nil) // created.Raw is shown once
claims, err := identityApp.AuthenticatePAT(ctx, rawBearer)  // same auth.Claims as a cookie session
```

## Hardening (all opt-in, off by default)

`auth.Config` (reached as `app.Config.Auth`) holds the keys and flags; the optional ports go in `app.Config.AuthExtensions` (`auth.Extensions{CeremonyStore, TOTPReplayGuard, SecretBox}`; the Valkey ones need no `Deps` entry).

```go
box, _ := authtotp.NewSecretBox(totpKey) // 16/24/32 bytes, AES-GCM
cfg := identityapp.Config{
    Auth: auth.Config{
        RequireCeremonyStore: true,           // WebAuthn fails closed without a store
        BackupCodePepper:     backupPepper,   // >= auth.MinSecretKeyBytes (32), keep out of the DB
        MaxSessionAge:        30 * 24 * time.Hour,
        MaxCredentialsPerUser: 10,
    },
    AuthExtensions: auth.Extensions{
        CeremonyStore:   authvalkey.NewCeremonies(vk),    // single-use, GETDEL
        TOTPReplayGuard: authvalkey.NewTOTPReplay(vk, 0), // 0 = 90s
        SecretBox:       box,
    },
    ProtectLastAdmin: true,
    PATMaxTTL:        90 * 24 * time.Hour,
    AsyncMail:        true, // call identityApp.WaitMail on shutdown
}
```

| Option | Effect and caveats |
|---|---|
| `RequireCeremonyStore` / `CeremonySigningKey` | with a `CeremonyStore`, WebAuthn ceremonies are server-held and single-use. Without one, the ceremony is caller-held: unsigned and replayable (client can downgrade `UserVerification`), deprecated, one warning logged. `CeremonySigningKey` (>= 32 bytes, ignored when a store exists) HMAC-signs and expires the ceremony (`ErrCeremonyTampered`) but it stays replayable until expiry. `RequireCeremonyStore` makes every `Begin*`/`Finish*` return `ErrCeremonyStoreRequired` without a store. `ErrKeyTooShort` for a short key |
| `webauthn.Config.RequireUserVerification` | login demands PIN/biometric, so a passkey is a real second factor. Clone-detected sign counts fail with `ErrWebAuthnCloneDetected` |
| `BackupCodePepper` | new codes stored as `h2:` + HMAC-SHA256; legacy SHA-256 codes still validate until regenerated; no migration. Losing or changing the pepper invalidates peppered codes |
| `BackupCodeLength` | default 16 chars (80 bits, `xxxx-xxxx-xxxx-xxxx`); old 8-char codes still validate. `TOTPPort.GenerateBackupCodes` is no longer called |
| `MaxTwoFAAttempts` (5) | per-user cap on `CompleteTwoFA` guesses, counted atomically when the limiter is an `AttemptIncrementer` (Valkey `Throttle` is); exceeding it deletes the pre-auth token |
| `MaxSessionAge` | caps a rotating session from its first token (migration 017 `session_started_at`; older sessions measured from their latest token) |
| `MaxCredentialsPerUser` | `ErrTooManyCredentials` on `BeginWebAuthnRegistration` |
| `SecretBox` | TOTP secrets stored as `enc:v1:`+AES-GCM; plaintext legacy secrets still read |
| `ProtectLastAdmin` | `ErrLastAdmin` from `DeleteRole`, `RemoveRoleFromUser`, `SuspendAccount`, `DeleteAccount`, `AdminDeleteUser` (best-effort read-before-write; optional `RoleMemberCounter` makes it one query) |
| `PATMaxTTL` | `CreatePAT` needs an expiry within the cap (`ErrPATTTLExceeded`, wraps `pat.ErrInvalidExpiry`); zero is unlimited |
| `AsyncMail` (+ `AsyncMailWorkers` 8, `AsyncMailQueue` 64, `AsyncMailTimeout`, `MailErrorHook`) | reset and resend-verification mail sent from a bounded background job so latency does not reveal account existence; a full queue drops the job (`ErrMailQueueFull` to the hook) and the caller still gets nil |
| `LinkEmailComposer.TokenInFragment` | token in the URL fragment, out of logs and Referer; also serve the frontend with `Referrer-Policy: no-referrer`. `BaseURL` must be http(s) (`ErrInvalidBaseURL`) |

### What App applies (v1.10.0)

`auth.Service` has `VerifyTOTPCode` (normalises to six ASCII digits, opens a sealed secret, claims the code with the replay guard), `SealTOTPSecret`, `RegenerateBackupCodes` and the `...For` finishers; only some are reachable through `*app.App`:

- **Applied:** the login second factor (`CompleteTwoFA`, via `Login`/`CompleteTwoFA`) goes through `VerifyTOTPCode`, so the replay guard, `SecretBox` opening and 2FA attempt cap apply. Backup-code length, pepper and `RequireCeremonyStore` apply to every WebAuthn `Begin*`/`Finish*`. `FinishWebAuthnRegistrationFor` / `FinishWebAuthnLoginFor` exist on `App` and bind the ceremony to the user (`ErrCeremonyUserMismatch`; login `expectedUserID` may be empty only for discoverable flows); the plain `FinishWebAuthn*` are deprecated.
- **Not applied:** `App.EnableTOTP` calls `ValidateTOTPCode` (no replay claim) and `App.GenerateTOTP` stores `setup.Secret`, the plaintext, never `SealedSecret`, so enrolled secrets are not sealed at rest through `App` alone. `App.DisableTOTP` checks only the password. `App` has no `RegenerateBackupCodes` or `VerifyTOTPCode`; reach them by holding the `*auth.Service` yourself (it also satisfies `auth.TOTPVerifier`, `TOTPSecretSealer`, `BackupCodeRegenerator`) and enforce fresh re-authentication before regenerating codes or disabling TOTP.
- Role/permission edits do not invalidate the permission cache by themselves: call `Service.InvalidateAllPermCache` (`auth.PermCacheAllInvalidator`, `ErrUnsupported` when the cache cannot) after them. The Valkey cache is epoch-based, so a set read before an invalidation is never cached after it.

## Rules

- **HTTP stays in the app.** identity ships no handlers or middleware. Use [IDENTITYGIN.md](IDENTITYGIN.md) for gin; otherwise write the thin adapter over `app.App` yourself.
- **Email is a dependency.** identity mails verification, reset and change links through `github.com/ralvarezdev/email` (see [EMAIL.md](EMAIL.md)); pass a `Mailer`. The default `EmailComposer` mails the raw token as plain text, so set `LinkEmailComposer` to point at the frontend route.
- **Migrations**: run `identity/migrate.Up(ctx, db)` at startup or deploy; it has its own version table, so it coexists with the app's migrations. Migrations are forward-only: after a release that changes identity's schema, the previous app image may not run against it, so rehearse the upgrade on a copy of live data first. Since v1.6: 016 `refresh_tokens.last_used_at` (session last-used), 017 `session_started_at`, 018 drops the duplicate `refresh_tokens_rotated_from_idx` that 017 created.
- **Sessions and passwords.** `ListSessions` and `ChangePassword` take the caller's **raw** refresh token (`auth.HashToken` is exported if you need the hash). `ChangePassword` keeps that session, revokes the others and all PATs, and revokes every access token, so the caller must `Refresh`; with an empty or unknown token every session is revoked. Suspend, delete and admin-delete revoke credentials first, then write the status (retryable); `Delete` is legal from `pending`. A revoked refresh token only triggers theft handling (`ErrTokenReuse`, all sessions revoked) if it was already rotated; plain logout does not.
- **Errors do not leak account state.** `Authenticate` checks the password first and burns a dummy hash for unknown emails; `BeginWebAuthnLogin` answers `ErrInvalidCredentials` for unknown, inactive or credential-less accounts; `Refresh`, `CompleteTwoFA` and WebAuthn login reject disabled or unverified users with `ErrUserInactive`. `Register`/`Login` still reveal existence by design. Login throttle keys are `login:<lowercased email>`, 2FA keys `2fa:<user>`; a fixed window, so failures cannot extend a lockout.
- **Config compatibility.** `auth.Config`, `auth.Service` and `app.Config` are not comparable with `==`, and fields were reordered in v1.10: use keyed literals. `PasswordMinLen: 0` now means 12 bytes (negative disables); passwords over 72 bytes fail `ErrPasswordTooLong`. `UserLookup` must return `ErrNotFound` for a missing user.
- **One-time tokens.** Password change/reset, email-change request/confirm, resend-verification and verify-email revoke earlier unused tokens (needs an OTT store with `RevokeOTTs`; the Postgres one has `RevokeOTTsByUser`). An email-change token confirms whichever address is pending.
- **`EnsureAdmin`** is a no-op (password unchecked) for a verified account already holding the admin role; any repair (pending account, role missing) needs the matching password (`ErrAdminPasswordMismatch`, `ErrAdminAccountInactive` for a suspended one) and never reactivates it. Run it before serving traffic.
- **Known limitations.** Access tokens are not bound to sessions: a revoked session's access token lives until expiry plus `Leeway`, and `ValidateAccess` does not check user status. `LogoutOthers` is not atomic with a concurrent refresh.
- **Two pools, two Postgres roles.** Apps reach identity data only through identity's interface, enforced at the database level by separate roles and connection pools.
- **Authorization is separate from authentication.** identity answers who the caller is and their RBAC roles; an app derives its own permissions from that (for example a resolver that maps identity roles to app permissions).
- **Login throttling is not rate limiting.** `auth.AttemptLimiter` counts failed logins per account. API throughput is [RATELIMIT.md](RATELIMIT.md).
- **The domain imports no infrastructure**; adapters point inward (hexagonal). New storage or cache backends implement identity's ports in an adapter subpackage.
- **Never log or return credentials.** A PAT's raw value exists once, at creation; only its hash is stored. `PATTouchInterval` (default 1m) throttles `last_used_at` writes.

## Not here

Response shapes, HTTP middleware, cookies (identitygin), throughput limiting (ratelimit), email delivery (email).
