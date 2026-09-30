# email

`github.com/ralvarezdev/email` — sending email as a mechanism, independent of any domain: the `Mailer` port, the `Message` payload and a validated `Address`. The root package has no external dependencies; only the SMTP adapter imports a mail library (gomail). The module has no README, so this file and `go doc` are the reference.

## Packages

| Package | Owns |
|---|---|
| `email` | `Mailer` (`Send(ctx, Message) error`), `Message`, `Address` (`NewAddress(addr, name)`), `Message.Validate` (`ErrNoRecipients`, `ErrNoBody`, joined) |
| `email/smtp` | production: `smtp.NewMailer(smtp.Config{From, Host, Port, Username, Password})`, STARTTLS on 587 or implicit TLS on 465, multipart/alternative; `*SendError{Permanent}` |
| `email/logger` | development: `logger.NewMailer(slog.Logger)` logs the message instead of delivering |
| `email/memory` | tests: `memory.NewMailer()` captures messages; `Sent()` returns them, `Reset()` clears |

## Recipe

```go
from, err := email.NewAddress("no-reply@example.com", "Example")
to, err := email.NewAddress(userEmail, "")

msg := email.Message{From: from, To: []email.Address{to}, Subject: "Reset your password", HTML: html, Text: text}

var mailer email.Mailer = logger.NewMailer(slog.Default())           // dev
mailer = smtp.NewMailer(smtp.Config{From: from, Host: "smtp.example.com", Port: 587, Username: u, Password: p}) // prod

err = mailer.Send(ctx, msg)

var sendErr *smtp.SendError
if errors.As(err, &sendErr) && !sendErr.Permanent { /* 4xx: retry later */ }
```

## Rules

- **Depend on `email.Mailer`, not an adapter.** Choose the adapter in the bootstrap package by environment; handlers and domain code only see the port.
- **Never use a raw string as an address.** Build an `Address` with `NewAddress` so a malformed one fails at construction.
- **At least one of `HTML` or `Text`, and at least one `To`**, or `Send` fails validation. Set `Text` as the plaintext fallback when sending HTML.
- **Set `From` explicitly.** Only the SMTP adapter substitutes `Config.From` for a zero `From`; `logger` and `memory` do not, so relying on the default behaves differently per adapter.
- **Permanent errors are not retried.** `SendError.Permanent` is true for SMTP 5xx; 4xx are transient.
- **`ctx` is accepted but not propagated to the SMTP dial** (gomail does not support it), so set deadlines at the caller, not through the context.
- **Test with `memory`.** Assert on `Sent()`; do not stub `Send` by hand.
- **Real delivery is a deployment step:** production needs the SMTP credentials wired and the `logger` mailer removed. A service still on `logger` is not sending mail.

## Reference consumers

finance-platform wires `email/logger` in `backend/internal/bootstrap/bootstrap.go` (real SMTP still pending); identity sends its verification and reset mail through this port, so any app on identity needs a `Mailer` ([IDENTITY.md](IDENTITY.md)).

## Not here

Templating, queues and retry policy (the app owns those), push notifications ([WEBPUSH.md](WEBPUSH.md)).
