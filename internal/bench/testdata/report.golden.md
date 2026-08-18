# Bench report: ds-deslop

- Reproduce: `devskills bench ds-deslop --runs 2 --model claude-sonnet-5 --format pr-md`
- Versions: old `1111111111111111111111111111111111111111` (main branch), new `2222222222222222222222222222222222222222` (working tree)

## Claude Code — model `claude-sonnet-5`

### narrated-greeting

| run | old | new |
|---|---|---|
| 1 | 1/3 hits, 0 extra | 3/3 hits, 1 extra |
| 2 | failed | 2/3 hits, 0 extra |
| **aggregate** | 1/6 hits, 0 extra | 5/6 hits, 1 extra |

<details>
<summary>narrated-greeting transcripts</summary>

#### old run 1

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

stdout:

````
cleaned all three
plus a ```code``` fence
````

#### new run 2

stdout:

````
cleaned two
````

</details>
