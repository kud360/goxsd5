# goxsd5 Design Document

## Overview

XSD 1.1 Schema Parser, Go Code Generator, and JSON + XML + BER Serializer/Deserializer.

### Goals

- Streaming XSD 1.1 parser using `xml.Decoder` with accurate source location tracking
- Incremental model construction with lazy reference resolution via callbacks
- Full XSD type system: built-in primitives bootstrapped from HFP facet definitions
- Correct lexical → value space transformation per spec
- Multi-schema support with namespace-aware import/include handling
- Circular import detection
- Deterministic output (no map iteration for ordered data)
- Clean public API; no internal state leaks
- Structured logging via `slog` (caller-configurable)
- Schema content resolver interface (caller provides `io.Reader` for schema URIs)
- No goroutines or concurrency

---

## Repository Layout

Single module (`github.com/kud360/goxsd5`), multiple packages. Users install
the toolchain via:
```
go get -tool github.com/kud360/goxsd5/cmd/goxsd5
```
This pins parser and codegen together. Generated code has **no runtime
dependency on `goxsd5`** — codegen emits all helper functions it needs
directly into the output package alongside the generated types. The only
possible runtime dependency is `xstype/` (only if `StrictMapper` is used).

```
goxsd5/                          # github.com/kud360/goxsd5
├── xstype/                      # Strict XSD value-space types; no parser dep
│   ├── date.go                  # xstype.Date
│   ├── time.go                  # xstype.Time
│   ├── datetime.go              # xstype.DateTime (timezone-aware)
│   ├── duration.go              # xstype.Duration (xs:duration, yearMonth, dayTime)
│   ├── gyear.go                 # xstype.GYear, GMonth, GDay, GYearMonth, GMonthDay
│   └── decimal.go               # xstype.Decimal (wraps math/big.Rat)
├── schema/                      # XSD schema model — no dep on anything else
│   ├── schema.go                # Schema, ElementDecl, ComplexType, SimpleType, …
│   ├── facets.go                # Facet kinds, FacetPipeline, value-space types
│   └── content.go               # Compositor types: Sequence, Choice, All, Group
├── builtins/                    # XSD 1.1 built-in type instances; imports schema/
│   ├── primitives.go            # BuiltinString, BuiltinDecimal, BuiltinBoolean, …
│   ├── derived.go               # BuiltinInteger, BuiltinLong, BuiltinToken, …
│   ├── facets.go                # Facet pipeline literals for each built-in
│   └── builtins.go              # AllBuiltins() — returns the full ordered slice
├── parser/
│   ├── parser.go                # Public API: Parser, Option, Parse()
│   ├── decoder.go               # xml.Decoder wrapper with line/col tracking
│   ├── resolver.go              # Deferred reference resolver + callback registry
│   ├── namespace.go             # xmlns prefix stack for attribute value resolution
│   ├── facets.go                # Lexical → value space transformation pipeline
│   └── imports.go               # Import/Include loading + circular-ref guard
├── codegen/
│   ├── codegen.go               # Orchestrator: drives all generators
│   ├── typemap.go               # TypeMapper interface + StrictMapper/LooseMapper
│   ├── gen_structs.go           # Generate Go struct declarations
│   ├── gen_json.go              # Generate MarshalJSON/UnmarshalJSON + json helpers
│   ├── gen_xml.go               # Generate MarshalXML/UnmarshalXML + xml helpers
│   └── gen_ber.go               # Generate MarshalBER/UnmarshalBER + ber helpers
├── cmd/goxsd5/
│   └── main.go                  # CLI: parse | generate | validate
└── go.mod
```

---

## Public API

### Parser

```go
package parser

// Parser is the entry point. Zero-value is not usable; construct with New().
type Parser struct { /* opaque */ }

// New constructs a Parser with the given options.
func New(opts ...Option) *Parser

// Option is a functional option for Parser.
type Option func(*parserConfig)

// WithLogger sets the slog.Logger used during parsing (debug + info levels).
// Defaults to slog.Default(). Pass slog.New(slog.DiscardHandler) to silence.
func WithLogger(l *slog.Logger) Option

// WithResolver sets the SchemaResolver used to load imported/included schemas.
// Defaults to FileResolver rooted at the directory of the initial schema.
func WithResolver(r SchemaResolver) Option

// Parse reads an XSD document from r, using location as the canonical URI
// for relative-reference resolution and error messages.
func (p *Parser) Parse(ctx context.Context, r io.Reader, location string) (*schema.Schema, error)

// ParseFile is a convenience wrapper that opens path and calls Parse.
func (p *Parser) ParseFile(ctx context.Context, path string) (*schema.Schema, error)
```

### SchemaResolver

```go
package parser

// SchemaResolver loads schema content by URI.
// base is the URI of the schema requesting the load (for relative resolution).
// Implementations must be safe for sequential (non-concurrent) use.
type SchemaResolver interface {
    Resolve(ctx context.Context, location, base string) (io.ReadCloser, error)
}

// FileResolver resolves locations relative to BaseDir (or the process working
// directory when BaseDir is empty).
type FileResolver struct {
    BaseDir string
}

func (f *FileResolver) Resolve(ctx context.Context, location, base string) (io.ReadCloser, error)
```

### Source Location

```go
package schema

// Pos records the source location of a schema component.
type Pos struct {
    URI    string // canonical schema document URI
    Line   int    // 1-based
    Column int    // 1-based, byte offset within line
}
```

---

## Parser Architecture

### Streaming Decoder (`parser/decoder.go`)

`xml.Decoder` is used in token-by-token mode. Line/column tracking is not
natively exposed by `xml.Decoder`, so we wrap the `io.Reader`:

```
io.Reader
  └── lineTracker   (tees bytes read, builds offset→line:col index on demand)
        └── xml.Decoder
```

`xml.Decoder.InputOffset()` returns the byte offset of the **start** of the
most recently returned token. After each `Token()` call we convert that offset
to line:col via the `lineTracker` index. The index is built lazily: we scan
forward through bytes already buffered rather than pre-reading the document.

```go
// internal to parser/decoder.go
type lineTracker struct {
    r       io.Reader
    buf     []byte // bytes read so far
    lineEnd []int  // lineEnd[i] = byte offset just after '\n' on line i (0-based)
}

// pos converts a byte offset to 1-based (line, col).
func (t *lineTracker) pos(offset int64) (line, col int)
```

### Three-Phase Architecture

Parsing is split into three sequential phases. This separation enables circular
imports (mutual namespace references) and `xs:override` support while keeping
each phase simple.

```
Phase 1: DISCOVER — load all schema documents, build dependency graph
Phase 2: PARSE    — parse each document into model components, register definitions
Phase 3: RESOLVE  — resolve all cross-references, apply override scoping, validate
```

#### Phase 1: Discover (`parser/imports.go`)

