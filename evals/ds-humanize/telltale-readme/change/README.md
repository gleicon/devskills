# taskr

taskr runs the tasks in `tasks.yaml`, one per line. Run `taskr <name>`
to run one task; `taskr --all` runs them all in order.

## Exit codes

taskr exits 0 when every task succeeds and 1 on the first failure.

## 🚀 Watch Mode

Here's the updated section on watch mode — I hope this helps!

Watch mode isn't just a file watcher — it's a complete feedback loop.

Let's dive into how it works. In order to enable it, it's important to
note that you pass `--watch`; taskr then reruns a task whenever one of
its inputs changes.
