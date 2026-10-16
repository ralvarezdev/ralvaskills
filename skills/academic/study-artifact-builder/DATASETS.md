# Studying a dataset or benchmark

How to take a benchmark or corpus from its description to something you can defend. Apply it before building the dataset views in [SKILL.md](SKILL.md#4-build-order).

## 1. Role first

A dataset plays one or more roles. Studying it as the wrong role produces the wrong questions.

| Role | The question that matters |
|---|---|
| Evaluation benchmark | What does a score on it mean, and what does it not show |
| Training corpus | What is learned, what is contaminated |
| Retrieval corpus | What can be found, and does it contain the answers |
| Knowledge base | Who curated it, how it is indexed, how it is refreshed |

One object can be several at once. When the same library is both the benchmark and the retrieval corpus, treat that as the central finding ([§4](#4-leakage-and-overlap)).

## 2. Dataset card

Fill every row. A blank is a finding.

| Field | Content |
|---|---|
| Name and version | As published; version or snapshot date |
| Provenance | Who produced it, from what source, under what license |
| Construction | How items were collected, filtered, annotated, by whom |
| Size | Items per split and per subset |
| Item schema | Input fields, ground truth, metadata |
| Ground truth | Human-written, reference implementation, formal spec, model output |
| Metric | Exact definition, and what counts as a pass |
| Evaluation harness | Tools and versions that produce the verdict |
| Known limitations | What the authors state, and what you found |

## 3. Metric definitions

Papers reuse metric names for different things.

- **Write the operational definition.** "Pass rate" may mean compiles, passes tests, or passes a formal check. Say which.
- **Name the oracle.** A real vendor tool, a test suite, a model judge, a human panel.
- **Layered metrics.** When there are several (syntax, functional, verified), draw the funnel and the denominators.
- **Averages.** Average of per-item counts, average over passed items only, or total over total.
- **Filters in the denominator.** Items dropped for any reason change what a percentage means; record the rule.

## 4. Leakage and overlap

The question to answer: can the system see the answer?

| Check | How |
|---|---|
| Name overlap between evaluation items and the retrieval corpus | Case-insensitive match in both directions, and containment |
| What the paper's filter catches | State its rule, for example mutual substring containment |
| What it cannot catch | Semantic duplicates, renamed copies, shared helper code |
| Whether the corpus is built from the benchmark's own source | Check the construction section, not only the file names |
| Pretraining exposure | Public libraries are likely in a model's pretraining data |

Compute the overlap yourself when the corpus is available, and report it as a percentage per benchmark. Then interpret each retrieval-related result with it in view: a large retrieval effect on a benchmark that fully overlaps its corpus is partly a lookup effect.

## 5. Strata and noise

- **Strata.** Difficulty levels, task types (control logic versus utility functions), size or complexity bins. Report per stratum when the paper does, and note when it does not.
- **Sample size.** Convert one item into percentage points for each subset. State it next to every table that depends on it.
- **Sample-based claims.** For any "we sampled N" in the paper or in your own checks, record the seed and the selection rule.
- **Pairing.** Comparisons across methods should use the same items; unpaired subsets are weaker.

## 6. Reproducibility artifacts

| Artifact | Check |
|---|---|
| Generated outputs published | Allows recompiling or re-scoring without regenerating |
| Configs and prompts | Needed to rerun; note which are in the repository |
| Evaluation tooling | Version, license, availability (some vendor tools are not public) |
| Randomness | Seeds, temperatures, number of runs |

Re-scoring published outputs is the cheapest independent check available. It tests the verdict, not the generation.

## 7. Views to put in the artifact

| View | Content | Use |
|---|---|---|
| Dataset card | The §2 table | Everyone |
| Size and strata | Items per subset, per stratum, in one chart | Defend, compare |
| Metric funnel | Definitions and denominators | Defend |
| Overlap panel | Percent overlap per benchmark and the filter's rule | Defend, replicate |
| Example browser | A few items with schema fields | Memorize, implement |
| Noise note | Points per item per subset | Defend |
| Reproducibility list | The §6 table | Replicate |

## 8. Focus by goal

| Goal | Lead with | Keep short |
|---|---|---|
| Memorize | Sizes, subset names, metric definitions | Construction detail |
| Defend | Leakage, noise, limitations, metric definition | Item examples |
| Replicate | Harness, versions, seeds, artifacts | Provenance prose |
| Compare | Same card rows for each dataset | Per-dataset narrative |
| Implement | Item schema, ground truth format | Results |

## 9. Failure modes

| Symptom | Cause | Guard |
|---|---|---|
| A percentage nobody can reproduce | Denominator filter not recorded | Metric definitions block |
| "Retrieval is decisive" | Corpus overlaps the benchmark | Overlap panel |
| A ranking built on a few items | Subset too small | Noise note |
| Two datasets compared as equals | Different oracles | Compare the §2 cards first |
| Counts that do not add up | Subsets overlap or filters differ | Recompute totals from the item list |

## 10. Derived reference sets

Many studies lean on a map built by a script: item to the things it uses (functions, types, variables), item to its plan, item to its tags. Before it is used as ground truth or as a retrieval source, check it.

| Check | How | What it catches |
|---|---|---|
| Whole-word presence | For each (item, entry) pair, does the entry appear as a whole word in the item's source | Matches by substring ("tan" inside "instance") |
| Frequency outliers | List the most frequent entries and check one against the source | An entry listed for most items and used by one |
| Language keywords | Count entries that are keywords of the language | Noise that the consumer may drop silently |
| Coverage | How many benchmark items have an entry | A reference that exists for some subsets only |
| Consumer behavior | What happens to an entry the dictionary does not have | Silent drops that hide the noise from the output |

Report the share of pairs that pass the whole-word check per subset. A reference set below the level you need (for a recall score, a high share) is rebuilt by parsing the source, not patched by hand.

## 11. Numeric audit of a results table

1. For each rate of a count, find an integer k such that k over the stated n gives the printed value at its precision. Report the cells that do not, and the other n that does fit.
2. If no common denominator exists across a column, say so; averaged runs or exclusions are possible, and a typo is too.
3. Recompute each cited percentage from the table. State the base the text used.
4. Compare the table across tables: the full configuration of an ablation should equal the corresponding row of the main table; the denominators of an efficiency table should equal the compile counts of the main one.
5. Compare the definitions of a same-named metric in the code of both papers before comparing numbers.

Script in [RECIPES.md](RECIPES.md#9-numeric-audit).