Starting from the root schema URI, recursively walk `xs:import`, `xs:include`,
and `xs:override` directives. Each schema document is loaded exactly once
(via the `SchemaResolver`). The output is the full set of schema documents and
the dependency graph.

This phase does **not** parse type declarations — it only scans for directives.
A lightweight pre-scan reads just the top-level `xs:schema` children looking
for `xs:import`, `xs:include`, `xs:override` elements.

```go
type schemaDoc struct {
    uri             string
    bytes           []byte
    targetNamespace string       // from xs:schema/@targetNamespace
    imports         []importEdge // xs:import directives
    includes        []importEdge // xs:include directives
    overrides       []overrideEdge // xs:override directives (target URI + decls)
}

type importManager struct {
    docs  map[string]*schemaDoc   // URI → loaded document (O(1) lookup)
    order []string                // URIs in discovery order (deterministic iteration)
}
```

Circular imports are not an error. When discovery encounters a URI that is
already in `docs`, it records the dependency edge and moves on. Every schema
in the transitive closure is loaded exactly once.

`xs:import` without `schemaLocation` is permitted — the namespace is noted as
a dependency; it must be satisfied by some schema discovered through another
path, or it becomes an error in Phase 3.

#### Phase 2: Parse (`parser/parser.go`)

Each loaded document is parsed using the streaming decoder (see above). The
token-by-token parse loop processes one `xml.Token` at a time:

```
for {
    tok = decoder.Token()
    switch tok.(type) {
    case xml.StartElement:
        nsStack.push(tok.Attr)
        currentPos = lineTracker.pos(decoder.InputOffset())
        dispatch handler for {Space, Local}
    case xml.EndElement:
        currentHandler.end()
        nsStack.pop()
        pop handler stack
    case xml.CharData:
        currentHandler.charData(data)
    }
}
```

Handlers are pushed/popped on a stack as elements open/close. Each handler
corresponds to one XSD construct (`schema`, `element`, `complexType`,
`sequence`, `restriction`, …). A handler is an interface:

```go
type handler interface {
    start(el xml.StartElement, pos schema.Pos) (handler, error) // child handler
    charData(data []byte)
    end(pos schema.Pos) error
}
```

On `end()`, each handler does two things:
1. Wires its produced component into the parent handler's in-progress model.
2. If the component is a named definition (global type, element, group, …),
   **registers it** in the global definition registry.

References encountered during parsing (`type="tns:Foo"`, `ref="bar"`, …) are
**not resolved** in this phase. Instead they are collected as `unresolvedRef`
entries:

```go
type unresolvedRef struct {
    name   schema.QName
    kind   refKind           // type, element, group, attrGroup
    pos    schema.Pos
    schema string            // URI of the schema containing this reference
    target any               // pointer into the model where the resolved def goes
}
```

All schemas are parsed in `order` (the discovery order from Phase 1). After
Phase 2, the global registry contains every definition from every schema, and
`unresolvedRefs` contains every reference that needs resolution.

**Override declarations** (the children of `xs:override`) are parsed like
normal declarations but are registered in the overriding schema's **local
override scope** rather than the global registry. See Phase 3.

#### Phase 3: Resolve (`parser/resolver.go`)

All definitions from all schemas are now registered. Resolution is a single
loop over the collected `unresolvedRefs` slice:

```go
// Global registry — all definitions from all schemas
type refRegistry struct {
    types      map[schema.QName]schema.TypeDefinition
    elements   map[schema.QName]*schema.ElementDecl
    groups     map[schema.QName]*schema.Group
    attrGroups map[schema.QName]*schema.AttributeGroup
}

// Per-schema scoped view — overrides shadow the global registry
type scopedRegistry struct {
    overrides map[schema.QName]any   // local override definitions (xs:override)
    global    *refRegistry
}

func (s *scopedRegistry) lookup(name schema.QName, kind refKind) (any, bool) {
    if def, ok := s.overrides[name]; ok {
        return def, true  // override takes precedence
    }
    return s.global.lookup(name, kind)
}
```

Resolution loop:

```go
for _, ref := range unresolvedRefs {
    scope := scopeFor(ref.schema)            // scoped registry for this schema
    def, ok := scope.lookup(ref.name, ref.kind)
    if !ok {
        errs = append(errs, unresolvedError(ref))
        continue
    }
    assign(ref.target, def)                  // wire into the model
    if err := validateRef(ref, def); err != nil {
        errs = append(errs, err)
    }
}
```

The loop is deterministic — `unresolvedRefs` is a slice in declaration order
across schemas (discovery order).

**Validation on resolution**: `validateRef` dispatches by `ref.kind` to check
consistency of the resolved definition against the reference context:
- Attribute `default` only valid with `use="optional"`
- `fixed` and `default` are mutually exclusive on the same declaration
- `nillable="true"` requires the resolved type to be a complex type or
  a simple type that is not a union
- Substitution group head must not have `final` blocking substitution
- Override definitions must be valid derivations of the original they replace

After the loop, any entries in `errs` are reported with full source position
and import-chain context.

### Namespace Prefix Tracking (`parser/namespace.go`)

`xml.Decoder` resolves namespace URIs on element and attribute **names**, but
NOT on attribute **values** (e.g., `type="xs:string"`). We maintain a prefix
map mirroring the in-scope namespace declarations:

```go
type nsStack struct {
    frames [][]nsBinding // stack of frames; each frame is a slice of bindings
}

type nsBinding struct {
    prefix string
    uri    string
}

func (ns *nsStack) push(attrs []xml.Attr) // extract xmlns:* and push new frame
func (ns *nsStack) pop()
func (ns *nsStack) resolve(prefixed string) (schema.QName, error)
// resolve("xs:string") → QName{Namespace:"http://…/XMLSchema", Local:"string"}
// resolve("myType")    → QName{Namespace: targetNamespace, Local:"myType"}
```

`push` is called on every `StartElement` (including non-XSD elements inside
`xs:annotation`), `pop` on every matching `EndElement`.

---

## XSD Regular Expressions (`parser/facets.go`)

XSD `xs:pattern` facets use a regex dialect defined in **Appendix F** of the
XSD spec (`docs/xsd-part2-primer.html`, §F "Regular Expressions"). This dialect
is **not compatible with Go's `regexp` package** (RE2). A translation layer is
required before compiling patterns with `regexp.Compile`.

### Differences from Go RE2

**1. Implicit anchoring.**
Every XSD pattern implicitly matches the entire string — no `^` or `$` needed.
Go's `regexp` is unanchored by default. The translator must wrap all patterns:
`pattern` → `\A(?:pattern)\z` before compilation.

**2. Character class subtraction.**
XSD supports `[base-[excluded]]` syntax to subtract one class from another:
```
[a-z-[aeiou]]   →  consonants (a-z minus vowels)
[\p{L}-[\p{Lu}]]  →  all letters except uppercase
```
Go RE2 has no subtraction syntax. The translator must evaluate the subtraction
and emit an explicit character class or negative lookahead equivalent.

