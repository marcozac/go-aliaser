# aliaser

[![Go Reference](https://pkg.go.dev/badge/github.com/marcozac/go-aliaser.svg)](https://pkg.go.dev/github.com/marcozac/go-aliaser)
[![CI](https://github.com/marcozac/go-aliaser/actions/workflows/ci.yml/badge.svg)](https://github.com/marcozac/go-aliaser/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/marcozac/go-aliaser/graph/badge.svg?token=veikSVxhpw)](https://codecov.io/gh/marcozac/go-aliaser)
[![Go Report Card](https://goreportcard.com/badge/github.com/marcozac/go-aliaser)](https://goreportcard.com/report/github.com/marcozac/go-aliaser)

`aliaser` is a Go package designed to streamline the process of generating
aliases for Go packages.

In Go projects, naming conflicts between imported packages and your own code can
lead to verbose and cumbersome code where you have to frequently alias imported
package names.

aliaser solves this problem by automatically generating Go code that provides
aliases for all the exported constants, variables, functions, and types in an
external package. This allows you to seamlessly integrate and extend external
packages without cluttering your code with manual aliases.

## Installation

You need a working Go environment.

```bash
go get github.com/marcozac/go-aliaser
```

## Usage

To use aliaser in your project, follow these steps:

1. Create an Aliaser: Start by creating a new `aliaser.Aliaser` with the target
   package name and the pattern of the package you want to alias.

```go
import "github.com/marcozac/go-aliaser"

a, err := aliaser.New(&aliaser.Config{TargetPackage: "mypkg", Pattern: "github.com/example/package"})
if err != nil {
  // ...
}
```

2. Generate Aliases: With the `Aliaser` created, you can generate the aliases
   writing them to a `io.Writer` or to directly to a file.

```go
if err := a.Generate(io.Discard); err != nil {
  // ...
}

if err := a.GenerateFile("mypkg/alias.go"); err != nil {
  // ...
}
```

## CLI

In addition to the library, `aliaser` comes with a CLI tool to simplify
generating aliases directly from the command line.

After installing aliaser CLI with `go install` (e.g.
`go install github.com/marcozac/go-aliaser/cmd/aliaser@latest`) ~~or by
downloading a release~~ (available soon), you can use the `aliaser` command to
generate aliases for a package.

```bash
aliaser generate \
  --from "github.com/example/package" \
  --package "myalias" \
  --file "path/to/output/file.go"
```

## Examples

For simple, but more detailed examples of how to use the `aliaser` library and
CLI, see the [examples](/examples) directory.

## Contributing

Contributions to aliaser are welcome! Whether it's reporting bugs, discussing
improvements, or contributing code, all contributions are appreciated.

## License

This project is licensed under the MIT License. See the [LICENSE](/LICENSE) file
for details.
