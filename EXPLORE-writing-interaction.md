# Writing & interaction rules — exploration

Captured 2026-08-05. Research and decisions from one session. Nothing here is built
yet. The purpose is to preserve findings so the build session can decide from
evidence instead of re-deriving.

---

## 1. The problem

Stated by the user, about assistant behaviour:

- Replies routinely contain **several questions**.
- Even a single question is phrased so the answer is **ambiguous**.
- Simple yes/no questions get transformed into "yes, no, **or**…" and branch outward.
- Long questions **pivot mid-sentence**: the reader starts composing an answer to the
  first clause, and by the time they finish reading, the answer they typed no longer
  fits.
- One answer can **match more than one** of the questions posed, so neither party
  knows which was answered.
- The same defects **leak into comments and documentation**, not just chat.

Confirmed live in-session. Three examples from this conversation:

- "Run it? Or want to look at `feat/structural-diff` first?"
- "Want me to check anything else — open PRs, or the local tree state?"
- "Want me to run that, or would you rather first decide whether this skill belongs
  in the fork at all?" — the second clause silently reframed the whole task.

And the collision happened for real: a question about where security findings should
come from was answered with a branch link that addressed a *different* question left
open two turns earlier. Two live questions, one answer.

---

## 2. Two problems, two tracks

They share a root but need different instruments.

| | Track 1 — Interaction | Track 2 — Text |
|---|---|---|
| Surface | questions asked of the user | comments, docs, commit bodies, any `.md` |
| Defect | cardinality, placement, internal structure | ambiguity, drift, slop, AI tells |
| Level | above the sentence | at and below the sentence |
| Delivery | `agents-md/system/` block **+** a `-mode` skill | rules into existing skills **+** two new skills |

**Key finding:** no controlled-language or plain-language standard addresses Track 1.
The double-barreled question is not a sentence-level property, so ASD-STE100,
ISO 24495-1, and the Federal Plain Language Guidelines are all silent on it.
Conversely, the ADHD-output work addresses cardinality and placement but not the
internal structure of a single question. **Neither body of work alone is sufficient.**

---

## 3. Track 1 — Interaction

### 3.1 The named defect

**Double-barreled question** — the precise technical term from survey methodology: a
single question asking about two or more topics while permitting one answer. The
literature's diagnosis matches the user's complaint exactly: the respondent may feel
differently about each part, the single answer cannot express that, and the result is
unusable.

Canonical fix, mechanical enough to encode as a rule:

> One topic per question. Whenever tempted to join two ideas with "and" or "or",
> split them.

### 3.1b Second defect — mid-message pivot

Found in-session, after the first was diagnosed, and **not covered by any of the three
sources**.

A handback listed actions in order — create branch, add, commit, then run
`/ds-project-checkpoint` — and only afterwards stated the condition that made the
checkpoint premature. The user reads and executes at the same time, so by the time the
condition arrived, the checkpoint call was already typed and had to be cleared.

Every individual line was correct. The message was still broken.

> A message containing actions must be executable top to bottom without backtracking.
> Anything that defers, conditions, or invalidates an instruction goes **before** that
> instruction, never after. Check: could the reader run every line in order and be
> correct?

Same shape as the mid-sentence pivot, one level up. Survey design has no execution
step; plain language has no reader-as-actor; and i-have-adhd's rule 3 — "end with one
concrete next action" — points the wrong way for an agent handing back several
commands, so it would have produced the same bug.

### 3.2 Convergent prior art — `ayghri/i-have-adhd`

Reached part of the same answer from ADHD working-memory limits rather than survey
design.

**Converges with us:**

- Rule 4: *"A question that comes up mid-work is not a tangent: answer it yourself if
  you can and fold the result in. If it still needs the reader, surface it once, at
  the end."*
- Rule 3: *"name ONE thing the reader can do in under two minutes."*
- Break-glass 4: *"One short clarifying question beats guessing and rewriting."*

**Does not cover us.** Its rules govern *how many* questions and *where* they sit,
never what is inside one. The offending example above is one question, at the end of
the turn, no tangent, no closer — it passes its rules 3, 4, and 10 cleanly.

**Worth taking:**

