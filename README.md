# github.com/robertarktes/go-bazel-starter

Minimal Go monorepo with Bazel (bzlmod), rules_go, gazelle, buildifier, and rules_oci.

## Quickstart

1. Install Bazelisk (as `bazel` alias).
2. Run `make bazel.gazelle` to generate/update BUILD files.
3. Run `make bazel.build` to build everything.
4. Run `bazel run //cmd/tool` to execute the CLI (e.g., `bazel run //cmd/tool -- --url=https://example.com`).

## Commands

- `make bazel.gazelle`: Update BUILD files with gazelle.
- `make bazel.test`: Run all tests.
- `make bazel.build`: Build all targets.
- `make bazel.buildifier`: Format all Bazel files.
- `bazel coverage //...`: Generate coverage (view with `genhtml bazel-out/_coverage/_coverage_report.dat`).
- Coverage badge: Add to CI with `bazel coverage //...` and upload to a service like Codecov.

## Bazel Targets

- `//cmd/tool`: go_binary for the CLI.
- `//cmd/tool:tool_image`: oci_image for containerizing the CLI.
- `//pkg/httpx:httpx`: go_library for HTTP client.
- `//pkg/httpx:httpx_test`: go_test with coverage.
- `//pkg/retry:retry`: go_library for retry logic.
- `//pkg/retry:retry_test`: go_test.
- `//:gazelle`: Gazelle target for BUILD generation.

## Enable Remote Cache Locally

Add to `.bazelrc`:
```
build --remote_cache=http://your-cache:port
build --remote_upload_local_results=true
```