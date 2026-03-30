# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-03-30

### Added

- **Core Engine:** Implement dynamic worker scaling logic (up to MaxWorkers and down to MinWorkers).
- **Task Queue:** Add a bounded channel-based task queue with blocking submission for backpressure.
- **Task Options:** Support for per-task retry counts, exponential backoff delays, and execution timeouts.
- **Observability:** Implement `Stats()` method for real-time monitoring of worker and queue states.
- **Safety:** Add worker-level panic recovery to prevent cascading pool failures.
- **Lifecycle:** Implement `Shutdown()` with context support for graceful task draining.
- **Configuration:** Provide functional options for pool initialization (`WithMaxWorkers`, `WithQueueSize`, etc.).
- **Documentation:** Complete Go documentation for all public symbols and a comprehensive README.
- **Examples:** Add a basic usage example demonstrating concurrency and backpressure.
- **CI/CD:** Setup GitHub Actions workflow for automated testing and linting.

[1.0.0]: https://github.com/AshvinBambhaniya/autopool/releases/tag/v1.0.0
