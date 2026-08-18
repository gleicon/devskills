# taskr

taskr runs the tasks in `tasks.yaml`, one per line. Run `taskr <name>`
to run one task; `taskr --all` runs them all in order.

## Exit codes

taskr exits 0 when every task succeeds and 1 on the first failure.
