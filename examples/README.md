# Examples

This directory contains examples of how to use the `aliaser` library and CLI.

## Library: main.go

The [`main.go`](/examples/main.go) file contains a simple example of how to use
the `aliaser` library to generate aliases for `github.com/gin-gonic/gin`
package.

```go
func main() {
	a, err := aliaser.New("gin", "github.com/gin-gonic/gin")
	if err != nil {
		log.Fatal(err)
	}
	if err := a.GenerateFile("gin/alias.go"); err != nil {
		log.Fatal(err)
	}
}
```

## CLI: aliaser-uuid.sh

The [`aliaser-uuid.sh`](/examples/aliaser-uuid.sh) file contains a simple shell
script that uses the `aliaser` CLI to generate aliases for
`github.com/google/uuid` package.

```sh
go run -mod=mod github.com/marcozac/go-aliaser/cmd/aliaser generate \
	--target="uuid" \
	--pattern="github.com/google/uuid" \
	--file="uuid/alias.go"
```
