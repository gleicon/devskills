# Tiger Style Reference

Source: https://tigerstyle.dev/
Full document: https://github.com/tigerbeetle/tigerbeetle/blob/main/docs/TIGER_STYLE.md

Tiger Style is TigerBeetle's engineering philosophy. TigerBeetle is a distributed financial database written in Zig that handles millions of transactions per second with strong correctness guarantees. The style guide is the encoding of the lessons learned building that system.

The central principle: **Do the hard thing today to make tomorrow easy.**

---

## Priority Order

**1. Safety** — Programs must not merely execute correctly, they must employ defense-in-depth. A correct program that fails in an unclear way is still a dangerous program.

**2. Performance** — Performance is a design-time concern, not a retrofit. Thinking about it late costs more than thinking about it early.

**3. Developer Experience** — Quality code is easier to maintain and understand. Simplicity and elegance are the outcome of hard work, not shortcuts.

---

## Safety Principles

### Assertions

Assertions are executable documentation of invariants. They are not error handling — they document properties that *must* hold, and if they do not, the program has a bug.

- Minimum 2 assertions per function.
- Validate arguments at function entry.
- Validate return values and invariants before returning.
- Assert at every external boundary (disk, network, OS).
- Use pair assertions: find at least two independent code paths that enforce the same property.

As Gerard J. Holzmann (NASA JPL) noted about safety rules: "initially they are perhaps a little uncomfortable, but after a while their use becomes second-nature and not using them becomes unimaginable."

### Control Flow

- No recursion unless termination is formally proven. Use iterative equivalents.
- No `goto`. No equivalent flow-jumping constructs.
- Every loop has an explicit iteration bound. Assert the bound is not exceeded.
- Minimize branches. Flatten nesting. Early returns are fine.

### Memory

- No dynamic allocation after initialization. Allocate statically upfront.
- Every queue, buffer, and collection has a fixed maximum capacity.
- Assert capacity before every append.

### Error Handling

- Every error is handled explicitly.
- No exceptions for control flow.
- Errors propagate via return values, not side channels.

### Dependencies

- Zero dependencies except the language runtime/stdlib.
- Each dependency is a liability: maintenance cost, security surface, upgrade friction, failure modes.
- If you must add a dependency, document the justification.

---

## Performance Principles

Performance optimization done late is expensive. Performance considered during design is cheap.

**Before writing a non-trivial system or function:**
1. Sketch the performance budget across four dimensions: network, disk, memory, compute.
2. Identify which dimension is the bottleneck.
3. Separate control plane from data plane through batching.

**Back-of-envelope sketch template:**
```
Network: <requests/sec> × <payload_bytes> = <bandwidth>
Disk:    <writes/sec> × <record_size_bytes> = <throughput>
Memory:  <concurrent_items> × <item_size_bytes> = <working_set>
Compute: <operations/sec> × <cycles_per_op> = <CPU budget>
```

**Hot path rules:**
- No allocation.
- Batch operations to amortize per-call overhead.
- Measure before optimizing.

---

## Code Structure

**Functions**
- Maximum 70 lines per function.
- One purpose per function.
- If you cannot name it with a verb-noun pair, it does too much.

**Naming**
- snake_case throughout.
- Include units: `timeout_ms`, `size_bytes`, `offset_sectors`.
- Include qualifiers: `is_valid`, `has_pending`, `was_committed`, `count_max`.
- No abbreviations unless the domain term is universally understood.
- Big-endian naming: general concept before specific qualifier.
  - Correct: `timestamp_created_at`
  - Incorrect: `created_at_timestamp`

**Scope**
- Declare variables as close to their use as possible.
- No global mutable state in performance paths.

**Comments**
- No comments that describe what code does — code should be self-describing.
- Comments document WHY: hidden constraints, non-obvious invariants, specific bug workarounds.
- If you want to write a comment saying "x must be Y here", write an assertion instead.

---

## Zero Technical Debt

TigerBeetle's position: fixing problems in production is exponentially more expensive than fixing them at design time. Deferred problems compound.

- Fix problems during design and implementation.
- If a fix cannot happen now, file a tracked issue with a deadline.
- No TODO comments without a linked ticket.

Edsger Dijkstra: "Simple and elegant systems tend to be easier and faster to design and get right, more efficient in execution, and much more reliable."

---

## Application to This Package

The `/ds-tiger-style-mode` command encodes these principles as session-level constraints for Claude Code, OpenCode, and equivalent tools. The Cursor rules apply them automatically based on file type. The language profiles include Tiger Style integration notes specific to each language's idioms.

Tiger Style is not a mechanical checklist — it is a habit of engineering rigor that protects against the most common causes of production failure: unchecked assumptions, deferred correctness, and complexity growth.