1. **The pre-send check.** A delete-list applied before sending rather than a
   principle to hold in mind: drop the announcing first sentence, drop the "anything
   else?" last sentence, drop the "by the way" sidebar, drop empty hedging adverbs,
   drop idioms. Then a falsifiable test — *"if the reader reads only the first line
   and the last line, do they know (a) what to do next and (b) what just happened?"*
2. **Its hedging caveat** — *"Keep a hedge that carries real uncertainty; deleting it
   manufactures confidence."* This is the plain-language-flattens-nuance risk already
   solved in one sentence. Lift the idea.
3. **`evals/`.** A weighted rubric (Correctness 35%, **Autonomy 25%**, Actionability
   20%, Safety 10%, Concision 10%), `cases.jsonl`, a runner, and a release gate: no
   blocking findings, correctness and safety within 0.1 of baseline, weighted score
   above baseline. The Autonomy dimension — *"does not push avoidable work to the
   user"* — is a direct counterweight to over-asking.

**Not worth taking:**

- Rule 6 (time estimates in concrete minutes) — invites fabrication from an agent;
  their own break-glass 6 already retreats from it.
- Rule 5 (restate state every turn) — fights concision and duplicates what
  `/ds-project-resume` does deliberately.
- The always-on SessionStart hook — Claude-only hook mechanism, same category already
  rejected from mattpocock/skills.

### 3.3 Theoretical frame

**Grice's maxims** — Manner (avoid obscurity, avoid ambiguity, be brief, be orderly)
plus Quantity (do not say more than required). Good as the stated *why*. Useless as a
checkable *what*. Cite, do not encode.

### 3.4 Decisions

- Ship in **`agents-md/system/`** (always-on, preventive) **and** as a **`-mode`
  skill** (for users who do not use AGENTS.md, use a different one, or want it only
  some sessions). Two audiences, not two copies — see §6.1.
- The two artifacts get **different lengths on purpose**: the block is the minimum
  that prevents the defect, in the register of `concise.md` (five bullets); the mode
  carries rationale, before/after pairs, the pre-send check, and break-glass cases.
- The short version should be a **strict subset** of the long one, so it cannot
  contradict it.

---

## 4. Track 2 — Text and writing

### 4.1 Standards evaluated

| Source | Value | Cost / blocker |
|---|---|---|
| **ASD-STE100** | ~60 writing rules are genuinely good | Licensed, cannot embed. ~900-word aerospace dictionary is dead weight. Deliberately hostile to nuance, which fights our own "comments explain WHY" rule. |
| **ISO 24495-1:2023** | Best *frame*: reader-first. Four principles — relevant, findable, understandable, usable. Scope explicitly names technical writing and controlled languages. | Paid standard, cannot embed. Principle level only; nothing mechanically checkable. |
| **Federal Plain Language Guidelines** | Concrete, rule-shaped, **public domain** — usable with citation only | US-government register; needs adaptation for technical docs. |
| **`danyuchn/asd-ste100-skill`** | Prior art for exactly this repurposing. Drops the official dictionary, keeps the principles: *"plainest available word, used the same way every time."* Targets agent-facing English. | Packaged as a skill, which is the wrong level for the preventive half. Not certified STE. |
| **`blader/humanizer`** | 33 patterns from Wikipedia's *Signs of AI writing* (WikiProject AI Cleanup) — empirical, drawn from thousands of observed instances | Different objective, see §4.2. Source is CC BY-SA, see §6.4. |

**Licensing is decisive.** STE cannot be embedded, ISO cannot be embedded, the Federal
Guidelines can. For text shipping in `agents-md/system/` to every devskills user, that
settles which source supplies quotable material.

### 4.2 The objective conflict — why two skills, not one

Humanizer optimizes for **not sounding like AI**. Plain language optimizes for **being
understood**. They overlap, then diverge hard.

Evidence from humanizer's own worked example: its "after" contains *"By the second day
my calves had opinions"* and *"Lisbon does not bend over backward to make things easy
for you."* That is better travel writing and *further* from plain language than the
input — longer sentences, figurative phrasing, an idiom. Its pattern **#31 prescribes
"varied sentence lengths"**, which directly fights every sentence-cap rule in the test
prompt.

A single "make text better" skill holding both objectives resolves that conflict
arbitrarily, per sentence, invisibly. **Two skills with named objectives do not.**

### 4.3 Humanizer patterns — triage

**Transfer to our surfaces (comments, docs, commit bodies):**

