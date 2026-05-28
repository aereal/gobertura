[![status][ci-status-badge]][ci-status]
[![PkgGoDev][pkg-go-dev-badge]][pkg-go-dev]

# gobertura

`gobertura` converts `go test -coverprofile` output into [Cobertura XML](http://cobertura.sourceforge.net/xml/coverage-04.dtd) format, making Go coverage reports compatible with CI tools that consume Cobertura (e.g. GitLab, Jenkins, Codecov).

## Installation

```bash
go get -tool github.com/aereal/gobertura/cmd/gobertura
go tool gobertura
```

## Synopsis

Read from stdin, write to stdout:

```bash
go test -coverprofile=/dev/stdout ./... | go tool gobertura > coverage.xml
```

Use explicit file paths with `-input` / `-output` flags:

```bash
go test -coverprofile=coverage.out ./...
go tool gobertura -input coverage.out -output coverage.xml
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-input` | stdin | Input coverage profile file path |
| `-output` | stdout | Output Cobertura XML file path |

## License

See LICENSE file.

[pkg-go-dev]: https://pkg.go.dev/github.com/aereal/gobertura
[pkg-go-dev-badge]: https://pkg.go.dev/badge/github.com/aereal/gobertura.svg
[ci-status-badge]: https://github.com/aereal/gobertura/workflows/CI/badge.svg?branch=main
[ci-status]: https://github.com/aereal/gobertura/actions/workflows/CI
