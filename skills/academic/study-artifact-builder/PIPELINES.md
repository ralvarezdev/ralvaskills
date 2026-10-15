# Studying a pipeline

How to take a method from a paper and its code to a component-level understanding. Apply it before building the diagram in [SKILL.md](SKILL.md#5-diagram-rules).

## 1. Classify the control flow first

The classification decides what the diagram can honestly show.

| Type | Signature | What to draw |
|---|---|---|
| Fixed pipeline | One function calls stages in a set order, each toggled by a flag | Lanes in order, flags as ablations |
| Closed loop | Generate, check, repair, repeat until a cap | The loop with its cap and exit conditions |
| Multi-agent | Roles with their own prompts, a planner chooses | Roles, who sends what to whom, who decides to continue |
| Agent with tools | One model picks tools turn by turn | The tool surface, guards and stop checks, not a fixed order |

If the paper's figure and the code disagree on the type, say so; the paper may describe a multi-agent system that the code runs as a fixed pipeline.

## 2. Component inventory

One row per component. Do not merge components that the paper names separately.

| Field | Question |
|---|---|
| Name | The paper's name, and the code's name if different |
| Input | What exactly enters (text, plan, retrieved items, diagnostics) |
| Output | What exactly leaves, and its format |
| Model or algorithm | LLM call, retriever, rule, compiler or external tool |
| Parameters | Counts, thresholds, temperatures, top-k, caps |
| Disabled by | The flag, config or switch that turns it off |
| Source | Paper section and code file and line |

## 3. Retrieval components

Retrieval is where papers are vaguest and where the details matter most in a defense.

- **What is indexed.** Whole items, metadata, or generated summaries. The text embedded is often not the text returned.
- **Dense or sparse.** Name the embedding model if declared; if it is behind a managed service and undeclared, write that.
- **Candidates and selection.** How many are fetched, whether there is a rerank, what ranks them, how many survive.
- **What is injected.** The item, its plan, its code, only its metadata. Say which parts are withheld and why.
- **Leak control.** The filter that keeps the answer out of the retrieved set, and what it cannot catch.
- **Corpus construction.** Who built it, from what, and whether it overlaps the evaluation data ([DATASETS.md](DATASETS.md#4-leakage-and-overlap)).
- **Second-order retrieval.** One stage reusing another's output as a source (for example recommendations drawn from retrieved cases) is a data flow to draw.

## 4. Loops and caps

- **Unit of the cap.** Iterations, compile calls, tool calls, tokens. Confirm by reading the loop, not the config key's name.
- **Order inside the loop.** Check, then repair, or repair, then check. The order decides whether the last repair is verified.
- **Exit conditions.** Success, cap, stuck detection, budget.
- **Default per variant.** Caps often differ by language, dataset or config file. List all of them.
- **What the repair sees.** Raw diagnostics, filtered, prioritized, or only the first error.
- **Repair shape.** Full regeneration or patches; a patch format has its own failure modes.

## 5. Mapping ablations to code

A paper's ablation rows rarely equal the code's flags one to one.

| Check | Why |
|---|---|
| Each ablation row maps to a named flag or config | Otherwise the numbers cannot be reproduced |
| Flags without an ablation row | Features the paper did not measure |
| Ablation rows without a flag | Possibly hand-edited runs; note it |
| What "removed" means | Skipped, replaced by a null, or replaced by a weaker variant |
| Shared infrastructure | Stores used by several stages are usually not ablated; their effect shows through the dependent stages |

## 6. Paper-versus-code discrepancy checklist

Run this on every study. Each hit becomes a discrepancy block on the component.

- [ ] Stage order
- [ ] Number of stages, and components that exist only in code
- [ ] Caps and defaults, per variant
- [ ] Model, temperature and sampling defaults against the paper's setup
- [ ] Retrieval parameters (top-k, thresholds, rerank)
- [ ] Features added locally that the paper does not have (alternative backends, extra filters)
- [ ] Dead code and components not wired into the workflow
- [ ] Defaults that disable a paper feature
- [ ] Environment versions (tools, vendors, runtimes)

## 7. Failure taxonomies

When the paper reports error types:

1. Copy the category definitions as the paper states them.
2. Compute the reductions from the table and check them against the paper's prose.
3. Record which component the paper credits for each category, and mark it as an attribution, not a measurement, unless a per-component run backs it.
4. Note the categories that improve little or get worse. They show where no component is aimed.

## 8. Comparing two pipelines

Use the same rows for both and the same vocabulary. Do not compare a paper to a code base.

| Concern | Method A | Method B |
|---|---|---|
| Control | | |
| Planning | | |
| Retrieval of examples | | |
| Retrieval of reference or API material | | |
| Verification | | |
| Repair | | |
| Stopping | | |
| What the ablation says matters most | | |

A weak ablation (few tasks, one model) is a finding in itself; do not rank components from it as if it were strong.

## 9. What goes in the artifact

| Element | Content |
|---|---|
| Know-by-heart box | The ranking, the numbers, the caps, the counts |
| Detailed diagram | Lanes, flows, stores, loop with cap, ablation badges |
| Component panel | Paper, code, discrepancies |
| Results views | Relative drop, resulting value, secondary metric, switchable |
| Cap table | Every source of the cap with its value |
| Failure taxonomy | Paired bars per category with attribution |
| Not verified list | One line per gap, at the bottom |
