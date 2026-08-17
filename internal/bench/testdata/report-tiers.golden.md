# Bench report: ds-project-checkpoint

- Reproduce: `devskills bench ds-project-checkpoint --runs 2 --model claude-sonnet-5 --format pr-md`
- Versions: old `1111111111111111111111111111111111111111` (main branch), new `2222222222222222222222222222222222222222` (working tree)

## Claude Code — model `claude-sonnet-5`

### fresh-checkpoint

| run | old | new |
|---|---|---|
| 1 | 3/4 elements | 4/4 elements |
| 2 | 4/4 elements | 4/4 elements |
| **aggregate** | 7/8 elements | 8/8 elements |

<details>
<summary>fresh-checkpoint transcripts</summary>

#### old run 1

stdout:

````
wrote state.md
````

#### old run 2

stdout:

````
wrote state.md
````

#### new run 1

stdout:

````
wrote state.md
````

#### new run 2

stdout:

````
wrote state.md
````

</details>

### invoke

| run | old | new |
|---|---|---|
| 1 | ok | ok |
| 2 | failed | no output |
| **aggregate** | 1/2 ok | 1/2 ok |

<details>
<summary>invoke transcripts</summary>

#### old run 1

stdout:

````
Git mode active.
````

#### old run 2

run failed: exit status 1

stderr:

````
boom
````

#### new run 1

stdout:

````
Git mode active.
````

#### new run 2

</details>
