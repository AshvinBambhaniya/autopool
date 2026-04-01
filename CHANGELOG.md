# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-04-01

### Added (Major Core Engine Overhaul)

- **Priority-Aware Task Queue:** Replaced the standard FIFO channel with a high-performance, bounded priority queue supporting `Low`, `Normal`, `High`, and `Critical` levels.
- **Virtual Runtime Aging:** Implemented a sophisticated aging mechanism to prevent task starvation by dynamically boosting the urgency of older tasks (Static Virtual Runtime model).
- **New Examples:** Added a dedicated priority queue demonstration in `examples/priority`.

### Fixed

- **Scaling during Bursts:** Improved the `ScaleUp` logic to correctly handle rapid task submissions when workers are busy but the queue is building up.

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

[2.0.0]: https://github.com/AshvinBambhaniya/autopool/compare/v1.0.0...v2.0.0
[1.0.0]: https://github.com/AshvinBambhaniya/autopool/releases/tag/v1.0.0