- **#30 Diff-anchored writing** — *"This function was added to replace…"* → describe
  what it does, not what changed. Best single catch in the list for us; already half
  encoded in AGENTS.md as "never cite plan/ticket IDs", but wider.
- #9 negative parallelism ("It's not just X, it's Y")
- #11 synonym cycling
- #16 inline-header lists (`**Performance:** Performance improved`)
- #17 Title Case Headings
- #18 emoji
- #20 chatbot artifacts
- #22 sycophancy
- #23 filler phrases ("in order to" → "to")
- #24 stacked hedging
- #28 signposting ("Let's dive in")
- #33 rhetorical openers ("Honestly? It depends…")

**Do not transfer:**

- #1–#6 — Wikipedia-neutrality concerns (significance inflation, notability
  name-dropping, promotional language). Not our failure mode.
- #31 — actively conflicts with sentence caps.
- Voice calibration — meaningless for a code comment.

### 4.4 Triple convergence

Three independent sources landed on the same rules: synonym cycling / term drift,
filler phrases, and stacked hedging appear in ASD-STE100's principles, in the Federal
Plain Language Guidelines, and in Wikipedia's AI-cleanup patterns. Derived from
controlled-language work, government plain-language work, and encyclopedia cleanup
respectively. That is good evidence these rules are real rather than taste.

### 4.5 Decisions

- Apply the shared rules inside **existing** skills — `ds-doc-quality-review` first.
- Add **two** new skills:
  1. a **generic clarity** skill, not tied to comments or docs, usable on any text;
  2. a **humanizer-type** skill for AI-tell removal.
- Boundary must be written into each description explicitly — see §6.3.

---

## 5. The test instrument

Run against `doc/overquota/` — 14 files, 4,650 lines, 4,279 prose units. A technical
reference set, already carefully written. Report-only, before/after, no file edits: for
a test, the rewrite is what proves a rule earns its place.

**Results in §5.2.** The prompt is reproduced first so the two can be read together.

```
Review the Markdown files in this project for plain-language defects.
Report only — do not edit any file.

Scope: all .md files.
Skip: fenced code blocks, inline code, URLs, and material quoted from
elsewhere. Rewrite prose only.

Flag a sentence when it breaks one of these:

1.  Two ideas, one sentence. It makes two claims, or states a thing and
    then qualifies it into a different thing. Split it.
2.  Mid-sentence pivot. It sets up one direction and turns: "X — though
    actually Y", "A, or rather B", "not X so much as Y". Commit to one.
3.  Length. Over 25 words describing, over 20 instructing. Split it;
    do not compress it.
4.  Term drift. The same thing called two names. Pick one, use it
    everywhere, and report every site.
5.  Actorless passive. Passive voice where the actor matters and is
    missing. Name who acts.
6.  Stacked hedges. More than one hedge in a sentence ("might sometimes
    possibly"). Keep one, or state the real condition.
7.  Noun pile. Three or more nouns in a row acting as one modifier.
    Unpack it.
8.  Dropped articles. Telegraphic style: "Set flag to enable feature."
    Restore "the" and "a".
9.  Elegant variation. A synonym chosen for variety rather than meaning.
    Use the same word every time.
10. A word with a plainer twin: utilize→use, in order to→to,
    prior to→before, in the event that→if.

Report each finding as:

    <file>:<line>
    rule:   <number and name>
    before: <the sentence as written>
    after:  <the rewrite>

Constraints on the rewrite:
- Preserve every fact, condition, and qualifier. If a rewrite would lose
  nuance, do not rewrite it — report it as an exception and say what
  would be lost.
- Do not shorten for its own sake. Precision beats brevity.
- Keep the document's existing terminology even where you dislike it,
  unless rule 4 applies.

Close with:
- Findings per file, counted by rule.
- The three rules that fired most, and whether they cluster in
  particular files.
- Any place where two rules disagreed, or where a rule fought the
  document's purpose.
```

### 5.2 Results

31 distinct findings — 28 single-site, 3 cross-file — plus an 802-item census under
rule 3. **Caveat that governs every number below: one corpus, and a good one.** These
documents are careful technical reference prose, not raw AI output. Low yield here is
evidence about *this* register, not proof a rule is worthless.

