# MCP Server Plan — `rsk mcp`

Diseño e implementación de un servidor MCP que permite al agente **descubrir skills por proyecto** e **instalarlas**, expuesto como subcomando del binario `rsk`.

Estado: plan aprobado en diseño, pendiente de implementación.
Última revisión: 2026-10-02

---

## Table of Contents

1. [Contexto y problema](#1-contexto-y-problema)
2. [Decisiones de diseño](#2-decisiones-de-diseño)
3. [Alcance y no-alcance](#3-alcance-y-no-alcance)
4. [Arquitectura](#4-arquitectura)
5. [Primitivas MCP](#5-primitivas-mcp)
6. [Modelo de consentimiento](#6-modelo-de-consentimiento)
7. [Detección de señales](#7-detección-de-señales)
8. [Estrategia de caché](#8-estrategia-de-caché)
9. [Registro en clientes](#9-registro-en-clientes)
10. [Entrega al agente: `CLAUDE.md` y `rsk-guide`](#10-entrega-al-agente-claudemd-y-rsk-guide)
11. [Plan de PRs](#11-plan-de-prs)
12. [Estrategia de tests](#12-estrategia-de-tests)
13. [Riesgos y decisiones resueltas](#13-riesgos-y-decisiones-resueltas)
14. [Definition of Done](#14-definition-of-done)

---

## 1. Contexto y problema

Hoy el agente que trabaja en un proyecto tiene tres fricciones:

1. **Sintaxis adivinada.** Para instalar una skill debe recordar la forma exacta (`rsk install <name> --pin`, `--global`, `--for claude-code`, `rsk pin --remove`) leyéndola del README. Es frágil y consume contexto en cada intento fallido.
2. **Cero descubrimiento.** No hay forma de que el agente sepa qué skills existen sin ejecutar `rsk catalog` y leer 17KB de descripciones.
3. **Sin señal del proyecto.** Nada le dice al agente "este repo es Go + gRPC + Postgres, y estas son las skills que aplican".

`rsk-guide` compensa parte de esto con texto, pero es una skill más que el agente tiene que recordar cargar, y su trigger actual (*"cuando el usuario menciona rsk"*) no cubre el caso de uso real: **empezar trabajo en un proyecto y necesitar saber qué estándares aplican**.

Este plan añade un servidor MCP con cuatro primitivas que cierran los tres huecos, reutilizando el código ya existente en `internal/source`, `internal/skill` y `internal/manifest`.

---

## 2. Decisiones de diseño

Decisiones tomadas en la fase de diseño, con su justificación. Todas cerradas; las que quedaron abiertas se resolvieron en §13.

### 2.1 Subcomando `rsk mcp`, no binario nuevo

El servidor es un subcomando de `rsk` (`rsk mcp`) en lugar de un binario separado.

- Un solo artefacto, una sola línea de versión, una sola entrada en goreleaser.
- Cero fricción de instalación: quien ya tiene `rsk` lo tiene.
- Descubrible vía `rsk --help`.

### 2.2 Transporte stdio, no HTTP

El cliente lanza el proceso; el proceso muere con el cliente.

| Criterio | stdio | HTTP local |
|---|---|---|
| Auth | Ninguna; el cliente lanza y confía | Validación de `Origin` (DNS rebinding), binding loopback, tokens |
| Ciclo de vida | El cliente lo arranca y lo mata | Daemon a gestionar, puertos que chocan, procesos zombis |
| Aislamiento de estado | Proceso fresco por sesión | Estado persistente, riesgo de manifiesto cacheado |
| Overhead de arranque | ~5–10 ms | Similar o peor (daemon ya caliente, pero hay que arrancarlo antes) |

Es además la regla explícita de `mcp-architect` §7: *"Use stdio when the server is bundled with the client"*, y §9: *"Local stdio servers: no auth — the client spawns the process and trusts it. Don't bolt OAuth onto stdio."*

**Concesión registrada:** el core stateless del spec 2026-07-28 (sin handshake `initialize`, sin `Mcp-Session-Id`) aligeró HTTP respecto a 2025, donde la negociación de sesión era lo peor de Streamable HTTP. Sigue sin compensar la categoría de trabajo extra (auth + lifecycle).

### 2.3 Transporte desacoplado de la lógica

El registro de tools **no** depende del transporte:

```
internal/mcp/
  registry.go     map[string]Tool{Name, Schema, Handler, Annotations}
  server.go       ServeStdio(ctx, reg)   ← adaptador
  tools/          project_profile.go, search_skills.go, install_skills.go
  catalog.go      recurso rsk://catalog
```

Añadir `rsk serve` (Streamable HTTP) después es un adaptador de ~30 líneas. Añadirlo desde un diseño acoplado a stdio es reescribir los handlers.

### 2.4 La recomendación se divide entre Go y el agente

Reparto explícito, porque esta es la decisión de diseño más importante del plan:

| Capa | Responsabilidad | Por qué |
|---|---|---|
| **Go (determinista)** | Detectar señales en el filesystem; mapear señal → skills candidatas | Testeable, revisable en un PR, cero falsos positivos silenciosos |
| **Agente (juicio)** | Decidir cuáles de las candidatas aplican realmente, y proponer | Un modelo razona "monolito Go con gRPC y Postgres" mejor que cualquier heurística |

**Lo que NO hacemos:** un ranking por scoring en Go. Sería un motor de reglas que hay que mantener durante meses y que fallaría en silencio. Cuando falla un lookup table, se ve en review; cuando falla un scorer, se descubre meses después porque el agente simplemente deja de seguir recomendando.

La tabla de §7 es un **lookup**, no un ranking: no puntúa, no ordena por score, solo propone candidatas.

### 2.5 Sin modelo embebido

Ninguna llamada de red a un LLM desde el server. El server es determinista y testeable; el juicio lo pone el cliente.

### 2.6 Abstracción de catálogo sobre ambas fuentes (no solo registry)

Corrección a la premisa original de §4. El índice rico (`Description`, `Latest`, `Personal`, `Versions`) **solo existe en modo registry**: vive en `source.Registry.Index()` y en `source.IndexEntry` (`internal/source/registry.go:35-96`). La interfaz común `source.Resolver` expone únicamente `All`/`Find` (`internal/source/resolver.go:12-19`), y en modo local-clone `newLocalSource` devuelve `source.NewLocal` (`cmd/rsk/resolve.go:24-29`), que no tiene `Index()`. Además `skill.Skill` no lleva `Description` (`internal/skill/skill.go:5-11`); `skill.Walk` solo parsea `version` (`internal/skill/registry.go:41-59`).

Sin resolver esto, `search_skills` (que busca en `description`) y `rsk://catalog` (que necesita `description`+`latest`+`personal`) **solo funcionarían en modo registry**, que es exactamente el modo que no usa quien desarrolla el catálogo.

**Decisión:** introducir `internal/catalog`, un adaptador que produce `[]CatalogEntry{Name, Description, Latest, Personal}` a partir de cualquier `source.Resolver`:

- **Modo registry:** usa `Registry.Index()` (con la caché en disco de §8).
- **Modo local / official:** usa `All()`; `Latest` es la versión en disco y `Description` se lee del frontmatter.
- Para que lo anterior sea posible, `skill.Skill` gana un campo `Description` que `parseSkill` rellena desde el frontmatter (junto a `version`).

**Consecuencia:** el MCP no depende del modo de la config; el precio es que PR 2 toca `internal/skill` y añade `internal/catalog` (ver §11). No es reutilización pura, es una capa fina nueva sobre código existente.

---

## 3. Alcance y no-alcance

### En alcance

- Servidor stdio con 4 primitivas (3 tools + 1 resource).
- Funcionamiento en modo local-clone y registry por igual, vía `internal/catalog` (§2.6).
- `project_profile`: detección de señales del proyecto + estado de instalación.
- `search_skills`: búsqueda sobre el catálogo con estado de instalación.
- `install_skills`: instalación vía manifest existente.
- Recurso `rsk://catalog`.
- `rsk new`: registro del server en los clientes configurados.
- Workflow pointer en `./CLAUDE.md`.
- Extensión de `rsk-guide` con la sección del MCP.

### Fuera de alcance (explícitamente)

- **Streamable HTTP / `rsk serve`.** Se deja la puerta abierta (§2.3) pero no se implementa.
- **Búsqueda semántica / embeddings.** Si algún día hace falta, es un **caché en disco** (§8), no un daemon.
- **Ranking o scoring de skills.** Ver §2.4.
- **`rsk recommend` como CLI.** Se descarta deliberadamente: es el mismo acierto en otra forma, y debe consumir `project_profile` en vez de duplicarlo. Si aparece, será después y encima de este.
- **Registro en Cursor.** Se difiere; ver §13.
- **Ejecución automática de installs.** El server nunca instala sin que el usuario lo apruebe (§6).

---

## 4. Arquitectura

```
┌─────────────────────────────────────────────────────────────┐
│ Cliente LLM (Claude Code / opencode)                        │
│  · tool definitions siempre en contexto                     │
│  · pide aprobación en install_skills (destructiveHint: true)│
└────────────────────────┬────────────────────────────────────┘
                         │ stdio (JSON-RPC 2.0)
┌────────────────────────▼────────────────────────────────────┐
│ rsk mcp  (cmd/rsk/mcp.go)                                    │
│  └─ adaptador stdio → internal/mcp.Server                    │
│                                                              │
│  internal/mcp/                                               │
│   ├─ registry   registro de tools, independiente de transporte│
│   ├─ profile    detección de señales (determinista)          │
│   ├─ search     match sobre internal/catalog                 │
│   └─ catalog    recurso rsk://catalog                        │
│                                                              │
│  Capas nuevas / reutilizadas:                                │
│   internal/catalog/             unifica local y registry     │
│   internal/install/             core de instalación (nuevo)  │
│   internal/source/registry.go   Index() All() Find()         │
│   internal/skill/               linker, manifest, Description│
│   internal/config/              config, registry cache path  │
└────────────────────────┬────────────────────────────────────┘
                         │
        ┌────────────────┴─────────────────┐
        │                                  │
  <RegistryCache>/index.json         registry HTTP
  (~/.ralvaskills/cache/registry/)   skills.ralvarez.dev
```

### Dependencia nueva

`github.com/modelcontextprotocol/go-sdk` **v1.8.0** (SDK oficial, Google). No está en `go.mod` todavía — PR 3 lo añade. **Corrección:** el plan original pedía `≥2.0.0`, pero el módulo **no tiene ninguna v2** (máximo publicado: `v1.8.0`) y esa versión **ya implementa el spec 2026-07-28** (`mcp/shared.go`: `latestProtocolVersion = protocolVersion20260728`). El floor `≥2.0.0` de `mcp-architect/STACK.md` es incorrecto para Go; conviene corregirlo allí. No usar `mark3labs/mcp-go`.

### mcpkit (helpers reutilizados)

`github.com/ralvarezdev/mcpkit` (local, publicado `v0.5.0`) es un módulo v0.x de helpers sobre el mismo go-sdk. PR 3 lo añade y usa dos piezas:

| Helper | Uso en el MCP |
|---|---|
| `RecoverWith(cfg)` / `RecoverMiddleware` | Envuelve los handlers: un panic en `install_skills` se convierte en error JSON-RPC en vez de tumbar el proceso stdio y la sesión. |
| `ErrorMsg.ToolError(logger, err)` | Puentea `ErrorMsg` con la firma del SDK: el handler devuelve el error saneado (el SDK lo marca `isError: true`), y la causa se queda en el log del server. El agente nunca ve internals. |

No se usan `BackendVerifier`, `CallerToken` ni `Callers[V]`: stdio no tiene auth ni multiusuario.

API confirmada contra v1.8.0: `mcp.NewServer(&mcp.Implementation{...}, nil)`, `server.AddReceivingMiddleware(...)`, `mcp.AddTool[In, Out](server, tool, handler)`, `server.Run(ctx, &mcp.StdioTransport{})`.

### Reutilización clave

`internal/source/registry.go` expone parte de lo necesario, pero **no todo y no en todos los modos**:

| Método | Modo | Uso en el MCP |
|---|---|---|
| `Index(ctx)` | registry | Mapa del índice para `search_skills` y el recurso de catálogo. **Hoy hace HTTP en cada llamada** — ver §8. |
| `All(ctx)` | local / official | Catálogo completo para el recurso en modo filesystem. |
| `Find(ctx, name)` | todos | Resolver nombre → skill. |
| `ensureCached(...)` | registry | Es **no exportado** (`registry.go:167`): desde `internal/mcp` solo se llega vía `Find`/`FindVersion`, nunca se llama directo. |

`skill.Skill` no lleva `Description`; se añade en PR 2 (§2.6) para que `search_skills` funcione también en modo local.

**La lógica de instalación no se reimplementa, pero tampoco es invocable tal cual.** La mutación vive dentro de comandos cobra: `runInstallGlobal`/`runInstallProject` reciben `*cobra.Command`, escriben en `cmd.OutOrStdout()` y llaman `termkit.Confirm(...)` (`cmd/rsk/install.go:189,201,214,326`). En stdio **stdout es el canal JSON-RPC y stdin el de peticiones**, así que ese camino no es usable desde el server. PR 3 **extrae el core de instalación** a un paquete interno que devuelve resultados estructurados y no pregunta; el comando `rsk install` pasa a ser un cliente de ese core, de modo que CLI y MCP comparten una única implementación y no pueden divergir.

---

## 5. Primitivas MCP

Criterio de selección de primitiva según `mcp-architect` §2 (*"If the LLM should autonomously call it → tool; if it's data it should read → resource"*).

| Primitive | Nombre | Tipo | Efecto |
|---|---|---|---|
| Resource | `rsk://catalog` | Read | Ninguno |
| Tool | `project_profile` | Read | Ninguno |
| Tool | `search_skills` | Read | Ninguno |
| Tool | `install_skills` | **Write** | Manifest + symlinks + ficheros |

### 5.1 `rsk://catalog` (resource)

Índice completo, **descripciones recortadas**. No un volcado completo del índice: las tool definitions viven siempre en contexto, y el contenido debe entrar bajo demanda.

```
name:        "rsk-catalog"
uri:         "rsk://catalog"
mimeType:    "application/json"
description: "Catálogo de skills disponibles en ralvaskills. Descripciones recortadas;
              usa search_skills para filtrar por proyecto."
```

Payload: `{name, description (recortada), latest, personal: bool}` por skill, servido por `internal/catalog` (§2.6). Se **excluye** `versions` completo — lo que se necesita es `latest`. En modo local `latest` es la versión en disco.

### 5.2 `project_profile` (tool)

El corazón del sistema. Determinista.

**In:**
```go
type ProjectProfileIn struct {
    Path string `json:"path,omitempty" jsonschema:"Ruta del proyecto; por defecto el cwd"`
}
```

**Out:**
```go
type ProjectProfileOut struct {
    ProjectRoot string         `json:"project_root"`
    Signals     []Signal       `json:"signals"`
    Candidates  []Candidate    `json:"candidates"`
    Installed   []InstalledRef `json:"installed"`
    Manifest    *ManifestState `json:"manifest,omitempty"`
}

type Signal struct {
    Kind  string `json:"kind"`   // "language:go", "infra:docker", "ci:github-actions"
    Proof string `json:"proof"`  // fichero o línea que lo evidencia — nunca evidencia a ciegas
}

type Candidate struct {
    Skill     string   `json:"skill"`
    Because   []string `json:"because"`  // signals que lo dispararon
    Latest    string   `json:"latest"`
    Installed bool     `json:"installed"`
}

type InstalledRef struct {
    Name     string `json:"name"`
    Version  string `json:"version"`
    Pinned   bool   `json:"pinned"`
    Personal bool   `json:"personal"`
}

type ManifestState struct {
    Exists bool `json:"exists"`  // false si no hay .rsk/rsk.mod
    Skills int  `json:"skills"`  // entradas en rsk.mod
    Pinned int  `json:"pinned"`  // entradas en la lista pinned
}
```

Dos propiedades no negociables:

- **`Proof` es obligatorio.** Cada señal cita el fichero que la evidencia. Una skill candidata sin señal que la justifique es ruido.
- **`Installed` viaja en cada candidata.** El agente no necesita una tool separada para saber qué ya está.

> **`Candidate.Skill` admite skill o bundle.** Nombres como `go-grpc` son bundles (`rsk-guide` SKILL.md:153), no skills. Es válido proponerlos —`install_skills` resuelve ambos vía `resolveNames`— pero el contrato debe decirlo, o `Skill` significa dos cosas distintas según de dónde venga.

### 5.3 `search_skills` (tool)

Búsqueda por keywords sobre `name` + `description` del catálogo. Match simple (tokenización + scoring por campo, peso en `name`), sin embeddings.

**In:**
```go
type SearchSkillsIn struct {
    Query string `json:"query"`
    Limit int    `json:"limit,omitempty" jsonschema:"Por defecto 10, máximo 25"`
}
```

**Out:** `[]Candidate` — la misma forma que devuelve `project_profile`, para que el agente use un único modelo mental y no dos. `Because` va vacío en búsquedas (no hay señales que citar); es la única diferencia semántica.

### 5.4 `install_skills` (tool)

Única tool con efecto. Delgada por diseño.

**In:**
```go
type InstallSkillsIn struct {
    Names []string `json:"names"`
    Scope string   `json:"scope,omitempty" jsonschema:"project (por defecto) | global"`
    Pin   bool     `json:"pin,omitempty"`
    For   string   `json:"for,omitempty" jsonschema:"claude-code | opencode"`
}
```

**Out:**
```go
type InstallSkillsOut struct {
    Installed []InstallResult `json:"installed"`
    Failed    []InstallResult `json:"failed"`
}
```

Cada resultado lleva `name`, `version`, `scope`, `files_changed`, y `error` (string legible) si falló. `scope: project` (default) exige un `.rsk/` existente — sin él, `resolveTargetDirs` falla (`cmd/rsk/resolve.go:55-63`); el resultado debe explicarlo. `for` solo tiene efecto con `scope: global`; con `project` se ignora (los targets salen de los tools del manifest).

**Reglas de error** (`mcp-architect` §10): los fallos se devuelven como resultado de tool con `isError: true` y un bloque `content[]` describiendo qué pasó — **nunca** como error JSON-RPC. El modelo nunca ve los errores de transporte como contenido; necesita leerlos para poder reintentar con otros argumentos o preguntar al usuario. El texto lo produce `mcpkit.ErrorMsg.ToolError`: el SDK convierte el error devuelto en `isError: true` usando `err.Error()`, así que el handler devuelve el mensaje público del `ErrorMsg` y la causa solo va al log.

---

## 6. Modelo de consentimiento

El punto de diseño que más cuidado requiere. Hay **dos actos distintos con dos capas de consentimiento distintas**, y confundirlos es el error fácil.

| Acto | Cuándo | Quién pide | Mecanismo |
|---|---|---|---|
| **Registrar** el server en config de cliente | Una vez, en `rsk new` | Humano, en el prompt de init | `huh` interactivo |
| **Instalar** una skill | Cada vez, mid-sesión | Cliente LLM | Tool annotation `destructiveHint: true` |

### 6.1 Por qué las annotations son el mecanismo real

Un servidor MCP **no tiene canal para preguntar al usuario**. No puede abrir un prompt. Por tanto el consentimiento de instalación no puede implementarse en el server, y no debe intentarse.

Vive en las **tool annotations** (`mcp-architect` §4):

```go
project_profile  → readOnlyHint: true,  destructiveHint: false
search_skills    → readOnlyHint: true,  destructiveHint: false
install_skills   → readOnlyHint: false, destructiveHint: true, idempotentHint: false
```

**Consecuencia (a validar en PR 3):** los clientes que honran `destructiveHint` muestran el prompt de aprobación sin que el modelo haya leído ninguna skill. `mcp-architect` §3/§4 es explícito en que las annotations son *hints* que manejan UX del cliente y **no imponen política en el server**; el comportamiento concreto de Claude Code ante una tool MCP con `destructiveHint: true` es una **premisa de §6, no un hecho verificado**. Se comprueba en el criterio de aceptación de PR 3.

### 6.2 Defensa en profundidad

`rsk-guide` §"Always ask the user before taking action" establece la política en texto. La annotation la **refuerza en el cliente** (es un hint, no una imposición server-side).

Esta asimetría es intencional: la skill es *asesoramiento* y un modelo puede saltársela; la annotation es una restricción del cliente. Juntas cubren los dos casos:

- El modelo leyó `rsk-guide` → sabe que debe preguntar y por qué.
- El modelo no la leyó → el cliente (si honra la hint) le pide aprobación igual.

### 6.3 Lo que el server nunca hace

- Instalar sin que el cliente haya aprobado (`install_skills` es siempre explícito).
- Escribir fuera del scope solicitado.
- Auto-registrarse: el server no modifica su propia config; eso es `rsk new`.

---

## 7. Detección de señales

Lookup table señal → skills candidatas. **No** puntúa ni ordena; solo propone.

### 7.1 Señales

| Señal | Evidencia | Candidatas |
|---|---|---|
| `language:go` | `go.mod` | `go-architect` |
| `language:python` | `pyproject.toml`, `requirements.txt`, `uv.lock` | `python-architect` |
| `language:typescript` | `tsconfig.json`, `package.json` | — |
| `framework:nextjs` | `next.config.*`, dep `next` | `nextjs-architect`, `react-architect` |
| `framework:astro` | `astro.config.*` | `astro-architect` |
| `protocol:grpc` | `*.proto`, dep `google.golang.org/grpc` | `grpc-architect`, `protobuf-architect` |
| `protocol:rest` | `openapi.yaml`, `swagger.json`, `*.http` | `rest-api-architect` |
| `infra:docker` | `Dockerfile`, `docker-compose.yml` | `docker-architect` |
| `infra:k8s` | `Chart.yaml`, `kustomization.yaml` | — (añadir con skill k8s) |
| `infra:ci` | `.github/workflows/` | `ci-cd-architect` |
| `repo:tooling` | `.golangci.yml`, `mise.toml`, `Taskfile.yml` | `repo-tooling-architect` |
| `repo:cli` | `cmd/` con `main.go`, `main.go` en raíz | `cli-tool-architect` |
| `data:postgres` | `*.sql` + `sqlx`/`psycopg` | (cubierto por las de lenguaje) |
| `quality:refactor` | — | `logic-cleaner`, `improve-codebase-architecture` (nunca por señal; sólo bajo petición) |

Las candidatas pueden ser skills o bundles (`go-grpc`, `gin`); `install_skills` resuelve ambos.

### 7.2 Reglas

- **Toda señal cita su evidencia.** Sin `Proof`, no entra en la salida.
- **Sin señal, no hay candidata.** La última fila es explícita: las skills de refactoring nunca se proponen solas — se proponen cuando el agente ya está en esa tarea.
- **La tabla se versiona como código**, en `internal/mcp/signals.go`, con tests. Cambiar la tabla es un PR revisable, no un ajuste de prompt.
- **`personal: true` se excluye** de las candidatas por defecto (son privadas del usuario). Se marcan con el flag, pero no se proponen.
- **El escaneo no desciende a `testdata/`** (ni a `.git`, `.rsk`, `node_modules`, `vendor` o caches): una `.proto` o un `Dockerfile` de fixture no dicen nada del stack del proyecto. Detectado en el gate de PR 1 — el repo se auto-detectaba como `protocol:grpc` por sus propias fixtures.

---

## 8. Estrategia de caché

Compute y transporte son ortogonales. **stdio no implica recalcular.**

> **Corrección:** la caché de índice en disco **no existe hoy**. `source.Registry.Index()` hace HTTP GET en cada llamada (`internal/source/registry.go:73-96`), `ensureCached` cachea solo tarballs (`:167`) y `internal/update/check.go:134` también va a red. La ruta `~/.cache/rsk/index.json` del diseño original tampoco existe: la caché registry es hermana de `official_cache`, por defecto `~/.ralvaskills/cache/registry` (`internal/config/fs.go:73-84`). Persistir el índice es **trabajo nuevo de PR 2**, no reutilización.

Cada arranque del proceso hará:

1. leer `<cfg.RegistryCache()>/index.json` (por defecto `~/.ralvaskills/cache/registry/index.json`) → sin red si está fresco
2. parsear a estructuras en memoria → ~0.5 ms
3. servir tools desde memoria

**Por qué caché en disco y no daemon caliente:** sobrevive a reinicios del cliente, sobrevive a un crash del server, y el warm cache se comparte entre proyectos y entre sesiones. La calidez en memoria de un daemon se pierde en cada reinicio del cliente; el disco no.

- **TTL de refresco:** 24h. `install`/`update` refrescan de inmediato.
- **Fuente de verdad:** `skills.ralvarez.dev/index.json`.
- **Fallback:** si la caché no existe, se hace fetch (con `context` timeout) y se persiste. Si el fetch falla y hay caché vieja, se usa la vieja anotando su antigüedad; el server arranca **sin red** solo si hay caché.
- **Modo local:** no hay índice que cachear — `internal/catalog` (§2.6) lee del filesystem.
- **Escritura atómica:** reutilizar `fsx.WriteAtomic` (`internal/fsx/atomic.go`).
- **Se queda local por ahora:** la caché es stdlib puro, pero no hay un segundo consumidor en `mcpkit`; si otro server necesita lo mismo, se extrae entonces (regla de `mcpkit`: extender con el segundo consumidor).

Coste real de arranque: ~5–10 ms (spawn) + ~0.5 ms (parse de 48K). No justifica nada más.

---

## 9. Registro en clientes

### 9.1 Dónde

| Cliente | Fichero | Config |
|---|---|---|
| Claude Code | `.mcp.json` (raíz del proyecto) | `{"mcpServers": {"rsk": {"command": "rsk", "args": ["mcp"]}}}` |
| opencode | `opencode.json` (raíz del proyecto) | `{"mcp": {"rsk": {"type": "local", "command": ["rsk", "mcp"], "enabled": true}}}` |

**Proyecto, no usuario.** El punto de `rsk.mod` es que el set de skills se comparte con el equipo; registrar el MCP en el proyecto significa que un `git pull` lo activa para todos.

**Fricción conocida:** ambos exigen `rsk` en el PATH de cada compañero. `rsk new` **verifica que `rsk` resuelve en PATH** antes de registrar y, si no, avisa y ofrece no registrar; el README lo documenta. Un teammate sin `rsk` verá el server fallar al conectar — es el mismo contrato que cualquier CLI-based MCP server (npx, uvx). Nótese que config de usuario **no** elimina el requisito de PATH: el comando debe ser ejecutable en cada máquina igual, así que no es una alternativa que resuelva la fricción.

### 9.2 Cómo

`internal/tool/` ya escribe config de Claude y opencode (`claude.go`, `opencode.go`), pero **no** `.mcp.json` ni la clave `mcp` de `opencode.json`: eso es escritura nueva. Se **extiende ese paquete**, no se crea un camino nuevo:

```
internal/tool/tool.go     interfaz existente + capacidad nueva
internal/tool/claude.go   escritura en .mcp.json
internal/tool/opencode.go escritura en el bloque mcp de opencode.json
```

Regla: ambas escrituras deben ser **idempotentes y no destructivas** — preservar claves ajenas y no pisar entradas existentes de otro server.

### 9.3 En `rsk new`

```
rsk new
  ...preguntas existentes (source, tools)...
  ? ¿Registrar el servidor MCP de rsk en este proyecto?  [Sí / No]
       Sí → verifica que rsk está en PATH (§9.1)
           → escribe .mcp.json / opencode.json
           → escribe el workflow pointer en ./CLAUDE.md (§10.2)
       No → no se escribe nada de MCP ni el pointer
  ? ¿Instalar rsk-guide en este proyecto?  [Sí / No]   ← pregunta separada
```

**Dos actos, dos preguntas.** Registrar el MCP e instalar una skill son cambios de estado distintos; la regla #1 de `rsk-guide` exige confirmación explícita por cada instalación, y `rsk new` hoy no instala nada. Acoplarlas en un solo sí/no haría implícita la instalación. El gate del pointer es **el mismo "sí/no" del MCP**: si el usuario declina el MCP, no se escribe el pointer, porque apuntaría a un tool que no existe. La instalación de `rsk-guide` es independiente y opcional (la annotation de §6 protege aunque no se instale).

---

## 10. Entrega al agente: `CLAUDE.md` y `rsk-guide`

### 10.1 Por qué no una skill nueva para el MCP

Una skill tipo "cómo usar el MCP de rsk" sería mayoritariamente redundante:

- **Las tool descriptions ya están siempre en contexto** cuando el server está conectado. El agente ya sabe qué hace `project_profile`. Una skill que le vuelva a explicar los parámetros son tokens muertos.
- **Claude Code autodescubre las skills instaladas.** Aparecen en la lista de skills disponibles con su description. Si `project_profile` recomienda `go-architect` y está instalada, **ya aparece sola** — no hace falta ninguna skill que lo diga.

Lo que falta en los schemas es **la política**, y ésa ya está escrita en `rsk-guide`.

**Decisión: no se crea una skill nueva. Se extiende `rsk-guide`.** Dos meta-skills sobre rsk competirían por el mismo espacio de triggers y confundirían al modelo.

### 10.2 Workflow pointer en `./CLAUDE.md` (no en `.rsk/CLAUDE.md`)

**Corrección:** `.rsk/CLAUDE.md` está reservado. `writePinnedClaude` lo reescribe con **solo** las líneas `@../.claude/skills/<name>/SKILL.md` en cada mutación (`internal/tool/claude.go:95-108`), y el contrato lo fija: *"one line per entry, no other content"* (`docs/SPECS.md:642`). Cualquier pointer añadido ahí lo borra el siguiente `rsk install`/`rsk pin`.

El canal correcto es `./CLAUDE.md`, que `rsk new` ya crea/edita de forma idempotente y **preserva el resto del contenido** (`internal/tool/claude.go:110-134`). Tres líneas bastan:

```markdown
## Project standards

Before scaffolding or a non-trivial change, call the rsk MCP `project_profile`
tool to learn which skills apply here. Installed skills appear automatically in
your skill list — invoke them with the Skill tool. Never install without asking.
```

Tres líneas, siempre en contexto, sin trigger que afinar, sin skill que cargar. `rsk destroy` debe eliminar también el bloque del pointer (hoy `removeClaudeImport` solo quita la línea `@.rsk/CLAUDE.md`).

### 10.3 Extensión de `rsk-guide`

Dos cambios:

1. **Sección nueva** (§ MCP): las cuatro primitivas, el flujo recomendado
   (`project_profile` → `search_skills` → proponer → `install_skills`), y la nota de que la aprobación la da la annotation en el cliente, no la skill.
2. **Cirugía de description.** El trigger actual es *"when the user mentions rsk"* — equivocado para el caso de uso real. Necesita una segunda clause que dispare con *"starting work in a new project / scaffolding / what standards apply here"*, sin dejar de disparar con *"user mentions rsk"*.

### 10.4 Reparto final

| Capa | Contenido |
|---|---|
| `./CLAUDE.md` | Las 3 líneas del workflow pointer |
| `rsk-guide` | Detalle del MCP + trigger corregido |
| `install_skills` annotation | `destructiveHint: true` — el cliente, si honra la hint, fuerza el prompt |
| **Skill nueva** | **Nada** |

---

## 11. Plan de PRs

**Cinco PRs** (el plan original tenía cuatro; PR 2 se parte porque las correcciones de §2.6/§4/§8 añaden catálogo, caché de índice y extracción del core de instalación — demasiado para un PR revisable). El orden importa: cada uno deja el repo en estado verde y usable, y el primero valida la pieza de mayor incertidumbre (¿la predicción sirve?) sin haber invertido en transporte ni registro.

> **Estado:** PR 1 ✅ completado y en `main`; PR 2–5 pendientes.

### PR 1 — `project_profile` + tabla de señales ✅

Sin servidor MCP. Es lógica pura, testeable en aislamiento, y es la parte que puede estar equivocada.

| Fichero | Cambio |
|---|---|
| `internal/mcp/profile.go` | Detección de señales (`Scan` → `[]Signal`) |
| `internal/mcp/signals.go` | Lookup table señal → skills |
| `internal/mcp/profile_test.go` | Tests contra fixtures (§12) |
| `internal/mcp/mcp.go` | Doc del paquete, por qué existe |

- Sin dependencia nueva.
- Sin cambios de CLI.
- **Criterio de aceptación:** `go test ./internal/mcp/...` verde; los fixtures de §12 producen las candidatas esperadas; la salida cita `Proof` en el 100% de las señales.

**Gate (ejecutado en PR 1):** perfilado sobre 8 proyectos reales; las candidatas fueron sensatas salvo un falso positivo de `testdata/` (el repo se auto-detectaba como `protocol:grpc` por sus propias fixtures), corregido en §7.2.

### PR 2 — Catálogo unificado + caché de índice

Sin transporte. Es la capa que §2.6 y §8 añaden, testeable en aislamiento.

| Fichero | Cambio |
|---|---|
| `internal/skill/skill.go` | Campo `Description` en `Skill` |
| `internal/skill/registry.go` | `parseSkill` parsea `description` del frontmatter |
| `internal/catalog/catalog.go` | `CatalogEntry{Name, Description, Latest, Personal}` desde cualquier `source.Resolver` |
| `internal/catalog/catalog_test.go` | Modo local y registry (con índice stubbeado) |
| `internal/mcp/cache.go` | Caché de índice en disco (`<RegistryCache>/index.json`) + TTL 24h + escritura atómica vía `fsx.WriteAtomic` |

- Sin dependencia nueva. Sin cambios de CLI.
- **Criterio de aceptación:** el catálogo produce entradas equivalentes en modo local y registry para un mismo skill; la caché arranca sin red si el índice está fresco; el TTL caduca y re-fetcha; la exclusión de `personal: true` se testea aquí con índice stub.

### PR 3 — Servidor stdio + 4 primitivas + core de instalación

| Fichero | Cambio |
|---|---|
| `cmd/rsk/mcp.go` | Subcomando `rsk mcp` |
| `cmd/rsk/setup.go` | `rootCmd.AddCommand(mcpCmd)` (los subcomandos se registran aquí, no en `root.go`) |
| `internal/mcp/registry.go` | Registro de tools, desacoplado de transporte |
| `internal/mcp/server.go` | `ServeStdio`; `mcpkit.RecoverWith` + logging |
| `internal/mcp/tools/*.go` | `project_profile`, `search_skills`, `install_skills` |
| `internal/mcp/catalog.go` | Recurso `rsk://catalog` sobre `internal/catalog` |
| `internal/install/` (nuevo) | Core de instalación extraído de `cmd/rsk/install.go`: sin cobra, sin prompt, resultados estructurados |
| `cmd/rsk/install.go` | Pasa a ser cliente del core extraído |
| `go.mod` | `github.com/modelcontextprotocol/go-sdk` v1.8.0 + `github.com/ralvarezdev/mcpkit` v0.5.0 |

- `internal/skill`, `internal/manifest` y `internal/source` solo se consumen (los cambios de `internal/skill` fueron en PR 2).
- **Criterio de aceptación:** arranca con Claude Code vía `claude mcp add`; las cuatro primitivas responden; `install_skills` produce el mismo efecto que `rsk install` **sin tocar stdout/stdin del protocolo**; se verifica que Claude Code pide aprobación ante `destructiveHint: true` (§6.1); un install falla con `isError: true`, no con error JSON-RPC.

### PR 4 — Registro en clientes vía `rsk new`

| Fichero | Cambio |
|---|---|
| `internal/tool/claude.go` | `.mcp.json` (nuevo) + pointer en `./CLAUDE.md` + limpieza del pointer en destroy |
| `internal/tool/opencode.go` | Clave `mcp` en `opencode.json` (nueva) |
| `internal/tool/tool.go` | Método de interfaz para registro MCP |
| `internal/tool/*_test.go` | Idempotencia y no destructividad |
| `cmd/rsk/new.go` | Verificación de PATH + pregunta de MCP + pregunta separada de `rsk-guide` |

- **Criterio de aceptación:** `rsk new` avisa si `rsk` no está en PATH y no registra; registra sin perder claves ajenas; correrlo dos veces no duplica entradas; el pointer sobrevive a `rsk install`/`rsk pin`; `rsk destroy` limpia registro y pointer.

### PR 5 — `rsk-guide` + README

| Fichero | Cambio |
|---|---|
| `skills/meta/rsk-guide/SKILL.md` | Sección MCP + description con trigger corregido; bump de versión |
| `skills/meta/rsk-guide/CHANGELOG` (si existe) | Entrada |
| `README.md` | Sección del MCP, requisito de `rsk` en PATH |
| `docs/SPECS.md` | Sección CLI para `rsk mcp` (§CLI Design) |

- Al bumpear `rsk-guide`, CI publica la release y actualiza el índice automáticamente (`docs/REGISTRY_FLOW.md`).

---

## 12. Estrategia de tests

### 12.1 Señales — la parte que más importa

Fixtures de directorios mínimos bajo `internal/mcp/testdata/`:

```
fixtures/gomod-only/           go.mod
fixtures/go-grpc/              go.mod + api/v1/service.proto
fixtures/nextjs/               package.json (dep next) + next.config.ts
fixtures/fullstack/            go.mod + Dockerfile + .github/workflows/ci.yml
fixtures/nested-testdata/      go.mod + testdata/sample.proto (el escaneo ignora testdata/)
```

Tests:
- Detección correcta por fixture, con `Proof` apuntando al fichero real.
- Proyecto vacío → cero candidatas, **sin error** (se prueba con `t.TempDir()`; git no versiona directorios vacíos).
- El escaneo no desciende a `testdata/`: `fixtures/nested-testdata` no produce `protocol:grpc`.
- La exclusión de `personal: true` se testea en `internal/catalog` con un índice stub (es propiedad de catálogo, no señal de proyecto), **no** con un fixture de filesystem.
- Una señal nunca produce candidata sin `Proof`.

### 12.2 Búsqueda

- Match por nombre tiene prioridad sobre match en description.
- `Limit` se respeta; default 10; máximo 25.
- Query sin resultados → lista vacía, **no error**.
- Normalización: query en mayúsculas y con acentos encuentra lo mismo.

### 12.3 Instalación

- `install_skills` delega en `internal/install` (mismo resultado que `rsk install`).
- Alcance `project` vs `global` respetado.
- Nombre inexistente → resultado de tool con `isError: true` y mensaje legible (§5.4).
- Fallo parcial: 2 de 3 skills instala → las 2 en `installed`, la tercera en `failed` con error legible.

### 12.4 Transporte

Tests **in-process**, no HTTP real — el SDK Go trae transporte en memoria (`mcp-architect` §13). Un servidor HTTP real solo para checks de transporte, y no es el caso aquí.

- Las 4 primitivas responden por el pipe en memoria.
- Las 3 read llevan `readOnlyHint: true`; `install_skills` lleva `destructiveHint: true`.
- El `inputSchema` de cada tool no está vacío (el SDK lo infiere del struct `In`; un tag roto o un `In` mal tipado lo dejaría vacío sin fallar).

### 12.5 Registro en config

- Preserva claves ajenas en `.mcp.json` y `opencode.json`.
- Idempotente: dos ejecuciones, un resultado.
- No pisa entradas de otros servers MCP.
- `rsk destroy` limpia el registro y el workflow pointer.

---

## 13. Riesgos y decisiones resueltas

### Riesgos

| Riesgo | Impacto | Mitigación |
|---|---|---|
| La tabla de señales predice mal | Recommendations ruidosas | Gate en PR 1: validar contra ≥5 proyectos reales antes de invertir en el MCP |
| El meta-comportamiento satura el contexto | El agente ignora el pointer | 3 líneas, no una skill; validar con uso real |
| `.mcp.json` exige `rsk` en PATH | El teammate ve el server fallar | Verificación de PATH en `rsk new` + sección en README |
| `rsk-guide` dispara en el momento equivocado | El workflow nunca se activa | Cirugía de description (PR 5) |
| Modo local sin índice ni descripciones | `search_skills`/`rsk://catalog` no funcionan fuera de registry | `internal/catalog` sobre ambas fuentes (§2.6) |
| Las annotations no fuerzan aprobación en todos los clientes | Instalación sin consentimiento | Verificar el prompt de Claude Code en PR 3; `rsk-guide` como defensa en texto |
| El pointer de workflow se borra al mutar el manifest | El agente nunca ve el pointer | Escribirlo en `./CLAUDE.md`, no en `.rsk/CLAUDE.md` (§10.2) |
| Duplicación de lógica de instalación | CLI y MCP divergen | Core de instalación extraído a `internal/install`, compartido por CLI y MCP |
| El STACK de `mcp-architect` fija go-sdk ≥2.0.0, que no existe | Confusión de versiones; PR 3 bloqueado en falso | Pinear `v1.8.0` (ya habla 2026-07-28) y corregir `STACK.md` |
| `mcpkit` es v0.x y puede mover su API | Rompe el build al actualizar | Pinear tag (`v0.5.0`); ralvaskills es su segundo consumidor |

### Decisiones resueltas

1. **`.mcp.json` de proyecto vs. config de usuario → proyecto.** El valor diferencial es compartir el set con el equipo vía `rsk.mod`, y el requisito de PATH no desaparece con config de usuario (el comando debe ser ejecutable en cada máquina igual). Se mantiene proyecto; `rsk new` verifica PATH y avisa (§9.1).
2. **Cursor → fuera de este plan.** `.cursor/mcp.json` es el bloque de config más volátil y multiplica la validación del primer PR que toca config sin enseñar nada nuevo. La capacidad de registro MCP se añade como método de interfaz para que Cursor sea "un fichero nuevo + `Register`", no un refactor (`internal/tool/tool.go:44-47`).
3. **`rsk-guide` en `rsk new` → pregunta separada.** Registrar el MCP e instalar una skill son dos actos de estado; la regla #1 de `rsk-guide` exige confirmación por instalación, y `rsk new` hoy no instala nada. Se ofrece como pregunta propia, no dentro del mismo sí/no (§9.3).

---

## 14. Definition of Done

- [ ] `rsk mcp` arranca como subcomando y responde por stdio.
- [ ] Las 4 primitivas funcionan: `rsk://catalog`, `project_profile`, `search_skills`, `install_skills`.
- [ ] El catálogo funciona en modo local-clone y registry por igual (§2.6).
- [ ] `install_skills` está anotada `destructiveHint: true`; se verificó que el cliente pide aprobación sin depender de que el modelo lea ninguna skill.
- [ ] La instalación comparte un core único con `rsk install` (`internal/install`), sin duplicar lógica ni escribir en stdout del protocolo.
- [ ] Los handlers usan `mcpkit.RecoverWith` (un panic no tumba la sesión) y `mcpkit.ErrorMsg.ToolError` (el agente no ve causas internas).
- [ ] Los fallos llegan como resultado de tool con `isError: true`, nunca como error JSON-RPC.
- [ ] El arranque no hace red si hay caché; caché de índice en disco con TTL de 24h.
- [ ] El registro de tools no depende del transporte (un `rsk serve` futuro es un adaptador).
- [ ] `rsk new` verifica `rsk` en PATH, registra el server en Claude Code y opencode, idempotente y no destructivo.
- [ ] El workflow pointer va en `./CLAUDE.md` y sobrevive a mutaciones del manifest.
- [ ] `rsk-guide` tiene sección del MCP y su trigger cubre "empezar trabajo en un proyecto".
- [ ] No se ha creado ninguna skill nueva.
- [ ] `go test ./...` verde; linters (`golangci-lint`) sin findings nuevos.
- [ ] Ningún install se ejecuta sin aprobación del usuario.

---

## Referencias internas

- [`skills/protocols/mcp-architect/`](../skills/protocols/mcp-architect/SKILL.md) — primitivas, tool design, annotations, transporte, auth, errores, testing
- [`docs/REGISTRY_FLOW.md`](REGISTRY_FLOW.md) — cómo se publica el índice que consume el server
- [`skills/meta/rsk-guide/`](../skills/meta/rsk-guide/SKILL.md) — política de consentimiento y workflows del CLI
- `internal/source/registry.go` — `Index()`, `All()`, `Find()`, `ensureCached()`
- `internal/tool/` — escritura de config de Claude Code y opencode
- `github.com/ralvarezdev/mcpkit` — `RecoverWith`, `ErrorMsg.ToolError`; helpers genéricos sobre el go-sdk
