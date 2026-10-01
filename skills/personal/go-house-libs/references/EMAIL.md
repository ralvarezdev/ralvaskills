# email

`github.com/ralvarezdev/email` — sending email as a mechanism, independent of any domain: the `Mailer` port, the `Message` payload and a validated `Address`. The root package imports only `golang.org/x/net/idna`; only the SMTP adapter imports a mail library (go-mail). The module has a README, a CHANGELOG and godoc.

## Packages

| Package | Owns |
|---|---|
| `email` | `Mailer` (`Send(ctx, Message) error`), `Message` (incl. optional `MessageID`), `Address` (`NewAddress(addr, name)`), `Message.Validate` (`ErrNoRecipients`, `ErrNoBody`, `ErrInvalidMessageID`, joined), `Classified` (`IsPermanent()`), `ErrNonASCIIAddress` |
| `email/smtp` | production over go-mail: `smtp.NewMailer(smtp.Config{From, Host, Port, Username, Password, TLS, TLSConfig, Timeout, LocalName, Logger})`, `TLSMode` (`TLSAuto`, `TLSStartTLSRequired`, `TLSImplicit`, `TLSNone`), multipart/alternative; `Config.Validate()`/`Warnings()`; `*SendError{Permanent}` with `IsPermanent()`/`Temporary()`; `ErrTLSRequired` |
| `email/logger` | development: `logger.NewMailer(log, ...Option)` logs the message instead of delivering; `logger.WithPreviewRunes(n)` (default 200, negative disables the body preview) |
| `email/memory` | tests: `memory.NewMailer()` captures messages; `Sent()` returns them, `Reset()` clears |

## Recipe

```go
from, err := email.NewAddress("no-reply@example.com", "Example")
to, err := email.NewAddress(userEmail, "")

msg := email.Message{From: from, To: []email.Address{to}, Subject: "Reset your password", HTML: html, Text: text, MessageID: "<reset-" + id + "@example.com>"} // MessageID optional

var mailer email.Mailer = logger.NewMailer(slog.Default())           // dev
cfg := smtp.Config{From: from, Host: "smtp.example.com", Port: 587, Username: u, Password: p, TLS: smtp.TLSStartTLSRequired}
if err := cfg.Validate(); err != nil { return err } // NewMailer does not validate; call it at startup
mailer = smtp.NewMailer(cfg) // prod

err = mailer.Send(ctx, msg)

var c email.Classified // adapter-neutral; webpush's SendError has the same IsPermanent()
if errors.As(err, &c) && !c.IsPermanent() { /* transient: retry later with the same MessageID */ }
```

## Rules

- **Depend on `email.Mailer`, not an adapter.** Choose the adapter in the bootstrap package by environment; handlers and domain code only see the port.
- **Never use a raw string as an address.** Build an `Address` with `NewAddress` so a malformed one fails at construction. It stores the parsed address (`"Name <a@b.c>"` works), converts an IDN domain to punycode and rejects a non-ASCII local part with `ErrNonASCIIAddress` (no SMTPUTF8).
- **At least one of `HTML` or `Text`, and at least one `To`**, or `Send` fails validation. Set `Text` as the plaintext fallback when sending HTML.
- **Set `From` explicitly.** Only the SMTP adapter substitutes `Config.From` for a zero `From`; `logger` and `memory` do not, so relying on the default behaves differently per adapter.
- **Permanent errors are not retried.** Permanent: SMTP 5xx, certificate verification failures, `ErrTLSRequired`, no usable AUTH mechanism. Transient: 4xx, network errors, context cancellation or deadline.
- **Use `smtp.TLSStartTLSRequired` (587) or `TLSImplicit` (465) in production.** `TLSAuto` (the zero value) is opportunistic: without credentials a stripped STARTTLS means cleartext delivery; with credentials `Send` fails with the permanent `ErrTLSRequired` before any AUTH byte (v0.4.0 fix; v0.3.x authenticated with SCRAM/CRAM-MD5 in cleartext). `TLSNone` is the only mode that authenticates unencrypted, for local dev. `Config.Warnings()` flags `TLSAuto` with credentials off port 465; `NewMailer` logs it.
- **`ctx` is honored for the whole SMTP exchange** (since v0.3.0, go-mail replaced gomail): cancel or deadline closes the connection and returns a transient `*SendError` wrapping the context error. `Config.Timeout` (default 30s) also caps each `Send`.
- **A reset or timeout after DATA is ambiguous**: the server may have accepted the message. Set the same `Message.MessageID` on every retry so receivers can deduplicate; without it each `Send` generates a new Message-ID.
- **Test with `memory`.** Assert on `Sent()`; do not stub `Send` by hand.
- **`logger` is development only**: it logs subject and a body preview, which hold reset tokens and magic links. Never ship it to production.
- **Real delivery is a deployment step:** production needs the SMTP credentials wired and the `logger` mailer removed. A service still on `logger` is not sending mail.

## Reference consumers

finance-platform wires `email/logger` in `backend/internal/bootstrap/bootstrap.go` (real SMTP still pending as of the last check; set `TLS` and call `Config.Validate()` when wiring it); identity sends its verification and reset mail through this port, so any app on identity needs a `Mailer` ([IDENTITY.md](IDENTITY.md)).

## Not here

Templating, queues and retry policy (the app owns those), push notifications ([WEBPUSH.md](WEBPUSH.md)).
