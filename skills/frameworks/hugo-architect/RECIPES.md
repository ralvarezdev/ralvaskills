# Hugo Recipes

Reference implementations referenced from [SKILL.md](SKILL.md). Loaded on demand.

## Project tree

```text
my-site/
├── hugo.toml                 # site config (TOML)
├── go.mod                    # Hugo Modules manifest
├── archetypes/
│   └── posts.md              # front-matter template for `hugo new posts/foo.md`
├── content/
│   ├── _index.md             # home page content
│   └── posts/
│       └── launch-day/
│           ├── index.md      # leaf Page Bundle
│           └── cover.jpg
├── layouts/
│   ├── _default/
│   │   ├── baseof.html
│   │   ├── single.html
│   │   └── list.html
│   ├── partials/
│   │   ├── head.html
│   │   └── footer.html
│   └── shortcodes/
│       └── figure.html
├── assets/
│   ├── css/app.scss
│   └── js/app.ts
├── static/
│   ├── favicon.ico
│   └── robots.txt
├── data/
│   └── authors.yaml
├── i18n/
│   ├── en.toml
│   └── es.toml
└── public/                   # build output (gitignored)
```

## Template hierarchy

`baseof.html` — the single shell:

```html
<!doctype html>
<html lang="{{ site.Language.Lang }}">
  <head>{{ partial "head.html" . }}</head>
  <body>
    {{ block "main" . }}{{ end }}
    {{ partial "footer.html" . }}
  </body>
</html>
```

`single.html` extends it:

```html
{{ define "main" }}
  <article>
    <h1>{{ .Title }}</h1>
    {{ .Content }}
  </article>
{{ end }}
```

## Hugo Module setup

```toml
# hugo.toml
[module]
  [[module.imports]]
    path = "github.com/example/hugo-theme-foo"
  [[module.imports.mounts]]
    source = "layouts/partials/footer.html"
    target = "layouts/partials/footer.html"   # override theme partial
```

```bash
hugo mod init github.com/me/my-site
hugo mod get github.com/example/hugo-theme-foo@v1.4.0
hugo mod vendor   # commit _vendor/ for reproducible builds
```

## Asset pipeline chain

```html
{{ $css := resources.Get "css/app.scss"
  | toCSS (dict "outputStyle" "compressed" "enableSourceMap" false)
  | postCSS
  | minify
  | fingerprint }}
<link rel="stylesheet" href="{{ $css.Permalink }}"
      integrity="{{ $css.Data.Integrity }}" crossorigin="anonymous">
```

## Render hook examples

`layouts/_default/_markup/render-image.html` — lazy-load every Markdown image:

```html
<img src="{{ .Destination | safeURL }}" alt="{{ .Text }}"
     loading="lazy" decoding="async">
```

`layouts/_default/_markup/render-link.html` — open external links in a new tab:

```html
{{- $u := urls.Parse .Destination -}}
<a href="{{ .Destination | safeURL }}"
   {{- if $u.IsAbs }} target="_blank" rel="noopener noreferrer"{{ end }}>
  {{ .Text | safeHTML }}
</a>
```

## GitHub Pages deploy workflow

```yaml
# .github/workflows/deploy.yml
name: Deploy Hugo
on:
  push:
    branches: [main]
permissions:
  contents: read
  pages: write
  id-token: write
concurrency:
  group: pages
  cancel-in-progress: true
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { submodules: recursive, fetch-depth: 0 }
      - uses: peaceiris/actions-hugo@v3
        with: { hugo-version: '0.161.1', extended: true }
      - run: hugo --minify --baseURL "${{ steps.pages.outputs.base_url }}/"
      - uses: actions/upload-pages-artifact@v3
        with: { path: ./public }
  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment: { name: github-pages, url: ${{ steps.deployment.outputs.page_url }} }
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

## Cache layers reference

(Stub — fill in once first review identifies which Hugo cache layers cause real-world friction. Candidates: build cache, image cache, module cache, partialCached.)