**3. XSD-specific multi-character escapes.**
Several escape sequences are unique to XSD and have no RE2 equivalent:

| XSD escape | Meaning | Translation |
|------------|---------|-------------|
| `\i` | Initial XML name char: `Letter \| '_' \| ':'` | Expand to explicit `[\p{L}_:]` |
| `\I` | Complement of `\i` | `[^\p{L}_:]` |
| `\c` | XML name char (`NameChar`) | Expand per XML spec |
| `\C` | Complement of `\c` | Complement expansion |
| `\d` | `\p{Nd}` (Unicode decimal digits) | `\p{Nd}` — Go supports this |
| `\w` | `[^[\p{P}\p{Z}\p{C}]]` (not punct/sep/other) | Expand to Unicode property complement |
| `\W` | Complement of `\w` | `[\p{P}\p{Z}\p{C}]` |

Note: XSD `\d` is `\p{Nd}` (all Unicode decimal digits), not Go's `[0-9]`.
XSD `\w` is similarly broader than Go's `\w`.

**4. Unicode block escapes.**
`\p{IsBasicLatin}`, `\p{IsGreek}`, `\p{IsCyrillic}`, … identify Unicode blocks.
Go RE2 does not support `\p{IsXxx}`. The translator must expand these to
explicit Unicode code point ranges using the block table in §F.

**5. Wildcard `.` semantics.**
XSD `.` matches `[^\n\r]` (everything except `#xA` and `#xD`).
Go `.` matches everything except `\n` by default. The translator replaces
bare `.` with `[^\n\r]`.

**6. No `{,m}` quantifier.**
XSD does not allow `{,m}`; only `{0,m}` is valid. This is already compatible
with RE2 — no translation needed, but the parser must reject `{,m}` as invalid.

### Translation approach

The translator lives in `parser/facets.go` alongside the facet pipeline. It
is a single-pass recursive descent parser over the XSD pattern string that
emits a valid RE2 string. It does not use `regexp` internally — it operates
on bytes. The output is compiled once with `regexp.MustCompile` and stored
in the `PatternFacet`.

```go
// translateXSDPattern converts an XSD §F pattern to a Go RE2 pattern.
// Returns an error if the pattern is not valid XSD regex syntax.
func translateXSDPattern(xsd string) (re2 string, err error)
```

Block escape expansion (§F, `\p{IsXxx}`) uses a generated lookup table
`blockRanges` in `parser/xsd_blocks.go`, derived from the block table in
`docs/xsd-part2-primer.html` §F at build time or hardcoded as a package-level
`var`.

---

## xstype — Date/Time Arithmetic

The `xstype/` package must implement date/time arithmetic exactly as specified
in **Appendix E** of the XSD spec (`docs/xsd-part2-primer.html`,
§E "Adding durations to dateTimes").

This algorithm is **incompatible with Go's stdlib** in two critical ways:

**1. Day pinning, not overflow.**
When adding a duration causes a day to exceed the month's maximum, XSD clamps
it to the last valid day. Go's `time.AddDate` overflows into the next month:

```
XSD:  2000-01-31 + P1M  →  2000-02-29  (pinned to Feb max in leap year)
Go:   time.AddDate(0,1,0) →  2000-03-02  (overflows into March)
```

**2. `fQuotient` / `modulo`, not Go's `%`.**
XSD arithmetic uses floor division (`fQuotient`) and its derived modulo.
Go's `%` truncates toward zero, which gives wrong results for negative values:

```
fQuotient(-1, 3) = -1        // floor(-1/3) = floor(-0.333) = -1
(-1) % 3 in Go   =  -1       // (happens to match here…)

modulo(-1, 3)    =  2        // -1 - (-1)*3 = 2
(-1) % 3 in Go   = -1        // wrong for XSD purposes
```

Month arithmetic uses the ranged form `fQuotient(a, 1, 13)` / `modulo(a, 1, 13)`
so that months stay in [1, 12] with correct carry into years.

**Algorithm summary (Appendix E, §E.1):**

```
Given dateTime S and duration D, compute E = S + D:

1. temp   = S[month] + D[month]
   E[month] = modulo(temp, 1, 13)
   carry    = fQuotient(temp, 1, 13)

2. E[year] = S[year] + D[year] + carry
   E[zone] = S[zone]

3. Seconds → Minutes → Hours: standard carry chain using modulo(x,60), modulo(x,24)

4. Days (pinning step):
   tempDays = clamp(S[day], 1, maximumDayInMonthFor(E[year], E[month]))
   E[day]   = tempDays + D[day] + carry
   LOOP while E[day] < 1 or E[day] > max:
     adjust E[day] and carry, then propagate carry into E[month]/E[year]

maximumDayInMonthFor(Y, M) handles leap years:
  Feb → 29 if (Y % 400 == 0) or (Y % 100 != 0 and Y % 4 == 0), else 28
```

**Leap seconds**: treated as 60s overflow (not as real UTC leap second points),
so `PT1M` and `PT60S` always produce identical results when added to any dateTime.

**Applies to all date/time types**: the algorithm is used for `DateTime`, `Date`,
`GYearMonth`, `GYear`, `GMonthDay`, `GDay`, `GMonth` — the duration is added to
the first (starting) dateTime in the set that the partial type represents.

Each `xstype` date/time type must therefore implement:

```go
func (d DateTime) Add(dur Duration) DateTime   // Appendix E algorithm
func (d Date) Add(dur Duration) Date
func (d GYearMonth) Add(dur Duration) GYearMonth
// … etc.
```

Do **not** delegate to `time.Time.Add` or `time.AddDate` for any of these.

---

## Type System & Built-in Types

### Bootstrap from HFP (`builtins/`)

The XSD 1.1 spec (HFP appendix) defines primitive types and their applicable
facets. The `builtins/` package declares each built-in as an individual
package-level `var` of type `*schema.SimpleType`. `schema/` contains only
model type declarations — no built-in instances.

Go initializes `var` declarations in dependency order within a package, so
derived types can safely reference primitive vars by pointer — no `init()`,
no map, no cycles:

