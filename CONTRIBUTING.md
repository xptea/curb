# Contributing to Curb

Start with the [development guide](docs/development.md) and [architecture](docs/architecture.md). Curb aims to keep its interface small and predictable; discuss substantial feature changes in an issue before implementing them.

## Changes

Keep download logic independent of MyGo, persistence in the settings package, and reusable visual controls in the UI components package. Reuse the shared style tokens. Follow the surrounding Go code and run `make fmt`.

Tests belong beside the affected package. Add regression coverage for a behavior fix and run `make check`. Download-engine changes should also pass `make test-integration`; these tests use local media and do not require a third-party service.

For visible changes, include a screenshot in the pull request and check both themes, settings, narrow windows, maximize/fullscreen restore, and resizing. Describe the user-facing behavior and the validation you performed.

## Repository hygiene

Do not add generated executables, upstream tool archives, personal download URLs, credentials, or local preferences. `build/`, `dist/`, and `tools/` are ignored intentionally. Keep changes focused and update documentation when commands, package boundaries, or behavior change.

## License

Contributions to Curb's own code are submitted under the repository's MIT license. Preserve third-party copyright and license notices when changing dependencies or packaging.