| Rule | Fired | Verdict |
|---|---:|---|
| 1 — two ideas, one sentence | 7 | **Keep.** Highest-value rule. Caught a vacuous comparison in the opening argument of the whole set — "the clearest comprehension defect in the set". |
| 2 — mid-sentence pivot | 3 | **Keep.** Low count, high precision. One hit was the named pattern verbatim ("not X so much as Y"). Our signature defect. |
| 3 — length | 802 | **Redesign.** See §5.3. |
| 4 — term drift | 2 (×42 and ×36 sites) | **Keep, highest leverage.** Cross-file, invisible to a human reading one file at a time. |
| 5 — actorless passive | 8 | **Keep, with a carve-out.** See §5.4. |
| 6 — stacked hedges | 1 | Near-dead here. |
| 7 — noun pile | 3 | Weak-keep. Insight: the corpus hyphenates compounds, and hyphenation *is* the fix — so the rule is quiet wherever that discipline holds. |
| 8 — dropped articles | 2 | Near-dead here. Both justified by inconsistency with their own table, not by the rule. |
| 9 — elegant variation | 2 | **Merge into rule 4.** Both are "one thing, two names". |
| 10 — plainer twin | **0** | Zero hits in 4,650 lines, after checking ~30 words and clearing 3 false positives. |

**Review yield and prevention value are different axes.** Rules 6, 8 and 10 barely
fire on well-written prose, but stating them preventively costs one clause each and
they are trivially cheap to obey. They may belong in the agents.md block while failing
to earn a place in the review skill. Do not conflate "found nothing" with "not worth
saying".

### 5.3 Rule 3 broke, and how it broke matters

At the specified threshold it fires on **802 of 4,279 prose units — 19% of all prose in
the set**. That is not a finding list, it is a description of the house style.

The diagnosis is the valuable part. These sentences routinely carry a claim, its
measured magnitude, its direction and its citation *together*, because the set's own
stated standard is that every mechanism is verified and every measurement is stated
with the query that produced it. The reviewer's example: a 30-word sentence that is
"four facts that must be read together to be usable. Splitting it into four sentences
would satisfy the rule and make the passage harder to act on."

**This is the flattens-nuance risk from §4.1, confirmed empirically rather than
predicted.**

Three things follow:

1. **The instruction cap held; the description cap did not.** 20 words for an
   instruction was sound — the worst single finding in the set is a 68-word
   *instruction* opening a ~400-word paragraph containing three further actions. 25
   words for description was far too tight.
2. **A rule that must be silently reinterpreted to be useful is badly specified.** The
   reviewer re-thresholded on its own initiative to ~45 words describing / 25
   instructing, then reported a census instead of 800 findings. That was the right
   call, and it should be *specified* rather than improvised.
3. **The output contract needs a bound.** Left unbounded, rule 3 would have buried the
   25 findings under rules 1, 2, 4, 5, 6, 7, 8 and 9 — the ones actually worth acting
   on. A review skill needs "report the N worst, census the rest" written into it.

Candidate reformulation for the build session: drop the raw word count for description
and test the *structure* instead — a sentence may carry one claim plus its evidence;
it may not carry a chain of qualifications. Word count stays as the instruction rule,
where it worked.

### 5.4 Rule conflicts observed in practice

These are the most transferable findings, because each one is a resolution we would
otherwise have had to guess at.

- **Rule 1 beats rule 3.** Splitting for ideas sometimes leaves a 33-word first half.
  The reviewer took rule 1 as governing and noted the residue, reasoning that a clear
  33-word sentence beats two muddled ones — and that rule 3's own instruction is
  "split it; do not compress it". Correct, and worth encoding as a precedence rule.
- **Rule 5 has a person/system boundary.** Where the missing actor is a *system* — the
  Demand API, the three services, this document — naming it costs nothing and adds
  precision. Where the missing actor is a *person*, the passive is required and
  correct. The reviewer declined those, citing a standing no-blame constraint. The
  carve-out has to be in the rule, not discovered per-run.
- **Rule 4 can create ambiguity instead of removing it.** For one contested pair,
  collapsing onto a single name was wrong because that name was already overloaded —
  a code identifier, a gate, and a measurement column. Resolution: **declare the
  distinction once, then hold to it.** "Pick one name" is the wrong instruction when
  the collision is with code.