```go
// builtins/primitives.go
var String  = &schema.SimpleType{Name: "string",  Namespace: schema.XSDNS, Builtin: true, Facets: stringFacets}
var Boolean = &schema.SimpleType{Name: "boolean", Namespace: schema.XSDNS, Builtin: true, Facets: booleanFacets}
var Decimal = &schema.SimpleType{Name: "decimal", Namespace: schema.XSDNS, Builtin: true, Facets: decimalFacets}
var Float   = &schema.SimpleType{Name: "float",   Namespace: schema.XSDNS, Builtin: true, Facets: floatFacets}
// … one var per XSD 1.1 primitive

// builtins/derived.go
var Integer = &schema.SimpleType{Name: "integer", Namespace: schema.XSDNS, Builtin: true,
                                  BaseType: Decimal, Facets: integerFacets}
var Long    = &schema.SimpleType{Name: "long",    Namespace: schema.XSDNS, Builtin: true,
                                  BaseType: Integer, Facets: longFacets}
var Int     = &schema.SimpleType{Name: "int",     Namespace: schema.XSDNS, Builtin: true,
                                  BaseType: Long,    Facets: intFacets}
// … one var per XSD 1.1 derived built-in

// builtins/facets.go — facet pipeline literals
var stringFacets  = schema.FacetPipeline{WhiteSpace: &schema.WhiteSpaceFacet{Value: schema.Preserve}}
var decimalFacets = schema.FacetPipeline{WhiteSpace: &schema.WhiteSpaceFacet{Value: schema.Collapse, Fixed: true}}

// builtins/builtins.go
// AllBuiltins returns all XSD 1.1 built-in types in spec order.
// The parser calls this once to seed the global ref registry.
func AllBuiltins() []*schema.SimpleType {
    return []*schema.SimpleType{
        String, Boolean, Decimal, Float, Double, Duration,
        DateTime, Time, Date, GYearMonth, GYear, GMonthDay, GDay, GMonth,
        HexBinary, Base64Binary, AnyURI, QName, Notation,
        // XSD 1.1 additions:
        YearMonthDuration, DayTimeDuration, DateTimeStamp, AnyAtomicType, UntypedAtomic,
        // Derived:
        NormalizedString, Token, Language, Name, NCName, ID, IDREF, IDREFS,
        ENTITY, ENTITIES, NMTOKEN, NMTOKENS,
        Integer, NonPositiveInteger, NegativeInteger,
        Long, Int, Short, Byte,
        NonNegativeInteger, UnsignedLong, UnsignedInt, UnsignedShort, UnsignedByte,
        PositiveInteger,
    }
}
```

Callers reference built-ins directly — `builtins.String`, `builtins.Integer` —
without any registry lookup. `schema/` remains a pure model package with no
built-in instances.

**XSD 1.1 primitive types:**
`string`, `boolean`, `decimal`, `float`, `double`, `duration`,
`dateTime`, `time`, `date`, `gYearMonth`, `gYear`, `gMonthDay`, `gDay`,
`gMonth`, `hexBinary`, `base64Binary`, `anyURI`, `QName`, `NOTATION`,
`yearMonthDuration`, `dayTimeDuration`, `dateTimeStamp`, `anyAtomicType`,
`untypedAtomic`.

**XSD derived built-ins** (`integer`, `long`, `int`, `short`, `byte`,
`nonNegativeInteger`, `positiveInteger`, …, `normalizedString`, `token`,
`language`, `Name`, `NCName`, `ID`, `IDREF`, `IDREFS`, `NMTOKEN`,
`NMTOKENS`, `ENTITY`, `ENTITIES`) are bootstrapped as restrictions of their
parent types with the appropriate fixed facets applied.

### Type Inheritance

The derivation chain is represented directly in the model via typed fields:

```go
type SimpleType struct {
    // …
    Primitive  *SimpleType  // root primitive ancestor (non-nil even for builtins)
    BaseType   *SimpleType  // direct base; nil only for anySimpleType
    Derivation Derivation   // *Restriction | *List | *Union
}

type ComplexType struct {
    // …
    BaseType   TypeDefinition // *SimpleType or *ComplexType; nil = anyType
    Derivation Derivation     // *Extension | *Restriction
}
```

`EffectiveFacets()` on `SimpleType` walks the chain upward collecting the
merged facet set — child facets narrow parent facets. Invalid narrowing (e.g.,
`maxLength` in child exceeds `maxLength` of parent) is reported at parse time
with source position of the offending facet.

### Facet Application Pipeline (`schema/facets.go`)

Per XSD spec, facets apply in a fixed order. The pipeline is a **struct with
typed fields**, not a slice, so ordering is enforced by the data structure
rather than by convention:

```go
// FacetPipeline is the ordered sequence of constraining facets for a type.
// A nil or zero field means "not applicable / not constrained".
type FacetPipeline struct {
    // Stage 1 — lexical space
    WhiteSpace *WhiteSpaceFacet     // preserve | replace | collapse

    // Stage 2 — pattern (still lexical space, pre-conversion)
    Patterns []PatternFacet         // all must match

    // Stage 3 — value parser (type-specific, converts lexical → value space)
    // Not stored here; looked up from the primitive type's registry entry.

    // Stage 4 — value space constraints
    Bounds      *BoundsFacets       // min/maxInclusive, min/maxExclusive
    Digits      *DigitsFacets       // totalDigits, fractionDigits
    Length      *LengthFacets       // length, minLength, maxLength
    Enumeration []EnumerationFacet  // allowed values

    // Stage 5 — XSD 1.1 assertions (post-value-parse)
    Assertions []AssertionFacet
}
```

```go
// AssertionFacet holds the parsed xs:assert expression.
// Evaluation is delegated to a user-supplied AssertionEvaluator.
type AssertionFacet struct {
    Test        string // XPath 2.0 expression string (stored as-is)
    Annotations []*Annotation
    Pos         Pos
}

// AssertionEvaluator is called at validation time for each xs:assert facet.
// Implementations receive the value (as interface{}) and the XPath expression.
// Return nil to indicate the assertion passes, a non-nil error to indicate failure.
// The default evaluator always returns an UnsupportedError — callers that need
// assertion evaluation must supply their own via parser.WithAssertionEvaluator.
type AssertionEvaluator interface {
    Evaluate(expr string, value any, ctx AssertionContext) error
}

// AssertionContext provides the evaluation context required by XPath 2.0.
type AssertionContext struct {
    TypeDef *SimpleType
    Pos     Pos
}
```

---

## Import, Include & Override Handling

### Semantics

| Directive      | Namespace effect                                            |
|----------------|-------------------------------------------------------------|
| `xs:import`    | Brings in definitions from a **different** namespace        |
| `xs:include`   | Merges definitions into the **same** namespace              |
| `xs:override`  | (XSD 1.1) Like include, but selective declarations replaced |

All three are handled in Phase 1 (Discover). Each schema is loaded exactly
once regardless of how many times it is referenced.

### Circular Imports

Circular imports are supported. Schema A may import namespace B while schema B
imports namespace A. This is valid XSD — each schema references types from the
other.

The three-phase architecture makes this straightforward:
- **Phase 1** loads each URI once. If the URI is already in `docs`, the
  dependency edge is recorded and discovery moves on. No error, no stack.
- **Phase 2** parses all schemas. Cross-schema references are collected as
  `unresolvedRef` entries without attempting resolution.
- **Phase 3** resolves all references. Every definition from every schema is
  available; mutual references resolve naturally.

