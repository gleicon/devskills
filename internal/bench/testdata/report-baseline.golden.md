# Bench report: ds-deslop

- Reproduce: `devskills bench ds-deslop --runs 2 --model claude-sonnet-5 --format pr-md`
- Baseline mode: skill absent on the main branch; new version only, `2222222222222222222222222222222222222222` (working tree)

## Claude Code — model `claude-sonnet-5`

### narrated-greeting

| run | new |
|---|---|
| 1 | 3/3 hits, 1 extra |
| 2 | 2/3 hits, 0 extra |
| **aggregate** | 5/6 hits, 1 extra |

<details>
<summary>narrated-greeting transcripts</summary>

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
