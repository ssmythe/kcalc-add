# kcalc-add

A simple microservice written in Go that performs addition.  
Designed as the first component of a distributed calculator (`kcalc`).

---

## Prerequisites

- [Go](https://go.dev/) **>= 1.22**
- [Make](https://www.gnu.org/software/make/) (standard on Linux/macOS; Windows users can use WSL or Git Bash)
- Optional developer tools:
  - [golangci-lint](https://golangci-lint.run/) for linting
  - [gh](https://cli.github.com/) for GitHub integration
  - `curl` for manual HTTP testing

---

## Building

To build the service binary with version metadata:

```bash
make build
````

The compiled binary is written to `bin/kcalc-add`.
Check version info:

```bash
./bin/kcalc-add --version
```

---

## Running

Run the service with:

```bash
make run
```

* Default port: **8080**
* Override port:

  ```bash
  PORT=18080 make run
  ```

Stop the running service:

```bash
make stop
```

Logs are written to `.run_server.log`. The PID is tracked in `.run_server.pid`.

---

## Usage

Test the `/add` endpoint:

```bash
curl -s -X POST http://localhost:8080/add \
  -H 'Content-Type: application/json' \
  -d '{"a":2,"b":3}'
# {"result":5}
```

Health check:

```bash
curl http://localhost:8080/healthz
# {"status":"ok"}
```

---

## Testing

Run all unit tests:

```bash
make test
```

Verbose mode:

```bash
make testv
```

Data race detector:

```bash
make race
```

Integration tests (spins up a live server on port 18080):

```bash
make integration-test
```

---

## Coverage

Generate coverage report (summary in terminal):

```bash
make cover
```

Generate and open HTML coverage report:

```bash
make cover-html
open coverage.html   # macOS
```

---

## Makefile

The project is driven by `Makefile` targets. List all available commands:

```bash
make help
```

Example output:

```
Targets:
  build               Build the service binary (with version info)
  build-linux         Cross-compile linux/amd64
  build-darwin        Cross-compile darwin/arm64 (Apple Silicon)
  print-version       Show resolved version metadata
  test                Run unit tests
  testv               Run unit tests (verbose)
  race                Run tests with data race detector
  cover-html          Create HTML coverage report
  run                 run service
  stop                stop service
  integration-test    Run integration tests against a live server
  integration-clean   Clean integration artifacts
  bench               Run benchmarks (if any)
  fuzz                Run fuzzing for functions named Fuzz* (example: FuzzAdd)
  fmt                 Format code
  vet                 Static analysis
  tidy                Sync go.mod/go.sum
  lint                Lint if golangci-lint is installed; otherwise print a hint
  ci                  CI bundle: tidy, fmt, vet, lint, race tests, cover, cover-html
  ci-full             CI bundle: tidy, fmt, vet, lint, race, cover, cover-html, integration tests (with cleanup)
  clean               Clean generated artifacts
  help                Show this help
```

---

## CI

Two targets exist for automation:

* `make ci` — fast checks (tidy, fmt, vet, lint, race, coverage).
* `make ci-full` — includes integration tests and cleanup.