### xs:override — Scoped Registry

`xs:override` allows a schema to replace specific declarations from an
included schema. The effect is **local**: only the overriding schema sees
the replacement; other schemas that include the same target see the original.

**Phase 1** records override directives and their contents:

```go
type overrideEdge struct {
    targetURI string            // schema being overridden
    decls     []overrideDecl    // replacement declarations (raw XML)
}

type overrideDecl struct {
    kind  refKind  // type, element, group, attrGroup
    name  string   // local name of the declaration being replaced
    raw   []byte   // raw XML of the replacement declaration
}
```

**Phase 2** parses override declarations the same way as normal declarations,
but registers them in the overriding schema's **local override scope** instead
of the global registry:

```go
// Built during Phase 2 for each schema that has xs:override directives.
// Key: QName of the declaration being replaced.
overrideScopes map[string]map[schema.QName]any  // schemaURI → overrides
```

**Phase 3** uses `scopedRegistry` (see above) to resolve references. When
schema A references type `Foo` and A has an override for `Foo`, the override
is returned. When schema C references the same `Foo`, the original is returned.

**Override validation** also happens in Phase 3: each override declaration is
checked to be a valid derivation of (or identical to) the original it replaces.
If validation fails, an error is reported with positions of both the original
and the override.

### xs:import without schemaLocation

`xs:import` may omit `schemaLocation`. The namespace is recorded as a
dependency during Phase 1. In Phase 3, if the namespace is provided by some
other schema discovered through a different path, it resolves normally. If no
schema provides the namespace, it is reported as an error.

---

## Schema Model (`schema/`)

All types are exported with no unexported fields. Computed properties are
provided as methods that walk the model rather than materialising derived
state into stored fields (e.g., `EffectiveFacets()` on `SimpleType`).

```go
// schema/schema.go

type Schema struct {
    Location             string
    TargetNamespace      string
    Version              string
    ElementFormDefault   Form          // Qualified | Unqualified
    AttributeFormDefault Form
    BlockDefault         DerivationSet
    FinalDefault         DerivationSet
    Elements             []*ElementDecl       // declaration order preserved
    Types                []TypeDefinition     // *SimpleType | *ComplexType; order preserved
    Attributes           []*AttributeDecl
    AttributeGroups      []*AttributeGroup
    Groups               []*Group
    Notations            []*Notation
    Imports              []*Import
    Includes             []*Include
    Overrides            []*Override          // xs:override — applied via scoped registry
    Annotations          []*Annotation
    Pos                  Pos
}

type ElementDecl struct {
    Name              string
    Namespace         string
    TypeDef           TypeDefinition   // resolved; nil only for abstract ur-type ref
    Nillable          bool
    Abstract          bool
    Default           *string          // nil = absent; ptr to "" = present empty string
    Fixed             *string
    SubstitutionGroup *ElementDecl
    Block             DerivationSet
    Final             DerivationSet
    Annotations       []*Annotation
    Pos               Pos
}

type ComplexType struct {
    Name        string
    Namespace   string
    Abstract    bool
    Mixed       bool
    Content     ContentModel     // Sequence | Choice | All | Empty | SimpleContent
    Attributes  []*AttributeUse
    AnyAttr     *AnyAttribute
    BaseType    TypeDefinition   // nil = anyType (ur-type)
    Derivation  Derivation       // *Extension | *Restriction
    Block       DerivationSet
    Final       DerivationSet
    Annotations []*Annotation
    Pos         Pos
}

type SimpleType struct {
    Name        string
    Namespace   string
    Primitive   *SimpleType      // root primitive (non-nil for all types)
    BaseType    *SimpleType      // nil only for anySimpleType
    Derivation  Derivation       // *Restriction | *List | *Union
    Facets      FacetPipeline    // own facets only; use EffectiveFacets() for full set
    Final       DerivationSet
    Builtin     bool
    Annotations []*Annotation
    Pos         Pos
}

// EffectiveFacets walks the derivation chain and returns the merged FacetPipeline.
func (s *SimpleType) EffectiveFacets() FacetPipeline
```

```go
// schema/content.go

// ContentModel is a sealed interface; implementations are in this package only.
type ContentModel interface{ contentModel() }

type Sequence struct {
    Items     []Particle
    MinOccurs int
    MaxOccurs int  // -1 = unbounded
    Pos       Pos
}

type Choice struct {
    Items     []Particle
    MinOccurs int
    MaxOccurs int
    Pos       Pos
}

type All struct {
    Items     []*ElementParticle // xs:all only allows elements, not groups
    MinOccurs int
    MaxOccurs int
    Pos       Pos
}

type SimpleContent struct {
    BaseType  TypeDefinition
    Derivation Derivation // *Extension | *Restriction
    Pos       Pos
}

// Particle is a sealed interface; implementations: *ElementParticle, *GroupRef,
// *Sequence, *Choice, *All, *Wildcard.
type Particle interface{ particle() }

type ElementParticle struct {
    Decl      *ElementDecl
    MinOccurs int
    MaxOccurs int
    Pos       Pos
}

type GroupRef struct {
    Group     *Group
    MinOccurs int
    MaxOccurs int
    Pos       Pos
}

type Wildcard struct {
    Namespace       WildcardNamespace // Any | Other | List
    ProcessContents ProcessContents   // Strict | Lax | Skip
    MinOccurs       int
    MaxOccurs       int
    Pos             Pos
}
```

---

## Logger

The parser logs at two levels:

| Level       | Events                                                               |
|-------------|----------------------------------------------------------------------|
| `LevelDebug`| Each XML token processed, each ref registered/resolved, each facet parsed |
| `LevelInfo` | Schema file loaded, type resolved from import, xs:override warning   |

Every log call includes these structured attributes:
- `schema_uri` — canonical URI of the schema being parsed
- `line` — source line (1-based)
- `col` — source column (1-based)

```go
// Embed in the parser config; not exported directly.
// Callers use WithLogger to supply their own *slog.Logger.
// Default: slog.Default().
```

The CLI (`cmd/goxsd5`) constructs:
```go
slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelInfo, // --debug → slog.LevelDebug
}))
```

---

## Codegen (`codegen/`)

Input: `[]*schema.Schema` + `TypeMapper` + `WriteFS`
Output: written directly to an `WriteFS` — one or more packages, each with
multiple files

Codegen is the single source of truth for all serialization formats. Because
the schema is fully known at generation time, the generated marshaler/unmarshaler
code is explicit, allocation-minimal, and reflection-free.

**Codegen also generates its own helper functions.** Each format generator
emits a `*_helpers.go` file into the output package containing the low-level
primitives the generated type methods call (JSON scanner, BER TLV append, XML
element helpers, etc.). These helpers are generated once per output package, not
once per type. Generated packages have **zero runtime dependency on `goxsd5`** —
only stdlib (and `xstype/` if `StrictMapper` is used).

