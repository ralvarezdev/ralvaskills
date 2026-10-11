# viperkit

`github.com/ralvarezdev/viperkit` — file-sourced secrets for viper configs: for an environment variable `VAR`, a set `VAR_FILE` has its contents become `VAR` at config-load time. This is the `<VAR>_FILE` indirection Docker Compose `secrets:` and systemd `LoadCredential=` need. Reach for it before hand-writing `<VAR>_FILE` handling or letting a secret live only in the environment.

## Packages

- `viperkit` (stdlib only): `FileSecretName`, `ResolveFileSecret`, `ResolveFileSecrets`.
- `viperconfig` (owns viper): `RejectFileSecrets`, `BindEnv`.

## Recipe

```go
// before the env bindings run
if err := viperkit.ResolveFileSecrets("FINANCE_DATABASE_DSN", "FINANCE_AUTH_JWT_SECRET"); err != nil {
    return err
}

// with viper: reject a committed secret, and resolve+bind in one call
if err := viperconfig.RejectFileSecrets(v, "auth.jwt_secret", "database.dsn"); err != nil {
    return err
}
if err := viperconfig.BindEnv(v, "auth.jwt_secret", "FINANCE_AUTH_JWT_SECRET"); err != nil {
    return err
}
```

## Rules

- **Resolve before binding.** `_FILE` resolution must run before the env bindings read the variables, so a file-sourced value and an inline one are indistinguishable to the rest of the config.
- **Unset/empty `<VAR>_FILE` is a no-op; a set but unreadable file is an error** — fail closed rather than run without the secret. A trailing `\r\n` is trimmed.
- **Only the app/bootstrap layer calls it.** Domain libraries never read env or viper ([go-library-builder §5](../go-library-builder/SKILL.md)).
- **The root stays stdlib-only;** anything that imports viper goes in the `viperconfig` subpackage, so a consumer that only needs file resolution never compiles viper.
- **The consumer owns the key list.** `viperkit` resolves whatever variables it is given; which keys are secrets is the app's decision.

## Not here

Config-file parsing, secret sources other than a file (Vault, cloud KMS), and non-viper config frameworks.
