# goxsd5

goxsd5 is an XSD 1.1 schema parser and instance validator for Go. It ships as
a Go **library** (packages `xsd`, `parser`, `xsdvalidate`, `xpath`, `xsdwalk`,
`xsdtemporal`, `xsdregex`, `xsdedit`, `builtin/xsdtype`, ...) and a **CLI** at
`cmd/goxsd5`. Requires Go 1.26+.

## Install

```
go get github.com/kud360/goxsd5
```

## Library usage

### Parse a schema

`parser.Parse` loads an XSD 1.1 schema document (and the transitive closure of
its imports, includes, redefines, and overrides) from a file path or URL.  It
returns one `*xsd.Schema` per target namespace — the root document's namespace
is always first.

```go
schemas, err := parser.Parse("path/to/schema.xsd", nil)
if err != nil {
    // err is an *xsd.ErrorList; the schemas are still usable for inspection.
    log.Println("schema errors:", err)
}
s := schemas[0]
fmt.Println("target namespace:", s.TargetNamespace)
fmt.Println("global elements:", len(s.Elements))
```

Full runnable version: `parser/example_test.go` (`ExampleParse`).

To serve schemas from memory (useful in tests), supply an `Options.Resolver`:

```go
schemas, err := parser.Parse("root.xsd", &parser.Options{
    Resolver: myResolver, // implements parser.SchemaResolver
})
```

### Validate an instance

Build a `*xsdvalidate.Validator` once from the parsed schemas, then call
`Assess` (or the convenience wrapper `xmlsrc.Validate`) for each instance
document.

```go
v := xsdvalidate.New(schemas, nil)

res, err := xmlsrc.Validate(v, strings.NewReader(`<age>42</age>`), "doc.xml")
if err != nil {
    log.Fatal("not well-formed XML:", err)
}
if res.Valid() {
    fmt.Println("valid")
} else {
    fmt.Println("invalid:", res.Err())
    // Each error carries a cvc-* spec id — e.g. "cvc-minInclusive-valid".
    fmt.Println("error ids:", xsd.RefIDs(res.Err()))
}
```

Full runnable version: `xsdvalidate/example_test.go` (`ExampleNew`).

Errors are structured `*xsd.Error` values with a `Ref.ID` field that matches
the XSD 1.1 spec clause (e.g. `cvc-elt`, `cvc-minInclusive-valid`).

## CLI

```bash
go build -o goxsd5 ./cmd/goxsd5

# Parse a schema and print a per-namespace component summary:
goxsd5 path/to/schema.xsd

# Quiet mode — print only errors:
goxsd5 -q path/to/schema.xsd

# Validate an XML instance (exit 0 = valid, exit 1 = invalid):
goxsd5 -q -validate instance.xml schema.xsd
```

Errors render as `uri:line:col: [constraint-id] message`.  Schema errors use
`src-*` ids; instance errors use `cvc-*` ids.

For the full CLI reference and the W3C conformance suite details, see
`.claude/skills/run-goxsd5/SKILL.md`.

## Testing

```bash
go test ./...          # all unit tests (~5 s)
tools/gate.sh          # full gate: fmt, vet, lint, tests, smoke, conformance ratchets
```