### WriteFS

Go's `fs.FS` is read-only. Codegen writes multiple files across multiple
package directories, so we define a minimal writable interface:

```go
// WriteFS is the write target for codegen. Paths are slash-separated and
// relative to the root (e.g. "orders/orders.go", "orders/orders_json.go").
// Implementations must create intermediate directories as needed.
type WriteFS interface {
    Create(path string) (io.WriteCloser, error)
}
```

Two implementations are provided:

```go
// DirWriteFS writes to a real directory on disk.
type DirWriteFS struct{ Root string }

func (d DirWriteFS) Create(path string) (io.WriteCloser, error) {
    full := filepath.Join(d.Root, filepath.FromSlash(path))
    if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil { return nil, err }
    return os.Create(full)
}

// MemWriteFS holds generated files in memory; useful for tests and previews.
type MemWriteFS struct {
    Files map[string][]byte // path → content; populated after Generate()
}

func (m *MemWriteFS) Create(path string) (io.WriteCloser, error)
```

### Codegen API

```go
// Generate writes all generated Go packages to out.
// schemas is the full import closure returned by the parser.
// Each schema's targetNamespace maps to one output package (see Package Layout).
func Generate(schemas []*schema.Schema, mapper TypeMapper, out WriteFS) error
```

### Type Mapping

```go
// TypeMapper maps an XSD SimpleType to a Go type string and its required imports.
// Called for every simple-type reference in the schema.
type TypeMapper interface {
    MapType(t *schema.SimpleType) (goType string, pkgImports []string)
}
```

Two built-in implementations:

**`StrictMapper`** — semantically accurate, uses `github.com/kud360/goxsd5/xstype`:

| XSD type                        | Go type              |
|---------------------------------|----------------------|
| `xs:decimal`                    | `xstype.Decimal`     |
| `xs:integer`, `xs:long`, …      | `*big.Int`           |
| `xs:int`, `xs:short`, `xs:byte` | `int32`, `int16`, `int8` |
| `xs:date`                       | `xstype.Date`        |
| `xs:time`                       | `xstype.Time`        |
| `xs:dateTime`                   | `xstype.DateTime`    |
| `xs:dateTimeStamp`              | `xstype.DateTime`    |
| `xs:duration`                   | `xstype.Duration`    |
| `xs:gYear`, `xs:gMonth`, …      | `xstype.GYear`, …    |
| `xs:boolean`                    | `bool`               |
| `xs:string` + derived           | `string`             |
| `xs:float`                      | `float32`            |
| `xs:double`                     | `float64`            |
| `xs:hexBinary`, `xs:base64Binary` | `[]byte`           |

**`LooseMapper`** — idiomatic Go stdlib types:

| XSD type                                      | Go type      |
|-----------------------------------------------|--------------|
| `xs:decimal`                                  | `float64`    |
| `xs:integer` + all derived integer types      | `int64`      |
| `xs:date`, `xs:dateTime`, `xs:dateTimeStamp`  | `time.Time`  |
| `xs:duration`, `xs:time`, `xs:gYear`, …       | `string`     |
| `xs:boolean`                                  | `bool`       |
| `xs:string` + derived                         | `string`     |
| `xs:hexBinary`, `xs:base64Binary`             | `[]byte`     |

Callers can compose mappers:

```go
type MyMapper struct{ codegen.LooseMapper }

func (m MyMapper) MapType(t *schema.SimpleType) (string, []string) {
    if t.Name == "decimal" && t.Namespace == schema.XSDNS {
        return "apd.Decimal", []string{"github.com/cockroachdb/apd/v3"}
    }
    return m.LooseMapper.MapType(t)
}
```

### Package Layout — One Package per Namespace

Each XSD `targetNamespace` maps to one generated Go package. Schemas that
share a `targetNamespace` (via `xs:include`) contribute to the same package.
Imported schemas (different `targetNamespace`) become separate packages; the
generated code imports them as needed.

**Namespace URI → package name:**

1. Take the last non-empty path segment, or last colon-delimited segment for URNs.
2. Strip common suffixes: `.xsd`, `.xml`, `.json`, `/schema`, `/types`.
3. Lowercase, replace hyphens/dots with underscores, drop other non-identifier chars.
4. If the result is a Go keyword or empty → append `_pkg`.
5. If the result still conflicts with another generated package name → append `_2`, `_3`, …

```
http://example.com/orders/v2      → orders_v2  (both segments kept for clarity)
urn:example:inventory             → inventory
http://schemas.example.org/common → common
http://www.w3.org/XML/1998/namespace → namespace_pkg  (too generic; user can override)
```

The CLI accepts `--pkg-name <ns>=<name>` overrides for any namespace.

---

### Name Resolution and Conflict Avoidance

All generated identifiers pass through a per-package **name resolver** that
enforces uniqueness and avoids reserved words. This is a codegen-only concern —
the schema model uses XSD names verbatim; only generated Go names are mangled.

#### Reserved name sets

There are two distinct reserved sets because type/field names and package names
have different casing rules.

**Type and field names** are always PascalCase (exported). Go's predeclared
identifiers (`string`, `bool`, `int`, `error`, `len`, …) are all lowercase, so
after PascalCase conversion they become `String`, `Bool`, `Int`, `Error`, `Len`
— none of which are reserved. The reserved set for type/field names is:

- **Go keywords** (syntactically forbidden as any identifier):
  `break case chan const continue default defer else fallthrough for func go
  goto if import interface map package range return select struct switch type var`
- **Generated helper function names** in the same package (known at codegen
  time per format enabled): `jsonAppendString`, `jsonNewScanner`, `berWrapTLV`,
  `berUnwrapTLV`, `berNewScanner`, `xmlEncodeString`, `xmlDecodeString`, …
- **Generated method names** that every type receives:
  `MarshalJSON`, `UnmarshalJSON`, `MarshalXML`, `UnmarshalXML`,
  `MarshalBER`, `UnmarshalBER`, `Validate`, `String`

**Package names** are lowercase (Go convention). Here predeclared identifiers
*do* matter — a package named `string` or `error` is valid Go but extremely
confusing and would shadow the predeclared identifier within its own files.
The reserved set for package names adds:

- All Go predeclared identifiers: `bool byte complex64 complex128 error
  float32 float64 int int8 int16 int32 int64 rune string uint uint8 uint16
  uint32 uint64 uintptr true false nil append cap clear close complex copy
  delete imag len make max min new panic print println real recover`

If a derived package name conflicts with either set, append `_pkg`; if still
conflicting, append `_2`, `_3`, …

#### Identifier derivation

XSD names are converted to Go identifiers before conflict checking:

1. Split on non-alphanumeric characters (hyphens, dots, underscores, spaces).
2. Title-case each segment.
3. Join. Result is PascalCase for types; camelCase for struct fields (first
   segment lowercased).

