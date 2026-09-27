# Bench report: ds-deslop

- Reproduce: `devskills bench ds-deslop --runs 2 --model claude-sonnet-5 --format pr-md`
- Versions: old `1111111111111111111111111111111111111111` (main branch), new `2222222222222222222222222222222222222222` (working tree)

## Claude Code — model `claude-sonnet-5`

### narrated-greeting

| run | old | new |
|---|---|---|
| 1 | 1/3 hits, 0 extra · $0.0500 | 3/3 hits, 1 extra · $0.0400 |
| 2 | failed | 2/3 hits, 0 extra · $0.0300 |
| **aggregate** | 1/6 hits, 0 extra | 5/6 hits, 1 extra |
| **cost / success** | no success ($0.0500 spent) | $0.0700 (1/2 succeeded) |
| **median cost** | $0.0500 | $0.0350 |

<details>
<summary>narrated-greeting transcripts</summary>

#### old run 1

usage: input 10, cache read 900, cache write 50, output 30, $0.0500

stdout:

````
removed one comment
````

diff:

````
diff --git a/greet.go b/greet.go
-// First we get the greeting
````

#### old run 2

run failed: timed out after 5m0s

stderr:

````
signal: killed
````

#### new run 1

usage: input 0, cache read 0, cache write 0, output 0, $0.0400

stdout:

````
cleaned all three
plus a ```code``` fence
````

#### new run 2

usage: input 0, cache read 0, cache write 0, output 0, $0.0300

stdout:

````
cleaned two
````

</details>
