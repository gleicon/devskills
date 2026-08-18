# Using relay

Point relay at a target by adding it to the config file:

    relay add github localhost:8080

Each entry in the config file names a source and a local port.

## Retries

Utilization of the `--retry` flag should be undertaken when delivery
failures are experienced.

Not infrequently, a target that is not unreachable will still refuse
payloads; relay logs each refusal to stderr.

## Removing targets

`relay remove github` deletes the entry from the settings file.
