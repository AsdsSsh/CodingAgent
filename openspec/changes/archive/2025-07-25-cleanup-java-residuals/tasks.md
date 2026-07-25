## 1. Remove Java build and source files

- [x] 1.1 Delete `pom.xml` — Maven project definition
- [x] 1.2 Delete `src/` — all 37 Java source files
- [x] 1.3 Delete `target/` — Maven build output (`.jar`, `.class`, etc.)

## 2. Remove IDE and CI configuration

- [x] 2.1 Delete `.idea/` — IntelliJ IDEA project files
- [x] 2.2 Delete `.github/modernize/` — Java upgrade CI workflow

## 3. Remove old documentation

- [x] 3.1 Delete `docs/` — Java-era docs

## 4. Remove redundant config and stray files

- [x] 4.1 Delete `config/permissions.yaml` — merged into `default_config.yaml`
- [x] 4.2 Delete `config/default_config.yaml.bak` — refactor backup
- [x] 4.3 Delete `./-p` — stray empty directory

## 5. Add Go binary to gitignore

- [x] 5.1 Add `coding-agent.exe` to `.gitignore`

## 6. Verification

- [x] 6.1 Verify `go build ./...` still succeeds after cleanup
- [x] 6.2 Verify `go test ./...` still passes after cleanup
- [x] 6.3 Verify `config/default_config.yaml` is intact and loadable
