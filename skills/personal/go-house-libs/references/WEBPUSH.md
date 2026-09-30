# webpush

`github.com/ralvarezdev/webpush` — Web Push notifications to browser subscriptions: the `Sender` and `SubscriptionStore` ports, a VAPID sender and a Postgres store. Infrastructure, the same category as email; not an identity concern.

## Packages

| Package | Owns |
|---|---|
| `webpush` | `Sender`, `SubscriptionStore`, `Subscription`, `Notification`, `Subscription.Validate`, `ErrSubscriptionGone`; no third-party imports |
| `webpush/vapid` | `vapid.New(vapid.Config{PublicKey, PrivateKey, Subscriber, TTL})` over `webpush-go` |
| `webpush/postgres` | `postgres.NewStore(pool)` (`Save` upserts by endpoint, `FindByUserID`, `DeleteByEndpoint`), sqlc over pgx |
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

// Notify a user: finding the recipient's subscriptions is the caller's job.
subs, _ := store.FindByUserID(ctx, userID)
for _, sub := range subs {
    err := sender.Send(ctx, sub, webpush.Notification{Title: "...", Body: "...", URL: "/inbox"})
    if errors.Is(err, webpush.ErrSubscriptionGone) {
        _ = store.DeleteByEndpoint(ctx, sub.Endpoint) // unsubscribed or expired; it will never work again
    }
}
```

## Rules

- **Generate the VAPID keypair once** (`webpushgo.GenerateVAPIDKeys()`) and keep it; changing it invalidates every existing subscription.
- **`Subscriber` is required.** Chrome and FCM silently drop delivery without the JWT `sub` claim. Either `mailto:you@example.com` or a URL works; the sender strips an existing `mailto:` so it is not doubled (Apple's push service rejects `mailto:mailto:`).
- **Check `Send`'s error.** `Send` returns nil only when the push service accepts the notification (2xx). A 404 or 410 wraps `ErrSubscriptionGone`, so delete the subscription; any other status is an error carrying the code and the start of the body (rejected VAPID token, throttling, oversized payload, outage). Tags up to v0.4.3 did not read the response at all, so a rejected send returned nil. This behavior is on `master` (commit `ebede8f`) and not yet tagged (2026-09-29): until a release carries it, check `git tag` and [../STACK.md](../STACK.md) before assuming `Send` reports rejections.
- **No "send to user" helper on purpose.** Who to notify and how to handle partial failure across a user's devices is app-specific.
- **A subscription needs `Endpoint`, `Auth` and `P256dh`**; `Validate` reports `ErrNoEndpoint`/`ErrNoKeys`.
- **Apply the schema with `migrate.Up`** (own version table) before using the Postgres store.
- **Test against a real Postgres container** (`pgkit/pgtest`), not a mocked store.

## Consumers

None yet (go.mod scan, 2026-09-29). finance-platform has the push-notification opt-in UI still to build and is the intended first consumer.

## Not here

Email ([EMAIL.md](EMAIL.md)), the service worker and the browser subscription flow (frontend), scheduling or fan-out.
