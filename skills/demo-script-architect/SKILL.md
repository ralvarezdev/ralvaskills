# Demo Script Architect Skill

## Purpose
Design and refine presenter-centric demo scripts that guide audiences through system capabilities in a natural, engaging narrative flow. Works for any technical or product demo.

## When to Use
- Creating live demonstration scripts
- Building structured walkthroughs of systems/features
- Designing progressive capability reveals
- Turning technical specs into engaging narratives
- Adding visual guidance for presenters

## Core Principles

### 1. **Narrative Flow Over Feature Lists**
- Each section builds on previous learning
- Demos are interconnected with bridge narration explaining the "why"
- Story arc: observe → understand → apply → verify
- Conversational tone, not technical documentation

### 2. **Element Documentation**
- Always include element purpose when referenced: `YI-100 (pump status indicator)`
- Keep names/references explicit in queries (for learning and precision)
- Help audience understand what they're looking at

### 3. **Visual Guidance Mapping**
Map abstract references to actual visible elements on screen:
- Reference invisible elements → Point to **visible visual counterpart**
- Example: Data field "status" → Point to **status icon/light**
- Never point to something the audience can't see on screen

### 4. **Query-to-Explanation Alignment**
- **Queries/Commands** should be explicit and clear
- **Processing Explanation** must accurately describe what's happening:
  - Match what's actually being queried (don't claim context-learning if tags are explicit)
  - Explain the mechanism (parallelization, routing, lookup, etc.)
  - Set up expectation for the response that follows

### 5. **Narrative Section Structure**

```
**Demo Setup** 
→ Explains purpose and what to watch for

**Visual Cue** 
→ "[On screen: Point to X, watch Y]"

**Action/Query**
→ What the system is being asked to do (explicit)

**Processing Explanation**
→ Bridge that explains what's happening (2-3 sentences, audience-friendly)

**Expected Response(s)**
→ Multiple outcomes based on system state

**Commentary**
→ Interpreter results, relate to broader picture
```

### 6. **Avoid Redundancy**
- Don't repeat the same query/capability in consecutive demos
- Build progressive revelation: basic → complex → combined → edge-cases
- Stack elements only when combining for a **new insight**
- Remove queries that just re-confirm what was already shown

**Example progression:**
- Demo A: Query individual data points
- Demo B: Query related data groups
- Demo C: Query for diagnostic/troubleshooting
- Demo D: Write commands and verify

NOT:
- Demo A: Query X
- Demo B: Query X again (same purpose)
- Demo C: Query X again (same purpose)

### 7. **Natural Language & Conversationalism**
- Use contractions: "it's", "we've", "that's", "pump's"
- Direct audience: "Watch", "Notice", "Observe", "See how..."
- Short, punchy sentences over long lists
- Tell a story, don't enumerate features

**Bad:** "The system retrieves data from three distinct protocols simultaneously and correlates them into a unified response."

**Good:** "Three different systems are being hit at once — behind the scenes, they're all talking. You get back one clean answer."

### 8. **Response-to-Commentary Alignment**
- Narrator commentary should interpret results, not just restate them
- Commentary explains significance, not mechanics
- Sets up the next logical question or demo

## Demo Progression Template

| Position | Focus | New Capability | Approach |
|----------|-------|-----------------|----------|
| 1 | Basics | Single element access | Explicit reference |
| 2 | Integration | Multiple related elements | Grouped query |
| 3 | Abstraction | Natural language, no explicit refs | System infers intent |
| 4 | Dynamics | Live data, change over time | Continuous monitoring |
| 5 | Accessibility | Different access methods | Voice/touch/alternative interface |
| 6 | Automation | System behavior without command | Observe automatic logic |
| 7 | Diagnostics | Troubleshooting scenario | Bundle related elements |
| 8 | Knowledge | Search/lookup capability | Find answers in resources |
| 9 | Control | Issue commands | Write operations + verification |

## Checklist for Script Review

- [ ] Each demo has a **Setup** explaining the purpose
- [ ] **Actions/Queries** are explicit and clear
- [ ] **Processing Explanations** accurately describe what's happening
- [ ] Visual cues point to **visible screen elements**, not abstract references
- [ ] Element names include purpose description: `NAME (what it does)`
- [ ] **Commentary** interprets results (not just describes them)
- [ ] **Bridge narration** connects each demo to the next with a "why"
- [ ] No redundant queries (same action shown twice for same purpose)
- [ ] Multiple expected responses cover realistic system states
- [ ] No robotic technical jargon in explanations
- [ ] Tone is conversational and engaging throughout

## Common Anti-Patterns

### Problem: Processing Explanation vs. Query Mismatch
```
Query mentions: "YI-100, SI-100, and PIT-100"
WRONG explanation: "Notice we didn't mention specific tags this time..."
RIGHT explanation: "Three different data points from two systems in parallel..."
```

### Problem: Pointing to What's Not Visible
```
WRONG: "[Point to field named 'PIT-100' on screen]" (if not labeled there)
RIGHT: "[Point to the pressure gauge display]" (what the audience actually sees)
```

### Problem: Robotic Explanations
```
WRONG: "System processes query, retrieves values from database, returns formatted response."
RIGHT: "Watch the system grab the values and hand them back as one answer."
```

### Problem: Redundant Demos
```
WRONG: 
- Demo 1: Show status query
- Demo 2: Show same status query (again)
- Demo 3: Show different element

RIGHT:
- Demo 1: Show status query
- Demo 2: Show how to combine status with related data
- Demo 3: Show different element or new capability
```

### Problem: Weak Bridge Narration
```
WRONG: "Now let's do a different demo." (disconnected)
RIGHT: "Now that we know the status, let's check the pressure to see what it's delivering." (logical progression)
```

## What Makes a Good Demo Section

1. **Clear purpose** - Audience knows why they're watching
2. **Visual guidance** - Presenter knows what to point at, audience knows what to see
3. **Explicit action** - Audience understands the request clearly
4. **Meaningful explanation** - Not just description, but interpretation
5. **Logical next step** - Sets up the following demo
6. **No repetition** - New insight or capability each time

## Output Deliverables

1. **Structured script** with proper narrative sections
2. **Visual element map** (reference → visible screen element)
3. **Demo flow diagram** showing connections and progression
4. **Redundancy audit** (what was removed/consolidated and why)
5. **Presenter notes** (what to watch for, where to point, what to emphasize)

## Key Questions When Reviewing a Demo

- **Purpose:** Is it clear *why* we're doing this? Does it build on the previous demo?
- **Clarity:** Can the audience understand the action without technical jargon?
- **Visuals:** Can the audience see what we're referring to?
- **Progression:** Is this demo different from previous ones, or redundant?
- **Tone:** Does it sound like a person explaining, or a manual reading?
- **Significance:** Does the commentary explain why the result matters?