```
"postal-code"   → PostalCode  (type) / postalCode  (field)
"xs.decimal"    → XsDecimal
"MyType"        → MyType
```

#### Ancestor-based conflict resolution

When a proposed name is taken, the resolver prepends ancestor names one level
at a time until a unique name is found:

```
Ancestors collected while descending the schema tree:
  schema → complexType "Customer" → element "Address" → anonymous complexType

Proposed name for anonymous type: "Address"
  → "Address" taken (global complexType)
  → prepend "Customer"  → "CustomerAddress" → free ✓  → use it

Later, another anonymous type inside Customer also wants "Address":
  → "Address" taken
  → "CustomerAddress" taken
  → prepend "Order" (grandparent) → "OrderCustomerAddress" → free ✓
```

If the ancestor stack is exhausted without finding a unique name, append a
decimal counter: `Address2`, `Address3`, …

The resolver is a single struct per output package, populated before code
emission so that forward references are handled correctly:

```go
// internal to codegen, not exported
type nameResolver struct {
    hardReserved map[string]struct{} // keywords + helpers + method names
    softReserved map[string]struct{} // predeclared identifiers
    used         map[string]struct{} // all names already assigned in this package
}

func (r *nameResolver) assign(proposed string, ancestors []string) string
```

`assign` is deterministic: same schema in same declaration order always
produces the same names. It does not use maps for iteration — it walks the
`ancestors` slice in reverse (innermost first).

#### Named global declarations

Even globally-named XSD types and elements go through the resolver. A schema
declaring `<xs:complexType name="Error">` will get its name checked: `Error`
is soft-reserved (shadows `error`); a warning is emitted and the name is used
as-is unless it hard-conflicts, in which case the resolver tries `ErrorType`,
then appends ancestors.

---

### xs:choice — Type Selection

`xs:choice` is a discriminated union: exactly one of the listed alternatives
is present. Go has no built-in sum type, so we generate a sealed interface:

```go
// Generated for <xs:choice> inside complexType "Address"
// The interface name is resolved through the standard name resolver.
type AddressLocation interface {
    addressLocationSeal() // unexported: seals the interface to this package
}

// One concrete type per alternative element in the choice.
// Names are resolved with parent "AddressLocation" as immediate ancestor.
type AddressLocationStreet struct{ Value string }
type AddressLocationPOBox  struct{ Value int32  }

func (AddressLocationStreet) addressLocationSeal() {}
func (AddressLocationPOBox)  addressLocationSeal() {}

// Containing type holds the interface field.
type Address struct {
    Location AddressLocation // required: exactly one alternative
}
```

The seal method name is the lowercase version of the interface name + `Seal`,
e.g. `addressLocationSeal`. This name also goes through the resolver to
ensure it doesn't shadow any method the alternatives' value types might have.

**Choice naming:**
- The interface name: parent type name + choice element name if the choice
  element is named, otherwise parent type name + `"Choice"` (fixed suffix, not configurable).
- Concrete alternative names: interface name + PascalCase(element name).
- All names go through the resolver as usual.

**Marshaling a choice:**

*JSON* — each alternative marshals to an object with a single key matching the
element name; unmarshaling reads the key to select the concrete type:

```json
{ "location": { "street": "123 Main St" } }
{ "location": { "poBox": 42 } }
```

Generated `MarshalJSON` on `Address` does a type switch:
```go
switch loc := v.Location.(type) {
case AddressLocationStreet:
    buf = jsonAppendKey(buf, "street")
    buf = jsonAppendString(buf, loc.Value)
case AddressLocationPOBox:
    buf = jsonAppendKey(buf, "poBox")
    buf = jsonAppendInt32(buf, loc.Value)
}
```

*XML* — each alternative marshals to the element with its declared name;
unmarshaling peeks at the child element name to select the type.

*BER* — each alternative is tagged with its APPLICATION index (assigned by
declaration order within the choice); the decoder reads the tag to select.

**Nested choices** — a `xs:choice` inside a `xs:sequence` inside another
`xs:choice` produces nested interfaces. Each level is resolved independently
through the name resolver with the full ancestor stack.

---

### Struct Generation (`gen_structs.go`)

- `complexType` → exported `struct` with fields
- `simpleType` restriction of string with enumerations → `type Foo string` + `const` block
- `simpleType` with value-space constraints → named type + `Validate() error`
- Element `maxOccurs > 1` or unbounded → slice field (`[]T`)
- Optional element (`minOccurs=0`) without default → pointer field (`*T`)
- Optional element with `default` value → value field; default applied in `UnmarshalXXX`
- `nillable="true"` → pointer field
- Attributes → fields; required attrs have no `omitempty`; optional attrs do

### JSON Marshaler Generation (`gen_json.go`)

Generated `MarshalJSON` / `UnmarshalJSON` are explicit and reflection-free.

**Marshaling** — appends directly to a `[]byte`, no intermediate structures:

```go
// Generated for: <xs:complexType name="Address">
func (v *Address) MarshalJSON() ([]byte, error) {
    var b [128]byte          // initial stack buffer; heap-escapes only if needed
    buf := appendAddressJSON(b[:0], v)
    return buf, nil
}

func appendAddressJSON(buf []byte, v *Address) []byte {
    buf = append(buf, '{')
    buf = append(buf, `"street":`...)
    buf = jsonAppendString(buf, v.Street) // generated helper in *_json_helpers.go
    buf = append(buf, `,"city":`...)
    buf = jsonAppendString(buf, v.City)
    if v.Zip != nil {
        buf = append(buf, `,"zip":`...)
        buf = jsonAppendString(buf, *v.Zip)
    }
    buf = append(buf, '}')
    return buf
}
```

**Unmarshaling** — generated field-dispatch switch over known keys; no map,
no reflection. Uses a generated scanner (emitted into `*_json_helpers.go`):

```go
func (v *Address) UnmarshalJSON(data []byte) error {
    s := jsonNewScanner(data)
    if err := s.expectObject(); err != nil { return err }
    for s.nextKey() {
        switch s.key() {
        case "street":
            v.Street = s.string()
        case "city":
            v.City = s.string()
        case "zip":
            z := s.string(); v.Zip = &z
        default:
            s.skip()
        }
        if s.err() != nil { return s.err() }
    }
    return s.err()
}
```

`gen_json.go` emits `*_json_helpers.go` into the output package containing
`jsonAppendString`, `jsonAppendInt`, `jsonAppendBool`, `jsonAppendFloat`,
`jsonScanner` — all operating on `[]byte`, no `interface{}` values. Generated
once per output package, not once per type.

### XML Marshaler Generation (`gen_xml.go`)

Generated code implements `xml.Marshaler` / `xml.Unmarshaler` using
`xml.Encoder` / `xml.Decoder` (stdlib streaming APIs). No reflection.
Field-by-field writes with explicit element names from the schema:

