# 🧐 StructInit

Go Linter to check struct's field initialization order.

## ⬇️ Getting Started

### Run it as a standalone linter

To install it, run:

```bash
go install github.com/manuelarte/structinit@latest
```

And then use it with:

```bash
structinit [--fix] ./...
```

- fix: `true|false` (default `false`) Fix the field instantiation order.

### Run it as a module plugin in golangci-lint

You can integrate this linter with [golangci-lint](https://golangci-lint.run/)
by using the [module plugin](https://golangci-lint.run/docs/plugins/module-plugins/).

Example of a `custom-gcl.yml` file that includes this linter:

```yaml
version: v2.12.2
plugins:
  - module: "github.com/manuelarte/structinit"
    import: "github.com/manuelarte/structinit/plugin"
    version: latest
```

## 🚀 Feature

By marking a struct with `//go:structinit`, that indicates that whenever you initialize that struct, the fields
need to be initialized in the same order as they are declared in the struct.

```go
//go:structinit
type MyStruct struct {
    Field1 string
    Field2 int

func NewMyStruct() *MyStruct {
    return &MyStruct{
        Field1: "value1",
        Field2: 42,
    }
}

func OtherMyStruct() *MyStruct {
    return &MyStruct{ // Linter will complain about this, since it's expecting Field1 to be initialized first.
        Field2: 42,
        Field1: "value1",
    }
}
```

## ❓ FAQ

### Why not lint every instantiation?

There are several reasons why it is not "easy" to lint every instantiation of a struct.
When instantiating a struct not using "static" values, there could be some side effects.
Check [example.go](./analyzer/testdata/src/side-effects/example.go).
