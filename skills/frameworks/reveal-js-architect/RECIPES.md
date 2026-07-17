# Reveal.js Architect — Recipes

Reference implementations. Load on demand when the user actually needs a skeleton.

## Vanilla project skeleton

```
deck/
├── index.html          # shell: reveal.js CSS/JS links, plugin scripts, <div class="reveal">
├── slides.md           # data-markdown source (if Markdown-mode)
├── css/
│   └── custom.css       # theme override, loaded after the built-in theme
├── js/
│   └── main.js           # Reveal.initialize({...}), plugin registration
├── assets/
│   └── images/, media/
└── package.json          # if using npm + Vite instead of CDN links
```

`index.html` shell (CDN form):

```html
<!doctype html>
<html>
<head>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/reveal.js@6.0.1/dist/reveal.css">
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/reveal.js@6.0.1/dist/theme/black.css">
  <link rel="stylesheet" href="css/custom.css">
</head>
<body>
  <div class="reveal"><div class="slides">
    <section data-markdown="slides.md" data-separator="^\n---\n$" data-separator-vertical="^\n--\n$"></section>
  </div></div>
  <script src="https://cdn.jsdelivr.net/npm/reveal.js@6.0.1/dist/reveal.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/reveal.js@6.0.1/plugin/markdown/markdown.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/reveal.js@6.0.1/plugin/notes/notes.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/reveal.js@6.0.1/plugin/highlight/highlight.js"></script>
  <script src="js/main.js"></script>
</body>
</html>
```

`js/main.js`:

```js
Reveal.initialize({
  hash: true,
  controls: true,
  progress: true,
  center: true,
  transition: 'slide',
  plugins: [ RevealMarkdown, RevealHighlight, RevealNotes ],
});
```

npm + Vite project instead (`js/main.js` as the Vite entry, referenced from `index.html` with `<script type="module" src="/js/main.js"></script>`):

```js
import Reveal from 'reveal.js';
import 'reveal.js/reveal.css';
import 'reveal.js/theme/black.css';

import Markdown from 'reveal.js/plugin/markdown';
import Highlight from 'reveal.js/plugin/highlight';
import Notes from 'reveal.js/plugin/notes';

Reveal.initialize({
  hash: true,
  controls: true,
  progress: true,
  center: true,
  transition: 'slide',
  plugins: [ Markdown, Highlight, Notes ],
});
```

**The 6.0.1 package `exports` map has no `dist/` prefix and no `.esm.js` suffix** on these subpaths — `reveal.js/dist/reveal.css` and `reveal.js/plugin/notes/notes.esm.js` (patterns that work with the CDN build above) both 404 under npm/Vite. Verify with `npm view reveal.js@6.0.1 exports` before assuming a path.

## React project skeleton

```
src/
├── App.tsx              # <Reveal> root, plugin config
├── slides/
│   ├── IntroSlide.tsx     # one component per <Slide>
│   ├── DataSlide.tsx      # example: slide fed by live app state
│   └── index.ts
└── styles/
    └── custom.css
```

`App.tsx`:

```tsx
import { Reveal, Slide, Fragment } from '@revealjs/react';
import 'reveal.js/reveal.css';
import 'reveal.js/theme/black.css';

export function App() {
  return (
    <Reveal plugins={{ highlight: true, notes: true }} transition="slide">
      <Slide>
        <h1>Title</h1>
        <Fragment><p>Revealed second</p></Fragment>
      </Slide>
      <Slide>{/* next section */}</Slide>
    </Reveal>
  );
}
```

Vertical slides nest a `<Slide>` array under a parent group per the library's own grouping prop — check `@revealjs/react`'s current API surface (0.2.x is pre-1.0 and the grouping API may still shift) before relying on it for a production deck.
