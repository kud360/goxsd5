# goxsd5 Design Document

## Overview
XSD 1.1 Schema Parser, Go Code Generator, and Serializer/Deserializer with BER support.

## Architecture

### Package Structure
```
goxsd5/
├── internal/
│   └── models/
│       └── types.go          # In-memory model types
├── parser/
│   └── parser.go             # XSD 1.1 schema parser
├── codegen/
│   └── codegen.go            # Go code generator
├── serializer/
│   ├── json.go
│   ├── xml.go
│   └── binary.go             # BER serializer
└── go.mod
```

### In-Memory Model (`internal/models/types.go`)

#### Core Types
```go
type Schema struct {
    TargetNamespace string
    Elements        []*Element
    ComplexTypes    []*ComplexType
    SimpleTypes     []*SimpleType
    AttributeGroups []*AttributeGroup
    Imports         []*Import
    Includes        []*Include
    Annotations    *Annotations
}
```

#### Type System
```go
type SimpleType struct {
    Name        string
    BaseType    string              // Built-in or user-defined
    Facets      []*Facet            // Constraint chain
    UnionTypes  []*SimpleType       // Union type members
    Annotations *Annotations
}

type ComplexType struct {
    Name              string
    Abstract          bool
    Mixed             bool
    Content           ContentModel        // Sequence/Choice/All
    Attributes        []*Attribute        // Direct attributes
    AttributeGroups   []string            // Referenced groups
    BaseType          string            // Extension base
    Annotations       *Annotations
}
```

#### Facet Types (HFP Appendix)
- **Numeric**: minLength, maxLength, length
- **Numeric**: minLength, maxLength, length
- **Numeric**: minExclusive, minInclusive, maxExclusive, maxInclusive
- **Restrictions**: pattern, enumeration
- **Numeric**: minExcl
- **Numeric**: fractionDigits, totalDigits
- **Facets**: whitespace
- **Other**: anyType, anyURI, base64Binary, date, dateTime, decimal, string, etc.

## Parsing Strategy

1. **Lexing**: Tokenize XSD XML
2. **Schema Resolution**: Handle imports/includes
3. **Type Resolution**: Resolve references (ref=) to definitions
4. **Facet Application**: Apply facets in order (lexical preprocessing → value derivation)
5. **Inheritance**: Support derivation (extension/restriction) chain

## Code Generation

Generate Go types from resolved schema:
- simpleType → Go primitive type or wrapper
- complexType → struct with JSON tags
- elements → exported variables or functions

## Serialization

### JSON/JSON
- Standard encoding/json integration
- Custom marshaling for XSD types

### BER (Binary Encoding Rules)
- ASN.1-style binary format
- Tag-Length-Value encoding
- Support for complex structures

## Testing
- Schema validation tests
- Codegen round-trip tests
- JSON/XML/BER serialization tests

## References
- docs/xsd-part1-primer.html
- docs/xsd-part2-primer.html