- **A missing rule surfaced.** The reviewer flagged a British spelling in a bolded
  instruction, against the set's own documented convention — outside all ten rules.
  That is *convention consistency*, distinct from term drift. Candidate rule 11.

### 5.5 The exceptions list

Three declines, each with a reason worth keeping:

1. A four-part parallel sentence where "the single-sentence form **is** the argument" —
   splitting it would lose the claim that these are one work item, not four.
2. A derivation where each clause's bound depends on the previous one — "splitting it
   turns a derivation into a list of assertions, and the reader loses the chain that
   makes the conclusion follow."
3. A passive whose actor is a person, protected by a standing no-blame constraint.

**These three sentences are the shape of what the agents.md block must not forbid.**
Parallel structure, derivation chains, and protective passives are all correct writing
that a naive plain-language rule would destroy.

---

## 6. Cross-cutting decisions

### 6.1 The duplication question — resolved

Concern raised: the same rule set in `agents-md/system/` **and** in a `-mode` skill is
two copies that will drift, which is what retired `ds-senior-mode`.

**Resolved: the pattern already ships in this repo, unguarded, without pain.**

`agents-md/system/agents-base.md:20`

> **Comments target humans and explain WHY, not WHAT** — a non-obvious constraint,
> invariant, or workaround. Default to one line, only where the reason isn't clear
> from the code; never restate code or cite plan/ticket IDs.

`skills/ds-comment-review/SKILL.md:7`

> comments are for humans and explain **WHY, not WHAT**, in domain terms a reader with
> only the code can follow — one line by default, only where the reason isn't obvious,
> never restating the code and never pointing into planning artifacts (decision IDs,
> roadmap/step/spec references)

Same rule, same phrasing skeleton, two files, already shipped.

**Why this differs from `ds-senior-mode`:** that mode duplicated four *sibling modes*
— same altitude, same audience, same activation — so turning it on was supposed to
equal turning on the other four, and any drift made it lie. This is one rule at **two
altitudes**: the base states a compressed constraint competing for attention against
sixty other lines; the skill applies a discipline on explicit scope, with `--fix`, and
imposes the standard regardless of local habits.

**The test:** *can one substitute for the other?* Yes → composite → forbidden.
No → altitudes → fine.

A guard test is available later using the `internal/catalog/docs_test.go` pattern, but
know its ceiling: it catches a rule going **missing**, not a rule quietly reworded into
something different.

### 6.2 Bundling and commits

- **One branch, one PR, one `VERSION` bump.** Changes to `agents-md/system/` are
  spliced into every user's AGENTS.md on the next `devskills install`, so each
  separate change is a churn event on files we do not own.
- **Per-rule commits inside that branch.** Git mode requires it, and more importantly:
  `agents-base.md` governs behaviour, and we are adding rules *and* revisiting
  existing ones in one pass. An undifferentiated bundle leaves nothing to bisect if
  the result comes out worse.

### 6.3 Skill-boundary collision

`ds-deslop` and `ds-comment-review` already exist with carved, stated boundaries —
comment-review's description literally says *"This is **not** `/ds-deslop`."* The
settled rule is that each pass answers a different question and they do not overlap.

Four skills will then sit close together:

| Skill | Lens | Scope |
|---|---|---|
| `ds-deslop` | code slop | branch diff |
| `ds-comment-review` | comment purpose | any scope |
| *new: clarity* | is it understandable | any text |
| *new: humanizer-type* | does it read as machine-written | any text |

The boundary is defensible, but it must be **written into each description** the way
comment-review already does it, or the new skills quietly become supersets of the old
ones.

### 6.4 Licensing

- devskills is **MIT** (Gleicon Moraes, 2026).
- `blader/humanizer` is MIT, but its source — Wikipedia's *Signs of AI writing* — is
  **CC BY-SA** (share-alike).
- **Ideas travel; text does not.** The pattern list as concepts is free to use.
  Copying the guide's phrasing wholesale would drag share-alike obligations into an
  MIT repo.
- Same shape for ASD-STE100 (licensed) and ISO 24495-1 (paid). Only the Federal Plain
  Language Guidelines are quotable directly.

---

## 7. Candidate rules from an external AGENTS.md

From <https://x.com/MarcosHernanz/status/2083954734487212511>. Seven rules, all
architectural. Assessed against `agents-md/system/agents-base.md`.

