# webpush

`github.com/ralvarezdev/webpush` — Web Push notifications to browser subscriptions: the `Sender` and `SubscriptionStore` ports, a VAPID sender and a Postgres store. Infrastructure, the same category as email; not an identity concern.

## Packages

| Package | Owns |
|---|---|
| `webpush` | `Sender`, `SubscriptionStore`, `Subscription` (plus `UserAgent`, `ExpiresAt`), `Notification`, `Subscription.Validate`, `Notify`/`NotifyConfig`/`Result`/`Outcome`, `SendError`, `MaxPayloadSize`, the optional store interfaces `UserEndpointDeleter`, `StaleDeleter`, `OwnerReader`, `UserDeleter`, and the sentinels `ErrSubscriptionGone`, `ErrNoEndpoint`, `ErrNoKeys`, `ErrInvalidKeys`, `ErrInvalidEndpoint`, `ErrEndpointForbidden`, `ErrEndpointNotAllowed`, `ErrPayloadTooLarge`; no third-party imports |
| `webpush/vapid` | `vapid.New(vapid.Config{PublicKey, PrivateKey, Subscriber, TTL, Timeout, AllowedHosts, MaxRetryAfter, AllowPrivateNetworks, HTTPClient})` over `webpush-go`; `Sender.PublicKey()`, `vapid.GenerateKeys()` |
| `webpush/cmd/vapidgen` | `go run github.com/ralvarezdev/webpush/cmd/vapidgen` prints `VAPID_PUBLIC_KEY=` and `VAPID_PRIVATE_KEY=` |
| `webpush/postgres` | `postgres.NewStore(pool)` (`Save` upserts by endpoint, `FindByUserID`, `DeleteByEndpoint`, plus `DeleteByUserAndEndpoint`, `DeleteStale`, `Owner`, `DeleteByUserID`), sqlc over pgx; keys stored as `bytea` |
| `webpush/migrations`, `webpush/migrate` | embedded schema and `migrate.Up(ctx, db)`, version table `webpush_schema_version` |

## Recipe

```go
sender := vapid.New(vapid.Config{
    PublicKey:  cfg.VAPIDPublicKey,  // also ships to the frontend; the private key never leaves the backend
    PrivateKey: cfg.VAPIDPrivateKey,
    Subscriber: "mailto:ops@example.com",
})
store := postgres.NewStore(pool)

// POST /push/subscribe
_ = store.Save(ctx, webpush.Subscription{UserID: userID, Endpoint: body.Endpoint, Auth: body.Keys.Auth, P256dh: body.Keys.P256dh})

// Notify a user: load, send with bounded concurrency, prune gone or malformed subscriptions.
res, err := webpush.Notify(ctx, store, sender, userID, webpush.Notification{Title: "...", Body: "...", URL: "/inbox"})
// err only if the subscriptions could not be loaded; res.Outcomes has one entry per subscription.
log.Info("push", "sent", res.Sent(), "pruned", res.Pruned(), "failed", res.Failed())
// webpush.NotifyConfig{Concurrency: 8}.Notify(...) changes the default of 4.

// One send, by hand.
var se *webpush.SendError
if err := sender.Send(ctx, sub, n); errors.As(err, &se) {
    switch {
    case errors.Is(err, webpush.ErrSubscriptionGone): // 404/410: delete it
    case se.Permanent():                              // 400/401/403/413, bad endpoint or keys: do not retry
    case se.Temporary():                              // network, 429, 5xx: retry after se.RetryAfter
    }
}

// Refuse to take over an endpoint another user owns (Save transfers it to the caller).
if owner, ok, _ := store.Owner(ctx, ep); ok && owner != userID { /* 409 */ }
```

## Rules

- **Generate the VAPID keypair once** (`vapid.GenerateKeys()` or `cmd/vapidgen`) and keep it; changing it invalidates every existing subscription.
- **`Subscriber` is required.** Chrome and FCM silently drop delivery without the JWT `sub` claim. Either `mailto:you@example.com` or a URL works; the sender strips an existing `mailto:` so it is not doubled (Apple's push service rejects `mailto:mailto:`).
- **Check `Send`'s error.** `Send` returns nil only when the push service accepts the notification (2xx). Every failure is a `*webpush.SendError` (`Err`, `StatusCode`, `RetryAfter`; `Permanent()`/`IsPermanent()`, `Temporary()`; `IsPermanent` matches `email`'s). A 404 or 410 wraps `ErrSubscriptionGone`; 429, 5xx and network errors are temporary (`Retry-After` parsed on 429/503, clamped by `Config.MaxRetryAfter`, default 24h); 400/401/403/413 and other 4xx are permanent. The body is quoted in the message. Tags up to v0.4.3 did not read the response at all, so a rejected send returned nil; v0.5.0 and later report it.
- **`Notify` only loads, sends and prunes.** Who to notify and what to do about failed sends stays with the app. It prunes through `StaleDeleter` (keeps a subscription re-saved with fresh keys), else `UserEndpointDeleter`, else `DeleteByEndpoint`; prefer the user-scoped deletes in handlers so a stale delete cannot remove an endpoint someone else took over, and call `DeleteByUserID` on account deletion.
- **Endpoints are untrusted (SSRF).** `Validate` requires an https URL with a host and no credentials. The default client refuses at dial time (on the resolved address, so DNS rebinding fails) loopback, private, link-local, multicast, CGNAT, NAT64/6to4/Teredo-embedded and other non-global addresses (`ErrEndpointForbidden`), follows no redirects, ignores proxy env vars and times out after `Timeout` (10s). Set `AllowedHosts: []string{"fcm.googleapis.com", ...}` to restrict to push-service hosts and subdomains. `AllowPrivateNetworks` is for tests only; `HTTPClient` replaces the client, guard included.
- **`Save` ownership**: the endpoint is unique across users and `Save` transfers it to the caller (keys, `UserAgent`, `ExpiresAt` replaced). A user submitting a victim's endpoint cannot read anything but can cut the victim off, so save only endpoints from the authenticated user's own browser and consider checking `OwnerReader.Owner` first.
- **Payload limit**: JSON over `MaxPayloadSize` (3993 bytes) fails with `ErrPayloadTooLarge` before sending.
- **A subscription needs `Endpoint`, `Auth` and `P256dh`**; `Validate` reports `ErrNoEndpoint`/`ErrNoKeys`, and `ErrInvalidKeys` unless `Auth` is base64url of 16 bytes and `P256dh` a 65-byte on-curve P-256 point. `Save` and `Send` both validate; a `Send` with bad stored keys is a permanent error and `Notify` prunes it.
- **Apply the schema with `migrate.Up`** (own version table, four migrations incl. `bytea` keys and nullable `user_agent`/`expires_at`) before using the Postgres store.
- **Test against a real Postgres container** (`pgkit/pgtest`), not a mocked store.

## Consumers

None yet (as of the 2026-09-29 go.mod scan; not re-checked). finance-platform has the push-notification opt-in UI still to build and is the intended first consumer.

## Not here

Email ([EMAIL.md](EMAIL.md)), the service worker and the browser subscription flow (frontend), scheduling or fan-out.
