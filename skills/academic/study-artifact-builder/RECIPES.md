# Recipes

Skeletons for the artifact in [SKILL.md](SKILL.md). Names are generic; adapt the data, keep the shape.

## 1. File layout

One HTML file, no external requests. Sections in this order:

1. Title line with the source versions and the review date
2. Know-by-heart box
3. Toggles, legend, detailed diagram
4. Detail panel under the diagram
5. Component cards (secondary navigation)
6. Results tables with a metric switch
7. Caps table, failure taxonomy
8. Sources and not-verified list

## 2. Data model

Keep every number in data, never inside markup, so the views and the know-by-heart box read from one place.

```js
const BENCHMARKS = ["Bench A", "Bench B", "Bench C"];
const FULL_RESULT = [92.7, 91.1, 92.9];

const ABLATIONS = {
  retrieval: { label: "Retrieval", result: [60.3, 73.3, 85.1], drop: [35.0, 19.5, 8.4] },
  repair: { label: "Repair loop", result: [66.2, 64.4, 84.8], drop: [28.6, 29.3, 8.8] }
};
const ABLATION_ORDER = ["retrieval", "repair"];

const COMPONENTS = {
  retrieval: {
    kind: "stage",
    title: "Case retrieval",
    ablation: "retrieval",
    paper: ["What the paper says, one claim per item"],
    code: ["What the code does, with file and line"],
    flags: [["diff", "A mismatch between the two"], ["unv", "A claim not verified"]]
  }
};
```

Derived values (means, ranks, relative changes) are functions over this data. Compute them; do not retype them.

```js
const mean = (values) => values.reduce((sum, v) => sum + v, 0) / values.length;

function meanDrop(key) {
  return mean(ABLATIONS[key].drop);
}

function rankOf(key) {
  const sorted = ABLATION_ORDER.slice().sort((a, b) => meanDrop(b) - meanDrop(a));
  return sorted.indexOf(key) + 1;
}
```

## 3. SVG helpers

Generate the diagram from helpers so coordinates stay in one place and a layout change is a data change.

```js
const LINE_HEIGHT = 15;

function textLines(lines, cx, cy) {
  const first = cy - ((lines.length - 1) * LINE_HEIGHT) / 2 + 4;
  const spans = lines.map((line, i) => {
    const dy = i === 0 ? 0 : LINE_HEIGHT;
    return `<tspan x="${cx}" dy="${dy}">${line}</tspan>`;
  });
  return `<text x="${cx}" y="${first}">${spans.join("")}</text>`;
}

function box(id, x, y, w, h, lines, cls) {
  const label = textLines(lines, x + w / 2, y + h / 2);
  return `<g class="dbox ${cls || ""}" data-id="${id}">` +
    `<rect x="${x}" y="${y}" width="${w}" height="${h}" rx="7"/>${label}</g>`;
}

function edge(d, color, dashed) {
  const style = dashed ? " dash" : "";
  return `<path class="edge e-${color}${style}" d="${d}" marker-end="url(#mk-${color})"/>`;
}
```

Layout rules that avoid the usual defects:

- Leave a gap of at least 40 px between lanes and route cross-lane flows through the gaps, one vertical per flow, 10 px apart.
- Put cross-lane buses outside the lanes (above for inputs, below for reuse), then up through a gap to the target.
- Keep text within the box: about 8 px per character at 12 px type, so a 220 px box takes about 26 characters per line.
- Give labels `text-anchor: start` next to vertical arrows, so the arrow does not cross them.
- Set a minimum width on the SVG and let the container scroll; do not shrink type below about 10 px.

## 4. Interaction pattern

State is two values: the selected component and the removed ablation. One `render()` updates every view from them.

```js
let selected = "retrieval";
let removed = null;

function render() {
  renderNodes();
  renderDiagram();
  renderDetail();
  renderAblationTable();
}

function toggleRemoved(key) {
  removed = removed === key ? null : key;
  render();
}
```

Clicking inside a lane selects the lane's component. Stores and code-only boxes are selectable on their own. The detail panel sits directly under the diagram.

Detail panel, three blocks:

```js
function renderDetail() {
  const c = COMPONENTS[selected];
  const blocks = [
    ["paper", "Paper", c.paper],
    ["code", "Code", c.code],
    ["flag", "Discrepancies and gaps", c.flags]
  ];
  // one column per block; chips: diff = does not match, unv = not verified
}
```

## 5. Headless verification

Use a headless browser that is already installed rather than adding a dependency. Two checks: no script errors, and one interaction per control.

```bash
# 1. render and dump the DOM
chrome-headless-shell --no-sandbox --disable-gpu \
  --virtual-time-budget=3000 --dump-dom file:///path/artifact.html > dom.html

# 2. screenshot, then look at it
chrome-headless-shell --no-sandbox --disable-gpu --window-size=1400,1500 \
  --virtual-time-budget=3000 --screenshot=shot.png file:///path/artifact.html
```

Interaction test: copy the file, append a script before `</body>` that clicks a toggle and dispatches a click on a store node, dump the DOM, and assert the expected classes and headings are present.

```js
document.querySelector("#toggles button[data-abl=retrieval]").click();
document.querySelector("#diagram .dbox.kbbox").dispatchEvent(
  new MouseEvent("click", { bubbles: true })
);
```

Assertions to grep for: the lane has the `removed` class, the detail heading names the store, the store has the `sel` class.

## 6. Extracting a paper table

Fetch the paper's HTML, find the table by its caption, and print cells with math converted to its alt text. Do not read a table from a search snippet.

```python
import re
import html
import urllib.request

def clean(cell):
    cell = re.sub(
        r"<math.*?</math>",
        lambda m: " " + (re.findall(r'alttext="([^"]*)"', m.group(0)) or [""])[0] + " ",
        cell,
        flags=re.S,
    )
    text = re.sub(r"<[^>]+>", "", cell)
    return re.sub(r"\s+", " ", html.unescape(text)).strip()

def table_rows(page, caption):
    start = page.find(caption)
    segment = page[start:start + 60000]
    rows = re.findall(r"<tr.*?</tr>", segment, flags=re.S)
    return [[clean(c) for c in re.findall(r"<t[dh].*?</t[dh]>", r, flags=re.S)] for r in rows]
```

## 7. Pre-handover checklist

- [ ] Every number traced to a source table or a computation
- [ ] Two derived numbers recomputed by hand
- [ ] Every component has paper, code and discrepancy blocks, even when a block says "nothing to note"
- [ ] Caps listed per variant with the unit stated
- [ ] Code-only components drawn apart and dashed
- [ ] The know-by-heart box matches the final tables
- [ ] Rendered with no script errors, one interaction tested per control
- [ ] Screenshot reviewed for clipped lanes and overlapping labels
- [ ] Not-verified list at the bottom is complete
