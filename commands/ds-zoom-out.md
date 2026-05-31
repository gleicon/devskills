Step up a layer of abstraction and map how the current code fits the bigger picture.

When invoked, the user is unfamiliar with this area of code and wants the broader context — not a line-by-line read.

## Process

- Identify the module or area in focus.
- Locate before reading: if `PROJECT.md` exists, start from its map; search for the callers and neighbours rather than reading files wholesale. You need boundaries and relationships, not line-by-line internals.
- Go up one layer: describe the role it plays in the system, not its internals.
- Map the relevant neighbouring modules and the callers that depend on this one.
- Name the data and control flow that crosses its boundary.
- Use the project's existing naming and vocabulary.

## Output

A concise map: the area's responsibility, its neighbours, its callers, and the boundaries between them. No code dumps.
