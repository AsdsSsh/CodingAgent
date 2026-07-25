## ADDED Requirements

### Requirement: Java source files are removed
The repository SHALL NOT contain Java source files after the Go refactor.

#### Scenario: Java source tree deleted
- **WHEN** the cleanup is applied
- **THEN** the `src/` directory SHALL no longer exist in the repository root

### Requirement: Maven build files are removed
The repository SHALL NOT contain Maven build configuration or artifacts.

#### Scenario: pom.xml deleted
- **WHEN** the cleanup is applied
- **THEN** `pom.xml` SHALL no longer exist in the repository root

#### Scenario: target directory deleted
- **WHEN** the cleanup is applied
- **THEN** the `target/` directory SHALL no longer exist in the repository root

### Requirement: IntelliJ IDEA configuration is removed
The repository SHALL NOT contain IDE-specific project configuration.

#### Scenario: .idea directory deleted
- **WHEN** the cleanup is applied
- **THEN** the `.idea/` directory SHALL no longer exist in the repository root

### Requirement: Java documentation is removed
The repository SHALL NOT contain documentation specific to the Java implementation.

#### Scenario: docs directory deleted
- **WHEN** the cleanup is applied
- **THEN** the `docs/` directory SHALL no longer exist in the repository root

### Requirement: Redundant configuration files removed
The repository SHALL NOT contain duplicate or backup configuration files from the refactor.

#### Scenario: permissions.yaml deleted
- **WHEN** the cleanup is applied
- **THEN** `config/permissions.yaml` SHALL no longer exist

#### Scenario: config backup deleted
- **WHEN** the cleanup is applied
- **THEN** `config/default_config.yaml.bak` SHALL no longer exist

### Requirement: GitHub CI for Java is removed
The repository SHALL NOT contain CI workflows for the Java project.

#### Scenario: Java upgrade workflow deleted
- **WHEN** the cleanup is applied
- **THEN** `.github/modernize/` SHALL no longer exist

### Requirement: Stray empty directories are removed
The repository SHALL NOT contain unintentionally created empty directories.

#### Scenario: -p directory deleted
- **WHEN** the cleanup is applied
- **THEN** the `-p` directory SHALL no longer exist in the repository root

### Requirement: Go build artifact is gitignored
The Go compiled binary SHALL be excluded from version control.

#### Scenario: coding-agent.exe in gitignore
- **WHEN** the cleanup is applied
- **THEN** `.gitignore` SHALL contain an entry for `coding-agent.exe`
