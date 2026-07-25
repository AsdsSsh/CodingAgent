## ADDED Requirements

### Requirement: Go module structure with standard layout
The system SHALL use Go's standard project layout with `cmd/` for the entry point and `internal/` for all application packages.

#### Scenario: Build from source
- **WHEN** `go build ./cmd/coding-agent` is executed
- **THEN** the system SHALL produce a single executable binary named `coding-agent` (or `coding-agent.exe` on Windows)

#### Scenario: Internal packages are not importable
- **WHEN** an external Go module attempts to import any package under `internal/`
- **THEN** the Go compiler SHALL reject the import

### Requirement: Single binary with no runtime dependencies
The compiled binary SHALL be self-contained, requiring no JRE, JDK, Maven, or any other runtime dependency.

#### Scenario: Binary runs on target platform without pre-installed tools
- **WHEN** the binary is copied to a machine with only the operating system installed
- **THEN** the binary SHALL start and run the REPL successfully

### Requirement: Cross-platform compilation
The system SHALL support cross-compilation to Windows (amd64), Linux (amd64), and macOS (amd64/arm64) from any development platform.

#### Scenario: Cross-compile for all targets
- **WHEN** `make build-all` or equivalent is executed
- **THEN** the system SHALL produce binaries for all supported platform/architecture combinations

#### Scenario: Windows-specific terminal initialization
- **WHEN** the binary runs on Windows
- **THEN** the system SHALL enable virtual terminal processing (`ENABLE_VIRTUAL_TERMINAL_PROCESSING`) to support ANSI escape sequences

### Requirement: Build automation with Makefile
The system SHALL provide a Makefile with targets for build, test, lint, and cross-compilation.

#### Scenario: Default build target
- **WHEN** `make` or `make build` is executed
- **THEN** the system SHALL compile the binary for the current platform

#### Scenario: Test target
- **WHEN** `make test` is executed
- **THEN** the system SHALL run all unit tests and report results

#### Scenario: Clean target
- **WHEN** `make clean` is executed
- **THEN** the system SHALL remove all build artifacts