```go
func (v *Address) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
    if err := e.EncodeToken(start); err != nil { return err }
    if err := xmlEncodeString(e, "street", v.Street); err != nil { return err } // generated helper
    if err := xmlEncodeString(e, "city",   v.City);   err != nil { return err }
    if v.Zip != nil {
        if err := xmlEncodeString(e, "zip", *v.Zip); err != nil { return err }
    }
    return e.EncodeToken(start.End())
}

func (v *Address) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
    for {
        tok, err := d.Token()
        if err != nil { return err }
        switch el := tok.(type) {
        case xml.StartElement:
            switch el.Name.Local {
            case "street": v.Street, err = xmlDecodeString(d, el) // generated helper
            case "city":   v.City,   err = xmlDecodeString(d, el)
            case "zip":    z, e2 := xmlDecodeString(d, el); v.Zip = &z; err = e2
            default:       err = d.Skip()
            }
        case xml.EndElement:
            return nil
        }
        if err != nil { return err }
    }
}
```

### BER Marshaler Generation (`gen_ber.go`)

Generated `MarshalBER` / `UnmarshalBER` use ASN.1 BER TLV encoding with no
reflection. Because encoded lengths of nested structures are not known until
the content is encoded, we use a **two-buffer approach**:

1. Encode content into a scratch `[]byte`.
2. Prepend the TLV header (tag + length) into the final buffer.
3. Copy content in. For fixed-size leaf types (bool, int32, etc.) the length
   is known at codegen time and step 1 is skipped.

```go
// XSD → BER universal tag assignments (generated as constants per type file)
const (
    berTagBoolean         = 0x01
    berTagInteger         = 0x02
    berTagOctetString     = 0x04
    berTagUTF8String      = 0x0C
    berTagSequence        = 0x30 // constructed
    berTagGeneralizedTime = 0x18
    berTagReal            = 0x09
    // Application-class tag for named complex types — assigned by codegen
    // using a stable index derived from declaration order in the schema.
    berTagAddress         = 0x60 | 0 // APPLICATION 0, constructed
)

func (v *Address) MarshalBER() ([]byte, error) {
    var content []byte
    content = berAppendUTF8String(content, v.Street) // generated helper
    content = berAppendUTF8String(content, v.City)
    if v.Zip != nil {
        content = berAppendUTF8String(content, *v.Zip)
    }
    return berWrapTLV(berTagAddress, content), nil
}

func (v *Address) UnmarshalBER(data []byte) error {
    content, err := berUnwrapTLV(berTagAddress, data) // generated helper
    if err != nil { return err }
    s := berNewScanner(content)
    v.Street, err = s.utf8String(berTagUTF8String); if err != nil { return err }
    v.City,   err = s.utf8String(berTagUTF8String); if err != nil { return err }
    if s.remaining() > 0 {
        z, err := s.utf8String(berTagUTF8String); if err != nil { return err }
        v.Zip = &z
    }
    return s.end()
}
```

`gen_ber.go` emits `*_ber_helpers.go` into the output package containing
`berAppendUTF8String`, `berAppendInteger`, `berAppendBoolean`, `berAppendReal`,
`berWrapTLV`, `berUnwrapTLV`, `berScanner` — all operating on `[]byte`.
Generated once per output package.

**XSD → BER tag mapping** (used by `gen_ber.go` to emit constants):

| XSD primitive                  | BER universal tag              |
|--------------------------------|--------------------------------|
| `xs:boolean`                   | 0x01 BOOLEAN                   |
| `xs:integer`, `xs:long`, …     | 0x02 INTEGER (two's complement)|
| `xs:hexBinary`, `xs:base64Binary` | 0x04 OCTET STRING           |
| `xs:string`, `xs:token`, …     | 0x0C UTF8String                |
| `xs:float`, `xs:double`        | 0x09 REAL (ISO 6093 NR3)       |
| `xs:date`, `xs:dateTime`, …    | 0x18 GeneralizedTime           |
| `complexType` (sequence)       | 0x30 SEQUENCE, constructed     |
| `complexType` (choice)         | 0x31 SET, constructed          |
| named complex types            | APPLICATION class, constructed; index = schema declaration order |
| `simpleType` restriction       | inherits from primitive        |

---

## Decisions Log

| # | Decision | Choice |
|---|----------|--------|
| 1 | Module structure | Single module `github.com/kud360/goxsd5`; `xstype/`, `schema/` are separate packages with no parser/codegen dep |
| 2 | `xstype` location | `goxsd5/xstype/` — same module, importable standalone |
| 3 | `xstype.Decimal` | Wraps `math/big.Rat`; convenience higher-level decimal types deferred to a future separate package |
| 4 | Parser type boundary | Parser always validates via strict facet pipeline; `TypeMapper` is codegen-only |
| 5 | Parser architecture | Three-phase: Discover (load all schemas) → Parse (register definitions, collect refs) → Resolve (direct lookup, validate) |
| 6 | Circular imports | Supported; each URI loaded once; mutual namespace references resolve in Phase 3 |
| 7 | `xs:override` | Supported via scoped registry: override definitions shadow originals locally to the overriding schema only |
| 8 | Reference resolution | Direct lookup in Phase 3 (single loop over collected `unresolvedRef` slice); no callbacks |
| 9 | Assertion evaluation | Stub: `AssertionEvaluator` interface; default returns `UnsupportedError`; `WithAssertionEvaluator` option |
| 10 | Marshaler strategy | Codegen generates all JSON/XML/BER methods + their helper functions; no reflection; zero runtime dep on `goxsd5` |
| 11 | Generated file layout | One file per format per schema: `foo_json.go`, `foo_xml.go`, `foo_ber.go`; helpers in `foo_json_helpers.go` etc. |
| 19 | Codegen output | `WriteFS` interface (`Create(path) (io.WriteCloser, error)`); `DirWriteFS` for disk, `MemWriteFS` for tests/preview |
| 12 | Package layout | One Go package per XSD `targetNamespace`; namespace URI → package name via path segment extraction + sanitisation |
| 13 | `xs:choice` representation | Sealed interface + one concrete struct per alternative; unexported marker method; fixed `"Choice"` suffix for unnamed choices |
| 14 | Name conflict resolution | Prepend ancestors innermost-first until unique; two reserved sets: (a) type/field names: keywords + generated helpers + method names; (b) package names: same + predeclared identifiers; numeric suffix as last resort |
| 15 | Allocation strategy | Append-to-slice (`[]byte`), stack-allocated initial buffers; BER uses two-buffer (content then TLV header) |
| 16 | BER APPLICATION tag | Assigned by schema declaration order; future plugin interface for user overrides |
| 17 | Concurrency | None; parser is strictly sequential |
| 18 | Ordering guarantees | Slices for declaration-order data; maps only for O(1) name lookup in ref registry |