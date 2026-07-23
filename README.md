# Bearm

Bearm is a safe, native replacement for Unix `rm`. Compatibility-mode removal
moves targets to operating-system trash locations instead of permanently
deleting them.

## Status

The project is under active development. Do not alias `rm` to Bearm until the
compatibility and platform test suites pass for your operating system.

## Development

```bash
make verify
make build
./bin/bearm version
```

## Project Rules

- Code and technical documentation are written in English.
- Commits follow Conventional Commits.
- Branches use conventional English names such as `feat/linux-trash-backend`.
- Bearm has no telemetry and performs no network requests.
