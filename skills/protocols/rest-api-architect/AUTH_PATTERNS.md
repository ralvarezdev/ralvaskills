# Authentication Patterns

Two patterns cover almost every API. Choose by team / system shape; framework-specific implementation lives in the matching framework architect skill ([fastapi-architect](../../frameworks/fastapi-architect/SKILL.md#6-authentication--authorization), [gin-architect](../../frameworks/gin-architect/SKILL.md#7-authentication--authorization), [nethttp-architect](../../frameworks/nethttp-architect/SKILL.md#8-authentication--authorization)).

## A. In-house OAuth2 password + JWT

For monoliths and single-service deployments. The service both issues tokens (on `POST /v1/auth/login`) and verifies them on every authenticated request.

- **Password hashing: Argon2id** — OWASP 2026 default. Never bcrypt or PBKDF2 for new code (acceptable only on platforms that lack Argon2 bindings). MD5 / SHA1 storage is a CVE waiting to happen.
- **JWT signing:** **HS256** (shared secret) for single-service. Switch to **RS256** (asymmetric) the moment a second service needs to verify the same tokens without holding the signing secret.
- **Token lifetimes:** short-lived access tokens (15 min). Pair with refresh tokens that are **opaque random strings stored server-side**, not JWTs — server-side storage is the only way to make revocation real.
- **Claims discipline:** minimal — `sub` (user id), `iat`, `exp`, `scopes` / `roles`. **Never embed PII** (email, name, address) in claims. Tokens end up in logs, and refresh-token responses go to the client.
- **Refresh flow:** `POST /v1/auth/refresh` with the opaque refresh token. Rotate refresh tokens on every use; revoke the old one. **Detect refresh-token reuse → revoke the entire chain** (token was stolen and the legitimate client just attempted with an old one).

## B. External IdP

(Keycloak / Auth0 / Cognito / Entra.) For any system with more than one service, MFA requirements, social login, or compliance demanding SSO.

- **Services only verify tokens; never issue them.** Identity lives in the IdP.
- **Verify via JWKS:** fetch `<idp>/.well-known/jwks.json`, validate the JWT's `kid` against a JWK in the set.
- **Cache JWKS** with a TTL (10 min default). On `kid` miss, refresh once — that's how key rotation is handled.
- **Audience and issuer checks are mandatory** — `aud` matches your service, `iss` matches your IdP's URL. Skipping them means any token from any IdP-issued client can hit your API.

## When to switch A → B

- A second service needs to trust the same user identity. (Two services each issuing tokens = identity drift, separate password stores, deeply confused users.)
- MFA, social login, or SSO is required and you don't want to build it.
- Compliance (HIPAA, SOC 2 Type 2) is in scope and you'd rather audit one IdP than every service.

## Authorization placement

**Per-endpoint, not middleware-global.** Scopes / roles are checked at the route after authentication, by a dependency or middleware-wrap specific to that endpoint. A blanket "authenticated users can do anything" middleware is the most common source of accidentally-public endpoints.