**Take:**

- **"Do not assume a library lacks a capability without checking its documentation and
  types."** Strongest line. Names a specific, observable agent failure — reimplementing
  what the dependency already does, because it never looked. Nothing in our base covers
  it; `ds-code-quality-review` catches the *result* after the fact, not the cause.
- **"Grow the system in layers… never trade a working product for unfinished
  complexity."** Guards against the big-bang rewrite that leaves nothing running. Our
  §4 is about verification, not about always having something that works end to end.

**Take, but scope first:**

- **"Do not preserve backward compatibility."** Right instinct — agents pile up compat
  shims nobody asked for. Unconditionally wrong for anyone shipping a library, an API,
  or a CLI with users, which includes devskills itself (the retired-skills ledger
  exists precisely to handle old installs). Needs a scope condition.
- **"Prefer established, well-maintained libraries."** Pulls *against*
  `ds-code-quality-review`'s existing "unjustified dependencies" check. Both are
  defensible; shipping both unreconciled gives an agent cover for either choice.

**Reject:**

- **"Make architectural decisions for the long term. Do not accept a stopgap."**
  Contradicts **rule 2 of its own list** — "the simplest implementation that fully
  meets the *current* requirements… avoid speculative abstractions" — and contradicts
  our §2 directly.
- **"Keep components modular and concerns clearly separated."** True, and too vague to
  change any decision.
- **"Choose the simplest implementation…"** already covered, and our version is
  sharper: "no error handling for impossible scenarios", "if you write 200 lines and it
  could be 50, rewrite it".

**Gap this exposed:** `agents-base.md` currently says **nothing about dependencies**.

---

## 8. Open — decide in the build session

1. **Reformulate rule 3** (§5.3) — structural test for description, word count kept
   for instruction. Then re-run against the same corpus to confirm the fire rate drops
   without losing the genuine finds.
2. **Run a second corpus, ideally raw AI output.** Rules 6, 8 and 10 scored near zero
   on careful prose. That is one sample of one register, and it cannot settle whether
   they belong in the review skill.
3. **Name the artifacts** — the `-mode` skill and the two new skills. Convention:
   `ds-` prefix on every skill, `-mode` suffix on modes.
4. **Merge rule 9 into rule 4**, and add the "declare the distinction once" escape for
   names that collide with code identifiers (§5.4).
5. **Write the person/system carve-out into rule 5** rather than leaving it to be
   rediscovered per-run.
6. **Encode rule 1 > rule 3 precedence**, and decide whether other pairs need one.
7. **Decide on candidate rule 11** — convention consistency, distinct from term drift
   (§5.4).
8. **Specify the output contract** for the review skill: report the N worst, census the
   rest. Unbounded output buried the findings that mattered (§5.3).
9. **Decide the source-of-truth direction** between the agents.md block and the mode,
   and whether a guard test is worth its ceiling.
10. **Reconcile** "prefer established libraries" against "unjustified dependencies".
11. **Write the scope condition** for the backward-compatibility rule.
12. **Decide whether to build `evals/`.** Without it, no agents.md wording change can be
    shown to have helped. The 14-file run is a manual eval; encoding it is the
    difference between measuring and guessing.
13. **Audit the existing base** — `agents-base.md`, `concise.md`, `phase-hints.md` — for
    behaviour-related items to fold into the same branch.

---

## 9. Sources

- ISO 24495-1:2023 — <https://www.iso.org/standard/78907.html>
- International Plain Language Federation — <https://www.iplfederation.org/iso-standard/>
- Federal Plain Language Guidelines — <https://www.gsa.gov/governmentwide-initiatives/plain-language>
- ASD-STE100 — <https://www.asd-ste100.org/about_STE.html>
- `danyuchn/asd-ste100-skill` — <https://github.com/danyuchn/asd-ste100-skill>
- `ayghri/i-have-adhd` — <https://github.com/ayghri/i-have-adhd>
- `blader/humanizer` — <https://github.com/blader/humanizer>
- Wikipedia, *Signs of AI writing* — <https://en.wikipedia.org/wiki/Wikipedia:Signs_of_AI_writing>
- Double-barreled questions — <https://www.scribbr.com/methodology/double-barreled-question/>
- Grice's maxims — <https://plato.stanford.edu/entries/implicature/>
