# W3C XML Schema Definition Language (XSD) 1.1 Part 2: Datatypes

## <a id="abstract"></a>Abstract

 *XML Schema: Datatypes* is part 2 of the specification of the XML Schema language. It defines facilities for defining datatypes to be used in XML Schemas as well as other XML specifications. The datatype language, which is itself represented in XML, provides a superset of the capabilities found in XML document type definitions (DTDs) for specifying datatypes on elements and attributes.

## <a id="contents"></a>Table of Contents

1 [Introduction](#Intro)
     1.1 [Introduction to Version 1.1](#intro1.1)
     1.2 [Purpose](#purpose)
     1.3 [Dependencies on Other Specifications](#intro-relatedWork)
     1.4 [Requirements](#requirements)
     1.5 [Scope](#scope)
     1.6 [Terminology](#terminology)
     1.7 [Constraints and Contributions](#constraints-and-contributions)
 2 [Datatype System](#typesystem)
     2.1 [Datatype](#datatype)
     2.2 [Value space](#value-space)  [Identity](#identity) · [Equality](#equality) · [Order](#order)     2.3 [The Lexical Space and Lexical Mapping](#lexical-space)  [Canonical Mapping](#canonical-lexical-representation)     2.4 [Datatype Distinctions](#datatype-dichotomies)  [Atomic vs. List vs. Union Datatypes](#atomic-vs-list) · [Special vs. Primitive vs. Ordinary Datatypes](#primitive-vs-derived) · [Definition, Derivation, Restriction, and Construction](#derivation) · [Built-in vs. User-Defined Datatypes](#built-in-vs-user-derived) 3 [Built-in Datatypes and Their Definitions](#built-in-datatypes)
     3.1 [Namespace considerations](#namespaces)
     3.2 [Special Built-in Datatypes](#special-datatypes)  [anySimpleType](#anySimpleType) · [anyAtomicType](#anyAtomicType)     3.3 [Primitive Datatypes](#built-in-primitive-datatypes)  [string](#string) · [boolean](#boolean) · [decimal](#decimal) · [float](#float) · [double](#double) · [duration](#duration) · [dateTime](#dateTime) · [time](#time) · [date](#date) · [gYearMonth](#gYearMonth) · [gYear](#gYear) · [gMonthDay](#gMonthDay) · [gDay](#gDay) · [gMonth](#gMonth) · [hexBinary](#hexBinary) · [base64Binary](#base64Binary) · [anyURI](#anyURI) · [QName](#QName) · [NOTATION](#NOTATION)     3.4 [Other Built-in Datatypes](#ordinary-built-ins)  [normalizedString](#normalizedString) · [token](#token) · [language](#language) · [NMTOKEN](#NMTOKEN) · [NMTOKENS](#NMTOKENS) · [Name](#Name) · [NCName](#NCName) · [ID](#ID) · [IDREF](#IDREF) · [IDREFS](#IDREFS) · [ENTITY](#ENTITY) · [ENTITIES](#ENTITIES) · [integer](#integer) · [nonPositiveInteger](#nonPositiveInteger) · [negativeInteger](#negativeInteger) · [long](#long) · [int](#int) · [short](#short) · [byte](#byte) · [nonNegativeInteger](#nonNegativeInteger) · [unsignedLong](#unsignedLong) · [unsignedInt](#unsignedInt) · [unsignedShort](#unsignedShort) · [unsignedByte](#unsignedByte) · [positiveInteger](#positiveInteger) · [yearMonthDuration](#yearMonthDuration) · [dayTimeDuration](#dayTimeDuration) · [dateTimeStamp](#dateTimeStamp) 4 [Datatype components](#datatype-components)
     4.1 [Simple Type Definition](#rf-defn)  [The Simple Type Definition Schema Component](#dc-defn) · [XML Representation of Simple Type Definition Schema Components](#xr-defn) · [Constraints on XML Representation of Simple Type Definition](#defn-rep-constr) · [Simple Type Definition Validation Rules](#defn-validation-rules) · [Constraints on Simple Type Definition Schema Components](#defn-coss) · [Built-in Simple Type Definitions](#builtin-stds)     4.2 [Fundamental Facets](#rf-fund-facets)  [ordered](#rf-ordered) · [bounded](#rf-bounded) · [cardinality](#rf-cardinality) · [numeric](#rf-numeric)     4.3 [Constraining Facets](#rf-facets)  [length](#rf-length) · [minLength](#rf-minLength) · [maxLength](#rf-maxLength) · [pattern](#rf-pattern) · [enumeration](#rf-enumeration) · [whiteSpace](#rf-whiteSpace) · [maxInclusive](#rf-maxInclusive) · [maxExclusive](#rf-maxExclusive) · [minExclusive](#rf-minExclusive) · [minInclusive](#rf-minInclusive) · [totalDigits](#rf-totalDigits) · [fractionDigits](#rf-fractionDigits) · [Assertions](#rf-assertions) · [explicitTimezone](#rf-explicitTimezone) 5 [Conformance](#conformance)
     5.1 [Host Languages](#hostlangs)
     5.2 [Independent implementations](#independent-impl)
     5.3 [Conformance of data](#data-conformance)
     5.4 [Partial Implementation of Infinite Datatypes](#partial-implementation)

### <a id="appendices"></a>Appendices

A [Schema for Schema Documents (Datatypes) (normative)](#schema)
 B [DTD for Datatype Definitions (non-normative)](#dtd-for-datatypeDefs)
 C [Illustrative XML representations for the built-in simple type definitions](#prim.nxsd)
     C.1 [Illustrative XML representations for the built-in primitive type definitions](#sec-prim-nxsd)
     C.2 [Illustrative XML representations for the built-in ordinary type definitions](#drvd.nxsd)
 D [Built-up Value Spaces](#constructedValueSpaces)
     D.1 [Numerical Values](#sec-numericalValues)  [Exact Lexical Mappings](#sec-exactmaps)     D.2 [Date/time Values](#d-t-values)  [The Seven-property Model](#theSevenPropertyModel) · [Lexical Mappings](#rf-lexicalMappings-datetime) E [Function Definitions](#ap-funcDefs)
     E.1 [Generic Number-related Functions](#sec-generic-number-functions)
     E.2 [Duration-related Definitions](#sec-duration-functions)
     E.3 [Date/time-related Definitions](#sec-dt-functions)  [Normalization of property values](#sec-normalization) · [Auxiliary Functions](#sec-aux-functions) · [Adding durations to dateTimes](#sec-dt-arith) · [Time on timeline](#sec-timeontimeline) · [Lexical mappings](#sec-dt-lexmaps) · [Canonical Mappings](#sec-dt-canmaps)     E.4 [Lexical and Canonical Mappings for Other Datatypes](#sec-misc-lexmaps)  [Lexical and canonical mappings for](#sec-hexbin-lexmaps) F [Datatypes and Facets](#sec-datatypes-and-facets)
     F.1 [Fundamental Facets](#app-fundamental-facets)
 G [Regular Expressions](#regexs)
     G.1 [Regular expressions and branches](#regex-branch)
     G.2 [Pieces, atoms, quantifiers](#regex-piece)
     G.3 [Characters and metacharacters](#regex-char-metachar)
     G.4 [Character Classes](#charcter-classes)  [Character class expressions](#charclassexps) · [Character Class Escapes](#cces) H [Implementation-defined and implementation-dependent features (normative)](#idef-idep)
     H.1 [Implementation-defined features](#impl-def)
     H.2 [Implementation-dependent features](#impl-dep)
 I [Changes since version 1.0](#changes)
     I.1 [Datatypes and Facets](#sec-chdtfacets)
     I.2 [Numerical Datatypes](#sec-chnum)
     I.3 [Date/time Datatypes](#sec-chdt)
     I.4 [Other changes](#sec-chother)
 J [Glossary (non-normative)](#normative-glossary)
 K [References](#biblio)
     K.1 [Normative](#normative-biblio)
     K.2 [Non-normative](#non-normative-biblio)
 L [Acknowledgements (non-normative)](#acknowledgments)

## <a id="Intro"></a>1 Introduction

### <a id="intro1.1"></a>1.1 Introduction to Version 1.1

The Working Group has two main goals for this version of W3C XML Schema:

- Significant improvements in simplicity of design and clarity of exposition *without* loss of backward *or* forward compatibility;
- Provision of support for versioning of XML languages defined using the XML Schema specification, including the XML transfer syntax for schemas itself.

These goals are slightly in tension with one another -- the following summarizes the Working Group's strategic guidelines for changes between versions 1.0 and 1.1:

1. Add support for versioning (acknowledging that this *may* be slightly disruptive to the XML transfer syntax at the margins)
2. Allow bug fixes (unless in specific cases we decide that the fix is too disruptive for a point release)
3. Allow editorial changes
4. Allow design cleanup to change behavior in edge cases
5. Allow relatively non-disruptive changes to type hierarchy (to better support current and forthcoming international standards and W3C recommendations)
6. Allow design cleanup to change component structure (changes to functionality restricted to edge cases)
7. Do not allow any significant changes in functionality
8. Do not allow any changes to XML transfer syntax except those required by version control hooks and bug fixes

The overall aim as regards compatibility is that

- All schema documents conformant to version 1.0 of this specification should also conform to version 1.1, and should have the same validation behavior across 1.0 and 1.1 implementations (except possibly in edge cases and in the details of the resulting PSVI);
- The vast majority of schema documents conformant to version 1.1 of this specification should also conform to version 1.0, leaving aside any incompatibilities arising from support for versioning, and when they are conformant to version 1.0 (or are made conformant by the removal of versioning information), should have the same validation behavior across 1.0 and 1.1 implementations (again except possibly in edge cases and in the details of the resulting PSVI);

### <a id="purpose"></a>1.2 Purpose

 The [[XML]](#XML) specification defines limited facilities for applying datatypes to document content in that documents may contain or refer to DTDs that assign types to elements and attributes. However, document authors, including authors of traditional *documents* and those transporting *data* in XML, often require a higher degree of type checking to ensure robustness in document understanding and data interchange.

 The table below offers two typical examples of XML instances in which datatypes are implicit: the instance on the left represents a billing invoice, the instance on the right a memo or perhaps an email message in XML.

<table><thead><tr><th>Data oriented</th><th>Document oriented</th></tr></thead><tbody><tr><td>
<pre><invoice>
  <orderDate>1999-01-21</orderDate>
  <shipDate>1999-01-25</shipDate>
  <billingAddress>
   <name>Ashok Malhotra</name>
   <street>123 Microsoft Ave.</street>
   <city>Hawthorne</city>
   <state>NY</state>
   <zip>10532-0000</zip>
  </billingAddress>
  <voice>555-1234</voice>
  <fax>555-4321</fax>
</invoice></pre>
</td><td>
<pre><memo importance='high'
      date='1999-03-23'>
  <from>Paul V. Biron</from>
  <to>Ashok Malhotra</to>
  <subject>Latest draft</subject>
  <body>
    We need to discuss the latest
    draft <emph>immediately</emph>.
    Either email me at <email>
    mailto:paul.v.biron@kp.org</email>
    or call <phone>555-9876</phone>
  </body>
</memo></pre>
</td></tr></tbody></table>

 The invoice contains several dates and telephone numbers, the postal abbreviation for a state (which comes from an enumerated list of sanctioned values), and a ZIP code (which takes a definable regular form).  The memo contains many of the same types of information: a date, telephone number, email address and an "importance" value (from an enumerated list, such as "low", "medium" or "high").  Applications which process invoices and memos need to raise exceptions if something that was supposed to be a date or telephone number does not conform to the rules for valid dates or telephone numbers.

 In both cases, validity constraints exist on the content of the instances that are not expressible in XML DTDs.  The limited datatyping facilities in XML have prevented validating XML processors from supplying the rigorous type checking required in these situations.  The result has been that individual applications writers have had to implement type checking in an ad hoc manner.  This specification addresses the need of both document authors and applications writers for a robust, extensible datatype system for XML which could be incorporated into XML processors.  As discussed below, these datatypes could be used in other XML-related standards as well.

### <a id="intro-relatedWork"></a>1.3 Dependencies on Other Specifications

Other specifications on which this one depends are listed in [References (§K)](#biblio).

This specification defines some datatypes which depend on definitions in [[XML]](#XML) and [[Namespaces in XML]](#XMLNS); those definitions, and therefore the datatypes based on them, vary between version 1.0 ([[XML 1.0]](#XML1.0), [[Namespaces in XML 1.0]](#XMLNS1.0)) and version 1.1 ([[XML]](#XML), [[Namespaces in XML]](#XMLNS)) of those specifications. In any given use of this specification, the choice of the 1.0 or the 1.1 definition of those datatypes is [implementation-defined](#key-impl-def).

 Conforming implementations of this specification may provide either the 1.1-based datatypes or the 1.0-based datatypes, or both. If both are supported, the choice of which datatypes to use in a particular assessment episode **should** be under user control.

**Note:** When this specification is used to check the datatype validity of XML input, implementations **may** provide the heuristic of using the 1.1 datatypes if the input is labeled as XML 1.1, and using the 1.0 datatypes if the input is labeled 1.0, but this heuristic **should** be subject to override by users, to support cases where users wish to accept XML 1.1 input but validate it using the 1.0 datatypes, or accept XML 1.0 input and validate it using the 1.1 datatypes.

This specification makes use of the EBNF notation used in the [[XML]](#XML) specification. Note that some constructs of the EBNF notation used here resemble the regular-expression syntax defined in this specification ([Regular Expressions (§G)](#regexs)), but that they are not identical: there are differences. For a fuller description of the EBNF notation, see [Section 6. Notation](https://www.w3.org/TR/xml11/#sec-notation) of the [[XML]](#XML) specification.

### <a id="requirements"></a>1.4 Requirements

 The [[XML Schema Requirements]](#schema-requirements) document spells out concrete requirements to be fulfilled by this specification, which state that the XML Schema Language must:

1.  provide for primitive data typing, including byte, date, integer, sequence, SQL and Java primitive datatypes, etc.;
2.  define a type system that is adequate for import/export from database systems (e.g., relational, object, OLAP);
3.  distinguish requirements relating to lexical data representation vs. those governing an underlying information set;
4.  allow creation of user-defined datatypes, such as datatypes that are derived from existing datatypes and which may constrain certain of its properties (e.g., range, precision, length, format).

### <a id="scope"></a>1.5 Scope

 This specification defines datatypes that can be used in an XML Schema.  These datatypes can be specified for element content that would be specified as [#PCDATA](https://www.w3.org/TR/xml11/#dt-chardata) and attribute values of [various types](https://www.w3.org/TR/xml11/#sec-attribute-types) in a DTD.  It is the intention of this specification that it be usable outside of the context of XML Schemas for a wide range of other XML-related activities such as [[XSL]](#XSL) and [[RDF Schema]](#RDFSchema).

### <a id="terminology"></a>1.6 Terminology

 The terminology used to describe XML Schema Datatypes is defined in the body of this specification. The terms defined in the following list are used in building those definitions and in describing the actions of a datatype processor:

 <a id="dt-compatibility">[Definition:]</a>for compatibility  A feature of this specification included solely to ensure that schemas which use this feature remain compatible with [[XML]](#XML).  <a id="dt-match">[Definition:]</a>**match**  *(Of strings or names:)* Two strings or names being compared must be identical. Characters with multiple possible representations in ISO/IEC 10646 (e.g. characters with both precomposed and base+diacritic forms) match only if they have the same representation in both strings. No case folding is performed. *(Of strings and rules in the grammar:)* A string matches a grammatical production if and only if it belongs to the language generated by that production.  <a id="dt-may">[Definition:]</a>**may**  Schemas, schema documents, and processors are permitted to but need not behave as described. <a id="dt-should">[Definition:]</a>**should**It is recommended that schemas, schema documents, and processors behave as described, but there can be valid reasons for them not to; it is important that the full implications be understood and carefully weighed before adopting behavior at variance with the recommendation. <a id="dt-must">[Definition:]</a>**must**  *(Of schemas and schema documents:)* Schemas and documents are required to behave as described; otherwise they are in [error](#dt-error). *(Of processors:)* Processors are required to behave as described. <a id="dt-mustnot">[Definition:]</a>**must not**Schemas, schema documents and processors are forbidden to behave as described; schemas and documents which nevertheless do so are in [error](#dt-error). <a id="dt-error">[Definition:]</a>**error** A failure of a schema or schema document to conform to the rules of this specification. Except as otherwise specified, processors **must** distinguish error-free (conforming) schemas and schema documents from those with errors; if a schema used in type-validation or a schema document used in constructing a schema is in error, processors **must** report the fact; if more than one is in error, it is [implementation-dependent](#key-impl-dep) whether more than one is reported as being in error. If more than one of the constraints given in this specification is violated, it is [implementation-dependent](#key-impl-dep) how many of the violations, and which, are reported. **Note:** Failure of an XML element or attribute to be datatype-valid against a particular datatype in a particular schema is not in itself a failure to conform to this specification and thus, for purposes of this specification, not an error.  <a id="dt-useroption">[Definition:]</a>**user option** A choice left under the control of the user of a processor, rather than being fixed for all users or uses of the processor. Statements in this specification that "Processors **may** at user option" behave in a certain way mean that processors **may** provide mechanisms to allow users (i.e. invokers of the processor) to enable or disable the behavior indicated. Processors which do not provide such user-operable controls **must not** behave in the way indicated. Processors which do provide such user-operable controls **must** make it possible for the user to disable the optional behavior. **Note:** The normal expectation is that the default setting for such options will be to disable the optional behavior in question, enabling it only when the user explicitly requests it. This is not, however, a requirement of conformance: if the processor's documentation makes clear that the user can disable the optional behavior, then invoking the processor without requesting that it be disabled can be taken as equivalent to a request that it be enabled. It is required, however, that it in fact be possible for the user to disable the optional behavior. **Note:** Nothing in this specification constrains the manner in which processors allow users to control user options. Command-line options, menu choices in a graphical user interface, environment variables, alternative call patterns in an application programming interface, and other mechanisms may all be taken as providing user options.

### <a id="constraints-and-contributions"></a>1.7 Constraints and Contributions

 This specification provides three different kinds of normative statements about schema components, their representations in XML and their contribution to the schema-validation of information items:

 <a id="dt-cos">[Definition:]</a> **Constraint on Schemas**   Constraints on the schema components themselves, i.e. conditions components [**must**](#dt-must) satisfy to be components at all. Largely to be found in [Datatype components (§4)](#datatype-components).  <a id="dt-src">[Definition:]</a> **Schema Representation Constraint**   Constraints on the representation of schema components in XML.  Some but not all of these are expressed in [Schema for Schema Documents (Datatypes) (normative) (§A)](#schema) and [DTD for Datatype Definitions (non-normative) (§B)](#dtd-for-datatypeDefs).  <a id="dt-cvc">[Definition:]</a> **Validation Rule**   Constraints expressed by schema components which information items [**must**](#dt-must) satisfy to be schema-valid.  Largely to be found in [Datatype components (§4)](#datatype-components).

## <a id="typesystem"></a>2 Datatype System

This section describes the conceptual framework behind the datatype system defined in this specification.  The framework has been influenced by the [[ISO 11404]](#ISO11404) standard on language-independent datatypes as well as the datatypes for [[SQL]](#SQL) and for programming languages such as Java.

The datatypes discussed in this specification are for the most part well known abstract concepts such as *integer* and *date*. It is not the place of this specification to thoroughly define these abstract concepts; many other publications provide excellent definitions. However, this specification will attempt to describe the abstract concepts well enough that they can be readily recognized and distinguished from other abstractions with which they may be confused.

**Note:** Only those operations and relations needed for schema processing are defined in this specification. Applications using these datatypes are generally expected to implement appropriate additional functions and/or relations to make the datatype generally useful.  For example, the description herein of the [float](#float) datatype does not define addition or multiplication, much less all of the operations defined for that datatype in [[IEEE 754-2008]](#ieee754-2008) on which it is based.  For some datatypes (e.g. [language](#language) or [anyURI](#anyURI)) defined in part by reference to other specifications which impose constraints not part of the datatypes as defined here, applications may also wish to check that values conform to the requirements given in the current version of the relevant external specification.

### <a id="datatype"></a>2.1 Datatype

<a id="dt-datatype">[Definition:]</a>In this specification, a **datatype** has three properties:

- A [value space](#dt-value-space), which is a set of values.
- A [lexical space](#dt-lexical-space), which is a set of [literals](#dt-literal) used to denote the values.
- A small collection of *functions, relations, and procedures* associated with the datatype.  Included are equality and (for some datatypes) order relations on the [value space](#dt-value-space), and a [lexical mapping](#dt-lexical-mapping), which is a mapping from the [lexical space](#dt-lexical-space) into the [value space](#dt-value-space).
  **Note:** This specification only defines the operations and relations needed for schema processing.  The choice of terminology for describing/naming the datatypes is selected to guide users and implementers in how to expand the datatype to be generally useful—i.e., how to recognize the "real world" datatypes and their variants for which the datatypes defined herein are meant to be used for data interchange.

Along with the [lexical mapping](#dt-lexical-mapping) it is often useful to have an inverse which provides a standard [lexical representation](#dt-lexical-representation) for each value.  Such a [canonical mapping](#dt-canonical-mapping) is not required for schema processing, but is described herein for the benefit of users of this specification, and other specifications which might find it useful to reference these descriptions normatively. For some datatypes, notably [QName](#QName) and [NOTATION](#NOTATION), the mapping from lexical representations to values is context-dependent; for these types, no [canonical mapping](#dt-canonical-mapping) is defined.

**Note:** Where [canonical mappings](#dt-canonical-mapping) are defined in this specification, they are defined for [primitive](#dt-primitive) datatypes. When a datatype is derived using facets which directly constrain the [value space](#dt-value-space), then for each value eliminated from the [value space](#dt-value-space), the corresponding lexical representations are dropped from the lexical space. The [canonical mapping](#dt-canonical-mapping) for such a datatype is a subset of the [canonical mapping](#dt-canonical-mapping) for its [primitive](#dt-primitive) type and provides a [canonical representation](#dt-canonical-representation) for each value remaining in the [value space](#dt-value-space).  The [pattern](#dt-pattern) facet, on the other hand, and any other ([implementation-defined](#key-impl-def)) [lexical](#dt-lexical) facets, restrict the [lexical space](#dt-lexical-space) directly. When more than one lexical representation is provided for a given value, such facets may remove the [canonical representation](#dt-canonical-representation) while permitting a different lexical representation; in this case, the value remains in the [value space](#dt-value-space) but has no [canonical representation](#dt-canonical-representation). This specification provides no recourse in such situations. Applications are free to deal with it as they see fit. **Note:** This specification sometimes uses the shorter form "type" where one might strictly speaking expect the longer form "datatype" (e.g. in the phrases "union type", "list type", "base type", "item type", etc. No systematic distinction is intended between the forms of these phrase with "type" and those with "datatype"; the two forms are used interchangeably.The distinction between "datatype" and "simple type definition", by contrast, carries more information: the datatype is characterized by its [value space](#dt-value-space), [lexical space](#dt-lexical-space), [lexical mapping](#dt-lexical-mapping), etc., as just described, independently of the specific facets or other definitional mechanisms used in the simple type definition to describe that particular [value space](#dt-value-space) or [lexical space](#dt-lexical-space). Different simple type definitions with different selections of facets can describe the same datatype.

### <a id="value-space"></a>2.2 Value space

        2.2.1 [Identity](#identity)
         2.2.2 [Equality](#equality)
         2.2.3 [Order](#order)

<a id="dt-value-space">[Definition:]</a>The **value space** *of a datatype* is the set of values for that datatype.  Associated with each value space are selected operations and relations necessary to permit proper schema processing.  Each value in the value space of a [primitive](#dt-primitive) or [ordinary](#dt-ordinary) datatype is denoted by one or more character strings in its [lexical space](#dt-lexical-space), according to [the lexical mapping](#dt-lexical-mapping); [special](#dt-special) datatypes, by contrast, may include "ineffable" values not mapped to by any lexical representation. (If the mapping is restricted during a derivation in such a way that a value has no denotation, that value is dropped from the value space.)

The value spaces of datatypes are abstractions, and are defined in [Built-in Datatypes and Their Definitions (§3)](#built-in-datatypes) to the extent needed to clarify them for readers.  For example, in defining the numerical datatypes, we assume some general numerical concepts such as number and integer are known.  In many cases we provide references to other documents providing more complete definitions.

**Note:** *The value spaces and the values therein are abstractions.*  This specification does not prescribe any particular internal representations that must be used when implementing these datatypes.  In some cases, there are references to other specifications which do prescribe specific internal representations; these specific internal representations must be used to comply with those other specifications, but need not be used to comply with this specification.In addition, other applications are expected to define additional appropriate operations and/or relations on these value spaces (e.g., addition and multiplication on the various numerical datatypes' value spaces), and are permitted where appropriate to even redefine the operations and relations defined within this specification, provided that *for schema processing the relations and operations used are those defined herein*.The [value space](#dt-value-space) of a datatype can be defined in one of the following ways:

- defined elsewhere axiomatically from fundamental notions (intensional definition) [see [primitive](#dt-primitive)]
- enumerated outright from values of an already defined datatype (extensional definition) [see [enumeration](#dt-enumeration)]
- defined by restricting the [value space](#dt-value-space) of an already defined datatype to a particular subset with a given set of properties [see [derived](#dt-derived)]
- defined as a combination of values from one or more already defined [value space](#dt-value-space)(s) by a specific construction procedure [see [list](#dt-list) and [union](#dt-union)]

The relations of *identity* and *equality* are required for each value space. An order relation is specified for some value spaces, but not all. A very few datatypes have other relations or operations prescribed for the purposes of this specification.

#### <a id="identity"></a>2.2.1 Identity

The identity relation is always defined. Every value space inherently has an identity relation. Two things are *identical* if and only if they are actually the same thing: i.e., if there is no way whatever to tell them apart. 

**Note:** This does not preclude implementing datatypes by using more than one *internal* representation for a given value, provided no mechanism inherent in the datatype implementation (i.e., other than bit-string-preserving "casting" of the datum to a different datatype) will distinguish between the two representations.

In the identity relation defined herein, values from different [primitive](#dt-primitive) datatypes' [value spaces](#dt-value-space) are made artificially distinct if they might otherwise be considered identical.  For example, there is a number *two* in the [decimal](#decimal) datatype and a number *two* in the [float](#float) datatype.  In the identity relation defined herein, these two values are considered distinct.  Other applications making use of these datatypes may choose to consider values such as these identical, but for the view of [primitive](#dt-primitive) datatypes' [value spaces](#dt-value-space) used herein, they are distinct.

*WARNING:*  Care must be taken when identifying values across distinct primitive datatypes.  The [literals](#dt-literal) '`0.1`' and '`0.10000000009`' map to the same value in [float](#float) (neither 0.1 nor 0.10000000009 is in the value space, and each literal is mapped to the nearest value, namely 0.100000001490116119384765625), but map to distinct values in [decimal](#decimal).

**Note:** Datatypes [constructed](#dt-constructed) by [facet-based restriction](#dt-fb-restriction) do not create new values; they define subsets of some [primitive](#dt-primitive) datatype's [value space](#dt-value-space). A consequence of this fact is that the [literals](#dt-literal) '`+2`', treated as a [decimal](#decimal), '`+2`', treated as an [integer](#integer), and '`+2`', treated as a [byte](#byte), all denote the same value. They are not only equal but identical.

Given a list *A* and a list *B*, *A* and *B* are the same list if they are the same sequence of atomic values. The necessary and sufficient conditions for this identity are that *A* and *B* have the same length and that the items of *A* are pairwise identical to the items of *B*.

**Note:** It is a consequence of the rule just given for list identity that there is only one empty list. An empty list declared as having [item type](#dt-itemType) [decimal](#decimal) and an empty list declared as having [item type](#dt-itemType) [string](#string) are not only equal but identical.

#### <a id="equality"></a>2.2.2 Equality

Each [primitive](#dt-primitive) datatype has prescribed an equality relation for its value space.  The equality relation for most datatypes is the identity relation.  In the few cases where it is not, equality has been carefully defined so that for most operations of interest to the datatype, if two values are equal and one is substituted for the other as an argument to any of the operations, the results will always also be equal.

On the other hand, equality need not cover the entire value space of the datatype (though it usually does). In particular, NaN is not equal to itself in the [float](#float) and [double](#double) datatypes.

This equality relation is used in conjunction with identity when making [facet-based restrictions](#dt-fb-restriction) by *enumeration*, when checking identity constraints (in the context of [[XSD 1.1 Part 1: Structures]](#structural-schemas)) and when checking value constraints. It is used in conjunction with order when making [facet-based restrictions](#dt-fb-restriction) involving order. The equality relation used in the evaluation of XPath expressions may differ.  When [processing XPath expressions](https://www.w3.org/TR/xpath20/#id-expression-processing) as part of XML schema-validity [assessment](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-va) or otherwise testing membership in the [value space](#dt-value-space) of a datatype whose derivation involves [assertions](#dt-assertions), equality (like all other relations) within those expressions is interpreted using the rules of XPath ([[XPath 2.0]](#XPATH2)).  All comparisons for "sameness" prescribed by this specification test for either equality or identity, not for identity alone.

**Note:** In the prior version of this specification (1.0), equality was always identity.  This has been changed to permit the datatypes defined herein to more closely match the "real world" datatypes for which they are intended to be used as transmission formats.For example, the [float](#float) datatype has an equality which is not the identity ( −0 = +0 , but they are not identical—although they *were* identical in the 1.0 version of this specification), and whose domain excludes one value, NaN, so that  NaN ≠ NaN .For another example, the [dateTime](#dateTime) datatype previously lost any time-zone offset information in the [lexical representation](#dt-lexical-representation) as the value was converted to [UTC](#dt-utc); now the time zone offset is retained and two values representing the same "moment in time" but with different remembered time zone offsets are now *equal* but not *identical*.

In the equality relation defined herein, values from different primitive data spaces are made artificially unequal even if they might otherwise be considered equal.  For example, there is a number *two* in the [decimal](#decimal) datatype and a number *two* in the [float](#float) datatype.  In the equality relation defined herein, these two values are considered unequal.  Other applications making use of these datatypes may choose to consider values such as these equal; nonetheless, in the equality relation defined herein, they are unequal.

Two lists *A* and *B* are equal if and only if they have the same length and their items are pairwise equal. A list of length one containing a value *V1* and an atomic value *V2* are equal if and only if *V1* is equal to *V2*.

For the purposes of this specification, there is one equality relation for all values of all datatypes (the union of the various datatype's individual equalities, if one consider relations to be sets of ordered pairs).  The *equality* relation is denoted by '=' and its negation by '≠', each used as a binary infix predicate:  *x* = *y*  and  *x* ≠ *y* .  On the other hand, *identity* relationships are always described in words.

#### <a id="order"></a>2.2.3 Order

For some datatypes, an order relation is prescribed for use in checking upper and lower bounds of the [value space](#dt-value-space).  This order may be a *partial* order, which means that there may be values in the [value space](#dt-value-space) which are neither equal, less-than, nor greater-than.  Such value pairs are *incomparable*.  In many cases, no order is prescribed; each pair of values is either equal or [incomparable](#dt-incomparable). <a id="dt-incomparable">[Definition:]</a>Two values that are neither equal, less-than, nor greater-than are **incomparable**. Two values that are not [incomparable](#dt-incomparable) are **comparable**.

The order relation is used in conjunction with equality when making [facet-based restrictions](#dt-fb-restriction) involving order.  This is the only use of this order relation for schema processing.  Of course, when [processing XPath expressions](https://www.w3.org/TR/xpath20/#id-expression-processing) as part of XML schema-validity [assessment](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-va) or otherwise testing membership in the [value space](#dt-value-space) of a datatype whose derivation involves [assertions](#dt-assertions), order (like all other relations) within those expressions is interpreted using the rules of XPath ([[XPath 2.0]](#XPATH2)).

In this specification, this less-than order relation is denoted by '<' (and its inverse by '>'), the weak order by '≤' (and its inverse by '≥'), and the resulting [incomparable](#dt-incomparable) relation by '<>', each used as a binary infix predicate:  *x* < *y* ,  *x* ≤ *y* ,  *x* > *y* ,  *x* ≥ *y* , and  *x* <> *y* .

**Note:** The weak order "less-than-or-equal" means "less-than" or "equal" *and one can tell which*.  For example, the [duration](#duration) P1M (one month) is *not* less-than-or-equal P31D (thirty-one days) because P1M is not less than P31D, nor is P1M equal to P31D.  Instead, P1M is [incomparable](#dt-incomparable) with P31D.)  The formal definition of order for [duration](#duration) ([duration (§3.3.6)](#duration)) ensures that this is true.

For purposes of this specification, the value spaces of primitive datatypes are disjoint, even in cases where the abstractions they represent might be thought of as having values in common.  In the order relations defined in this specification, values from different value spaces are [incomparable](#dt-incomparable).  For example, the numbers two and three are values in both the decimal datatype and the float datatype.  In the order relation defined here, the two in the decimal datatype is not less than the three in the float datatype; the two values are incomparable.  Other applications making use of these datatypes may choose to consider values such as these comparable.

**Note:** Comparison of values from different [primitive](#dt-primitive) datatypes can sometimes be an error and sometimes not, depending on context.  When made for purposes of checking an enumeration constraint, such a comparison is not in itself an error, but since no two values from different [primitive](#dt-primitive) [value spaces](#dt-value-space) are equal, any comparison of [incomparable](#dt-incomparable) values will invariably be false.  Specifying an upper or lower bound which is of the wrong primitive datatype (and therefore [incomparable](#dt-incomparable) with the values of the datatype it is supposed to restrict) is, by contrast, always an error. It is a consequence of the rules for [facet-based restriction](#dt-fb-restriction) that in conforming simple type definitions, the values of upper and lower bounds, and enumerated values, **must** be drawn from the value space of the [base type](#dt-basetype), which necessarily means from the same [primitive](#dt-primitive) datatype.  Comparison of [incomparable](#dt-incomparable) values in the context of an XPath expression (e.g. in an assertion or in the rules for conditional type assignment) can raise a dynamic error in the evaluation of the XPath expression; see [[XQuery 1.0 and XPath 2.0 Functions and Operators]](#F_O) for details.

### <a id="lexical-space"></a>2.3 The Lexical Space and Lexical Mapping

<a id="dt-lexical-mapping">[Definition:]</a>The **lexical mapping** for a datatype is a prescribed relation which maps from the [lexical space](#dt-lexical-space) of the datatype into its [value space](#dt-value-space).

<a id="dt-lexical-space">[Definition:]</a>The **lexical space** of a datatype is the prescribed set of strings which [the lexical mapping](#dt-lexical-mapping) for that datatype maps to values of that datatype.

<a id="dt-lexical-representation">[Definition:]</a>The members of the [lexical space](#dt-lexical-space) are **lexical representations** of the values to which they are mapped.

**Note:** For the [special](#dt-special) datatypes, the [lexical mappings](#dt-lexical-mapping) defined here map from the [lexical space](#dt-lexical-space) into, but not onto, the [value space](#dt-value-space). The [value spaces](#dt-value-space) of the [special](#dt-special) datatypes include "ineffable" values for which the [lexical mappings](#dt-lexical-mapping) defined in this specification provide no lexical representation.For the [primitive](#dt-primitive) and [ordinary](#dt-ordinary) atomic datatypes, the [lexical mapping](#dt-lexical-mapping) is a (total) function on the entire [lexical space](#dt-lexical-space) *onto* (not merely *into*) the [value space](#dt-value-space): every member of the [lexical space](#dt-lexical-space) maps into the [value space](#dt-value-space), and every value is mapped to by some member of the [lexical space](#dt-lexical-space).For [union](#dt-union) datatypes, the [lexical mapping](#dt-lexical-mapping) is not necessarily a function, since the same [literal](#dt-literal) may map to different values in different member types. For [list](#dt-list) datatypes, the [lexical mapping](#dt-lexical-mapping) is a function if and only if the [lexical mapping](#dt-lexical-mapping) of the list's [item type](#dt-itemType) is a function.

 <a id="dt-literal">[Definition:]</a>A sequence of zero or more characters in the Universal Character Set (UCS) which may or may not prove upon inspection to be a member of the [lexical space](#dt-lexical-space) of a given datatype and thus a [lexical representation](#dt-lexical-representation) of a given value in that datatype's [value space](#dt-value-space), is referred to as a **literal**. The term is used indifferently both for character sequences which are members of a particular [lexical space](#dt-lexical-space) and for those which are not.

If a derivation introduces a [pre-lexical](#dt-pre-lexical) facet value (a new value for [whiteSpace](#f-w) or an implementation-defined [pre-lexical](#dt-pre-lexical) facet), the corresponding [pre-lexical](#dt-pre-lexical) transformation of a character string, if indeed it changed that string, could prevent that string from ever having the [lexical mapping](#dt-lexical-mapping) of the derived datatype applied to it.  Character strings that a [pre-lexical](#dt-pre-lexical) transformation blocks in this way (i.e., they are not in the range of the [pre-lexical](#dt-pre-lexical) facet's transformation) are always dropped from the derived datatype's [lexical space](#dt-lexical-space).

**Note:** One should be aware that in the context of XML schema-validity [assessment](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-va), there are [pre-lexical](#dt-pre-lexical) transformations of the input character string (controlled by the [whiteSpace](#f-w) facet and any implementation-defined [pre-lexical](#dt-pre-lexical) facets) which result in the intended [literal](#dt-literal).  Systems other than XML schema-validity [assessment](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-va) utilizing this specification may or may not implement these transformations.  If they do not, then input character strings that would have been transformed into correct [lexical representations](#dt-lexical-representation), when taken "raw", may not be correct [lexical representations](#dt-lexical-representation).

Should a derivation be made using a derivation mechanism that removes [lexical representations](#dt-lexical-representation) from the[lexical space](#dt-lexical-space) to the extent that one or more values cease to have any [lexical representation](#dt-lexical-representation), then those values are dropped from the [value space](#dt-value-space).

**Note:** This could happen by means of a [pattern](#f-p) or other [lexical](#dt-lexical) facet, or by a [pre-lexical](#dt-pre-lexical) facet as described above.

Conversely, should a derivation remove values then their [lexical representations](#dt-lexical-representation) are dropped from the [lexical space](#dt-lexical-space) unless there is a facet value whose impact is defined to cause the otherwise-dropped [lexical representation](#dt-lexical-representation) to be mapped to another value instead.

**Note:** There are currently no facets with such an impact.  There may be in the future.

For example, '100' and '1.0E2' are two different [lexical representations](#dt-lexical-representation) from the [float](#float) datatype which both denote the same value.  The datatype system defined in this specification provides mechanisms for schema designers to control the [value space](#dt-value-space) and the corresponding set of acceptable [lexical representations](#dt-lexical-representation) of those values for a datatype.

#### <a id="canonical-lexical-representation"></a>2.3.1 Canonical Mapping

While the datatypes defined in this specification often have a single [lexical representation](#dt-lexical-representation) for each value (i.e., each value in the datatype's [value space](#dt-value-space) is denoted by a single [representation](#dt-lexical-representation) in its [lexical space](#dt-lexical-space)), this is not always the case.  The example in the previous section shows two [lexical representations](#dt-lexical-representation) from the [float](#float) datatype which denote the same value.

<a id="dt-canonical-mapping">[Definition:]</a>The **canonical mapping** is a prescribed subset of the inverse of a [lexical mapping](#dt-lexical-mapping) which is one-to-one and whose domain (where possible) is the entire range of the [lexical mapping](#dt-lexical-mapping) (the [value space](#dt-value-space)).  Thus a [canonical mapping](#dt-canonical-mapping) selects one [lexical representation](#dt-lexical-representation) for each value in the [value space](#dt-value-space).

<a id="dt-canonical-representation">[Definition:]</a>The **canonical representation** of a value in the [value space](#dt-value-space) of a datatype is the [lexical representation](#dt-lexical-representation) associated with that value by the datatype's [canonical mapping](#dt-canonical-mapping).

[Canonical mappings](#dt-canonical-mapping) are not available for datatypes whose [lexical mappings](#dt-lexical-mapping) are context dependent (i.e., mappings for which the value of a [lexical representation](#dt-lexical-representation) depends on the context in which it occurs, or for which a character string may or may not be a valid [lexical representation](#dt-lexical-representation) similarly depending on its context)

**Note:** [Canonical representations](#dt-canonical-representation) are provided where feasible for the use of other applications; they are not required for schema processing itself.  *A conforming schema processor implementation is not required to implement [canonical mappings](#dt-canonical-mapping).*

### <a id="datatype-dichotomies"></a>2.4 Datatype Distinctions

        2.4.1 [Atomic vs. List vs. Union Datatypes](#atomic-vs-list)
             2.4.1.1 [Atomic Datatypes](#atomic)
             2.4.1.2 [List Datatypes](#list-datatypes)
             2.4.1.3 [Union datatypes](#union-datatypes)
         2.4.2 [Special vs. Primitive vs. Ordinary Datatypes](#primitive-vs-derived)
             2.4.2.1 [Facet-based Restriction](#restriction)
             2.4.2.2 [Construction by List](#list)
             2.4.2.3 [Construction by Union](#union)
         2.4.3 [Definition, Derivation, Restriction, and Construction](#derivation)
         2.4.4 [Built-in vs. User-Defined Datatypes](#built-in-vs-user-derived)

It is useful to categorize the datatypes defined in this specification along various dimensions, defining terms which can be used to characterize datatypes and the [Simple Type Definition](#std)s which define them.

#### <a id="atomic-vs-list"></a>2.4.1 Atomic vs. List vs. Union Datatypes

First, we distinguish [atomic](#dt-atomic), [list](#dt-list), and [union](#dt-union) datatypes.

<a id="dt-atomic-value">[Definition:]</a>An **atomic value** is an elementary value, not constructed from simpler values by any user-accessible means defined by this specification.

- <a id="dt-atomic">[Definition:]</a>**Atomic** datatypes are those whose [value spaces](#dt-value-space) contain only [atomic values](#dt-atomic-value).  **Atomic** datatypes are [anyAtomicType](#anyAtomicType) and all datatypes [derived](#dt-derived) from it.
- <a id="dt-list">[Definition:]</a>**List** datatypes are those having values each of which consists of a finite-length (possibly empty) sequence of [atomic values](#dt-atomic-value). The values in a list are drawn from some [atomic](#dt-atomic) datatype (or from a [union](#dt-union) of [atomic](#dt-atomic) datatypes), which is the [item type](#dt-itemType) of the **list**. **Note:** It is a consequence of constraints normatively specified elsewhere in this document (in particular, the component properties specified in [The Simple Type Definition Schema Component (§4.1.1)](#dc-defn)) that the [item type](#dt-itemType) of a list **may** be any [atomic](#dt-atomic) datatype, or any [union](#dt-union) datatype whose [basic members](#dt-basicmember) are all [atomic](#dt-atomic) datatypes (so a [list](#dt-list) of a [union](#dt-union) of [atomic](#dt-atomic) datatypes is possible, but not a [list](#dt-list) of a [union](#dt-union) of [lists](#dt-list)). The [item type](#dt-itemType) of a list **must not** itself be a list datatype.
- <a id="dt-union">[Definition:]</a>**Union** datatypes are (a) those whose [value spaces](#dt-value-space), [lexical spaces](#dt-lexical-space), and [lexical mappings](#dt-lexical-mapping) are the union of the [value spaces](#dt-value-space), [lexical spaces](#dt-lexical-space), and [lexical mappings](#dt-lexical-mapping) of one or more other datatypes, which are the [member types](#dt-memberTypes) of the union, or (b) those derived by [facet-based restriction](#dt-fb-restriction) of another union datatype. **Note:** It is a consequence of constraints normatively specified elsewhere in this document (in particular, the component properties specified in [The Simple Type Definition Schema Component (§4.1.1)](#dc-defn)) that any [primitive](#dt-primitive) or [ordinary](#dt-ordinary) datatype **may** occur among the [member types](#dt-memberTypes) of a [union](#dt-union). (In particular, [union](#dt-union) datatypes may themselves be members of [unions](#dt-union), as may [lists](#dt-list).) The only prohibition is that no [special](#dt-special) datatype may be a member of a [union](#dt-union).

For example, a single token which [matches](#dt-match) [Nmtoken](https://www.w3.org/TR/xml11/#NT-Nmtoken) from [[XML]](#XML) is in the value space of the [atomic](#dt-atomic) datatype [NMTOKEN](#NMTOKEN), while a sequence of such tokens is in the value space of the [list](#dt-list) datatype [NMTOKENS](#NMTOKENS).

##### <a id="atomic"></a>2.4.1.1 Atomic Datatypes

An [atomic](#dt-atomic) datatype has a [value space](#dt-value-space) consisting of a set of "atomic" or elementary values.

**Note:** Atomic values are sometimes regarded, and described, as "not decomposable", but in fact the values in several datatypes defined here are described with internal structure, which is appealed to in checking whether particular values satisfy various constraints (e.g. upper and lower bounds on a datatype). Other specifications which use the datatypes defined here may define operations which attribute internal structure to values and expose or act upon that structure.

 The [lexical space](#dt-lexical-space) of an [atomic](#dt-atomic) datatype is a set of [literals](#dt-literal) whose internal structure is specific to the datatype in question.

There is one [special](#dt-special) [atomic](#dt-atomic) datatype ([anyAtomicType](#anyAtomicType)), and a number of [primitive](#dt-primitive) [atomic](#dt-atomic) datatypes which have [anyAtomicType](#anyAtomicType) as their [base type](#dt-basetype).  All other [atomic](#dt-atomic) datatypes are [derived](#dt-derived) either from one of the [primitive](#dt-primitive) [atomic](#dt-atomic) datatypes or from another [ordinary](#dt-ordinary) [atomic](#dt-atomic) datatype.  No [user-defined](#dt-user-defined) datatype **may** have [anyAtomicType](#anyAtomicType) as its [base type](#dt-basetype).

##### <a id="list-datatypes"></a>2.4.1.2 List Datatypes

[List](#dt-list) datatypes are always [constructed](#dt-constructed) from some other type; they are never [primitive](#dt-primitive). The [value space](#dt-value-space) of a [list](#dt-list) datatype is the set of finite-length sequences of zero or more [atomic](#dt-atomic) values where each [atomic](#dt-atomic) value is drawn from the [value space](#dt-value-space) of the lists's [item type](#dt-itemType) and has a [lexical representation](#dt-lexical-representation) containing no whitespace. The [lexical space](#dt-lexical-space) of a [list](#dt-list) datatype is a set of [literals](#dt-literal) each of which is a space-separated sequence of [literals](#dt-literal) of the [item type](#dt-itemType).

<a id="dt-itemType">[Definition:]</a> The [atomic](#dt-atomic) or [union](#dt-union) datatype that participates in the definition of a [list](#dt-list) datatype is the **item type** of that [list](#dt-list) datatype.  If the [item type](#dt-itemType) is a [union](#dt-union), each of its [basic members](#dt-basicmember) **must** be [atomic](#dt-atomic).

*Example*

```
<simpleType name='sizes'>
  <list itemType='decimal'/>
</simpleType>

```

```
<cerealSizes xsi:type='sizes'> 8 10.5 12 </cerealSizes>

```

A [list](#dt-list) datatype can be [constructed](#dt-constructed) from an ordinary or [primitive](#dt-primitive) [atomic](#dt-atomic) datatype whose [lexical space](#dt-lexical-space) allows whitespace (such as [string](#string) or [anyURI](#anyURI)) or a [union](#dt-union) datatype any of whose [{member type definitions}](#std-member_type_definitions)'s [lexical space](#dt-lexical-space) allows space. Since [list](#dt-list) items are separated at whitespace before the [lexical representations](#dt-lexical-representation) of the items are mapped to values, no whitespace will ever occur in the [lexical representation](#dt-lexical-representation) of a [list](#dt-list) item, even when the item type would in principle allow it.  For the same reason, when every possible [lexical representation](#dt-lexical-representation) of a given value in the [value space](#dt-value-space) of the [item type](#dt-itemType) includes whitespace, that value can never occur as an item in any value of the [list](#dt-list) datatype.

*Example*

```
<simpleType name='listOfString'>
  <list itemType='string'/>
</simpleType>

```

```
<someElement xsi:type='listOfString'>
this is not list item 1
this is not list item 2
this is not list item 3
</someElement>

```

 In the above example, the value of the *someElement* element is not a [list](#dt-list) of [length](#dt-length) 3; rather, it is a [list](#dt-list) of [length](#dt-length) 18.

When a datatype is [derived](#dt-derived) by [restricting](#dt-fb-restriction) a [list](#dt-list) datatype, the following [constraining facets](#dt-constraining-facet) apply:

- [length](#dt-length)
- [maxLength](#dt-maxLength)
- [minLength](#dt-minLength)
- [enumeration](#dt-enumeration)
- [pattern](#dt-pattern)
- [whiteSpace](#dt-whiteSpace)
- [assertions](#dt-assertions)

For each of [length](#dt-length), [maxLength](#dt-maxLength) and [minLength](#dt-minLength), the *length* is measured in number of list items.  The value of [whiteSpace](#dt-whiteSpace) is fixed to the value ***collapse***.

For [list](#dt-list) datatypes the [lexical space](#dt-lexical-space) is composed of space-separated [literals](#dt-literal) of the [item type](#dt-itemType).  Any [pattern](#dt-pattern) specified when a new datatype is [derived](#dt-derived) from a [list](#dt-list) datatype applies to the members of the [list](#dt-list) datatype's [lexical space](#dt-lexical-space), not to the members of the [lexical space](#dt-lexical-space) of the [item type](#dt-itemType).  Similarly, enumerated values are compared to the entire [list](#dt-list), not to individual list items, and [assertions](#f-a) apply to the entire [list](#dt-list) too. Lists are identical if and only if they have the same length and their items are pairwise identical; they are equal if and only if they have the same length and their items are pairwise equal. And a list of length one whose item is an atomic value *V1* is equal or identical to an atomic value *V2* if and only if *V1* is equal or identical to *V2*.

*Example*

```
<xs:simpleType name='myList'>
	<xs:list itemType='xs:integer'/>
</xs:simpleType>
<xs:simpleType name='myRestrictedList'>
	<xs:restriction base='myList'>
		<xs:pattern value='123 (\d+\s)*456'/>
	</xs:restriction>
</xs:simpleType>
<someElement xsi:type='myRestrictedList'>123 456</someElement>
<someElement xsi:type='myRestrictedList'>123 987 456</someElement>
<someElement xsi:type='myRestrictedList'>123 987 567 456</someElement>

```

The [canonical mapping](#dt-canonical-mapping) of a [list](#dt-list) datatype maps each value onto the space-separated concatenation of the [canonical representations](#dt-canonical-representation) of all the items in the value (in order), using the [canonical mapping](#dt-canonical-mapping) of the [item type](#dt-itemType).

##### <a id="union-datatypes"></a>2.4.1.3 Union datatypes

 Union types may be defined in either of two ways. When a union type is [constructed](#dt-constructed) by [union](#dt-union), its [value space](#dt-value-space), [lexical space](#dt-lexical-space), and [lexical mapping](#dt-lexical-mapping) are the "ordered unions" of the [value spaces](#dt-value-space), [lexical spaces](#dt-lexical-space), and [lexical mappings](#dt-lexical-mapping) of its [member types](#dt-memberTypes).

It will be observed that the [lexical mapping](#dt-lexical-mapping) of a union, so defined, is not necessarily a function: a given [literal](#dt-literal) may map to one value or to several values of different [primitive](#dt-primitive) datatypes, and it may be indeterminate which value is to be preferred in a particular context. When the datatypes defined here are used in the context of [[XSD 1.1 Part 1: Structures]](#structural-schemas), the `xsi:type` attribute defined by that specification in section [xsi:type](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#xsi_type) can be used to indicate which value a [literal](#dt-literal) which is the content of an element should map to. In other contexts, other rules (such as type coercion rules) may be employed to determine which value is to be used.

When a union type is defined by [restricting](#dt-fb-restriction) another [union](#dt-union), its [value space](#dt-value-space), [lexical space](#dt-lexical-space), and [lexical mapping](#dt-lexical-mapping) are subsets of the [value spaces](#dt-value-space), [lexical spaces](#dt-lexical-space), and [lexical mappings](#dt-lexical-mapping) of its [base type](#dt-basetype).

[Union](#dt-union) datatypes are always [constructed](#dt-constructed) from other datatypes; they are never [primitive](#dt-primitive). Currently, there are no [built-in](#dt-built-in) [union](#dt-union) datatypes.

*Example*

 A prototypical example of a [union](#dt-union) type is the [maxOccurs attribute](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#p-max_occurs) on the [element element](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#element-element) in XML Schema itself: it is a union of nonNegativeInteger and an enumeration with the single member, the string "unbounded", as shown below.

```
  <attributeGroup name="occurs">
    <attribute name="minOccurs" type="nonNegativeInteger"
    	use="optional" default="1"/>
    <attribute name="maxOccurs"use="optional" default="1">
      <simpleType>
        <union>
          <simpleType>
            <restriction base='nonNegativeInteger'/>
          </simpleType>
          <simpleType>
            <restriction base='string'>
              <enumeration value='unbounded'/>
            </restriction>
          </simpleType>
        </union>
      </simpleType>
    </attribute>
  </attributeGroup>

```

Any number (zero or more) of ordinary or [primitive](#dt-primitive) [datatypes](#dt-datatype) can participate in a [union](#dt-union) type.

<a id="dt-memberTypes">[Definition:]</a> The datatypes that participate in the definition of a [union](#dt-union) datatype are known as the **member types** of that [union](#dt-union) datatype.

**Note:** When datatypes are represented using XSD schema components, as described in [Datatype components (§4)](#datatype-components), the member types of a union are those simple type definitions given in the [{member type definitions}](#std-member_type_definitions) property.

<a id="dt-transitivemembership">[Definition:]</a>The **transitive membership** of a [union](#dt-union) is the set of its own [member types](#dt-memberTypes), and the [member types](#dt-memberTypes) of its members, and so on. More formally, if *U* is a [union](#dt-union), then (a) its [member types](#dt-memberTypes) are in the transitive membership of *U*, and (b) for any datatypes *T1* and *T2*, if *T1* is in the transitive membership of *U* and *T2* is one of the [member types](#dt-memberTypes) of *T1*, then *T2* is also in the transitive membership of *U*.

The [transitive membership](#dt-transitivemembership) of a [union](#dt-union) **must not** contain the [union](#dt-union) itself, nor any datatype [derived](#dt-derived) or [constructed](#dt-constructed) from the [union](#dt-union).

<a id="dt-basicmember">[Definition:]</a>Those members of the [transitive membership](#dt-transitivemembership) of a [union](#dt-union) datatype *U* which are themselves not [union](#dt-union) datatypes are the **basic members** of *U*.

<a id="dt-interveningunion">[Definition:]</a>If a datatype *M* is in the [transitive membership](#dt-transitivemembership) of a [union](#dt-union) datatype *U*, but not one of *U*'s [member types](#dt-memberTypes), then a sequence of one or more [union](#dt-union) datatypes necessarily exists, such that the first is one of the [member types](#dt-memberTypes) of *U*, each is one of the [member types](#dt-memberTypes) of its predecessor in the sequence, and *M* is one of the [member types](#dt-memberTypes) of the last in the sequence. The [union](#dt-union) datatypes in this sequence are said to **intervene** between *M* and *U*. When *U* and *M* are given by the context, the datatypes in the sequence are referred to as the **intervening unions**. When *M* is one of the [member types](#dt-memberTypes) of *U*, the set of **intervening unions** is the empty set.

 <a id="dt-active-member">[Definition:]</a>In a valid instance of any [union](#dt-union), the first of its members in order which accepts the instance as valid is the **active member type**. <a id="dt-active-basic-member">[Definition:]</a>If the [active member type](#dt-active-member) is itself a [union](#dt-union), one of *its* members will be *its* [active member type](#dt-active-member), and so on, until finally a [basic (non-union) member](#dt-basicmember) is reached. That [basic member](#dt-basicmember) is the **active basic member** of the union.

The order in which the [member types](#dt-memberTypes) are specified in the definition (that is, in the case of datatypes defined in a schema document, the order of the <simpleType> children of the <union> element, or the order of the [QName](#QName)s in the `memberTypes` attribute) is significant. During validation, an element or attribute's value is validated against the [member types](#dt-memberTypes) in the order in which they appear in the definition until a match is found.  As noted above, the evaluation order can be overridden with the use of [xsi:type](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#xsi_type).

*Example*

 For example, given the definition below, the first instance of the <size> element validates correctly as an [integer (§3.4.13)](#integer), the second and third as [string (§3.3.1)](#string).

```
  <xs:element name='size'>
    <xs:simpleType>
      <xs:union>
        <xs:simpleType>
          <xs:restriction base='integer'/>
        </xs:simpleType>
        <xs:simpleType>
          <xs:restriction base='string'/>
        </xs:simpleType>
      </xs:union>
    </xs:simpleType>
  </xs:element>

```

```
  <size>1</size>
  <size>large</size>
  <size xsi:type='xs:string'>1</size>

```

The [canonical mapping](#dt-canonical-mapping) of a [union](#dt-union) datatype maps each value onto the [canonical representation](#dt-canonical-representation) of that value obtained using the [canonical mapping](#dt-canonical-mapping) of the first [member type](#dt-memberTypes) in whose value space it lies.

 When a datatype is [derived](#dt-derived) by [restricting](#dt-fb-restriction) a [union](#dt-union) datatype, the following [constraining facets](#dt-constraining-facet) apply:

- [enumeration](#dt-enumeration)
- [pattern](#dt-pattern)
- [assertions](#dt-assertions)

#### <a id="primitive-vs-derived"></a>2.4.2 Special vs. Primitive vs. Ordinary Datatypes

Next, we distinguish [special](#dt-special), [primitive](#dt-primitive), and [ordinary](#dt-ordinary) (or [constructed](#dt-constructed)) datatypes.  Each datatype defined by or in accordance with this specification falls into exactly one of these categories.

- <a id="dt-special">[Definition:]</a>The **special** datatypes are [anySimpleType](#anySimpleType) and [anyAtomicType](#anyAtomicType). They are special by virtue of their position in the type hierarchy.
- <a id="dt-primitive">[Definition:]</a>**Primitive** datatypes are those datatypes that are not [special](#dt-special) and are not defined in terms of other datatypes; they exist *ab initio*. All [primitive](#dt-primitive) datatypes have [anyAtomicType](#anyAtomicType) as their [base type](#dt-basetype), but their [value](#dt-value-space) and [lexical spaces](#dt-lexical-space) must be given in prose; they cannot be described as [restrictions](#dt-fb-restriction) of [anyAtomicType](#anyAtomicType) by the application of particular [constraining facets](#dt-constraining-facet).**Note:** As normatively specified elsewhere, conforming processors **must** support all the primitive datatypes defined in this specification; it is [implementation-defined](#key-impl-def) whether other primitive datatypes are supported.Processors **may**, for example, support the floating-point decimal datatype specified in [[Precision Decimal]](#pd-note).
- <a id="dt-ordinary">[Definition:]</a>**Ordinary** datatypes are all datatypes other than the [special](#dt-special) and [primitive](#dt-primitive) datatypes.  [Ordinary](#dt-ordinary) datatypes can be understood fully in terms of their [Simple Type Definition](#std) and the properties of the datatypes from which they are [constructed](#dt-constructed).

For example, in this specification, [float](#float) is a [primitive](#dt-primitive) datatype based on a well-defined mathematical concept and not defined in terms of other datatypes, while [integer](#integer) is [constructed](#dt-constructed) from the more general datatype [decimal](#decimal).

##### <a id="restriction"></a>2.4.2.1 Facet-based Restriction

<a id="dt-fb-restriction">[Definition:]</a>A datatype is defined by **facet-based restriction** of another datatype (its [base type](#dt-basetype)), when values for zero or more [constraining facets](#dt-constraining-facet) are specified that serve to constrain its [value space](#dt-value-space) and/or its [lexical space](#dt-lexical-space) to a subset of those of the [base type](#dt-basetype). The [base type](#dt-basetype) of a [facet-based restriction](#dt-fb-restriction) **must** be a [primitive](#dt-primitive) or [ordinary](#dt-ordinary) datatype.

##### <a id="list"></a>2.4.2.2 Construction by List

A [list](#dt-list) datatype can be [constructed](#dt-constructed) from another datatype (its [item type](#dt-itemType)) by creating a [value space](#dt-value-space) that consists of finite-length sequences of zero or more values of its [item type](#dt-itemType). Datatypes so [constructed](#dt-constructed) have [anySimpleType](#anySimpleType) as their [base type](#dt-basetype). Note that since the [value space](#dt-value-space) and [lexical space](#dt-lexical-space) of any [list](#dt-list) datatype are necessarily subsets of the [value space](#dt-value-space) and [lexical space](#dt-lexical-space) of [anySimpleType](#anySimpleType), any datatype [constructed](#dt-constructed) as a [list](#dt-list) is a [restriction](#dt-restriction) of its base type.

##### <a id="union"></a>2.4.2.3 Construction by Union

One datatype can be [constructed](#dt-constructed) from one or more datatypes by unioning their [lexical mappings](#dt-lexical-mapping) and, consequently, their [value spaces](#dt-value-space) and [lexical spaces](#dt-lexical-space).  Datatypes so [constructed](#dt-constructed) also have [anySimpleType](#anySimpleType) as their [base type](#dt-basetype). Note that since the [value space](#dt-value-space) and [lexical space](#dt-lexical-space) of any [union](#dt-union) datatype are necessarily subsets of the [value space](#dt-value-space) and [lexical space](#dt-lexical-space) of [anySimpleType](#anySimpleType), any datatype [constructed](#dt-constructed) as a [union](#dt-union) is a [restriction](#dt-restriction) of its base type.

#### <a id="derivation"></a>2.4.3 Definition, Derivation, Restriction, and Construction

Definition, derivation, restriction, and construction are conceptually distinct, although in practice they are frequently performed by the same mechanisms.

By 'definition' is meant the explicit identification of the relevant properties of a datatype, in particular its [value space](#dt-value-space), [lexical space](#dt-lexical-space), and [lexical mapping](#dt-lexical-mapping).

The properties of the [special](#dt-special) and the standard [primitive](#dt-primitive) datatypes are defined by this specification. A [Simple Type Definition](#std) is present for each of these datatypes in every valid schema; it serves as a representation of the datatype, but by itself it does not capture all the relevant information and does not suffice (without knowledge of this specification) to *define* the datatype.

**Note:** The properties of any [implementation-defined](#key-impl-def) [primitive](#dt-primitive) datatypes are given not here but in the documentation for the implementation in question. Alternatively, a primitive datatype not specified in this document can be specified in a document of its own not tied to a particular implementation; [[Precision Decimal]](#pd-note) is an example of such a document.

For all other datatypes, a [Simple Type Definition](#std) does suffice. The properties of an [ordinary](#dt-ordinary) datatype can be inferred from the datatype's [Simple Type Definition](#std) and the properties of the [base type](#dt-basetype), [item type](#dt-itemType) if any, and [member types](#dt-memberTypes) if any. All [ordinary](#dt-ordinary) datatypes can be defined in this way.

By 'derivation' is meant the relation of a datatype to its [base type](#dt-basetype), or to the [base type](#dt-basetype) of its [base type](#dt-basetype), and so on.

<a id="dt-basetype"></a>Every datatype other than [anySimpleType](#anySimpleType) is associated with another datatype, its **base type**. **Base types** can be [special](#dt-special), [primitive](#dt-primitive), or [ordinary](#dt-ordinary).

<a id="dt-immediately-derived">[Definition:]</a>A datatype *T* is **immediately derived** from another datatype *X* if and only if *X* is the [base type](#dt-basetype) of *T*.

**Note:** The above does not preclude the [Simple Type Definition](#std) for [anySimpleType](#anySimpleType) from having a value for its [{base type definition}](#std-base_type_definition).  (It does, and its value is [anyType](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#any-type-itself).) More generally, <a id="dt-derived"></a>A datatype *R* is **derived** from another datatype *B* if and only if one of the following is true:

- *B* is the [base type](#dt-basetype) of *R*.
- There is some datatype *X* such that *X* is the [base type](#dt-basetype) of *R*, and *X* is derived from *B*.

A datatype **must not** be [derived](#dt-derived) from itself. That is, the base type relation must be acyclic.

It is a consequence of the above that every datatype other than [anySimpleType](#anySimpleType) is [derived](#dt-derived) from [anySimpleType](#anySimpleType).

Since each datatype has exactly one [base type](#dt-basetype), and every datatype other than [anySimpleType](#anySimpleType) is [derived](#dt-derived) directly or indirectly from [anySimpleType](#anySimpleType), it follows that the [base type](#dt-basetype) relation arranges all simple types into a tree structure, which is conventionally referred to as the *derivation hierarchy*.

By 'restriction' is meant the definition of a datatype whose [value space](#dt-value-space) and [lexical space](#dt-lexical-space) are subsets of those of its [base type](#dt-basetype).

Formally, <a id="dt-restriction"></a>A datatype *R* is a **restriction** of another datatype *B* when

- the [value space](#dt-value-space) of *R* is a subset of the [value space](#dt-value-space) of *B*, and
- the [lexical space](#dt-lexical-space) of *R* is a subset of the [lexical space](#dt-lexical-space) of *B*.

 Note that all three forms of datatype [construction](#dt-constructed) produce [restrictions](#dt-restriction) of the [base type](#dt-basetype): [facet-based restriction](#dt-fb-restriction) does so by means of [constraining facets](#dt-constraining-facet), while [construction](#dt-constructed) by [list](#dt-list) or [union](#dt-union) does so because those [constructions](#dt-constructed) take [anySimpleType](#anySimpleType) as the [base type](#dt-basetype). It follows that all datatypes are [restrictions](#dt-restriction) of [anySimpleType](#anySimpleType). This specification provides no means by which a datatype may be defined so as to have a larger [lexical space](#dt-lexical-space) or [value space](#dt-value-space) than its [base type](#dt-basetype).

By 'construction' is meant the creation of a datatype by defining it in terms of another.

 <a id="dt-constructed">[Definition:]</a>All [ordinary](#dt-ordinary) datatypes are defined in terms of, or **constructed** from, other datatypes, either by [restricting](#dt-fb-restriction) the [value space](#dt-value-space) or [lexical space](#dt-lexical-space) of a [base type](#dt-basetype) using zero or more [constraining facets](#dt-constraining-facet) or by specifying the new datatype as a [list](#dt-list) of items of some [item type](#dt-itemType), or by defining it as a [union](#dt-union) of some specified sequence of [member types](#dt-memberTypes). These three forms of [construction](#dt-constructed) are often called "[facet-based restriction](#dt-fb-restriction)", "[construction](#dt-constructed) by [list](#dt-list)", and "[construction](#dt-constructed) by [union](#dt-union)", respectively. Datatypes so constructed may be understood fully (for purposes of a type system) in terms of (a) the properties of the datatype(s) from which they are constructed, and (b) their [Simple Type Definition](#std). This distinguishes [ordinary](#dt-ordinary) datatypes from the [special](#dt-special) and [primitive](#dt-primitive) datatypes, which can be understood only in the light of documentation (namely, their descriptions elsewhere in this specification, or, for [implementation-defined](#key-impl-def) [primitives](#dt-primitive), in the appropriate implementation-specific documentation). All [ordinary](#dt-ordinary) datatypes are [constructed](#dt-constructed), and all [constructed](#dt-constructed) datatypes are [ordinary](#dt-ordinary).

#### <a id="built-in-vs-user-derived"></a>2.4.4 Built-in vs. User-Defined Datatypes

-  <a id="dt-built-in">[Definition:]</a>**Built-in** datatypes are those which are defined in this specification; they can be [special](#dt-special), [primitive](#dt-primitive), or [ordinary](#dt-ordinary) datatypes .
-  <a id="dt-user-defined">[Definition:]</a> **User-defined** datatypes are those datatypes that are defined by individual schema designers.

The [built-in](#dt-built-in) datatypes are intended to be available automatically whenever this specification is implemented or used, whether by itself or embedded in a host language. In the language defined by [[XSD 1.1 Part 1: Structures]](#structural-schemas), the [built-in](#dt-built-in) datatypes are automatically included in every valid schema. Other host languages **should** specify that all of the datatypes decribed here as built-ins are automatically available; they **may** specify that additional datatypes are also made available automatically.

**Note:** [Implementation-defined](#key-impl-def) datatypes, whether [primitive](#dt-primitive) or [ordinary](#dt-ordinary), may sometimes be included automatically in any schemas processed by that implementation; nevertheless, they are not built in to *every* schema, and are thus not included in the term 'built-in', as that term is used in this specification.

The mechanism for making [user-defined](#dt-user-defined) datatypes available for use is not defined in this specification; if [user-defined](#dt-user-defined) datatypes are to be available, some such mechanism **must** be specified by the host language.

<a id="dt-unknown-dt">[Definition:]</a>A datatype which is not available for use is said to be **unknown**.

**Note:** From the schema author's perspective, a reference to a datatype which proves to be [unknown](#dt-unknown-dt) might reflect any of the following causes, or others:

- 1<a id="unkown.type"></a>An error has been made in giving the name of the datatype.
- 2<a id="unkown.sdoc"></a>The datatype is a [user-defined](#dt-user-defined) datatype which has not been made available using the means defined by the host language (e.g. because the appropriate schema document has not been consulted).
- 3<a id="unkown.id-primitive"></a>The datatype is an [implementation-defined](#key-impl-def) [primitive](#dt-primitive) datatype not supported by the implementation being used.
- 4<a id="unkown.id-derived"></a>The datatype is an [implementation-defined](#key-impl-def) [ordinary](#dt-ordinary) datatype which is made automatically available by some implementations, but not by the implementation being used.
- 5<a id="unkown.contaminated"></a>The datatype is a [user-defined](#dt-user-defined) [ordinary](#dt-ordinary) datatype whose base type is [unknown](#dt-unknown-dt)
 From the point of view of the implementation, these cases are likely to be indistinguishable. **Note:** In the terminology of [[XSD 1.1 Part 1: Structures]](#structural-schemas), the datatypes here called [unknown](#dt-unknown-dt) are referred to as [absent](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-null).

 Conceptually there is no difference between the [ordinary](#dt-ordinary) [built-in](#dt-built-in) datatypes included in this specification and the [user-defined](#dt-user-defined) datatypes which will be created by individual schema designers. The [built-in](#dt-built-in) [constructed](#dt-constructed) datatypes are those which are believed to be so common that if they were not defined in this specification many schema designers would end up reinventing them.  Furthermore, including these [constructed](#dt-constructed) datatypes in this specification serves to demonstrate the mechanics and utility of the datatype generation facilities of this specification.

## <a id="built-in-datatypes"></a>3 Built-in Datatypes and Their Definitions

Diagram showing the derivation relations in the built-in type hierarchy. (A [long description of the diagram](type-hierarchy-201104.longdesc.html) is available separately.)

                                                   Each built-in datatype defined in this specification can be uniquely addressed via a URI Reference constructed as follows:

1. the base URI is the URI of the XML Schema namespace
2. the fragment identifier is the name of the datatype
 For example, to address the [int](#int) datatype, the URI is:

- `http://www.w3.org/2001/XMLSchema#int`
 Additionally, each facet definition element can be uniquely addressed via a URI constructed as follows:

1. the base URI is the URI of the XML Schema namespace
2. the fragment identifier is the name of the facet
 For example, to address the maxInclusive facet, the URI is:

- `http://www.w3.org/2001/XMLSchema#maxInclusive`
 Additionally, each facet usage in a built-in [Simple Type Definition](#std) can be uniquely addressed via a URI constructed as follows:

1. the base URI is the URI of the XML Schema namespace
2. the fragment identifier is the name of the [Simple Type Definition](#std), followed by a period ('`.`') followed by the name of the facet
 For example, to address the usage of the maxInclusive facet in the definition of int, the URI is:

- `http://www.w3.org/2001/XMLSchema#int.maxInclusive`

### <a id="namespaces"></a>3.1 Namespace considerations

 The [built-in](#dt-built-in) datatypes defined by this specification are designed to be used with the XML Schema definition language as well as other XML specifications. To facilitate usage within the XML Schema definition language, the [built-in](#dt-built-in) datatypes in this specification have the namespace name:

- http://www.w3.org/2001/XMLSchema

 To facilitate usage in specifications other than the XML Schema definition language, such as those that do not want to know anything about aspects of the XML Schema definition language other than the datatypes, each [built-in](#dt-built-in) datatype is also defined in the namespace whose URI is:

- http://www.w3.org/2001/XMLSchema-datatypes

 Each [user-defined](#dt-user-defined) datatype may also be associated with a target namespace.  If it is constructed from a schema document, then its namespace is typically the target namespace of that schema document. (See [XML Representation of Schemas](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-schema) in [[XSD 1.1 Part 1: Structures]](#structural-schemas).)

### <a id="special-datatypes"></a>3.2 Special Built-in Datatypes

        3.2.1 [anySimpleType](#anySimpleType)
             3.2.1.1 [Value space](#sec-ast-vs)
             3.2.1.2 [Lexical mapping](#sec-ast-lex)
             3.2.1.3 [Facets](#sec-ast-f)
         3.2.2 [anyAtomicType](#anyAtomicType)
             3.2.2.1 [Value space](#sec-aat-vs)
             3.2.2.2 [Lexical mapping](#sec-aat-lex)
             3.2.2.3 [Facets](#sec-aat-f)

The two datatypes at the root of the hierarchy of simple types are [anySimpleType](#anySimpleType) and [anyAtomicType](#anyAtomicType).

#### <a id="anySimpleType"></a>3.2.1 anySimpleType

<a id="dt-anySimpleType"></a> The definition of **anySimpleType** is a special [restriction](#dt-restriction) of ***anyType***.  The [lexical space](#dt-lexical-space) of **anySimpleType** is the set of all sequences of Unicode characters, and its [value space](#dt-value-space) includes all [atomic values](#dt-atomic-value) and all finite-length lists of zero or more [atomic values](#dt-atomic-value).

For further details of [anySimpleType](#anySimpleType) and its representation as a [Simple Type Definition](#std), see [Built-in Simple Type Definitions (§4.1.6)](#builtin-stds).

##### <a id="sec-ast-vs"></a>3.2.1.1 Value space

The [value space](#dt-value-space) of [anySimpleType](#anySimpleType) is the set of all [atomic values](#dt-atomic-value) and of all finite-length lists of zero or more [atomic values](#dt-atomic-value).

**Note:** It is a consequence of this definition, together with the definition of the [lexical mapping](#dt-lexical-mapping) in the next section, that some values of this datatype have no [lexical representation](#dt-lexical-representation) using the [lexical mappings](#dt-lexical-mapping) defined by this specification. That is, the "potential" [value space](#dt-value-space) and the "effable" or "nameable" [value space](#dt-value-space) diverge for this datatype. As far as this specification is concerned, there is no operational difference between the potential and effable [value spaces](#dt-value-space) and the distinction is of mostly formal interest. Since some host languages for the type system defined here may allow means of construction values other than mapping from a [lexical representation](#dt-lexical-representation), the difference may have practical importance in some contexts. In those contexts, the term [value space](#dt-value-space) should unless otherwise qualified be taken to mean the potential [value space](#dt-value-space).

##### <a id="sec-ast-lex"></a>3.2.1.2 Lexical mapping

The [lexical space](#dt-lexical-space) of [anySimpleType](#anySimpleType) is the set of all finite-length sequences of zero or more [character](https://www.w3.org/TR/xml11/#dt-character)s (as defined in [[XML]](#XML)) that [match](#dt-match) the [Char](https://www.w3.org/TR/xml11/#NT-Char) production from [[XML]](#XML). This is equivalent to the union of the [lexical spaces](#dt-lexical-space) of all [primitive](#dt-primitive) and all possible [ordinary](#dt-ordinary) datatypes.

It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports the [Char](https://www.w3.org/TR/xml11/#NT-Char) production from [[XML]](#XML), or that from [[XML 1.0]](#XML1.0), or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

The [lexical mapping](#dt-lexical-mapping) of [anySimpleType](#anySimpleType) is the union of the [lexical mappings](#dt-lexical-mapping) of all [primitive](#dt-primitive) datatypes and all list datatypes. It will be noted that this mapping is not a function: a given [literal](#dt-literal) may map to one value or to several values of different [primitive](#dt-primitive) datatypes, and it may be indeterminate which value is to be preferred in a particular context. When the datatypes defined here are used in the context of [[XSD 1.1 Part 1: Structures]](#structural-schemas), the `xsi:type` attribute defined by that specification in section [xsi:type](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#xsi_type) can be used to indicate which value a [literal](#dt-literal) which is the content of an element should map to. In other contexts, other rules (such as type coercion rules) may be employed to determine which value is to be used.

##### <a id="sec-ast-f"></a>3.2.1.3 Facets

When a new datatype is defined by [facet-based restriction](#dt-fb-restriction), [anySimpleType](#anySimpleType) **must not** be used as the [base type](#dt-basetype). So no [constraining facets](#dt-constraining-facet) are directly applicable to [anySimpleType](#anySimpleType).

#### <a id="anyAtomicType"></a>3.2.2 anyAtomicType

<a id="dt-anyAtomicType">[Definition:]</a> **anyAtomicType** is a special [restriction](#dt-restriction) of [anySimpleType](#anySimpleType). The [value](#dt-value-space) and [lexical spaces](#dt-lexical-space) of **anyAtomicType** are the unions of the [value](#dt-value-space) and [lexical spaces](#dt-lexical-space) of all the [primitive](#dt-primitive) datatypes, and **anyAtomicType** is their [base type](#dt-basetype).

For further details of [anyAtomicType](#anyAtomicType) and its representation as a [Simple Type Definition](#std), see [Built-in Simple Type Definitions (§4.1.6)](#builtin-stds).

##### <a id="sec-aat-vs"></a>3.2.2.1 Value space

The [value space](#dt-value-space) of [anyAtomicType](#anyAtomicType) is the union of the [value spaces](#dt-value-space) of all the [primitive](#dt-primitive) datatypes defined here or supplied as [implementation-defined](#key-impl-def) [primitives](#dt-primitive).

##### <a id="sec-aat-lex"></a>3.2.2.2 Lexical mapping

The [lexical space](#dt-lexical-space) of [anyAtomicType](#anyAtomicType) is the set of all finite-length sequences of zero or more [character](https://www.w3.org/TR/xml11/#dt-character)s (as defined in [[XML]](#XML)) that [match](#dt-match) the [Char](https://www.w3.org/TR/xml11/#NT-Char) production from [[XML]](#XML). This is equivalent to the union of the [lexical spaces](#dt-lexical-space) of all [primitive](#dt-primitive) datatypes.

It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports the [Char](https://www.w3.org/TR/xml11/#NT-Char) production from [[XML]](#XML), or that from [[XML 1.0]](#XML1.0), or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

The [lexical mapping](#dt-lexical-mapping) of [anyAtomicType](#anyAtomicType) is the union of the [lexical mappings](#dt-lexical-mapping) of all [primitive](#dt-primitive) datatypes. It will be noted that this mapping is not a function: a given [literal](#dt-literal) may map to one value or to several values of different [primitive](#dt-primitive) datatypes, and it may be indeterminate which value is to be preferred in a particular context. When the datatypes defined here are used in the context of [[XSD 1.1 Part 1: Structures]](#structural-schemas), the `xsi:type` attribute defined by that specification in section [xsi:type](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#xsi_type) can be used to indicate which value a [literal](#dt-literal) which is the content of an element should map to. In other contexts, other rules (such as type coercion rules) may be employed to determine which value is to be used.

##### <a id="sec-aat-f"></a>3.2.2.3 Facets

When a new datatype is defined by [facet-based restriction](#dt-fb-restriction), [anyAtomicType](#anyAtomicType) **must not** be used as the [base type](#dt-basetype). So no [constraining facets](#dt-constraining-facet) are directly applicable to [anyAtomicType](#anyAtomicType).

### <a id="built-in-primitive-datatypes"></a>3.3 Primitive Datatypes

        3.3.1 [string](#string)
             3.3.1.1 [Value Space](#sec-vs-string)
             3.3.1.2 [Lexical Mapping](#string-lexical-mapping)
             3.3.1.3 [Facets](#string-facets)
             3.3.1.4 [Derived datatypes](#string-derived-types)
         3.3.2 [boolean](#boolean)
             3.3.2.1 [Value Space](#sec-vs-boolean)
             3.3.2.2 [Lexical Mapping](#boolean-lexical-mapping)
             3.3.2.3 [Facets](#boolean-facets)
         3.3.3 [decimal](#decimal)
             3.3.3.1 [Lexical Mapping](#decimal-lexical-representation)
             3.3.3.2 [Facets](#decimal-facets)
             3.3.3.3 [Datatypes based on decimal](#decimal-derived-types)
         3.3.4 [float](#float)
             3.3.4.1 [Value Space](#sec-vs-float)
             3.3.4.2 [Lexical Mapping](#sec-lex-float)
             3.3.4.3 [Facets](#float-facets)
         3.3.5 [double](#double)
             3.3.5.1 [Value Space](#sec-vs-double)
             3.3.5.2 [Lexical Mapping](#sec-lex-double)
             3.3.5.3 [Facets](#double-facets)
         3.3.6 [duration](#duration)
             3.3.6.1 [Value Space](#sec-vs-duration)
             3.3.6.2 [Lexical Mapping](#duration-lexical-space)
             3.3.6.3 [Facets](#duration-facets)
             3.3.6.4 [Related Datatypes](#duration-derived-types)
         3.3.7 [dateTime](#dateTime)
             3.3.7.1 [Value Space](#dateTime-value-space)
             3.3.7.2 [Lexical Mapping](#dateTime-lexical-mapping)
             3.3.7.3 [Facets](#dateTime-facets)
             3.3.7.4 [Related Datatypes](#dateTime-derived-types)
         3.3.8 [time](#time)
             3.3.8.1 [Value Space](#time-value-space)
             3.3.8.2 [Lexical Mappings](#time-lexical-mapping)
             3.3.8.3 [Facets](#time-facets)
         3.3.9 [date](#date)
             3.3.9.1 [Value Space](#date-value-space)
             3.3.9.2 [Lexical Mapping](#date-lexical-mapping)
             3.3.9.3 [Facets](#date-facets)
         3.3.10 [gYearMonth](#gYearMonth)
             3.3.10.1 [Value Space](#gYearMonth-value-space)
             3.3.10.2 [Lexical Mapping](#gYearMonth-lexical-repr)
             3.3.10.3 [Facets](#gYearMonth-facets)
         3.3.11 [gYear](#gYear)
             3.3.11.1 [Value Space](#gYear-value-space)
             3.3.11.2 [Lexical Mapping](#gYear-lexical-repr)
             3.3.11.3 [Facets](#gYear-facets)
         3.3.12 [gMonthDay](#gMonthDay)
             3.3.12.1 [Value Space](#gMonthDay-value-space)
             3.3.12.2 [Lexical Mapping](#gMonthDay-lexical-repr)
             3.3.12.3 [Facets](#gMonthDay-facets)
         3.3.13 [gDay](#gDay)
             3.3.13.1 [Value Space](#sec-vs-gDay)
             3.3.13.2 [Lexical Mapping](#gDay-lexical-mapping)
             3.3.13.3 [Facets](#gDay-facets)
         3.3.14 [gMonth](#gMonth)
             3.3.14.1 [Value Space](#gMonth-value-space)
             3.3.14.2 [Lexical Mapping](#gMonth-lexical-repr)
             3.3.14.3 [Facets](#gMonth-facets)
         3.3.15 [hexBinary](#hexBinary)
             3.3.15.1 [Value Space](#sec-vs-hexbin)
             3.3.15.2 [Lexical Mapping](#hexBinary-lexical-representation)
             3.3.15.3 [Facets](#hexBinary-facets)
         3.3.16 [base64Binary](#base64Binary)
             3.3.16.1 [Value Space](#sec-vs-b46b)
             3.3.16.2 [Lexical Mapping](#sec-lex-b64b)
             3.3.16.3 [Facets](#base64Binary-facets)
         3.3.17 [anyURI](#anyURI)
             3.3.17.1 [Value Space](#anyURI-vs)
             3.3.17.2 [Lexical Mapping](#anyURI-lexical-representation)
             3.3.17.3 [Facets](#anyURI-facets)
         3.3.18 [QName](#QName)
             3.3.18.1 [Facets](#QName-facets)
         3.3.19 [NOTATION](#NOTATION)
             3.3.19.1 [Facets](#NOTATION-facets)

 The [primitive](#dt-primitive) datatypes defined by this specification are described below.  For each datatype, the [value space](#dt-value-space) is described; the [lexical space](#dt-lexical-space) is defined using an extended Backus Naur Format grammar (and in most cases also a regular expression using the regular expression language of [Regular Expressions (§G)](#regexs)); [constraining facets](#dt-constraining-facet) which apply to the datatype are listed; and any datatypes [constructed](#dt-constructed) from this datatype are specified.

 Conforming processors **must** support the [primitive](#dt-primitive) datatypes defined in this specification; it is [implementation-defined](#key-impl-def) whether they support others. [Primitive](#dt-primitive) datatypes may be added by revisions to this specification.

**Note:** Processors **may**, for example, support the floating-point decimal datatype specified in [[Precision Decimal]](#pd-note).

#### <a id="string"></a>3.3.1 string

<a id="dt-string">[Definition:]</a>The **string** datatype represents character strings in XML.

**Note:** Many human languages have writing systems that require child elements for control of aspects such as bidirectional formatting or ruby annotation (see [[Ruby]](#ruby) and Section 8.2.4 [Overriding the bidirectional algorithm: the BDO element](https://www.w3.org/TR/html401/struct/dirlang.html#h-8.2.4) of [[HTML 4.01]](#html4)).  Thus, [string](#string), as a simple type that can contain only characters but not child elements, is often not suitable for representing text. In such situations, a complex type that allows mixed content should be considered. For more information, see Section 5.5 [Any Element, Any Attribute](https://www.w3.org/TR/2001/REC-xmlschema-0-20010502/#textType) of [[XML Schema Language: Part 0 Primer]](#schema-primer).

##### <a id="sec-vs-string"></a>3.3.1.1 Value Space

The [value space](#dt-value-space) of [string](#string) is the set of finite-length sequences of zero or more [character](https://www.w3.org/TR/xml11/#dt-character)s (as defined in [[XML]](#XML)) that [match](#dt-match) the [Char](https://www.w3.org/TR/xml11/#NT-Char) production from [[XML]](#XML). A [character](https://www.w3.org/TR/xml11/#dt-character) is an atomic unit of communication; it is not further specified except to note that every [character](https://www.w3.org/TR/xml11/#dt-character) has a corresponding Universal Character Set (UCS) code point, which is an integer.

It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports the [Char](https://www.w3.org/TR/xml11/#NT-Char) production from [[XML]](#XML), or that from [[XML 1.0]](#XML1.0), or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

Equality for [string](#string) is identity. No order is prescribed.

**Note:** As noted in [ordered](#ff-o), the fact that this specification does not specify an order relation for [string](#dt-string) does not preclude other applications from treating strings as being ordered.

##### <a id="string-lexical-mapping"></a>3.3.1.2 Lexical Mapping

The [lexical space](#dt-lexical-space) of [string](#string) is the set of finite-length sequences of zero or more [character](https://www.w3.org/TR/xml11/#dt-character)s (as defined in [[XML]](#XML)) that [match](#dt-match) the [Char](https://www.w3.org/TR/xml11/#NT-Char) production from [[XML]](#XML).  Lexical Space  <a id="nt-stringRep"></a>[1]   *stringRep* ::= [Char](https://www.w3.org/TR/xml11/#NT-Char)*  /* *(as defined in [[XML]](#XML))* */

It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports the [Char](https://www.w3.org/TR/xml11/#NT-Char) production from [[XML]](#XML), or that from [[XML 1.0]](#XML1.0), or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

The [lexical mapping](#dt-lexical-mapping) for [string](#string) is [*stringLexicalMap*](#f-stringLexmap), and the [canonical mapping](#dt-canonical-mapping) is [*stringCanonicalMap*](#f-stringCanmap); each is a subset of the identity function.

##### <a id="string-facets"></a>3.3.1.3 Facets

The [string](#string) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [whiteSpace](#rf-whiteSpace) = ***preserve***

Datatypes derived by restriction from [string](#string) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [string](#string) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

##### <a id="string-derived-types"></a>3.3.1.4 Derived datatypes

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [string](#string)

- [normalizedString](#normalizedString)

#### <a id="boolean"></a>3.3.2 boolean

<a id="dt-boolean">[Definition:]</a>**boolean** represents the values of two-valued logic.

##### <a id="sec-vs-boolean"></a>3.3.2.1 Value Space

[boolean](#boolean) has the [value space](#dt-value-space) of two-valued logic:  {***true***, ***false***}.

##### <a id="boolean-lexical-mapping"></a>3.3.2.2 Lexical Mapping

[boolean](#boolean)'s lexical space is a set of four [literals](#dt-literal):  Lexical Space  <a id="nt-booleanRep"></a>[2]   *booleanRep* ::= '`true`' | '`false`' | '`1`' | '`0`'

The [lexical mapping](#dt-lexical-mapping) for [boolean](#boolean) is [*booleanLexicalMap*](#f-booleanLexmap); the [canonical mapping](#dt-canonical-mapping) is [*booleanCanonicalMap*](#f-booleanCanmap).

##### <a id="boolean-facets"></a>3.3.2.3 Facets

The [boolean](#boolean) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

Datatypes derived by restriction from [boolean](#boolean) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [assertions](#rf-assertions)

The [boolean](#boolean) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***finite***
- [numeric](#rf-numeric) = ***false***

#### <a id="decimal"></a>3.3.3 decimal

<a id="dt-decimal-datatype">[Definition:]</a>**decimal** represents a subset of the real numbers, which can be represented by decimal numerals. The [value space](#dt-value-space) of **decimal** is the set of numbers that can be obtained by dividing an integer by a non-negative power of ten, i.e., expressible as *i* / 10^*n*^ where *i* and *n* are integers and *n* ≥ 0. Precision is not reflected in this value space; the number 2.0 is not distinct from the number 2.00. The order relation on **decimal** is the order relation on real numbers, restricted to this subset.

**Note:** For a decimal datatype whose values do reflect precision, see [[Precision Decimal]](#pd-note).

##### <a id="decimal-lexical-representation"></a>3.3.3.1 Lexical Mapping

**decimal** has a lexical representation consisting of a non-empty finite-length sequence of decimal digits (#x30–#x39) separated by a period as a decimal indicator.  An optional leading sign is allowed.  If the sign is omitted, "+" is assumed.  Leading and trailing zeroes are optional.  If the fractional part is zero, the period and following zero(es) can be omitted. For example:  '`-1.23`', '`12678967.543233`', '`+100000.00`', '`210`'.

 The [decimal](#decimal) Lexical Representation  <a id="nt-decimalRep"></a>[3]   *decimalLexicalRep* ::= [*decimalPtNumeral*](#nt-decNuml) | [*noDecimalPtNumeral*](#nt-noDecNuml) The lexical space of decimal is the set of lexical representations which match the grammar given above, or (equivalently) the regular expression

> > `(\+|-)?([0-9]+(\.[0-9]*)?|\.[0-9]+)`

 The mapping from lexical representations to values is the usual one for decimal numerals; it is given formally in [*decimalLexicalMap*](#f-decimalLexmap).

The definition of the [canonical representation](#dt-canonical-representation) has the effect of prohibiting certain options from the [Lexical Mapping (§3.3.3.1)](#decimal-lexical-representation).  Specifically, for integers, the decimal point and fractional part are prohibited. For other values, the preceding optional "+" sign is prohibited.  The decimal point is required.  In all cases, leading and trailing zeroes are prohibited subject to the following:  there **must** be at least one digit to the right and to the left of the decimal point which may be a zero.

 The mapping from values to [canonical representations](#dt-canonical-representation) is given formally in [*decimalCanonicalMap*](#f-decimalCanmap).

##### <a id="decimal-facets"></a>3.3.3.2 Facets

The [decimal](#decimal) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

Datatypes derived by restriction from [decimal](#decimal) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [fractionDigits](#rf-fractionDigits)
- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [decimal](#decimal) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***true***

##### <a id="decimal-derived-types"></a>3.3.3.3 Datatypes based on decimal

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [decimal](#decimal)

- [integer](#integer)

#### <a id="float"></a>3.3.4 float

<a id="dt-float">[Definition:]</a>The **float** datatype is patterned after the IEEE single-precision 32-bit floating point datatype [[IEEE 754-2008]](#ieee754-2008).  Its value space is a subset of the rational numbers.  Floating point numbers are often used to approximate arbitrary real numbers.

##### <a id="sec-vs-float"></a>3.3.4.1 Value Space

The [value space](#dt-value-space) of [float](#float) contains the non-zero numbers  *m* × 2^*e*^ , where *m* is an integer whose absolute value is less than 2^24^, and *e* is an integer between −149 and 104, inclusive.  In addition to these values, the [value space](#dt-value-space) of [float](#float) also contains the following [special values](#dt-specialvalue):  ***positiveZero***, ***negativeZero***, ***positiveInfinity***, ***negativeInfinity***, and ***notANumber***.

**Note:** As explained below, the [lexical representation](#dt-lexical-representation) of the [float](#float) value ***notANumber*** is '`NaN`'.  Accordingly, in English text we generally use 'NaN' to refer to that value.  Similarly, we use 'INF' and '−INF' to refer to the two values ***positiveInfinity*** and ***negativeInfinity***, and '0' and '−0' to refer to ***positiveZero*** and ***negativeZero***.Equality and order for [float](#float) are defined as follows:

- Equality is identity, except that  0 = −0  (although they are not identical) and  NaN ≠ NaN  (although NaN is of course identical to itself).0 and −0 are thus equivalent for purposes of enumerations and identity constraints, as well as for minimum and maximum values.
- For the basic values, the order relation on float is the order relation for rational numbers.  INF is greater than all other non-NaN values; −INF is less than all other non-NaN values.  NaN is [incomparable](#dt-incomparable) with any value in the [value space](#dt-value-space) including itself.  0 and −0 are greater than all the negative numbers and less than all the positive numbers.
 **Note:** Any value [incomparable](#dt-incomparable) with the value used for the four bounding facets ([minInclusive](#dt-minInclusive), [maxInclusive](#dt-maxInclusive), [minExclusive](#dt-minExclusive), and [maxExclusive](#dt-maxExclusive)) will be excluded from the resulting restricted [value space](#dt-value-space).  In particular, when NaN is used as a facet value for a bounding facet, since no [float](#float) values are [comparable](#dt-incomparable) with it, the result is a [value space](#dt-value-space) that is empty.  If any other value is used for a bounding facet, NaN will be excluded from the resulting restricted [value space](#dt-value-space); to add NaN back in requires union with the NaN-only space (which may be derived using the pattern '`NaN`').**Note:** The Schema 1.0 version of this datatype did not differentiate between 0 and −0 and NaN was equal to itself.  The changes were made to make the datatype more closely mirror [[IEEE 754-2008]](#ieee754-2008).

##### <a id="sec-lex-float"></a>3.3.4.2 Lexical Mapping

The [lexical space](#dt-lexical-space) of [float](#float) is the set of all decimal numerals with or without a decimal point, numerals in scientific (exponential) notation, and the [literals](#dt-literal) '`INF`', '`+INF`', '`-INF`', and '`NaN`'  Lexical Space  <a id="nt-floatRep"></a>[4]   *floatRep* ::= [*noDecimalPtNumeral*](#nt-noDecNuml) | [*decimalPtNumeral*](#nt-decNuml) | [*scientificNotationNumeral*](#nt-sciNuml) | [*numericalSpecialRep*](#nt-numSpecReps) The [*floatRep*](#nt-floatRep) production is equivalent to this regular expression (after whitespace is removed from the regular expression):

> `(\+|-)?([0-9]+(\.[0-9]*)?|\.[0-9]+)([Ee](\+|-)?[0-9]+)?
>  |(\+|-)?INF|NaN`

The [float](#float) datatype is designed to implement for schema processing the single-precision floating-point datatype of [[IEEE 754-2008]](#ieee754-2008).  That specification does not specify specific [lexical representations](#dt-lexical-representation), but does prescribe requirements on any [lexical mapping](#dt-lexical-mapping) used.  Any [lexical mapping](#dt-lexical-mapping) that maps the [lexical space](#dt-lexical-space) just described onto the [value space](#dt-value-space), is a function, satisfies the requirements of [[IEEE 754-2008]](#ieee754-2008), and correctly handles the mapping of the literals '`INF`', '`NaN`', etc., to the [special values](#dt-specialvalue), satisfies the conformance requirements of this specification.

Since IEEE allows some variation in rounding of values, processors conforming to this specification may exhibit some variation in their [lexical mappings](#dt-lexical-mapping).

The [lexical mapping](#dt-lexical-mapping) [*floatLexicalMap*](#f-floatLexmap) is provided as an example of a simple algorithm that yields a conformant mapping, and that provides the most accurate rounding possible—and is thus useful for insuring inter-implementation reproducibility and inter-implementation round-tripping.  The simple rounding algorithm used in [*floatLexicalMap*](#f-floatLexmap) may be more efficiently implemented using the algorithms of [[Clinger, WD (1990)]](#clinger1990).

**Note:** The Schema 1.0 version of this datatype did not permit rounding algorithms whose results differed from [[Clinger, WD (1990)]](#clinger1990).

The [canonical mapping](#dt-canonical-mapping) [*floatCanonicalMap*](#f-floatCanmap) is provided as an example of a mapping that does not produce unnecessarily long [canonical representations](#dt-canonical-representation).  Other algorithms which do not yield identical results for mapping from float values to character strings are permitted by [[IEEE 754-2008]](#ieee754-2008).

##### <a id="float-facets"></a>3.3.4.3 Facets

The [float](#float) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

Datatypes derived by restriction from [float](#float) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [float](#float) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***true***
- [cardinality](#rf-cardinality) = ***finite***
- [numeric](#rf-numeric) = ***true***

#### <a id="double"></a>3.3.5 double

 <a id="dt-double">[Definition:]</a>The **double** datatype is patterned after the IEEE double-precision 64-bit floating point datatype [[IEEE 754-2008]](#ieee754-2008).  Each floating point datatype has a value space that is a subset of the rational numbers.  Floating point numbers are often used to approximate arbitrary real numbers.

**Note:** The only significant differences between float and double are the three defining constants 53 (vs 24), −1074 (vs −149), and 971 (vs 104).

##### <a id="sec-vs-double"></a>3.3.5.1 Value Space

The [value space](#dt-value-space) of [double](#double) contains the non-zero numbers  *m* × 2^*e*^ , where *m* is an integer whose absolute value is less than 2^53^, and *e* is an integer between −1074 and 971, inclusive.  In addition to these values, the [value space](#dt-value-space) of [double](#double) also contains the following [special values](#dt-specialvalue):  ***positiveZero***, ***negativeZero***, ***positiveInfinity***, ***negativeInfinity***, and ***notANumber***.

**Note:** As explained below, the [lexical representation](#dt-lexical-representation) of the [double](#double) value ***notANumber*** is '`NaN`'.  Accordingly, in English text we generally use 'NaN' to refer to that value.  Similarly, we use 'INF' and '−INF' to refer to the two values ***positiveInfinity*** and ***negativeInfinity***, and '0' and '−0' to refer to ***positiveZero*** and ***negativeZero***.Equality and order for [double](#double) are defined as follows:

- Equality is identity, except that  0 = −0  (although they are not identical) and  NaN ≠ NaN  (although NaN is of course identical to itself).0 and −0 are thus equivalent for purposes of enumerations, identity constraints, and minimum and maximum values.
- For the basic values, the order relation on double is the order relation for rational numbers.  INF is greater than all other non-NaN values; −INF is less than all other non-NaN values.  NaN is [incomparable](#dt-incomparable) with any value in the [value space](#dt-value-space) including itself.  0 and −0 are greater than all the negative numbers and less than all the positive numbers.
 **Note:** Any value [incomparable](#dt-incomparable) with the value used for the four bounding facets ([minInclusive](#dt-minInclusive), [maxInclusive](#dt-maxInclusive), [minExclusive](#dt-minExclusive), and [maxExclusive](#dt-maxExclusive)) will be excluded from the resulting restricted [value space](#dt-value-space).  In particular, when NaN is used as a facet value for a bounding facet, since no [double](#double) values are [comparable](#dt-incomparable) with it, the result is a [value space](#dt-value-space) that is empty.  If any other value is used for a bounding facet, NaN will be excluded from the resulting restricted [value space](#dt-value-space); to add NaN back in requires union with the NaN-only space (which may be derived using the pattern '`NaN`').**Note:** The Schema 1.0 version of this datatype did not differentiate between 0 and −0 and NaN was equal to itself.  The changes were made to make the datatype more closely mirror [[IEEE 754-2008]](#ieee754-2008).

##### <a id="sec-lex-double"></a>3.3.5.2 Lexical Mapping

The [lexical space](#dt-lexical-space) of [double](#double) is the set of all decimal numerals with or without a decimal point, numerals in scientific (exponential) notation, and the [literals](#dt-literal) '`INF`', '`+INF`', '`-INF`', and '`NaN`'  Lexical Space  <a id="nt-doubleRep"></a>[5]   *doubleRep* ::= [*noDecimalPtNumeral*](#nt-noDecNuml) | [*decimalPtNumeral*](#nt-decNuml) | [*scientificNotationNumeral*](#nt-sciNuml) | [*numericalSpecialRep*](#nt-numSpecReps) The [*doubleRep*](#nt-doubleRep) production is equivalent to this regular expression (after whitespace is eliminated from the expression):

> `(\+|-)?([0-9]+(\.[0-9]*)?|\.[0-9]+)([Ee](\+|-)?[0-9]+)? |(\+|-)?INF|NaN`

The [double](#double) datatype is designed to implement for schema processing the double-precision floating-point datatype of [[IEEE 754-2008]](#ieee754-2008).  That specification does not specify specific [lexical representations](#dt-lexical-representation), but does prescribe requirements on any [lexical mapping](#dt-lexical-mapping) used.  Any [lexical mapping](#dt-lexical-mapping) that maps the [lexical space](#dt-lexical-space) just described onto the [value space](#dt-value-space), is a function, satisfies the requirements of [[IEEE 754-2008]](#ieee754-2008), and correctly handles the mapping of the literals '`INF`', '`NaN`', etc., to the [special values](#dt-specialvalue), satisfies the conformance requirements of this specification.

Since IEEE allows some variation in rounding of values, processors conforming to this specification may exhibit some variation in their [lexical mappings](#dt-lexical-mapping).

The [lexical mapping](#dt-lexical-mapping) [*doubleLexicalMap*](#f-doubleLexmap) is provided as an example of a simple algorithm that yields a conformant mapping, and that provides the most accurate rounding possible—and is thus useful for insuring inter-implementation reproducibility and inter-implementation round-tripping.  The simple rounding algorithm used in [*doubleLexicalMap*](#f-doubleLexmap) may be more efficiently implemented using the algorithms of [[Clinger, WD (1990)]](#clinger1990).

**Note:** The Schema 1.0 version of this datatype did not permit rounding algorithms whose results differed from [[Clinger, WD (1990)]](#clinger1990).

The [canonical mapping](#dt-canonical-mapping) [*doubleCanonicalMap*](#f-doubleCanmap) is provided as an example of a mapping that does not produce unnecessarily long [canonical representations](#dt-canonical-representation).  Other algorithms which do not yield identical results for mapping from float values to character strings are permitted by [[IEEE 754-2008]](#ieee754-2008).

##### <a id="double-facets"></a>3.3.5.3 Facets

The [double](#double) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

Datatypes derived by restriction from [double](#double) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [double](#double) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***true***
- [cardinality](#rf-cardinality) = ***finite***
- [numeric](#rf-numeric) = ***true***

#### <a id="duration"></a>3.3.6 duration

 <a id="dt-duration">[Definition:]</a>**duration** is a datatype that represents durations of time.  The concept of duration being captured is drawn from those of [[ISO 8601]](#ISO8601), specifically *durations without fixed endpoints*.  For example, "15 days" (whose most common lexical representation in [duration](#duration) is "'`P15D`'") is a [duration](#duration) value; "15 days beginning 12 July 1995" and "15 days ending 12 July 1995" are not [duration](#duration) values.  [duration](#duration) can provide addition and subtraction operations between [duration](#duration) values and between [duration](#duration)/[dateTime](#dateTime) value pairs, and can be the result of subtracting [dateTime](#dateTime) values.  However, only addition to [dateTime](#dateTime) is required for XML Schema processing and is defined in the function [*dateTimePlusDuration*](#vp-dt-dateTimePlusDuration).

##### <a id="sec-vs-duration"></a>3.3.6.1 Value Space

 Duration values can be modelled as two-property tuples. Each value consists of an integer number of months and a decimal number of seconds. The [*seconds*](#vp-du-second) value **must not** be negative if the [*months*](#vp-du-month) value is positive and **must not** be positive if the [*months*](#vp-du-month) is negative.  Properties of [duration](#duration) Values  **<a id="vp-du-month">*months*</a>** [integer](#integer)  **<a id="vp-du-second">*seconds*</a>** a [decimal](#decimal) value; **must not** be negative if [*months*](#vp-du-month) is positive, and **must not** be positive if [*months*](#vp-du-month) is negative.  [duration](#duration) is partially ordered.  Equality of [duration](#duration) is defined in terms of equality of [dateTime](#dateTime); order for [duration](#duration) is defined in terms of the order of [dateTime](#dateTime). Specifically, the equality or order of two [duration](#duration) values is determined by adding each [duration](#duration) in the pair to each of the following four [dateTime](#dateTime) values:

- 1696-09-01T00:00:00Z
- 1697-02-01T00:00:00Z
- 1903-03-01T00:00:00Z
- 1903-07-01T00:00:00Z
 If all four resulting [dateTime](#dateTime) value pairs are ordered the same way (less than, equal, or greater than), then the original pair of [duration](#duration) values is ordered the same way; otherwise the original pair is [incomparable](#dt-incomparable).**Note:** These four values are chosen so as to maximize the possible differences in results that could occur, such as the difference when adding P1M and P30D:  1697-02-01T00:00:00Z + P1M < 1697-02-01T00:00:00Z + P30D , but 1903-03-01T00:00:00Z + P1M > 1903-03-01T00:00:00Z + P30D , so that  P1M <> P30D .  If two [duration](#duration) values are ordered the same way when added to each of these four [dateTime](#dateTime) values, they will retain the same order when added to *any* other [dateTime](#dateTime) values.  Therefore, two [duration](#duration) values are incomparable if and only if they can *ever* result in different orders when added to *any* [dateTime](#dateTime) value.

Under the definition just given, two [duration](#duration) values are equal if and only if they are identical.

<a id="two_totally_ordered_subtypes"></a>**Note:** Two totally ordered datatypes ([yearMonthDuration](#yearMonthDuration) and [dayTimeDuration](#dayTimeDuration)) are derived from [duration](#duration) in [Other Built-in Datatypes (§3.4)](#ordinary-built-ins).**Note:** There are many ways to implement [duration](#duration), some of which do not base the implementation on the two-component model.  This specification does not prescribe any particular implementation, as long as the visible results are isomorphic to those described herein.**Note:** See the conformance notes in [Partial Implementation of Infinite Datatypes (§5.4)](#partial-implementation), which apply to this datatype.

##### <a id="duration-lexical-space"></a>3.3.6.2 Lexical Mapping

The [lexical representations](#dt-lexical-representation) of [duration](#duration) are more or less based on the pattern:

> `P*n*Y*n*M*n*DT*n*H*n*M*n*S`

 More precisely, the [lexical space](#dt-lexical-space) of [duration](#duration) is the set of character strings that satisfy [*durationLexicalRep*](#nt-durationRep) as defined by the following productions:   Lexical Representation Fragments  <a id="nt-duYrFrag"></a>[6]   *duYearFrag* ::= [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) '`Y`' <a id="nt-duMoFrag"></a>[7]   *duMonthFrag* ::= [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) '`M`' <a id="nt-duDaFrag"></a>[8]   *duDayFrag* ::= [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) '`D`' <a id="nt-duHrFrag"></a>[9]   *duHourFrag* ::= [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) '`H`' <a id="nt-duMiFrag"></a>[10]   *duMinuteFrag* ::= [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) '`M`' <a id="nt-duSeFrag"></a>[11]   *duSecondFrag* ::= ([*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) | [*unsignedDecimalPtNumeral*](#nt-unsDecNuml)) '`S`' <a id="nt-duYMFrag"></a>[12]   *duYearMonthFrag* ::= ([*duYearFrag*](#nt-duYrFrag) [*duMonthFrag*](#nt-duMoFrag)?) | [*duMonthFrag*](#nt-duMoFrag) <a id="nt-duTFrag"></a>[13]   *duTimeFrag* ::= '`T`' (([*duHourFrag*](#nt-duHrFrag) [*duMinuteFrag*](#nt-duMiFrag)? [*duSecondFrag*](#nt-duSeFrag)?) | ([*duMinuteFrag*](#nt-duMiFrag) [*duSecondFrag*](#nt-duSeFrag)?) | [*duSecondFrag*](#nt-duSeFrag)) <a id="nt-duDTFrag"></a>[14]   *duDayTimeFrag* ::= ([*duDayFrag*](#nt-duDaFrag) [*duTimeFrag*](#nt-duTFrag)?) | [*duTimeFrag*](#nt-duTFrag)  Lexical Representation  <a id="nt-durationRep"></a>[15]   *durationLexicalRep* ::= '`-`'? '`P`' (([*duYearMonthFrag*](#nt-duYMFrag) [*duDayTimeFrag*](#nt-duDTFrag)?) | [*duDayTimeFrag*](#nt-duDTFrag))

Thus, a [*durationLexicalRep*](#nt-durationRep) consists of one or more of a [*duYearFrag*](#nt-duYrFrag), [*duMonthFrag*](#nt-duMoFrag), [*duDayFrag*](#nt-duDaFrag), [*duHourFrag*](#nt-duHrFrag), [*duMinuteFrag*](#nt-duMiFrag), and/or [*duSecondFrag*](#nt-duSeFrag), in order, with letters '`P`' and '`T`' (and perhaps a '`-`') where appropriate.

The language accepted by the [*durationLexicalRep*](#nt-durationRep) production is the set of strings which satisfy all of the following three regular expressions:

- The expression

> `-?P[0-9]+Y?([0-9]+M)?([0-9]+D)?(T([0-9]+H)?([0-9]+M)?([0-9]+(\.[0-9]+)?S)?)?`

 matches only strings in which the fields occur in the proper order.
- The expression '`.*[YMDHS].*`' matches only strings in which at least one field occurs.
- The expression '`.*[^T]`' matches only strings in which '`T`' is not the final character, so that if '`T`' appears, something follows it. The first rule ensures that what follows '`T`' will be an hour, minute, or second field.
 The intersection of these three regular expressions is equivalent to the following (after removal of the white space inserted here for legibility):

```
-?P( ( ( [0-9]+Y([0-9]+M)?([0-9]+D)?
       | ([0-9]+M)([0-9]+D)?
       | ([0-9]+D)
       )
       (T ( ([0-9]+H)([0-9]+M)?([0-9]+(\.[0-9]+)?S)?
          | ([0-9]+M)([0-9]+(\.[0-9]+)?S)?
          | ([0-9]+(\.[0-9]+)?S)
          )
       )?
    )
  | (T ( ([0-9]+H)([0-9]+M)?([0-9]+(\.[0-9]+)?S)?
       | ([0-9]+M)([0-9]+(\.[0-9]+)?S)?
       | ([0-9]+(\.[0-9]+)?S)
       )
    )
  )

```

The [lexical mapping](#dt-lexical-mapping) for [duration](#duration) is [*durationMap*](#f-durationMap).

[The canonical mapping](#dt-canonical-mapping) for [duration](#duration) is [*durationCanonicalMap*](#f-durationCanMap).

##### <a id="duration-facets"></a>3.3.6.3 Facets

The [duration](#duration) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

Datatypes derived by restriction from [duration](#duration) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [duration](#duration) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

##### <a id="duration-derived-types"></a>3.3.6.4 Related Datatypes

 The following [built-in](#dt-built-in) datatypes are [derived](#dt-derived) from [duration](#duration)

- [yearMonthDuration](#yearMonthDuration)
- [dayTimeDuration](#dayTimeDuration)

#### <a id="dateTime"></a>3.3.7 dateTime

[dateTime](#dateTime) represents instants of time, optionally marked with a particular time zone offset.  Values representing the same instant but having different time zone offsets are equal but not identical.

##### <a id="dateTime-value-space"></a>3.3.7.1 Value Space

[dateTime](#dateTime) uses the [date/timeSevenPropertyModel](#dt-dt-7PropMod), with no properties except [*timezoneOffset*](#vp-dt-timezone) permitted to be ***absent***. The [*timezoneOffset*](#vp-dt-timezone) property remains [optional](#dt-optional).

**Note:** In version 1.0 of this specification, the [*year*](#vp-dt-year) property was not permitted to have the value zero. The year before the year 1 in the proleptic Gregorian calendar, traditionally referred to as 1 BC or as 1 BCE, was represented by a [*year*](#vp-dt-year) value of −1, 2 BCE by −2, and so forth. Of course, many, perhaps most, references to 1 BCE (or 1 BC) actually refer not to a year in the proleptic Gregorian calendar but to a year in the Julian or "old style" calendar; the two correspond approximately but not exactly to each other.  In this version of this specification, two changes are made in order to agree with existing usage. First, [*year*](#vp-dt-year) is permitted to have the value zero. Second, the interpretation of [*year*](#vp-dt-year) values is changed accordingly: a [*year*](#vp-dt-year) value of zero represents 1 BCE, −1 represents 2 BCE, etc. This representation simplifies interval arithmetic and leap-year calculation for dates before the common era (which may be why astronomers and others interested in such calculations with the proleptic Gregorian calendar have adopted it), and is consistent with the current edition of [[ISO 8601]](#ISO8601).  Note that 1 BCE, 5 BCE, and so on (years 0000, -0004, etc. in the lexical representation defined here) are leap years in the proleptic Gregorian calendar used for the date/time datatypes defined here. Version 1.0 of this specification was unclear about the treatment of leap years before the common era. If existing schemas or data specify dates of 29 February for any years before the common era, then some values giving a date of 29 February which were valid under a plausible interpretation of XSD 1.0 will be invalid under this specification, and some which were invalid will be valid. With that possible exception, schemas and data valid under the old interpretation remain valid under the new. <a id="con-dateTime-dayValue"></a>**Constraint: Day-of-month Values**

> The [*day*](#vp-dt-day) value **must** be no more than 30 if [*month*](#vp-dt-month) is one of 4, 6, 9, or 11; no more than 28 if [*month*](#vp-dt-month) is 2 and [*year*](#vp-dt-year) is not divisible by 4, or is divisible by 100 but not by 400; and no more than 29 if [*month*](#vp-dt-month) is 2 and [*year*](#vp-dt-year) is divisible by 400, or by 4 but not by 100.

**Note:** See the conformance note in [Partial Implementation of Infinite Datatypes (§5.4)](#partial-implementation) which applies to the [*year*](#vp-dt-year) and [*second*](#vp-dt-second) values of this datatype.

Equality and order are as prescribed in [The Seven-property Model (§D.2.1)](#theSevenPropertyModel).  [dateTime](#dateTime) values are ordered by their [*timeOnTimeline*](#vp-dt-timeOnTimeline) value.

**Note:** Since the order of a [dateTime](#dateTime) value having a [*timezoneOffset*](#vp-dt-timezone) relative to another value whose [*timezoneOffset*](#vp-dt-timezone) is ***absent*** is determined by imputing time zone offsets of both +14:00 and −14:00 to the value with no time zone offset, many such combinations will be [incomparable](#dt-incomparable) because the two imputed time zone offsets yield different orders.Although [dateTime](#dateTime) and other types related to dates and times have only a partial order, it is possible for datatypes derived from [dateTime](#dateTime) to have total orders, if they are restricted (e.g. using the [pattern](#f-p) facet) to the subset of values with, or the subset of values without, time zone offsets. Similar restrictions on other date- and time-related types will similarly produce totally ordered subtypes. Note, however, that such restrictions do not affect the value shown, for a given [Simple Type Definition](#std), in the [ordered](#ff-o) facet.**Note:** Order and equality are essentially the same for [dateTime](#dateTime) in this version of this specification as they were in version 1.0.  However, since values now distinguish time zone offsets, equal values with different [*timezoneOffset*](#vp-dt-timezone)s are not *identical*, and values with extreme [*timezoneOffset*](#vp-dt-timezone)s may no longer be equal to any value with a smaller [*timezoneOffset*](#vp-dt-timezone).

##### <a id="dateTime-lexical-mapping"></a>3.3.7.2 Lexical Mapping

The lexical representations for [dateTime](#dateTime) are as follows:  Lexical Space  <a id="nt-dateTimeRep"></a>[16]   *dateTimeLexicalRep* ::= [*yearFrag*](#nt-yrFrag) '`-`' [*monthFrag*](#nt-moFrag) '`-`' [*dayFrag*](#nt-daFrag) '`T`' (([*hourFrag*](#nt-hrFrag) '`:`' [*minuteFrag*](#nt-miFrag) '`:`' [*secondFrag*](#nt-seFrag)) | [*endOfDayFrag*](#nt-eodFrag)) [*timezoneFrag*](#nt-tzFrag)?   **Constraint:**  Day-of-month Representations <a id="con-dateTime-day"></a>**Constraint: Day-of-month Representations**

> Within a [*dateTimeLexicalRep*](#nt-dateTimeRep), a [*dayFrag*](#nt-daFrag) **must not** begin with the digit '`3`' or be '`29`' unless the value to which it would map would satisfy the value constraint on [*day*](#vp-dt-day) values ("Constraint: Day-of-month Values") given above.

 In such representations:

- [*yearFrag*](#nt-yrFrag) is a numeral consisting of at least four decimal digits, optionally preceded by a minus sign; leading '`0`' digits are prohibited except to bring the digit count up to four.  It represents the [*year*](#vp-dt-year) value.
- Subsequent '`-`', '`T`', and '`:`', separate the various numerals.
- [*monthFrag*](#nt-moFrag), [*dayFrag*](#nt-daFrag), [*hourFrag*](#nt-hrFrag), and [*minuteFrag*](#nt-miFrag) are numerals consisting of exactly two decimal digits.  They represent the [*month*](#vp-dt-month), [*day*](#vp-dt-day), [*hour*](#vp-dt-hour), and [*minute*](#vp-dt-minute) values respectively.
- [*secondFrag*](#nt-seFrag) is a numeral consisting of exactly two decimal digits, or two decimal digits, a decimal point, and one or more trailing digits.  It represents the [*second*](#vp-dt-second) value.
- Alternatively, [*endOfDayFrag*](#nt-eodFrag) combines the [*hourFrag*](#nt-hrFrag), [*minuteFrag*](#nt-miFrag), [*minuteFrag*](#nt-miFrag), and their separators to represent midnight of the day, which is the first moment of the next day.
- [*timezoneFrag*](#nt-tzFrag), if present, specifies an offset between UTC and local time. Time zone offsets are a count of minutes (expressed in [*timezoneFrag*](#nt-tzFrag) as a count of hours and minutes) that are added or subtracted from UTC time to get the "local" time.  '`Z`' is an alternative representation of the time zone offset '`00:00`', which is, of course, zero minutes from UTC.For example, 2002-10-10T12:00:00−05:00 (noon on 10 October 2002, Central Daylight Savings Time as well as Eastern Standard Time in the U.S.) is equal to 2002-10-10T17:00:00Z, five hours later than 2002-10-10T12:00:00Z.**Note:** For the most part, this specification adopts the distinction between 'timezone' and 'timezone offset' laid out in [[Timezones]](#ref-timezones). Version 1.0 of this specification did not make this distinction, but used the term 'timezone' for the time zone offset information associated with date- and time-related datatypes. Some traces of the earlier usage remain visible in this and other specifications. The names [*timezoneFrag*](#nt-tzFrag) and [explicitTimezone](#f-tz) are such traces ; others will be found in the names of functions defined in [[XQuery 1.0 and XPath 2.0 Functions and Operators]](#F_O), or in references in this specification to "timezoned" and "non-timezoned" values.
 The [*dateTimeLexicalRep*](#nt-dateTimeRep) production is equivalent to this regular expression once whitespace is removed.

```
-?([1-9][0-9]{3,}|0[0-9]{3})
-(0[1-9]|1[0-2])
-(0[1-9]|[12][0-9]|3[01])
T(([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]+)?|(24:00:00(\.0+)?))
(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))?
```

 Note that neither the [*dateTimeLexicalRep*](#nt-dateTimeRep) production nor this regular expression alone enforce the constraint on [*dateTimeLexicalRep*](#nt-dateTimeRep) given above.

The [lexical mapping](#dt-lexical-mapping) for [dateTime](#dateTime) is [*dateTimeLexicalMap*](#vp-dateTimeLexRep). The [canonical mapping](#dt-canonical-mapping) is [*dateTimeCanonicalMap*](#vp-dateTimeCanRep).

##### <a id="dateTime-facets"></a>3.3.7.3 Facets

The [dateTime](#dateTime) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [dateTime](#dateTime) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [explicitTimezone](#rf-explicitTimezone) = ***optional***

Datatypes derived by restriction from [dateTime](#dateTime) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [dateTime](#dateTime) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

##### <a id="dateTime-derived-types"></a>3.3.7.4 Related Datatypes

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [dateTime](#dateTime)

- [dateTimeStamp](#dateTimeStamp)

#### <a id="time"></a>3.3.8 time

[time](#time) represents instants of time that recur at the same point in each calendar day, or that occur in some arbitrary calendar day.

##### <a id="time-value-space"></a>3.3.8.1 Value Space

[time](#time) uses the [date/timeSevenPropertyModel](#dt-dt-7PropMod), with [*year*](#vp-dt-year), [*month*](#vp-dt-month), and [*day*](#vp-dt-day) required to be ***absent***.  [*timezoneOffset*](#vp-dt-timezone) remains [optional](#dt-optional).

**Note:** See the conformance note in [Partial Implementation of Infinite Datatypes (§5.4)](#partial-implementation) which applies to the [*second*](#vp-dt-second) value of this datatype.

Equality and order are as prescribed in [The Seven-property Model (§D.2.1)](#theSevenPropertyModel).  [time](#time) values (points in time in an "arbitrary" day) are ordered taking into account their [*timezoneOffset*](#vp-dt-timezone).

A calendar (or "local time") day with a larger positive time zone offset begins earlier than the same calendar day with a smaller (or negative) time zone offset. Since the time zone offsets allowed spread over 28 hours, it is possible for the period denoted by a given calendar day with one time zone offset to be completely disjoint from the period denoted by the same calendar day with a different offset — the earlier day ends before the later one starts.  The moments in time represented by a single calendar day are spread over a 52-hour interval, from the beginning of the day in the +14:00 time zone offset to the end of that day in the −14:00 time zone offset.

**Note:** The relative order of two [time](#time) values, one of which has a [*timezoneOffset*](#vp-dt-timezone) of ***absent*** is determined by imputing time zone offsets of both +14:00 and −14:00 to the value without an offset. Many such combinations will be [incomparable](#dt-incomparable) because the two imputed time zone offsets yield different orders.  However, for a given non-timezoned value, there will always be timezoned values at one or both ends of the 52-hour interval that are [comparable](#dt-incomparable) (because the interval of [incomparability](#dt-incomparable) is only 28 hours wide). Some pairs of [time](#time) literals which in the 1.0 version of this specification denoted the same value now (in this version) denote distinct values instead, because values now include time zone offset information. Some such pairs, such as '`05:00:00-03:00`' and '`10:00:00+02:00`', now denote equal though distinct values (because they identify the same points on the time line); others, such as '`23:00:00-03:00`' and '`02:00:00Z`', now denote unequal values (23:00:00−03:00 > 02:00:00Z because 23:00:00−03:00 on any given day is equal to 02:00:00Z on *the next day*).

##### <a id="time-lexical-mapping"></a>3.3.8.2 Lexical Mappings

The lexical representations for [time](#time) are "projections" of those of [dateTime](#dateTime), as follows:  Lexical Space  <a id="nt-timeRep"></a>[17]   *timeLexicalRep* ::= (([*hourFrag*](#nt-hrFrag) '`:`' [*minuteFrag*](#nt-miFrag) '`:`' [*secondFrag*](#nt-seFrag)) | [*endOfDayFrag*](#nt-eodFrag)) [*timezoneFrag*](#nt-tzFrag)? The [*timeLexicalRep*](#nt-timeRep) production is equivalent to this regular expression, once whitespace is removed:

> > ` (([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]+)?|(24:00:00(\.0+)?))(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))?`

 Note that neither the [*timeLexicalRep*](#nt-timeRep) production nor this regular expression alone enforce the constraint on [*timeLexicalRep*](#nt-timeRep) given above.

The [lexical mapping](#dt-lexical-mapping) for [time](#time) is [*timeLexicalMap*](#vp-timeLexRep); the [canonical mapping](#dt-canonical-mapping) is [*timeCanonicalMap*](#vp-timeCanRep).

**Note:** The [lexical mapping](#dt-lexical-mapping) maps '`00:00:00`' and '`24:00:00`' to the same value, namely midnight ([*hour*](#vp-dt-hour) = 0 , [*minute*](#vp-dt-minute) = 0 , [*second*](#vp-dt-second) = 0).

##### <a id="time-facets"></a>3.3.8.3 Facets

The [time](#time) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [time](#time) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [explicitTimezone](#rf-explicitTimezone) = ***optional***

Datatypes derived by restriction from [time](#time) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [time](#time) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="date"></a>3.3.9 date

<a id="dt-date">[Definition:]</a> **date** represents top-open intervals of exactly one day in length on the timelines of [dateTime](#dateTime), beginning on the beginning moment of each day, up to but not including the beginning moment of the next day).  For non-timezoned values, the top-open intervals disjointly cover the non-timezoned timeline, one per day.  For timezoned values, the intervals begin at every minute and therefore overlap.

##### <a id="date-value-space"></a>3.3.9.1 Value Space

[date](#date) uses the [date/timeSevenPropertyModel](#dt-dt-7PropMod), with [*hour*](#vp-dt-hour), [*minute*](#vp-dt-minute), and [*second*](#vp-dt-second) required to be ***absent***.  [*timezoneOffset*](#vp-dt-timezone) remains [optional](#dt-optional).

<a id="con-date-dayValue"></a>**Constraint: Day-of-month Values**

> The [*day*](#vp-dt-day) value **must** be no more than 30 if [*month*](#vp-dt-month) is one of 4, 6, 9, or 11, no more than 28 if [*month*](#vp-dt-month) is 2 and [*year*](#vp-dt-year) is not divisible by 4, or is divisible by 100 but not by 400, and no more than 29 if [*month*](#vp-dt-month) is 2 and [*year*](#vp-dt-year) is divisible by 400, or by 4 but not by 100.

**Note:** See the conformance note in [Partial Implementation of Infinite Datatypes (§5.4)](#partial-implementation) which applies to the [*year*](#vp-dt-year) value of this datatype.

Equality and order are as prescribed in [The Seven-property Model (§D.2.1)](#theSevenPropertyModel).

**Note:** In version 1.0 of this specification, [date](#date) values did not retain a time zone offset explicitly, but for offsets not too far from zero their time zone offset could be recovered based on their value's first moment on the timeline.  The [date/timeSevenPropertyModel](#dt-dt-7PropMod) retains all time zone offsets.Some [date](#date) values with different time zone offsets that were identical in the 1.0 version of this specification, such as 2000-01-01+13:00 and 1999-12-31−11:00, are in this version of this specification equal (because they begin at the same moment on the time line) but are not identical (because they have and retain different time zone offsets).  This situation will arise for dates only if one has a far-from-zero time zone offset and hence in 1.0 its "recoverable time zone offset" was different from the the time zone offset which is retained in the [date/timeSevenPropertyModel](#dt-dt-7PropMod) used in this version of this specification.

##### <a id="date-lexical-mapping"></a>3.3.9.2 Lexical Mapping

The lexical representations for [date](#date) are "projections" of those of [dateTime](#dateTime), as follows:  Lexical Space  <a id="nt-dateRep"></a>[18]   *dateLexicalRep* ::= [*yearFrag*](#nt-yrFrag) '`-`' [*monthFrag*](#nt-moFrag) '`-`' [*dayFrag*](#nt-daFrag) [*timezoneFrag*](#nt-tzFrag)?   **Constraint:**  Day-of-month Representations <a id="con-date-day"></a>**Constraint: Day-of-month Representations**

> Within a [*dateLexicalRep*](#nt-dateRep), a [*dayFrag*](#nt-daFrag) **must not** begin with the digit '`3`' or be '`29`' unless the value to which it would map would satisfy the value constraint on [*day*](#vp-dt-day) values ("Constraint: Day-of-month Values") given above.

 The [*dateLexicalRep*](#nt-dateRep) production is equivalent to this regular expression:

> `-?([1-9][0-9]{3,}|0[0-9]{3})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))?`

 Note that neither the [*dateLexicalRep*](#nt-dateRep) production nor this regular expression alone enforce the constraint on [*dateLexicalRep*](#nt-dateRep) given above.

The [lexical mapping](#dt-lexical-mapping) for [date](#date) is [*dateLexicalMap*](#vp-dateLexRep). The [canonical mapping](#dt-canonical-mapping) is [*dateCanonicalMap*](#vp-dateCanRep).

##### <a id="date-facets"></a>3.3.9.3 Facets

The [date](#date) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [date](#date) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [explicitTimezone](#rf-explicitTimezone) = ***optional***

Datatypes derived by restriction from [date](#date) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [date](#date) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="gYearMonth"></a>3.3.10 gYearMonth

 **gYearMonth** represents specific whole Gregorian months in specific Gregorian years.

**Note:** Because month/year combinations in one calendar only rarely correspond to month/year combinations in other calendars, values of this type are not, in general, convertible to simple values corresponding to month/year combinations in other calendars.  This type should therefore be used with caution in contexts where conversion to other calendars is desired.

##### <a id="gYearMonth-value-space"></a>3.3.10.1 Value Space

[gYearMonth](#gYearMonth) uses the [date/timeSevenPropertyModel](#dt-dt-7PropMod), with [*day*](#vp-dt-day), [*hour*](#vp-dt-hour), [*minute*](#vp-dt-minute), and [*second*](#vp-dt-second) required to be ***absent***.  [*timezoneOffset*](#vp-dt-timezone) remains [optional](#dt-optional).

**Note:** See the conformance note in [Partial Implementation of Infinite Datatypes (§5.4)](#partial-implementation) which applies to the [*year*](#vp-dt-year) value of this datatype.

Equality and order are as prescribed in [The Seven-property Model (§D.2.1)](#theSevenPropertyModel).

##### <a id="gYearMonth-lexical-repr"></a>3.3.10.2 Lexical Mapping

The lexical representations for [gYearMonth](#gYearMonth) are "projections" of those of [dateTime](#dateTime), as follows:  Lexical Space  <a id="nt-gYearMonthRep"></a>[19]   *gYearMonthLexicalRep* ::= [*yearFrag*](#nt-yrFrag) '`-`' [*monthFrag*](#nt-moFrag) [*timezoneFrag*](#nt-tzFrag)? The [*gYearMonthLexicalRep*](#nt-gYearMonthRep) is equivalent to this regular expression:

> `-?([1-9][0-9]{3,}|0[0-9]{3})-(0[1-9]|1[0-2])(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))?`

The [lexical mapping](#dt-lexical-mapping) for [gYearMonth](#gYearMonth) is [*gYearMonthLexicalMap*](#vp-gYearMonthLexRep). The [canonical mapping](#dt-canonical-mapping) is [*gYearMonthCanonicalMap*](#vp-gYearMonthCanRep).

##### <a id="gYearMonth-facets"></a>3.3.10.3 Facets

The [gYearMonth](#gYearMonth) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [gYearMonth](#gYearMonth) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [explicitTimezone](#rf-explicitTimezone) = ***optional***

Datatypes derived by restriction from [gYearMonth](#gYearMonth) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [gYearMonth](#gYearMonth) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="gYear"></a>3.3.11 gYear

**gYear** represents Gregorian calendar years.

**Note:** Because years in one calendar only rarely correspond to years in other calendars, values of this type are not, in general, convertible to simple values corresponding to years in other calendars.  This type should therefore be used with caution in contexts where conversion to other calendars is desired.

##### <a id="gYear-value-space"></a>3.3.11.1 Value Space

[gYear](#gYear) uses the [date/timeSevenPropertyModel](#dt-dt-7PropMod), with [*month*](#vp-dt-month), [*day*](#vp-dt-day), [*hour*](#vp-dt-hour), [*minute*](#vp-dt-minute), and [*second*](#vp-dt-second) required to be ***absent***.  [*timezoneOffset*](#vp-dt-timezone) remains [optional](#dt-optional).

**Note:** See the conformance note in [Partial Implementation of Infinite Datatypes (§5.4)](#partial-implementation) which applies to the [*year*](#vp-dt-year) value of this datatype.

Equality and order are as prescribed in [The Seven-property Model (§D.2.1)](#theSevenPropertyModel).

##### <a id="gYear-lexical-repr"></a>3.3.11.2 Lexical Mapping

The lexical representations for [gYear](#gYear) are "projections" of those of [dateTime](#dateTime), as follows:  Lexical Space  <a id="nt-gYearRep"></a>[20]   *gYearLexicalRep* ::= [*yearFrag*](#nt-yrFrag) [*timezoneFrag*](#nt-tzFrag)? The [*gYearLexicalRep*](#nt-gYearRep) is equivalent to this regular expression:

> `-?([1-9][0-9]{3,}|0[0-9]{3})(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))?`

The [lexical mapping](#dt-lexical-mapping) for [gYear](#gYear) is [*gYearLexicalMap*](#vp-gYearLexRep). The [canonical mapping](#dt-canonical-mapping) is [*gYearCanonicalMap*](#vp-gYearCanRep).

##### <a id="gYear-facets"></a>3.3.11.3 Facets

The [gYear](#gYear) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [gYear](#gYear) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [explicitTimezone](#rf-explicitTimezone) = ***optional***

Datatypes derived by restriction from [gYear](#gYear) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [gYear](#gYear) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="gMonthDay"></a>3.3.12 gMonthDay

[gMonthDay](#gMonthDay) represents whole calendar days that recur at the same point in each calendar year, or that occur in some arbitrary calendar year.  (Obviously, days beyond 28 cannot occur in all Februaries; 29 is nonetheless permitted.)

This datatype can be used, for example, to record birthdays; an instance of the datatype could be used to say that someone's birthday occurs on the 14th of September every year.

**Note:** Because day/month combinations in one calendar only rarely correspond to day/month combinations in other calendars, values of this type do not, in general, have any straightforward or intuitive representation in terms of most other calendars. This type should therefore be used with caution in contexts where conversion to other calendars is desired.

##### <a id="gMonthDay-value-space"></a>3.3.12.1 Value Space

[gMonthDay](#gMonthDay) uses the [date/timeSevenPropertyModel](#dt-dt-7PropMod), with [*year*](#vp-dt-year), [*hour*](#vp-dt-hour), [*minute*](#vp-dt-minute), and [*second*](#vp-dt-second) required to be ***absent***.  [*timezoneOffset*](#vp-dt-timezone) remains [optional](#dt-optional).

<a id="con-gMonthDay-dayValue"></a>**Constraint: Day-of-month Values**

> The [*day*](#vp-dt-day) value **must** be no more than 30 if [*month*](#vp-dt-month) is one of 4, 6, 9, or 11, and no more than 29 if [*month*](#vp-dt-month) is 2.

Equality and order are as prescribed in [The Seven-property Model (§D.2.1)](#theSevenPropertyModel).

**Note:** In version 1.0 of this specification, [gMonthDay](#gMonthDay) values did not retain a time zone offset explicitly, but for time zone offsets not too far from [UTC](#dt-utc) their time zone offset could be recovered based on their value's first moment on the timeline.  The [date/timeSevenPropertyModel](#dt-dt-7PropMod) retains all time zone offsets.An example that shows the difference from version 1.0 (see [Lexical Mapping (§3.3.12.2)](#gMonthDay-lexical-repr) for the notations):

- A day is a calendar (or "local time") day offset from [UTC](#dt-utc) by the appropriate interval; this is now true for all [*day*](#vp-dt-day) values, including those with time zone offsets outside the range +12:00 through -11:59 inclusive:--12-12+13:00 < --12-12+11:00  (just as --12-12+12:00 has always been less than --12-12+11:00, but in version 1.0  --12-12+13:00 > --12-12+11:00 , since --12-12+13:00's "recoverable time zone offset" was −11:00)

##### <a id="gMonthDay-lexical-repr"></a>3.3.12.2 Lexical Mapping

The lexical representations for [gMonthDay](#gMonthDay) are "projections" of those of [dateTime](#dateTime), as follows:  Lexical Space  <a id="nt-gMonthDayRep"></a>[21]   *gMonthDayLexicalRep* ::= '`--`' [*monthFrag*](#nt-moFrag) '`-`' [*dayFrag*](#nt-daFrag) [*timezoneFrag*](#nt-tzFrag)?   **Constraint:**  Day-of-month Representations <a id="con-gMonthDay-day"></a>**Constraint: Day-of-month Representations**

> Within a [*gMonthDayLexicalRep*](#nt-gMonthDayRep), a [*dayFrag*](#nt-daFrag) **must not** begin with the digit '`3`' or be '`29`' unless the value to which it would map would satisfy the value constraint on [*day*](#vp-dt-day) values ("Constraint: Day-of-month Values") given above.

 The [*gMonthDayLexicalRep*](#nt-gMonthDayRep) is equivalent to this regular expression:

> `--(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))?`

 Note that neither the [*gMonthDayLexicalRep*](#nt-gMonthDayRep) production nor this regular expression alone enforce the constraint on [*gMonthDayLexicalRep*](#nt-gMonthDayRep) given above.

The [lexical mapping](#dt-lexical-mapping) for [gMonthDay](#gMonthDay) is [*gMonthDayLexicalMap*](#vp-gMonthDayLexRep). The [canonical mapping](#dt-canonical-mapping) is [*gMonthDayCanonicalMap*](#vp-gMonthDayCanRep).

##### <a id="gMonthDay-facets"></a>3.3.12.3 Facets

The [gMonthDay](#gMonthDay) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [gMonthDay](#gMonthDay) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [explicitTimezone](#rf-explicitTimezone) = ***optional***

Datatypes derived by restriction from [gMonthDay](#gMonthDay) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [gMonthDay](#gMonthDay) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="gDay"></a>3.3.13 gDay

<a id="dt-gday">[Definition:]</a>**gDay** represents whole days within an arbitrary month—days that recur at the same point in each (Gregorian) month. This datatype is used to represent a specific day of the month. To indicate, for example, that an employee gets a paycheck on the 15th of each month.  (Obviously, days beyond 28 cannot occur in *all* months; they are nonetheless permitted, up to 31.)

**Note:** Because days in one calendar only rarely correspond to days in other calendars, [gDay](#gDay) values do not, in general, have any straightforward or intuitive representation in terms of most non-Gregorian calendars. [gDay](#gDay) should therefore be used with caution in contexts where conversion to other calendars is desired.

##### <a id="sec-vs-gDay"></a>3.3.13.1 Value Space

[gDay](#gDay) uses the [date/timeSevenPropertyModel](#dt-dt-7PropMod), with [*year*](#vp-dt-year), [*month*](#vp-dt-month), [*hour*](#vp-dt-hour), [*minute*](#vp-dt-minute), and [*second*](#vp-dt-second) required to be ***absent***.  [*timezoneOffset*](#vp-dt-timezone) remains [optional](#dt-optional) and [*day*](#vp-dt-day) **must** be between 1 and 31 inclusive.

Equality and order are as prescribed in [The Seven-property Model (§D.2.1)](#theSevenPropertyModel).  Since [gDay](#gDay) values (days) are ordered by their first moments, it is possible for apparent anomalies to appear in the order when [*timezoneOffset*](#vp-dt-timezone) values differ by at least 24 hours.  (It is possible for [*timezoneOffset*](#vp-dt-timezone) values to differ by up to 28 hours.)

 Examples that may appear anomalous (see [Lexical Mapping (§3.3.13.2)](#gDay-lexical-mapping) for the notations):

- ---15 < ---16 , but  ---15−13:00 > ---16+13:00
- ---15−11:00 = ---16+13:00
- ---15−13:00 <> ---16 , because  ---15−13:00 > ---16+14:00  and ---15−13:00 < 16−14:00
 **Note:** Time zone offsets do not cause wrap-around at the end of the month:  the last day of a given month with a time zone offset of −13:00 may start after the first day of the *next* month with offset +13:00, as measured on the global timeline, but nonetheless  ---01+13:00 < ---31−13:00 .

##### <a id="gDay-lexical-mapping"></a>3.3.13.2 Lexical Mapping

 The lexical representations for [gDay](#gDay) are "projections" of those of [dateTime](#dateTime), as follows:  Lexical Space  <a id="nt-gDayRep"></a>[22]   *gDayLexicalRep* ::= '`---`' [*dayFrag*](#nt-daFrag) [*timezoneFrag*](#nt-tzFrag)? The [*gDayLexicalRep*](#nt-gDayRep) is equivalent to this regular expression:

> `---(0[1-9]|[12][0-9]|3[01])(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))?`

The [lexical mapping](#dt-lexical-mapping) for [gDay](#gDay) is [*gDayLexicalMap*](#vp-gDayLexRep). The [canonical mapping](#dt-canonical-mapping) is [*gDayCanonicalMap*](#vp-gDayCanRep).

##### <a id="gDay-facets"></a>3.3.13.3 Facets

The [gDay](#gDay) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [gDay](#gDay) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [explicitTimezone](#rf-explicitTimezone) = ***optional***

Datatypes derived by restriction from [gDay](#gDay) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [gDay](#gDay) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="gMonth"></a>3.3.14 gMonth

**gMonth** represents whole (Gregorian) months within an arbitrary year—months that recur at the same point in each year.  It might be used, for example, to say what month annual Thanksgiving celebrations fall in different countries (--11 in the United States, --10 in Canada, and possibly other months in other countries).

**Note:** Because months in one calendar only rarely correspond to months in other calendars, values of this type do not, in general, have any straightforward or intuitive representation in terms of most other calendars. This type should therefore be used with caution in contexts where conversion to other calendars is desired.

##### <a id="gMonth-value-space"></a>3.3.14.1 Value Space

[gMonth](#gMonth) uses the [date/timeSevenPropertyModel](#dt-dt-7PropMod), with [*year*](#vp-dt-year), [*day*](#vp-dt-day), [*hour*](#vp-dt-hour), [*minute*](#vp-dt-minute), and [*second*](#vp-dt-second) required to be ***absent***.  [*timezoneOffset*](#vp-dt-timezone) remains [optional](#dt-optional).

Equality and order are as prescribed in [The Seven-property Model (§D.2.1)](#theSevenPropertyModel).

##### <a id="gMonth-lexical-repr"></a>3.3.14.2 Lexical Mapping

The lexical representations for [gMonth](#gMonth) are "projections" of those of [dateTime](#dateTime), as follows:  Lexical Space  <a id="nt-gMonthRep"></a>[23]   *gMonthLexicalRep* ::= '`--`' [*monthFrag*](#nt-moFrag) [*timezoneFrag*](#nt-tzFrag)? The [*gMonthLexicalRep*](#nt-gMonthRep) is equivalent to this regular expression:

> `--(0[1-9]|1[0-2])(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))?`

 The [lexical mapping](#dt-lexical-mapping) for [gMonth](#gMonth) is [*gMonthLexicalMap*](#vp-gMonthLexRep). The [canonical mapping](#dt-canonical-mapping) is [*gMonthCanonicalMap*](#vp-gMonthCanRep).

##### <a id="gMonth-facets"></a>3.3.14.3 Facets

The [gMonth](#gMonth) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [gMonth](#gMonth) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [explicitTimezone](#rf-explicitTimezone) = ***optional***

Datatypes derived by restriction from [gMonth](#gMonth) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [gMonth](#gMonth) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="hexBinary"></a>3.3.15 hexBinary

<a id="dt-hexBinary">[Definition:]</a>**hexBinary** represents arbitrary hex-encoded binary data. 

##### <a id="sec-vs-hexbin"></a>3.3.15.1 Value Space

The [value space](#dt-value-space) of [hexBinary](#hexBinary) is the set of finite-length sequences of zero or more binary octets.  The length of a value is the number of octets.

##### <a id="hexBinary-lexical-representation"></a>3.3.15.2 Lexical Mapping

[hexBinary](#hexBinary)'s [lexical space](#dt-lexical-space) consists of strings of hex (hexadecimal) digits, two consecutive digits representing each octet in the corresponding value (treating the octet as the binary representation of a number between 0 and 255).  For example, '`0FB7`' is a [lexical representation](#dt-lexical-representation) of the two-octet value 00001111 10110111.

More formally, the [lexical space](#dt-lexical-space) of [hexBinary](#hexBinary) is the set of literals matching the [*hexBinary*](#nt-hexBinary) production.  Lexical space of hexBinary  <a id="nt-hexDigit"></a>[24]   *hexDigit* ::= [`0-9a-fA-F`] <a id="nt-hexOctet"></a>[25]   *hexOctet* ::= [*hexDigit*](#nt-hexDigit) [*hexDigit*](#nt-hexDigit) <a id="nt-hexBinary"></a>[26]   *hexBinary* ::= [*hexOctet*](#nt-hexOctet)*

The set recognized by [*hexBinary*](#nt-hexBinary) is the same as that recognized by the regular expression '`([0-9a-fA-F]{2})*`'.

The [lexical mapping](#dt-lexical-mapping) of [hexBinary](#hexBinary) is [*hexBinaryMap*](#f-hexBinaryMap).

The [canonical mapping](#dt-canonical-mapping) of [hexBinary](#hexBinary) is given formally in [*hexBinaryCanonical*](#f-hexBinaryCanonical).

##### <a id="hexBinary-facets"></a>3.3.15.3 Facets

The [hexBinary](#hexBinary) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

Datatypes derived by restriction from [hexBinary](#hexBinary) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [hexBinary](#hexBinary) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="base64Binary"></a>3.3.16 base64Binary

<a id="dt-base64Binary">[Definition:]</a> **base64Binary** represents arbitrary Base64-encoded binary data.  For **base64Binary** data the entire binary stream is encoded using the Base64 Encoding defined in [[RFC 3548]](#RFC3548), which is derived from the encoding described in [[RFC 2045]](#RFC2045).

##### <a id="sec-vs-b46b"></a>3.3.16.1 Value Space

The [value space](#dt-value-space) of [base64Binary](#base64Binary) is the set of finite-length sequences of zero or more binary octets.  The length of a value is the number of octets.

##### <a id="sec-lex-b64b"></a>3.3.16.2 Lexical Mapping

The [lexical representations](#dt-lexical-representation) of [base64Binary](#base64Binary) values are limited to the 65 characters of the Base64 Alphabet defined in [[RFC 3548]](#RFC3548), i.e., `a-z`, `A-Z`, `0-9`, the plus sign (+), the forward slash (/) and the equal sign (=), together with the space character (#x20). No other characters are allowed.

 For compatibility with older mail gateways, [[RFC 2045]](#RFC2045) suggests that Base64 data should have lines limited to at most 76 characters in length.  This line-length limitation is not required by [[RFC 3548]](#RFC3548) and is not mandated in the [lexical representations](#dt-lexical-representation) of [base64Binary](#base64Binary) data.  It **must not** be enforced by XML Schema processors.

The [lexical space](#dt-lexical-space) of [base64Binary](#base64Binary) is the set of literals which [match](#dt-match) the [base64Binary](#base64Binary)production.

  Lexical space of base64Binary  <a id="nt-Base64Binary"></a>[27]   *Base64Binary* ::= ([*B64quad*](#nt-B64quad)* [*B64final*](#nt-B64final))? <a id="nt-B64quad"></a>[28]   *B64quad* ::= ([*B64*](#nt-B64) [*B64*](#nt-B64) [*B64*](#nt-B64) [*B64*](#nt-B64)) /* *[*B64quad*](#nt-B64quad) represents three octets of binary data.* */  <a id="nt-B64final"></a>[29]   *B64final* ::= [*B64finalquad*](#nt-B64finalquad) | [*Padded16*](#nt-Padded16) | [*Padded8*](#nt-Padded8) <a id="nt-B64finalquad"></a>[30]   *B64finalquad* ::= ([*B64*](#nt-B64) [*B64*](#nt-B64) [*B64*](#nt-B64) [*B64char*](#nt-B64char)) /* *[*B64finalquad*](#nt-B64finalquad) represents three octets of binary data without trailing space.* */  <a id="nt-Padded16"></a>[31]   *Padded16* ::= [*B64*](#nt-B64) [*B64*](#nt-B64) [*B16*](#nt-B16) '`=`' /* *[*Padded16*](#nt-Padded16) represents a two-octet at the end of the data.* */  <a id="nt-Padded8"></a>[32]   *Padded8* ::= [*B64*](#nt-B64) [*B04*](#nt-B04) '`=`' #x20? '`=`' /* *[*Padded8*](#nt-Padded8) represents a single octet at the end of the data.* */  <a id="nt-B64"></a>[33]   *B64* ::= [*B64char*](#nt-B64char) #x20? <a id="nt-B64char"></a>[34]   *B64char* ::= [A-Za-z0-9+/] <a id="nt-B16"></a>[35]   *B16* ::= [*B16char*](#nt-B16char) #x20? <a id="nt-B16char"></a>[36]   *B16char* ::= [AEIMQUYcgkosw048] /* *Base64 characters whose bit-string value ends in '00'* */ <a id="nt-B04"></a>[37]   *B04* ::= [*B04char*](#nt-B04char) #x20? <a id="nt-B04char"></a>[38]   *B04char* ::= [AQgw] /* *Base64 characters whose bit-string value ends in '0000'* */  The [*Base64Binary*](#nt-Base64Binary) production is equivalent to the following regular expression.

> `((([A-Za-z0-9+/] ?){4})*(([A-Za-z0-9+/] ?){3}[A-Za-z0-9+/]|([A-Za-z0-9+/] ?){2}[AEIMQUYcgkosw048] ?=|[A-Za-z0-9+/] ?[AQgw] ?= ?=))?`

 Note that each '`?`' except the last is preceded by a single space character.

Note that this grammar requires the number of non-whitespace characters in the [lexical representation](#dt-lexical-representation) to be a multiple of four, and for equals signs to appear only at the end of the [lexical representation](#dt-lexical-representation); literals which do not meet these constraints are not legal [lexical representations](#dt-lexical-representation) of [base64Binary](#base64Binary).

The [lexical mapping](#dt-lexical-mapping) for [base64Binary](#base64Binary) is as given in [[RFC 2045]](#RFC2045) and [[RFC 3548]](#RFC3548).

**Note:** The above definition of the [lexical space](#dt-lexical-space) is more restrictive than that given in [[RFC 2045]](#RFC2045) as regards whitespace — and less restrictive than [[RFC 3548]](#RFC3548). This is not an issue in practice.  Any string compatible with either RFC can occur in an element or attribute validated by this type, because the [whiteSpace](#dt-whiteSpace) facet of this type is fixed to ***collapse***, which means that all leading and trailing whitespace will be stripped, and all internal whitespace collapsed to single space characters, *before* the above grammar is enforced. The possibility of ignoring whitespace in Base64 data is foreseen in clause 2.3 of [[RFC 3548]](#RFC3548), but for the reasons given there this specification does not allow implementations to ignore non-whitespace characters which are not in the Base64 Alphabet.

The canonical [lexical representation](#dt-lexical-representation) of a [base64Binary](#base64Binary) data value is the Base64 encoding of the value which matches the Canonical-base64Binary production in the following grammar:

  Canonical representation of base64Binary  <a id="nt-Canonical-base64Binary"></a>[39]   *Canonical-base64Binary* ::= [*CanonicalQuad*](#nt-CanonicalQuad)* [*CanonicalPadded*](#nt-CanonicalPadded)? <a id="nt-CanonicalQuad"></a>[40]   *CanonicalQuad* ::= [*B64char*](#nt-B64char) [*B64char*](#nt-B64char) [*B64char*](#nt-B64char) [*B64char*](#nt-B64char) <a id="nt-CanonicalPadded"></a>[41]   *CanonicalPadded* ::= [*B64char*](#nt-B64char) [*B64char*](#nt-B64char) [*B16char*](#nt-B16char) '`=`' | [*B64char*](#nt-B64char) [*B04char*](#nt-B04char) '`==`'

That is, the [canonical representation](#dt-canonical-representation) of a [base64Binary](#base64Binary) value is the [lexical representation](#dt-lexical-representation) which maps to that value and contains no whitespace. The [canonical mapping](#dt-canonical-mapping) for [base64Binary](#base64Binary) is thus the encoding algorithm for Base64 data given in [[RFC 2045]](#RFC2045) and [[RFC 3548]](#RFC3548), with the proviso that no characters except those in the Base64 Alphabet are to be written out.

**Note:** For some values the [canonical representation](#dt-canonical-representation) defined above does not conform to [[RFC 2045]](#RFC2045), which requires breaking with linefeeds at appropriate intervals. It does conform with [[RFC 3548]](#RFC3548).

The length of a [base64Binary](#base64Binary) value may be calculated from the [lexical representation](#dt-lexical-representation) by removing whitespace and padding characters and performing the calculation shown in the pseudo-code below:

 ` lex2   := killwhitespace(lexform)    -- remove whitespace characters
 lex3   := strip_equals(lex2)         -- strip padding characters at end
 length := floor (length(lex3) * 3 / 4)         -- calculate length `

 Note on encoding:  [[RFC 2045]](#RFC2045) and [[RFC 3548]](#RFC3548) explicitly reference US-ASCII encoding.  However, decoding of **base64Binary** data in an XML entity is to be performed on the Unicode characters obtained after character encoding processing as specified by [[XML]](#XML).

##### <a id="base64Binary-facets"></a>3.3.16.3 Facets

The [base64Binary](#base64Binary) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

Datatypes derived by restriction from [base64Binary](#base64Binary) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [base64Binary](#base64Binary) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="anyURI"></a>3.3.17 anyURI

<a id="dt-anyURI">[Definition:]</a> **anyURI** represents an Internationalized Resource Identifier Reference (IRI).  An **anyURI** value can be absolute or relative, and may have an optional fragment identifier (i.e., it may be an IRI Reference).  This type should be used when the value fulfills the role of an IRI, as defined in [[RFC 3987]](#RFC3987) or its successor(s) in the IETF Standards Track.

**Note:** IRIs may be used to locate resources or simply to identify them. In the case where they are used to locate resources using a URI, applications should use the mapping from [anyURI](#anyURI) values to URIs given by the reference escaping procedure defined in [[LEIRI]](#LEIRIs) and in Section 3.1 [Mapping of IRIs to URIs](http://www.ietf.org/rfc/rfc3987.txt) of [[RFC 3987]](#RFC3987) or its successor(s) in the IETF Standards Track.  This means that a wide range of internationalized resource identifiers can be specified when an [anyURI](#anyURI) is called for, and still be understood as URIs per [[RFC 3986]](#RFC3986) and its successor(s).

##### <a id="anyURI-vs"></a>3.3.17.1 Value Space

The value space of [anyURI](#anyURI) is the set of finite-length sequences of zero or more [character](https://www.w3.org/TR/xml11/#dt-character)s (as defined in [[XML]](#XML)) that [match](#dt-match) the [Char](https://www.w3.org/TR/xml11/#NT-Char) production from [[XML]](#XML).

##### <a id="anyURI-lexical-representation"></a>3.3.17.2 Lexical Mapping

The [lexical space](#dt-lexical-space) of [anyURI](#anyURI) is the set of finite-length sequences of zero or more [character](https://www.w3.org/TR/xml11/#dt-character)s (as defined in [[XML]](#XML)) that [match](#dt-match) the [Char](https://www.w3.org/TR/xml11/#NT-Char) production from [[XML]](#XML).

**Note:** For an [anyURI](#anyURI) value to be usable in practice as an IRI, the result of applying to it the algorithm defined in Section 3.1 of [[RFC 3987]](#RFC3987) should be a string which is a legal URI according to [[RFC 3986]](#RFC3986). (This is true at the time this document is published; if in the future [[RFC 3987]](#RFC3987) and [[RFC 3986]](#RFC3986) are replaced by other specifications in the IETF Standards Track, the relevant constraints will be those imposed by those successor specifications.)Each URI scheme imposes specialized syntax rules for URIs in that scheme, including restrictions on the syntax of allowed fragment identifiers. Because it is impractical for processors to check that a value is a context-appropriate URI reference, neither the syntactic constraints defined by the definitions of individual schemes nor the generic syntactic constraints defined by [[RFC 3987]](#RFC3987) and [[RFC 3986]](#RFC3986) and their successors are part of this datatype as defined here. Applications which depend on [anyURI](#anyURI) values being legal according to the rules of the relevant specifications should make arrangements to check values against the appropriate definitions of IRI, URI, and specific schemes.**Note:** Spaces are, in principle, allowed in the [lexical space](#dt-lexical-space) of [anyURI](#anyURI), however, their use is highly discouraged (unless they are encoded by '`%20`').

The [lexical mapping](#dt-lexical-mapping) for [anyURI](#anyURI) is the identity mapping.

**Note:** The definitions of URI in the current IETF specifications define certain URIs as equivalent to each other. Those equivalences are not part of this datatype as defined here: if two "equivalent" URIs or IRIs are different character sequences, they map to different values in this datatype.

##### <a id="anyURI-facets"></a>3.3.17.3 Facets

The [anyURI](#anyURI) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

Datatypes derived by restriction from [anyURI](#anyURI) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [anyURI](#anyURI) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="QName"></a>3.3.18 QName

<a id="dt-QName">[Definition:]</a> **QName** represents [XML qualified names](https://www.w3.org/TR/xml-names11/#dt-qualname). The [value space](#dt-value-space) of **QName** is the set of tuples {[namespace name](https://www.w3.org/TR/xml-names11/#dt-NSName), [local part](https://www.w3.org/TR/xml-names11/#dt-localname)}, where [namespace name](https://www.w3.org/TR/xml-names11/#dt-NSName) is an [anyURI](#anyURI) and [local part](https://www.w3.org/TR/xml-names11/#dt-localname) is an [NCName](#NCName). The [lexical space](#dt-lexical-space) of **QName** is the set of strings that [match](#dt-match) the [QName](https://www.w3.org/TR/xml-names11/#NT-QName) production of [[Namespaces in XML]](#XMLNS).

It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports the [QName](https://www.w3.org/TR/xml-names11/#NT-QName) production from [[Namespaces in XML]](#XMLNS), or that from [[Namespaces in XML 1.0]](#XMLNS1.0), or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

 The mapping from lexical space to value space for a particular [QName](#QName) [literal](#dt-literal) depends on the namespace bindings in scope where the literal occurs.

When [QName](#QName)s appear in an XML context, the bindings to be used in the [lexical mapping](#dt-lexical-mapping) are those in the [[in-scope namespaces]](https://www.w3.org/TR/xml-infoset/#infoitem.element) property of the relevant element. When this datatype is used in a non-XML host language, the host language **must** specify what namespace bindings are to be used.

 The host language, whether XML-based or otherwise, **may** specify whether unqualified names are bound to the default namespace (if any) or not; the host language may also place this under user control. If the host language does not specify otherwise, unqualified names are bound to the default namespace.

**Note:** The default treatment of unqualified names parallels that specified in [[Namespaces in XML]](#XMLNS) for element names (as opposed to that specified for attribute names). **Note:** The mapping between [literals](#dt-literal) in the [lexical space](#dt-lexical-space) and values in the [value space](#dt-value-space) of [QName](#QName) depends on the set of namespace declarations in scope for the context in which [QName](#QName) is used.  Because the lexical representations available for any value of type [QName](#QName) vary with context, no [canonical representation](#dt-canonical-representation) is defined for [QName](#QName) in this specification.

##### <a id="QName-facets"></a>3.3.18.1 Facets

The [QName](#QName) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

Datatypes derived by restriction from [QName](#QName) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [QName](#QName) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="NOTATION"></a>3.3.19 NOTATION

<a id="dt-NOTATION">[Definition:]</a>**NOTATION** represents the [NOTATION](https://www.w3.org/TR/xml11/#NT-NotationType) attribute type from [[XML]](#XML). The [value space](#dt-value-space) of **NOTATION** is the set of [QName](#QName)s of notations declared in the current schema. The [lexical space](#dt-lexical-space) of **NOTATION** is the set of all names of [notations](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-notation) declared in the current schema (in the form of [QName](#QName)s).

**Note:** Because its [value space](#dt-value-space) depends on the notion of a "current schema", as instantiated for example by [[XSD 1.1 Part 1: Structures]](#structural-schemas), the [NOTATION](#NOTATION) datatype is unsuitable for use in other contexts which lack the notion of a current schema.

 The lexical mapping rules for [NOTATION](#NOTATION) are as given for [QName](#QName) in [QName (§3.3.18)](#QName).

<a id="enumeration-required-notation"></a>**Schema Component Constraint: enumeration facet value required for NOTATION**

> It is (with one exception) an [error](#dt-error) for [NOTATION](#NOTATION) to be used directly to validate a literal as described in [Datatype Valid (§4.1.4)](#cvc-datatype-valid): only datatypes [derived](#dt-derived) from [NOTATION](#NOTATION) by specifying a value for [enumeration](#dt-enumeration) can be used to validate literals. The exception is that in the [derivation](#dt-derived) of a new type the [literals](#dt-literal) used to enumerate the allowed values **may** be (and in the context of [XSD 1.1 Part 1: Structures] will be) validated directly against [NOTATION](#NOTATION); this amounts to verifying that the value is a [QName](#QName) and that the [QName](#QName) is the name of a **NOTATION** declared in the current schema.

For compatibility (see [Terminology (§1.6)](#terminology)) [NOTATION](#NOTATION) should be used only on attributes and should only be used in schemas with no target namespace.

 **Note:** Because the lexical representations available for any given value of [NOTATION](#NOTATION) vary with context, this specification defines no [canonical representation](#dt-canonical-representation) for [NOTATION](#NOTATION) values.

##### <a id="NOTATION-facets"></a>3.3.19.1 Facets

The [NOTATION](#NOTATION) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

Datatypes derived by restriction from [NOTATION](#NOTATION) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [NOTATION](#NOTATION) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

 The use of [length](#dt-length), [minLength](#dt-minLength) and [maxLength](#dt-maxLength) on [NOTATION](#NOTATION) or datatypes [derived](#dt-derived) from [NOTATION](#NOTATION) is deprecated.  Future versions of this specification may remove these facets for this datatype.

### <a id="ordinary-built-ins"></a>3.4 Other Built-in Datatypes

        3.4.1 [normalizedString](#normalizedString)
             3.4.1.1 [Facets](#normalizedString-facets)
             3.4.1.2 [Derived datatypes](#normalizedString-derived-types)
         3.4.2 [token](#token)
             3.4.2.1 [Facets](#token-facets)
             3.4.2.2 [Derived datatypes](#token-derived-types)
         3.4.3 [language](#language)
             3.4.3.1 [Facets](#language-facets)
         3.4.4 [NMTOKEN](#NMTOKEN)
             3.4.4.1 [Facets](#NMTOKEN-facets)
             3.4.4.2 [Related datatypes](#NMTOKEN-derived-types)
         3.4.5 [NMTOKENS](#NMTOKENS)
             3.4.5.1 [Facets](#NMTOKENS-facets)
         3.4.6 [Name](#Name)
             3.4.6.1 [Facets](#Name-facets)
             3.4.6.2 [Derived datatypes](#Name-derived-types)
         3.4.7 [NCName](#NCName)
             3.4.7.1 [Facets](#NCName-facets)
             3.4.7.2 [Derived datatypes](#NCName-derived-types)
         3.4.8 [ID](#ID)
             3.4.8.1 [Facets](#ID-facets)
         3.4.9 [IDREF](#IDREF)
             3.4.9.1 [Facets](#IDREF-facets)
             3.4.9.2 [Related datatypes](#IDREF-derived-types)
         3.4.10 [IDREFS](#IDREFS)
             3.4.10.1 [Facets](#IDREFS-facets)
         3.4.11 [ENTITY](#ENTITY)
             3.4.11.1 [Facets](#ENTITY-facets)
             3.4.11.2 [Related datatypes](#ENTITY-derived-types)
         3.4.12 [ENTITIES](#ENTITIES)
             3.4.12.1 [Facets](#ENTITIES-facets)
         3.4.13 [integer](#integer)
             3.4.13.1 [Lexical representation](#integer-lexical-representation)
             3.4.13.2 [Canonical representation](#integer-canonical-repr)
             3.4.13.3 [Facets](#integer-facets)
             3.4.13.4 [Derived datatypes](#integer-derived-types)
         3.4.14 [nonPositiveInteger](#nonPositiveInteger)
             3.4.14.1 [Lexical representation](#nonPositiveInteger-lexical-representation)
             3.4.14.2 [Canonical representation](#nonPositiveInteger-canonical-repr)
             3.4.14.3 [Facets](#nonPositiveInteger-facets)
             3.4.14.4 [Derived datatypes](#nonPositiveInteger-derived-types)
         3.4.15 [negativeInteger](#negativeInteger)
             3.4.15.1 [Lexical representation](#negativeInteger-lexical-representation)
             3.4.15.2 [Canonical representation](#negativeInteger-canonical-repr)
             3.4.15.3 [Facets](#negativeInteger-facets)
         3.4.16 [long](#long)
             3.4.16.1 [Lexical Representation](#long-lexical-representation)
             3.4.16.2 [Canonical Representation](#long-canonical-repr)
             3.4.16.3 [Facets](#long-facets)
             3.4.16.4 [Derived datatypes](#long-derived-types)
         3.4.17 [int](#int)
             3.4.17.1 [Lexical Representation](#int-lexical-representation)
             3.4.17.2 [Canonical representation](#int-canonical-repr)
             3.4.17.3 [Facets](#int-facets)
             3.4.17.4 [Derived datatypes](#int-derived-types)
         3.4.18 [short](#short)
             3.4.18.1 [Lexical representation](#short-lexical-representation)
             3.4.18.2 [Canonical representation](#short-canonical-repr)
             3.4.18.3 [Facets](#short-facets)
             3.4.18.4 [Derived datatypes](#short-derived-types)
         3.4.19 [byte](#byte)
             3.4.19.1 [Lexical representation](#byte-lexical-representation)
             3.4.19.2 [Canonical representation](#byte-canonical-repr)
             3.4.19.3 [Facets](#byte-facets)
         3.4.20 [nonNegativeInteger](#nonNegativeInteger)
             3.4.20.1 [Lexical representation](#nonNegativeInteger-lexical-representation)
             3.4.20.2 [Canonical representation](#nonNegativeInteger-canonical-repr)
             3.4.20.3 [Facets](#nonNegativeInteger-facets)
             3.4.20.4 [Derived datatypes](#nonNegativeInteger-derived-types)
         3.4.21 [unsignedLong](#unsignedLong)
             3.4.21.1 [Lexical representation](#unsignedLong-lexical-representation)
             3.4.21.2 [Canonical representation](#unsignedLong-canonical-repr)
             3.4.21.3 [Facets](#unsignedLong-facets)
             3.4.21.4 [Derived datatypes](#unsignedLong-derived-types)
         3.4.22 [unsignedInt](#unsignedInt)
             3.4.22.1 [Lexical representation](#unsignedInt-lexical-representation)
             3.4.22.2 [Canonical representation](#unsignedInt-canonical-repr)
             3.4.22.3 [Facets](#unsignedInt-facets)
             3.4.22.4 [Derived datatypes](#unsignedInt-derived-types)
         3.4.23 [unsignedShort](#unsignedShort)
             3.4.23.1 [Lexical representation](#unsignedShort-lexical-representation)
             3.4.23.2 [Canonical representation](#unsignedShort-canonical-repr)
             3.4.23.3 [Facets](#unsignedShort-facets)
             3.4.23.4 [Derived datatypes](#unsignedShort-derived-types)
         3.4.24 [unsignedByte](#unsignedByte)
             3.4.24.1 [Lexical representation](#unsignedByte-lexical-representation)
             3.4.24.2 [Canonical representation](#unsignedByte-canonical-repr)
             3.4.24.3 [Facets](#unisngedByte-facets)
         3.4.25 [positiveInteger](#positiveInteger)
             3.4.25.1 [Lexical representation](#positiveInteger-lexical-representation)
             3.4.25.2 [Canonical representation](#positiveInteger-canonical-repr)
             3.4.25.3 [Facets](#positiveInteger-facets)
         3.4.26 [yearMonthDuration](#yearMonthDuration)
             3.4.26.1 [The Lexical Mapping](#yearMonthDuration-lexical-mapping)
             3.4.26.2 [Facets](#YearMonthDuration-facets)
         3.4.27 [dayTimeDuration](#dayTimeDuration)
             3.4.27.1 [The Lexical Space](#dayTimeDuration-lexical-mapping)
             3.4.27.2 [Facets](#dayTimeDuration-facets)
         3.4.28 [dateTimeStamp](#dateTimeStamp)
             3.4.28.1 [The Lexical Space](#dateTimeStamp-lexical-mapping)
             3.4.28.2 [Facets](#dateTimeStamp-facets)

 This section gives conceptual definitions for all [built-in](#dt-built-in) [ordinary](#dt-ordinary) datatypes defined by this specification. The XML representation used to define [ordinary](#dt-ordinary) datatypes (whether [built-in](#dt-built-in) or [user-defined](#dt-user-defined)) is given in [XML Representation of Simple Type Definition Schema Components (§4.1.2)](#xr-defn) and the complete definitions of the [built-in](#dt-built-in) [ordinary](#dt-ordinary) datatypes are provided in the appendix [Schema for Schema Documents (Datatypes) (normative) (§A)](#schema).

#### <a id="normalizedString"></a>3.4.1 normalizedString

 <a id="dt-normalizedString">[Definition:]</a> **normalizedString** represents white space normalized strings.  The [value space](#dt-value-space) of **normalizedString** is the set of strings that do not contain the carriage return (#xD), line feed (#xA) nor tab (#x9) characters.  The [lexical space](#dt-lexical-space) of **normalizedString** is the set of strings that do not contain the carriage return (#xD), line feed (#xA) nor tab (#x9) characters.  The [base type](#dt-basetype) of **normalizedString** is [string](#string).

##### <a id="normalizedString-facets"></a>3.4.1.1 Facets

The [normalizedString](#normalizedString) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [whiteSpace](#rf-whiteSpace) = ***replace***

Datatypes derived by restriction from [normalizedString](#normalizedString) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [normalizedString](#normalizedString) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

##### <a id="normalizedString-derived-types"></a>3.4.1.2 Derived datatypes

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [normalizedString](#normalizedString)

- [token](#token)

#### <a id="token"></a>3.4.2 token

 <a id="dt-token">[Definition:]</a> **token** represents tokenized strings. The [value space](#dt-value-space) of **token** is the set of strings that do not contain the carriage return (#xD), line feed (#xA) nor tab (#x9) characters, that have no leading or trailing spaces (#x20) and that have no internal sequences of two or more spaces. The [lexical space](#dt-lexical-space) of **token** is the set of strings that do not contain the carriage return (#xD), line feed (#xA) nor tab (#x9) characters, that have no leading or trailing spaces (#x20) and that have no internal sequences of two or more spaces. The [base type](#dt-basetype) of **token** is [normalizedString](#normalizedString).

##### <a id="token-facets"></a>3.4.2.1 Facets

The [token](#token) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse***

Datatypes derived by restriction from [token](#token) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [token](#token) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

##### <a id="token-derived-types"></a>3.4.2.2 Derived datatypes

 The following [built-in](#dt-built-in) datatypes are [derived](#dt-derived) from [token](#token)

- [language](#language)
- [NMTOKEN](#NMTOKEN)
- [Name](#Name)

#### <a id="language"></a>3.4.3 language

<a id="dt-language">[Definition:]</a>**language** represents formal natural language identifiers, as defined by [[BCP 47]](#BCP47) (currently represented by [[RFC 4646]](#RFC4646) and [[RFC 4647]](#RFC4647)) or its successor(s).  The [value space](#dt-value-space) and [lexical space](#dt-lexical-space) of [language](#language) are the set of all strings that conform to the pattern

> > `[a-zA-Z]{1,8}(-[a-zA-Z0-9]{1,8})*`

 This is the set of strings accepted by the grammar given in [[RFC 3066]](#RFC3066), which is now obsolete; the current specification of language codes is more restrictive.  The [base type](#dt-basetype) of [language](#language) is [token](#token). **Note:** The regular expression above provides the only normative constraint on the lexical and value spaces of this type. The additional constraints imposed on language identifiers by [[BCP 47]](#BCP47) and its successor(s), and in particular their requirement that language codes be registered with IANA or ISO if not given in ISO 639, are not part of this datatype as defined here.**Note:** [[BCP 47]](#BCP47) specifies that language codes "are to be treated as case insensitive; there exist conventions for capitalization of some of the subtags, but these MUST NOT be taken to carry meaning." Since the [language](#language) datatype is derived from [string](#string), it inherits from [string](#string) a one-to-one mapping from lexical representations to values. The literals '`MN`' and '`mn`' (for Mongolian) therefore correspond to distinct values and have distinct canonical forms. Users of this specification should be aware of this fact, the consequence of which is that the case-insensitive treatment of language values prescribed by [[BCP 47]](#BCP47) does not follow from the definition of this datatype given here; applications which require case-insensitivity should make appropriate adjustments.<a id="xml.lang.and.language"></a>**Note:** The empty string is not a member of the [value space](#dt-value-space) of [language](#language). Some constructs which normally take language codes as their values, however, also allow the empty string. The attribute `xml:lang` defined by [[XML]](#XML) is one example; there, the empty string overrides a value which would otherwise be inherited, but without specifying a new value.One way to define the desired set of possible values is illustrated by the schema document for the XML namespace at [http://www.w3.org/2001/xml.xsd](https://www.w3.org/2001/xml.xsd), which defines the attribute `xml:lang` as having a type which is a union of [language](#language) and an anonymous type whose only value is the empty string:

```
 <xs:attribute name="lang">
   <xs:annotation>
     <xs:documentation>
       See RFC 3066 at http://www.ietf.org/rfc/rfc3066.txt
       and the IANA registry at
       http://www.iana.org/assignments/lang-tag-apps.htm for
       further information.

       The union allows for the 'un-declaration' of xml:lang with
       the empty string.
     </xs:documentation>
   </xs:annotation>
   <xs:simpleType>
     <xs:union memberTypes="xs:language">
       <xs:simpleType>
         <xs:restriction base="xs:string">
           <xs:enumeration value=""/>
         </xs:restriction>
       </xs:simpleType>
     </xs:union>
   </xs:simpleType>
 </xs:attribute>

```

##### <a id="language-facets"></a>3.4.3.1 Facets

The [language](#language) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[a-zA-Z]{1,8}(-[a-zA-Z0-9]{1,8})****
- [whiteSpace](#rf-whiteSpace) = ***collapse***

Datatypes derived by restriction from [language](#language) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [language](#language) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="NMTOKEN"></a>3.4.4 NMTOKEN

<a id="dt-NMTOKEN">[Definition:]</a> **NMTOKEN** represents the [NMTOKEN attribute type](https://www.w3.org/TR/xml11/#NT-TokenizedType) from [[XML]](#XML). The [value space](#dt-value-space) of **NMTOKEN** is the set of tokens that [match](#dt-match) the [Nmtoken](https://www.w3.org/TR/xml11/#NT-Nmtoken) production in [[XML]](#XML). The [lexical space](#dt-lexical-space) of **NMTOKEN** is the set of strings that [match](#dt-match) the [Nmtoken](https://www.w3.org/TR/xml11/#NT-Nmtoken) production in [[XML]](#XML).  The [base type](#dt-basetype) of **NMTOKEN** is [token](#token).

It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports the [NMTOKEN](https://www.w3.org/TR/xml11/#NT-Nmtoken) production from [[XML]](#XML), or that from [[XML 1.0]](#XML1.0), or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

For compatibility (see [Terminology (§1.6)](#terminology) [NMTOKEN](#NMTOKEN) should be used only on attributes.

##### <a id="NMTOKEN-facets"></a>3.4.4.1 Facets

The [NMTOKEN](#NMTOKEN) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***\c+***
- [whiteSpace](#rf-whiteSpace) = ***collapse***

Datatypes derived by restriction from [NMTOKEN](#NMTOKEN) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [NMTOKEN](#NMTOKEN) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

##### <a id="NMTOKEN-derived-types"></a>3.4.4.2 Related datatypes

 The following [built-in](#dt-built-in) datatype is [constructed](#dt-constructed) from [NMTOKEN](#NMTOKEN)

- [NMTOKENS](#NMTOKENS)

#### <a id="NMTOKENS"></a>3.4.5 NMTOKENS

<a id="dt-NMTOKENS">[Definition:]</a>**NMTOKENS** represents the [NMTOKENS attribute type](https://www.w3.org/TR/xml11/#NT-TokenizedType) from [[XML]](#XML). The [value space](#dt-value-space) of **NMTOKENS** is the set of finite, non-zero-length sequences of [NMTOKEN](#dt-NMTOKEN)s.  The [lexical space](#dt-lexical-space) of **NMTOKENS** is the set of space-separated lists of tokens, of which each token is in the [lexical space](#dt-lexical-space) of [NMTOKEN](#NMTOKEN).  The [item type](#dt-itemType) of **NMTOKENS** is [NMTOKEN](#NMTOKEN). [NMTOKENS](#NMTOKENS) is derived from [`anySimpleType`](#dt-anySimpleType) in two steps: an anonymous list type is defined, whose [item type](#dt-itemType) is [NMTOKEN](#NMTOKEN); this is the [base type](#dt-basetype) of [NMTOKENS](#NMTOKENS), which restricts its value space to lists with at least one item.

 For compatibility (see [Terminology (§1.6)](#terminology)) [NMTOKENS](#NMTOKENS) should be used only on attributes.

##### <a id="NMTOKENS-facets"></a>3.4.5.1 Facets

The [NMTOKENS](#NMTOKENS) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [minLength](#rf-minLength) = ***1***
- [whiteSpace](#rf-whiteSpace) = ***collapse***

Datatypes derived by restriction from [NMTOKENS](#NMTOKENS) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [maxLength](#rf-maxLength)
- [enumeration](#rf-enumeration)
- [pattern](#rf-pattern)
- [assertions](#rf-assertions)

The [NMTOKENS](#NMTOKENS) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="Name"></a>3.4.6 Name

 <a id="dt-Name">[Definition:]</a> **Name** represents [XML Names](https://www.w3.org/TR/xml11/#dt-name). The [value space](#dt-value-space) of **Name** is the set of all strings which [match](#dt-match) the [Name](https://www.w3.org/TR/xml11/#NT-Name) production of [[XML]](#XML).  The [lexical space](#dt-lexical-space) of **Name** is the set of all strings which [match](#dt-match) the [Name](https://www.w3.org/TR/xml11/#NT-Name) production of [[XML]](#XML). The [base type](#dt-basetype) of **Name** is [token](#token).

It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports the [Name](https://www.w3.org/TR/xml11/#NT-Name) production from [[XML]](#XML), or that from [[XML 1.0]](#XML1.0), or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

##### <a id="Name-facets"></a>3.4.6.1 Facets

The [Name](#Name) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***\i\c****
- [whiteSpace](#rf-whiteSpace) = ***collapse***

Datatypes derived by restriction from [Name](#Name) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [Name](#Name) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

##### <a id="Name-derived-types"></a>3.4.6.2 Derived datatypes

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [Name](#Name)

- [NCName](#NCName)

#### <a id="NCName"></a>3.4.7 NCName

 <a id="dt-NCName">[Definition:]</a> **NCName** represents XML "non-colonized" Names.  The [value space](#dt-value-space) of **NCName** is the set of all strings which [match](#dt-match) the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production of [[Namespaces in XML]](#XMLNS).  The [lexical space](#dt-lexical-space) of **NCName** is the set of all strings which [match](#dt-match) the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production of [[Namespaces in XML]](#XMLNS).  The [base type](#dt-basetype) of **NCName** is [Name](#Name).

It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production from [[Namespaces in XML]](#XMLNS), or that from [[Namespaces in XML 1.0]](#XMLNS1.0), or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

##### <a id="NCName-facets"></a>3.4.7.1 Facets

The [NCName](#NCName) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***\i\c* ∩ [\i-[:]][\c-[:]]****
- [whiteSpace](#rf-whiteSpace) = ***collapse***

Datatypes derived by restriction from [NCName](#NCName) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [NCName](#NCName) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

##### <a id="NCName-derived-types"></a>3.4.7.2 Derived datatypes

 The following [built-in](#dt-built-in) datatypes are [derived](#dt-derived) from [NCName](#NCName)

- [ID](#ID)
- [IDREF](#IDREF)
- [ENTITY](#ENTITY)

#### <a id="ID"></a>3.4.8 ID

<a id="dt-ID">[Definition:]</a>**ID** represents the [ID attribute type](https://www.w3.org/TR/xml11/#NT-TokenizedType) from [[XML]](#XML).  The [value space](#dt-value-space) of **ID** is the set of all strings that [match](#dt-match) the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production in [[Namespaces in XML]](#XMLNS).  The [lexical space](#dt-lexical-space) of **ID** is the set of all strings that [match](#dt-match) the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production in [[Namespaces in XML]](#XMLNS). The [base type](#dt-basetype) of **ID** is [NCName](#NCName).

**Note:** It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production from [[Namespaces in XML]](#XMLNS), or that from [[Namespaces in XML 1.0]](#XMLNS1.0), or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

For compatibility (see [Terminology (§1.6)](#terminology)), [ID](#ID) should be used only on attributes.

**Note:** Uniqueness of items validated as [ID](#ID) is not part of this datatype as defined here. When this specification is used in conjunction with [[XSD 1.1 Part 1: Structures]](#structural-schemas), uniqueness is enforced at a different level, not as part of datatype validity; see [Validation Rule: Validation Root Valid (ID/IDREF)](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#cvc-id) in [[XSD 1.1 Part 1: Structures]](#structural-schemas).

##### <a id="ID-facets"></a>3.4.8.1 Facets

The [ID](#ID) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***\i\c* ∩ [\i-[:]][\c-[:]]****
- [whiteSpace](#rf-whiteSpace) = ***collapse***

Datatypes derived by restriction from [ID](#ID) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [ID](#ID) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="IDREF"></a>3.4.9 IDREF

 <a id="dt-IDREF">[Definition:]</a> **IDREF** represents the [IDREF attribute type](https://www.w3.org/TR/xml11/#NT-TokenizedType) from [[XML]](#XML).  The [value space](#dt-value-space) of **IDREF** is the set of all strings that [match](#dt-match) the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production in [[Namespaces in XML]](#XMLNS).  The [lexical space](#dt-lexical-space) of **IDREF** is the set of strings that [match](#dt-match) the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production in [[Namespaces in XML]](#XMLNS). The [base type](#dt-basetype) of **IDREF** is [NCName](#NCName).

**Note:** It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production from [[Namespaces in XML]](#XMLNS), or that from [[Namespaces in XML 1.0]](#XMLNS1.0), or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

 For compatibility (see [Terminology (§1.6)](#terminology)) this datatype should be used only on attributes.

**Note:** Existence of referents for items validated as [IDREF](#IDREF) is not part of this datatype as defined here. When this specification is used in conjunction with [[XSD 1.1 Part 1: Structures]](#structural-schemas), referential integrity is enforced at a different level, not as part of datatype validity; see [Validation Rule: Validation Root Valid (ID/IDREF)](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#cvc-id) in [[XSD 1.1 Part 1: Structures]](#structural-schemas).

##### <a id="IDREF-facets"></a>3.4.9.1 Facets

The [IDREF](#IDREF) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***\i\c* ∩ [\i-[:]][\c-[:]]****
- [whiteSpace](#rf-whiteSpace) = ***collapse***

Datatypes derived by restriction from [IDREF](#IDREF) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [IDREF](#IDREF) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

##### <a id="IDREF-derived-types"></a>3.4.9.2 Related datatypes

 The following [built-in](#dt-built-in) datatype is [constructed](#dt-constructed) from [IDREF](#IDREF)

- [IDREFS](#IDREFS)

#### <a id="IDREFS"></a>3.4.10 IDREFS

<a id="dt-IDREFS">[Definition:]</a> **IDREFS** represents the [IDREFS attribute type](https://www.w3.org/TR/xml11/#NT-TokenizedType) from [[XML]](#XML).  The [value space](#dt-value-space) of **IDREFS** is the set of finite, non-zero-length sequences of [IDREF](#IDREF)s. The [lexical space](#dt-lexical-space) of **IDREFS** is the set of space-separated lists of tokens, of which each token is in the [lexical space](#dt-lexical-space) of [IDREF](#IDREF).  The [item type](#dt-itemType) of **IDREFS** is [IDREF](#IDREF). [IDREFS](#IDREFS) is derived from [`anySimpleType`](#dt-anySimpleType) in two steps: an anonymous list type is defined, whose [item type](#dt-itemType) is [IDREF](#IDREF); this is the [base type](#dt-basetype) of [IDREFS](#IDREFS), which restricts its value space to lists with at least one item.

For compatibility (see [Terminology (§1.6)](#terminology)) [IDREFS](#IDREFS) should be used only on attributes.

**Note:** Existence of referents for items validated as [IDREFS](#IDREFS) is not part of this datatype as defined here. When this specification is used in conjunction with [[XSD 1.1 Part 1: Structures]](#structural-schemas), referential integrity is enforced at a different level, not as part of datatype validity; see [Validation Rule: Validation Root Valid (ID/IDREF)](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#cvc-id) in [[XSD 1.1 Part 1: Structures]](#structural-schemas).

##### <a id="IDREFS-facets"></a>3.4.10.1 Facets

The [IDREFS](#IDREFS) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [minLength](#rf-minLength) = ***1***
- [whiteSpace](#rf-whiteSpace) = ***collapse***

Datatypes derived by restriction from [IDREFS](#IDREFS) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [maxLength](#rf-maxLength)
- [enumeration](#rf-enumeration)
- [pattern](#rf-pattern)
- [assertions](#rf-assertions)

The [IDREFS](#IDREFS) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="ENTITY"></a>3.4.11 ENTITY

 <a id="dt-ENTITY">[Definition:]</a> **ENTITY** represents the [ENTITY](https://www.w3.org/TR/xml11/#NT-TokenizedType) attribute type from [[XML]](#XML).  The [value space](#dt-value-space) of **ENTITY** is the set of all strings that [match](#dt-match) the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production in [[Namespaces in XML]](#XMLNS) and have been declared as an [unparsed entity](https://www.w3.org/TR/xml11/#dt-unparsed) in a [document type definition](https://www.w3.org/TR/xml11/#dt-doctype). The [lexical space](#dt-lexical-space) of **ENTITY** is the set of all strings that [match](#dt-match) the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production in [[Namespaces in XML]](#XMLNS). The [base type](#dt-basetype) of **ENTITY** is [NCName](#NCName).

**Note:** It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports the [NCName](https://www.w3.org/TR/xml-names11/#NT-NCName) production from [[Namespaces in XML]](#XMLNS), or that from [[Namespaces in XML 1.0]](#XMLNS1.0), or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork). **Note:** The [value space](#dt-value-space) of [ENTITY](#ENTITY) is scoped to a specific instance document.

For compatibility (see [Terminology (§1.6)](#terminology)) [ENTITY](#ENTITY) should be used only on attributes.

##### <a id="ENTITY-facets"></a>3.4.11.1 Facets

The [ENTITY](#ENTITY) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***\i\c* ∩ [\i-[:]][\c-[:]]****
- [whiteSpace](#rf-whiteSpace) = ***collapse***

Datatypes derived by restriction from [ENTITY](#ENTITY) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [minLength](#rf-minLength)
- [maxLength](#rf-maxLength)
- [enumeration](#rf-enumeration)
- [assertions](#rf-assertions)

The [ENTITY](#ENTITY) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

##### <a id="ENTITY-derived-types"></a>3.4.11.2 Related datatypes

 The following [built-in](#dt-built-in) datatype is [constructed](#dt-constructed) from [ENTITY](#ENTITY)

- [ENTITIES](#ENTITIES)

#### <a id="ENTITIES"></a>3.4.12 ENTITIES

<a id="dt-ENTITIES">[Definition:]</a>**ENTITIES** represents the [ENTITIES attribute type](https://www.w3.org/TR/xml11/#NT-TokenizedType) from [[XML]](#XML).  The [value space](#dt-value-space) of **ENTITIES** is the set of finite, non-zero-length sequences of [ENTITY](#dt-ENTITY) values that have been declared as [unparsed entities](https://www.w3.org/TR/xml11/#dt-unparsed) in a [document type definition](https://www.w3.org/TR/xml11/#dt-doctype).  The [lexical space](#dt-lexical-space) of **ENTITIES** is the set of space-separated lists of tokens, of which each token is in the [lexical space](#dt-lexical-space) of [ENTITY](#ENTITY).  The [item type](#dt-itemType) of **ENTITIES** is [ENTITY](#ENTITY). [ENTITIES](#ENTITIES) is derived from [`anySimpleType`](#dt-anySimpleType) in two steps: an anonymous list type is defined, whose [item type](#dt-itemType) is [ENTITY](#ENTITY); this is the [base type](#dt-basetype) of [ENTITIES](#ENTITIES), which restricts its value space to lists with at least one item.

**Note:** The [value space](#dt-value-space) of [ENTITIES](#ENTITIES) is scoped to a specific instance document.

For compatibility (see [Terminology (§1.6)](#terminology)) [ENTITIES](#ENTITIES) should be used only on attributes.

##### <a id="ENTITIES-facets"></a>3.4.12.1 Facets

The [ENTITIES](#ENTITIES) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [minLength](#rf-minLength) = ***1***
- [whiteSpace](#rf-whiteSpace) = ***collapse***

Datatypes derived by restriction from [ENTITIES](#ENTITIES) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [length](#rf-length)
- [maxLength](#rf-maxLength)
- [enumeration](#rf-enumeration)
- [pattern](#rf-pattern)
- [assertions](#rf-assertions)

The [ENTITIES](#ENTITIES) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***false***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***

#### <a id="integer"></a>3.4.13 integer

<a id="dt-integer-datatype">[Definition:]</a> **integer** is [derived](#dt-derived) from [decimal](#decimal) by fixing the value of [fractionDigits](#dt-fractionDigits) to be 0 and disallowing the trailing decimal point.  This results in the standard mathematical concept of the integer numbers.  The [value space](#dt-value-space) of **integer** is the infinite set {...,-2,-1,0,1,2,...}.  The [base type](#dt-basetype) of **integer** is [decimal](#decimal).

##### <a id="integer-lexical-representation"></a>3.4.13.1 Lexical representation

[integer](#integer) has a lexical representation consisting of a finite-length sequence of one or more decimal digits (#x30-#x39) with an optional leading sign.  If the sign is omitted, "+" is assumed.  For example: -1, 0, 12678967543233, +100000.

##### <a id="integer-canonical-repr"></a>3.4.13.2 Canonical representation

The [canonical representation](#dt-canonical-representation) for [integer](#integer) is defined by prohibiting certain options from the [Lexical representation (§3.4.13.1)](#integer-lexical-representation).  Specifically, the preceding optional "+" sign is prohibited and leading zeroes are prohibited.

##### <a id="integer-facets"></a>3.4.13.3 Facets

The [integer](#integer) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [integer](#integer) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***

Datatypes derived by restriction from [integer](#integer) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [integer](#integer) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***true***

##### <a id="integer-derived-types"></a>3.4.13.4 Derived datatypes

 The following [built-in](#dt-built-in) datatypes are [derived](#dt-derived) from [integer](#integer)

- [nonPositiveInteger](#nonPositiveInteger)
- [long](#long)
- [nonNegativeInteger](#nonNegativeInteger)

#### <a id="nonPositiveInteger"></a>3.4.14 nonPositiveInteger

<a id="dt-nonPositiveInteger">[Definition:]</a> **nonPositiveInteger** is [derived](#dt-derived) from [integer](#integer) by setting the value of [maxInclusive](#dt-maxInclusive) to be 0.  This results in the standard mathematical concept of the non-positive integers. The [value space](#dt-value-space) of **nonPositiveInteger** is the infinite set {...,-2,-1,0}.  The [base type](#dt-basetype) of **nonPositiveInteger** is [integer](#integer).

##### <a id="nonPositiveInteger-lexical-representation"></a>3.4.14.1 Lexical representation

[nonPositiveInteger](#nonPositiveInteger) has a lexical representation consisting of an optional preceding sign followed by a non-empty finite-length sequence of decimal digits (#x30-#x39).  The sign may be "+" or may be omitted only for lexical forms denoting zero; in all other lexical forms, the negative sign ('`-`') **must** be present.  For example: -1, 0, -12678967543233, -100000.

##### <a id="nonPositiveInteger-canonical-repr"></a>3.4.14.2 Canonical representation

The [canonical representation](#dt-canonical-representation) for [nonPositiveInteger](#nonPositiveInteger) is defined by prohibiting certain options from the [Lexical representation (§3.4.14.1)](#nonPositiveInteger-lexical-representation).  In the canonical form for zero, the sign **must** be omitted.  Leading zeroes are prohibited.

##### <a id="nonPositiveInteger-facets"></a>3.4.14.3 Facets

The [nonPositiveInteger](#nonPositiveInteger) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [nonPositiveInteger](#nonPositiveInteger) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [maxInclusive](#rf-maxInclusive) = ***0***

Datatypes derived by restriction from [nonPositiveInteger](#nonPositiveInteger) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [nonPositiveInteger](#nonPositiveInteger) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***true***

##### <a id="nonPositiveInteger-derived-types"></a>3.4.14.4 Derived datatypes

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [nonPositiveInteger](#nonPositiveInteger)

- [negativeInteger](#negativeInteger)

#### <a id="negativeInteger"></a>3.4.15 negativeInteger

<a id="dt-negativeInteger">[Definition:]</a> **negativeInteger** is [derived](#dt-derived) from [nonPositiveInteger](#nonPositiveInteger) by setting the value of [maxInclusive](#dt-maxInclusive) to be -1.  This results in the standard mathematical concept of the negative integers.  The [value space](#dt-value-space) of **negativeInteger** is the infinite set {...,-2,-1}.  The [base type](#dt-basetype) of **negativeInteger** is [nonPositiveInteger](#nonPositiveInteger).

##### <a id="negativeInteger-lexical-representation"></a>3.4.15.1 Lexical representation

[negativeInteger](#negativeInteger) has a lexical representation consisting of a negative sign ('`-`') followed by a non-empty finite-length sequence of decimal digits (#x30-#x39), at least one of which **must** be a digit other than '`0`'.  For example: -1, -12678967543233, -100000.

##### <a id="negativeInteger-canonical-repr"></a>3.4.15.2 Canonical representation

The [canonical representation](#dt-canonical-representation) for [negativeInteger](#negativeInteger) is defined by prohibiting certain options from the [Lexical representation (§3.4.15.1)](#negativeInteger-lexical-representation).  Specifically, leading zeroes are prohibited.

##### <a id="negativeInteger-facets"></a>3.4.15.3 Facets

The [negativeInteger](#negativeInteger) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [negativeInteger](#negativeInteger) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [maxInclusive](#rf-maxInclusive) = ***-1***

Datatypes derived by restriction from [negativeInteger](#negativeInteger) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [negativeInteger](#negativeInteger) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***true***

#### <a id="long"></a>3.4.16 long

<a id="dt-long">[Definition:]</a> **long** is [derived](#dt-derived) from [integer](#integer) by setting the value of [maxInclusive](#dt-maxInclusive) to be 9223372036854775807 and [minInclusive](#dt-minInclusive) to be -9223372036854775808. The [base type](#dt-basetype) of **long** is [integer](#integer).

##### <a id="long-lexical-representation"></a>3.4.16.1 Lexical Representation

[long](#long) has a lexical representation consisting of an optional sign followed by a non-empty finite-length sequence of decimal digits (#x30-#x39).  If the sign is omitted, "+" is assumed.  For example: -1, 0, 12678967543233, +100000.

##### <a id="long-canonical-repr"></a>3.4.16.2 Canonical Representation

The [canonical representation](#dt-canonical-representation) for [long](#long) is defined by prohibiting certain options from the [Lexical Representation (§3.4.16.1)](#long-lexical-representation).  Specifically, the the optional "+" sign is prohibited and leading zeroes are prohibited.

##### <a id="long-facets"></a>3.4.16.3 Facets

The [long](#long) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [long](#long) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [maxInclusive](#rf-maxInclusive) = ***9223372036854775807***
- [minInclusive](#rf-minInclusive) = ***-9223372036854775808***

Datatypes derived by restriction from [long](#long) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxExclusive](#rf-maxExclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [long](#long) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***true***
- [cardinality](#rf-cardinality) = ***finite***
- [numeric](#rf-numeric) = ***true***

##### <a id="long-derived-types"></a>3.4.16.4 Derived datatypes

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [long](#long)

- [int](#int)

#### <a id="int"></a>3.4.17 int

<a id="dt-int">[Definition:]</a> **int** is [derived](#dt-derived) from [long](#long) by setting the value of [maxInclusive](#dt-maxInclusive) to be 2147483647 and [minInclusive](#dt-minInclusive) to be -2147483648.  The [base type](#dt-basetype) of **int** is [long](#long).

##### <a id="int-lexical-representation"></a>3.4.17.1 Lexical Representation

[int](#int) has a lexical representation consisting of an optional sign followed by a non-empty finite-length sequence of decimal digits (#x30-#x39).  If the sign is omitted, "+" is assumed. For example: -1, 0, 126789675, +100000.

##### <a id="int-canonical-repr"></a>3.4.17.2 Canonical representation

The [canonical representation](#dt-canonical-representation) for [int](#int) is defined by prohibiting certain options from the [Lexical Representation (§3.4.17.1)](#int-lexical-representation).  Specifically, the the optional "+" sign is prohibited and leading zeroes are prohibited.

##### <a id="int-facets"></a>3.4.17.3 Facets

The [int](#int) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [int](#int) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [maxInclusive](#rf-maxInclusive) = ***2147483647***
- [minInclusive](#rf-minInclusive) = ***-2147483648***

Datatypes derived by restriction from [int](#int) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxExclusive](#rf-maxExclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [int](#int) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***true***
- [cardinality](#rf-cardinality) = ***finite***
- [numeric](#rf-numeric) = ***true***

##### <a id="int-derived-types"></a>3.4.17.4 Derived datatypes

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [int](#int)

- [short](#short)

#### <a id="short"></a>3.4.18 short

<a id="dt-short">[Definition:]</a> **short** is [derived](#dt-derived) from [int](#int) by setting the value of [maxInclusive](#dt-maxInclusive) to be 32767 and [minInclusive](#dt-minInclusive) to be -32768.  The [base type](#dt-basetype) of **short** is [int](#int).

##### <a id="short-lexical-representation"></a>3.4.18.1 Lexical representation

[short](#short) has a lexical representation consisting of an optional sign followed by a non-empty finite-length sequence of decimal digits (#x30-#x39).  If the sign is omitted, "+" is assumed. For example: -1, 0, 12678, +10000.

##### <a id="short-canonical-repr"></a>3.4.18.2 Canonical representation

The [canonical representation](#dt-canonical-representation) for [short](#short) is defined by prohibiting certain options from the [Lexical representation (§3.4.18.1)](#short-lexical-representation).  Specifically, the the optional "+" sign is prohibited and leading zeroes are prohibited.

##### <a id="short-facets"></a>3.4.18.3 Facets

The [short](#short) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [short](#short) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [maxInclusive](#rf-maxInclusive) = ***32767***
- [minInclusive](#rf-minInclusive) = ***-32768***

Datatypes derived by restriction from [short](#short) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxExclusive](#rf-maxExclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [short](#short) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***true***
- [cardinality](#rf-cardinality) = ***finite***
- [numeric](#rf-numeric) = ***true***

##### <a id="short-derived-types"></a>3.4.18.4 Derived datatypes

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [short](#short)

- [byte](#byte)

#### <a id="byte"></a>3.4.19 byte

<a id="dt-byte">[Definition:]</a> **byte** is [derived](#dt-derived) from [short](#short) by setting the value of [maxInclusive](#dt-maxInclusive) to be 127 and [minInclusive](#dt-minInclusive) to be -128. The [base type](#dt-basetype) of **byte** is [short](#short).

##### <a id="byte-lexical-representation"></a>3.4.19.1 Lexical representation

[byte](#byte) has a lexical representation consisting of an optional sign followed by a non-empty finite-length sequence of decimal digits (#x30-#x39).  If the sign is omitted, "+" is assumed. For example: -1, 0, 126, +100.

##### <a id="byte-canonical-repr"></a>3.4.19.2 Canonical representation

The [canonical representation](#dt-canonical-representation) for [byte](#byte) is defined by prohibiting certain options from the [Lexical representation (§3.4.19.1)](#byte-lexical-representation).  Specifically, the the optional "+" sign is prohibited and leading zeroes are prohibited.

##### <a id="byte-facets"></a>3.4.19.3 Facets

The [byte](#byte) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [byte](#byte) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [maxInclusive](#rf-maxInclusive) = ***127***
- [minInclusive](#rf-minInclusive) = ***-128***

Datatypes derived by restriction from [byte](#byte) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxExclusive](#rf-maxExclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [byte](#byte) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***true***
- [cardinality](#rf-cardinality) = ***finite***
- [numeric](#rf-numeric) = ***true***

#### <a id="nonNegativeInteger"></a>3.4.20 nonNegativeInteger

<a id="dt-nonNegativeInteger">[Definition:]</a> **nonNegativeInteger** is [derived](#dt-derived) from [integer](#integer) by setting the value of [minInclusive](#dt-minInclusive) to be 0.  This results in the standard mathematical concept of the non-negative integers. The [value space](#dt-value-space) of **nonNegativeInteger** is the infinite set {0,1,2,...}.  The [base type](#dt-basetype) of **nonNegativeInteger** is [integer](#integer).

##### <a id="nonNegativeInteger-lexical-representation"></a>3.4.20.1 Lexical representation

[nonNegativeInteger](#nonNegativeInteger) has a lexical representation consisting of an optional sign followed by a non-empty finite-length sequence of decimal digits (#x30-#x39).  If the sign is omitted, the positive sign ('`+`') is assumed. If the sign is present, it **must** be "+" except for lexical forms denoting zero, which may be preceded by a positive ('`+`') or a negative ('`-`') sign. For example: 1, 0, 12678967543233, +100000.

##### <a id="nonNegativeInteger-canonical-repr"></a>3.4.20.2 Canonical representation

The [canonical representation](#dt-canonical-representation) for [nonNegativeInteger](#nonNegativeInteger) is defined by prohibiting certain options from the [Lexical representation (§3.4.20.1)](#nonNegativeInteger-lexical-representation).  Specifically, the the optional "+" sign is prohibited and leading zeroes are prohibited.

##### <a id="nonNegativeInteger-facets"></a>3.4.20.3 Facets

The [nonNegativeInteger](#nonNegativeInteger) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [nonNegativeInteger](#nonNegativeInteger) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [minInclusive](#rf-minInclusive) = ***0***

Datatypes derived by restriction from [nonNegativeInteger](#nonNegativeInteger) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [nonNegativeInteger](#nonNegativeInteger) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***true***

##### <a id="nonNegativeInteger-derived-types"></a>3.4.20.4 Derived datatypes

 The following [built-in](#dt-built-in) datatypes are [derived](#dt-derived) from [nonNegativeInteger](#nonNegativeInteger)

- [unsignedLong](#unsignedLong)
- [positiveInteger](#positiveInteger)

#### <a id="unsignedLong"></a>3.4.21 unsignedLong

 <a id="dt-unsignedLong">[Definition:]</a> **unsignedLong** is [derived](#dt-derived) from [nonNegativeInteger](#nonNegativeInteger) by setting the value of [maxInclusive](#dt-maxInclusive) to be 18446744073709551615.  The [base type](#dt-basetype) of **unsignedLong** is [nonNegativeInteger](#nonNegativeInteger).

##### <a id="unsignedLong-lexical-representation"></a>3.4.21.1 Lexical representation

[unsignedLong](#unsignedLong) has a lexical representation consisting of an optional sign followed by a non-empty finite-length sequence of decimal digits (#x30-#x39).  If the sign is omitted, the positive sign ('`+`') is assumed.  If the sign is present, it **must** be '`+`' except for lexical forms denoting zero, which may be preceded by a positive ('`+`') or a negative ('`-`') sign. For example: 0, 12678967543233, 100000.

##### <a id="unsignedLong-canonical-repr"></a>3.4.21.2 Canonical representation

The [canonical representation](#dt-canonical-representation) for [unsignedLong](#unsignedLong) is defined by prohibiting certain options from the [Lexical representation (§3.4.21.1)](#unsignedLong-lexical-representation).  Specifically, leading zeroes are prohibited.

##### <a id="unsignedLong-facets"></a>3.4.21.3 Facets

The [unsignedLong](#unsignedLong) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [unsignedLong](#unsignedLong) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [maxInclusive](#rf-maxInclusive) = ***18446744073709551615***
- [minInclusive](#rf-minInclusive) = ***0***

Datatypes derived by restriction from [unsignedLong](#unsignedLong) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxExclusive](#rf-maxExclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [unsignedLong](#unsignedLong) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***true***
- [cardinality](#rf-cardinality) = ***finite***
- [numeric](#rf-numeric) = ***true***

##### <a id="unsignedLong-derived-types"></a>3.4.21.4 Derived datatypes

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [unsignedLong](#unsignedLong)

- [unsignedInt](#unsignedInt)

#### <a id="unsignedInt"></a>3.4.22 unsignedInt

<a id="dt-unsignedInt">[Definition:]</a> **unsignedInt** is [derived](#dt-derived) from [unsignedLong](#unsignedLong) by setting the value of [maxInclusive](#dt-maxInclusive) to be 4294967295.  The [base type](#dt-basetype) of **unsignedInt** is [unsignedLong](#unsignedLong).

##### <a id="unsignedInt-lexical-representation"></a>3.4.22.1 Lexical representation

[unsignedInt](#unsignedInt) has a lexical representation consisting of an optional sign followed by a non-empty finite-length sequence of decimal digits (#x30-#x39).  If the sign is omitted, the positive sign ('`+`') is assumed.  If the sign is present, it **must** be '`+`' except for lexical forms denoting zero, which may be preceded by a positive ('`+`') or a negative ('`-`') sign. For example: 0, 1267896754, 100000.

##### <a id="unsignedInt-canonical-repr"></a>3.4.22.2 Canonical representation

The [canonical representation](#dt-canonical-representation) for [unsignedInt](#unsignedInt) is defined by prohibiting certain options from the [Lexical representation (§3.4.22.1)](#unsignedInt-lexical-representation).  Specifically, leading zeroes are prohibited.

##### <a id="unsignedInt-facets"></a>3.4.22.3 Facets

The [unsignedInt](#unsignedInt) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [unsignedInt](#unsignedInt) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [maxInclusive](#rf-maxInclusive) = ***4294967295***
- [minInclusive](#rf-minInclusive) = ***0***

Datatypes derived by restriction from [unsignedInt](#unsignedInt) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxExclusive](#rf-maxExclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [unsignedInt](#unsignedInt) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***true***
- [cardinality](#rf-cardinality) = ***finite***
- [numeric](#rf-numeric) = ***true***

##### <a id="unsignedInt-derived-types"></a>3.4.22.4 Derived datatypes

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [unsignedInt](#unsignedInt)

- [unsignedShort](#unsignedShort)

#### <a id="unsignedShort"></a>3.4.23 unsignedShort

<a id="dt-unsignedShort">[Definition:]</a> **unsignedShort** is [derived](#dt-derived) from [unsignedInt](#unsignedInt) by setting the value of [maxInclusive](#dt-maxInclusive) to be 65535.  The [base type](#dt-basetype) of **unsignedShort** is [unsignedInt](#unsignedInt).

##### <a id="unsignedShort-lexical-representation"></a>3.4.23.1 Lexical representation

[unsignedShort](#unsignedShort) has a lexical representation consisting of an optional sign followed by a non-empty finite-length sequence of decimal digits (#x30-#x39). If the sign is omitted, the positive sign ('`+`') is assumed.  If the sign is present, it **must** be '`+`' except for lexical forms denoting zero, which may be preceded by a positive ('`+`') or a negative ('`-`') sign.  For example: 0, 12678, 10000.

##### <a id="unsignedShort-canonical-repr"></a>3.4.23.2 Canonical representation

The [canonical representation](#dt-canonical-representation) for [unsignedShort](#unsignedShort) is defined by prohibiting certain options from the [Lexical representation (§3.4.23.1)](#unsignedShort-lexical-representation).  Specifically, the leading zeroes are prohibited.

##### <a id="unsignedShort-facets"></a>3.4.23.3 Facets

The [unsignedShort](#unsignedShort) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [unsignedShort](#unsignedShort) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [maxInclusive](#rf-maxInclusive) = ***65535***
- [minInclusive](#rf-minInclusive) = ***0***

Datatypes derived by restriction from [unsignedShort](#unsignedShort) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxExclusive](#rf-maxExclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [unsignedShort](#unsignedShort) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***true***
- [cardinality](#rf-cardinality) = ***finite***
- [numeric](#rf-numeric) = ***true***

##### <a id="unsignedShort-derived-types"></a>3.4.23.4 Derived datatypes

 The following [built-in](#dt-built-in) datatype is [derived](#dt-derived) from [unsignedShort](#unsignedShort)

- [unsignedByte](#unsignedByte)

#### <a id="unsignedByte"></a>3.4.24 unsignedByte

<a id="dt-unsignedByte">[Definition:]</a> **unsignedByte** is [derived](#dt-derived) from [unsignedShort](#unsignedShort) by setting the value of [maxInclusive](#dt-maxInclusive) to be 255.  The [base type](#dt-basetype) of **unsignedByte** is [unsignedShort](#unsignedShort).

##### <a id="unsignedByte-lexical-representation"></a>3.4.24.1 Lexical representation

[unsignedByte](#unsignedByte) has a lexical representation consisting of an optional sign followed by a non-empty finite-length sequence of decimal digits (#x30-#x39). If the sign is omitted, the positive sign ('`+`') is assumed.  If the sign is present, it **must** be '`+`' except for lexical forms denoting zero, which may be preceded by a positive ('`+`') or a negative ('`-`') sign.  For example: 0, 126, 100.

##### <a id="unsignedByte-canonical-repr"></a>3.4.24.2 Canonical representation

 The [canonical representation](#dt-canonical-representation) for [unsignedByte](#unsignedByte) is defined by prohibiting certain options from the [Lexical representation (§3.4.24.1)](#unsignedByte-lexical-representation).  Specifically, leading zeroes are prohibited.

##### <a id="unisngedByte-facets"></a>3.4.24.3 Facets

The [unsignedByte](#unsignedByte) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [unsignedByte](#unsignedByte) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [maxInclusive](#rf-maxInclusive) = ***255***
- [minInclusive](#rf-minInclusive) = ***0***

Datatypes derived by restriction from [unsignedByte](#unsignedByte) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxExclusive](#rf-maxExclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [unsignedByte](#unsignedByte) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***true***
- [cardinality](#rf-cardinality) = ***finite***
- [numeric](#rf-numeric) = ***true***

#### <a id="positiveInteger"></a>3.4.25 positiveInteger

<a id="dt-positiveInteger">[Definition:]</a> **positiveInteger** is [derived](#dt-derived) from [nonNegativeInteger](#nonNegativeInteger) by setting the value of [minInclusive](#dt-minInclusive) to be 1.  This results in the standard mathematical concept of the positive integer numbers.  The [value space](#dt-value-space) of **positiveInteger** is the infinite set {1,2,...}.  The [base type](#dt-basetype) of **positiveInteger** is [nonNegativeInteger](#nonNegativeInteger).

##### <a id="positiveInteger-lexical-representation"></a>3.4.25.1 Lexical representation

[positiveInteger](#positiveInteger) has a lexical representation consisting of an optional positive sign ('`+`') followed by a non-empty finite-length sequence of decimal digits (#x30-#x39), at least one of which **must** be a digit other than '`0`'.  For example: 1, 12678967543233, +100000.

##### <a id="positiveInteger-canonical-repr"></a>3.4.25.2 Canonical representation

 The [canonical representation](#dt-canonical-representation) for [positiveInteger](#positiveInteger) is defined by prohibiting certain options from the [Lexical representation (§3.4.25.1)](#positiveInteger-lexical-representation).  Specifically, the optional "+" sign is prohibited and leading zeroes are prohibited.

##### <a id="positiveInteger-facets"></a>3.4.25.3 Facets

The [positiveInteger](#positiveInteger) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [fractionDigits](#rf-fractionDigits) = ***0*** (fixed)
- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [positiveInteger](#positiveInteger) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[\-+]?[0-9]+***
- [minInclusive](#rf-minInclusive) = ***1***

Datatypes derived by restriction from [positiveInteger](#positiveInteger) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [totalDigits](#rf-totalDigits)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [positiveInteger](#positiveInteger) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***total***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***true***

#### <a id="yearMonthDuration"></a>3.4.26 yearMonthDuration

<a id="dt-yearMonthDuration">[Definition:]</a> **yearMonthDuration** is a datatype [derived](#dt-derived) from [duration](#duration) by restricting its [lexical representations](#dt-lexical-representation) to instances of [*yearMonthDurationLexicalRep*](#nt-yearMonthDurationRep).  The [value space](#dt-value-space) of **yearMonthDuration** is therefore that of [duration](#duration) restricted to those whose [*seconds*](#vp-du-second) property is 0.  This results in a duration datatype which is totally ordered.

**Note:** The always-zero [*seconds*](#vp-du-second) is formally retained in order that [yearMonthDuration](#yearMonthDuration)'s (abstract) value space truly be a subset of that of [duration](#duration)  An obvious implementation optimization is to ignore the zero and implement [yearMonthDuration](#yearMonthDuration) values simply as [integer](#integer) values.

##### <a id="yearMonthDuration-lexical-mapping"></a>3.4.26.1 The [yearMonthDuration](#yearMonthDuration) Lexical Mapping

The lexical space is reduced from that of [duration](#duration) by disallowing [*duDayFrag*](#nt-duDaFrag) and [*duTimeFrag*](#nt-duTFrag) fragments in the [lexical representations](#dt-lexical-representation).  The [yearMonthDuration](#yearMonthDuration) Lexical Representation  <a id="nt-yearMonthDurationRep"></a>[42]   *yearMonthDurationLexicalRep* ::= '`-`'? '`P`' [*duYearMonthFrag*](#nt-duYMFrag)

The lexical space of [yearMonthDuration](#yearMonthDuration) consists of strings which match the regular expression '`-?P((([0-9]+Y)([0-9]+M)?)|([0-9]+M))`' or the expression '`-?P[0-9]+(Y([0-9]+M)?|M)`', but the formal definition of [yearMonthDuration](#yearMonthDuration) uses a simpler regular expression in its [pattern](#dt-pattern) facet: '`[^DT]*`'. This pattern matches only strings of characters which contain no 'D' and no 'T', thus restricting the [lexical space](#dt-lexical-space) of [duration](#duration) to strings with no day, hour, minute, or seconds fields.

The [canonical mapping](#dt-canonical-mapping) is that of [duration](#duration) restricted in its range to the [lexical space](#dt-lexical-space) (which reduces its domain to omit any values not in the [yearMonthDuration](#yearMonthDuration) value space).

**Note:** The [yearMonthDuration](#yearMonthDuration) value whose [*months*](#vp-du-month) and [*seconds*](#vp-du-second) are both zero has no [canonical representation](#dt-canonical-representation) in this datatype since its [canonical representation](#dt-canonical-representation) in [duration](#duration) ('`PT0S`') is not in the [lexical space](#dt-lexical-space) of [yearMonthDuration](#yearMonthDuration).

##### <a id="YearMonthDuration-facets"></a>3.4.26.2 Facets

The [yearMonthDuration](#yearMonthDuration) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [yearMonthDuration](#yearMonthDuration) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[^DT]****

Datatypes derived by restriction from [yearMonthDuration](#yearMonthDuration) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [yearMonthDuration](#yearMonthDuration) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***
**Note:** The [ordered](#ff-o) facet has the value ***partial*** even though the datatype is in fact totally ordered, because (as explained in [ordered (§4.2.1)](#rf-ordered)), the value of that facet is unchanged by derivation.

#### <a id="dayTimeDuration"></a>3.4.27 dayTimeDuration

 <a id="dt-dayTimeDuration">[Definition:]</a> **dayTimeDuration** is a datatype [derived](#dt-derived) from [duration](#duration) by restricting its [lexical representations](#dt-lexical-representation) to instances of [*dayTimeDurationLexicalRep*](#nt-dayTimeDurationRep). The [value space](#dt-value-space) of **dayTimeDuration** is therefore that of [duration](#duration) restricted to those whose [*months*](#vp-du-month) property is 0.  This results in a duration datatype which is totally ordered.

##### <a id="dayTimeDuration-lexical-mapping"></a>3.4.27.1 The [dayTimeDuration](#dayTimeDuration) Lexical Space

 The lexical space is reduced from that of [duration](#duration) by disallowing [*duYearFrag*](#nt-duYrFrag) and [*duMonthFrag*](#nt-duMoFrag) fragments in the [lexical representations](#dt-lexical-representation).

 The [dayTimeDuration](#dayTimeDuration) Lexical Representation  <a id="nt-dayTimeDurationRep"></a>[43]   *dayTimeDurationLexicalRep* ::= '`-`'? '`P`' [*duDayTimeFrag*](#nt-duDTFrag)

The lexical space of [dayTimeDuration](#dayTimeDuration) consists of strings in the [lexical space](#dt-lexical-space) of [duration](#duration) which match the regular expression '`[^YM]*[DT].*`'; this pattern eliminates all durations with year or month fields, leaving only those with day, hour, minutes, and/or seconds fields.

 The [canonical mapping](#dt-canonical-mapping) is that of [duration](#duration) restricted in its range to the [lexical space](#dt-lexical-space) (which reduces its domain to omit any values not in the [dayTimeDuration](#dayTimeDuration) value space).

##### <a id="dayTimeDuration-facets"></a>3.4.27.2 Facets

The [dayTimeDuration](#dayTimeDuration) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)

The [dayTimeDuration](#dayTimeDuration) datatype has the following [constraining facets](#dt-constraining-facet) with the values shown; these facets **may** be specified in the derivation of new types, if the value given is at least as restrictive as the one shown:

- [pattern](#rf-pattern) = ***[^YM]*(T.*)?***

Datatypes derived by restriction from [dayTimeDuration](#dayTimeDuration) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [dayTimeDuration](#dayTimeDuration) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***
**Note:** The [ordered](#ff-o) facet has the value ***partial*** even though the datatype is in fact totally ordered, because (as explained in [ordered (§4.2.1)](#rf-ordered)), the value of that facet is unchanged by derivation.

#### <a id="dateTimeStamp"></a>3.4.28 dateTimeStamp

 <a id="dt-dateTimeStamp">[Definition:]</a> The **dateTimeStamp** datatype is [derived](#dt-derived) from [dateTime](#dateTime) by giving the value ***required*** to its [explicitTimezone](#f-tz) facet. The result is that all values of [dateTimeStamp](#dateTimeStamp) are required to have explicit time zone offsets and the datatype is totally ordered.

##### <a id="dateTimeStamp-lexical-mapping"></a>3.4.28.1 The [dateTimeStamp](#dateTimeStamp) Lexical Space

As a consequence of requiring an explicit time zone offset, the lexical space of [dateTimeStamp](#dateTimeStamp) is reduced from that of [dateTime](#dateTime) by requiring a [*timezoneFrag*](#nt-tzFrag) fragment in the [lexical representations](#dt-lexical-representation).

 The [dateTimeStamp](#dateTimeStamp) Lexical Representation  <a id="nt-dateTimeStampRep"></a>[44]   *dateTimeStampLexicalRep* ::= [*yearFrag*](#nt-yrFrag) '`-`' [*monthFrag*](#nt-moFrag) '`-`' [*dayFrag*](#nt-daFrag) '`T`' (([*hourFrag*](#nt-hrFrag) '`:`' [*minuteFrag*](#nt-miFrag) '`:`' [*secondFrag*](#nt-seFrag)) | [*endOfDayFrag*](#nt-eodFrag)) [*timezoneFrag*](#nt-tzFrag)   **Constraint:**  Day-of-month Representations **Note:** For details of the [Day-of-month Representations (§3.3.7.2)](#con-dateTime-day) constraint, see [dateTime](#dateTime), from which the constraint is inherited.

In other words, the lexical space of [dateTimeStamp](#dateTimeStamp) consists of strings which are in the [lexical space](#dt-lexical-space) of [dateTime](#dateTime) and which also match the regular expression '`.*(Z|(\+|-)[0-9][0-9]:[0-9][0-9])`'.

The [lexical mapping](#dt-lexical-mapping) is that of [dateTime](#dateTime) restricted to the [dateTimeStamp](#dateTimeStamp) lexical space.

The [canonical mapping](#dt-canonical-mapping) is that of [dateTime](#dateTime) restricted to the [dateTimeStamp](#dateTimeStamp) value space.

##### <a id="dateTimeStamp-facets"></a>3.4.28.2 Facets

The [dateTimeStamp](#dateTimeStamp) datatype and all datatypes derived from it by restriction have the following [constraining facets](#dt-constraining-facet) with ***fixed*** values; these facets **must not** be changed from the values shown:

- [whiteSpace](#rf-whiteSpace) = ***collapse*** (fixed)
- [explicitTimezone](#rf-explicitTimezone) = ***required*** (fixed)

Datatypes derived by restriction from [dateTimeStamp](#dateTimeStamp) **may** also specify values for the following [constraining facets](#dt-constraining-facet):

- [pattern](#rf-pattern)
- [enumeration](#rf-enumeration)
- [maxInclusive](#rf-maxInclusive)
- [maxExclusive](#rf-maxExclusive)
- [minInclusive](#rf-minInclusive)
- [minExclusive](#rf-minExclusive)
- [assertions](#rf-assertions)

The [dateTimeStamp](#dateTimeStamp) datatype has the following values for its [fundamental facets](#dt-fundamental-facet):

- [ordered](#rf-ordered) = ***partial***
- [bounded](#rf-bounded) = ***false***
- [cardinality](#rf-cardinality) = ***countably infinite***
- [numeric](#rf-numeric) = ***false***
**Note:** The [ordered](#ff-o) facet has the value ***partial*** even though the datatype is in fact totally ordered, because (as explained in [ordered (§4.2.1)](#rf-ordered)), the value of that facet is unchanged by derivation.

## <a id="datatype-components"></a>4 Datatype components

The preceding sections of this specification have described datatypes in a way largely independent of their use in the particular context of [schema-aware processing](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-va) as defined in [[XSD 1.1 Part 1: Structures]](#structural-schemas).

 This section presents the mechanisms necessary to integrate datatypes into the context of [[XSD 1.1 Part 1: Structures]](#structural-schemas), mostly in terms of the [schema component](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#c) abstraction introduced there. The account of datatypes given in this specification is also intended to be useful in other contexts. Any specification or other formal system intending to use datatypes as defined above, particularly if definition of new datatypes via facet-based restriction is envisaged, will need to provide analogous mechanisms for some, but not necessarily all, of what follows below. For example, the [{target namespace}](#std-target_namespace) and [{final}](#std-final) properties are required because of particular aspects of [[XSD 1.1 Part 1: Structures]](#structural-schemas) which are not in principle necessary for the use of datatypes as defined here.

The following sections provide full details on the properties and significance of each kind of schema component involved in datatype definitions. For each property, the kinds of values it is allowed to have is specified.  Any property not identified as optional is required to be present; optional properties which are not present have [absent](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-null) as their value. Any property identified as a having a set, subset or [list](#dt-list) value may have an empty value unless this is explicitly ruled out: this is not the same as [absent](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-null). Any property value identified as a superset or a subset of some set may be equal to that set, unless a proper superset or subset is explicitly called for.

 For more information on the notion of schema components, see [Schema Component Details](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#components) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

<a id="dt-owner">[Definition:]</a>A component may be referred to as the **owner** of its properties, and of the values of those properties.

### <a id="rf-defn"></a>4.1 Simple Type Definition

        4.1.1 [The Simple Type Definition Schema Component](#dc-defn)
         4.1.2 [XML Representation of Simple Type Definition Schema Components](#xr-defn)
         4.1.3 [Constraints on XML Representation of Simple Type Definition](#defn-rep-constr)
         4.1.4 [Simple Type Definition Validation Rules](#defn-validation-rules)
         4.1.5 [Constraints on Simple Type Definition Schema Components](#defn-coss)
         4.1.6 [Built-in Simple Type Definitions](#builtin-stds)

 Simple Type Definitions provide for:

- In the case of [primitive](#dt-primitive) datatypes, identifying a datatype with its definition in this specification.
- In the case of [constructed](#dt-constructed) datatypes, defining the datatype in terms of other datatypes.
- Attaching a [QName](#QName) to the datatype.

#### <a id="dc-defn"></a>4.1.1 The Simple Type Definition Schema Component

The Simple Type Definition schema component has the following properties:

---

---

  Schema Component: <a id="std">Simple Type Definition</a>   <a id="std-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="std-name"></a>{name}  An xs:NCName value. Optional.  <a id="std-target_namespace"></a>{target namespace}  An xs:anyURI value. Optional.  <a id="std-final"></a>{final}

A subset of `{`***restriction***, ***extension***, ***list***, ***union***`}`

   <a id="std-context"></a>{context}  Required if [{name}](#std-name) is ***absent***, otherwise **must** be ***absent***

Either an [Attribute Declaration](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#ad), an [Element Declaration](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#ed), a [Complex Type Definition](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#ctd) or a [Simple Type Definition](#std).

   <a id="std-base_type_definition"></a>{base type definition}  A [Type Definition](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#td) component. Required.

With one exception, the [{base type definition}](#std-base_type_definition) of any [Simple Type Definition](#std) is a [Simple Type Definition](#std). The exception is [anySimpleType](#anySimpleType-def), which has [anyType](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-anyType), a [Complex Type Definition](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#ctd), as its [{base type definition}](#std-base_type_definition).

   <a id="std-facets"></a>{facets}  A set of [Constraining Facet](#f) components.   <a id="std-fundamental_facets"></a>{fundamental facets}  A set of [Fundamental Facet](#ff) components.   <a id="std-variety"></a>{variety}  One of {atomic, list, union}. Required for all [Simple Type Definition](#std)s except [anySimpleType](#anySimpleType-def), in which it is ***absent***.  <a id="std-primitive_type_definition"></a>{primitive type definition}  A [Simple Type Definition](#std) component. With one exception, required if [{variety}](#std-variety) is ***atomic***, otherwise **must** be ***absent***. The exception is [anyAtomicType](#anyAtomicType-def), whose [{primitive type definition}](#std-primitive_type_definition) is ***absent***.

If not ***absent***, **must** be a [primitive](#dt-primitive) built-in definition.

   <a id="std-item_type_definition"></a>{item type definition}  A [Simple Type Definition](#std) component. Required if [{variety}](#std-variety) is ***list***, otherwise **must** be ***absent***.

 The value of this property **must** be a primitive or ordinary simple type definition with [{variety}](#std-variety) = ***atomic***, or an ordinary simple type definition with [{variety}](#std-variety) = ***union*** whose basic members are all atomic; the value **must not** itself be a list type (have [{variety}](#std-variety) = ***list***) or have any basic members which are list types.

   <a id="std-member_type_definitions"></a>{member type definitions}  A sequence of primitive or ordinary [Simple Type Definition](#std) components.

**Must** be present (but **may** be empty) if [{variety}](#std-variety) is ***union***, otherwise **must** be ***absent***.

The sequence may contain any primitive or ordinary simple type definition, but **must not** contain any special type definitions.

---

---

 Simple type definitions are identified by their [{name}](#std-name) and [{target namespace}](#std-target_namespace).  Except for anonymous [Simple Type Definition](#std)s (those with no [{name}](#std-name)), [Simple Type Definition](#std)s **must** be uniquely identified within a schema. Within a valid schema, each [Simple Type Definition](#std) uniquely determines one datatype. The [value space](#dt-value-space), [lexical space](#dt-lexical-space), [lexical mapping](#dt-lexical-mapping), etc., of a [Simple Type Definition](#std) are the [value space](#dt-value-space), [lexical space](#dt-lexical-space), etc., of the datatype uniquely determined (or "defined") by that [Simple Type Definition](#std).

 If [{variety}](#std-variety) is [atomic](#dt-atomic) then the [value space](#dt-value-space) of the datatype defined will be a subset of the [value space](#dt-value-space) of [{base type definition}](#std-base_type_definition) (which is a subset of the [value space](#dt-value-space) of [{primitive type definition}](#std-primitive_type_definition)). If [{variety}](#std-variety) is [list](#dt-list) then the [value space](#dt-value-space) of the datatype defined will be the set of (possibly empty) finite-length sequences of values from the [value space](#dt-value-space) of [{item type definition}](#std-item_type_definition). If [{variety}](#std-variety) is [union](#dt-union) then the [value space](#dt-value-space) of the datatype defined will be a subset (possibly an improper subset) of the union of the [value spaces](#dt-value-space) of each [Simple Type Definition](#std) in [{member type definitions}](#std-member_type_definitions).

 If [{variety}](#std-variety) is [atomic](#dt-atomic) then the [{variety}](#std-variety) of [{base type definition}](#std-base_type_definition) **must** be [atomic](#dt-atomic), unless the [{base type definition}](#std-base_type_definition) is [anySimpleType](#anySimpleType). If [{variety}](#std-variety) is [list](#dt-list) then the [{variety}](#std-variety) of [{item type definition}](#std-item_type_definition) **must** be either [atomic](#dt-atomic) or [union](#dt-union), and if [{item type definition}](#std-item_type_definition) is [union](#dt-union) then all its [basic members](#dt-basicmember) **must** be [atomic](#dt-atomic). If [{variety}](#std-variety) is [union](#dt-union) then [{member type definitions}](#std-member_type_definitions) **must** be a list of [Simple Type Definition](#std)s.

The [{facets}](#std-facets) property determines the [value space](#dt-value-space) and [lexical space](#dt-lexical-space) of the datatype being defined by imposing constraints which are to be satisfied by all valid values and [lexical representations](#dt-lexical-representation).

 The [{fundamental facets}](#std-fundamental_facets) property provides some basic information about the datatype being defined: its cardinality, whether an ordering is defined for it by this specification, whether it has upper and lower bounds, and whether it is numeric.

 If [{final}](#std-final) is the empty set then the type can be used in deriving other types; the explicit values ***restriction***, ***list*** and ***union*** prevent further derivations of [Simple Type Definition](#std)s by [facet-based restriction](#dt-fb-restriction), [list](#dt-list) and [union](#dt-union) respectively; the explicit value ***extension*** prevents any derivation of [Complex Type Definitions](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#ctd) by extension.

The [{context}](#std-context) property is only relevant for anonymous type definitions, for which its value is the component in which this type definition appears as the value of a property, e.g. [{item type definition}](#std-item_type_definition) or [{base type definition}](#std-base_type_definition).

#### <a id="xr-defn"></a>4.1.2 XML Representation of Simple Type Definition Schema Components

 The XML representation for a [Simple Type Definition](#std) schema component is a [<simpleType>](#element-simpleType) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `simpleType` Element Information Item et al.

<a id="element-simpleType"></a>

```xml
<simpleType
  final =

    (#all | List of (list | union | restriction | extension))

  id = ID
  name = NCName
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?, (restriction | list | union))
</simpleType>
```

<a id="element-restriction"></a>

```xml
<restriction
  base = QName
  id = ID
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?, (simpleType?, (minExclusive | minInclusive | maxExclusive | maxInclusive | totalDigits | fractionDigits | length | minLength | maxLength | enumeration | whiteSpace | pattern | assertion | explicitTimezone | {any with namespace: ##other})*))
</restriction>
```

<a id="element-list"></a>

```xml
<list
  id = ID
  itemType = QName
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?, simpleType?)
</list>
```

<a id="element-union"></a>

```xml
<union
  id = ID
  memberTypes = List of QName
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?, simpleType*)
</union>
```

**[Simple Type Definition](#dc-defn) **Schema Component****

- **Property**: **Representation**
- [{name}](#std-name): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `name` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present on the [<simpleType>](#element-simpleType) element, otherwise [absent](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-null)
- [{target namespace}](#std-target_namespace): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `targetNamespace` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of the parent `schema` element information item, if present, otherwise [absent](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-null).
- [{base type definition}](#std-base_type_definition): The appropriate **case** among the following:

- 1 **If **the [<restriction>](#element-restriction) alternative is chosen, **then **the type definition [resolved](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#src-resolve) to by the [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `base` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of [<restriction>](#element-restriction), if present, otherwise the type definition corresponding to the [<simpleType>](#element-simpleType) among the [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of [<restriction>](#element-restriction).
- 2 **If **the [<list>](#element-list) or [<union>](#element-union) alternative is chosen, **then **[anySimpleType](#anySimpleType-def).
- [{final}](#std-final): A subset of `{`***restriction***, ***extension***, ***list***, ***union***`}`, determined as follows. <a id="lt-vs">[Definition:]</a>Let **FS** be the [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `final` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise the [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `finalDefault` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of the ancestor `schema` element, if present, otherwise the empty string. Then the property value is the appropriate **case** among the following:

- 1 **If **[FS](#lt-vs) is the empty string, **then **the empty set;
- 2 **If **[FS](#lt-vs) is '`#all`', **then **`{`***restriction***, ***extension***, ***list***, ***union***`}`;
- 3 **otherwise **Consider [FS](#lt-vs) as a space-separated list, and include ***restriction*** if '`restriction`' is in that list, and similarly for ***extension***, ***list*** and ***union***.
- [{context}](#std-context): The appropriate **case** among the following:

- 1 **If **the `name` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) is present, **then **[absent](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-null)
- 2 **otherwise ** the appropriate **case** among the following:

- 2.1 **If **the parent element information item is [<attribute>](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#element-attribute), **then **the corresponding [Attribute Declaration](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#ad)
- 2.2 **If **the parent element information item is [<element>](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#element-element), **then **the corresponding [Element Declaration](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#ed)
- 2.3 **If **the parent element information item is [<list>](#element-list) or [<union>](#element-union), **then **the [Simple Type Definition](#std) corresponding to the grandparent [<simpleType>](#element-simpleType) element information item
- 2.4 **otherwise **(the parent element information item is [<restriction>](#element-restriction)), the appropriate **case** among the following:

- 2.4.1 **If **the grandparent element information item is [<simpleType>](#element-simpleType), **then **the [Simple Type Definition](#std) corresponding to the grandparent
- 2.4.2 **otherwise **(the grandparent element information item is [<simpleContent>](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#element-simpleContent)), the [Simple Type Definition](#std) which is the [{content type}](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#ctd-content_type) of the [Complex Type Definition](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#ctd) corresponding to the great-grandparent [<complexType>](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#element-complexType) element information item.
- [{variety}](#std-variety): If the [<list>](#element-list) alternative is chosen, then ***list***, otherwise if the [<union>](#element-union) alternative is chosen, then ***union***, otherwise (the [<restriction>](#element-restriction) alternative is chosen) the [{variety}](#std-variety) of the [{base type definition}](#std-base_type_definition).
- [{facets}](#std-facets): The appropriate **case** among the following:

- 1 **If **the [<restriction>](#element-restriction) alternative is chosen, **then **the set of [Constraining Facet](#f) components obtained by [overlaying](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-facets-overlay) the [{facets}](#std-facets) of the [{base type definition}](#std-base_type_definition) with the set of [Constraining Facet](#f) components corresponding to those [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of [<restriction>](#element-restriction) which specify facets, as defined in [Schema Component Constraint: Simple Type Restriction (Facets)](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#st-restrict-facets).
- 2 **If **the [<list>](#element-list) alternative is chosen, **then **a set with one member, a [whiteSpace](#f-w) facet with [{value}](#f-w-value) = ***collapse*** and [{fixed}](#f-w-fixed) = ***true***.
- 3 **otherwise **the empty set
- [{fundamental facets}](#std-fundamental_facets): Based on [{variety}](#std-variety), [{facets}](#std-facets), [{base type definition}](#std-base_type_definition) and [{member type definitions}](#std-member_type_definitions), a set of [Fundamental Facet](#ff) components, one each as specified in [The ordered Schema Component (§4.2.1.1)](#dc-ordered), [The bounded Schema Component (§4.2.2.1)](#dc-bounded), [The cardinality Schema Component (§4.2.3.1)](#dc-cardinality) and [The numeric Schema Component (§4.2.4.1)](#dc-numeric).
- [{annotations}](#std-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-set) of the set of elements containing the [<simpleType>](#element-simpleType), and the [<restriction>](#element-restriction), the [<list>](#element-list), or the [<union>](#element-union) [[child]](https://www.w3.org/TR/xml-infoset/#infoitem.element), whichever is present, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas). <a id="std-ancestor">[Definition:]</a>The **ancestors** of a [type definition](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#td) are its [{base type definition}](#std-base_type_definition) and the [ancestors](#std-ancestor) of its [{base type definition}](#std-base_type_definition). (The ancestors of a [Simple Type Definition](#std) *T* in the type hierarchy are themselves [type definitions](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#td); they are distinct from the XML elements which may be ancestors, in the XML document hierarchy, of the [<simpleType>](#element-simpleType) element which declares *T*.) If the [{variety}](#std-variety) is ***atomic***, the following additional property mapping also applies:

**[Atomic Simple Type Definition](#xr-defn) **Schema Component****

- **Property**: **Representation**
- [{primitive type definition}](#std-primitive_type_definition): From among the [ancestors](#std-ancestor) of this [Simple Type Definition](#std), that [Simple Type Definition](#std) which corresponds to a [primitive](#dt-primitive) datatype.

*Example*

  An electronic commerce schema might define a datatype called '`SKU`' (the barcode number that appears on products) from the [built-in](#dt-built-in) datatype [string](#string) by supplying a value for the [pattern](#dt-pattern) facet.

```
<simpleType name='SKU'>
    <restriction base='string'>
      <pattern value='\d{3}-[A-Z]{2}'/>
    </restriction>
</simpleType>
```

  In this case, '`SKU`' is the name of the new [user-defined](#dt-user-defined) datatype, [string](#string) is its [base type](#dt-basetype) and [pattern](#dt-pattern) is the facet.

If the [{variety}](#std-variety) is ***list***, the following additional property mappings also apply:

**[List Simple Type Definition](#xr-defn) **Schema Component****

- **Property**: **Representation**
- [{item type definition}](#std-item_type_definition): The appropriate **case** among the following:

- 1 **If **the [{base type definition}](#std-base_type_definition) is [anySimpleType](#anySimpleType-def), **then **the [Simple Type Definition](#std) (a) [resolved](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#src-resolve) to by the [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `itemType` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of [<list>](#element-list), or (b) corresponding to the [<simpleType>](#element-simpleType) among the [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of [<list>](#element-list), whichever is present. **Note:** In this case, a [<list>](#element-list) element will invariably be present; it will invariably have either an `itemType` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) or a [<simpleType>](#element-simpleType) [[child]](https://www.w3.org/TR/xml-infoset/#infoitem.element), but not both.
- 2 **otherwise **(that is, the [{base type definition}](#std-base_type_definition) is not [anySimpleType](#anySimpleType-def)), the [{item type definition}](#std-item_type_definition) of the [{base type definition}](#std-base_type_definition). **Note:** In this case, a [<restriction>](#element-restriction) element will invariably be present.

*Example*

  A system might want to store lists of floating point values.

```
<simpleType name='listOfFloat'>
  <list itemType='float'/>
</simpleType>

```

  In this case, *listOfFloat* is the name of the new [user-defined](#dt-user-defined) datatype, [float](#float) is its [item type](#dt-itemType) and [list](#dt-list) is the derivation method.

If the [{variety}](#std-variety) is ***union***, the following additional property mappings also apply:

**[Union Simple Type Definition](#xr-defn) **Schema Component****

- **Property**: **Representation**
- [{member type definitions}](#std-member_type_definitions): The appropriate **case** among the following:

- 1 **If **the [{base type definition}](#std-base_type_definition) is [anySimpleType](#anySimpleType-def), **then **the sequence of (a) the [Simple Type Definition](#std)s (a) [resolved](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#src-resolve) to by the items in the [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `memberTypes` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of [<union>](#element-union), if any, and (b) those corresponding to the [<simpleType>](#element-simpleType)s among the [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of [<union>](#element-union), if any, in order. **Note:** In this case, a [<union>](#element-union) element will invariably be present; it will invariably have either a `memberTypes` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) or one or more [<simpleType>](#element-simpleType) [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element), or both.
- 2 **otherwise **(that is, the [{base type definition}](#std-base_type_definition) is not [anySimpleType](#anySimpleType-def)), the [{member type definitions}](#std-member_type_definitions) of the [{base type definition}](#std-base_type_definition). **Note:** In this case, a [<restriction>](#element-restriction) element will invariably be present.

*Example*

 As an example, taken from a typical display oriented text markup language, one might want to express font sizes as an integer between 8 and 72, or with one of the tokens "small", "medium" or "large".  The [union](#dt-union) [Simple Type Definition](#std) below would accomplish that.

```
<xs:attribute name="size">
  <xs:simpleType>
    <xs:union>
      <xs:simpleType>
        <xs:restriction base="xs:positiveInteger">
          <xs:minInclusive value="8"/>
          <xs:maxInclusive value="72"/>
        </xs:restriction>
      </xs:simpleType>
      <xs:simpleType>
        <xs:restriction base="xs:NMTOKEN">
          <xs:enumeration value="small"/>
          <xs:enumeration value="medium"/>
          <xs:enumeration value="large"/>
        </xs:restriction>
      </xs:simpleType>
    </xs:union>
  </xs:simpleType>
</xs:attribute>

```

```
<p>
<font size='large'>A header</font>
</p>
<p>
<font size='12'>this is a test</font>
</p>

```

---

 A datatype can be [constructed](#dt-constructed) from a [primitive](#dt-primitive) datatype or an [ordinary](#dt-ordinary) datatype by one of three means: by *[facet-based restriction](#dt-fb-restriction)*, by *[list](#dt-list)* or by *[union](#dt-union)*.

#### <a id="defn-rep-constr"></a>4.1.3 Constraints on XML Representation of Simple Type Definition

<a id="src-list-itemType-or-simpleType"></a>**Schema Representation Constraint: itemType attribute or simpleType child**

> Either the `itemType` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) or the [<simpleType>](#element-simpleType) [[child]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of the [<list>](#element-list) element **must** be present, but not both.

<a id="src-restriction-base-or-simpleType"></a>**Schema Representation Constraint: base attribute or simpleType child**

> Either the `base` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) or the `simpleType` [[child]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of the [<restriction>](#element-restriction) element **must** be present, but not both.

<a id="src-union-memberTypes-or-simpleTypes"></a>**Schema Representation Constraint: memberTypes attribute or simpleType children**

> Either the `memberTypes` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of the [<union>](#element-union) element **must** be non-empty or there **must** be at least one `simpleType` [[child]](https://www.w3.org/TR/xml-infoset/#infoitem.element).

#### <a id="defn-validation-rules"></a>4.1.4 Simple Type Definition Validation Rules

<a id="cvc-facet-valid"></a>**Validation Rule: Facet Valid**

> A value in a [value space](#dt-value-space) is facet-valid with respect to a [constraining facet](#dt-constraining-facet) component if and only if:
>
> - 1  the value is facet-valid with respect to the particular [constraining facet](#dt-constraining-facet) as specified below.

<a id="cvc-datatype-valid"></a>**Validation Rule: Datatype Valid**

> A [literal](#dt-literal) is datatype-valid with respect to a [Simple Type Definition](#std) if and only if it is a member of the [lexical space](#dt-lexical-space) of the corresponding datatype.

**Note:** Since every value in the [value space](#dt-value-space) is denoted by some [literal](#dt-literal), and every [literal](#dt-literal) in the [lexical space](#dt-lexical-space) maps to some value, the requirement that the [literal](#dt-literal) be in the [lexical space](#dt-lexical-space) entails the requirement that the value it maps to should fulfill all of the constraints imposed by the [{facets}](#std-facets) of the datatype. If the datatype is a [list](#dt-list), the Datatype Valid constraint also entails that each whitespace-delimited token in the list be datatype-valid against the [item type](#dt-itemType) of the list. If the datatype is a [union](#dt-union), the Datatype Valid constraint entails that the [literal](#dt-literal) be datatype-valid against at least one of the [member types](#dt-memberTypes).That is, the constraints on [Simple Type Definition](#std)s and on datatype [derivation](#dt-derived) defined in this specification have as a consequence that a [literal](#dt-literal) *L* is datatype-valid with respect to a [Simple Type Definition](#std) *T* if and only if either *T* corresponds to a [special](#dt-special) datatype or **all** of the following are true:

- 1<a id="dv_pattern"></a>If there is a [pattern](#f-p) in [{facets}](#std-facets), then *L* is [pattern valid (§4.3.4.4)](#cvc-pattern-valid) with respect to the [pattern](#f-p). If there are other [lexical](#dt-lexical) facets in [{facets}](#std-facets), then *L* is facet-valid with respect to them.
- 2<a id="dv_lv"></a>The appropriate case among the following is true:

- 2.1<a id="dv_atomic"></a>If the [{variety}](#std-variety) of *T* is [atomic](#dt-atomic), then *L* is in the [lexical space](#dt-lexical-space) of the [{primitive type definition}](#std-primitive_type_definition) of *T*, as defined in the appropriate documentation. Let *V* be the member of the [value space](#dt-value-space) of the [{primitive type definition}](#std-primitive_type_definition) of *T* mapped to by *L*, as defined in the appropriate documentation.**Note:** For [built-in](#dt-built-in) [primitives](#dt-primitive), the "appropriate documentation" is the relevant section of this specification. For [implementation-defined](#key-impl-def) [primitives](#dt-primitive), it is the normative specification of the [primitive](#dt-primitive), which will typically be included in, or referred to from, the implementation's documentation.
- 2.2<a id="dv_list"></a>If the [{variety}](#std-variety) of *T* is [list](#dt-list), then each space-delimited substring of *L* is Datatype Valid with respect to the [{item type definition}](#std-item_type_definition) of *T*. Let *V* be the sequence consisting of the values identified by Datatype Valid for each of those substrings, in order.
- 2.3<a id="dv_union"></a>If the [{variety}](#std-variety) of *T* is [union](#dt-union), then *L* is Datatype Valid with respect to at least one member of the [{member type definitions}](#std-member_type_definitions) of *T*. Let *B* be the [active basic member](#dt-active-basic-member) of *T* for *L*. Let *V* be the value identified by Datatype Valid for *L* with respect to *B*.

- 3<a id="dv_vfacets"></a>*V*, as determined by the appropriate sub-clause of clause [2](#dv_lv) above, is [Facet Valid (§4.1.4)](#cvc-facet-valid) with respect to each member of the [{facets}](#std-facets) of *T* which is a [value-based](#dt-value-based) (and not a [pre-lexical](#dt-pre-lexical) or [lexical](#dt-lexical)) facet.
 Note that [whiteSpace](#f-w) facets and other [pre-lexical](#dt-pre-lexical) facets do not take part in checking Datatype Valid. In cases where this specification is used in conjunction with schema-validation of XML documents, such facets are used to normalize infoset values *before* the normalized results are checked for datatype validity. In the case of unions the [pre-lexical](#dt-pre-lexical) facets to use are those associated with *B* in clause [2.3](#dv_union) above. When more than one [pre-lexical](#dt-pre-lexical) facet applies, the [whiteSpace](#f-w) facet is applied first; the order in which [implementation-defined](#key-impl-def) facets are applied is [implementation-defined](#key-impl-def).

#### <a id="defn-coss"></a>4.1.5 Constraints on Simple Type Definition Schema Components

<a id="cos-applicable-facets"></a>**Schema Component Constraint: Applicable Facets**

> The [constraining facets](#dt-constraining-facet) which are allowed to be members of [{facets}](#std-facets) depend on the [{variety}](#std-variety) and [{primitive type definition}](#std-primitive_type_definition) of the type, as follows:
>
>  If [{variety}](#std-variety) is ***absent***, then no facets are applicable. (This is true for [anySimpleType](#anySimpleType-def).)
>
>
>
>  If [{variety}](#std-variety) is [list](#dt-list), then the applicable facets are [assertions](#dc-assertions), [length](#dt-length), [minLength](#dt-minLength), [maxLength](#dt-maxLength), [pattern](#dt-pattern), [enumeration](#dt-enumeration), and [whiteSpace](#dt-whiteSpace).
>
>
>
>  If [{variety}](#std-variety) is [union](#dt-union), then the applicable facets are [pattern](#dt-pattern), [enumeration](#dt-enumeration), and [assertions](#dc-assertions).
>
>
>
>  If [{variety}](#std-variety) is [atomic](#dt-atomic), and [{primitive type definition}](#std-primitive_type_definition) is ***absent*** then no facets are applicable. (This is true for [anyAtomicType](#anyAtomicType-def).)
>
>
>
>  In all other cases ([{variety}](#std-variety) is [atomic](#dt-atomic) and [{primitive type definition}](#std-primitive_type_definition) is not ***absent***), then the applicable facets are shown in the table below.
>
>
>
> <table>
>         <tbody><tr><th><a href="#std-primitive_type_definition">{primitive type definition}</a></th><th>applicable <a href="#std-facets">{facets}</a></th></tr>
>         <tr><td><a href="#string">string</a></td><td><a href="#dc-length">length</a>, <a href="#dc-minLength">minLength</a>, <a href="#dc-maxLength">maxLength</a>, <a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-assertions">assertions</a></td></tr>
>         <tr><td><a href="#boolean">boolean</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-assertions">assertions</a></td></tr>
>         <tr><td><a href="#float">float</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a></td></tr>
>         <tr><td><a href="#double">double</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a></td></tr>
>         <tr><td><a href="#decimal">decimal</a></td><td><a href="#dc-totalDigits">totalDigits</a>, <a href="#dc-fractionDigits">fractionDigits</a>, <a href="#dc-pattern">pattern</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a></td></tr>
>         <tr><td><a href="#duration">duration</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a></td></tr>
>         <tr><td><a href="#dateTime">dateTime</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a>, <a href="#dc-explicitTimezone">explicitTimezone</a></td></tr>
>         <tr><td><a href="#time">time</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a>, <a href="#dc-explicitTimezone">explicitTimezone</a></td></tr>
>         <tr><td><a href="#date">date</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a>, <a href="#dc-explicitTimezone">explicitTimezone</a></td></tr>
>         <tr><td><a href="#gYearMonth">gYearMonth</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a>, <a href="#dc-explicitTimezone">explicitTimezone</a></td></tr>
>         <tr><td><a href="#gYear">gYear</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a>, <a href="#dc-explicitTimezone">explicitTimezone</a></td></tr>
>         <tr><td><a href="#gMonthDay">gMonthDay</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a>, <a href="#dc-explicitTimezone">explicitTimezone</a></td></tr>
>         <tr><td><a href="#gDay">gDay</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a>, <a href="#dc-explicitTimezone">explicitTimezone</a></td></tr>
>         <tr><td><a href="#gMonth">gMonth</a></td><td><a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-maxInclusive">maxInclusive</a>, <a href="#dc-maxExclusive">maxExclusive</a>, <a href="#dc-minInclusive">minInclusive</a>, <a href="#dc-minExclusive">minExclusive</a>, <a href="#dc-assertions">assertions</a>, <a href="#dc-explicitTimezone">explicitTimezone</a></td></tr>
>         <tr><td><a href="#hexBinary">hexBinary</a></td><td><a href="#dc-length">length</a>, <a href="#dc-minLength">minLength</a>, <a href="#dc-maxLength">maxLength</a>, <a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-assertions">assertions</a></td></tr>
>         <tr><td><a href="#base64Binary">base64Binary</a></td><td><a href="#dc-length">length</a>, <a href="#dc-minLength">minLength</a>, <a href="#dc-maxLength">maxLength</a>, <a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-assertions">assertions</a></td></tr>
>         <tr><td><a href="#anyURI">anyURI</a></td><td><a href="#dc-length">length</a>, <a href="#dc-minLength">minLength</a>, <a href="#dc-maxLength">maxLength</a>, <a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-assertions">assertions</a></td></tr>
>         <tr><td><a href="#QName">QName</a></td><td><a href="#dc-length">length</a>, <a href="#dc-minLength">minLength</a>, <a href="#dc-maxLength">maxLength</a>, <a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-assertions">assertions</a></td></tr>
>         <tr><td><a href="#NOTATION">NOTATION</a></td><td><a href="#dc-length">length</a>, <a href="#dc-minLength">minLength</a>, <a href="#dc-maxLength">maxLength</a>, <a href="#dc-pattern">pattern</a>, <a href="#dc-enumeration">enumeration</a>, <a href="#dc-whiteSpace">whiteSpace</a>, <a href="#dc-assertions">assertions</a></td></tr>
>     </tbody></table>
>
> **Note:** For any [implementation-defined](#key-impl-def) primitive types, it is [implementation-defined](#key-impl-def) which constraining facets are applicable to them. Similarly, for any [implementation-defined](#key-impl-def) constraining facets, it is [implementation-defined](#key-impl-def) which [primitives](#dt-primitive) they apply to.

#### <a id="builtin-stds"></a>4.1.6 Built-in Simple Type Definitions

The [Simple Type Definition](#std) of [anySimpleType](#anySimpleType) is present in every schema.  It has the following properties:

<a id="anySimpleType-def">Simple type definition of `anySimpleType`</a>
  - **Property** = **Value**
  - [{name}](#std-name) = '`anySimpleType`'
  - [{target namespace}](#std-target_namespace) = '`http://www.w3.org/2001/XMLSchema`'
  - [{final}](#std-final) = The empty set
  - [{context}](#std-context) = ***absent***
  - [{base type definition}](#std-base_type_definition) = [anyType](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#any-type-itself)
  - [{facets}](#std-facets) = The empty set
  - [{fundamental facets}](#std-fundamental_facets) = The empty set
  - [{variety}](#std-variety) = ***absent***
  - [{primitive type definition}](#std-primitive_type_definition) = ***absent***
  - [{item type definition}](#std-item_type_definition) = ***absent***
  - [{member type definitions}](#std-member_type_definitions) = ***absent***
  - [{annotations}](#std-annotations) = The empty sequence

The definition of [anySimpleType](#anySimpleType) is the root of the Simple Type Definition hierarchy; as such it mediates between the other simple type definitions, which all eventually trace back to it via their [{base type definition}](#std-base_type_definition) properties, and the definition of ***anyType***, which is *its* [{base type definition}](#std-base_type_definition).

The [Simple Type Definition](#std) of [anyAtomicType](#anyAtomicType) is present in every schema.  It has the following properties:

<a id="anyAtomicType-def">Simple type definition of `anyAtomicType`</a>
  - **Property** = **Value**
  - [{name}](#std-name) = '`anyAtomicType`'
  - [{target namespace}](#std-target_namespace) = '`http://www.w3.org/2001/XMLSchema`'
  - [{final}](#std-final) = The empty set
  - [{context}](#std-context) = ***absent***
  - [{base type definition}](#std-base_type_definition) = [anySimpleType](#anySimpleType)
  - [{facets}](#std-facets) = The empty set
  - [{fundamental facets}](#std-fundamental_facets) = The empty set
  - [{variety}](#std-variety) = ***atomic***
  - [{primitive type definition}](#std-primitive_type_definition) = ***absent***
  - [{item type definition}](#std-item_type_definition) = ***absent***
  - [{member type definitions}](#std-member_type_definitions) = ***absent***
  - [{annotations}](#std-annotations) = The empty sequence

Simple type definitions for all the built-in primitive datatypes, namely [string](#string), [boolean](#boolean), [float](#float), [double](#double), [decimal](#decimal), [dateTime](#dateTime), [duration](#duration), [time](#time), [date](#date), [gMonth](#gMonth), [gMonthDay](#gMonthDay), [gDay](#gDay), [gYear](#gYear), [gYearMonth](#gYearMonth), [hexBinary](#hexBinary), [base64Binary](#base64Binary), [anyURI](#anyURI) are present by definition in every schema.  All have a very similar structure, with only the [{name}](#std-name), the [{primitive type definition}](#std-primitive_type_definition) (which is self-referential), the [{fundamental facets}](#std-fundamental_facets), and in one case the [{facets}](#std-facets) varying from one to the next:

<a id="dummy-def">Simple Type Definition corresponding to the built-in primitive datatypes</a>
  - **Property** = **Value**
  - [{name}](#std-name) = [as appropriate]
  - [{target namespace}](#std-target_namespace) = '`http://www.w3.org/2001/XMLSchema`'
  - [{base type definition}](#std-base_type_definition) = [anyAtomicType Definition](#anyAtomicType-def)
  - [{final}](#std-final) = The empty set
  - [{variety}](#std-variety) = ***atomic***
  - [{primitive type definition}](#std-primitive_type_definition) = [this [Simple Type Definition](#std) itself]
  - [{facets}](#std-facets) = {a [whiteSpace](#f-w) facet with [{value}](#f-w-value) = ***collapse*** and [{fixed}](#f-w-fixed) = ***true*** in all cases except [string](#string), which has [{value}](#f-w-value) = ***preserve*** and [{fixed}](#f-w-fixed) = ***false***}
  - [{fundamental facets}](#std-fundamental_facets) = [as appropriate]
  - [{context}](#std-context) = ***absent***
  - [{item type definition}](#std-item_type_definition) = ***absent***
  - [{member type definitions}](#std-member_type_definitions) = ***absent***
  - [{annotations}](#std-annotations) = The empty sequence [Implementation-defined](#key-impl-def) [primitives](#dt-primitive) **must** have a [Simple Type Definition](#std) with the values shown above, with the following exceptions.

1. The [{facets}](#std-facets) property **must** contain a [whiteSpace](#f-w) facet, the value of which is [implementation-defined](#key-impl-def). It **may** contain other facets, whether defined in this specification or [implementation-defined](#key-impl-def).
2. The value of [{fundamental facets}](#std-fundamental_facets) is [implementation-defined](#key-impl-def).
3. The value of [{annotations}](#std-annotations) **may** be empty, but need not be.
 **Note:** It is a consequence of the rule just given that each [implementation-defined](#key-impl-def) [primitive](#dt-primitive) will have an [expanded name](https://www.w3.org/TR/2004/REC-xml-names11-20040204/#dt-expname) by which it can be referred to.**Note:** [Implementation-defined](#key-impl-def) datatypes will normally have a value other than '`http://www.w3.org/2001/XMLSchema`' for the [{target namespace}](#std-target_namespace) property. That namespace is controlled by the W3C and datatypes will be added to it only by W3C or its designees.

Similarly, [Simple Type Definition](#std)s for all the built-in [ordinary](#dt-ordinary) datatypes are present by definition in every schema, with properties as specified in [Other Built-in Datatypes (§3.4)](#ordinary-built-ins) and as represented in XML in [Illustrative XML representations for the built-in ordinary type definitions (§C.2)](#drvd.nxsd).

<a id="dummy-ddef">Simple Type Definition corresponding to the built-in ordinary datatypes</a>
  - **Property** = **Value**
  - [{name}](#std-name) = [as appropriate]
  - [{target namespace}](#std-target_namespace) = '`http://www.w3.org/2001/XMLSchema`'
  - [{base type definition}](#std-base_type_definition) = [as specified in the appropriate sub-section of [Other Built-in Datatypes (§3.4)](#ordinary-built-ins)]
  - [{final}](#std-final) = The empty set
  - [{variety}](#std-variety) = [***atomic*** or ***list***, as specified in the appropriate sub-section of [Other Built-in Datatypes (§3.4)](#ordinary-built-ins)]
  - [{primitive type definition}](#std-primitive_type_definition) = [if [{variety}](#std-variety) is ***atomic***, then the [{primitive type definition}](#std-primitive_type_definition) of the [{base type definition}](#std-base_type_definition), otherwise ***absent***]
  - [{facets}](#std-facets) = [as specified in the appropriate sub-section of [Other Built-in Datatypes (§3.4)](#ordinary-built-ins)]
  - [{fundamental facets}](#std-fundamental_facets) = [as specified in the appropriate sub-section of [Other Built-in Datatypes (§3.4)](#ordinary-built-ins)]
  - [{context}](#std-context) = ***absent***
  - [{item type definition}](#std-item_type_definition) = if [{variety}](#std-variety) is ***atomic***, then ***absent***, otherwise as specified in the appropriate sub-section of [Other Built-in Datatypes (§3.4)](#ordinary-built-ins)]
  - [{member type definitions}](#std-member_type_definitions) = ***absent***
  - [{annotations}](#std-annotations) = As shown in the XML representations of the ordinary built-in datatypes in [Illustrative XML representations for the built-in ordinary type definitions (§C.2)](#drvd.nxsd)

### <a id="rf-fund-facets"></a>4.2 Fundamental Facets

        4.2.1 [ordered](#rf-ordered)
             4.2.1.1 [The ordered Schema Component](#dc-ordered)
         4.2.2 [bounded](#rf-bounded)
             4.2.2.1 [The bounded Schema Component](#dc-bounded)
         4.2.3 [cardinality](#rf-cardinality)
             4.2.3.1 [The cardinality Schema Component](#dc-cardinality)
         4.2.4 [numeric](#rf-numeric)
             4.2.4.1 [The numeric Schema Component](#dc-numeric)

 <a id="ff"></a> <a id="dt-fundamental-facet">[Definition:]</a> Each **fundamental facet** is a schema component that provides a limited piece of information about some aspect of each datatype.  All [fundamental facet](#dt-fundamental-facet) components are defined in this section.  For example, [cardinality](#ff-c) is a [fundamental facet](#dt-fundamental-facet).  Most [fundamental facets](#dt-fundamental-facet) are given a value fixed with each primitive datatype's definition, and this value is not changed by subsequent [derivations](#dt-derived) (even when it would perhaps be reasonable to expect an application to give a more accurate value based on the constraining facets used to define the [derivation](#dt-derived)).  The [cardinality](#ff-c) and [bounded](#ff-b) facets are exceptions to this rule; their values may change as a result of certain [derivations](#dt-derived).

**Note:** Schema components are identified by kind.  "Fundamental" is not a kind of component.  Each kind of [fundamental facet](#dt-fundamental-facet) ("ordered", "bounded", etc.) is a separate kind of schema component.

A [fundamental facet](#dt-fundamental-facet) can occur only in the [{fundamental facets}](#std-fundamental_facets) of a [Simple Type Definition](#std), and this is the only place where [fundamental facet](#dt-fundamental-facet) components occur.    Each kind of [fundamental facet](#dt-fundamental-facet) component occurs (once) in each [Simple Type Definition](#std)'s [{fundamental facets}](#std-fundamental_facets) set.

**Note:** The value of any [fundamental facet](#dt-fundamental-facet) component can always be calculated from other properties of its [owner](#dt-owner).  Fundamental facets are not required for schema processing, but some applications use them.

#### <a id="rf-ordered"></a>4.2.1 ordered

For some datatypes, this document specifies an order relation for their value spaces (see [Order (§2.2.3)](#order)); the *ordered* facet reflects this. It takes the values ***total***, ***partial***, and ***false***, with the meanings described below. For the [primitive](#dt-primitive) datatypes, the value of the *ordered* facet is specified in [Fundamental Facets (§F.1)](#app-fundamental-facets). For [ordinary](#dt-ordinary) datatypes, the value is inherited without change from the [base type](#dt-basetype). For a [list](#dt-list), the value is always ***false***; for a [union](#dt-union), the value is computed as described below.

A ***false*** value means no order is prescribed; a ***total*** value assures that the prescribed order is a total order; a ***partial*** value means that the prescribed order is a partial order, but not (for the primitive type in question) a total order.

**Note:** The value ***false*** in the *ordered* facet does not mean no partial or total ordering *exists* for the value space, only that none is specified by this document for use in checking upper and lower bounds. Mathematically, any set of values possesses at least one trivial partial ordering, in which every value pair that is not equal is incomparable.**Note:** When new datatypes are derived from datatypes with partial orders, the constraints imposed can sometimes result in a value space for which the ordering is total, or trivial. The value of the [ordered](#ff-o) facet is not, however, changed to reflect this. The value ***partial*** should therefore be interpreted with appropriate caution.

 <a id="dt-ordered">[Definition:]</a>A [value space](#dt-value-space), and hence a datatype, is said to be **ordered** if some members of the [value space](#dt-value-space) are drawn from a [primitive](#dt-primitive) datatype for which the table in [Fundamental Facets (§F.1)](#app-fundamental-facets) specifies the value ***total*** or ***partial*** for the *ordered* facet.

**Note:** Some of the "real-world" datatypes which are the basis for those defined herein are ordered in some applications, even though no order is prescribed for schema-processing purposes.  For example, [boolean](#boolean) is sometimes ordered, and [string](#string) and [list](#dt-list) datatypes [constructed](#dt-constructed) from ordered [atomic](#dt-atomic) datatypes are sometimes given "lexical" orderings.  They are *not* ordered for schema-processing purposes.

##### <a id="dc-ordered"></a>4.2.1.1 The ordered Schema Component

---

---

  Schema Component: <a id="ff-o">ordered</a>, a kind of [Fundamental Facet](#ff)   <a id="ff-o-value"></a>{value}  One of {false, partial, total}. Required.

---

---

 [{value}](#ff-o-value) depends on the [owner's](#dt-owner) [{variety}](#std-variety), [{facets}](#std-facets), and [{member type definitions}](#std-member_type_definitions). The appropriate **case** among the following **must** be true:

- 1<a id="x04042a"></a>**If **the [owner's](#dt-owner) [{variety}](#std-variety) is ***atomic***, **then ** the appropriate **case** among the following **must** be true:

- 1.1<a id="x040428b"></a>**If **the [owner](#dt-owner) is [primitive](#dt-primitive), **then **[{value}](#ff-o-value) is as specified in the table in [Fundamental Facets (§F.1)](#app-fundamental-facets).
- 1.2 **otherwise **[{value}](#ff-o-value) is the [owner's](#dt-owner) [{base type definition}](#std-base_type_definition)'s [ordered](#ff-o) [{value}](#ff-o-value).

- 2 **If **the [owner's](#dt-owner) [{variety}](#std-variety) is ***list***, **then **[{value}](#ff-o-value) is ***false***.
- 3 **otherwise **the [owner's](#dt-owner) [{variety}](#std-variety) is ***union***; the appropriate **case** among the following **must** be true:

- 3.1<a id="x040428"></a>**If **every [basic member](#dt-basicmember) of the [owner](#dt-owner) has [{variety}](#std-variety) atomic and has the same [{primitive type definition}](#std-primitive_type_definition), **then **[{value}](#ff-o-value) is the same as the [ordered](#ff-o) component's [{value}](#ff-o-value) in that primitive type definition's [{fundamental facets}](#std-fundamental_facets).
- 3.2 **If **each member of the [owner's](#dt-owner) [{member type definitions}](#std-member_type_definitions) has an [ordered](#ff-o) component in its [{fundamental facets}](#std-fundamental_facets) whose [{value}](#ff-o-value) is ***false***, **then **[{value}](#ff-o-value) is ***false***.
- 3.3 **otherwise **[{value}](#ff-o-value) is ***partial***.

#### <a id="rf-bounded"></a>4.2.2 bounded

Some ordered datatypes have the property that there is one value greater than or equal to every other value, and another that is less than or equal to every other value.  (In the case of [ordinary](#dt-ordinary) datatypes, these two values are not necessarily in the value space of the derived datatype, but they will always be in the value space of the primitive datatype from which they have been derived.) The *bounded* facet value is [boolean](#boolean) and is generally ***true*** for such *bounded* datatypes.  However, it will remain ***false*** when the mechanism for imposing such a bound is difficult to detect, as, for example, when the boundedness occurs because of derivation using a [pattern](#f-p) component.

##### <a id="dc-bounded"></a>4.2.2.1 The bounded Schema Component

---

---

  Schema Component: <a id="ff-b">bounded</a>, a kind of [Fundamental Facet](#ff)   <a id="ff-b-value"></a>{value}  An xs:boolean value. Required.

---

---

[{value}](#ff-b-value) depends on the [owner's](#dt-owner) [{variety}](#std-variety), [{facets}](#std-facets) and [{member type definitions}](#std-member_type_definitions).

When the [owner](#dt-owner) is [primitive](#dt-primitive), [{value}](#ff-b-value) is as specified in the table in [Fundamental Facets (§F.1)](#app-fundamental-facets).  Otherwise, when the [owner's](#dt-owner) [{variety}](#std-variety) is ***atomic***, if one of [minInclusive](#f-mii) or [minExclusive](#f-mie) and one of [maxInclusive](#f-mai) or [maxExclusive](#f-mae) are members of the [owner's](#dt-owner) [{facets}](#std-facets) set, then [{value}](#ff-b-value) is ***true***; otherwise [{value}](#ff-b-value) is ***false***.

When the [owner's](#dt-owner) [{variety}](#std-variety) is ***list***, [{value}](#ff-b-value) is ***false***.

 When the [owner's](#dt-owner) [{variety}](#std-variety) is ***union***, if [{value}](#ff-b-value) is ***true*** for every member of the [owner's](#dt-owner) [{member type definitions}](#std-member_type_definitions) set and all of the [owner's](#dt-owner) [basic members](#dt-basicmember) have the same [{primitive type definition}](#std-primitive_type_definition), then [{value}](#ff-b-value) is ***true***; otherwise [{value}](#ff-b-value) is ***false***.

#### <a id="rf-cardinality"></a>4.2.3 cardinality

Every value space has a specific number of members.  This number can be characterized as *finite* or *infinite*.  (Currently there are no datatypes with infinite value spaces larger than *countable*.)  The *cardinality* facet value is either ***finite*** or ***countably infinite*** and is generally ***finite*** for datatypes with finite value spaces.  However, it will remain ***countably infinite*** when the mechanism for causing finiteness is difficult to detect, as, for example, when finiteness occurs because of a derivation using a [pattern](#f-p) component.

##### <a id="dc-cardinality"></a>4.2.3.1 The cardinality Schema Component

---

---

  Schema Component: <a id="ff-c">cardinality</a>, a kind of [Fundamental Facet](#ff)   <a id="ff-c-value"></a>{value}  One of {finite, countably infinite}. Required.

---

---

 [{value}](#ff-c-value) depends on the [owner's](#dt-owner) [{variety}](#std-variety), [{facets}](#std-facets), and [{member type definitions}](#std-member_type_definitions).

When the [owner](#dt-owner) is [primitive](#dt-primitive), [{value}](#ff-c-value) is as specified in the table in [Fundamental Facets (§F.1)](#app-fundamental-facets).  Otherwise, when the [owner's](#dt-owner) [{variety}](#std-variety) is ***atomic***, [{value}](#ff-c-value) is ***countably infinite*** unless **any** of the following conditions are true, in which case [{value}](#ff-c-value) is ***finite***:

1. the [owner's](#dt-owner) [{base type definition}](#std-base_type_definition)'s [cardinality](#ff-c) [{value}](#ff-c-value) is ***finite***,
2. at least one of [length](#f-l), [maxLength](#f-mal), or [totalDigits](#f-td) is a member of the [owner's](#dt-owner) [{facets}](#std-facets) set,
3. **all** of the following are true:

  1. one of [minInclusive](#f-mii) or [minExclusive](#f-mie) is a member of the [owner's](#dt-owner) [{facets}](#std-facets) set
  2. one of [maxInclusive](#f-mai) or [maxExclusive](#f-mae) is a member of the [owner's](#dt-owner) [{facets}](#std-facets) set
  3. **either** of the following are true:

    1. [fractionDigits](#f-fd) is a member of the [owner's](#dt-owner) [{facets}](#std-facets) set
    2. [{primitive type definition}](#std-primitive_type_definition) is one of [date](#date), [gYearMonth](#gYearMonth), [gYear](#gYear), [gMonthDay](#gMonthDay), [gDay](#gDay) or [gMonth](#gMonth)

 When the [owner's](#dt-owner) [{variety}](#std-variety) is ***list***, if [length](#f-l) or both [minLength](#f-mil) and [maxLength](#f-mal) are members of the [owner's](#dt-owner) [{facets}](#std-facets) set and the [owner's](#dt-owner) [{item type definition}](#std-item_type_definition)'s [cardinality](#ff-c) [{value}](#ff-c-value) is ***finite*** then [{value}](#ff-c-value) is ***finite***; otherwise [{value}](#ff-c-value) is ***countably infinite***.

 When the [owner's](#dt-owner) [{variety}](#std-variety) is ***union***, if [cardinality](#ff-c)'s [{value}](#ff-c-value) is *finite* for every member of the [owner's](#dt-owner) [{member type definitions}](#std-member_type_definitions) set then [{value}](#ff-c-value) is ***finite***, otherwise [{value}](#ff-c-value) is ***countably infinite***.

#### <a id="rf-numeric"></a>4.2.4 numeric

Some value spaces are made up of things that are conceptually *numeric*, others are not. The *numeric* facet value indicates which are considered numeric.

##### <a id="dc-numeric"></a>4.2.4.1 The numeric Schema Component

---

---

  Schema Component: <a id="ff-n">numeric</a>, a kind of [Fundamental Facet](#ff)   <a id="ff-n-value"></a>{value}  An xs:boolean value. Required.

---

---

 [{value}](#ff-n-value) depends on the [owner's](#dt-owner) [{variety}](#std-variety), [{facets}](#std-facets), [{base type definition}](#std-base_type_definition) and [{member type definitions}](#std-member_type_definitions).

 When the [owner](#dt-owner) is [primitive](#dt-primitive), [{value}](#ff-n-value) is as specified in the table in [Fundamental Facets (§F.1)](#app-fundamental-facets).  Otherwise, when the [owner's](#dt-owner) [{variety}](#std-variety) is ***atomic***, [{value}](#ff-n-value) is inherited from the [owner's](#dt-owner) [{base type definition}](#std-base_type_definition)'s [numeric](#ff-n)[{value}](#ff-n-value).

 When the [owner's](#dt-owner) [{variety}](#std-variety) is ***list***, [{value}](#ff-n-value) is ***false***.

 When the [owner's](#dt-owner) [{variety}](#std-variety) is ***union***, if [numeric](#ff-n)'s [{value}](#ff-n-value) is ***true*** for every member of the [owner's](#dt-owner) [{member type definitions}](#std-member_type_definitions) set then [{value}](#ff-n-value) is ***true***, otherwise [{value}](#ff-n-value) is ***false***.

### <a id="rf-facets"></a>4.3 Constraining Facets

        4.3.1 [length](#rf-length)
             4.3.1.1 [The length Schema Component](#dc-length)
             4.3.1.2 [XML Representation of length Schema Components](#xr-length)
             4.3.1.3 [length Validation Rules](#length-validation-rules)
             4.3.1.4 [Constraints on length Schema Components](#length-coss)
         4.3.2 [minLength](#rf-minLength)
             4.3.2.1 [The minLength Schema Component](#dc-minLength)
             4.3.2.2 [XML Representation of minLength Schema Component](#xr-minLength)
             4.3.2.3 [minLength Validation Rules](#minLength-validation-rules)
             4.3.2.4 [Constraints on minLength Schema Components](#minLength-coss)
         4.3.3 [maxLength](#rf-maxLength)
             4.3.3.1 [The maxLength Schema Component](#dc-maxLength)
             4.3.3.2 [XML Representation of maxLength Schema Components](#xr-maxLength)
             4.3.3.3 [maxLength Validation Rules](#maxLength-validation-rules)
             4.3.3.4 [Constraints on maxLength Schema Components](#maxLength-coss)
         4.3.4 [pattern](#rf-pattern)
             4.3.4.1 [The pattern Schema Component](#dc-pattern)
             4.3.4.2 [XML Representation of pattern Schema Components](#xr-pattern)
             4.3.4.3 [Constraints on XML Representation of pattern](#pattern-rep-constr)
             4.3.4.4 [pattern Validation Rules](#pattern-validation-rules)
             4.3.4.5 [Constraints on pattern Schema Components](#pattern-constraints)
         4.3.5 [enumeration](#rf-enumeration)
             4.3.5.1 [The enumeration Schema Component](#dc-enumeration)
             4.3.5.2 [XML Representation of enumeration Schema Components](#xr-enumeration)
             4.3.5.3 [Constraints on XML Representation of enumeration](#enumeration-rep-constr)
             4.3.5.4 [enumeration Validation Rules](#enumeration-validation-rules)
             4.3.5.5 [Constraints on enumeration Schema Components](#enumeration-coss)
         4.3.6 [whiteSpace](#rf-whiteSpace)
             4.3.6.1 [The whiteSpace Schema Component](#dc-whiteSpace)
             4.3.6.2 [XML Representation of whiteSpace Schema Components](#xr-whiteSpace)
             4.3.6.3 [whiteSpace Validation Rules](#whiteSpace-validation-rules)
             4.3.6.4 [Constraints on whiteSpace Schema Components](#whiteSpace-coss)
         4.3.7 [maxInclusive](#rf-maxInclusive)
             4.3.7.1 [The maxInclusive Schema Component](#dc-maxInclusive)
             4.3.7.2 [XML Representation of maxInclusive Schema Components](#xr-maxInclusive)
             4.3.7.3 [maxInclusive Validation Rules](#maxInclusive-validation-rules)
             4.3.7.4 [Constraints on maxInclusive Schema Components](#maxInclusive-coss)
         4.3.8 [maxExclusive](#rf-maxExclusive)
             4.3.8.1 [The maxExclusive Schema Component](#dc-maxExclusive)
             4.3.8.2 [XML Representation of maxExclusive Schema Components](#xr-maxExclusive)
             4.3.8.3 [maxExclusive Validation Rules](#maxExclusive-validation-rules)
             4.3.8.4 [Constraints on maxExclusive Schema Components](#maxExclusive-coss)
         4.3.9 [minExclusive](#rf-minExclusive)
             4.3.9.1 [The minExclusive Schema Component](#dc-minExclusive)
             4.3.9.2 [XML Representation of minExclusive Schema Components](#xr-minExclusive)
             4.3.9.3 [minExclusive Validation Rules](#minExclusive-validation-rules)
             4.3.9.4 [Constraints on minExclusive Schema Components](#minExclusive-coss)
         4.3.10 [minInclusive](#rf-minInclusive)
             4.3.10.1 [The minInclusive Schema Component](#dc-minInclusive)
             4.3.10.2 [XML Representation of minInclusive Schema Components](#xr-minInclusive)
             4.3.10.3 [minInclusive Validation Rules](#minInclusive-validation-rules)
             4.3.10.4 [Constraints on minInclusive Schema Components](#minInclusive-coss)
         4.3.11 [totalDigits](#rf-totalDigits)
             4.3.11.1 [The totalDigits Schema Component](#dc-totalDigits)
             4.3.11.2 [XML Representation of totalDigits Schema Components](#xr-totalDigits)
             4.3.11.3 [totalDigits Validation Rules](#totalDigits-validation-rules)
             4.3.11.4 [Constraints on totalDigits Schema Components](#totalDigits-coss)
         4.3.12 [fractionDigits](#rf-fractionDigits)
             4.3.12.1 [The fractionDigits Schema Component](#dc-fractionDigits)
             4.3.12.2 [XML Representation of fractionDigits Schema Components](#xr-fractionDigits)
             4.3.12.3 [fractionDigits Validation Rules](#fractionDigits-validation-rules)
             4.3.12.4 [Constraints on fractionDigits Schema Components](#fractionDigits-coss)
         4.3.13 [Assertions](#rf-assertions)
             4.3.13.1 [The assertions Schema Component](#dc-assertions)
             4.3.13.2 [XML Representation of assertions Schema Components](#xr-assertions)
             4.3.13.3 [Assertions Validation Rules](#assertions-validation-rules)
             4.3.13.4 [Constraints on assertions Schema Components](#assertions-coss)
         4.3.14 [explicitTimezone](#rf-explicitTimezone)
             4.3.14.1 [The explicitTimezone Schema Component](#dc-explicitTimezone)
             4.3.14.2 [XML Representation of explicitTimezone Schema Components](#xr-timezone)
             4.3.14.3 [explicitTimezone Validation Rules](#timezone-vr)
             4.3.14.4 [Constraints on explicitTimezone Schema Components](#timezone-coss)

 <a id="f"></a> <a id="dt-constraining-facet">[Definition:]</a>**Constraining facets** are schema components whose values may be set or changed during [derivation](#dt-derived) (subject to facet-specific controls) to control various aspects of the derived datatype.  All [constraining facet](#dt-constraining-facet) components defined by this specification are defined in this section.  For example, [whiteSpace](#f-w) is a [constraining facet](#dt-constraining-facet).  [Constraining Facets](#dt-constraining-facet) are given a value as part of the [derivation](#dt-derived) when an [ordinary](#dt-ordinary) datatype is defined by [restricting](#dt-fb-restriction) a [primitive](#dt-primitive) or [ordinary](#dt-ordinary) datatype; a few [constraining facets](#dt-constraining-facet) have default values that are also provided for [primitive](#dt-primitive) datatypes.

**Note:** Schema components are identified by kind.  "Constraining" is not a kind of component.  Each kind of [constraining facet](#dt-constraining-facet) ("whiteSpace", "length", etc.) is a separate kind of schema component.  This specification distinguishes three kinds of constraining facets:

- <a id="dt-pre-lexical">[Definition:]</a>A constraining facet which is used to normalize an initial [literal](#dt-literal) before checking to see whether the resulting character sequence is a member of a datatype's [lexical space](#dt-lexical-space) is a **pre-lexical** facet.This specification defines just one [pre-lexical](#dt-pre-lexical) facet: [whiteSpace](#f-w).
- <a id="dt-lexical">[Definition:]</a>A constraining facet which directly restricts the [lexical space](#dt-lexical-space) of a datatype is a **lexical** facet.This specification defines just one [lexical](#dt-lexical) facet: [pattern](#f-p).**Note:** As specified normatively elsewhere, [lexical](#dt-lexical) facets can have an indirect effect on the [value space](#dt-value-space): if every lexical representation of a value is removed from the [lexical space](#dt-lexical-space), the value itself is removed from the [value space](#dt-value-space).
- <a id="dt-value-based">[Definition:]</a>A constraining facet which directly restricts the [value space](#dt-value-space) of a datatype is a **value-based** facet.Most of the constraining facets defined by this specification are [value-based](#dt-value-based) facets.**Note:** As specified normatively elsewhere, [value-based](#dt-value-based) facets can have an indirect effect on the [lexical space](#dt-lexical-space): if a value is removed from the [value space](#dt-value-space), its lexical representations are removed from the [lexical space](#dt-lexical-space).

 Conforming processors **must** support all the facets defined in this section. It is [implementation-defined](#key-impl-def) whether a processor supports other constraining facets. <a id="dt-unknown-f">[Definition:]</a>An [constraining facet](#dt-constraining-facet) which is not supported by the processor in use is **unknown**.

**Note:** A reference to an [unknown](#dt-unknown-f) facet might be a reference to an [implementation-defined](#key-impl-def) facet supported by some other processor, or might be the result of a typographic error, or might have some other explanation.

The descriptions of individual facets given below include both constraints on [Simple Type Definition](#std) components and rules for checking the datatype validity of a given literal against a given datatype. The validation rules typically depend upon having a full knowledge of the datatype; full knowledge of the datatype, in turn, depends on having a fully instantiated [Simple Type Definition](#std). A full instantiation of the [Simple Type Definition](#std), and the checking of the component constraints, require knowledge of the [base type](#dt-basetype). It follows that if a datatype's [base type](#dt-basetype) is [unknown](#dt-unknown-dt), the [Simple Type Definition](#std) defining the datatype will be incompletely instantiated, and the datatype itself will be [unknown](#dt-unknown-dt). Similarly, any datatype defined using an [unknown](#dt-unknown-f) [constraining facet](#dt-constraining-facet) will be [unknown](#dt-unknown-dt). It is not possible to perform datatype validation as defined here using [unknown](#dt-unknown-dt) datatypes.

**Note:** The preceding paragraph does not forbid implementations from attempting to make use of such partial information as they have about [unknown](#dt-unknown-dt) datatypes. But the exploitation of such partial knowledge is not datatype validity checking as defined here and is to be distinguished from it in the implementation's documentation and interface.

#### <a id="rf-length"></a>4.3.1 length

 <a id="dt-length">[Definition:]</a> **length** is the number of *units of length*, where *units of length* varies depending on the type that is being [derived](#dt-derived) from. The value of **length** [**must**](#dt-must) be a [nonNegativeInteger](#nonNegativeInteger).

 For [string](#string) and datatypes [derived](#dt-derived) from [string](#string), **length** is measured in units of [character](https://www.w3.org/TR/xml11/#dt-character)s as defined in [[XML]](#XML). For [anyURI](#anyURI), **length** is measured in units of characters (as for [string](#string)). For [hexBinary](#hexBinary) and [base64Binary](#base64Binary) and datatypes [derived](#dt-derived) from them, **length** is measured in octets (8 bits) of binary data. For datatypes [constructed](#dt-constructed) by [list](#dt-list), **length** is measured in number of list items.

**Note:** For [string](#string) and datatypes [derived](#dt-derived) from [string](#string), **length** will not always coincide with "string length" as perceived by some users or with the number of storage units in some digital representation.  Therefore, care should be taken when specifying a value for **length** and in attempting to infer storage requirements from a given value for **length**.

 [length](#dt-length) provides for:

-  Constraining a [value space](#dt-value-space) to values with a specific number of *units of length*, where *units of length* varies depending on [{base type definition}](#std-base_type_definition).

*Example*

  The following is the definition of a [user-defined](#dt-user-defined) datatype to represent product codes which must be exactly 8 characters in length.  By fixing the value of the **length** facet we ensure that types derived from productCode can change or set the values of other facets, such as **pattern**, but cannot change the length.

```
<simpleType name='productCode'>
   <restriction base='string'>
     <length value='8' fixed='true'/>
   </restriction>
</simpleType>
```

##### <a id="dc-length"></a>4.3.1.1 The length Schema Component

---

---

  Schema Component: <a id="f-l">length</a>, a kind of [Constraining Facet](#f)   <a id="f-l-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-l-value"></a>{value}  An xs:nonNegativeInteger value. Required.  <a id="f-l-fixed"></a>{fixed}  An xs:boolean value. Required.

---

---

 If [{fixed}](#f-l-fixed) is *true*, then types for which the current type is the [{base type definition}](#std-base_type_definition) cannot specify a value for [length](#f-l) other than [{value}](#f-l-value).

**Note:** The [{fixed}](#f-l-fixed) property is defined for parallelism with other facets and for compatibility with version 1.0 of this specification. But it is a consequence of [length valid restriction (§4.3.1.4)](#length-valid-restriction) that the value of the [length](#f-l) facet cannot be changed, regardless of whether [{fixed}](#f-l-fixed) is *true* or *false*.

##### <a id="xr-length"></a>4.3.1.2 XML Representation of length Schema Components

 The XML representation for a [length](#f-l) schema component is a [<length>](#element-length) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `length` Element Information Item

<a id="element-length"></a>

```xml
<length
  fixed = boolean : false
  id = ID
  value = nonNegativeInteger
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</length>
```

**[length](#dc-fractionDigits) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-l-value): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- [{fixed}](#f-l-fixed): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `fixed` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise ***false***
- [{annotations}](#f-l-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-one) of the [<length>](#element-length) element, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

##### <a id="length-validation-rules"></a>4.3.1.3 length Validation Rules

<a id="cvc-length-valid"></a>**Validation Rule: Length Valid**

> A value in a [value space](#dt-value-space) is facet-valid with respect to [length](#dt-length) if and only if:
>
> - 1  if the [{variety}](#std-variety) is [atomic](#dt-atomic) then
>
> - 1.1  if [{primitive type definition}](#std-primitive_type_definition) is [string](#string) or [anyURI](#anyURI), then the length of the value, as measured in [character](https://www.w3.org/TR/xml11/#dt-character)s [**must**](#dt-must) be equal to [{value}](#f-l-value);
> - 1.2  if [{primitive type definition}](#std-primitive_type_definition) is [hexBinary](#hexBinary) or [base64Binary](#base64Binary), then the length of the value, as measured in octets of the binary data, [**must**](#dt-must) be equal to [{value}](#f-l-value);
> - 1.3  if [{primitive type definition}](#std-primitive_type_definition) is [QName](#QName) or [NOTATION](#NOTATION), then any [{value}](#f-l-value) is facet-valid.
>
> - 2  if the [{variety}](#std-variety) is [list](#dt-list), then the length of the value, as measured in list items, [**must**](#dt-must) be equal to [{value}](#f-l-value)

 The use of [length](#dt-length) on [QName](#QName), [NOTATION](#NOTATION), and datatypes [derived](#dt-derived) from them is deprecated.  Future versions of this specification may remove this facet for these datatypes.

##### <a id="length-coss"></a>4.3.1.4 Constraints on length Schema Components

<a id="length-minLength-maxLength"></a>**Schema Component Constraint: length and minLength or maxLength**

> If [length](#f-l) is a member of [{facets}](#std-facets) then
>
> - 1 It is an error for [minLength](#f-mil) to be a member of [{facets}](#std-facets) unless
>
> - 1.1 the [{value}](#f-mil-value) of [minLength](#f-mil) <= the [{value}](#f-l-value) of [length](#f-l) and
> - 1.2 there is some type definition from which this one is derived by one or more [restriction](#dt-restriction) steps in which [minLength](#f-mil) has the same [{value}](#f-mil-value) and [length](#f-l) is not specified.
>
> - 2 It is an error for [maxLength](#f-mal) to be a member of [{facets}](#std-facets) unless
>
> - 2.1 the [{value}](#f-l-value) of [length](#f-l) <= the [{value}](#f-mal-value) of [maxLength](#f-mal) and
> - 2.2 there is some type definition from which this one is derived by one or more restriction steps in which [maxLength](#f-mal) has the same [{value}](#f-mal-value) and [length](#f-l) is not specified.

<a id="length-valid-restriction"></a>**Schema Component Constraint: length valid restriction**

> It is an [error](#dt-error) if [length](#f-l) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-l-value) is not equal to the [{value}](#f-l-value) of the parent [length](#f-l).

#### <a id="rf-minLength"></a>4.3.2 minLength

 <a id="dt-minLength">[Definition:]</a> **minLength** is the minimum number of *units of length*, where *units of length* varies depending on the type that is being [derived](#dt-derived) from. The value of **minLength**  [**must**](#dt-must) be a [nonNegativeInteger](#nonNegativeInteger).

 For [string](#string) and datatypes [derived](#dt-derived) from [string](#string), **minLength** is measured in units of [character](https://www.w3.org/TR/xml11/#dt-character)s as defined in [[XML]](#XML). For [hexBinary](#hexBinary) and [base64Binary](#base64Binary) and datatypes [derived](#dt-derived) from them, **minLength** is measured in octets (8 bits) of binary data. For datatypes [constructed](#dt-constructed) by [list](#dt-list), **minLength** is measured in number of list items.

**Note:** For [string](#string) and datatypes [derived](#dt-derived) from [string](#string), **minLength** will not always coincide with "string length" as perceived by some users or with the number of storage units in some digital representation. Therefore, care should be taken when specifying a value for **minLength** and in attempting to infer storage requirements from a given value for **minLength**.

 [minLength](#dt-minLength) provides for:

-  Constraining a [value space](#dt-value-space) to values with at least a specific number of *units of length*, where *units of length* varies depending on [{base type definition}](#std-base_type_definition).

*Example*

  The following is the definition of a [user-defined](#dt-user-defined) datatype which requires strings to have at least one character (i.e., the empty string is not in the [value space](#dt-value-space) of this datatype).

```
<simpleType name='non-empty-string'>
  <restriction base='string'>
    <minLength value='1'/>
  </restriction>
</simpleType>
```

##### <a id="dc-minLength"></a>4.3.2.1 The minLength Schema Component

---

---

  Schema Component: <a id="f-mil">minLength</a>, a kind of [Constraining Facet](#f)   <a id="f-mil-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-mil-value"></a>{value}  An xs:nonNegativeInteger value. Required.  <a id="f-mil-fixed"></a>{fixed}  An xs:boolean value. Required.

---

---

 If [{fixed}](#f-mil-fixed) is *true*, then types for which the current type is the [{base type definition}](#std-base_type_definition) cannot specify a value for [minLength](#f-mil) other than [{value}](#f-mil-value).

##### <a id="xr-minLength"></a>4.3.2.2 XML Representation of minLength Schema Component

 The XML representation for a [minLength](#f-mil) schema component is a [<minLength>](#element-minLength) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `minLength` Element Information Item

<a id="element-minLength"></a>

```xml
<minLength
  fixed = boolean : false
  id = ID
  value = nonNegativeInteger
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</minLength>
```

**[minLength](#dc-fractionDigits) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-mil-value): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- [{fixed}](#f-mil-fixed): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `fixed` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise ***false***
- [{annotations}](#f-mil-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-one) of the [<minLength>](#element-minLength) element, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

##### <a id="minLength-validation-rules"></a>4.3.2.3 minLength Validation Rules

<a id="cvc-minLength-valid"></a>**Validation Rule: minLength Valid**

> A value in a [value space](#dt-value-space) is facet-valid with respect to [minLength](#dt-minLength), determined as follows:
>
> - 1  if the [{variety}](#std-variety) is [atomic](#dt-atomic) then
>
> - 1.1  if [{primitive type definition}](#std-primitive_type_definition) is [string](#string) or [anyURI](#anyURI), then the length of the value, as measured in[character](https://www.w3.org/TR/xml11/#dt-character)s [**must**](#dt-must) be greater than or equal to [{value}](#f-mil-value);
> - 1.2  if [{primitive type definition}](#std-primitive_type_definition) is [hexBinary](#hexBinary) or [base64Binary](#base64Binary), then the length of the value, as measured in octets of the binary data, [**must**](#dt-must) be greater than or equal to [{value}](#f-mil-value);
> - 1.3  if [{primitive type definition}](#std-primitive_type_definition) is [QName](#QName) or [NOTATION](#NOTATION), then any [{value}](#f-mil-value) is facet-valid.
>
> - 2  if the [{variety}](#std-variety) is [list](#dt-list), then the length of the value, as measured in list items, [**must**](#dt-must) be greater than or equal to [{value}](#f-mil-value)

 The use of [minLength](#dt-minLength) on [QName](#QName), [NOTATION](#NOTATION), and datatypes [derived](#dt-derived) from them is deprecated.  Future versions of this specification may remove this facet for these datatypes.

##### <a id="minLength-coss"></a>4.3.2.4 Constraints on minLength Schema Components

<a id="minLength-less-than-equal-to-maxLength"></a>**Schema Component Constraint: minLength <= maxLength**

> If both [minLength](#f-mil) and [maxLength](#f-mal) are members of [{facets}](#std-facets), then the [{value}](#f-mil-value) of [minLength](#f-mil)  [**must**](#dt-must) be less than or equal to the [{value}](#f-mal-value) of [maxLength](#f-mal).

<a id="minLength-valid-restriction"></a>**Schema Component Constraint: minLength valid restriction**

> It is an [error](#dt-error) if [minLength](#f-mil) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mil-value) is less than the [{value}](#f-mil-value) of the parent [minLength](#f-mil).

#### <a id="rf-maxLength"></a>4.3.3 maxLength

 <a id="dt-maxLength">[Definition:]</a> **maxLength** is the maximum number of *units of length*, where *units of length* varies depending on the type that is being [derived](#dt-derived) from. The value of **maxLength**  [**must**](#dt-must) be a [nonNegativeInteger](#nonNegativeInteger).

 For [string](#string) and datatypes [derived](#dt-derived) from [string](#string), **maxLength** is measured in units of [character](https://www.w3.org/TR/xml11/#dt-character)s as defined in [[XML]](#XML). For [hexBinary](#hexBinary) and [base64Binary](#base64Binary) and datatypes [derived](#dt-derived) from them, **maxLength** is measured in octets (8 bits) of binary data. For datatypes [constructed](#dt-constructed) by [list](#dt-list), **maxLength** is measured in number of list items.

**Note:** For [string](#string) and datatypes [derived](#dt-derived) from [string](#string), **maxLength** will not always coincide with "string length" as perceived by some users or with the number of storage units in some digital representation. Therefore, care should be taken when specifying a value for **maxLength** and in attempting to infer storage requirements from a given value for **maxLength**.

 [maxLength](#dt-maxLength) provides for:

-  Constraining a [value space](#dt-value-space) to values with at most a specific number of *units of length*, where *units of length* varies depending on [{base type definition}](#std-base_type_definition).

*Example*

  The following is the definition of a [user-defined](#dt-user-defined) datatype which might be used to accept form input with an upper limit to the number of characters that are acceptable.

```
<simpleType name='form-input'>
  <restriction base='string'>
    <maxLength value='50'/>
  </restriction>
</simpleType>
```

##### <a id="dc-maxLength"></a>4.3.3.1 The maxLength Schema Component

---

---

  Schema Component: <a id="f-mal">maxLength</a>, a kind of [Constraining Facet](#f)   <a id="f-mal-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-mal-value"></a>{value}  An xs:nonNegativeInteger value. Required.  <a id="f-mal-fixed"></a>{fixed}  An xs:boolean value. Required.

---

---

 If [{fixed}](#f-mal-fixed) is *true*, then types for which the current type is the [{base type definition}](#std-base_type_definition) cannot specify a value for [maxLength](#f-mal) other than [{value}](#f-mal-value).

##### <a id="xr-maxLength"></a>4.3.3.2 XML Representation of maxLength Schema Components

 The XML representation for a [maxLength](#f-mal) schema component is a [<maxLength>](#element-maxLength) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `maxLength` Element Information Item

<a id="element-maxLength"></a>

```xml
<maxLength
  fixed = boolean : false
  id = ID
  value = nonNegativeInteger
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</maxLength>
```

**[maxLength](#dc-fractionDigits) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-mal-value): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- [{fixed}](#f-mal-fixed): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `fixed` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise ***false***
- [{annotations}](#f-mal-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-one) of the [<maxLength>](#element-maxLength) element, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

##### <a id="maxLength-validation-rules"></a>4.3.3.3 maxLength Validation Rules

<a id="cvc-maxLength-valid"></a>**Validation Rule: maxLength Valid**

> A value in a [value space](#dt-value-space) is facet-valid with respect to [maxLength](#dt-maxLength), determined as follows:
>
> - 1  if the [{variety}](#std-variety) is [atomic](#dt-atomic) then
>
> - 1.1  if [{primitive type definition}](#std-primitive_type_definition) is [string](#string) or [anyURI](#anyURI), then the length of the value, as measured in [character](https://www.w3.org/TR/xml11/#dt-character)s [**must**](#dt-must) be less than or equal to [{value}](#f-mal-value);
> - 1.2  if [{primitive type definition}](#std-primitive_type_definition) is [hexBinary](#hexBinary) or [base64Binary](#base64Binary), then the length of the value, as measured in octets of the binary data, [**must**](#dt-must) be less than or equal to [{value}](#f-mal-value);
> - 1.3  if [{primitive type definition}](#std-primitive_type_definition) is [QName](#QName) or [NOTATION](#NOTATION), then any [{value}](#f-mal-value) is facet-valid.
>
> - 2  if the [{variety}](#std-variety) is [list](#dt-list), then the length of the value, as measured in list items, [**must**](#dt-must) be less than or equal to [{value}](#f-mal-value)

 The use of [maxLength](#dt-maxLength) on [QName](#QName), [NOTATION](#NOTATION), and datatypes [derived](#dt-derived) from them is deprecated.  Future versions of this specification may remove this facet for these datatypes.

##### <a id="maxLength-coss"></a>4.3.3.4 Constraints on maxLength Schema Components

<a id="maxLength-valid-restriction"></a>**Schema Component Constraint: maxLength valid restriction**

> It is an [error](#dt-error) if [maxLength](#f-mal) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mal-value) is greater than the [{value}](#f-mal-value) of the parent [maxLength](#f-mal).

#### <a id="rf-pattern"></a>4.3.4 pattern

<a id="dt-pattern">[Definition:]</a> **pattern** is a constraint on the [value space](#dt-value-space) of a datatype which is achieved by constraining the [lexical space](#dt-lexical-space) to [literals](#dt-literal) which match each member of a set of [regular expressions](#dt-regex).  The value of **pattern**  **must** be a set of [regular expressions](#dt-regex).

**Note:** An XML [<restriction>](#element-restriction) containing more than one [<pattern>](#element-pattern) element gives rise to a single [regular expression](#dt-regex) in the set; this [regular expression](#dt-regex) is an "or" of the [regular expressions](#dt-regex) that are the content of the [<pattern>](#element-pattern) elements.

 [pattern](#dt-pattern) provides for:

- Constraining a [value space](#dt-value-space) to values that are denoted by [literals](#dt-literal) which match each of a set of [regular expressions](#dt-regex).

*Example*

  The following is the definition of a [user-defined](#dt-user-defined) datatype which is a better representation of postal codes in the United States, by limiting strings to those which are matched by a specific [regular expression](#dt-regex).

```
<simpleType name='better-us-zipcode'>
  <restriction base='string'>
    <pattern value='[0-9]{5}(-[0-9]{4})?'/>
  </restriction>
</simpleType>
```

##### <a id="dc-pattern"></a>4.3.4.1 The pattern Schema Component

---

---

  Schema Component: <a id="f-p">pattern</a>, a kind of [Constraining Facet](#f)   <a id="f-p-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-p-value"></a>{value}

A non-empty set of [regular expressions](#dt-regex).

---

---

##### <a id="xr-pattern"></a>4.3.4.2 XML Representation of pattern Schema Components

 The XML representation for a [pattern](#f-p) schema component is one or more [<pattern>](#element-pattern) element information items. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `pattern` Element Information Item

<a id="element-pattern"></a>

```xml
<pattern
  id = ID
  value = string
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</pattern>
```

**[pattern](#dc-pattern) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-p-value): <a id="l-R">[Definition:]</a>Let **R** be a regular expression given by  the appropriate **case** among the following:

- 1 **If **there is only one [<pattern>](#element-pattern) among the [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of a [<restriction>](#element-restriction), **then **the [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of its `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- 2 **otherwise **the concatenation of the [actual values](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of all the [<pattern>](#element-pattern) [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element)'s `value` [[attributes]](https://www.w3.org/TR/xml-infoset/#infoitem.element), in order, separated by '`|`', so forming a single regular expression with multiple [branches](#dt-branch).
 The value is then given by the appropriate **case** among the following:

- 1 **If **the [{base type definition}](#std-base_type_definition) of the [owner](#dt-owner) has a [pattern](#f-p) facet among its [{facets}](#std-facets), **then **the union of that [pattern](#f-p) facet's [{value}](#f-p-value) and {[R](#l-R)}
- 2 **otherwise **just {[R](#l-R)}
- [{annotations}](#f-p-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-set) of the set containing all of the [<pattern>](#element-pattern) elements among the [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of the [<restriction>](#element-restriction) element information item, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

**Note:** The [{value}](#f-p-value) property will only have more than one member when [facet-based restriction](#dt-fb-restriction) involves a [pattern](#f-p) facet at more than one step in a type derivation. During validation, lexical forms will be checked against every member of the resulting [{value}](#f-p-value), effectively creating a conjunction of patterns. In summary, [pattern](#dt-pattern) facets specified on the *same* step in a type derivation are **OR**ed together, while [pattern](#dt-pattern) facets specified on *different* steps of a type derivation are **AND**ed together.  Thus, to impose two [pattern](#dt-pattern) constraints simultaneously, schema authors may either write a single [pattern](#dt-pattern) which expresses the intersection of the two [pattern](#dt-pattern)s they wish to impose, or define each [pattern](#dt-pattern) on a separate type derivation step.

##### <a id="pattern-rep-constr"></a>4.3.4.3 Constraints on XML Representation of pattern

<a id="src-pattern-value"></a>**Schema Representation Constraint: Pattern value**

> The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) must be a [regular expression](#dt-regex) as defined in [Regular Expressions (§G)](#regexs).

##### <a id="pattern-validation-rules"></a>4.3.4.4 pattern Validation Rules

<a id="cvc-pattern-valid"></a>**Validation Rule: pattern valid**

> A [literal](#dt-literal) in a [lexical space](#dt-lexical-space) is pattern-valid (or: facet-valid with respect to [pattern](#dt-pattern)) if and only if for each [regular expression](#dt-regex) in its [{value}](#f-p-value), the [literal](#dt-literal) is among the set of character sequences denoted by the [regular expression](#dt-regex).

**Note:** As noted in [Datatype (§2.1)](#datatype), certain uses of the [pattern](#dt-pattern) facet may eliminate from the lexical space the canonical forms of some values in the value space; this can be inconvenient for applications which write out the canonical form of a value and rely on being able to read it in again as a legal lexical form. This specification provides no recourse in such situations; applications are free to deal with it as they see fit. Caution is advised.

##### <a id="pattern-constraints"></a>4.3.4.5 Constraints on pattern Schema Components

<a id="cos-pattern-restriction"></a>**Schema Component Constraint: Valid restriction of pattern**

> It is an [error](#dt-error) if there is any member of the [{value}](#f-p-value) of the [pattern](#f-p) facet on the [{base type definition}](#std-base_type_definition) which is not also a member of the [{value}](#f-p-value).**Note:** For components constructed from XML representations in schema documents, the satisfaction of this constraint is a consequence of the XML mapping rules: any pattern imposed by a simple type definition *S* will always also be imposed by any type derived from *S* by [facet-based restriction](#dt-fb-restriction). This constraint ensures that components constructed by other means (so-called "born-binary" components) similarly preserve [pattern](#f-p) facets across [facet-based restriction](#dt-fb-restriction).

#### <a id="rf-enumeration"></a>4.3.5 enumeration

 <a id="dt-enumeration">[Definition:]</a> **enumeration** constrains the [value space](#dt-value-space) to a specified set of values.

 **enumeration** does not impose an order relation on the [value space](#dt-value-space) it creates; the value of the [ordered](#dt-ordered) property of the [derived](#dt-derived) datatype remains that of the datatype from which it is [derived](#dt-derived).

 [enumeration](#dt-enumeration) provides for:

-  Constraining a [value space](#dt-value-space) to a specified set of values.

*Example*

  The following example is a [Simple Type Definition](#std) for a [user-defined](#dt-user-defined) datatype which limits the values of dates to the three US holidays enumerated. This [Simple Type Definition](#std) would appear in a schema authored by an "end-user" and shows how to define a datatype by enumerating the values in its [value space](#dt-value-space).  The enumerated values must be type-valid [literals](#dt-literal) for the [base type](#dt-basetype).

```
<simpleType name='holidays'>
    <annotation>
        <documentation>some US holidays</documentation>
    </annotation>
    <restriction base='gMonthDay'>
      <enumeration value='--01-01'>
        <annotation>
            <documentation>New Year's day</documentation>
        </annotation>
      </enumeration>
      <enumeration value='--07-04'>
        <annotation>
            <documentation>4th of July</documentation>
        </annotation>
      </enumeration>
      <enumeration value='--12-25'>
        <annotation>
            <documentation>Christmas</documentation>
        </annotation>
      </enumeration>
    </restriction>
</simpleType>
```

##### <a id="dc-enumeration"></a>4.3.5.1 The enumeration Schema Component

---

---

  Schema Component: <a id="f-e">enumeration</a>, a kind of [Constraining Facet](#f)   <a id="f-e-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-e-value"></a>{value}

 A set of values from the [value space](#dt-value-space) of the [{base type definition}](#std-base_type_definition).

---

---

##### <a id="xr-enumeration"></a>4.3.5.2 XML Representation of enumeration Schema Components

 The XML representation for an [enumeration](#f-e) schema component is one or more [<enumeration>](#element-enumeration) element information items. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `enumeration` Element Information Item

<a id="element-enumeration"></a>

```xml
<enumeration
  id = ID
  value = anySimpleType
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</enumeration>
```

**[enumeration](#dc-enumeration) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-e-value): The appropriate **case** among the following:

- 1 **If **there is only one [<enumeration>](#element-enumeration) among the [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of a [<restriction>](#element-restriction), **then **a set with one member, the [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of its `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), interpreted as an instance of the [{base type definition}](#std-base_type_definition).
- 2 **otherwise **a set of the [actual values](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of all the [<enumeration>](#element-enumeration) [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element)'s `value` [[attributes]](https://www.w3.org/TR/xml-infoset/#infoitem.element), interpreted as instances of the [{base type definition}](#std-base_type_definition).
 **Note:** The `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) is declared as having type [`anySimpleType`](#dt-anySimpleType), but the [{value}](#f-e-value) property of the [enumeration](#f-e) facet **must** be a member of the [{base type definition}](#std-base_type_definition). So in mapping from the XML representation to the [enumeration](#f-e) component, the [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) is identified by using the [lexical mapping](#dt-lexical-mapping) of the [{base type definition}](#std-base_type_definition).
- [{annotations}](#f-e-annotations): A (possibly empty) sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components, one for each [<annotation>](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#element-annotation) among the [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of the [<enumeration>](#element-enumeration)s among the [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of a [<restriction>](#element-restriction), in order.

---

##### <a id="enumeration-rep-constr"></a>4.3.5.3 Constraints on XML Representation of enumeration

<a id="src-enumeration-value"></a>**Schema Representation Constraint: Enumeration value**

> The [normalized value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-nv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element) must be [Datatype Valid (§4.1.4)](#cvc-datatype-valid) with respect to the [{base type definition}](#std-base_type_definition) of the [Simple Type Definition](#std) corresponding to the nearest [<simpleType>](#element-simpleType) ancestor element.

##### <a id="enumeration-validation-rules"></a>4.3.5.4 enumeration Validation Rules

<a id="cvc-enumeration-valid"></a>**Validation Rule: enumeration valid**

> A value in a [value space](#dt-value-space) is facet-valid with respect to [enumeration](#dt-enumeration) if and only if the value is equal or identical to one of the values specified in [{value}](#f-e-value).

**Note:** As specified normatively elsewhere, for purposes of checking enumerations, no distinction is made between an atomic value *V* and a list of length one containing *V* as its only item.In this question, the behavior of this specification is thus the same as the behavior specified by [[XQuery 1.0 and XPath 2.0 Functions and Operators]](#F_O) and related specifications.

##### <a id="enumeration-coss"></a>4.3.5.5 Constraints on enumeration Schema Components

<a id="enumeration-valid-restriction"></a>**Schema Component Constraint: enumeration valid restriction**

> It is an [error](#dt-error) if any member of [{value}](#f-e-value) is not in the [value space](#dt-value-space) of [{base type definition}](#std-base_type_definition).

#### <a id="rf-whiteSpace"></a>4.3.6 whiteSpace

 <a id="dt-whiteSpace">[Definition:]</a> **whiteSpace** constrains the [value space](#dt-value-space) of types [derived](#dt-derived) from [string](#string) such that the various behaviors specified in [Attribute Value Normalization](https://www.w3.org/TR/xml11/#AVNormalize) in [[XML]](#XML) are realized.  The value of **whiteSpace** must be one of {preserve, replace, collapse}.

preserve No normalization is done, the value is not changed (this is the behavior required by [[XML]](#XML) for element content) replace All occurrences of #x9 (tab), #xA (line feed) and #xD (carriage return) are replaced with #x20 (space) collapse After the processing implied by **replace**, contiguous sequences of #x20's are collapsed to a single #x20, and any #x20 at the start or end of the string is then removed. **Note:** The notation #xA used here (and elsewhere in this specification) represents the Universal Character Set (UCS) code point `hexadecimal A` (line feed), which is denoted by U+000A.  This notation is to be distinguished from `&#xA;`, which is the XML [character reference](https://www.w3.org/TR/xml11/#NT-CharRef) to that same UCS code point.

 **whiteSpace** is applicable to all [atomic](#dt-atomic) and [list](#dt-list) datatypes.  For all [atomic](#dt-atomic) datatypes other than [string](#string) (and types [derived](#dt-derived) by [facet-based restriction](#dt-fb-restriction) from it) the value of **whiteSpace** is `collapse` and cannot be changed by a schema author; for [string](#string) the value of **whiteSpace** is `preserve`; for any type [derived](#dt-derived) by [facet-based restriction](#dt-fb-restriction) from [string](#string) the value of **whiteSpace** can be any of the three legal values (as long as the value is at least as restrictive as the value of the [base type](#dt-basetype); see [Constraints on whiteSpace Schema Components (§4.3.6.4)](#whiteSpace-coss)).  For all datatypes [constructed](#dt-constructed) by [list](#dt-list) the value of **whiteSpace** is `collapse` and cannot be changed by a schema author.  For all datatypes [constructed](#dt-constructed) by [union](#dt-union)  **whiteSpace** does not apply directly; however, the normalization behavior of [union](#dt-union) types is controlled by the value of **whiteSpace** on that one of the [basic members](#dt-basicmember) against which the [union](#dt-union) is successfully validated.

**Note:** For more information on **whiteSpace**, see the discussion on white space normalization in [Schema Component Details](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#components) in [[XSD 1.1 Part 1: Structures]](#structural-schemas).

 [whiteSpace](#dt-whiteSpace) provides for:

-  Constraining a [value space](#dt-value-space) according to the white space normalization rules.

*Example*

  The following example is the [Simple Type Definition](#std) for the [built-in](#dt-built-in) [token](#token) datatype.

```
<simpleType name='token'>
    <restriction base='normalizedString'>
      <whiteSpace value='collapse'/>
    </restriction>
</simpleType>
```

**Note:** The values "`replace`" and "`collapse`" may appear to provide a convenient way to "unwrap" text (i.e. undo the effects of pretty-printing and word-wrapping). In some cases, especially highly constrained data consisting of lists of artificial tokens such as part numbers or other identifiers, this appearance is correct. For natural-language data, however, the whitespace processing prescribed for these values is not only unreliable but will systematically remove the information needed to perform unwrapping correctly. For Asian scripts, for example, a correct unwrapping process will replace line boundaries not with blanks but with zero-width separators or nothing. In consequence, it is normally unwise to use these values for natural-language data, or for any data other than lists of highly constrained tokens.

##### <a id="dc-whiteSpace"></a>4.3.6.1 The whiteSpace Schema Component

---

---

  Schema Component: <a id="f-w">whiteSpace</a>, a kind of [Constraining Facet](#f)   <a id="f-w-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-w-value"></a>{value}  One of {preserve, replace, collapse}. Required.  <a id="f-w-fixed"></a>{fixed}  An xs:boolean value. Required.

---

---

 If [{fixed}](#f-w-fixed) is *true*, then types for which the current type is the [{base type definition}](#std-base_type_definition) cannot specify a value for [whiteSpace](#f-w) other than [{value}](#f-w-value).

##### <a id="xr-whiteSpace"></a>4.3.6.2 XML Representation of whiteSpace Schema Components

 The XML representation for a [whiteSpace](#f-w) schema component is a [<whiteSpace>](#element-whiteSpace) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `whiteSpace` Element Information Item

<a id="element-whiteSpace"></a>

```xml
<whiteSpace
  fixed = boolean : false
  id = ID
  value = (collapse | preserve | replace)
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</whiteSpace>
```

**[whiteSpace](#dc-whiteSpace) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-w-value): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- [{fixed}](#f-w-fixed): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `fixed` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise ***false***
- [{annotations}](#f-w-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-one) of the [<whiteSpace>](#element-whiteSpace) element, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

##### <a id="whiteSpace-validation-rules"></a>4.3.6.3 whiteSpace Validation Rules

**Note:** There are no [Validation Rule](#dt-cvc)s associated with [whiteSpace](#dt-whiteSpace). For more information, see the discussion on white space normalization in [Schema Component Details](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#components) in [[XSD 1.1 Part 1: Structures]](#structural-schemas), in particular the section [3.1.4 White Space Normalization during Validation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#sec-wsnormalization).

##### <a id="whiteSpace-coss"></a>4.3.6.4 Constraints on whiteSpace Schema Components

<a id="whiteSpace-valid-restriction"></a>**Schema Component Constraint: whiteSpace valid restriction**

> It is an [error](#dt-error) if [whiteSpace](#f-w) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and any of the following conditions is true:
>
> - 1  [{value}](#f-w-value) is *replace* or *preserve* and the [{value}](#f-w-value) of the parent [whiteSpace](#f-w) is *collapse*
> - 2  [{value}](#f-w-value) is *preserve* and the [{value}](#f-w-value) of the parent [whiteSpace](#f-w) is *replace*

**Note:** In order of increasing restrictiveness, the legal values for the [whiteSpace](#f-w) facet are ***preserve***, ***collapse***, and ***replace***. The more restrictive keywords are more restrictive not in the sense of accepting progressively fewer instance documents but in the sense that each corresponds to a progressively smaller, more tightly restricted value space.

#### <a id="rf-maxInclusive"></a>4.3.7 maxInclusive

 <a id="dt-maxInclusive">[Definition:]</a> maxInclusive is the inclusive upper bound of the [value space](#dt-value-space) for a datatype with the [ordered](#dt-ordered) property.  The value of **maxInclusive** [**must**](#dt-must) be equal to some value in the [value space](#dt-value-space) of the [base type](#dt-basetype).

 [maxInclusive](#dt-maxInclusive) provides for:

-  Constraining a [value space](#dt-value-space) to values with a specific inclusive upper bound.

*Example*

  The following is the definition of a [user-defined](#dt-user-defined) datatype which limits values to integers less than or equal to 100, using [maxInclusive](#dt-maxInclusive).

```
<simpleType name='one-hundred-or-less'>
  <restriction base='integer'>
    <maxInclusive value='100'/>
  </restriction>
</simpleType>
```

##### <a id="dc-maxInclusive"></a>4.3.7.1 The maxInclusive Schema Component

---

---

  Schema Component: <a id="f-mai">maxInclusive</a>, a kind of [Constraining Facet](#f)   <a id="f-mai-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-mai-value"></a>{value}  Required.

A value from the [value space](#dt-value-space) of the [{base type definition}](#std-base_type_definition).

   <a id="f-mai-fixed"></a>{fixed}  An xs:boolean value. Required.

---

---

 If [{fixed}](#f-mai-fixed) is *true*, then types for which the current type is the [{base type definition}](#std-base_type_definition) cannot specify a value for [maxInclusive](#f-mai) other than [{value}](#f-mai-value).

##### <a id="xr-maxInclusive"></a>4.3.7.2 XML Representation of maxInclusive Schema Components

 The XML representation for a [maxInclusive](#f-mai) schema component is a [<maxInclusive>](#element-maxInclusive) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `maxInclusive` Element Information Item

<a id="element-maxInclusive"></a>

```xml
<maxInclusive
  fixed = boolean : false
  id = ID
  value = anySimpleType
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</maxInclusive>
```

 [{value}](#f-mai-value) [**must**](#dt-must) be equal to some value in the [value space](#dt-value-space) of [{base type definition}](#std-base_type_definition).

**[maxInclusive](#dt-maxInclusive) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-mai-value): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- [{fixed}](#f-mai-fixed): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `fixed` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise ***false***
- [{annotations}](#f-mai-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-one) of the [<maxInclusive>](#element-maxInclusive) element, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

##### <a id="maxInclusive-validation-rules"></a>4.3.7.3 maxInclusive Validation Rules

<a id="cvc-maxInclusive-valid"></a>**Validation Rule: maxInclusive Valid**

> A value in an [ordered](#dt-ordered) [value space](#dt-value-space) is facet-valid with respect to [maxInclusive](#dt-maxInclusive) if and only if the value is less than or equal to [{value}](#f-mie-value), according to the datatype's order relation.

##### <a id="maxInclusive-coss"></a>4.3.7.4 Constraints on maxInclusive Schema Components

<a id="minInclusive-less-than-equal-to-maxInclusive"></a>**Schema Component Constraint: minInclusive <= maxInclusive**

> It is an [error](#dt-error) for the value specified for [minInclusive](#dt-minInclusive) to be greater than the value specified for [maxInclusive](#dt-maxInclusive) for the same datatype.

<a id="maxInclusive-valid-restriction"></a>**Schema Component Constraint: maxInclusive valid restriction**

> It is an [error](#dt-error) if any of the following conditions is true:
>
> - 1  [maxInclusive](#f-mai) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mai-value) is greater than the [{value}](#f-mai-value) of that [maxInclusive](#f-mai).
> - 2  [maxExclusive](#f-mae) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mai-value) is greater than or equal to the [{value}](#f-mae-value) of that [maxExclusive](#f-mae).
> - 3  [minInclusive](#f-mii) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mai-value) is less than the [{value}](#f-mii-value) of that [minInclusive](#f-mii).
> - 4  [minExclusive](#f-mie) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mai-value) is less than or equal to the [{value}](#f-mie-value) of that [minExclusive](#f-mie).

#### <a id="rf-maxExclusive"></a>4.3.8 maxExclusive

 <a id="dt-maxExclusive">[Definition:]</a> **maxExclusive** is the exclusive upper bound of the [value space](#dt-value-space) for a datatype with the [ordered](#dt-ordered) property.  The value of **maxExclusive**  [**must**](#dt-must) be equal to some value in the [value space](#dt-value-space) of the [base type](#dt-basetype) or be equal to [{value}](#f-mae-value) in [{base type definition}](#std-base_type_definition).

 [maxExclusive](#dt-maxExclusive) provides for:

-  Constraining a [value space](#dt-value-space) to values with a specific exclusive upper bound.

*Example*

  The following is the definition of a [user-defined](#dt-user-defined) datatype which limits values to integers less than or equal to 100, using [maxExclusive](#dt-maxExclusive).

```
<simpleType name='less-than-one-hundred-and-one'>
  <restriction base='integer'>
    <maxExclusive value='101'/>
  </restriction>
</simpleType>
```

  Note that the [value space](#dt-value-space) of this datatype is identical to the previous one (named 'one-hundred-or-less').

##### <a id="dc-maxExclusive"></a>4.3.8.1 The maxExclusive Schema Component

---

---

  Schema Component: <a id="f-mae">maxExclusive</a>, a kind of [Constraining Facet](#f)   <a id="f-mae-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-mae-value"></a>{value}  Required.

A value from the [value space](#dt-value-space) of the [{base type definition}](#std-base_type_definition).

   <a id="f-mae-fixed"></a>{fixed}  An xs:boolean value. Required.

---

---

 If [{fixed}](#f-mae-fixed) is *true*, then types for which the current type is the [{base type definition}](#std-base_type_definition) cannot specify a value for [maxExclusive](#f-mae) other than [{value}](#f-mae-value).

##### <a id="xr-maxExclusive"></a>4.3.8.2 XML Representation of maxExclusive Schema Components

 The XML representation for a [maxExclusive](#f-mae) schema component is a [<maxExclusive>](#element-maxExclusive) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `maxExclusive` Element Information Item

<a id="element-maxExclusive"></a>

```xml
<maxExclusive
  fixed = boolean : false
  id = ID
  value = anySimpleType
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</maxExclusive>
```

 [{value}](#f-mae-value) [**must**](#dt-must) be equal to some value in the [value space](#dt-value-space) of [{base type definition}](#std-base_type_definition).

**[maxExclusive](#dt-maxExclusive) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-mae-value): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- [{fixed}](#f-mae-fixed): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `fixed` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise ***false***
- [{annotations}](#f-mae-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-one) of the [<maxExclusive>](#element-maxExclusive) element, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

##### <a id="maxExclusive-validation-rules"></a>4.3.8.3 maxExclusive Validation Rules

<a id="cvc-maxExclusive-valid"></a>**Validation Rule: maxExclusive Valid**

> A value in an [ordered](#dt-ordered) [value space](#dt-value-space) is facet-valid with respect to [maxExclusive](#dt-maxExclusive) if and only if the value is less than [{value}](#f-mie-value), according to the datatype's order relation.

##### <a id="maxExclusive-coss"></a>4.3.8.4 Constraints on maxExclusive Schema Components

<a id="maxInclusive-maxExclusive"></a>**Schema Component Constraint: maxInclusive and maxExclusive**

> It is an [error](#dt-error) for both [maxInclusive](#dt-maxInclusive) and [maxExclusive](#dt-maxExclusive) to be specified in the same derivation step of a [Simple Type Definition](#std).

<a id="minExclusive-less-than-equal-to-maxExclusive"></a>**Schema Component Constraint: minExclusive <= maxExclusive**

> It is an [error](#dt-error) for the value specified for [minExclusive](#dt-minExclusive) to be greater than the value specified for [maxExclusive](#dt-maxExclusive) for the same datatype.

<a id="maxExclusive-valid-restriction"></a>**Schema Component Constraint: maxExclusive valid restriction**

> It is an [error](#dt-error) if any of the following conditions is true:
>
> - 1  [maxExclusive](#f-mae) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mae-value) is greater than the [{value}](#f-mae-value) of that [maxExclusive](#f-mae).
> - 2  [maxInclusive](#f-mai) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mae-value) is greater than the [{value}](#f-mai-value) of that [maxInclusive](#f-mai).
> - 3  [minInclusive](#f-mii) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mae-value) is less than or equal to the [{value}](#f-mii-value) of that [minInclusive](#f-mii).
> - 4  [minExclusive](#f-mie) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mae-value) is less than or equal to the [{value}](#f-mie-value) of that [minExclusive](#f-mie).

#### <a id="rf-minExclusive"></a>4.3.9 minExclusive

 <a id="dt-minExclusive">[Definition:]</a> **minExclusive** is the exclusive lower bound of the [value space](#dt-value-space) for a datatype with the [ordered](#dt-ordered) property. The value of **minExclusive** [**must**](#dt-must) be equal to some value in the [value space](#dt-value-space) of the [base type](#dt-basetype) or be equal to [{value}](#f-mie-value) in [{base type definition}](#std-base_type_definition).

 [minExclusive](#dt-minExclusive) provides for:

-  Constraining a [value space](#dt-value-space) to values with a specific exclusive lower bound.

*Example*

  The following is the definition of a [user-defined](#dt-user-defined) datatype which limits values to integers greater than or equal to 100, using [minExclusive](#dt-minExclusive).

```
<simpleType name='more-than-ninety-nine'>
  <restriction base='integer'>
    <minExclusive value='99'/>
  </restriction>
</simpleType>
```

  Note that the [value space](#dt-value-space) of this datatype is identical to the following one (named 'one-hundred-or-more').

##### <a id="dc-minExclusive"></a>4.3.9.1 The minExclusive Schema Component

---

---

  Schema Component: <a id="f-mie">minExclusive</a>, a kind of [Constraining Facet](#f)   <a id="f-mie-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-mie-value"></a>{value}  Required.

A value from the [value space](#dt-value-space) of the [{base type definition}](#std-base_type_definition).

   <a id="f-mie-fixed"></a>{fixed}  An xs:boolean value. Required.

---

---

 If [{fixed}](#f-mie-fixed) is *true*, then types for which the current type is the [{base type definition}](#std-base_type_definition) cannot specify a value for [minExclusive](#f-mie) other than [{value}](#f-mie-value).

##### <a id="xr-minExclusive"></a>4.3.9.2 XML Representation of minExclusive Schema Components

 The XML representation for a [minExclusive](#f-mie) schema component is a [<minExclusive>](#element-minExclusive) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `minExclusive` Element Information Item

<a id="element-minExclusive"></a>

```xml
<minExclusive
  fixed = boolean : false
  id = ID
  value = anySimpleType
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</minExclusive>
```

 [{value}](#f-mie-value) [**must**](#dt-must) be equal to some value in the [value space](#dt-value-space) of [{base type definition}](#std-base_type_definition).

**[minExclusive](#dt-minExclusive) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-mie-value): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- [{fixed}](#f-mie-fixed): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `fixed` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise ***false***
- [{annotations}](#f-mie-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-one) of the [<minExclusive>](#element-minExclusive) element, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

##### <a id="minExclusive-validation-rules"></a>4.3.9.3 minExclusive Validation Rules

<a id="cvc-minExclusive-valid"></a>**Validation Rule: minExclusive Valid**

> A value in an [ordered](#dt-ordered) [value space](#dt-value-space) is facet-valid with respect to [minExclusive](#dt-minExclusive) if and only if the value is greater than [{value}](#f-mie-value), according to the datatype's order relation.

##### <a id="minExclusive-coss"></a>4.3.9.4 Constraints on minExclusive Schema Components

<a id="minInclusive-minExclusive"></a>**Schema Component Constraint: minInclusive and minExclusive**

> It is an [error](#dt-error) for both [minInclusive](#dt-minInclusive) and [minExclusive](#dt-minExclusive) to be specified in the same derivation step of a [Simple Type Definition](#std).

<a id="minExclusive-less-than-maxInclusive"></a>**Schema Component Constraint: minExclusive < maxInclusive**

> It is an [error](#dt-error) for the value specified for [minExclusive](#dt-minExclusive) to be greater than or equal to the value specified for [maxInclusive](#dt-maxInclusive) for the same datatype.

<a id="minExclusive-valid-restriction"></a>**Schema Component Constraint: minExclusive valid restriction**

> It is an [error](#dt-error) if any of the following conditions is true:
>
> - 1  [minExclusive](#f-mie) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mie-value) is less than the [{value}](#f-mie-value) of that [minExclusive](#f-mie).
> - 2  [minInclusive](#f-mii) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mie-value) is less than the [{value}](#f-mii-value) of that [minInclusive](#f-mii).
> - 3  [maxInclusive](#f-mai) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mie-value) is greater than or equal to the [{value}](#f-mai-value) of that [maxInclusive](#f-mai).
> - 4  [maxExclusive](#f-mae) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mae-value) is greater than or equal to the [{value}](#f-mae-value) of that [maxExclusive](#f-mae).

#### <a id="rf-minInclusive"></a>4.3.10 minInclusive

 <a id="dt-minInclusive">[Definition:]</a> **minInclusive** is the inclusive lower bound of the [value space](#dt-value-space) for a datatype with the [ordered](#dt-ordered) property.  The value of **minInclusive**  [**must**](#dt-must) be equal to some value in the [value space](#dt-value-space) of the [base type](#dt-basetype).

 [minInclusive](#dt-minInclusive) provides for:

-  Constraining a [value space](#dt-value-space) to values with a specific inclusive lower bound.

*Example*

  The following is the definition of a [user-defined](#dt-user-defined) datatype which limits values to integers greater than or equal to 100, using [minInclusive](#dt-minInclusive).

```
<simpleType name='one-hundred-or-more'>
  <restriction base='integer'>
    <minInclusive value='100'/>
  </restriction>
</simpleType>
```

##### <a id="dc-minInclusive"></a>4.3.10.1 The minInclusive Schema Component

---

---

  Schema Component: <a id="f-mii">minInclusive</a>, a kind of [Constraining Facet](#f)   <a id="f-mii-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-mii-value"></a>{value}  Required.

A value from the [value space](#dt-value-space) of the [{base type definition}](#std-base_type_definition).

   <a id="f-mii-fixed"></a>{fixed}  An xs:boolean value. Required.

---

---

 If [{fixed}](#f-mii-fixed) is *true*, then types for which the current type is the [{base type definition}](#std-base_type_definition) cannot specify a value for [minInclusive](#f-mii) other than [{value}](#f-mii-value).

##### <a id="xr-minInclusive"></a>4.3.10.2 XML Representation of minInclusive Schema Components

 The XML representation for a [minInclusive](#f-mii) schema component is a [<minInclusive>](#element-minInclusive) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `minInclusive` Element Information Item

<a id="element-minInclusive"></a>

```xml
<minInclusive
  fixed = boolean : false
  id = ID
  value = anySimpleType
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</minInclusive>
```

 [{value}](#f-mii-value) [**must**](#dt-must) be equal to some value in the [value space](#dt-value-space) of [{base type definition}](#std-base_type_definition).

**[minInclusive](#dt-minInclusive) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-mii-value): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- [{fixed}](#f-mii-fixed): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `fixed` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise ***false***
- [{annotations}](#f-mii-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-one) of the [<minInclusive>](#element-minInclusive) element, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

##### <a id="minInclusive-validation-rules"></a>4.3.10.3 minInclusive Validation Rules

<a id="cvc-minInclusive-valid"></a>**Validation Rule: minInclusive Valid**

> A value in an [ordered](#dt-ordered) [value space](#dt-value-space) is facet-valid with respect to [minInclusive](#dt-minInclusive) if and only if the value is greater than or equal to [{value}](#f-mie-value), according to the datatype's order relation.

##### <a id="minInclusive-coss"></a>4.3.10.4 Constraints on minInclusive Schema Components

<a id="minInclusive-less-than-maxExclusive"></a>**Schema Component Constraint: minInclusive < maxExclusive**

> It is an [error](#dt-error) for the value specified for [minInclusive](#dt-minInclusive) to be greater than or equal to the value specified for [maxExclusive](#dt-maxExclusive) for the same datatype.

<a id="minInclusive-valid-restriction"></a>**Schema Component Constraint: minInclusive valid restriction**

> It is an [error](#dt-error) if any of the following conditions is true:
>
> - 1  [minInclusive](#f-mii) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mii-value) is less than the [{value}](#f-mii-value) of that [minInclusive](#f-mii).
> - 2  [maxInclusive](#f-mai) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mii-value) is greater the [{value}](#f-mai-value) of that [maxInclusive](#f-mai).
> - 3  [minExclusive](#f-mie) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mii-value) is less than or equal to the [{value}](#f-mie-value) of that [minExclusive](#f-mie).
> - 4  [maxExclusive](#f-mae) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-mii-value) is greater than or equal to the [{value}](#f-mae-value) of that [maxExclusive](#f-mae).

#### <a id="rf-totalDigits"></a>4.3.11 totalDigits

 <a id="dt-totalDigits">[Definition:]</a> **totalDigits** restricts the magnitude and arithmetic precision of values in the [value spaces](#dt-value-space) of [decimal](#decimal) and datatypes derived from it.

 For [decimal](#decimal), if the [{value}](#f-td-value) of [totalDigits](#f-td) is *t*, the effect is to require that values be equal to *i* / 10^*n*^, for some integers *i* and *n*, with | *i* | < 10^*t*^ and 0 ≤ *n* ≤ *t*. This has as a consequence that the values are expressible using at most *t* digits in decimal notation.

 The [{value}](#f-td-value) of [totalDigits](#f-td) **must** be a [positiveInteger](#positiveInteger).

 The term 'totalDigits' is chosen to reflect the fact that it restricts the [value space](#dt-value-space) to those values that can be represented lexically using at most *totalDigits* digits in decimal notation, or at most *totalDigits* digits for the coefficient, in scientific notation.  Note that it does not restrict the [lexical space](#dt-lexical-space) directly; a lexical representation that adds non-significant leading or trailing zero digits is still permitted. It also has no effect on the values NaN, INF, and -INF.

##### <a id="dc-totalDigits"></a>4.3.11.1 The totalDigits Schema Component

---

---

  Schema Component: <a id="f-td">totalDigits</a>, a kind of [Constraining Facet](#f)   <a id="f-td-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-td-value"></a>{value}  An xs:positiveInteger value. Required.  <a id="f-td-fixed"></a>{fixed}  An xs:boolean value. Required.

---

---

 If [{fixed}](#f-td-fixed) is *true*, then types for which the current type is the [{base type definition}](#std-base_type_definition) **must** not specify a value for [totalDigits](#f-td) other than [{value}](#f-td-value).

##### <a id="xr-totalDigits"></a>4.3.11.2 XML Representation of totalDigits Schema Components

 The XML representation for a [totalDigits](#f-td) schema component is a [<totalDigits>](#element-totalDigits) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `totalDigits` Element Information Item

<a id="element-totalDigits"></a>

```xml
<totalDigits
  fixed = boolean : false
  id = ID
  value = positiveInteger
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</totalDigits>
```

**[totalDigits](#dc-totalDigits) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-td-value): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- [{fixed}](#f-td-fixed): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `fixed` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise ***false***
- [{annotations}](#f-td-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-one) of the [<totalDigits>](#element-totalDigits) element, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

##### <a id="totalDigits-validation-rules"></a>4.3.11.3 totalDigits Validation Rules

<a id="cvc-totalDigits-valid"></a>**Validation Rule: totalDigits Valid**

> A value *v* is facet-valid with respect to a [totalDigits](#f-td) facet with a [{value}](#f-td-value) of *t* if and only if *v* is a [decimal](#decimal) value equal to *i* / 10^*n*^, for some integers *i* and *n*, with | *i* | < 10^*t*^ and 0 ≤ *n* ≤ *t*.

##### <a id="totalDigits-coss"></a>4.3.11.4 Constraints on totalDigits Schema Components

<a id="totalDigits-valid-restriction"></a>**Schema Component Constraint: totalDigits valid restriction**

> It is an [error](#dt-error) if the [owner](#dt-owner)'s [{base type definition}](#std-base_type_definition) has a [totalDigits](#f-td) facet among its [{facets}](#std-facets) and [{value}](#f-td-value) is greater than the [{value}](#f-td-value) of that [totalDigits](#f-td) facet.

#### <a id="rf-fractionDigits"></a>4.3.12 fractionDigits

 <a id="dt-fractionDigits">[Definition:]</a> **fractionDigits** places an upper limit on the arithmetic precision of [decimal](#decimal) values: if the [{value}](#f-fd-value) of **fractionDigits** = *f*, then the value space is restricted to values equal to *i* / 10^*n*^ for some integers *i* and *n* and 0 ≤ *n* ≤ *f*. The value of **fractionDigits** [**must**](#dt-must) be a [nonNegativeInteger](#nonNegativeInteger)

 The term **fractionDigits** is chosen to reflect the fact that it restricts the [value space](#dt-value-space) to those values that can be represented lexically in decimal notation using at most *fractionDigits* to the right of the decimal point. Note that it does not restrict the [lexical space](#dt-lexical-space) directly; a lexical representation that adds non-significant leading or trailing zero digits is still permitted.

*Example*

  The following is the definition of a [user-defined](#dt-user-defined) datatype which could be used to represent the magnitude of a person's body temperature on the Celsius scale. This definition would appear in a schema authored by an "end-user" and shows how to define a datatype by specifying facet values which constrain the range of the [base type](#dt-basetype).

```
<simpleType name='celsiusBodyTemp'>
  <restriction base='decimal'>
    <fractionDigits value='1'/>
    <minInclusive value='32'/>
	 <maxInclusive value='41.7'/>
  </restriction>
</simpleType>

```

##### <a id="dc-fractionDigits"></a>4.3.12.1 The fractionDigits Schema Component

---

---

  Schema Component: <a id="f-fd">fractionDigits</a>, a kind of [Constraining Facet](#f)   <a id="f-fd-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-fd-value"></a>{value}  An xs:nonNegativeInteger value. Required.  <a id="f-fd-fixed"></a>{fixed}  An xs:boolean value. Required.

---

---

 If [{fixed}](#f-fd-fixed) is *true*, then types for which the current type is the [{base type definition}](#std-base_type_definition) **must** not specify a value for [fractionDigits](#f-fd) other than [{value}](#f-fd-value).

##### <a id="xr-fractionDigits"></a>4.3.12.2 XML Representation of fractionDigits Schema Components

 The XML representation for a [fractionDigits](#f-fd) schema component is a [<fractionDigits>](#element-fractionDigits) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `fractionDigits` Element Information Item

<a id="element-fractionDigits"></a>

```xml
<fractionDigits
  fixed = boolean : false
  id = ID
  value = nonNegativeInteger
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</fractionDigits>
```

**[fractionDigits](#dc-fractionDigits) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-fd-value): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- [{fixed}](#f-fd-fixed): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `fixed` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise ***false***
- [{annotations}](#f-fd-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-one) of the [<fractionDigits>](#element-fractionDigits) element, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

##### <a id="fractionDigits-validation-rules"></a>4.3.12.3 fractionDigits Validation Rules

<a id="cvc-fractionDigits-valid"></a>**Validation Rule: fractionDigits Valid**

> A value is facet-valid with respect to [fractionDigits](#dt-fractionDigits) if and only if that value is equal to *i* / 10^*n*^ for integer *i* and *n*, with 0 ≤ *n* ≤ [{value}](#f-fd-value).

##### <a id="fractionDigits-coss"></a>4.3.12.4 Constraints on fractionDigits Schema Components

<a id="fractionDigits-totalDigits"></a>**Schema Component Constraint: fractionDigits less than or equal to totalDigits**

> It is an [error](#dt-error) for the [{value}](#f-fd-value) of [fractionDigits](#f-fd) to be greater than that of [totalDigits](#f-td).

<a id="fractionDigits-valid-restriction"></a>**Schema Component Constraint: fractionDigits valid restriction**

> It is an [error](#dt-error) if [fractionDigits](#dt-fractionDigits) is among the members of [{facets}](#std-facets) of [{base type definition}](#std-base_type_definition) and [{value}](#f-fd-value) is greater than the [{value}](#f-fd-value) of that [fractionDigits](#dt-fractionDigits).

#### <a id="rf-assertions"></a>4.3.13 Assertions

<a id="dt-assertions">[Definition:]</a> **Assertions** constrain the [value space](#dt-value-space) by requiring the values to satisfy specified XPath ([[XPath 2.0]](#XPATH2)) expressions. The value of the [assertions](#f-a) facet is a sequence of [Assertion](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#as) components as defined in [[XSD 1.1 Part 1: Structures]](#structural-schemas).

*Example*

The following is the definition of a [user-defined](#dt-user-defined) datatype which allows all integers but 0 by using an assertion to disallow the value 0.

```
<simpleType name='nonZeroInteger'>
  <restriction base='integer'>
    <assertion test='$value ne 0'/>
  </restriction>
</simpleType>

```

*Example*

The following example defines the datatype "triple", whose [value space](#dt-value-space) is the set of integers evenly divisible by three.

```
<simpleType name='triple'>
  <restriction base='integer'>
    <assertion test='$value mod 3 eq 0'/>
  </restriction>
</simpleType>
```

The same datatype can be defined without the use of assertions, but the pattern necessary to represent the set of triples is long and error-prone:

```
<simpleType name='triple'>
  <restriction base='integer'>
    <pattern value=
    "([0369]|[147][0369]*[258]|(([258]|[147][0369]*[147])([0369]|[258][0369]*[147])*([147]|[258][0369]*[258]))*"/>
  </restriction>
</simpleType>
```

The assertion used in the first version of "triple" is likely to be clearer for many readers of the schema document.

##### <a id="dc-assertions"></a>4.3.13.1 The assertions Schema Component

---

---

  Schema Component: <a id="f-a">assertions</a>, a kind of [Constraining Facet](#f)   <a id="f-a-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-a-value"></a>{value}

A sequence of [Assertion](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#as) components.

---

---

##### <a id="xr-assertions"></a>4.3.13.2 XML Representation of assertions Schema Components

The XML representation for an [assertions](#f-a) schema component is one or more [<assertion>](#element-assertion) element information items. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `assertion` Element Information Item

<a id="element-assertion"></a>

```xml
<assertion
  id = ID
  test = an XPath expression
  xpathDefaultNamespace =

    (anyURI | (##defaultNamespace | ##targetNamespace | ##local))

  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</assertion>
```

**[assertions](#dc-assertions) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-a-value): A sequence whose members are [Assertion](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#as)s drawn from the following sources, in order:

- 1 If the [{base type definition}](#std-base_type_definition) of the [owner](#dt-owner) has an [assertions](#f-a) facet among its [{facets}](#std-facets), then the [Assertion](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#as)s which appear in the [{value}](#f-p-value) of that [assertions](#f-a) facet.
- 2 [Assertion](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#as)s corresponding to the [<assertion>](#element-assertion) element information items among the [[children]](https://www.w3.org/TR/xml-infoset/#infoitem.element) of [<restriction>](#element-restriction), if any, in document order. For details of the construction of the [Assertion](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#as) components, see [section 3.13.2](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-assertion) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).
- [{annotations}](#f-a-annotations): The empty sequence. **Note:** Annotations specified within an [<assertion>](#element-assertion) element are captured by the individual [Assertion](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#as) component to which it maps.

---

##### <a id="assertions-validation-rules"></a>4.3.13.3 Assertions Validation Rules

The following rule refers to "the nearest built-in" datatype and to the "XDM representation" of a value under a datatype. <a id="dt-optype">[Definition:]</a>For any datatype *T*, the **nearest built-in datatype** to *T* is the first [built-in](#dt-built-in) datatype encountered in following the chain of links connecting each datatype to its [base type](#dt-basetype). If *T* is a [built-in](#dt-built-in) datatype, then the nearest built-in datatype of *T* is *T* itself; otherwise, it is the nearest built-in datatype of *T*'s [base type](#dt-basetype).

 <a id="dt-xdmrep">[Definition:]</a>For any value *V* and any datatype *T*, the **XDM representation of *V* under *T*** is defined recursively as follows. Call the XDM representation *X*. Then

- 1 If *T* = [`xs:anySimpleType`](#dt-anySimpleType) or [`xs:anyAtomicType`](#dt-anyAtomicType) then *X* is *V*, and the [dynamic type](https://www.w3.org/TR/xpath20/#dt-dynamic-type) of *X* is `xs:untypedAtomic`.
- 2 If *T* . [{variety}](#std-variety) = ***atomic***, then let *T2* be the [nearest built-in datatype](#dt-optype) to *T*. If *V* is a member of the [value space](#dt-value-space) of *T2*, then *X* is *V* and the [dynamic type](https://www.w3.org/TR/xpath20/#dt-dynamic-type) of *X* is *T2*. Otherwise (i.e. if *V* is not a member of the [value space](#dt-value-space) of *T2*), *X* is the [XDM representation](#dt-xdmrep) of *V* under *T2* . [{base type definition}](#std-base_type_definition).
- 3 If *T* . [{variety}](#std-variety) = ***list***, then *X* is a sequence of atomic values, each atomic value being the [XDM representation](#dt-xdmrep) of the corresponding item in the list *V* under *T* . [{item type definition}](#std-item_type_definition).
- 4 If *T* . [{variety}](#std-variety) = ***union***, then *X* is the [XDM representation](#dt-xdmrep) of *V* under the [active basic member](#dt-active-basic-member) of *V* when validated against *T*. If there is no [active basic member](#dt-active-basic-member), then *V* has no [XDM representation](#dt-xdmrep) under *T*.
  **Note:** If the [{item type definition}](#std-item_type_definition) of a [list](#dt-list) is a [union](#dt-union), or the [active basic member](#dt-active-basic-member) is a [list](#dt-list), then several steps may be necessary before the [atomic](#dt-atomic) datatype which serves as the [dynamic type](https://www.w3.org/TR/xpath20/#dt-dynamic-type) of *X* is found. Because the [{item type definition}](#std-item_type_definition) of a [list](#dt-list) is required to be an [atomic](#dt-atomic) or [union](#dt-union) datatype, and the [active basic member](#dt-active-basic-member) of a [union](#dt-union) which accepts the value *V* is by definition not a [union](#dt-union), the recursive rule given above is guaranteed to terminate in a sequence of one or more [atomic](#dt-atomic) values, each belonging to an [atomic](#dt-atomic) datatype. <a id="cvc-assertions-valid"></a>**Validation Rule: Assertions Valid**

> A value *V* is facet-valid with respect to an [assertions](#f-a) facet belonging to a simple type *T* if and only if the {test} property of each [Assertion](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#as) in its [{value}](#f-a-value) evaluates to `true` under the conditions laid out below, without raising any [dynamic error](https://www.w3.org/TR/2007/REC-xpath20-20070123/#dt-dynamic-error) or [type error](https://www.w3.org/TR/2007/REC-xpath20-20070123/#dt-type-error).Evaluation of {test} is performed as defined in [[XPath 2.0]](#XPATH2), with the following conditions:
>
> - 1 The XPath expression {test} is evaluated, following the rules given in [XPath Evaluation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#cvc-xpath) of [[XSD 1.1 Part 1: Structures]](#structural-schemas), with the following modifications.
>
> - 1.1  The [in-scope variables](https://www.w3.org/TR/2007/REC-xpath20-20070123/#dt-in-scope-variables) in the [static context](https://www.w3.org/TR/2007/REC-xpath20-20070123/#dt-static-context) is a set with a single member. The `expanded QName` of that member has no namespace URI and has '`value`' as the local name. The (static) `type` of the member is `anyAtomicType*`. **Note:** The XDM type label `anyAtomicType*` simply says that for static typing purposes the variable `$value` will have a value consisting of a sequence of zero or more atomic values.
> - 1.2 There is no [context item](https://www.w3.org/TR/xpath20/#dt-context-item) for the evaluation of the XPath expression. **Note:** In the terminology of [[XPath 2.0]](#XPATH2), the [context item](https://www.w3.org/TR/xpath20/#dt-context-item) is "undefined". **Note:** As a consequence the expression '`.`', or any implicit or explicit reference to the context item, will raise a dynamic error, which will cause the assertion to be treated as false. If an error is detected statically, then the assertion violates the schema component constraint [XPath Valid](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#xpath-valid) and causes an error to be flagged in the schema. The variable "`$value`" can be used to refer to the value being checked.
> - 1.3 There is likewise no value for the [context size](https://www.w3.org/TR/xpath20/#dt-context-size) and the [context position](https://www.w3.org/TR/xpath20/#dt-context-position) in the [dynamic context](https://www.w3.org/TR/2007/REC-xpath20-20070123/#dt-dynamic-context) used for evaluation of the assertion.
> - 1.4  The [variable values](https://www.w3.org/TR/2007/REC-xpath20-20070123/#dt-variable-values) in the [dynamic context](https://www.w3.org/TR/2007/REC-xpath20-20070123/#dt-dynamic-context) is a set with a single member. The `expanded QName` of that member has no namespace URI and '`value`' as the local name. The `value` of the member is the [XDM representation](#dt-xdmrep) of *V* under *T*.
> - 1.5 If *V* has no [XDM representation](#dt-xdmrep) under *T*, then the XPath expression cannot usefully be evaluated, and *V* is not facet-valid against the [assertions](#f-a) facet of *T*.
>
> - 2 The evaluation result is converted to either `true` or `false` as if by a call to the XPath [fn:boolean](https://www.w3.org/TR/2007/REC-xpath-functions-20070123/#func-boolean) function.

##### <a id="assertions-coss"></a>4.3.13.4 Constraints on assertions Schema Components

<a id="cos-assertions-restriction"></a>**Schema Component Constraint: Valid restriction of assertions**

> The [{value}](#f-a-value) of the [assertions](#f-a) facet on the [{base type definition}](#std-base_type_definition) **must** be a prefix of the [{value}](#f-a-value).**Note:** For components constructed from XML representations in schema documents, the satisfaction of this constraint is a consequence of the XML mapping rules: any assertion imposed by a simple type definition *S* will always also be imposed by any type derived from *S* by [facet-based restriction](#dt-fb-restriction). This constraint ensures that components constructed by other means (so-called "born-binary" components) similarly preserve [assertions](#f-a) facets across [facet-based restriction](#dt-fb-restriction).

#### <a id="rf-explicitTimezone"></a>4.3.14 explicitTimezone

<a id="dt-timezone">[Definition:]</a> **explicitTimezone** is a three-valued facet which can can be used to require or prohibit the time zone offset in date/time datatypes.

*Example*

  The following [user-defined](#dt-user-defined) datatype accepts only [date](#date) values without a time zone offset, using the [explicitTimezone](#f-tz) facet.

```
<simpleType name='bare-date'>
  <restriction base='date'>
    <explicitTimezone value='prohibited'/>
  </restriction>
</simpleType>
```

  The same effect could also be achieved using the [pattern](#f-p) facet, as shown below, but it is somewhat less clear what is going on in this derivation, and it is better practice to use the more straightforward [explicitTimezone](#f-tz) for this purpose.

```
<simpleType name='bare-date'>
  <restriction base='date'>
    <pattern value='[^:Z]*'/>
  </restriction>
</simpleType>
```

##### <a id="dc-explicitTimezone"></a>4.3.14.1 The explicitTimezone Schema Component

---

---

  Schema Component: <a id="f-tz">explicitTimezone</a>, a kind of [Constraining Facet](#f)   <a id="f-tz-annotations"></a>{annotations}  A sequence of [Annotation](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#a) components.   <a id="f-tz-value"></a>{value}  One of {required, prohibited, optional}. Required.  <a id="f-tz-fixed"></a>{fixed}  An xs:boolean value. Required.

---

---

If [{fixed}](#f-tz-fixed) is *true*, then datatypes for which the current type is the [{base type definition}](#std-base_type_definition) cannot specify a value for [explicitTimezone](#f-tz) other than [{value}](#f-tz-value).

**Note:** It is a consequence of [timezone valid restriction (§4.3.14.4)](#timezone-valid-restriction) that the value of the [explicitTimezone](#f-tz) facet cannot be changed unless that value is ***optional***, regardless of whether [{fixed}](#f-tz-fixed) is ***true*** or ***false***.  Accordingly, [{fixed}](#f-tz-fixed) is relevant only when [{value}](#f-tz-value) is ***optional***.

##### <a id="xr-timezone"></a>4.3.14.2 XML Representation of explicitTimezone Schema Components

 The XML representation for an [explicitTimezone](#f-tz) schema component is an [<explicitTimezone>](#element-explicitTimezone) element information item. The correspondences between the properties of the information item and properties of the component are as follows:

---

XML Representation Summary: `explicitTimezone` Element Information Item

<a id="element-explicitTimezone"></a>

```xml
<explicitTimezone
  fixed = boolean : false
  id = ID
  value = NCName
  {any attributes with non-schema namespace . . .}>
  Content: (annotation?)
</explicitTimezone>
```

**[explicitTimezone](#dc-explicitTimezone) **Schema Component****

- **Property**: **Representation**
- [{value}](#f-tz-value): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `value` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element)
- [{fixed}](#f-tz-fixed): The [actual value](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-vv) of the `fixed` [[attribute]](https://www.w3.org/TR/xml-infoset/#infoitem.element), if present, otherwise ***false***
- [{annotations}](#f-tz-annotations): The [annotation mapping](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#key-am-one) of the [<explicitTimezone>](#element-explicitTimezone) element, as defined in section [XML Representation of Annotation Schema Components](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#declare-annotation) of [[XSD 1.1 Part 1: Structures]](#structural-schemas).

---

##### <a id="timezone-vr"></a>4.3.14.3 explicitTimezone Validation Rules

<a id="cvc-explicitTimezone-valid"></a>**Validation Rule: explicitOffset Valid**

> A [dateTime](#dateTime) value *V* is facet-valid with respect to [explicitTimezone](#dt-timezone) if and only if **one** of the following is true
>
> - 1 The [{value}](#f-tz-value) of the facet is ***required*** and *V* has a (non-***absent***) value for the [*timezoneOffset*](#vp-dt-timezone) property.
> - 2 The [{value}](#f-tz-value) of the facet is ***prohibited*** and the value for the [*timezoneOffset*](#vp-dt-timezone) property in *V* is ***absent***.
> - 3 The [{value}](#f-tz-value) of the facet is ***optional***.

##### <a id="timezone-coss"></a>4.3.14.4 Constraints on explicitTimezone Schema Components

<a id="timezone-valid-restriction"></a>**Schema Component Constraint: timezone valid restriction**

> If the [explicitTimezone](#f-tz) facet on the [{base type definition}](#std-base_type_definition) has a [{value}](#f-tz-value) other than ***optional***, then the [{value}](#f-tz-value) of the facet on the [restriction](#dt-restriction) **must** be equal to the [{value}](#f-tz-value) on the [{base type definition}](#std-base_type_definition); otherwise it is an [error](#dt-error).**Note:** The effect of this rule is to allow datatypes with a [explicitTimezone](#f-tz) value of ***optional*** to be restricted by specifying a value of ***required*** or ***prohibited***, and to forbid any other derivations using this facet.

## <a id="conformance"></a>5 Conformance

*XSD 1.1: Datatypes* is intended to be usable in a variety of contexts.

In the usual case, it will embedded in a **host language** such as [[XSD 1.1 Part 1: Structures]](#structural-schemas), which refers to this specification normatively to define some part of the host language. In some cases, *XSD 1.1: Datatypes* may be implemented independently of any host language.

Certain aspects of the behavior of conforming processors are described in this specification as [implementation-defined](#key-impl-def) or [implementation-dependent](#key-impl-dep).

- <a id="key-impl-def">[Definition:]</a>Something which **may** vary among conforming implementations, but which **must** be specified by the implementor for each particular implementation, is **implementation-defined**.
- <a id="key-impl-dep">[Definition:]</a>Something which **may** vary among conforming implementations, is not specified by this or any W3C specification, and is not required to be specified by the implementor for any particular implementation, is **implementation-dependent**.
 Anything described in this specification as [implementation-defined](#key-impl-def) or [implementation-dependent](#key-impl-dep) **may** be further constrained by the specifications of a host language in which the datatypes and other material specified here are used. A list of implementation-defined and implementation-dependent features can be found in [Implementation-defined and implementation-dependent features (normative) (§H)](#idef-idep)

### <a id="hostlangs"></a>5.1 Host Languages

When *XSD 1.1: Datatypes* is embedded in a host language, the definition of conformance is specified by the host language, not by this specification. That is, when this specification is implemented in the context of an implementation of a host language, the question of conformance to this specification (separate from the host language) does not arise.

This specification imposes certain constraints on the embedding of *XSD 1.1: Datatypes* by a host language; these are indicated in the normative text by the use of the verbs '**must**', etc., with the phrase "host language" as the subject of the verb.

**Note:** For convenience, the most important of these constraints are noted here:

- Host languages **should** specify that all of the datatypes described here as built-ins are automatically available.
- Host languages **may** specify that additional datatypes are also made available automatically.
- If user-defined datatypes are to be supported in the host language, then the host language **must** specify how user-defined datatypes are defined and made available for use.

 In addition, host languages **must** require conforming implementations of the host language to obey all of the constraints and rules specified here.

### <a id="independent-impl"></a>5.2 Independent implementations

<a id="dt-minimally-conforming">[Definition:]</a>Implementations claiming **minimal conformance** to this specification independent of any host language **must** do **all** of the following:

- 1<a id="support-all-primitives"></a>Support all the [built-in](#dt-built-in) datatypes defined in this specification.
- 2<a id="implement-all-cos"></a>Completely and correctly implement all of the [constraints on schemas](#dt-cos) defined in this specification.
- 3<a id="implement-all-vr"></a>Completely and correctly implement all of the [Validation Rules](#dt-cvc) defined in this specification, when checking the datatype validity of literals against datatypes.
  Implementations claiming **schema-document-aware conformance** to this specification, independent of any host language **must** be minimally conforming. In addition, they must do **all** of the following:

- 1<a id="accept-std"></a>Accept simple type definitions in the form specified in [Datatype components (§4)](#datatype-components).
- 2<a id="implement-all-xrc"></a>Completely and correctly implement all of rules governing the XML representation of simple type definitions specified in [Datatype components (§4)](#datatype-components).
- 3<a id="map-xml-component"></a>Map the XML representations of simple type definitions to simple type definition components as specified in the mapping rules given in [Datatype components (§4)](#datatype-components).
 **Note:** The term **schema-document aware** is used here for parallelism with the corresponding term in [[XSD 1.1 Part 1: Structures]](#structural-schemas). The reference to schema documents may be taken as referring to the fact that schema-document-aware implementations accept the XML representation of simple type definitions found in XSD schema documents. It does *not* mean that the simple type definitions must themselves be free-standing XML documents, nor that they typically will be.

### <a id="data-conformance"></a>5.3 Conformance of data

Abstract representations of simple type definitions conform to this specification if and only if they obey all of the [constraints on schemas](#dt-cos) defined in this specification.

XML representations of simple type definitions conform to this specification if they obey all of the applicable rules defined in this specification.

**Note:** Because the conformance of the resulting simple type definition component depends not only on the XML representation of a given simple type definition, but on the properties of its [base type](#dt-basetype), the conformance of an XML representation of a simple type definition does not guarantee that, in the context of other schema components, it will map to a conforming component.

### <a id="partial-implementation"></a>5.4 Partial Implementation of Infinite Datatypes

Some [primitive](#dt-primitive) datatypes defined in this specification have infinite [value spaces](#dt-value-space); no finite implementation can completely handle all their possible values. For some such datatypes, minimum implementation limits are specified below. For other infinite types such as [string](#string), [hexBinary](#hexBinary), and [base64Binary](#base64Binary), no minimum implementation limits are specified.

When this specification is used in the context of other languages (as it is, for example, by [[XSD 1.1 Part 1: Structures]](#structural-schemas)), the host language may specify other minimum implementation limits.

 When presented with a literal or value exceeding the capacity of its partial implementation of a datatype, a minimally conforming implementation of this specification will sometimes be unable to determine with certainty whether the value is datatype-valid or not. Sometimes it will be unable to represent the value correctly through its interface to any downstream application.

 When either of these is so, a conforming processor **must** indicate to the user and/or downstream application that it cannot process the input data with assured correctness (much as it would indicate if it ran out of memory). When the datatype validity of a value or literal is uncertain because it exceeds the capacity of a partial implementation, the literal or value **must not** be treated as invalid, and the unsupported value **must not** be quietly changed to a supported value.

 This specification does not constrain the method used to indicate that a literal or value in the input data has exceeded the capacity of the implementation, or the form such indications take.

[Minimally conforming](#dt-minimally-conforming) processors which set an application- or [implementation-defined](#key-impl-def) limit on the size of the values supported **must** clearly document that limit.

These are the partial-implementation [minimal conformance](#dt-minimally-conforming) requirements:

- All [minimally conforming](#dt-minimally-conforming) processors **must** support [decimal](#decimal) values whose absolute value can be expressed as *i* / 10^*k*^, where *i* and *k* are nonnegative integers such that *i* < 10^16^ and *k* ≤ 16 (i.e., those expressible with sixteen total digits).
- All [minimally conforming](#dt-minimally-conforming) processors **must** support nonnegative [*year*](#vp-dt-year) values less than 10000 (i.e., those expressible with four digits) in all datatypes which use the seven-property model defined in [The Seven-property Model (§D.2.1)](#theSevenPropertyModel) and have a non-[absent](#key-null) value for [*year*](#vp-dt-year) (i.e. [dateTime](#dateTime), [dateTimeStamp](#dateTimeStamp), [date](#date), [gYearMonth](#gYearMonth), and [gYear](#gYear)). .
- All [minimally conforming](#dt-minimally-conforming) processors **must** support [*second*](#vp-dt-second) values to milliseconds (i.e. those expressible with three fraction digits) in all datatypes which use the seven-property model defined in [The Seven-property Model (§D.2.1)](#theSevenPropertyModel) and have a non-[absent](#key-null) value for [*second*](#vp-dt-second) (i.e. [dateTime](#dateTime), [dateTimeStamp](#dateTimeStamp), and [time](#time)). .
- All [minimally conforming](#dt-minimally-conforming) processors **must** support fractional-second [duration](#duration) values to milliseconds (i.e. those expressible with three fraction digits).
- All [minimally conforming](#dt-minimally-conforming) processors **must** support [duration](#duration) values with [*months*](#vp-du-month) values in the range −119999 to 119999 months (9999 years and 11 months) and [*seconds*](#vp-du-second) values in the range −31622400 to 31622400 seconds (one leap-year).

## <a id="schema"></a>A Schema for Schema Documents (Datatypes) (normative)

The XML representation of the datatypes-relevant part of the schema for schema documents is presented here as a normative part of the specification. Independent copies of this material are available in an undated (mutable) version at [http://www.w3.org/2009/XMLSchema/datatypes.xsd](https://www.w3.org/2009/XMLSchema/datatypes.xsd) and in a dated (immutable) version at [http://www.w3.org/2012/04/datatypes.xsd](https://www.w3.org/2012/04/datatypes.xsd) — the mutable version will be updated with future revisions of this specification, and the immutable one will not.

Like any other XML document, schema documents may carry XML and document type declarations. An XML declaration and a document type declaration are provided here for convenience. Since this schema document describes the XML Schema language, the `targetNamespace` attribute on the `schema` element refers to the XML Schema namespace itself.

 Schema documents conforming to this specification may be in XML 1.0 or XML 1.1. Conforming implementations may accept input in XML 1.0 or XML 1.1 or both. See [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

 <a id="schema-for-datatypes"></a> Schema for Schema Documents (Datatypes)

```
<?xml version='1.0'?>
<!DOCTYPE xs:schema PUBLIC "-//W3C//DTD XSD 1.1//EN" "XMLSchema.dtd" [

<!--
        Make sure that processors that do not read the external
        subset will know about the various IDs we declare
  -->
        <!ATTLIST xs:simpleType id ID #IMPLIED>
        <!ATTLIST xs:maxExclusive id ID #IMPLIED>
        <!ATTLIST xs:minExclusive id ID #IMPLIED>
        <!ATTLIST xs:maxInclusive id ID #IMPLIED>
        <!ATTLIST xs:minInclusive id ID #IMPLIED>
        <!ATTLIST xs:totalDigits id ID #IMPLIED>
        <!ATTLIST xs:fractionDigits id ID #IMPLIED>
        <!ATTLIST xs:length id ID #IMPLIED>
        <!ATTLIST xs:minLength id ID #IMPLIED>
        <!ATTLIST xs:maxLength id ID #IMPLIED>
        <!ATTLIST xs:enumeration id ID #IMPLIED>
        <!ATTLIST xs:pattern id ID #IMPLIED>
        <!ATTLIST xs:assertion id ID #IMPLIED>
        <!ATTLIST xs:explicitTimezone id ID #IMPLIED>
        <!ATTLIST xs:appinfo id ID #IMPLIED>
        <!ATTLIST xs:documentation id ID #IMPLIED>
        <!ATTLIST xs:list id ID #IMPLIED>
        <!ATTLIST xs:union id ID #IMPLIED>
        ]>

<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           elementFormDefault="qualified"
           xml:lang="en"
           targetNamespace="http://www.w3.org/2001/XMLSchema"
           version="datatypes.xsd (rec-20120405)">
  <xs:annotation>
    <xs:documentation source="../datatypes/datatypes.html">
      The schema corresponding to this document is normative,
      with respect to the syntactic constraints it expresses in the
      XML Schema language.  The documentation (within 'documentation'
      elements) below, is not normative, but rather highlights important
      aspects of the W3C Recommendation of which this is a part.

      See below (at the bottom of this document) for information about
      the revision and namespace-versioning policy governing this
      schema document.
    </xs:documentation>
  </xs:annotation>

  <xs:simpleType name="derivationControl">
    <xs:annotation>
      <xs:documentation>
   A utility type, not for public use</xs:documentation>
    </xs:annotation>
    <xs:restriction base="xs:NMTOKEN">
      <xs:enumeration value="substitution"/>
      <xs:enumeration value="extension"/>
      <xs:enumeration value="restriction"/>
      <xs:enumeration value="list"/>
      <xs:enumeration value="union"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:group name="simpleDerivation">
    <xs:choice>
      <xs:element ref="xs:restriction"/>
      <xs:element ref="xs:list"/>
      <xs:element ref="xs:union"/>
    </xs:choice>
  </xs:group>
  <xs:simpleType name="simpleDerivationSet">
    <xs:annotation>
      <xs:documentation>
   #all or (possibly empty) subset of {restriction, extension, union, list}
   </xs:documentation>
      <xs:documentation>
   A utility type, not for public use</xs:documentation>
    </xs:annotation>
    <xs:union>
      <xs:simpleType>
        <xs:restriction base="xs:token">
          <xs:enumeration value="#all"/>
        </xs:restriction>
      </xs:simpleType>
      <xs:simpleType>
        <xs:list>
          <xs:simpleType>
            <xs:restriction base="xs:derivationControl">
              <xs:enumeration value="list"/>
              <xs:enumeration value="union"/>
              <xs:enumeration value="restriction"/>
              <xs:enumeration value="extension"/>
            </xs:restriction>
          </xs:simpleType>
        </xs:list>
      </xs:simpleType>
    </xs:union>
  </xs:simpleType>
  <xs:complexType name="simpleType" abstract="true">
    <xs:complexContent>
      <xs:extension base="xs:annotated">
        <xs:group ref="xs:simpleDerivation"/>
        <xs:attribute name="final" type="xs:simpleDerivationSet"/>
        <xs:attribute name="name" type="xs:NCName">
          <xs:annotation>
            <xs:documentation>
              Can be restricted to required or forbidden
            </xs:documentation>
          </xs:annotation>
        </xs:attribute>
      </xs:extension>
    </xs:complexContent>
  </xs:complexType>
  <xs:complexType name="topLevelSimpleType">
    <xs:complexContent>
      <xs:restriction base="xs:simpleType">
        <xs:sequence>
          <xs:element ref="xs:annotation" minOccurs="0"/>
          <xs:group ref="xs:simpleDerivation"/>
        </xs:sequence>
        <xs:attribute name="name" type="xs:NCName" use="required">
          <xs:annotation>
            <xs:documentation>
              Required at the top level
            </xs:documentation>
          </xs:annotation>
        </xs:attribute>
        <xs:anyAttribute namespace="##other" processContents="lax"/>
      </xs:restriction>
    </xs:complexContent>
  </xs:complexType>
  <xs:complexType name="localSimpleType">
    <xs:complexContent>
      <xs:restriction base="xs:simpleType">
        <xs:sequence>
          <xs:element ref="xs:annotation" minOccurs="0"/>
          <xs:group ref="xs:simpleDerivation"/>
        </xs:sequence>
        <xs:attribute name="name" use="prohibited">
          <xs:annotation>
            <xs:documentation>
              Forbidden when nested
            </xs:documentation>
          </xs:annotation>
        </xs:attribute>
        <xs:attribute name="final" use="prohibited"/>
        <xs:anyAttribute namespace="##other" processContents="lax"/>
      </xs:restriction>
    </xs:complexContent>
  </xs:complexType>
  <xs:element name="simpleType" type="xs:topLevelSimpleType" id="simpleType">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-simpleType"/>
    </xs:annotation>
  </xs:element>
  <xs:element name="facet" abstract="true">
    <xs:annotation>
      <xs:documentation>
        An abstract element, representing facets in general.
        The facets defined by this spec are substitutable for
        this element, and implementation-defined facets should
        also name this as a substitution-group head.
      </xs:documentation>
    </xs:annotation>
  </xs:element>
  <xs:group name="simpleRestrictionModel">
    <xs:sequence>
      <xs:element name="simpleType" type="xs:localSimpleType" minOccurs="0"/>
      <xs:choice minOccurs="0"
          maxOccurs="unbounded">
        <xs:element ref="xs:facet"/>
        <xs:any processContents="lax"
            namespace="##other"/>
      </xs:choice>
    </xs:sequence>
  </xs:group>
  <xs:element name="restriction" id="restriction">
    <xs:complexType>
      <xs:annotation>
        <xs:documentation
             source="http://www.w3.org/TR/xmlschema11-2/#element-restriction">
          base attribute and simpleType child are mutually
          exclusive, but one or other is required
        </xs:documentation>
      </xs:annotation>
      <xs:complexContent>
        <xs:extension base="xs:annotated">
          <xs:group ref="xs:simpleRestrictionModel"/>
          <xs:attribute name="base" type="xs:QName" use="optional"/>
        </xs:extension>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
  <xs:element name="list" id="list">
    <xs:complexType>
      <xs:annotation>
        <xs:documentation
             source="http://www.w3.org/TR/xmlschema11-2/#element-list">
          itemType attribute and simpleType child are mutually
          exclusive, but one or other is required
        </xs:documentation>
      </xs:annotation>
      <xs:complexContent>
        <xs:extension base="xs:annotated">
          <xs:sequence>
            <xs:element name="simpleType" type="xs:localSimpleType"
                        minOccurs="0"/>
          </xs:sequence>
          <xs:attribute name="itemType" type="xs:QName" use="optional"/>
        </xs:extension>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
  <xs:element name="union" id="union">
    <xs:complexType>
      <xs:annotation>
        <xs:documentation
             source="http://www.w3.org/TR/xmlschema11-2/#element-union">
          memberTypes attribute must be non-empty or there must be
          at least one simpleType child
        </xs:documentation>
      </xs:annotation>
      <xs:complexContent>
        <xs:extension base="xs:annotated">
          <xs:sequence>
            <xs:element name="simpleType" type="xs:localSimpleType"
                        minOccurs="0" maxOccurs="unbounded"/>
          </xs:sequence>
          <xs:attribute name="memberTypes" use="optional">
            <xs:simpleType>
              <xs:list itemType="xs:QName"/>
            </xs:simpleType>
          </xs:attribute>
        </xs:extension>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
  <xs:complexType name="facet">
    <xs:complexContent>
      <xs:extension base="xs:annotated">
        <xs:attribute name="value" use="required"/>
        <xs:attribute name="fixed" type="xs:boolean" default="false"
                      use="optional"/>
      </xs:extension>
    </xs:complexContent>
  </xs:complexType>
  <xs:complexType name="noFixedFacet">
    <xs:complexContent>
      <xs:restriction base="xs:facet">
        <xs:sequence>
          <xs:element ref="xs:annotation" minOccurs="0"/>
        </xs:sequence>
        <xs:attribute name="fixed" use="prohibited"/>
        <xs:anyAttribute namespace="##other" processContents="lax"/>
      </xs:restriction>
    </xs:complexContent>
  </xs:complexType>
  <xs:element name="minExclusive" type="xs:facet"
    id="minExclusive"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-minExclusive"/>
    </xs:annotation>
  </xs:element>
  <xs:element name="minInclusive" type="xs:facet"
    id="minInclusive"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-minInclusive"/>
    </xs:annotation>
  </xs:element>
  <xs:element name="maxExclusive" type="xs:facet"
    id="maxExclusive"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-maxExclusive"/>
    </xs:annotation>
  </xs:element>
  <xs:element name="maxInclusive" type="xs:facet"
    id="maxInclusive"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-maxInclusive"/>
    </xs:annotation>
  </xs:element>
  <xs:complexType name="numFacet">
    <xs:complexContent>
      <xs:restriction base="xs:facet">
        <xs:sequence>
          <xs:element ref="xs:annotation" minOccurs="0"/>
        </xs:sequence>
        <xs:attribute name="value"
            type="xs:nonNegativeInteger" use="required"/>
        <xs:anyAttribute namespace="##other" processContents="lax"/>
      </xs:restriction>
    </xs:complexContent>
  </xs:complexType>

  <xs:complexType name="intFacet">
    <xs:complexContent>
      <xs:restriction base="xs:facet">
        <xs:sequence>
          <xs:element ref="xs:annotation" minOccurs="0"/>
        </xs:sequence>
        <xs:attribute name="value" type="xs:integer" use="required"/>
        <xs:anyAttribute namespace="##other" processContents="lax"/>
      </xs:restriction>
    </xs:complexContent>
  </xs:complexType>

  <xs:element name="totalDigits" id="totalDigits"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-totalDigits"/>
    </xs:annotation>
    <xs:complexType>
      <xs:complexContent>
        <xs:restriction base="xs:numFacet">
          <xs:sequence>
            <xs:element ref="xs:annotation" minOccurs="0"/>
          </xs:sequence>
          <xs:attribute name="value" type="xs:positiveInteger" use="required"/>
          <xs:anyAttribute namespace="##other" processContents="lax"/>
        </xs:restriction>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
  <xs:element name="fractionDigits" type="xs:numFacet"
    id="fractionDigits"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-fractionDigits"/>
    </xs:annotation>
  </xs:element>

  <xs:element name="length" type="xs:numFacet" id="length"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-length"/>
    </xs:annotation>
  </xs:element>
  <xs:element name="minLength" type="xs:numFacet"
    id="minLength"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-minLength"/>
    </xs:annotation>
  </xs:element>
  <xs:element name="maxLength" type="xs:numFacet"
    id="maxLength"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-maxLength"/>
    </xs:annotation>
  </xs:element>
  <xs:element name="enumeration" type="xs:noFixedFacet"
    id="enumeration"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-enumeration"/>
    </xs:annotation>
  </xs:element>
  <xs:element name="whiteSpace" id="whiteSpace"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-whiteSpace"/>
    </xs:annotation>
    <xs:complexType>
      <xs:complexContent>
        <xs:restriction base="xs:facet">
          <xs:sequence>
            <xs:element ref="xs:annotation" minOccurs="0"/>
          </xs:sequence>
          <xs:attribute name="value" use="required">
            <xs:simpleType>
              <xs:restriction base="xs:NMTOKEN">
                <xs:enumeration value="preserve"/>
                <xs:enumeration value="replace"/>
                <xs:enumeration value="collapse"/>
              </xs:restriction>
            </xs:simpleType>
          </xs:attribute>
          <xs:anyAttribute namespace="##other" processContents="lax"/>
        </xs:restriction>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
  <xs:element name="pattern" id="pattern"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-pattern"/>
    </xs:annotation>
    <xs:complexType>
      <xs:complexContent>
        <xs:restriction base="xs:noFixedFacet">
          <xs:sequence>
            <xs:element ref="xs:annotation" minOccurs="0"/>
          </xs:sequence>
          <xs:attribute name="value" type="xs:string"
              use="required"/>
          <xs:anyAttribute namespace="##other"
              processContents="lax"/>
        </xs:restriction>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>
  <xs:element name="assertion" type="xs:assertion"
              id="assertion" substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-assertion"/>
    </xs:annotation>
  </xs:element>
  <xs:element name="explicitTimezone" id="explicitTimezone"
    substitutionGroup="xs:facet">
    <xs:annotation>
      <xs:documentation
           source="http://www.w3.org/TR/xmlschema11-2/#element-explicitTimezone"/>
    </xs:annotation>
    <xs:complexType>
      <xs:complexContent>
        <xs:restriction base="xs:facet">
          <xs:sequence>
            <xs:element ref="xs:annotation" minOccurs="0"/>
          </xs:sequence>
          <xs:attribute name="value" use="required">
            <xs:simpleType>
              <xs:restriction base="xs:NMTOKEN">
                <xs:enumeration value="optional"/>
                <xs:enumeration value="required"/>
                <xs:enumeration value="prohibited"/>
              </xs:restriction>
            </xs:simpleType>
          </xs:attribute>
          <xs:anyAttribute namespace="##other" processContents="lax"/>
        </xs:restriction>
      </xs:complexContent>
    </xs:complexType>
  </xs:element>

  <xs:annotation>
    <xs:documentation>
      In keeping with the XML Schema WG's standard versioning policy,
      this schema document will persist at the URI
      http://www.w3.org/2012/04/datatypes.xsd.

      At the date of issue it can also be found at the URI
      http://www.w3.org/2009/XMLSchema/datatypes.xsd.

      The schema document at that URI may however change in the future,
      in order to remain compatible with the latest version of XSD
      and its namespace.  In other words, if XSD or the XML Schema
      namespace change, the version of this document at
      http://www.w3.org/2009/XMLSchema/datatypes.xsd will change accordingly;
      the version at http://www.w3.org/2012/04/datatypes.xsd will not change.

      Previous dated (and unchanging) versions of this schema document
      include:

        http://www.w3.org/2012/01/datatypes.xsd
          (XSD 1.1 Proposed Recommendation)

        http://www.w3.org/2011/07/datatypes.xsd
          (XSD 1.1 Candidate Recommendation)

        http://www.w3.org/2009/04/datatypes.xsd
          (XSD 1.1 Candidate Recommendation)

        http://www.w3.org/2004/10/datatypes.xsd
          (XSD 1.0 Recommendation, Second Edition)

        http://www.w3.org/2001/05/datatypes.xsd
          (XSD 1.0 Recommendation, First Edition)

    </xs:documentation>
  </xs:annotation>

</xs:schema>

```

## <a id="dtd-for-datatypeDefs"></a>B DTD for Datatype Definitions (non-normative)

The DTD for the datatypes-specific aspects of schema documents is given below. Note there is *no* implication here that `schema` **must** be the root element of a document.

 <a id="dtd-for-datatypes"></a> DTD for datatype definitions

```
<!--
        DTD for XML Schemas: Part 2: Datatypes

        Id: datatypes.dtd,v 1.1.2.4 2005/01/31 18:40:42 cmsmcq Exp
        Note this DTD is NOT normative, or even definitive.
  -->

<!--
        This DTD cannot be used on its own, it is intended
        only for incorporation in XMLSchema.dtd, q.v.
  -->

<!-- Define all the element names, with optional prefix -->
<!ENTITY % simpleType "%p;simpleType">
<!ENTITY % restriction "%p;restriction">
<!ENTITY % list "%p;list">
<!ENTITY % union "%p;union">
<!ENTITY % maxExclusive "%p;maxExclusive">
<!ENTITY % minExclusive "%p;minExclusive">
<!ENTITY % maxInclusive "%p;maxInclusive">
<!ENTITY % minInclusive "%p;minInclusive">
<!ENTITY % totalDigits "%p;totalDigits">
<!ENTITY % fractionDigits "%p;fractionDigits">

<!ENTITY % length "%p;length">
<!ENTITY % minLength "%p;minLength">
<!ENTITY % maxLength "%p;maxLength">
<!ENTITY % enumeration "%p;enumeration">
<!ENTITY % whiteSpace "%p;whiteSpace">
<!ENTITY % pattern "%p;pattern">

<!ENTITY % assertion "%p;assertion">

<!ENTITY % explicitTimezone "%p;explicitTimezone">

<!--
        Customization entities for the ATTLIST of each element
        type. Define one of these if your schema takes advantage
        of the anyAttribute='##other' in the schema for schemas
  -->

<!ENTITY % simpleTypeAttrs "">
<!ENTITY % restrictionAttrs "">
<!ENTITY % listAttrs "">
<!ENTITY % unionAttrs "">
<!ENTITY % maxExclusiveAttrs "">
<!ENTITY % minExclusiveAttrs "">
<!ENTITY % maxInclusiveAttrs "">
<!ENTITY % minInclusiveAttrs "">
<!ENTITY % totalDigitsAttrs "">
<!ENTITY % fractionDigitsAttrs "">
<!ENTITY % lengthAttrs "">
<!ENTITY % minLengthAttrs "">
<!ENTITY % maxLengthAttrs "">

<!ENTITY % enumerationAttrs "">
<!ENTITY % whiteSpaceAttrs "">
<!ENTITY % patternAttrs "">
<!ENTITY % assertionAttrs "">
<!ENTITY % explicitTimezoneAttrs "">

<!-- Define some entities for informative use as attribute
        types -->
<!ENTITY % URIref "CDATA">
<!ENTITY % XPathExpr "CDATA">
<!ENTITY % QName "NMTOKEN">
<!ENTITY % QNames "NMTOKENS">
<!ENTITY % NCName "NMTOKEN">
<!ENTITY % nonNegativeInteger "NMTOKEN">
<!ENTITY % boolean "(true|false)">
<!ENTITY % simpleDerivationSet "CDATA">
<!--
        #all or space-separated list drawn from derivationChoice
  -->

<!--
        Note that the use of 'facet' below is less restrictive
        than is really intended:  There should in fact be no
        more than one of each of minInclusive, minExclusive,
        maxInclusive, maxExclusive, totalDigits, fractionDigits,
        length, maxLength, minLength within datatype,
        and the min- and max- variants of Inclusive and Exclusive
        are mutually exclusive. On the other hand,  pattern and
        enumeration and assertion may repeat.
  -->
<!ENTITY % minBound "(%minInclusive; | %minExclusive;)">
<!ENTITY % maxBound "(%maxInclusive; | %maxExclusive;)">
<!ENTITY % bounds "%minBound; | %maxBound;">
<!ENTITY % numeric "%totalDigits; | %fractionDigits;">
<!ENTITY % ordered "%bounds; | %numeric;">
<!ENTITY % unordered
   "%pattern; | %enumeration; | %whiteSpace; | %length; |
   %maxLength; | %minLength; | %assertion;
   | %explicitTimezone;">
<!ENTITY % implementation-defined-facets "">
<!ENTITY % facet "%ordered; | %unordered; %implementation-defined-facets;">
<!ENTITY % facetAttr
        "value CDATA #REQUIRED
        id ID #IMPLIED">
<!ENTITY % fixedAttr "fixed %boolean; #IMPLIED">
<!ENTITY % facetModel "(%annotation;)?">
<!ELEMENT %simpleType;
        ((%annotation;)?, (%restriction; | %list; | %union;))>
<!ATTLIST %simpleType;
    name      %NCName; #IMPLIED
    final     %simpleDerivationSet; #IMPLIED
    id        ID       #IMPLIED
    %simpleTypeAttrs;>
<!-- name is required at top level -->
<!ELEMENT %restriction; ((%annotation;)?,
                         (%restriction1; |
                          ((%simpleType;)?,(%facet;)*)),
                         (%attrDecls;))>
<!ATTLIST %restriction;
    base      %QName;                  #IMPLIED
    id        ID       #IMPLIED
    %restrictionAttrs;>
<!--
        base and simpleType child are mutually exclusive,
        one is required.

        restriction is shared between simpleType and
        simpleContent and complexContent (in XMLSchema.xsd).
        restriction1 is for the latter cases, when this
        is restricting a complex type, as is attrDecls.
  -->
<!ELEMENT %list; ((%annotation;)?,(%simpleType;)?)>
<!ATTLIST %list;
    itemType      %QName;             #IMPLIED
    id        ID       #IMPLIED
    %listAttrs;>
<!--
        itemType and simpleType child are mutually exclusive,
        one is required
  -->
<!ELEMENT %union; ((%annotation;)?,(%simpleType;)*)>
<!ATTLIST %union;
    id            ID       #IMPLIED
    memberTypes   %QNames;            #IMPLIED
    %unionAttrs;>
<!--
        At least one item in memberTypes or one simpleType
        child is required
  -->

<!ELEMENT %maxExclusive; %facetModel;>
<!ATTLIST %maxExclusive;
        %facetAttr;
        %fixedAttr;
        %maxExclusiveAttrs;>
<!ELEMENT %minExclusive; %facetModel;>
<!ATTLIST %minExclusive;
        %facetAttr;
        %fixedAttr;
        %minExclusiveAttrs;>

<!ELEMENT %maxInclusive; %facetModel;>
<!ATTLIST %maxInclusive;
        %facetAttr;
        %fixedAttr;
        %maxInclusiveAttrs;>
<!ELEMENT %minInclusive; %facetModel;>
<!ATTLIST %minInclusive;
        %facetAttr;
        %fixedAttr;
        %minInclusiveAttrs;>

<!ELEMENT %totalDigits; %facetModel;>
<!ATTLIST %totalDigits;
        %facetAttr;
        %fixedAttr;
        %totalDigitsAttrs;>
<!ELEMENT %fractionDigits; %facetModel;>
<!ATTLIST %fractionDigits;
        %facetAttr;
        %fixedAttr;
        %fractionDigitsAttrs;>

<!ELEMENT %length; %facetModel;>
<!ATTLIST %length;
        %facetAttr;
        %fixedAttr;
        %lengthAttrs;>
<!ELEMENT %minLength; %facetModel;>
<!ATTLIST %minLength;
        %facetAttr;
        %fixedAttr;
        %minLengthAttrs;>
<!ELEMENT %maxLength; %facetModel;>
<!ATTLIST %maxLength;
        %facetAttr;
        %fixedAttr;
        %maxLengthAttrs;>

<!-- This one can be repeated -->
<!ELEMENT %enumeration; %facetModel;>
<!ATTLIST %enumeration;
        %facetAttr;
        %enumerationAttrs;>

<!ELEMENT %whiteSpace; %facetModel;>
<!ATTLIST %whiteSpace;
        %facetAttr;
        %fixedAttr;
        %whiteSpaceAttrs;>

<!-- This one can be repeated -->
<!ELEMENT %pattern; %facetModel;>
<!ATTLIST %pattern;
        %facetAttr;
        %patternAttrs;>

<!ELEMENT %assertion; %facetModel;>
<!ATTLIST %assertion;
        %facetAttr;
        %assertionAttrs;>

<!ELEMENT %explicitTimezone; %facetModel;>
<!ATTLIST %explicitTimezone;
        %facetAttr;
        %explicitTimezoneAttrs;>

```

## <a id="prim.nxsd"></a>C Illustrative XML representations for the built-in simple type definitions

### <a id="sec-prim-nxsd"></a>C.1 Illustrative XML representations for the built-in primitive type definitions

The following, although in the form of a schema document, does not conform to the rules for schema documents defined in this specification. It contains explicit XML representations of the primitive datatypes which need not be declared in a schema document, since they are automatically included in every schema, and indeed must not be declared in a schema document, since it is forbidden to try to derive types with [anyAtomicType](#anyAtomicType) as the base type definition. It is included here as a form of documentation.

 <a id="not-schema-for-primitives"></a> The (not a) schema document for primitive built-in type definitions

```
<?xml version='1.0'?>
<!DOCTYPE xs:schema SYSTEM "../namespace/XMLSchema.dtd" [

<!--
     keep this schema XML1.0 DTD valid
  -->
        <!ENTITY % schemaAttrs 'xmlns:hfp CDATA #IMPLIED'>

        <!ELEMENT hfp:hasFacet EMPTY>
        <!ATTLIST hfp:hasFacet
                name NMTOKEN #REQUIRED>

        <!ELEMENT hfp:hasProperty EMPTY>
        <!ATTLIST hfp:hasProperty
                name NMTOKEN #REQUIRED
                value CDATA #REQUIRED>
]>
<xs:schema
  xmlns:hfp="http://www.w3.org/2001/XMLSchema-hasFacetAndProperty"
  xmlns:xs="http://www.w3.org/2001/XMLSchema"
  elementFormDefault="qualified"
  xml:lang="en"
  targetNamespace="http://www.w3.org/2001/XMLSchema">

  <xs:annotation>
    <xs:documentation>
      This document contains XML elements which look like
      definitions for the primitive datatypes.  These definitions are for
      information only; the real built-in definitions are magic.
    </xs:documentation>
    <xs:documentation>
      For each built-in datatype in this schema (both primitive and
      derived) can be uniquely addressed via a URI constructed
      as follows:
        1) the base URI is the URI of the XML Schema namespace
        2) the fragment identifier is the name of the datatype

      For example, to address the int datatype, the URI is:

        http://www.w3.org/2001/XMLSchema#int

      Additionally, each facet definition element can be uniquely
      addressed via a URI constructed as follows:
        1) the base URI is the URI of the XML Schema namespace
        2) the fragment identifier is the name of the facet

      For example, to address the maxInclusive facet, the URI is:

        http://www.w3.org/2001/XMLSchema#maxInclusive

      Additionally, each facet usage in a built-in datatype definition
      can be uniquely addressed via a URI constructed as follows:
        1) the base URI is the URI of the XML Schema namespace
        2) the fragment identifier is the name of the datatype, followed
           by a period (".") followed by the name of the facet

      For example, to address the usage of the maxInclusive facet in
      the definition of int, the URI is:

        http://www.w3.org/2001/XMLSchema#int.maxInclusive

    </xs:documentation>
  </xs:annotation>
  <xs:simpleType name="string" id="string">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="length"/>
        <hfp:hasFacet name="minLength"/>
        <hfp:hasFacet name="maxLength"/>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="false"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#string"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace value="preserve" id="string.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="boolean" id="boolean">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="false"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="finite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#boolean"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="boolean.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="float" id="float">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="partial"/>
        <hfp:hasProperty name="bounded" value="true"/>
        <hfp:hasProperty name="cardinality" value="finite"/>
        <hfp:hasProperty name="numeric" value="true"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#float"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="float.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="double" id="double">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="partial"/>
        <hfp:hasProperty name="bounded" value="true"/>
        <hfp:hasProperty name="cardinality" value="finite"/>
        <hfp:hasProperty name="numeric" value="true"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#double"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="double.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="decimal" id="decimal">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="totalDigits"/>
        <hfp:hasFacet name="fractionDigits"/>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="total"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="true"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#decimal"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="decimal.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>

  <xs:simpleType name="duration" id="duration">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="partial"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#duration"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="duration.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="dateTime" id="dateTime">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasFacet name="explicitTimezone"/>
        <hfp:hasProperty name="ordered" value="partial"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#dateTime"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="dateTime.whiteSpace"/>
      <xs:explicitTimezone value="optional" id="dateTime.explicitTimezone"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="time" id="time">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasFacet name="explicitTimezone"/>
        <hfp:hasProperty name="ordered" value="partial"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#time"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="time.whiteSpace"/>
      <xs:explicitTimezone value="optional" id="time.explicitTimezone"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="date" id="date">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasFacet name="explicitTimezone"/>
        <hfp:hasProperty name="ordered" value="partial"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#date"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="date.whiteSpace"/>
      <xs:explicitTimezone value="optional" id="date.explicitTimezone"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="gYearMonth" id="gYearMonth">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasFacet name="explicitTimezone"/>
        <hfp:hasProperty name="ordered" value="partial"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#gYearMonth"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="gYearMonth.whiteSpace"/>
      <xs:explicitTimezone value="optional" id="gYearMonth.explicitTimezone"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="gYear" id="gYear">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasFacet name="explicitTimezone"/>
        <hfp:hasProperty name="ordered" value="partial"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#gYear"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="gYear.whiteSpace"/>
      <xs:explicitTimezone value="optional" id="gYear.explicitTimezone"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="gMonthDay" id="gMonthDay">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasFacet name="explicitTimezone"/>
        <hfp:hasProperty name="ordered" value="partial"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#gMonthDay"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="gMonthDay.whiteSpace"/>
      <xs:explicitTimezone value="optional" id="gMonthDay.explicitTimezone"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="gDay" id="gDay">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasFacet name="explicitTimezone"/>
        <hfp:hasProperty name="ordered" value="partial"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#gDay"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="gDay.whiteSpace"/>
      <xs:explicitTimezone value="optional" id="gDay.explicitTimezone"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="gMonth" id="gMonth">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="maxInclusive"/>
        <hfp:hasFacet name="maxExclusive"/>
        <hfp:hasFacet name="minInclusive"/>
        <hfp:hasFacet name="minExclusive"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasFacet name="explicitTimezone"/>
        <hfp:hasProperty name="ordered" value="partial"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#gMonth"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="gMonth.whiteSpace"/>
      <xs:explicitTimezone value="optional" id="gMonth.explicitTimezone"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="hexBinary" id="hexBinary">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="length"/>
        <hfp:hasFacet name="minLength"/>
        <hfp:hasFacet name="maxLength"/>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="false"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#hexBinary"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="hexBinary.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="base64Binary" id="base64Binary">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="length"/>
        <hfp:hasFacet name="minLength"/>
        <hfp:hasFacet name="maxLength"/>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="false"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#base64Binary"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="base64Binary.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="anyURI" id="anyURI">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="length"/>
        <hfp:hasFacet name="minLength"/>
        <hfp:hasFacet name="maxLength"/>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="false"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#anyURI"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="anyURI.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="QName" id="QName">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="length"/>
        <hfp:hasFacet name="minLength"/>
        <hfp:hasFacet name="maxLength"/>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="false"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#QName"/>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="QName.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="NOTATION" id="NOTATION">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="length"/>
        <hfp:hasFacet name="minLength"/>
        <hfp:hasFacet name="maxLength"/>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="false"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#NOTATION"/>
      <xs:documentation>
        NOTATION cannot be used directly in a schema; rather a type
        must be derived from it by specifying at least one enumeration
        facet whose value is the name of a NOTATION declared in the
        schema.
      </xs:documentation>
    </xs:annotation>
    <xs:restriction base="xs:anyAtomicType">
      <xs:whiteSpace fixed="true" value="collapse" id="NOTATION.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
</xs:schema>

```

### <a id="drvd.nxsd"></a>C.2 Illustrative XML representations for the built-in ordinary type definitions

The following, although in the form of a schema document, contains XML representations of components already present in all schemas by definition. It is included here as a form of documentation.

**Note:** These datatypes do not need to be declared in a schema document, since they are automatically included in every schema.<a id="B-1933"></a>

> **Issue (B-1933):**It is an open question whether this and similar XML documents should be accepted or rejected by software conforming to this specification. The XML Schema Working Group expects to resolve this question in connection with its work on issues relating to schema composition.In the meantime, some existing schema processors will accept declarations for them; other existing processors will reject such declarations as duplicates.

 <a id="schema-for-derived"></a> Illustrative schema document for derived built-in type definitions

```
<?xml version='1.0'?>
<!DOCTYPE xs:schema SYSTEM "../namespace/XMLSchema.dtd" [

<!--
     keep this schema XML1.0 DTD valid
  -->
        <!ENTITY % schemaAttrs 'xmlns:hfp CDATA #IMPLIED'>

        <!ELEMENT hfp:hasFacet EMPTY>
        <!ATTLIST hfp:hasFacet
                name NMTOKEN #REQUIRED>

        <!ELEMENT hfp:hasProperty EMPTY>
        <!ATTLIST hfp:hasProperty
                name NMTOKEN #REQUIRED
                value CDATA #REQUIRED>

]>
<xs:schema
  xmlns:hfp="http://www.w3.org/2001/XMLSchema-hasFacetAndProperty"
  xmlns:xs="http://www.w3.org/2001/XMLSchema"
  elementFormDefault="qualified"
  xml:lang="en"
  targetNamespace="http://www.w3.org/2001/XMLSchema">
 <xs:annotation>
    <xs:documentation>
      This document contains XML representations for the
     ordinary non-primitive built-in datatypes
    </xs:documentation>
  </xs:annotation>
  <xs:simpleType name="normalizedString" id="normalizedString">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#normalizedString"/>
    </xs:annotation>
    <xs:restriction base="xs:string">
      <xs:whiteSpace value="replace" id="normalizedString.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="token" id="token">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#token"/>
    </xs:annotation>
    <xs:restriction base="xs:normalizedString">
      <xs:whiteSpace value="collapse" id="token.whiteSpace"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="language" id="language">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#language"/>
    </xs:annotation>
    <xs:restriction base="xs:token">
      <xs:pattern value="[a-zA-Z]{1,8}(-[a-zA-Z0-9]{1,8})*" id="language.pattern">
        <xs:annotation>
          <xs:documentation source="http://www.ietf.org/rfc/bcp/bcp47.txt">
            pattern specifies the content of section 2.12 of XML 1.0e2
            and RFC 3066 (Revised version of RFC 1766).  N.B. RFC 3066 is now
            obsolete; the grammar of RFC4646 is more restrictive.  So strict
            conformance to the rules for language codes requires extra checking
            beyond validation against this type.
          </xs:documentation>
        </xs:annotation>
      </xs:pattern>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="IDREFS" id="IDREFS">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="length"/>
        <hfp:hasFacet name="minLength"/>
        <hfp:hasFacet name="maxLength"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="false"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#IDREFS"/>
    </xs:annotation>
    <xs:restriction>
      <xs:simpleType>
        <xs:list itemType="xs:IDREF"/>
      </xs:simpleType>
      <xs:minLength value="1" id="IDREFS.minLength"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="ENTITIES" id="ENTITIES">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="length"/>
        <hfp:hasFacet name="minLength"/>
        <hfp:hasFacet name="maxLength"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="false"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#ENTITIES"/>
    </xs:annotation>
    <xs:restriction>
      <xs:simpleType>
        <xs:list itemType="xs:ENTITY"/>
      </xs:simpleType>
      <xs:minLength value="1" id="ENTITIES.minLength"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="NMTOKEN" id="NMTOKEN">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#NMTOKEN"/>
    </xs:annotation>
    <xs:restriction base="xs:token">
      <xs:pattern value="\c+" id="NMTOKEN.pattern">
        <xs:annotation>
          <xs:documentation source="http://www.w3.org/TR/REC-xml#NT-Nmtoken">
            pattern matches production 7 from the XML spec
          </xs:documentation>
        </xs:annotation>
      </xs:pattern>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="NMTOKENS" id="NMTOKENS">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasFacet name="length"/>
        <hfp:hasFacet name="minLength"/>
        <hfp:hasFacet name="maxLength"/>
        <hfp:hasFacet name="enumeration"/>
        <hfp:hasFacet name="whiteSpace"/>
        <hfp:hasFacet name="pattern"/>
        <hfp:hasFacet name="assertions"/>
        <hfp:hasProperty name="ordered" value="false"/>
        <hfp:hasProperty name="bounded" value="false"/>
        <hfp:hasProperty name="cardinality" value="countably infinite"/>
        <hfp:hasProperty name="numeric" value="false"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#NMTOKENS"/>
    </xs:annotation>
    <xs:restriction>
      <xs:simpleType>
        <xs:list itemType="xs:NMTOKEN"/>
      </xs:simpleType>
      <xs:minLength value="1" id="NMTOKENS.minLength"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="Name" id="Name">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#Name"/>
    </xs:annotation>
    <xs:restriction base="xs:token">
      <xs:pattern value="\i\c*" id="Name.pattern">
        <xs:annotation>
          <xs:documentation source="http://www.w3.org/TR/REC-xml#NT-Name">
            pattern matches production 5 from the XML spec
          </xs:documentation>
        </xs:annotation>
      </xs:pattern>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="NCName" id="NCName">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#NCName"/>
    </xs:annotation>
    <xs:restriction base="xs:Name">
      <xs:pattern value="[\i-[:]][\c-[:]]*" id="NCName.pattern">
        <xs:annotation>
          <xs:documentation source="http://www.w3.org/TR/REC-xml-names/#NT-NCName">
            pattern matches production 4 from the Namespaces in XML spec
          </xs:documentation>
        </xs:annotation>
      </xs:pattern>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="ID" id="ID">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#ID"/>
    </xs:annotation>
    <xs:restriction base="xs:NCName"/>
  </xs:simpleType>
  <xs:simpleType name="IDREF" id="IDREF">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#IDREF"/>
    </xs:annotation>
    <xs:restriction base="xs:NCName"/>
  </xs:simpleType>
  <xs:simpleType name="ENTITY" id="ENTITY">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#ENTITY"/>
    </xs:annotation>
    <xs:restriction base="xs:NCName"/>
  </xs:simpleType>
  <xs:simpleType name="integer" id="integer">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#integer"/>
    </xs:annotation>
    <xs:restriction base="xs:decimal">
      <xs:fractionDigits fixed="true" value="0" id="integer.fractionDigits"/>
      <xs:pattern value="[\-+]?[0-9]+" id="integer.pattern"/>

    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="nonPositiveInteger" id="nonPositiveInteger">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#nonPositiveInteger"/>
    </xs:annotation>
    <xs:restriction base="xs:integer">
      <xs:maxInclusive value="0" id="nonPositiveInteger.maxInclusive"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="negativeInteger" id="negativeInteger">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#negativeInteger"/>
    </xs:annotation>
    <xs:restriction base="xs:nonPositiveInteger">
      <xs:maxInclusive value="-1" id="negativeInteger.maxInclusive"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="long" id="long">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasProperty name="bounded" value="true"/>
        <hfp:hasProperty name="cardinality" value="finite"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#long"/>
    </xs:annotation>
    <xs:restriction base="xs:integer">
      <xs:minInclusive value="-9223372036854775808" id="long.minInclusive"/>
      <xs:maxInclusive value="9223372036854775807" id="long.maxInclusive"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="int" id="int">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#int"/>
    </xs:annotation>
    <xs:restriction base="xs:long">
      <xs:minInclusive value="-2147483648" id="int.minInclusive"/>
      <xs:maxInclusive value="2147483647" id="int.maxInclusive"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="short" id="short">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#short"/>
    </xs:annotation>
    <xs:restriction base="xs:int">
      <xs:minInclusive value="-32768" id="short.minInclusive"/>
      <xs:maxInclusive value="32767" id="short.maxInclusive"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="byte" id="byte">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#byte"/>
    </xs:annotation>
    <xs:restriction base="xs:short">
      <xs:minInclusive value="-128" id="byte.minInclusive"/>
      <xs:maxInclusive value="127" id="byte.maxInclusive"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="nonNegativeInteger" id="nonNegativeInteger">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#nonNegativeInteger"/>
    </xs:annotation>
    <xs:restriction base="xs:integer">
      <xs:minInclusive value="0" id="nonNegativeInteger.minInclusive"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="unsignedLong" id="unsignedLong">
    <xs:annotation>
      <xs:appinfo>
        <hfp:hasProperty name="bounded" value="true"/>
        <hfp:hasProperty name="cardinality" value="finite"/>
      </xs:appinfo>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#unsignedLong"/>
    </xs:annotation>
    <xs:restriction base="xs:nonNegativeInteger">
      <xs:maxInclusive value="18446744073709551615" id="unsignedLong.maxInclusive"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="unsignedInt" id="unsignedInt">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#unsignedInt"/>
    </xs:annotation>
    <xs:restriction base="xs:unsignedLong">
      <xs:maxInclusive value="4294967295" id="unsignedInt.maxInclusive"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="unsignedShort" id="unsignedShort">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#unsignedShort"/>
    </xs:annotation>
    <xs:restriction base="xs:unsignedInt">
      <xs:maxInclusive value="65535" id="unsignedShort.maxInclusive"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="unsignedByte" id="unsignedByte">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#unsignedByte"/>
    </xs:annotation>
    <xs:restriction base="xs:unsignedShort">
      <xs:maxInclusive value="255" id="unsignedByte.maxInclusive"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="positiveInteger" id="positiveInteger">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#positiveInteger"/>
    </xs:annotation>
    <xs:restriction base="xs:nonNegativeInteger">
      <xs:minInclusive value="1" id="positiveInteger.minInclusive"/>
    </xs:restriction>
  </xs:simpleType>

  <xs:simpleType name="yearMonthDuration">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#yearMonthDuration">
        This type includes just those durations expressed in years and months.
        Since the pattern given excludes days, hours, minutes, and seconds,
        the values of this type have a seconds property of zero.  They are
        totally ordered.
      </xs:documentation>
    </xs:annotation>
    <xs:restriction base="xs:duration">
      <xs:pattern id="yearMonthDuration.pattern" value="[^DT]*"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="dayTimeDuration">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#dayTimeDuration">
        This type includes just those durations expressed in days, hours, minutes, and seconds.
        The pattern given excludes years and months, so the values of this type
        have a months property of zero.  They are totally ordered.
      </xs:documentation>
    </xs:annotation>
    <xs:restriction base="xs:duration">
      <xs:pattern id="dayTimeDuration.pattern" value="[^YM]*(T.*)?"/>
     </xs:restriction>
  </xs:simpleType>
    <xs:simpleType name="dateTimeStamp" id="dateTimeStamp">
    <xs:annotation>
      <xs:documentation source="http://www.w3.org/TR/xmlschema11-2/#dateTimeStamp">
        This datatype includes just those dateTime values Whose explicitTimezone
        is present.  They are totally ordered.
      </xs:documentation>
    </xs:annotation>
    <xs:restriction base="xs:dateTime">
      <xs:explicitTimezone fixed="true"
        id="dateTimeStamp.explicitTimezone" value="required"/>
     </xs:restriction>
  </xs:simpleType>

</xs:schema>
```

## <a id="constructedValueSpaces"></a>D Built-up Value Spaces

Some datatypes, such as [integer](#integer), describe well-known mathematically abstract systems.  Others, such as the date/time datatypes, describe "real-life", "applied" systems.  Certain of the systems described by datatypes, both abstract and applied, have values in their value spaces most easily described as things having several *properties*, which in turn have values which are in some sense "primitive" or are from the value spaces of simpler datatypes.

In this document, the arguments to functions are assumed to be "call by value" unless explicitly noted to the contrary, meaning that if the argument is modified during the processing of the algorithm, that modification is *not* reflected in the "outside world".  On the other hand, the arguments to procedures are assumed to be "call by location", meaning that modifications *are* so reflected, since that is the only way the processing of the algorithm can have any effect.

Properties always have values.  <a id="dt-optional">[Definition:]</a>An **optional** property is *permitted* but not *required* to have the distinguished value ***absent***.

<a id="key-null">[Definition:]</a>Throughout this specification, the value *****absent***** is used as a distinguished value to indicate that a given instance of a property "has no value" or "is absent".  This should not be interpreted as constraining implementations, as for instance between using a ***null*** value for such properties or not representing them at all.

Those values that are more primitive, and are used (among other things) herein to construct object value spaces but which we do not explicitly define are described here:

- A **number (without precision)** is an ordinary mathematical number; 1, 1.0, and 1.000000000000 are the same number.  The decimal numbers and integers generally used in the algorithms of appendix [Function Definitions (§E)](#ap-funcDefs) are such ordinary numbers, not carrying precision.
- <a id="dt-specialvalue">[Definition:]</a>A **special value** is an object whose only relevant properties for purposes of this specification are that it is distinct from, and unequal to, any other values (special or otherwise).  A few special values in different value spaces (e.g. ***positiveInfinity***, ***negativeInfinity***, and ***notANumber*** in [float](#float) and [double](#double)) share names.  Thus, special values can be distinguished from each other in the general case by considering both the name and the primitive datatype of the value; in some cases, of course, the name alone suffices to identify the value uniquely.<a id="b3226move.n1"></a>**Note:** In the case of [float](#float) and [double](#double), the [special values](#dt-specialvalue) are members of the datatype's [value space](#dt-value-space). 

### <a id="sec-numericalValues"></a>D.1 Numerical Values

The following standard operators are defined here in case the reader is unsure of their definition:

- <a id="dt-div">[Definition:]</a>If *m* and *n* are numbers, then *m* **div** *n* is the greatest integer less than or equal to *m* / *n* .
- <a id="dt-mod">[Definition:]</a>If *m* and *n* are numbers, then *m* **mod** *n* is  *m* − *n* × ( *m* [div](#dt-div) *n*) .
 **Note:** *n* [div](#dt-div) 1  is a convenient and short way of expressing "the greatest integer less than or equal to *n*".

#### <a id="sec-exactmaps"></a>D.1.1 Exact Lexical Mappings

 Numerals and Fragments Thereof  <a id="nt-digit"></a>[45]   *digit* ::= [`0-9`] <a id="nt-unsNoDecNuml"></a>[46]   *unsignedNoDecimalPtNumeral* ::= [*digit*](#nt-digit)+ <a id="nt-noDecNuml"></a>[47]   *noDecimalPtNumeral* ::= ('`+`' | '`-`')? [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) <a id="nt-fracFrag"></a>[48]   *fracFrag* ::= [*digit*](#nt-digit)+ <a id="nt-unsDecNuml"></a>[49]   *unsignedDecimalPtNumeral* ::= ([*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) '`.`' [*fracFrag*](#nt-fracFrag)?) | ('`.`' [*fracFrag*](#nt-fracFrag)) <a id="nt-unsFullDecNuml"></a>[50]   *unsignedFullDecimalPtNumeral* ::= [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) '`.`' [*fracFrag*](#nt-fracFrag) <a id="nt-decNuml"></a>[51]   *decimalPtNumeral* ::= ('`+`' | '`-`')? [*unsignedDecimalPtNumeral*](#nt-unsDecNuml) <a id="nt-unsSciNuml"></a>[52]   *unsignedScientificNotationNumeral* ::= ([*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) | [*unsignedDecimalPtNumeral*](#nt-unsDecNuml)) ('`e`' | '`E`') [*noDecimalPtNumeral*](#nt-noDecNuml) <a id="nt-sciNuml"></a>[53]   *scientificNotationNumeral* ::= ('`+`' | '`-`')? [*unsignedScientificNotationNumeral*](#nt-unsSciNuml)    Generic Numeral-to-Number Lexical Mappings **[*unsignedNoDecimalMap*](#f-unsNoDecVal)** (*N*) → integer Maps an [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) to its numerical value. **[*noDecimalMap*](#f-noDecVal)** (*N*) → integer Maps an [*noDecimalPtNumeral*](#nt-noDecNuml) to its numerical value. **[*unsignedDecimalPtMap*](#f-unsDecVal)** (*D*) → decimal number Maps an [*unsignedDecimalPtNumeral*](#nt-unsDecNuml) to its numerical value. **[*decimalPtMap*](#f-decVal)** (*N*) → decimal number Maps a [*decimalPtNumeral*](#nt-decNuml) to its numerical value. **[*scientificMap*](#f-sciVal)** (*N*) → decimal number Maps a [*scientificNotationNumeral*](#nt-sciNuml) to its numerical value.    Generic Number to Numeral Canonical Mappings **[*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)** (*i*) → [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) Maps a nonnegative integer to a [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml), its [canonical representation](#dt-canonical-representation). **[*noDecimalPtCanonicalMap*](#f-noDecCanMap)** (*i*) → [*noDecimalPtNumeral*](#nt-noDecNuml) Maps an integer to a [*noDecimalPtNumeral*](#nt-noDecNuml), its [canonical representation](#dt-canonical-representation). **[*unsignedDecimalPtCanonicalMap*](#f-unsDecCanFragMap)** (*n*) → [*unsignedDecimalPtNumeral*](#nt-unsDecNuml) Maps a nonnegative decimal number to a [*unsignedDecimalPtNumeral*](#nt-unsDecNuml), its [canonical representation](#dt-canonical-representation). **[*decimalPtCanonicalMap*](#f-decCanFragMap)** (*n*) → [*decimalPtNumeral*](#nt-decNuml) Maps a decimal number to a [*decimalPtNumeral*](#nt-decNuml), its [canonical representation](#dt-canonical-representation). **[*unsignedScientificCanonicalMap*](#f-unsSciCanFragMap)** (*n*) → [*unsignedScientificNotationNumeral*](#nt-unsSciNuml) Maps a nonnegative decimal number to a [*unsignedScientificNotationNumeral*](#nt-unsSciNuml), its [canonical representation](#dt-canonical-representation). **[*scientificCanonicalMap*](#f-sciCanFragMap)** (*n*) → [*scientificNotationNumeral*](#nt-sciNuml) Maps a decimal number to a [*scientificNotationNumeral*](#nt-sciNuml), its [canonical representation](#dt-canonical-representation).  Some numerical datatypes include some or all of three non-numerical [special values](#dt-specialvalue): ***positiveInfinity***, ***negativeInfinity***, and ***notANumber***.  Their lexical spaces include non-numeral lexical representations for these non-numeric values:  Special Non-numerical Lexical Representations Used With Numerical Datatypes  <a id="nt-minNumSpecReps"></a>[54]   *minimalNumericalSpecialRep* ::= '`INF`' | '`-INF`' | '`NaN`' <a id="nt-numSpecReps"></a>[55]   *numericalSpecialRep* ::= '`+INF`' | [*minimalNumericalSpecialRep*](#nt-minNumSpecReps)   Lexical Mapping for Non-numerical [Special Values](#dt-specialvalue) Used With Numerical Datatypes **[*specialRepValue*](#f-specRepVal)** (*S*) → a [special value](#dt-specialvalue) Maps the [lexical representations](#dt-lexical-representation) of [special values](#dt-specialvalue) used with some numerical datatypes to those [special values](#dt-specialvalue).    Canonical Mapping for Non-numerical [Special Values](#dt-specialvalue) Used with Numerical Datatypes **[*specialRepCanonicalMap*](#f-specValCanMap)** (*c*) → [*numericalSpecialRep*](#nt-numSpecReps) Maps the [special values](#dt-specialvalue) used with some numerical datatypes to their [canonical representations](#dt-canonical-representation).

### <a id="d-t-values"></a>D.2 Date/time Values

        D.2.1 [The Seven-property Model](#theSevenPropertyModel)
         D.2.2 [Lexical Mappings](#rf-lexicalMappings-datetime)

There are several different primitive but related datatypes defined in the specification which pertain to various combinations of dates and times, and parts thereof.  They all use related value-space models, which are described in detail in this section.  It is not difficult for a casual reader of the descriptions of the individual datatypes elsewhere in this specification to misunderstand some of the details of just what the datatypes are intended to represent, so more detail is presented here in this section.

All of the value spaces for dates and times described here represent moments or periods of time in Universal Coordinated Time (UTC).  <a id="dt-utc">[Definition:]</a>**Universal Coordinated Time** (**UTC**) is an adaptation of TAI which closely approximates UT1 by adding [leap-seconds](#dt-leapsec) to selected [UTC](#dt-utc) days.

<a id="dt-leapsec">[Definition:]</a>A **leap-second** is an additional second added to the last day of December, June, October, or March, when such an adjustment is deemed necessary by the International Earth Rotation and Reference Systems Service in order to keep [UTC](#dt-utc) within 0.9 seconds of observed astronomical time.  When leap seconds are introduced, the last minute in the day has more than sixty seconds.  In theory leap seconds can also be removed from a day, but this has not yet occurred. (See [[International Earth Rotation Service (IERS)]](#IERS), [[ITU-R TF.460-6]](#itu-r-460-6).) Leap seconds are *not* supported by the types defined here.

Because the [dateTime](#dateTime) type and other date- and time-related types defined in this specification do not support leap seconds, there are portions of the [UTC](#dt-utc) timeline which cannot be represented by values of these types. Users whose applications require that leap seconds be represented and that date/time arithmetic take historically occurring leap seconds into account will wish to make appropriate adjustments at the application level, or to use other types.

#### <a id="theSevenPropertyModel"></a>D.2.1 The Seven-property Model

There are two distinct ways to model moments in time:  either by tracking their year, month, day, hour, minute and second (with fractional seconds as needed), or by tracking their time (measured generally in seconds or days) from some starting moment.  Each has its advantages.  The two are isomorphic.  For definiteness, we choose to model the first using five integer and one decimal number properties.  We superimpose the second by providing one decimal number-valued function which gives the corresponding count of seconds from zero (the "time on the time line").

There is also a seventh [integer](#integer) property which specifies the time zone offset as the number of minutes of offset from UTC.  Values for the six primary properties are always stored in their "local" values (the values shown in the lexical representations), rather than converted to [UTC](#dt-utc).  Properties of <a id="dt-dt-7PropMod">Date/time Seven-property Models</a>  **<a id="vp-dt-year">*year*</a>** an integer  **<a id="vp-dt-month">*month*</a>** an integer between 1 and 12 inclusive  **<a id="vp-dt-day">*day*</a>** an integer between 1 and 31 inclusive, possibly restricted further depending on [*month*](#vp-dt-month) and [*year*](#vp-dt-year)  **<a id="vp-dt-hour">*hour*</a>** an integer between 0 and 23 inclusive  **<a id="vp-dt-minute">*minute*</a>** an integer between 0 and 59 inclusive  **<a id="vp-dt-second">*second*</a>** a decimal number greater than or equal to 0 and less than 60.  **<a id="vp-dt-timezone">*timezoneOffset*</a>** an [optional](#dt-optional) integer between −840 and 840 inclusive

Non-negative values of the properties map to the years, months, days of month, etc. of the Gregorian calendar in the obvious way. Values less than 1582 in the [*year*](#vp-dt-year) property represent years in the "proleptic Gregorian calendar". A value of zero in the [*year*](#vp-dt-year) property represents the year 1 BCE; a value of −1 represents the year 2 BCE, −2 is 3 BCE, etc.

**Note:** In version 1.0 of this specification, the [*year*](#vp-dt-year) property was not permitted to have the value zero. The year before the year 1 in the proleptic Gregorian calendar, traditionally referred to as 1 BC or as 1 BCE, was represented by a [*year*](#vp-dt-year) value of −1, 2 BCE by −2, and so forth. Of course, many, perhaps most, references to 1 BCE (or 1 BC) actually refer not to a year in the proleptic Gregorian calendar but to a year in the Julian or "old style" calendar; the two correspond approximately but not exactly to each other. In this version of this specification, two changes are made in order to agree with existing usage. First, [*year*](#vp-dt-year) is permitted to have the value zero. Second, the interpretation of [*year*](#vp-dt-year) values is changed accordingly: a [*year*](#vp-dt-year) value of zero represents 1 BCE, −1 represents 2 BCE, etc. This representation simplifies interval arithmetic and leap-year calculation for dates before the common era (which may be why astronomers and others interested in such calculations with the proleptic Gregorian calendar have adopted it), and is consistent with the current edition of [[ISO 8601]](#ISO8601).  Note that 1 BCE, 5 BCE, and so on (years 0000, −0004, etc. in the lexical representation defined here) are leap years in the proleptic Gregorian calendar used for the date/time datatypes defined here. Version 1.0 of this specification was unclear about the treatment of leap years before the common era. If existing schemas or data specify dates of 29 February for any years before the common era, then some values giving a date of 29 February which were valid under a plausible interpretation of XSD 1.0 will be invalid under this specification, and some which were invalid will be valid. With that possible exception, schemas and data valid under the old interpretation remain valid under the new.

The model just described is called herein the "seven-property" model for date/time datatypes.  It is used "as is" for [dateTime](#dateTime); all other date/time datatypes except [duration](#duration) use the same model except that some of the six primary properties are *required* to have the value ***absent***, instead of being required to have a numerical value.  (An *[optional](#dt-optional)* property, like [*timezoneOffset*](#vp-dt-timezone), is always *permitted* to have the value ***absent***.)

[*timezoneOffset*](#vp-dt-timezone) values are limited to 14 hours, which is 840 (= 60 × 14) minutes.

**Note:** Leap-seconds are not permitted

Readers interested in when leap-seconds have been introduced should consult [[USNO Historical List]](#USNavy_leaps), which includes a list of times when the difference between TAI and [UTC](#dt-utc) has changed.  Because the simple types defined here do not support leap seconds, they cannot be used to represent the final second, in [UTC](#dt-utc), of any of the days containing one.  If it is important, at the application level, to track the occurrence of leap seconds, then users will need to make special arrangements for special handling of them and of time intervals crossing them.

While calculating, property values from the [dateTime](#dateTime) 1972-12-31T00:00:00 are used to fill in for those that are ***absent***, except that if [*day*](#vp-dt-day) is ***absent*** but [*month*](#vp-dt-month) is not, the largest permitted day for that month is used.

   Time on Timeline for Date/time Seven-property Model Datatypes **[*timeOnTimeline*](#vp-dt-timeOnTimeline)** (*dt*) → decimal number Maps a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value to the decimal number representing its position on the "time line".  Values from any one date/time datatype using the seven-component model (all except [duration](#duration)) are ordered the same as their [*timeOnTimeline*](#vp-dt-timeOnTimeline) values, except that if one value's [*timezoneOffset*](#vp-dt-timezone) is ***absent*** and the other's is not, and using maximum and minimum [*timezoneOffset*](#vp-dt-timezone) values for the one whose [*timezoneOffset*](#vp-dt-timezone) is actually ***absent*** changes the resulting (strict) inequality, the original two values are incomparable.

#### <a id="rf-lexicalMappings-datetime"></a>D.2.2 Lexical Mappings

<a id="dt-dt-frag">[Definition:]</a>Each lexical representation is made up of certain **date/time fragments**, each of which corresponds to a particular property of the datatype value.  They are defined by the following productions.  Date/time Lexical Representation Fragments  <a id="nt-yrFrag"></a>[56]   *yearFrag* ::= '`-`'? (([`1-9`] [*digit*](#nt-digit) [*digit*](#nt-digit) [*digit*](#nt-digit)+)) | ('`0`' [*digit*](#nt-digit) [*digit*](#nt-digit) [*digit*](#nt-digit))) <a id="nt-moFrag"></a>[57]   *monthFrag* ::= ('`0`' [`1-9`]) | ('`1`' [`0-2`]) <a id="nt-daFrag"></a>[58]   *dayFrag* ::= ('`0`' [`1-9`]) | ([`12`] [*digit*](#nt-digit)) | ('`3`' [`01`]) <a id="nt-hrFrag"></a>[59]   *hourFrag* ::= ([`01`] [*digit*](#nt-digit)) | ('`2`' [`0-3`]) <a id="nt-miFrag"></a>[60]   *minuteFrag* ::= [`0-5`] [*digit*](#nt-digit) <a id="nt-seFrag"></a>[61]   *secondFrag* ::= ([`0-5`] [*digit*](#nt-digit)) ('`.`' [*digit*](#nt-digit)+)? <a id="nt-eodFrag"></a>[62]   *endOfDayFrag* ::= '`24:00:00`' ('`.`' '`0`'+)? <a id="nt-tzFrag"></a>[63]   *timezoneFrag* ::= '`Z`' | ('`+`' | '`-`') (('`0`' [*digit*](#nt-digit) | '`1`' [`0-3`]) '`:`' [*minuteFrag*](#nt-miFrag) | '`14:00`') Each fragment other than [*timezoneFrag*](#nt-tzFrag) defines a subset of the [lexical space](#dt-lexical-space) of [decimal](#decimal); the corresponding [lexical mapping](#dt-lexical-mapping) is the [decimal](#decimal) [lexical mapping](#dt-lexical-mapping) restricted to that subset.  These fragment [lexical mappings](#dt-lexical-mapping) are combined separately for each date/time datatype (other than [duration](#duration)) to make up [the complete lexical mapping](#dt-lexical-mapping) for that datatype.  The [*yearFragValue*](#f-dt-yrMap) mapping is used to obtain the value of the [*year*](#vp-dt-year) property, the [*monthFragValue*](#f-dt-moMap) mapping is used to obtain the value of the [*month*](#vp-dt-month) property, etc.  Each datatype which specifies some properties to be mandatorily ***absent*** also does not permit the corresponding lexical fragments in its lexical representations.   Partial Date/time Lexical Mappings **[*yearFragValue*](#f-dt-yrMap)** (*YR*) → integer Maps a [*yearFrag*](#nt-yrFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*year*](#vp-dt-year) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value. **[*monthFragValue*](#f-dt-moMap)** (*MO*) → integer Maps a [*monthFrag*](#nt-moFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*month*](#vp-dt-month) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value. **[*dayFragValue*](#f-dt-daMap)** (*DA*) → integer Maps a [*dayFrag*](#nt-daFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*day*](#vp-dt-day) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value. **[*hourFragValue*](#f-dt-hrMap)** (*HR*) → integer Maps a [*hourFrag*](#nt-hrFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*hour*](#vp-dt-hour) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value. **[*minuteFragValue*](#f-dt-miMap)** (*MI*) → integer Maps a [*minuteFrag*](#nt-miFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*minute*](#vp-dt-minute) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value. **[*secondFragValue*](#f-dt-seMap)** (*SE*) → decimal number Maps a [*secondFrag*](#nt-seFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto a decimal number, presumably the [*second*](#vp-dt-second) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value. **[*timezoneFragValue*](#f-dt-tzMap)** (*TZ*) → integer Maps a [*timezoneFrag*](#nt-tzFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*timezoneOffset*](#vp-dt-timezone) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value.  **Note:** The redundancy between '`Z`', '`+00:00`', and '`-00:00`', and the possibility of trailing fractional '`0`' digits for [*secondFrag*](#nt-seFrag), are the only redundancies preventing these mappings from being one-to-one. There is no [lexical mapping](#dt-lexical-mapping) for [*endOfDayFrag*](#nt-eodFrag); it is handled specially by the relevant [lexical mappings](#dt-lexical-mapping).  See, e.g., [*dateTimeLexicalMap*](#vp-dateTimeLexRep). The following fragment [canonical mappings](#dt-canonical-mapping) for each value-object property are combined as appropriate to make the [canonical mapping](#dt-canonical-mapping) for each date/time datatype (other than [duration](#duration)):   Partial Date/time Canonical Mappings **[*yearCanonicalFragmentMap*](#f-yrCanFragMap)** (*y*) → [*yearFrag*](#nt-yrFrag) Maps an integer, presumably the [*year*](#vp-dt-year) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*yearFrag*](#nt-yrFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation). **[*monthCanonicalFragmentMap*](#f-moCanFragMap)** (*m*) → [*monthFrag*](#nt-moFrag) Maps an integer, presumably the [*month*](#vp-dt-month) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*monthFrag*](#nt-moFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation). **[*dayCanonicalFragmentMap*](#f-daCanFragMap)** (*d*) → [*dayFrag*](#nt-daFrag) Maps an integer, presumably the [*day*](#vp-dt-day) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*dayFrag*](#nt-daFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation). **[*hourCanonicalFragmentMap*](#f-hrCanFragMap)** (*h*) → [*hourFrag*](#nt-hrFrag) Maps an integer, presumably the [*hour*](#vp-dt-hour) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*hourFrag*](#nt-hrFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation). **[*minuteCanonicalFragmentMap*](#f-miCanFragMap)** (*m*) → [*minuteFrag*](#nt-miFrag) Maps an integer, presumably the [*minute*](#vp-dt-minute) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*minuteFrag*](#nt-miFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation). **[*secondCanonicalFragmentMap*](#f-seCanFragMap)** (*s*) → [*secondFrag*](#nt-seFrag) Maps a decimal number, presumably the [*second*](#vp-dt-second) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*secondFrag*](#nt-seFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation). **[*timezoneCanonicalFragmentMap*](#f-tzCanFragMap)** (*t*) → [*timezoneFrag*](#nt-tzFrag) Maps an integer, presumably the [*timezoneOffset*](#vp-dt-timezone) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*timezoneFrag*](#nt-tzFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation).

## <a id="ap-funcDefs"></a>E Function Definitions

The more important functions and procedures defined here are summarized in the text  When there is a text summary, the name of the function in each is a "hot-link" to the same name in the other.  All other links to these functions link to the complete definition in this section.

### <a id="sec-generic-number-functions"></a>E.1 Generic Number-related Functions

The following functions are used with various numeric and date/time datatypes.

  Auxiliary Functions for Operating on Numeral Fragments **<a id="f-digitVal">*digitValue*</a>** (*d*) → integer Maps each digit to its numerical value.  **Arguments:**

<table><tbody>
    <tr><td>d</td><td>: </td><td>matches <a href="#nt-digit">digit</a></td></tr>

    </tbody></table>

 **Result:**a nonnegative integer less than ten **Algorithm:** Return

- 0   when  *d* = '`0`' ,
- 1   when  *d* = '`1`' ,
- 2   when  *d* = '`2`' ,
- *etc.*
  **<a id="f-digitSeqVal">*digitSequenceValue*</a>** (*S*) → integer Maps a sequence of digits to the position-weighted sum of the terms numerical values.  **Arguments:**

<table><tbody>
    <tr><td>S</td><td>: </td><td>a finite sequence of
        <a href="#dt-literal">·literals·</a>, each term matching <a href="#nt-digit">digit</a>.</td></tr>

    </tbody></table>

 **Result:**a nonnegative integer **Algorithm:** Return the sum of [*digitValue*](#f-digitVal)(*S*~*i*~) × 10^length(*S*)−*i*^  where *i* runs over the domain of *S*.  **<a id="f-fracDigitSeqVal">*fractionDigitSequenceValue*</a>** (*S*) → integer Maps a sequence of digits to the position-weighted sum of the terms numerical values, weighted appropriately for fractional digits.  **Arguments:**

<table><tbody>
    <tr><td>S</td><td>: </td><td>a finite sequence of
        <a href="#dt-literal">·literals·</a>, each term matching <a href="#nt-digit">digit</a>.</td></tr>

    </tbody></table>

 **Result:**a nonnegative integer **Algorithm:** Return the sum of [*digitValue*](#f-digitVal)(*S*~*i*~) − 10^−*i*^  where *i* runs over the domain of *S*.  **<a id="f-fracFragVal">*fractionFragValue*</a>** (*N*) → decimal number Maps a [*fracFrag*](#nt-fracFrag) to the appropriate fractional decimal number.  **Arguments:**

<table><tbody>
    <tr><td>N</td><td>: </td><td>matches <a href="#nt-fracFrag">fracFrag</a></td></tr>

    </tbody></table>

 **Result:**a nonnegative decimal number **Algorithm:** *N* is necessarily the left-to-right concatenation of a finite sequence *S* of [literals](#dt-literal), each term matching [*digit*](#nt-digit). Return [*fractionDigitSequenceValue*](#f-fracDigitSeqVal)(*S*).   Generic Numeral-to-Number Lexical Mappings **<a id="f-unsNoDecVal">*unsignedNoDecimalMap*</a>** (*N*) → integer Maps an [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) to its numerical value.  **Arguments:**

<table><tbody>
    <tr><td>N</td><td>: </td><td>matches <a href="#nt-unsNoDecNuml">unsignedNoDecimalPtNumeral</a></td></tr>

    </tbody></table>

 **Result:**a nonnegative integer **Algorithm:** *N* is the left-to-right concatenation of a finite sequence *S* of [literals](#dt-literal), each term matching [*digit*](#nt-digit). Return [*digitSequenceValue*](#f-digitSeqVal)(*S*). **<a id="f-noDecVal">*noDecimalMap*</a>** (*N*) → integer Maps an [*noDecimalPtNumeral*](#nt-noDecNuml) to its numerical value.  **Arguments:**

<table><tbody>
    <tr><td>N</td><td>: </td><td>matches <a href="#nt-noDecNuml">noDecimalPtNumeral</a></td></tr>

    </tbody></table>

 **Result:**an integer **Algorithm:** *N* necessarily consists of an optional sign('`+`' or '`-`') and then a [literal](#dt-literal) *U* that matches [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml). Return

- −1 × [*unsignedNoDecimalMap*](#f-unsNoDecVal)(*U*)   when '`-`' is present, and
- [*unsignedNoDecimalMap*](#f-unsNoDecVal)(*U*)   otherwise.
 **<a id="f-unsDecVal">*unsignedDecimalPtMap*</a>** (*D*) → decimal number Maps an [*unsignedDecimalPtNumeral*](#nt-unsDecNuml) to its numerical value.  **Arguments:**

<table><tbody>
    <tr><td>D</td><td>: </td><td>matches <a href="#nt-unsDecNuml">unsignedDecimalPtNumeral</a></td></tr>

    </tbody></table>

 **Result:**a nonnegative decimal number **Algorithm:** *D* necessarily consists of an optional [literal](#dt-literal) *N* matching [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml), a decimal point, and then an optional [literal](#dt-literal) *F* matching [*fracFrag*](#nt-fracFrag). Return

- [*unsignedNoDecimalMap*](#f-unsNoDecVal)(*N*)   when *F* is not present,
- [*fractionFragValue*](#f-fracFragVal)(*F*)   when *N* is not present, and
- [*unsignedNoDecimalMap*](#f-unsNoDecVal)(*N*) + [*fractionFragValue*](#f-fracFragVal)(*F*)   otherwise.
 **<a id="f-decVal">*decimalPtMap*</a>** (*N*) → decimal number Maps a [*decimalPtNumeral*](#nt-decNuml) to its numerical value.  **Arguments:**

<table><tbody>
    <tr><td>N</td><td>: </td><td>matches <a href="#nt-decNuml">decimalPtNumeral</a></td></tr>

    </tbody></table>

 **Result:**a decimal number **Algorithm:** *N* necessarily consists of an optional sign('`+`' or '`-`') and then an instance *U* of [*unsignedDecimalPtNumeral*](#nt-unsDecNuml).  Return

- −[*unsignedDecimalPtMap*](#f-unsDecVal)(*U*)   when '`-`' is present, and
- [*unsignedDecimalPtMap*](#f-unsDecVal)(*U*)   otherwise.
 **<a id="f-sciVal">*scientificMap*</a>** (*N*) → decimal number Maps a [*scientificNotationNumeral*](#nt-sciNuml) to its numerical value.  **Arguments:**

<table><tbody>
    <tr><td>N</td><td>: </td><td>matches <a href="#nt-sciNuml">scientificNotationNumeral</a></td></tr>

    </tbody></table>

 **Result:**a decimal number **Algorithm:** *N* necessarily consists of an instance *C* of either [*noDecimalPtNumeral*](#nt-noDecNuml) or [*decimalPtNumeral*](#nt-decNuml), either an '`e`' or an '`E`', and then an instance *E* of [*noDecimalPtNumeral*](#nt-noDecNuml). Return

-  [*decimalPtMap*](#f-decVal)(*C*) × 10 ^ [*unsignedDecimalPtMap*](#f-unsDecVal)(*E*)   when a '`.`' is present in *N*, and
-  [*noDecimalMap*](#f-noDecVal)(*C*) × 10 ^ [*unsignedDecimalPtMap*](#f-unsDecVal)(*E*)   otherwise.
   Auxiliary Functions for Producing Numeral Fragments **<a id="f-digit">*digit*</a>** (*i*) → [*digit*](#nt-digit) Maps each integer between 0 and 9 to the corresponding [*digit*](#nt-digit).  **Arguments:**

<table><tbody>
    <tr><td>i</td><td>: </td><td>between 0 and 9 inclusive</td></tr>

    </tbody></table>

 **Result:**matches [*digit*](#nt-digit) **Algorithm:** Return

- '`0`'   when  *i* = 0 ,
- '`1`'   when  *i* = 1 ,
- '`2`'   when  *i* = 2 ,
- etc.
 **<a id="f-digitRemSeq">*digitRemainderSeq*</a>** (*i*) → sequence of integers Maps each nonnegative integer to a sequence of integers used by [*digitSeq*](#f-digitSeq) to ultimately create an [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml).  **Arguments:**

<table><tbody>
    <tr><td>i</td><td>: </td><td>a nonnegative integer</td></tr>

    </tbody></table>

 **Result:**sequence of nonnegative integers **Algorithm:** Return that sequence *s* for which

- *s*~0~ = *i*  and
- *s*~*j*+1~ = *s*~*j*~ [div](#dt-div) 10 .
  **<a id="f-digitSeq">*digitSeq*</a>** (*i*) → sequence of integers Maps each nonnegative integer to a sequence of integers used by [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap) to create an [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml).  **Arguments:**

<table><tbody>
    <tr><td>i</td><td>: </td><td>a nonnegative integer</td></tr>

    </tbody></table>

 **Result:**sequence of integers where each term is between 0 and 9 inclusive **Algorithm:** Return that sequence *s* for which  *s*~*j*~ =[*digitRemainderSeq*](#f-digitRemSeq)(*i*)~*j*~ [mod](#dt-mod) 10 .  **<a id="f-lastSigDigit">*lastSignificantDigit*</a>** (*s*) → integer Maps a sequence of nonnegative integers to the index of the first zero term.  **Arguments:**

<table><tbody>
    <tr><td>s</td><td>: </td><td>a sequence of nonnegative integers</td></tr>

    </tbody></table>

 **Result:**a nonnegative integer **Algorithm:** Return the smallest nonnegative integer *j* such that *s*(*i*)~*j*+1~ is 0.  **<a id="f-fracDigitRemSeq">*FractionDigitRemainderSeq*</a>** (*f*) → sequence of decimal numbers Maps each nonnegative decimal number less than 1 to a sequence of decimal numbers used by [*fractionDigitSeq*](#f-fracDigitSeq) to ultimately create an [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml).  **Arguments:**

<table><tbody>
    <tr><td>f</td><td>: </td><td>nonnegative and less than 1</td></tr>

    </tbody></table>

 **Result:**a sequence of nonnegative decimal numbers **Algorithm:** Return that sequence *s* for which

- *s*~0~ = *f* − 10 , and
- *s*~*j*+1~ = (*s*~*j*~ [mod](#dt-mod) 1) − 10 .
  **<a id="f-fracDigitSeq">*fractionDigitSeq*</a>** (*f*) → sequence of integers Maps each nonnegative decimal number less than 1 to a sequence of integers used by [*fractionDigitsCanonicalFragmentMap*](#f-fracDigitsMap) to ultimately create an [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml).  **Arguments:**

<table><tbody>
    <tr><td>f</td><td>: </td><td>nonnegative and less than 1</td></tr>

    </tbody></table>

 **Result:**a sequence of integer;s where each term is between 0 and 9 inclusive **Algorithm:** Return that sequence *s* for which  *s*~*j*~ = [*FractionDigitRemainderSeq*](#f-fracDigitRemSeq)(*f*)~*j*~ [div](#dt-div) 1 .  **<a id="f-fracDigitsMap">*fractionDigitsCanonicalFragmentMap*</a>** (*f*) → [*fracFrag*](#nt-fracFrag) Maps each nonnegative decimal number less than 1 to a [literal](#dt-literal) used by [*unsignedDecimalPtCanonicalMap*](#f-unsDecCanFragMap) to create an [*unsignedDecimalPtNumeral*](#nt-unsDecNuml).  **Arguments:**

<table><tbody>
    <tr><td>f</td><td>: </td><td>nonnegative and less than 1</td></tr>

    </tbody></table>

 **Result:**matches [*fracFrag*](#nt-fracFrag) **Algorithm:** Return  [*digit*](#f-digit)([*fractionDigitSeq*](#f-fracDigitSeq)(*f*)~0~) & . . . & [*digit*](#f-digit)([*fractionDigitSeq*](#f-fracDigitSeq)(*f*)~[*lastSignificantDigit*](#f-lastSigDigit)([*FractionDigitRemainderSeq*](#f-fracDigitRemSeq)(*f*))~) .   Generic Number to Numeral Canonical Mappings **<a id="f-unsNoDecCanFragMap">*unsignedNoDecimalPtCanonicalMap*</a>** (*i*) → [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) Maps a nonnegative integer to a [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml), its [canonical representation](#dt-canonical-representation).  **Arguments:**

<table><tbody>
    <tr><td>i</td><td>: </td><td>a nonnegative integer</td></tr>

    </tbody></table>

 **Result:**matches [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) **Algorithm:** Return [*digit*](#f-digit)([*digitSeq*](#f-digitSeq)(*i*)~[*lastSignificantDigit*](#f-lastSigDigit)([*digitRemainderSeq*](#f-digitRemSeq)(*i*))~) & . . . & [*digit*](#f-digit)([*digitSeq*](#f-digitSeq)(*i*)~0~) .   (Note that the concatenation is in reverse order.) **<a id="f-noDecCanMap">*noDecimalPtCanonicalMap*</a>** (*i*) → [*noDecimalPtNumeral*](#nt-noDecNuml) Maps an integer to a [*noDecimalPtNumeral*](#nt-noDecNuml), its [canonical representation](#dt-canonical-representation).  **Arguments:**

<table><tbody>
    <tr><td>i</td><td>: </td><td>an integer</td></tr>

    </tbody></table>

 **Result:**matches [*noDecimalPtNumeral*](#nt-noDecNuml) **Algorithm:** Return

- '`-`' & [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(−*i*)   when *i* is negative,
- [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(*i*)   otherwise.
  **<a id="f-unsDecCanFragMap">*unsignedDecimalPtCanonicalMap*</a>** (*n*) → [*unsignedDecimalPtNumeral*](#nt-unsDecNuml) Maps a nonnegative decimal number to a [*unsignedDecimalPtNumeral*](#nt-unsDecNuml), its [canonical representation](#dt-canonical-representation).  **Arguments:**

<table><tbody>
    <tr><td>n</td><td>: </td><td>a nonnegative decimal number</td></tr>

    </tbody></table>

 **Result:**matches [*unsignedDecimalPtNumeral*](#nt-unsDecNuml) **Algorithm:** Return  [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(*n*[div](#dt-div)1) & '`.`' & [*fractionDigitsCanonicalFragmentMap*](#f-fracDigitsMap)(*n*[mod](#dt-mod)1) . **<a id="f-decCanFragMap">*decimalPtCanonicalMap*</a>** (*n*) → [*decimalPtNumeral*](#nt-decNuml) Maps a decimal number to a [*decimalPtNumeral*](#nt-decNuml), its [canonical representation](#dt-canonical-representation).  **Arguments:**

<table><tbody>
    <tr><td>n</td><td>: </td><td>a decimal number</td></tr>

    </tbody></table>

 **Result:**matches [*decimalPtNumeral*](#nt-decNuml) **Algorithm:** Return

- '`-`' & [*unsignedDecimalPtCanonicalMap*](#f-unsDecCanFragMap)(−*i*)   when *i* is negative,
- [*unsignedDecimalPtCanonicalMap*](#f-unsDecCanFragMap)(*i*)   otherwise.
  **<a id="f-unsSciCanFragMap">*unsignedScientificCanonicalMap*</a>** (*n*) → [*unsignedScientificNotationNumeral*](#nt-unsSciNuml) Maps a nonnegative decimal number to a [*unsignedScientificNotationNumeral*](#nt-unsSciNuml), its [canonical representation](#dt-canonical-representation).  **Arguments:**

<table><tbody>
    <tr><td>n</td><td>: </td><td>a nonnegative decimal number</td></tr>

    </tbody></table>

 **Result:**matches [*unsignedScientificNotationNumeral*](#nt-unsSciNuml) **Algorithm:**  Return  [*unsignedDecimalPtCanonicalMap*](#f-unsDecCanFragMap)(*n* / 10^log(*n*) [div](#dt-div) 1^) & '`E`' & [*noDecimalPtCanonicalMap*](#f-noDecCanMap)(log(*n*) [div](#dt-div) 1)  **<a id="f-sciCanFragMap">*scientificCanonicalMap*</a>** (*n*) → [*scientificNotationNumeral*](#nt-sciNuml) Maps a decimal number to a [*scientificNotationNumeral*](#nt-sciNuml), its [canonical representation](#dt-canonical-representation).  **Arguments:**

<table><tbody>
    <tr><td>n</td><td>: </td><td>a decimal number</td></tr>

    </tbody></table>

 **Result:**matches [*scientificNotationNumeral*](#nt-sciNuml) **Algorithm:** Return

- '`-`' & [*unsignedScientificCanonicalMap*](#f-unsSciCanFragMap)(−*n*)   when *n* is negative,
- [*unsignedScientificCanonicalMap*](#f-unsSciCanFragMap)(*i*)   otherwise.

For example:

- 123.4567 [mod](#dt-mod) 1 = 0.4567  and  123.4567 [div](#dt-div) 1 = 123 .
- [*digitRemainderSeq*](#f-digitRemSeq)(123)  is  123 , 12 , 1 , 0 , 0 , . . . .
- [*digitSeq*](#f-digitSeq)(123)  is  3 , 2 , 1 , 0 , 0 , . . . .
- [*lastSignificantDigit*](#f-lastSigDigit)([*digitRemainderSeq*](#f-digitRemSeq)(123)) = 2   (Sequences count from 0.)
- [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(123) = '`123`'
- [*FractionDigitRemainderSeq*](#f-fracDigitRemSeq)(0.4567)  is  4.567 , 5.67 , 6.7 , 7 , 0 , 0 , . . . .
- [*fractionDigitSeq*](#f-fracDigitSeq)(0.4567)  is  4 , 5 , 6 , 7 , 0 , 0 , . . . .
- [*lastSignificantDigit*](#f-lastSigDigit)([*FractionDigitRemainderSeq*](#f-fracDigitRemSeq)(0.4567)) = 3
- [*fractionDigitsCanonicalFragmentMap*](#f-fracDigitsMap)(0.4567) = '`4567`'
- [*unsignedDecimalPtCanonicalMap*](#f-unsDecCanFragMap)(123.4567) = '`123.4567`'
  Lexical Mapping for Non-numerical [Special Values](#dt-specialvalue) Used With Numerical Datatypes **<a id="f-specRepVal">*specialRepValue*</a>** (*S*) → a [special value](#dt-specialvalue) Maps the [lexical representations](#dt-lexical-representation) of [special values](#dt-specialvalue) used with some numerical datatypes to those [special values](#dt-specialvalue).  **Arguments:**

<table><tbody>
    <tr><td>S</td><td>: </td><td>matches <a href="#nt-numSpecReps">numericalSpecialRep</a></td></tr>

    </tbody></table>

 **Result:**one of ***positiveInfinity***, ***negativeInfinity***, or ***notANumber***. **Algorithm:**  Return

- ***positiveInfinity***   when *S* is '`INF`' or '`+INF`',
- ***negativeInfinity***   when *S* is '`-INF`', and
- ***notANumber***   when *S* is '`NaN`'
    Canonical Mapping for Non-numerical [Special Values](#dt-specialvalue) Used with Numerical Datatypes **<a id="f-specValCanMap">*specialRepCanonicalMap*</a>** (*c*) → [*numericalSpecialRep*](#nt-numSpecReps) Maps the [special values](#dt-specialvalue) used with some numerical datatypes to their [canonical representations](#dt-canonical-representation).  **Arguments:**

<table><tbody>
    <tr><td>c</td><td>: </td><td>one of positiveInfinity,
        negativeInfinity, and notANumber</td></tr>

    </tbody></table>

 **Result:**matches [*numericalSpecialRep*](#nt-numSpecReps) **Algorithm:** Return

- '`INF`'   when *c* is ***positiveInfinity***
- '`-INF`'   when *c* is ***negativeInfinity***
- '`NaN`'   when *c* is ***notANumber***
    Lexical Mapping **<a id="f-decimalLexmap">*decimalLexicalMap*</a>** (*LEX*) → [decimal](#decimal) Maps a [*decimalLexicalRep*](#nt-decimalRep) onto a [decimal](#decimal) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>matches <a href="#nt-decimalRep">decimalLexicalRep</a></td></tr>

    </tbody></table>

 **Result:**a [decimal](#decimal) value **Algorithm:**

<table><tbody><tr><td>Let </td><td>d be a <a href="#decimal">decimal</a> value.</td></tr></tbody></table>

1. Set *d* to

  - [*noDecimalMap*](#f-noDecVal)(*LEX*)   when *LEX* is an instance of [*noDecimalPtNumeral*](#nt-noDecNuml), and
  - [*decimalPtMap*](#f-decVal)(*LEX*)   when *LEX* is an instance of [*decimalPtNumeral*](#nt-decNuml),

2. Return *d*.
    Canonical Mapping **<a id="f-decimalCanmap">*decimalCanonicalMap*</a>** (*d*) → [*decimalLexicalRep*](#nt-decimalRep) Maps a [decimal](#decimal) to its [canonical representation](#dt-canonical-representation), a [*decimalLexicalRep*](#nt-decimalRep).  **Arguments:**

<table><tbody>
    <tr><td>d</td><td>: </td><td>a <a href="#decimal">decimal</a> value</td></tr>

    </tbody></table>

 **Result:**a [literal](#dt-literal) matching [*decimalLexicalRep*](#nt-decimalRep) **Algorithm:**

1. If *d* is an integer, then return [*noDecimalPtCanonicalMap*](#f-noDecCanMap)(*d*).
2. Otherwise, return [*decimalPtCanonicalMap*](#f-decCanFragMap)(*d*).
    Auxiliary Functions for Binary Floating-point Lexical/Canonical Mappings **<a id="f-floatPtRound">*floatingPointRound*</a>** (*nV*, *cWidth*, *eMin*, *eMax*) → decimal number or [special value](#dt-specialvalue) Rounds a non-zero decimal number to the nearest floating-point value.  **Arguments:**

<table><tbody>
    <tr><td>nV</td><td>: </td><td>an initially non-zero decimal number (may be
        set to zero during calculations)</td></tr>
    <tr><td>cWidth</td><td>: </td><td>a positive integer</td></tr>
    <tr><td>eMin</td><td>: </td><td>an integer</td></tr>
    <tr><td>eMax</td><td>: </td><td>an integer greater than eMin</td></tr>

    </tbody></table>

 **Result:**a decimal number or [special value](#dt-specialvalue) *(***INF*** or −***INF***)* **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            s be an integer
                initially 1,c be a nonnegative integer, ande be an integer.
        </td></tr></tbody></table>

1. Set *s* to −1   when  *nV* < 0 .
2. So select *e* that 2^*cWidth*^ × 2^(*e*−1)^ ≤ |*nV*| < 2^*cWidth*^ × 2^*e*^ .
3. So select *c* that  (*c* − 1) × 2^*e*^ ≤ |*nV* | <*c* × 2^*e*^   .
4.

  - when  *eMax* < *e*  *(overflow)* return:

    - ***positiveInfinity***   when *s* is positive, and
    - ***negativeInfinity***   otherwise.

  - otherwise:

    1. When *e* < *eMin*  *(underflow):*

      - Set  *e* = *eMin*
      - So select *c* that  (*c* − 1) × 2^*e*^ ≤ |*nV* | <*c* × 2^*e*^ .

    2. Set *nV* to

      - *c* × 2^*e*^   when  |*nV* | > *c* × 2^*e*^ − 2^(*e*−1)^ ;
      - (*c* − 1) × 2^e^   when  |*nV* | < *c* × 2^*e*^ − 2^(*e*−1)^ ;
      - *c* × 2^*e*^ or (*c* − 1) × 2^*e*^   according to whether *c* is even or  *c* − 1  is even, otherwise (i.e.,  |*nV* | = *c* × 2^*e*^ − 2^(*e*−1)^ , the midpoint between the two values).

    3. Return 

      - *s* × *nV*   when *nV* < 2^*cWidth*^ × 2^*eMax*^,
      - ***positiveInfinity***   when *s* is positive, and
      - ***negativeInfinity***   otherwise.

 **Note:** Implementers will find the algorithms of [[Clinger, WD (1990)]](#clinger1990) more efficient in memory than the simple abstract algorithm employed above.  **<a id="f-round">*round*</a>** (*n*, *k*) → decimal number Maps a decimal number to that value rounded by some power of 10.  **Arguments:**

<table><tbody>
    <tr><td>n</td><td>: </td><td>a decimal number</td></tr>
    <tr><td>k</td><td>: </td><td>a nonnegative integer</td></tr>

    </tbody></table>

 **Result:**a decimal number **Algorithm:** Return  ((*n* / 10^k^ + 0.5) [div](#dt-div)1) × 10^k^ . **<a id="f-floatApprox">*floatApprox*</a>** (*c*, *e*, *j*) → decimal number Maps a decimal number ( *c* × 10^e^ ) to successive approximations.  **Arguments:**

<table><tbody>
    <tr><td>c</td><td>: </td><td>a nonnegative integer</td></tr>
    <tr><td>e</td><td>: </td><td>an integer</td></tr>
    <tr><td>j</td><td>: </td><td>a nonnegative integer</td></tr>

    </tbody></table>

 **Result:**a decimal number **Algorithm:** Return  [*round*](#f-round)(*c*, *j* ) × 10^*e*^   Lexical Mapping **<a id="f-floatLexmap">*floatLexicalMap*</a>** (*LEX*) → [float](#float) Maps a [*floatRep*](#nt-floatRep) onto a [float](#float) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>matches <a href="#nt-floatRep">floatRep</a></td></tr>

    </tbody></table>

 **Result:**a [float](#float) value **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            nV be a decimal number or <a href="#dt-specialvalue">·special value·</a> (INF or −INF).
        </td></tr></tbody></table>

- Return [*specialRepValue*](#f-specRepVal)(*LEX*)   when *LEX* is an instance of [*numericalSpecialRep*](#nt-numSpecReps);
- otherwise (*LEX* is a numeral):

  1. Set *nV* to

    - [*noDecimalMap*](#f-noDecVal)(*LEX*)   when *LEX* is an instance of [*noDecimalPtNumeral*](#nt-noDecNuml),
    - [*decimalPtMap*](#f-decVal)(*LEX*)   when *LEX* is an instance of [*decimalPtNumeral*](#nt-decNuml), and
    - [*scientificMap*](#f-sciVal)(*LEX*)   otherwise (*LEX* is an instance of [*scientificNotationNumeral*](#nt-sciNuml)).

  2. Set *nV* to [*floatingPointRound*](#f-floatPtRound)(*nV*, 24, −149, 104)   when *nV* is not zero.  *([*floatingPointRound*](#f-floatPtRound) may nonetheless return zero, or INF or −INF.)*
  3. Return:

    - When *nV* is zero:

      - ***negativeZero***   when the first character of *LEX* is '`-`', and
      - ***positiveZero***   otherwise.

    - *nV*   otherwise.

 **Note:** This specification permits the substitution of any other rounding algorithm which conforms to the requirements of [[IEEE 754-2008]](#ieee754-2008).    Lexical Mapping **<a id="f-doubleLexmap">*doubleLexicalMap*</a>** (*LEX*) → [double](#double) Maps a [*doubleRep*](#nt-doubleRep) onto a [double](#double) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>matches <a href="#nt-doubleRep">doubleRep</a></td></tr>

    </tbody></table>

 **Result:**a [double](#double) value **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            nV be a decimal number or <a href="#dt-specialvalue">·special value·</a> (INF or −INF).
        </td></tr></tbody></table>

- Return [*specialRepValue*](#f-specRepVal)(*LEX*)   when *LEX* is an instance of [*numericalSpecialRep*](#nt-numSpecReps);
- otherwise (*LEX* is a numeral):

  1. Set *nV* to

    - [*noDecimalMap*](#f-noDecVal)(*LEX*)   when *LEX* is an instance of [*noDecimalPtNumeral*](#nt-noDecNuml),
    - [*decimalPtMap*](#f-decVal)(*LEX*)   when *LEX* is an instance of [*decimalPtNumeral*](#nt-decNuml), and
    - [*scientificMap*](#f-sciVal)(*LEX*)   otherwise (*LEX* is an instance of [*scientificNotationNumeral*](#nt-sciNuml)).

  2. Set *nV* to [*floatingPointRound*](#f-floatPtRound)(*nV*, 53, −1074, 971)   when *nV* is not zero.  *([*floatingPointRound*](#f-floatPtRound) may nonetheless return zero, or INF or −INF.)*
  3. Return:

    - When *nV* is zero:

      - ***negativeZero***   when the first character of *LEX* is '`-`', and
      - ***positiveZero***   otherwise.

    - *nV*   otherwise.

 **Note:** This specification permits the substitution of any other rounding algorithm which conforms to the requirements of [[IEEE 754-2008]](#ieee754-2008).    Canonical Mapping **<a id="f-floatCanmap">*floatCanonicalMap*</a>** (*f*) → [*floatRep*](#nt-floatRep) Maps a [float](#float) to its [canonical representation](#dt-canonical-representation), a [*floatRep*](#nt-floatRep).  **Arguments:**

<table><tbody>
    <tr><td>f</td><td>: </td><td>a <a href="#float">float</a> value</td></tr>

    </tbody></table>

 **Result:**a [literal](#dt-literal) matching [*floatRep*](#nt-floatRep) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            l be a nonnegative integers be an integer intially 1,c be a positive integer, ande be an integer.
        </td></tr></tbody></table>

- Return [*specialRepCanonicalMap*](#f-specValCanMap)(*f* )   when *f* is one of ***positiveInfinity***, ***negativeInfinity***, or ***notANumber***;
- return '`0.0E0`'   when *f* is ***positiveZero***;
- return '`-0.0E0`'   when *f* is ***negativeZero***;
- otherwise (*f* is numeric and non-zero):

  1. Set *s* to −1   when  *f* < 0 .
  2. Let *c* be the smallest integer for which there exists an integer *e* for which  |*f* | = *c* × 10^*e*^ .
  3. Let *e* be log~10~(|*f* | / *c*)   (so that  |*f* | = *c* × 10^*e*^ ).
  4. Let *l* be the largest nonnegative integer for which  *c* × 10^*e*^ = [*floatingPointRound*](#f-floatPtRound)([*floatApprox*](#f-floatApprox)(*c*, *e*, *l* ), 24, −149, 104)
  5. Return [*scientificCanonicalMap*](#f-sciCanFragMap)(*s* × [*floatApprox*](#f-floatApprox)(*c*, *e*, *l* )) .

    Canonical Mapping **<a id="f-doubleCanmap">*doubleCanonicalMap*</a>** (*f*) → [*doubleRep*](#nt-doubleRep) Maps a [double](#double) to its [canonical representation](#dt-canonical-representation), a [*doubleRep*](#nt-doubleRep).  **Arguments:**

<table><tbody>
    <tr><td>f</td><td>: </td><td>a <a href="#double">double</a> value</td></tr>

    </tbody></table>

 **Result:**a [literal](#dt-literal) matching [*doubleRep*](#nt-doubleRep) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            l be a nonnegative integers be an integer intially 1,c be a positive integer, ande be an integer.
        </td></tr></tbody></table>

- Return [*specialRepCanonicalMap*](#f-specValCanMap)(*f* )   when *f* is one of ***positiveInfinity***, ***negativeInfinity***, or ***notANumber***;
- return '`0.0E0`'   when *f* is ***positiveZero***;
- return '`-0.0E0`'   when *f* is ***negativeZero***;
- otherwise (*f* is numeric and non-zero):

  1. Set *s* to −1   when  *f* < 0 .
  2. Let *c* be the smallest integer for which there exists an integer *e* for which  |*f* | = *c* × 10^*e*^ .
  3. Let *e* be log~10~(|*f* | / *c*)   (so that  |*f* | = *c* × 10^*e*^ ).
  4. Let *l* be the largest nonnegative integer for which  *c* × 10^*e*^ = [*floatingPointRound*](#f-floatPtRound)([*floatApprox*](#f-floatApprox)(*c*, *e*, *l* ), 53, −1074, 971)
  5. Return [*scientificCanonicalMap*](#f-sciCanFragMap)(*s* × [*floatApprox*](#f-floatApprox)(*c*, *e*, *l* )) .

### <a id="sec-duration-functions"></a>E.2 Duration-related Definitions

The following functions are primarily used with the [duration](#duration) datatype and its derivatives.   Auxiliary [duration](#duration)-related Functions Operating on Representation Fragments **<a id="f-duYrMap">*duYearFragmentMap*</a>** (*Y*) → integer Maps a [*duYearFrag*](#nt-duYrFrag) to an integer, intended as part of the value of the [*months*](#vp-du-month) property of a [duration](#duration) value.  **Arguments:**

<table><tbody>
        <tr><td>Y</td><td>: </td><td>matches <a href="#nt-duYrFrag">duYearFrag</a></td></tr>

        </tbody></table>

 **Result:**a nonnegative integer **Algorithm:** *Y* is necessarily the letter '`Y`' followed by a numeral *N*: Return [*noDecimalMap*](#f-noDecVal)(*N*). **<a id="f-duMoMap">*duMonthFragmentMap*</a>** (*M*) → integer Maps a [*duMonthFrag*](#nt-duMoFrag) to an integer, intended as part of the value of the [*months*](#vp-du-month) property of a [duration](#duration) value.  **Arguments:**

<table><tbody>
        <tr><td>M</td><td>: </td><td>matches <a href="#nt-duYrFrag">duYearFrag</a></td></tr>

        </tbody></table>

 **Result:**a nonnegative integer **Algorithm:** *M* is necessarily the letter '`M`' followed by a numeral *N*: Return [*noDecimalMap*](#f-noDecVal)(*N*). **<a id="f-duDaMap">*duDayFragmentMap*</a>** (*D*) → integer Maps a [*duDayFrag*](#nt-duDaFrag) to an integer, intended as part of the value of the [*seconds*](#vp-du-second) property of a [duration](#duration) value.  **Arguments:**

<table><tbody>
        <tr><td>D</td><td>: </td><td>matches <a href="#nt-duDaFrag">duDayFrag</a></td></tr>

        </tbody></table>

 **Result:**a nonnegative integer **Algorithm:** *D* is necessarily the letter '`D`' followed by a numeral *N*: Return [*noDecimalMap*](#f-noDecVal)(*N*). **<a id="f-duHrMap">*duHourFragmentMap*</a>** (*H*) → integer Maps a [*duHourFrag*](#nt-duHrFrag) to an integer, intended as part of the value of the [*seconds*](#vp-du-second) property of a [duration](#duration) value.  **Arguments:**

<table><tbody>
        <tr><td>H</td><td>: </td><td>matches <a href="#nt-duHrFrag">duHourFrag</a></td></tr>

        </tbody></table>

 **Result:**a nonnegative integer **Algorithm:** *D* is necessarily the letter '`D`' followed by a numeral *N*: Return [*noDecimalMap*](#f-noDecVal)(*N*). **<a id="f-duMiMap">*duMinuteFragmentMap*</a>** (*M*) → integer Maps a [*duMinuteFrag*](#nt-duMiFrag) to an integer, intended as part of the value of the [*seconds*](#vp-du-second) property of a [duration](#duration) value.  **Arguments:**

<table><tbody>
        <tr><td>M</td><td>: </td><td>matches <a href="#nt-duMiFrag">duMinuteFrag</a></td></tr>

        </tbody></table>

 **Result:**a nonnegative integer **Algorithm:** *M* is necessarily the letter '`M`' followed by a numeral *N*: Return [*noDecimalMap*](#f-noDecVal)(*N*). **<a id="f-duSeMap">*duSecondFragmentMap*</a>** (*S*) → decimal number Maps a [*duSecondFrag*](#nt-duSeFrag) to a decimal number, intended as part of the value of the [*seconds*](#vp-du-second) property of a [duration](#duration) value.  **Arguments:**

<table><tbody>
        <tr><td>S</td><td>: </td><td>matches <a href="#nt-duSeFrag">duSecondFrag</a></td></tr>

        </tbody></table>

 **Result:**a nonnegative decimal number **Algorithm:** *S* is necessarily '`S`' followed by a numeral *N*: Return

- [*decimalPtMap*](#f-decVal)(*N*)   when '`.`' occurs in *N*, and
- [*noDecimalMap*](#f-noDecVal)(*N*)   otherwise.
  **<a id="f-duYMMap">*duYearMonthFragmentMap*</a>** (*YM*) → integer Maps a [*duYearMonthFrag*](#nt-duYMFrag) into an integer, intended as part of the [*months*](#vp-du-month) property of a [duration](#duration) value.  **Arguments:**

<table><tbody>
        <tr><td>YM</td><td>: </td><td>matches <a href="#nt-duYMFrag">duYearMonthFrag</a></td></tr>

        </tbody></table>

 **Result:**a nonnegative integer **Algorithm:** *YM* necessarily consists of an instance *Y* of [*duYearFrag*](#nt-duYrFrag) and/or an instance *M* of [*duMonthFrag*](#nt-duMoFrag):

<table><tbody><tr><td>Let </td><td>
                y be <a href="#f-duYrMap">·duYearFragmentMap·</a>(Y) (or 0 if Y is not present) andm be <a href="#f-duMoMap">·duMonthFragmentMap·</a>(M) (or 0 if M is not present).
            </td></tr></tbody></table>

 Return  12 × *y* + *m* . **<a id="f-duTMap">*duTimeFragmentMap*</a>** (*T*) → decimal number Maps a [*duTimeFrag*](#nt-duTFrag) into a decimal number, intended as part of the [*seconds*](#vp-du-second) property of a [duration](#duration) value.  **Arguments:**

<table><tbody>
        <tr><td>T</td><td>: </td><td>matches <a href="#nt-duTFrag">duTimeFrag</a></td></tr>

        </tbody></table>

 **Result:**a nonnegative decimal number **Algorithm:** *T* necessarily consists of an instance *H* of [*duHourFrag*](#nt-duHrFrag), and/or an instance *M* of [*duMinuteFrag*](#nt-duMiFrag), and/or an instance *S* of [*duSecondFrag*](#nt-duSeFrag).

<table><tbody><tr><td>Let </td><td>
                h be <a href="#f-duDaMap">·duDayFragmentMap·</a>(H)
                    (or 0 if H is not present),m be <a href="#f-duMiMap">·duMinuteFragmentMap·</a>(M)
                    (or 0 if M is not present), ands be <a href="#f-duSeMap">·duSecondFragmentMap·</a>(S)
                    (or 0 if S is not present).
            </td></tr></tbody></table>

 Return  3600 × *h* + 60 × *m* + s . **<a id="f-duDTMap">*duDayTimeFragmentMap*</a>** (*DT*) → decimal number Maps a [*duDayTimeFrag*](#nt-duDTFrag) into a decimal number, which is the potential value of the [*seconds*](#vp-du-second) property of a [duration](#duration) value.  **Arguments:**

<table><tbody>
        <tr><td>DT</td><td>: </td><td>matches <a href="#nt-duDTFrag">duDayTimeFrag</a></td></tr>

        </tbody></table>

 **Result:**a nonnegative decimal number **Algorithm:** *DT* necessarily consists of an instance *D* of [*duDayFrag*](#nt-duDaFrag) and/or an instance *T* of [*duTimeFrag*](#nt-duTFrag).

<table><tbody><tr><td>Let </td><td>
                d be <a href="#f-duDaMap">·duDayFragmentMap·</a>(D)
                    (or 0 if D is not present) andt be <a href="#f-duTMap">·duTimeFragmentMap·</a>(T)
                    (or 0 if T is not present).
            </td></tr></tbody></table>

 Return  86400 × *d* + *t* .    The [duration](#duration) Lexical Mapping **<a id="f-durationMap">*durationMap*</a>** (*DUR*) → [duration](#duration) Separates the [*durationLexicalRep*](#nt-durationRep) into the month part and the seconds part, then maps them into the [*months*](#vp-du-month) and [*seconds*](#vp-du-second) of the [duration](#duration) value.  **Arguments:**

<table><tbody>
        <tr><td>DUR</td><td>: </td><td>matches <a href="#nt-durationRep">durationLexicalRep</a></td></tr>

        </tbody></table>

 **Result:**a complete [duration](#duration) value **Algorithm:** *DUR* consists of possibly a leading '`-`', followed by '`P`' and then an instance *Y* of [*duYearMonthFrag*](#nt-duYMFrag) and/or an instance *D* of [*duDayTimeFrag*](#nt-duDTFrag): Return a [duration](#duration) whose

- [*months*](#vp-du-month) value is

  - 0   if *Y* is not present,
  - −[*duYearMonthFragmentMap*](#f-duYMMap)(*Y*)   if both '`-`' and *Y* are present, and
  - [*duYearMonthFragmentMap*](#f-duYMMap)(*Y*)   otherwise.

 and whose

- [*seconds*](#vp-du-second) value is

  - 0   if *D* is not present,
  - −[*duDayTimeFragmentMap*](#f-duDTMap)(*D*)   if both '`-`' and *D* are present, and
  - [*duDayTimeFragmentMap*](#f-duDTMap)(*D*)   otherwise.

     The [yearMonthDuration](#yearMonthDuration) Lexical Mapping **<a id="f-yearMonthDurationMap">*yearMonthDurationMap*</a>** (*YM*) → [yearMonthDuration](#yearMonthDuration) Maps the lexical representation into the [*months*](#vp-du-month) of a [yearMonthDuration](#yearMonthDuration) value.  (A [yearMonthDuration](#yearMonthDuration)'s [*seconds*](#vp-du-second) is always zero.)  [*yearMonthDurationMap*](#f-yearMonthDurationMap) is a restriction of [*durationMap*](#f-durationMap).  **Arguments:**

<table><tbody>
        <tr><td>YM</td><td>: </td><td>matches <a href="#nt-yearMonthDurationRep">yearMonthDurationLexicalRep</a></td></tr>

        </tbody></table>

 **Result:**a complete [yearMonthDuration](#yearMonthDuration) value **Algorithm:** *YM* necessarily consists of an optional leading '`-`', followed by '`P`' and then an instance *Y* of [*duYearMonthFrag*](#nt-duYMFrag): Return a [yearMonthDuration](#yearMonthDuration) whose

- [*months*](#vp-du-month) value is

  - −[*duYearMonthFragmentMap*](#f-duYMMap)(*Y*)   if '`-`' is present in *YM* and
  - [*duYearMonthFragmentMap*](#f-duYMMap)(*Y*)   otherwise, and

- [*seconds*](#vp-du-second) value is (necessarily) 0.
     The [dayTimeDuration](#dayTimeDuration) Lexical Mapping **<a id="f-dayTimeDurationMap">*dayTimeDurationMap*</a>** (*DT*) → [dayTimeDuration](#dayTimeDuration) Maps the lexical representation into the [*seconds*](#vp-du-second) of a [dayTimeDuration](#dayTimeDuration) value.  (A [dayTimeDuration](#dayTimeDuration)'s [*months*](#vp-du-month) is always zero.)  [*dayTimeDurationMap*](#f-dayTimeDurationMap) is a restriction of [*durationMap*](#f-durationMap).  **Arguments:**

<table><tbody>
        <tr><td>DT</td><td>: </td><td>a <a href="#dayTimeDuration">dayTimeDuration</a> value</td></tr>

        </tbody></table>

 **Result:**a complete [dayTimeDuration](#dayTimeDuration) value **Algorithm:** *DT* necessarily consists of possibly a leading '`-`', followed by '`P`' and then an instance *D* of [*duDayTimeFrag*](#nt-duDTFrag): Return a [dayTimeDuration](#dayTimeDuration) whose

- [*months*](#vp-du-month) value is (necessarily) 0, and
- [*seconds*](#vp-du-second) value is

  - −[*duDayTimeFragmentMap*](#f-duDTMap)(*D*)   if '`-`' is present in *DT* and
  - [*duDayTimeFragmentMap*](#f-duDTMap)(*D*)   otherwise.

     Auxiliary [duration](#duration)-related Functions Producing Representation Fragments **<a id="f-duYMCan">*duYearMonthCanonicalFragmentMap*</a>** (*ym*) → [*duYearMonthFrag*](#nt-duYMFrag) Maps a nonnegative integer, presumably the absolute value of the [*months*](#vp-du-month) of a [duration](#duration) value, to a [*duYearMonthFrag*](#nt-duYMFrag), a fragment of a [duration](#duration) [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
        <tr><td>ym</td><td>: </td><td>a nonnegative integer</td></tr>

        </tbody></table>

 **Result:**a [literal](#dt-literal) matching [*duYearMonthFrag*](#nt-duYMFrag) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
                y be  ym <a href="#dt-div">·div·</a> 12 , andm be  ym <a href="#dt-mod">·mod·</a> 12 ,
            </td></tr></tbody></table>

 Return

- [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(*y*) & '`Y`' & [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(*m*) & '`M`'   when neither *y* nor *m* is zero,
- [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(*y*) & '`Y`'   when *y* is not zero but *m* is, and
-  [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(*m*) & '`M`'   when *y* is zero.
  **<a id="f-duDCan">*duDayCanonicalFragmentMap*</a>** (*d*) → [*duDayFrag*](#nt-duDaFrag) Maps a nonnegative integer, presumably the day normalized value from the [*seconds*](#vp-du-second) of a [duration](#duration) value, to a [*duDayFrag*](#nt-duDaFrag), a fragment of a [duration](#duration) [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
        <tr><td>d</td><td>: </td><td>a nonnegative integer</td></tr>

        </tbody></table>

 **Result:**a [literal](#dt-literal) matching [*duDayFrag*](#nt-duDaFrag) **Algorithm:** Return

- [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(*d*) & '`D`'   when *d* is not zero, and
- the empty string ('')   when *d* is zero.
  **<a id="f-duHCan">*duHourCanonicalFragmentMap*</a>** (*h*) → [*duHourFrag*](#nt-duHrFrag) Maps a nonnegative integer, presumably the hour normalized value from the [*seconds*](#vp-du-second) of a [duration](#duration) value, to a [*duHourFrag*](#nt-duHrFrag), a fragment of a [duration](#duration) [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
        <tr><td>h</td><td>: </td><td>a nonnegative integer</td></tr>

        </tbody></table>

 **Result:**a [literal](#dt-literal) matching [*duHourFrag*](#nt-duHrFrag) **Algorithm:** Return

- [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(*h*) & '`H`'   when *h* is not zero, and
- the empty string ('')   when *h* is zero.
  **<a id="f-duMCan">*duMinuteCanonicalFragmentMap*</a>** (*m*) → [*duMinuteFrag*](#nt-duMiFrag) Maps a nonnegative integer, presumably the minute normalized value from the [*seconds*](#vp-du-second) of a [duration](#duration) value, to a [*duMinuteFrag*](#nt-duMiFrag), a fragment of a [duration](#duration) [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
        <tr><td>m</td><td>: </td><td>a nonnegative integer</td></tr>

        </tbody></table>

 **Result:**a [literal](#dt-literal) matching [*duMinuteFrag*](#nt-duMiFrag) **Algorithm:** Return

- [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(*m*) & '`M`'   when *m* is not zero, and
- the empty string ('')   when *m* is zero.
  **<a id="f-duSCan">*duSecondCanonicalFragmentMap*</a>** (*s*) → [*duSecondFrag*](#nt-duSeFrag) Maps a nonnegative decimal number, presumably the second normalized value from the [*seconds*](#vp-du-second) of a [duration](#duration) value, to a [*duSecondFrag*](#nt-duSeFrag), a fragment of a [duration](#duration) [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
        <tr><td>s</td><td>: </td><td>a nonnegative decimal number</td></tr>

        </tbody></table>

 **Result:**matches [*duSecondFrag*](#nt-duSeFrag) **Algorithm:** Return

- [*unsignedNoDecimalPtCanonicalMap*](#f-unsNoDecCanFragMap)(*s*) & '`S`'  when *s* is a non-zero integer,
- [*unsignedDecimalPtCanonicalMap*](#f-unsDecCanFragMap)(*s*) & '`S`'  when *s* is not an integer, and
- the empty string ('') when *s* is zero.
  **<a id="f-duTCan">*duTimeCanonicalFragmentMap*</a>** (*h*, *m*, *s*) → [*duTimeFrag*](#nt-duTFrag) Maps three nonnegative numbers, presumably the hour, minute, and second normalized values from a [duration](#duration)'s [*seconds*](#vp-du-second), to a [*duTimeFrag*](#nt-duTFrag), a fragment of a [duration](#duration) [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
        <tr><td>h</td><td>: </td><td>a nonnegative integer</td></tr>
        <tr><td>m</td><td>: </td><td>a nonnegative integer</td></tr>
        <tr><td>s</td><td>: </td><td>a nonnegative decimal number</td></tr>

        </tbody></table>

 **Result:**a [literal](#dt-literal) matching [*duTimeFrag*](#nt-duTFrag) **Algorithm:** Return

- '`T`' & [*duHourCanonicalFragmentMap*](#f-duHCan)(*h*) & [*duMinuteCanonicalFragmentMap*](#f-duMCan)(*m*) & [*duSecondCanonicalFragmentMap*](#f-duSCan)(*s*)   when *h*, *m*, and *s* are not all zero, and
- the empty string ('') when all arguments are zero.
  **<a id="f-duDTCan">*duDayTimeCanonicalFragmentMap*</a>** (*ss*) → [*duDayTimeFrag*](#nt-duDTFrag) Maps a nonnegative decimal number, presumably the absolute value of the [*seconds*](#vp-du-second) of a [duration](#duration) value, to a [*duDayTimeFrag*](#nt-duDTFrag), a fragment of a [duration](#duration) [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
        <tr><td>ss</td><td>: </td><td>a nonnegative decimal number</td></tr>

        </tbody></table>

 **Result:**matches [*duDayTimeFrag*](#nt-duDTFrag) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
                d is 
                    ss <a href="#dt-div">·div·</a> 86400 ,h is 
                    (ss <a href="#dt-mod">·mod·</a> 86400) <a href="#dt-div">·div·</a> 3600 ,m is 
                    (ss <a href="#dt-mod">·mod·</a> 3600) <a href="#dt-div">·div·</a> 60 , ands is 
                    ss <a href="#dt-mod">·mod·</a> 60 ,
            </td></tr></tbody></table>

 Return

- [*duDayCanonicalFragmentMap*](#f-duDCan)(*d*) & [*duTimeCanonicalFragmentMap*](#f-duTCan)(*h*, *m*, *s*)   when *ss* is not zero and
- '`T0S`'   when *ss* is zero.
     The [duration](#duration) Canonical Mapping **<a id="f-durationCanMap">*durationCanonicalMap*</a>** (*v*) → [*durationLexicalRep*](#nt-durationRep) Maps a [duration](#duration)'s property values to [*durationLexicalRep*](#nt-durationRep) fragments and combines the fragments into a complete [*durationLexicalRep*](#nt-durationRep).  **Arguments:**

<table><tbody>
        <tr><td>v</td><td>: </td><td>a complete <a href="#duration">duration</a> value</td></tr>

        </tbody></table>

 **Result:**matches [*durationLexicalRep*](#nt-durationRep) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
                m be v's <a href="#vp-du-month">·months·</a>,s be v's <a href="#vp-du-second">·seconds·</a>, andsgn be '<code>-</code>' if m or
                    s is negative and
                    the empty string ('') otherwise.
            </td></tr></tbody></table>

 Return

- *sgn* & '`P`' & [*duYearMonthCanonicalFragmentMap*](#f-duYMCan)(| *m* |) & [*duDayTimeCanonicalFragmentMap*](#f-duDTCan)(| *s* |)    when neither *m* nor *s* is zero,
- *sgn* & '`P`' & [*duYearMonthCanonicalFragmentMap*](#f-duYMCan)(| *m* |)    when *m* is not zero but *s* is, and
- *sgn* & '`P`' & [*duDayTimeCanonicalFragmentMap*](#f-duDTCan)(| *s* |)    when *m* is zero.
     The [yearMonthDuration](#yearMonthDuration) Canonical Mapping **<a id="f-yearMonthDurationCanMap">*yearMonthDurationCanonicalMap*</a>** (*ym*) → [*yearMonthDurationLexicalRep*](#nt-yearMonthDurationRep) Maps a [yearMonthDuration](#yearMonthDuration)'s [*months*](#vp-du-month) value to a [*yearMonthDurationLexicalRep*](#nt-yearMonthDurationRep).  (The [*seconds*](#vp-du-second) value is necessarily zero and is ignored.)  [*yearMonthDurationCanonicalMap*](#f-yearMonthDurationCanMap) is a restriction of [*durationCanonicalMap*](#f-durationCanMap).  **Arguments:**

<table><tbody>
        <tr><td>ym</td><td>: </td><td>a complete <a href="#yearMonthDuration">yearMonthDuration</a> value</td></tr>

        </tbody></table>

 **Result:**matches [*yearMonthDurationLexicalRep*](#nt-yearMonthDurationRep) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
                m be ym's <a href="#vp-du-month">·months·</a> andsgn be '<code>-</code>' if m is negative and
                    the empty string ('') otherwise.
            </td></tr></tbody></table>

  Return  *sgn* & '`P`' & [*duYearMonthCanonicalFragmentMap*](#f-duYMCan)(| *m* |) .     The [dayTimeDuration](#dayTimeDuration) Canonical Mapping **<a id="f-dayTimeDurationCanMap">*dayTimeDurationCanonicalMap*</a>** (*dt*) → [*dayTimeDurationLexicalRep*](#nt-dayTimeDurationRep) Maps a [dayTimeDuration](#dayTimeDuration)'s [*seconds*](#vp-du-second) value to a [*dayTimeDurationLexicalRep*](#nt-dayTimeDurationRep).  (The [*months*](#vp-du-month) value is necessarily zero and is ignored.)  [*dayTimeDurationCanonicalMap*](#f-dayTimeDurationCanMap) is a restriction of [*durationCanonicalMap*](#f-durationCanMap).  **Arguments:**

<table><tbody>
        <tr><td>dt</td><td>: </td><td>a complete <a href="#dayTimeDuration">dayTimeDuration</a> value</td></tr>

        </tbody></table>

 **Result:**matches [*dayTimeDurationLexicalRep*](#nt-dayTimeDurationRep) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
                s be dt's <a href="#vp-du-month">·months·</a> andsgn be '<code>-</code>' if s is negative and
                    the empty string ('') otherwise.
            </td></tr></tbody></table>

 Return *sgn* & '`P`' & [*duYearMonthCanonicalFragmentMap*](#f-duYMCan)(| *s* |) .

### <a id="sec-dt-functions"></a>E.3 Date/time-related Definitions

        E.3.1 [Normalization of property values](#sec-normalization)
         E.3.2 [Auxiliary Functions](#sec-aux-functions)
         E.3.3 [Adding durations to dateTimes](#sec-dt-arith)
         E.3.4 [Time on timeline](#sec-timeontimeline)
         E.3.5 [Lexical mappings](#sec-dt-lexmaps)
         E.3.6 [Canonical Mappings](#sec-dt-canmaps)

#### <a id="sec-normalization"></a>E.3.1 Normalization of property values

When adding and subtracting numbers from date/time properties, the immediate results may not conform to the limits specified.  Accordingly, the following procedures are used to "normalize" potential property values to corresponding values that do conform to the appropriate limits.  Normalization is required when dealing with time zone offset changes (as when converting to [UTC](#dt-utc) from "local" values) and when adding [duration](#duration) values to or subtracting them from [dateTime](#dateTime) values.   Date/time Datatype Normalizing Procedures **<a id="f-dt-normMo">*normalizeMonth*</a>** (*yr*, *mo*) If month (*mo*) is out of range, adjust month and year (*yr*) accordingly; otherwise, make no change.  **Arguments:**

<table><tbody>
        <tr><td>yr</td><td>: </td><td>an integer</td></tr>
        <tr><td>mo</td><td>: </td><td>an integer</td></tr>

        </tbody></table>

 **Algorithm:**

1.  Add  (*mo* − 1) [div](#dt-div) 12  to *yr*.
2.  Set *mo* to  (*mo* − 1) [mod](#dt-mod) 12 + 1 .
  **<a id="f-dt-normDa">*normalizeDay*</a>** (*yr*, *mo*, *da*) If month (*mo*) is out of range, or day (*da*) is out of range for the appropriate month, then adjust values accordingly, otherwise make no change.  **Arguments:**

<table><tbody>
        <tr><td>yr</td><td>: </td><td>an integer</td></tr>
        <tr><td>mo</td><td>: </td><td>an integer</td></tr>
        <tr><td>da</td><td>: </td><td>an integer</td></tr>

        </tbody></table>

 **Algorithm:**

1. [*normalizeMonth*](#f-dt-normMo)(*yr*, *mo*)
2. Repeat until *da* is positive and not greater than [*daysInMonth*](#f-daysInMonth)(*yr*, *mo*):

  1. If *da* exceeds [*daysInMonth*](#f-daysInMonth)(*yr*, *mo*) then:

    1. Subtract that limit from *da*.
    2. Add 1 to *mo*.
    3. [*normalizeMonth*](#f-dt-normMo)(*yr*, *mo*)

  2. If *da* is not positive then:

    1. Subtract 1 from *mo*.
    2. [*normalizeMonth*](#f-dt-normMo)(*yr*, *mo*)
    3. Add the new upper limit from the table to *da*.

  **<a id="f-dt-normMi">*normalizeMinute*</a>** (*yr*, *mo*, *da*, *hr*, *mi*) Normalizes minute, hour, month, and year values to values that obey the appropriate constraints.  **Arguments:**

<table><tbody>
        <tr><td>yr</td><td>: </td><td>an integer</td></tr>
        <tr><td>mo</td><td>: </td><td>an integer</td></tr>
        <tr><td>da</td><td>: </td><td>an integer</td></tr>
        <tr><td>hr</td><td>: </td><td>an integer</td></tr>
        <tr><td>mi</td><td>: </td><td>an integer</td></tr>

        </tbody></table>

 **Algorithm:**

1. Add   *mi* [div](#dt-div) 60  to *hr*.
2. Set *mi* to  *mi* [mod](#dt-mod) 60 .
3. Add  *hr* [div](#dt-div) 24  to *da*.
4. Set *hr* to  *hr* [mod](#dt-mod) 24 .
5. [*normalizeDay*](#f-dt-normDa)(*yr*, *mo*, *da*).
  **<a id="f-dt-normSe">*normalizeSecond*</a>** (*yr*, *mo*, *da*, *hr*, *mi*, *se*) Normalizes second, minute, hour, month, and year values to values that obey the appropriate constraints.  (This algorithm ignores leap seconds.)   **Arguments:**

<table><tbody>
        <tr><td>yr</td><td>: </td><td>an integer</td></tr>
        <tr><td>mo</td><td>: </td><td>an integer</td></tr>
        <tr><td>da</td><td>: </td><td>an integer</td></tr>
        <tr><td>hr</td><td>: </td><td>an integer</td></tr>
        <tr><td>mi</td><td>: </td><td>an integer</td></tr>
        <tr><td>se</td><td>: </td><td>a decimal number</td></tr>

        </tbody></table>

 **Algorithm:**

1. Add  *se* [div](#dt-div) 60  to *mi*.
2. Set *se* to *se* [mod](#dt-mod) 60 .
3. [*normalizeMinute*](#f-dt-normMi)(*yr*, *mo*, *da*, *hr*, *mi*).

#### <a id="sec-aux-functions"></a>E.3.2 Auxiliary Functions

   Date/time Auxiliary Functions **<a id="f-daysInMonth">*daysInMonth*</a>** (*y*, *m*) → integer Returns the number of the last day of the month for any combination of year and month.  **Arguments:**

<table><tbody>
        <tr><td>y</td><td>: </td><td>an <a href="#dt-optional">·optional·</a> integer</td></tr>
        <tr><td>m</td><td>: </td><td>an integer between 1 and 12</td></tr>

        </tbody></table>

 **Result:**between 28 and 31 inclusive **Algorithm:**  Return:

- 28   when *m* is 2 and *y* is not evenly divisible by 4, or is evenly divisible by 100 but not by 400, or is ***absent***,
- 29   when *m* is 2 and *y* is evenly divisible by 400, or is evenly divisible by 4 but not by 100,
- 30   when *m* is 4, 6, 9, or 11,
- 31   otherwise (*m* is 1, 3, 5, 7, 8, 10, or 12)
  **<a id="p-setDTFromRaw">*newDateTime*</a>** (*Yr*, *Mo*, *Da*, *Hr*, *Mi*, *Se*, *Tz*) → an instance of the [date/timeSevenPropertyModel](#dt-dt-7PropMod) Returns an instance of the [date/timeSevenPropertyModel](#dt-dt-7PropMod) with property values as specified in the arguments. If an argument is omitted, the corresponding property is set to ***absent***.  **Arguments:**

<table><tbody>
        <tr><td>Yr</td><td>: </td><td>an <a href="#dt-optional">·optional·</a>
            integer</td></tr>
        <tr><td>Mo</td><td>: </td><td>an <a href="#dt-optional">·optional·</a>
            integer between 1 and 12 inclusive</td></tr>
        <tr><td>Da</td><td>: </td><td>an <a href="#dt-optional">·optional·</a>
            integer between 1 and 31 inclusive</td></tr>
        <tr><td>Hr</td><td>: </td><td>an <a href="#dt-optional">·optional·</a>
            integer between 0 and 24 inclusive</td></tr>
        <tr><td>Mi</td><td>: </td><td>an <a href="#dt-optional">·optional·</a>
            integer between 0 and 59 inclusive</td></tr>
        <tr><td>Se</td><td>: </td><td>an <a href="#dt-optional">·optional·</a>
            decimal number greater than or equal to 0 and less than
            60</td></tr>
        <tr><td>Tz</td><td>: </td><td>an <a href="#dt-optional">·optional·</a>
            integer
            between −840 and 840 inclusive.</td></tr>

        </tbody></table>

 **Result:**  **Algorithm:**

<table><tbody><tr><td>Let </td><td>
                dt be an instance of the <a href="#dt-dt-7PropMod">date/timeSevenPropertyModel</a>yr
                    be Yr when Yr is not absent, otherwise 1mo
                    be Mowhen Mo is
                    not absent, otherwise 1da
                    be Dawhen Da is
                    not absent, otherwise 1hr
                    be Hrwhen Hr is
                    not absent, otherwise 0mi
                    be Miwhen Mi is
                    not absent, otherwise 0se
                    be Sewhen Se is
                    not absent, otherwise 0
            </td></tr></tbody></table>

1. [*normalizeSecond*](#f-dt-normSe)(*yr*, *mo*, *da*, *hr*, *mi*, *se*)
2. Set the [*year*](#vp-dt-year) property of *dt* to ***absent*** when *Yr* is ***absent***, otherwise *yr*.
3. Set the [*month*](#vp-dt-month) property of *dt* to ***absent*** when *Mo* is ***absent***, otherwise *mo*.
4. Set the [*day*](#vp-dt-day) property of *dt* to ***absent*** when *Da* is ***absent***, otherwise *da*.
5. Set the [*hour*](#vp-dt-hour) property of *dt* to ***absent*** when *Hr* is ***absent***, otherwise *hr*.
6. Set the [*minute*](#vp-dt-minute) property of *dt* to ***absent*** when *Mi* is ***absent***, otherwise *mi*.
7. Set the [*second*](#vp-dt-second) property of *dt* to ***absent*** when *Se* is ***absent***, otherwise *se*.
8. Set the [*timezoneOffset*](#vp-dt-timezone) property of *dt* to *Tz*
9. Return *dt*.

#### <a id="sec-dt-arith"></a>E.3.3 Adding durations to dateTimes

 Given a [dateTime](#dateTime) *S* and a [duration](#duration) *D*, function [*dateTimePlusDuration*](#vp-dt-dateTimePlusDuration) specifies how to compute a [dateTime](#dateTime) *E*, where *E* is the end of the time period with start *S* and duration *D* i.e. *E* = *S* + *D*.  Such computations are used, for example, to determine whether a [dateTime](#dateTime) is within a specific time period.  This algorithm can also be applied, when applications need the operation, to the addition of [duration](#duration)s to the datatypes [date](#date), [gYearMonth](#gYearMonth), [gYear](#gYear), [gDay](#gDay) and [gMonth](#gMonth), each of which can be viewed as denoting a set of [dateTime](#dateTime)s. In such cases, the addition is made to the first or starting [dateTime](#dateTime) in the set.  Note that the extension of this algorithm to types other than [dateTime](#dateTime) is not needed for schema-validity assessment.

 Essentially, this calculation adds the [*months*](#vp-du-month) and [*seconds*](#vp-du-second) properties of the [duration](#duration) value separately to the [dateTime](#dateTime) value. The [*months*](#vp-du-month) value is added to the starting [dateTime](#dateTime) value first. If the day is out of range for the new month value, it is *pinned* to be within range. Thus April 31 turns into April 30. Then the [*seconds*](#vp-du-second) value is added. This latter addition can cause the year, month, day, hour, and minute to change.

 Leap seconds are ignored by the computation. All calculations use 60 seconds per minute.

 Thus the addition of either PT1M or PT60S to any dateTime will always produce the same result. This is a special definition of addition which is designed to match common practice, and—most importantly—be stable over time.

 A definition that attempted to take leap-seconds into account would need to be constantly updated, and could not predict the results of future implementation's additions. The decision to introduce a leap second in [UTC](#dt-utc) is the responsibility of the [[International Earth Rotation Service (IERS)]](#IERS). They make periodic announcements as to when leap seconds are to be added, but this is not known more than a year in advance. For more information on leap seconds, see [[U.S. Naval Observatory Time Service Department]](#USNavy).

  Adding [duration](#duration) to [dateTime](#dateTime) **<a id="vp-dt-dateTimePlusDuration">*dateTimePlusDuration*</a>** (*du*, *dt*) → [dateTime](#dateTime) Adds a [duration](#duration) to a [dateTime](#dateTime) value, producing another [dateTime](#dateTime) value.  **Arguments:**

<table><tbody>
    <tr><td>du</td><td>: </td><td>a <a href="#duration">duration</a>
        value</td></tr>
    <tr><td>dt</td><td>: </td><td>a <a href="#dateTime">dateTime</a>
        value</td></tr>

    </tbody></table>

 **Result:**a [dateTime](#dateTime) value **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            yr be dt's <a href="#vp-dt-year">·year·</a>,mo be dt's <a href="#vp-dt-month">·month·</a>,da be dt's <a href="#vp-dt-day">·day·</a>,hr be dt's <a href="#vp-dt-hour">·hour·</a>,mi be dt's <a href="#vp-dt-minute">·minute·</a>, andse be dt's <a href="#vp-dt-second">·second·</a>.tz be dt's <a href="#vp-dt-timezone">·timezoneOffset·</a>.
        </td></tr></tbody></table>

1. Add *du*'s [*months*](#vp-du-month) to *mo*.
2. [*normalizeMonth*](#f-dt-normMo)(*yr*, *mo*). (I.e., carry any over- or underflow, adjust month.)
3. Set *da* to  min(*da*, [*daysInMonth*](#f-daysInMonth)(*yr*, *mo*)). (I.e., *pin* the value if necessary.)
4. Add *du*'s [*seconds*](#vp-du-second) to *se*.
5. [*normalizeSecond*](#f-dt-normSe)(*yr*, *mo*, *da*, *hr*, *mi*, *se*). (I.e., carry over- or underflow of seconds up to minutes, hours, etc.)
6. Return [*newDateTime*](#p-setDTFromRaw)(*yr*, *mo*, *da*, *hr*, *mi*, *se*, *tz*)

This algorithm may be applied to date/time types other than [dateTime](#dateTime), by

1. For each ***absent*** property, supply the minimum legal value for that property (1 for years, months, days, 0 for hours, minutes, seconds).
2. Call the function.
3. For each property ***absent*** in the initial value, set the corresponding property in the result value to ***absent***.

*Examples:*

<table><tbody><tr><th>dateTime</th><th>duration</th><th>result</th></tr><tr><td>2000-01-12T12:13:14Z</td><td>P1Y3M5DT7H10M3.3S</td><td>2001-04-17T19:23:17.3Z</td></tr><tr><td>2000-01</td><td>-P3M</td><td>1999-10</td></tr><tr><td>2000-01-12</td><td>PT33H</td><td>2000-01-13</td></tr></tbody></table>

 Note that the addition defined by [*dateTimePlusDuration*](#vp-dt-dateTimePlusDuration) differs from addition on integers or real numbers in not being commutative. The order of addition of durations to instants *is* significant. For example, there are cases where:

> > ((dateTime + duration1) + duration2) != ((dateTime + duration2) + duration1)

*Example:*

-  (2000-03-30 + P1D) + P1M = 2000-03-31 + P1M = 2000-**04-30**
-  (2000-03-30 + P1M) + P1D = 2000-04-30 + P1D = 2000-**05-01**

#### <a id="sec-timeontimeline"></a>E.3.4 Time on timeline

  Time on Timeline for Date/time Seven-property Model Datatypes **<a id="vp-dt-timeOnTimeline">*timeOnTimeline*</a>** (*dt*) → decimal number Maps a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value to the decimal number representing its position on the "time line".  **Arguments:**

<table><tbody>
    <tr><td>dt</td><td>: </td><td>a <a href="#dt-dt-7PropMod">date/timeSevenPropertyModel</a>
        value</td></tr>

    </tbody></table>

 **Result:**a decimal number **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            yr be
                1971
                when dt's <a href="#vp-dt-year">·year·</a> is absent,
                and  dt's
                <a href="#vp-dt-year">·year·</a> − 1  otherwise,mo be 12 or
                dt's <a href="#vp-dt-month">·month·</a> , similarly,da be 
                <a href="#f-daysInMonth">·daysInMonth·</a>(yr+1, mo) − 1 
                or  (dt's
                <a href="#vp-dt-day">·day·</a>) − 1 , similarly,hr be 0 or
                dt's <a href="#vp-dt-hour">·hour·</a> ,
                similarly,mi be 0 or
                dt's <a href="#vp-dt-minute">·minute·</a> ,
                similarly,
                andse be 0 or
                dt's <a href="#vp-dt-second">·second·</a> ,
                similarly.
        </td></tr></tbody></table>

1. Subtract [*timezoneOffset*](#vp-dt-timezone) from *mi*   when [*timezoneOffset*](#vp-dt-timezone) is not ***absent***.
2. ([*year*](#vp-dt-year))

  1. Set *ToTl* to  31536000 × *yr* .

3. (Leap-year Days, [*month*](#vp-dt-month), and [*day*](#vp-dt-day))

  1. Add  86400 × (*yr* [div](#dt-div) 400 − *yr* [div](#dt-div) 100 + *yr* [div](#dt-div) 4)  to *ToTl*.
  2. Add   86400 × Sum~*m* < *mo*~ [*daysInMonth*](#f-daysInMonth)(*yr* + 1, *m*) to *ToTl*
  3. Add   86400 × *da*  to *ToTl*.

4. ([*hour*](#vp-dt-hour), [*minute*](#vp-dt-minute), and [*second*](#vp-dt-second))

  1. Add  3600 × *hr* + 60 × *mi* + *se*  to *ToTl*.

5. Return *ToTl*.

#### <a id="sec-dt-lexmaps"></a>E.3.5 Lexical mappings

  Partial Date/time Lexical Mappings **<a id="f-dt-yrMap">*yearFragValue*</a>** (*YR*) → integer Maps a [*yearFrag*](#nt-yrFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*year*](#vp-dt-year) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value.  **Arguments:**

<table><tbody>
    <tr><td>YR</td><td>: </td><td>matches <a href="#nt-yrFrag">yearFrag</a></td></tr>

    </tbody></table>

 **Result:**an integer **Algorithm:** Return [*noDecimalMap*](#f-noDecVal)(*YR*) **<a id="f-dt-moMap">*monthFragValue*</a>** (*MO*) → integer Maps a [*monthFrag*](#nt-moFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*month*](#vp-dt-month) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value.  **Arguments:**

<table><tbody>
    <tr><td>MO</td><td>: </td><td>matches <a href="#nt-moFrag">monthFrag</a></td></tr>

    </tbody></table>

 **Result:**an integer **Algorithm:** Return [*unsignedNoDecimalMap*](#f-unsNoDecVal)(*MO*) **<a id="f-dt-daMap">*dayFragValue*</a>** (*DA*) → integer Maps a [*dayFrag*](#nt-daFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*day*](#vp-dt-day) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value.  **Arguments:**

<table><tbody>
    <tr><td>DA</td><td>: </td><td>matches <a href="#nt-daFrag">dayFrag</a></td></tr>

    </tbody></table>

 **Result:**an integer **Algorithm:** Return [*unsignedNoDecimalMap*](#f-unsNoDecVal)(*DA*) **<a id="f-dt-hrMap">*hourFragValue*</a>** (*HR*) → integer Maps a [*hourFrag*](#nt-hrFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*hour*](#vp-dt-hour) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value.  **Arguments:**

<table><tbody>
    <tr><td>HR</td><td>: </td><td>matches <a href="#nt-hrFrag">hourFrag</a></td></tr>

    </tbody></table>

 **Result:**an integer **Algorithm:** Return [*unsignedNoDecimalMap*](#f-unsNoDecVal)(*HR*) **<a id="f-dt-miMap">*minuteFragValue*</a>** (*MI*) → integer Maps a [*minuteFrag*](#nt-miFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*minute*](#vp-dt-minute) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value.  **Arguments:**

<table><tbody>
    <tr><td>MI</td><td>: </td><td>matches <a href="#nt-miFrag">minuteFrag</a></td></tr>

    </tbody></table>

 **Result:**an integer **Algorithm:** Return [*unsignedNoDecimalMap*](#f-unsNoDecVal)(*MI*) **<a id="f-dt-seMap">*secondFragValue*</a>** (*SE*) → decimal number Maps a [*secondFrag*](#nt-seFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto a decimal number, presumably the [*second*](#vp-dt-second) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value.  **Arguments:**

<table><tbody>
    <tr><td>SE</td><td>: </td><td>matches <a href="#nt-seFrag">secondFrag</a></td></tr>

    </tbody></table>

 **Result:**a decimal number **Algorithm:** Return

- [*unsignedNoDecimalMap*](#f-unsNoDecVal)(*SE*)   when no decimal point occurs in *SE*, and
- [*unsignedDecimalPtMap*](#f-unsDecVal)(*SE*)   otherwise.
 **<a id="f-dt-tzMap">*timezoneFragValue*</a>** (*TZ*) → integer Maps a [*timezoneFrag*](#nt-tzFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation), onto an integer, presumably the [*timezoneOffset*](#vp-dt-timezone) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value.  **Arguments:**

<table><tbody>
    <tr><td>TZ</td><td>: </td><td>matches <a href="#nt-tzFrag">timezoneFrag</a></td></tr>

    </tbody></table>

 **Result:**an integer **Algorithm:** *TZ* necessarily consists of either just '`Z`', or a sign ('`+`' or '`-`') followed by an instance *H* of [*hourFrag*](#nt-hrFrag), a colon, and an instance *M* of [*minuteFrag*](#nt-miFrag)  Return

- 0   when *TZ* is '`Z`',
- −([*unsignedDecimalPtMap*](#f-unsDecVal)(*H*) × 60 + [*unsignedDecimalPtMap*](#f-unsDecVal)(*M*))   when the sign is '`-`', and
- [*unsignedDecimalPtMap*](#f-unsDecVal)(*H*) × 60 + [*unsignedDecimalPtMap*](#f-unsDecVal)(*M*)   otherwise.
 **Note:** There is no [lexical mapping](#dt-lexical-mapping) for [*endOfDayFrag*](#nt-eodFrag); it is handled specially by the relevant [lexical mappings](#dt-lexical-mapping).  See, e.g., [*dateTimeLexicalMap*](#vp-dateTimeLexRep).  Lexical Mapping **<a id="vp-dateTimeLexRep">*dateTimeLexicalMap*</a>** (*LEX*) → [dateTime](#dateTime) Maps a [*dateTimeLexicalRep*](#nt-dateTimeRep) to a [dateTime](#dateTime) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>matches <a href="#nt-dateTimeRep">dateTimeLexicalRep</a></td></tr>

    </tbody></table>

 **Result:**a complete [dateTime](#dateTime) value **Algorithm:** *LEX* necessarily includes substrings that are instances of [*yearFrag*](#nt-yrFrag), [*monthFrag*](#nt-moFrag), and [*dayFrag*](#nt-daFrag) (below referred to as *Y*, *MO*, and *D* respectively); it also contains either instances of [*hourFrag*](#nt-hrFrag), [*minuteFrag*](#nt-miFrag), and [*secondFrag*](#nt-seFrag)(*Y*, *MI*, and *S*), or else an instance of [*endOfDayFrag*](#nt-eodFrag); finally, it may optionally contain an instance of[*timezoneFrag*](#nt-tzFrag) (*T*).

<table><tbody><tr><td>Let </td><td>tz
            be <a href="#f-dt-tzMap">·timezoneFragValue·</a>(T) when T
            is present, otherwise absent.</td></tr></tbody></table>

  Return

- [*newDateTime*](#p-setDTFromRaw)([*yearFragValue*](#f-dt-yrMap)(*Y*), [*monthFragValue*](#f-dt-moMap)(*MO*), [*dayFragValue*](#f-dt-daMap)(*D*), 24, 0, 0, *tz*) when [*endOfDayFrag*](#nt-eodFrag) is present, and
-  [*newDateTime*](#p-setDTFromRaw)([*yearFragValue*](#f-dt-yrMap)(*Y*), [*monthFragValue*](#f-dt-moMap)(*MO*), [*dayFragValue*](#f-dt-daMap)(*D*), [*hourFragValue*](#f-dt-hrMap)(*H*), [*minuteFragValue*](#f-dt-miMap)(*MI*), [*secondFragValue*](#f-dt-seMap)(*S*), *tz*) otherwise
    Lexical Mapping **<a id="vp-timeLexRep">*timeLexicalMap*</a>** (*LEX*) → [time](#time) Maps a [*timeLexicalRep*](#nt-timeRep) to a [time](#time) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>matches
        <a href="#nt-timeRep">timeLexicalRep</a></td></tr>

    </tbody></table>

 **Result:**a complete [time](#time) value **Algorithm:** *LEX* necessarily includes either substrings that are instances of [*hourFrag*](#nt-hrFrag), [*minuteFrag*](#nt-miFrag), and [*secondFrag*](#nt-seFrag), (below referred to as *H*, *M*, and *S* respectively), or else an instance of [*endOfDayFrag*](#nt-eodFrag); finally, it may optionally contain an instance of [*timezoneFrag*](#nt-tzFrag) (*T*).

<table><tbody><tr><td>Let </td><td>tz
            be <a href="#f-dt-tzMap">·timezoneFragValue·</a>(T) when T
            is present, otherwise absent</td></tr></tbody></table>

  Return

- [*newDateTime*](#p-setDTFromRaw)(***absent***, ***absent***, ***absent***, 0, 0, 0, *tz*) when [*endOfDayFrag*](#nt-eodFrag) is present, and
-  [*newDateTime*](#p-setDTFromRaw)(***absent***, ***absent***, ***absent***, [*hourFragValue*](#f-dt-hrMap)(*H*), [*minuteFragValue*](#f-dt-miMap)(*MI*), [*secondFragValue*](#f-dt-seMap)(*S*), *tz*) otherwise.
    Lexical Mapping **<a id="vp-dateLexRep">*dateLexicalMap*</a>** (*LEX*) → [date](#date) Maps a [*dateLexicalRep*](#nt-dateRep) to a [date](#date) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>matches
        <a href="#nt-dateRep">dateLexicalRep</a></td></tr>

    </tbody></table>

 **Result:**a complete [date](#date) value **Algorithm:** *LEX* necessarily includes an instance *Y* of [*yearFrag*](#nt-yrFrag), an instance *M* of [*monthFrag*](#nt-moFrag), and an instance *D* of [*dayFrag*](#nt-daFrag), hyphen-separated and optionally followed by an instance *T* of [*timezoneFrag*](#nt-tzFrag).

<table><tbody><tr><td>Let </td><td>tz
            be <a href="#f-dt-tzMap">·timezoneFragValue·</a>(T) when T
            is present, otherwise absent</td></tr></tbody></table>

  Return [*newDateTime*](#p-setDTFromRaw)([*yearFragValue*](#f-dt-yrMap)(*Y*), [*monthFragValue*](#f-dt-moMap)(*M*), [*dayFragValue*](#f-dt-daMap)(*D*), ***absent***, ***absent***, ***absent***, *tz*.)    Lexical Mapping **<a id="vp-gYearMonthLexRep">*gYearMonthLexicalMap*</a>** (*LEX*) → [gYearMonth](#gYearMonth) Maps a [*gYearMonthLexicalRep*](#nt-gYearMonthRep) to a [gYearMonth](#gYearMonth) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>matches
        <a href="#nt-gYearMonthRep">gYearMonthLexicalRep</a></td></tr>

    </tbody></table>

 **Result:**a complete [gYearMonth](#gYearMonth) value **Algorithm:** *LEX* necessarily includes an instance *Y* of [*yearFrag*](#nt-yrFrag) and an instance *M* of [*monthFrag*](#nt-moFrag), hyphen-separated and optionally followed by an instance *T* of [*timezoneFrag*](#nt-tzFrag).

<table><tbody><tr><td>Let </td><td>tz
            be <a href="#f-dt-tzMap">·timezoneFragValue·</a>(T) when T
            is present, otherwise absent.</td></tr></tbody></table>

  Return [*newDateTime*](#p-setDTFromRaw)([*yearFragValue*](#f-dt-yrMap)(*Y*), [*monthFragValue*](#f-dt-moMap)(*M*), ***absent***, ***absent***, ***absent***, ***absent***, *tz*).    Lexical Mapping **<a id="vp-gYearLexRep">*gYearLexicalMap*</a>** (*LEX*) → [gYear](#gYear) Maps a [*gYearLexicalRep*](#nt-gYearRep) to a [gYear](#gYear) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>matches <a href="#nt-gYearRep">gYearLexicalRep</a></td></tr>

    </tbody></table>

 **Result:**a complete [gYear](#gYear) value **Algorithm:** *LEX* necessarily includes an instance *Y* of [*yearFrag*](#nt-yrFrag), optionally followed by an instance *T* of [*timezoneFrag*](#nt-tzFrag).

<table><tbody><tr><td>Let </td><td>tz
            be <a href="#f-dt-tzMap">·timezoneFragValue·</a>(T) when T
            is present, otherwise absent.</td></tr></tbody></table>

  Return [*newDateTime*](#p-setDTFromRaw)([*yearFragValue*](#f-dt-yrMap)(*Y*), ***absent***, ***absent***, ***absent***, ***absent***, ***absent***, *tz*).    Lexical Mapping **<a id="vp-gMonthDayLexRep">*gMonthDayLexicalMap*</a>** (*LEX*) → [gMonthDay](#gMonthDay) Maps a [*gMonthDayLexicalRep*](#nt-gMonthDayRep) to a [gMonthDay](#gMonthDay) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>matches
        <a href="#nt-gMonthDayRep">gMonthDayLexicalRep</a></td></tr>

    </tbody></table>

 **Result:**a complete [gMonthDay](#gMonthDay) value **Algorithm:** *LEX* necessarily includes an instance *M* of [*monthFrag*](#nt-moFrag) and an instance *D* of [*dayFrag*](#nt-daFrag), hyphen-separated and optionally followed by an instance *T* of [*timezoneFrag*](#nt-tzFrag).

<table><tbody><tr><td>Let </td><td>tz
            be <a href="#f-dt-tzMap">·timezoneFragValue·</a>(T) when T
            is present, otherwise absent.</td></tr></tbody></table>

  Return [*newDateTime*](#p-setDTFromRaw)(***absent***, [*monthFragValue*](#f-dt-moMap)(*M*), [*dayFragValue*](#f-dt-daMap)(*D*), ***absent***, ***absent***, ***absent***, *tz*.    Lexical Mapping **<a id="vp-gDayLexRep">*gDayLexicalMap*</a>** (*LEX*) → [gDay](#gDay) Maps a [*gDayLexicalRep*](#nt-gDayRep) to a [gDay](#gDay) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>matches <a href="#nt-gDayRep">gDayLexicalRep</a></td></tr>

    </tbody></table>

 **Result:**a complete [gDay](#gDay) value **Algorithm:** *LEX* necessarily includes an instance *D* of [*dayFrag*](#nt-daFrag), optionally followed by an instance *T* of [*timezoneFrag*](#nt-tzFrag).

<table><tbody><tr><td>Let </td><td>tz
            be <a href="#f-dt-tzMap">·timezoneFragValue·</a>(T) when T
            is present, otherwise absent.</td></tr></tbody></table>

1. Return [*newDateTime*](#p-setDTFromRaw)(*gD*, ***absent***, ***absent***, [*dayFragValue*](#f-dt-daMap)(*D*), ***absent***, ***absent***, ***absent***, *tz*).
 Return [*newDateTime*](#p-setDTFromRaw)(***absent***, ***absent***, [*dayFragValue*](#f-dt-daMap)(*D*), ***absent***, ***absent***, ***absent***, *tz*).    Lexical Mapping **<a id="vp-gMonthLexRep">*gMonthLexicalMap*</a>** (*LEX*) → [gMonth](#gMonth) Maps a [*gMonthLexicalRep*](#nt-gMonthRep) to a [gMonth](#gMonth) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>matches <a href="#nt-gMonthRep">gMonthLexicalRep</a></td></tr>

    </tbody></table>

 **Result:**a complete [gMonth](#gMonth) value **Algorithm:** *LEX* necessarily includes an instance *M* of [*monthFrag*](#nt-moFrag), optionally followed by an instance *T* of [*timezoneFrag*](#nt-tzFrag).

<table><tbody><tr><td>Let </td><td>tz
            be <a href="#f-dt-tzMap">·timezoneFragValue·</a>(T) when T
            is present, otherwise absent.</td></tr></tbody></table>

  Return [*newDateTime*](#p-setDTFromRaw)(***absent***, [*monthFragValue*](#f-dt-moMap)(*M*), ***absent***, ***absent***, ***absent***, ***absent***, *tz*)

#### <a id="sec-dt-canmaps"></a>E.3.6 Canonical Mappings

  Auxiliary Functions for Date/time Canonical Mappings **<a id="f-unsTwoDigCanFragMap">*unsTwoDigitCanonicalFragmentMap*</a>** (*i*) → [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) Maps a nonnegative integer less than 100 onto an unsigned always-two-digit numeral.  **Arguments:**

<table><tbody>
    <tr><td>i</td><td>: </td><td>a nonnegative integer less than 100</td></tr>

    </tbody></table>

 **Result:**matches [*unsignedNoDecimalPtNumeral*](#nt-unsNoDecNuml) **Algorithm:** Return [*digit*](#f-digit)(*i* [div](#dt-div) 10) & [*digit*](#f-digit)(*i* [mod](#dt-mod) 10) **<a id="f-fourDigCanFragMap">*fourDigitCanonicalFragmentMap*</a>** (*i*) → [*noDecimalPtNumeral*](#nt-noDecNuml) Maps an integer between -10000 and 10000 onto an always-four-digit numeral.  **Arguments:**

<table><tbody>
    <tr><td>i</td><td>: </td><td>an integer whose absolute value is less than 10000</td></tr>

    </tbody></table>

 **Result:**matches [*noDecimalPtNumeral*](#nt-noDecNuml) **Algorithm:** Return

- '`-`' & [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(−*i* [div](#dt-div) 100) & [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(−*i* [mod](#dt-mod) 100)   when *i* is negative,
- [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(*i* [div](#dt-div) 100) & [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(*i* [mod](#dt-mod) 100)   otherwise.
   Partial Date/time Canonical Mappings **<a id="f-yrCanFragMap">*yearCanonicalFragmentMap*</a>** (*y*) → [*yearFrag*](#nt-yrFrag) Maps an integer, presumably the [*year*](#vp-dt-year) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*yearFrag*](#nt-yrFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
    <tr><td>y</td><td>: </td><td>an integer </td></tr>

    </tbody></table>

 **Result:**matches [*yearFrag*](#nt-yrFrag) **Algorithm:** Return

- [*noDecimalPtCanonicalMap*](#f-noDecCanMap)(*y*)   when  |*y*| > 9999 .
- [*fourDigitCanonicalFragmentMap*](#f-fourDigCanFragMap)(*y*)   otherwise.
 **<a id="f-moCanFragMap">*monthCanonicalFragmentMap*</a>** (*m*) → [*monthFrag*](#nt-moFrag) Maps an integer, presumably the [*month*](#vp-dt-month) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*monthFrag*](#nt-moFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
    <tr><td>m</td><td>: </td><td>an integer between 1 and 12 inclusive</td></tr>

    </tbody></table>

 **Result:**matches [*monthFrag*](#nt-moFrag) **Algorithm:** Return [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(*m*) **<a id="f-daCanFragMap">*dayCanonicalFragmentMap*</a>** (*d*) → [*dayFrag*](#nt-daFrag) Maps an integer, presumably the [*day*](#vp-dt-day) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*dayFrag*](#nt-daFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
    <tr><td>d</td><td>: </td><td>an integer between 1 and 31 inclusive 
        (may be limited further depending on associated <a href="#vp-dt-year">·year·</a> and <a href="#vp-dt-month">·month·</a>)</td></tr>

    </tbody></table>

 **Result:**matches [*dayFrag*](#nt-daFrag) **Algorithm:** Return [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(*d*) **<a id="f-hrCanFragMap">*hourCanonicalFragmentMap*</a>** (*h*) → [*hourFrag*](#nt-hrFrag) Maps an integer, presumably the [*hour*](#vp-dt-hour) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*hourFrag*](#nt-hrFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
    <tr><td>h</td><td>: </td><td>an integer between 0 and 23 inclusive.</td></tr>

    </tbody></table>

 **Result:**matches [*hourFrag*](#nt-hrFrag) **Algorithm:** Return [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(*h*) **<a id="f-miCanFragMap">*minuteCanonicalFragmentMap*</a>** (*m*) → [*minuteFrag*](#nt-miFrag) Maps an integer, presumably the [*minute*](#vp-dt-minute) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*minuteFrag*](#nt-miFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
    <tr><td>m</td><td>: </td><td>an integer between 0 and 59 inclusive.</td></tr>

    </tbody></table>

 **Result:**matches [*minuteFrag*](#nt-miFrag) **Algorithm:** Return [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(*m*) **<a id="f-seCanFragMap">*secondCanonicalFragmentMap*</a>** (*s*) → [*secondFrag*](#nt-seFrag) Maps a decimal number, presumably the [*second*](#vp-dt-second) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*secondFrag*](#nt-seFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
    <tr><td>s</td><td>: </td><td>a nonnegative decimal number less than 70</td></tr>

    </tbody></table>

 **Result:**matches [*secondFrag*](#nt-seFrag) **Algorithm:** Return

- [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(*s*)   when *s* is an integer, and
- [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(*s*[div](#dt-div)1) & '`.`' & [*fractionDigitsCanonicalFragmentMap*](#f-fracDigitsMap)(*s*[mod](#dt-mod)1)   otherwise.
 **<a id="f-tzCanFragMap">*timezoneCanonicalFragmentMap*</a>** (*t*) → [*timezoneFrag*](#nt-tzFrag) Maps an integer, presumably the [*timezoneOffset*](#vp-dt-timezone) property of a [date/timeSevenPropertyModel](#dt-dt-7PropMod) value, onto a [*timezoneFrag*](#nt-tzFrag), part of a [date/timeSevenPropertyModel](#dt-dt-7PropMod)'s [lexical representation](#dt-lexical-representation).  **Arguments:**

<table><tbody>
    <tr><td>t</td><td>: </td><td>an integer between −840 and 840 inclusive</td></tr>

    </tbody></table>

 **Result:**matches [*timezoneFrag*](#nt-tzFrag) **Algorithm:** Return

- '`Z`'   when *t* is zero,
- '`-`' & [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(−*t* [div](#dt-div) 60) & '`:`' & [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(−*t* [mod](#dt-mod) 60)   when *t* is negative, and
- '`+`' & [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(*t* [div](#dt-div) 60) & '`:`' & [*unsTwoDigitCanonicalFragmentMap*](#f-unsTwoDigCanFragMap)(*t* [mod](#dt-mod) 60)   otherwise.
   Canonical Mapping **<a id="vp-dateTimeCanRep">*dateTimeCanonicalMap*</a>** (*dt*) → [*dateLexicalRep*](#nt-dateRep) Maps a [dateTime](#dateTime) value to a [*dateTimeLexicalRep*](#nt-dateTimeRep).  **Arguments:**

<table><tbody>
    <tr><td>dt</td><td>: </td><td>a complete <a href="#dateTime">dateTime</a> value</td></tr>

    </tbody></table>

 **Result:**matches [*dateLexicalRep*](#nt-dateRep) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            DT be 
            <a href="#f-yrCanFragMap">·yearCanonicalFragmentMap·</a>(dt's <a href="#vp-dt-year">·year·</a>) &
            '<code>-</code>' &
            <a href="#f-moCanFragMap">·monthCanonicalFragmentMap·</a>(dt's <a href="#vp-dt-month">·month·</a>) &
            '<code>-</code>' &
            <a href="#f-daCanFragMap">·dayCanonicalFragmentMap·</a>(dt's <a href="#vp-dt-day">·day·</a>) &
            '<code>T</code>' &
            <a href="#f-hrCanFragMap">·hourCanonicalFragmentMap·</a>(dt's <a href="#vp-dt-hour">·hour·</a>) &
            '<code>:</code>' &
            <a href="#f-miCanFragMap">·minuteCanonicalFragmentMap·</a>(dt's <a href="#vp-dt-minute">·minute·</a>) &
            '<code>:</code>' &
            <a href="#f-seCanFragMap">·secondCanonicalFragmentMap·</a>(dt's <a href="#vp-dt-second">·second·</a>) .
        </td></tr></tbody></table>

 Return

- *DT*   when *dt*'s [*timezoneOffset*](#vp-dt-timezone) is ***absent***, and
- *DT* & [*timezoneCanonicalFragmentMap*](#f-tzCanFragMap)(*dt*'s [*timezoneOffset*](#vp-dt-timezone))   otherwise.
    Canonical Mapping **<a id="vp-timeCanRep">*timeCanonicalMap*</a>** (*ti*) → [*timeLexicalRep*](#nt-timeRep) Maps a [time](#time) value to a [*timeLexicalRep*](#nt-timeRep).  **Arguments:**

<table><tbody>
    <tr><td>ti</td><td>: </td><td>a complete <a href="#time">time</a> value</td></tr>

    </tbody></table>

 **Result:**matches [*timeLexicalRep*](#nt-timeRep) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            T be 
            <a href="#f-hrCanFragMap">·hourCanonicalFragmentMap·</a>(ti's <a href="#vp-dt-hour">·hour·</a>) &
            '<code>:</code>' &
            <a href="#f-miCanFragMap">·minuteCanonicalFragmentMap·</a>(ti's <a href="#vp-dt-minute">·minute·</a>) &
            '<code>:</code>' &
            <a href="#f-seCanFragMap">·secondCanonicalFragmentMap·</a>(ti's <a href="#vp-dt-second">·second·</a>) .
        </td></tr></tbody></table>

 Return

- *T*   when *ti*'s [*timezoneOffset*](#vp-dt-timezone) is ***absent***, and
- *T* & [*timezoneCanonicalFragmentMap*](#f-tzCanFragMap)(*ti*'s [*timezoneOffset*](#vp-dt-timezone))   otherwise.
    Canonical Mapping **<a id="vp-dateCanRep">*dateCanonicalMap*</a>** (*da*) → [*dateLexicalRep*](#nt-dateRep) Maps a [date](#date) value to a [*dateLexicalRep*](#nt-dateRep).  **Arguments:**

<table><tbody>
    <tr><td>da</td><td>: </td><td>a complete <a href="#date">date</a> value</td></tr>

    </tbody></table>

 **Result:**matches [*dateLexicalRep*](#nt-dateRep) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            D be 
            <a href="#f-yrCanFragMap">·yearCanonicalFragmentMap·</a>(da's <a href="#vp-dt-year">·year·</a>) &
            '<code>-</code>' &
            <a href="#f-moCanFragMap">·monthCanonicalFragmentMap·</a>(da's <a href="#vp-dt-month">·month·</a>) &
            '<code>-</code>' &
            <a href="#f-daCanFragMap">·dayCanonicalFragmentMap·</a>(da's <a href="#vp-dt-day">·day·</a>) .
        </td></tr></tbody></table>

 Return

- *D*   when *da*'s [*timezoneOffset*](#vp-dt-timezone) is ***absent***, and
- *D* & [*timezoneCanonicalFragmentMap*](#f-tzCanFragMap)(*da*'s [*timezoneOffset*](#vp-dt-timezone))   otherwise.
    Canonical Mapping **<a id="vp-gYearMonthCanRep">*gYearMonthCanonicalMap*</a>** (*ym*) → [*gYearMonthLexicalRep*](#nt-gYearMonthRep) Maps a [gYearMonth](#gYearMonth) value to a [*gYearMonthLexicalRep*](#nt-gYearMonthRep).  **Arguments:**

<table><tbody>
    <tr><td>ym</td><td>: </td><td>a complete <a href="#gYearMonth">gYearMonth</a> value</td></tr>

    </tbody></table>

 **Result:**matches [*gYearMonthLexicalRep*](#nt-gYearMonthRep) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            YM be 
            <a href="#f-yrCanFragMap">·yearCanonicalFragmentMap·</a>(ym's <a href="#vp-dt-year">·year·</a>) &
            '<code>-</code>' &
            <a href="#f-moCanFragMap">·monthCanonicalFragmentMap·</a>(ym's <a href="#vp-dt-month">·month·</a>) .
        </td></tr></tbody></table>

 Return

- *YM*   when *ym*'s [*timezoneOffset*](#vp-dt-timezone) is ***absent***, and
- *YM* & [*timezoneCanonicalFragmentMap*](#f-tzCanFragMap)(*ym*'s [*timezoneOffset*](#vp-dt-timezone))   otherwise.
    Canonical Mapping **<a id="vp-gYearCanRep">*gYearCanonicalMap*</a>** (*gY*) → [*gYearLexicalRep*](#nt-gYearRep) Maps a [gYear](#gYear) value to a [*gYearLexicalRep*](#nt-gYearRep).  **Arguments:**

<table><tbody>
    <tr><td>gY</td><td>: </td><td>a complete <a href="#gYear">gYear</a> value</td></tr>

    </tbody></table>

 **Result:**matches [*gYearLexicalRep*](#nt-gYearRep) **Algorithm:** Return

- [*yearCanonicalFragmentMap*](#f-yrCanFragMap)(*gY*'s [*year*](#vp-dt-year))   when *gY*'s [*timezoneOffset*](#vp-dt-timezone) is ***absent***, and
- [*yearCanonicalFragmentMap*](#f-yrCanFragMap)(*gY*'s [*year*](#vp-dt-year)) & [*timezoneCanonicalFragmentMap*](#f-tzCanFragMap)(*gY*'s [*timezoneOffset*](#vp-dt-timezone))   otherwise.
   Canonical Mapping **<a id="vp-gMonthDayCanRep">*gMonthDayCanonicalMap*</a>** (*md*) → [*gMonthDayLexicalRep*](#nt-gMonthDayRep) Maps a [gMonthDay](#gMonthDay) value to a [*gMonthDayLexicalRep*](#nt-gMonthDayRep).  **Arguments:**

<table><tbody>
    <tr><td>md</td><td>: </td><td>a complete <a href="#gMonthDay">gMonthDay</a> value</td></tr>

    </tbody></table>

 **Result:**matches [*gMonthDayLexicalRep*](#nt-gMonthDayRep) **Algorithm:**

<table><tbody><tr><td>Let </td><td>
            MD be  '<code>--</code>' &
            <a href="#f-moCanFragMap">·monthCanonicalFragmentMap·</a>(md's <a href="#vp-dt-month">·month·</a>) &
            '<code>-</code>' &
            <a href="#f-daCanFragMap">·dayCanonicalFragmentMap·</a>(md's <a href="#vp-dt-day">·day·</a>) .
        </td></tr></tbody></table>

 Return

- *MD*   when *md*'s [*timezoneOffset*](#vp-dt-timezone) is ***absent***, and
- *MD* & [*timezoneCanonicalFragmentMap*](#f-tzCanFragMap)(*md*'s [*timezoneOffset*](#vp-dt-timezone))   otherwise.
    Canonical Mapping **<a id="vp-gDayCanRep">*gDayCanonicalMap*</a>** (*gD*) → [*gDayLexicalRep*](#nt-gDayRep) Maps a [gDay](#gDay) value to a [*gDayLexicalRep*](#nt-gDayRep).  **Arguments:**

<table><tbody>
    <tr><td>gD</td><td>: </td><td>a complete <a href="#gDay">gDay</a> value</td></tr>

    </tbody></table>

 **Result:**matches [*gDayLexicalRep*](#nt-gDayRep) **Algorithm:** Return

- '`---`' & [*dayCanonicalFragmentMap*](#f-daCanFragMap)(*gD*'s [*day*](#vp-dt-day))   when *gD*'s [*timezoneOffset*](#vp-dt-timezone) is ***absent***, and
- '`---`' & [*dayCanonicalFragmentMap*](#f-daCanFragMap)(*gD*'s [*day*](#vp-dt-day)) & [*timezoneCanonicalFragmentMap*](#f-tzCanFragMap)(*gD*'s [*timezoneOffset*](#vp-dt-timezone))   otherwise.
   Canonical Mapping **<a id="vp-gMonthCanRep">*gMonthCanonicalMap*</a>** (*gM*) → [*gMonthLexicalRep*](#nt-gMonthRep) Maps a [gMonth](#gMonth) value to a [*gMonthLexicalRep*](#nt-gMonthRep).  **Arguments:**

<table><tbody>
    <tr><td>gM</td><td>: </td><td>a complete <a href="#gMonth">gMonth</a> value</td></tr>

    </tbody></table>

 **Result:**matches [*gMonthLexicalRep*](#nt-gMonthRep) **Algorithm:** Return

- '`--`' & [*monthCanonicalFragmentMap*](#f-moCanFragMap)(*gM*'s [*day*](#vp-dt-day))   when *gM*'s [*timezoneOffset*](#vp-dt-timezone) is ***absent***, and
- '`--`' & [*monthCanonicalFragmentMap*](#f-moCanFragMap)(*gM*'s [*day*](#vp-dt-day)) & [*timezoneCanonicalFragmentMap*](#f-tzCanFragMap)(*gM*'s [*timezoneOffset*](#vp-dt-timezone))   otherwise.

### <a id="sec-misc-lexmaps"></a>E.4 Lexical and Canonical Mappings for Other Datatypes

The following functions are used with various datatypes neither numeric nor date/time related.

  Lexical Mapping **<a id="f-stringLexmap">*stringLexicalMap*</a>** (*LEX*) → [string](#string) Maps a [literal](#dt-literal) matching the [*stringRep*](#nt-stringRep) production to a [string](#string) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>a <a href="#dt-literal">·literal·</a> matching
        <a href="#nt-stringRep">stringRep</a></td></tr>

    </tbody></table>

 **Result:**A [string](#string) value **Algorithm:** Return *LEX*.  (The function is the identity function on the domain.)   Lexical Mapping **<a id="f-booleanLexmap">*booleanLexicalMap*</a>** (*LEX*) → [boolean](#boolean) Maps a [literal](#dt-literal) matching the [*booleanRep*](#nt-booleanRep) production to a [boolean](#boolean) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>a <a href="#dt-literal">·literal·</a> matching
        <a href="#nt-booleanRep">booleanRep</a></td></tr>

    </tbody></table>

 **Result:**A [boolean](#boolean) value **Algorithm:** Return

- ***true***   when *LEX* is '`true`' or '`1`' , and
- ***false***   otherwise (*LEX* is '`false`' or '`0`').
    Canonical Mapping **<a id="f-stringCanmap">*stringCanonicalMap*</a>** (*s*) → [*stringRep*](#nt-stringRep) Maps a [string](#string) value to a [*stringRep*](#nt-stringRep).  **Arguments:**

<table><tbody>
    <tr><td>s</td><td>: </td><td>a <a href="#string">string</a> value</td></tr>

    </tbody></table>

 **Result:**matches [*stringRep*](#nt-stringRep) **Algorithm:** Return *s*.  (The function is the identity function on the domain.)   Canonical Mapping **<a id="f-booleanCanmap">*booleanCanonicalMap*</a>** (*b*) → [*booleanRep*](#nt-booleanRep) Maps a [boolean](#boolean) value to a [*booleanRep*](#nt-booleanRep).  **Arguments:**

<table><tbody>
    <tr><td>b</td><td>: </td><td>a <a href="#boolean">boolean</a> value</td></tr>

    </tbody></table>

 **Result:**matches [*booleanRep*](#nt-booleanRep) **Algorithm:** Return

- '`true`'   when *b* is ***true***, and
- '`false`'   otherwise (*b* is ***false***).

#### <a id="sec-hexbin-lexmaps"></a>E.4.1 Lexical and canonical mappings for [hexBinary](#hexBinary)

The [lexical mapping](#dt-lexical-mapping) for [hexBinary](#hexBinary) maps each pair of hexadecimal digits to an octet, in the conventional way:

  Lexical Mapping for hexBinary **<a id="f-hexBinaryMap">*hexBinaryMap*</a>** (*LEX*) → [hexBinary](#hexBinary) Maps a [literal](#dt-literal) matching the [*hexBinary*](#nt-hexBinary) production to a sequence of octets in the form of a [hexBinary](#hexBinary) value.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>a <a href="#dt-literal">·literal·</a> matching
        <a href="#nt-hexBinary">hexBinary</a></td></tr>

    </tbody></table>

 **Result:**A sequence of binary octets in the form of a [hexBinary](#hexBinary) value **Algorithm:** *LEX* necessarily includes a sequence of zero or more substrings matching the [*hexOctet*](#nt-hexOctet) production.

<table><tbody><tr><td>Let </td><td>o be the sequence of octets formed by applying
            <a href="#f-hexOctetMap">·hexOctetMap·</a> to each <a href="#nt-hexOctet">hexOctet</a> in
            LEX, in order, and concatenating the results.</td></tr></tbody></table>

 Return *o*.

The auxiliary functions [*hexOctetMap*](#f-hexOctetMap) and [*hexDigitMap*](#f-hexDigitMap) are used by [*hexBinaryMap*](#f-hexBinaryMap).

  Mappings for hexadecimal digits **<a id="f-hexOctetMap">*hexOctetMap*</a>** (*LEX*) → octet Maps a [literal](#dt-literal) matching the [*hexOctet*](#nt-hexOctet) production to a single octet.  **Arguments:**

<table><tbody>
    <tr><td>LEX</td><td>: </td><td>a <a href="#dt-literal">·literal·</a> matching
        <a href="#nt-hexOctet">hexOctet</a></td></tr>

    </tbody></table>

 **Result:**A single binary octet **Algorithm:** *LEX* necessarily includes exactly two hexadecimal digits.

<table><tbody><tr><td>Let </td><td>d0 be the first hexadecimal digit in LEX.
            Let d1 be the second hexadecimal digit in LEX.</td></tr></tbody></table>

 Return the octet whose four high-order bits are [*hexDigitMap*](#f-hexDigitMap)(*d0*) and whose four low-order bits are [*hexDigitMap*](#f-hexDigitMap)(*d1*).  **<a id="f-hexDigitMap">*hexDigitMap*</a>** (*d*) → a bit-sequence of length four Maps a hexadecimal digit (a character matching the [*hexDigit*](#nt-hexDigit) production) to a sequence of four binary digits.  **Arguments:**

<table><tbody>
    <tr><td>d</td><td>: </td><td>a hexadecimal digit</td></tr>

    </tbody></table>

 **Result:**a sequence of four binary digits **Algorithm:** Return

- 0000 when *d* = '`0`',
- 0001 when *d* = '`1`',
- 0010 when *d* = '`2`',
- 0011 when *d* = '`3`',
- ...
- 1110 when *d* = '`E`' or '`e`',
- 1111 when *d* = '`F`' or '`f`'.

The [canonical mapping](#dt-canonical-mapping) for [hexBinary](#hexBinary) uses only the uppercase forms of A-F.

  Canonical Mapping for hexBinary **<a id="f-hexBinaryCanonical">*hexBinaryCanonical*</a>** (*o*) → [*hexBinary*](#nt-hexBinary) Maps a [hexBinary](#hexBinary) value to a literal matching the [*hexBinary*](#nt-hexBinary) production.  **Arguments:**

<table><tbody>
    <tr><td>o</td><td>: </td><td>a <a href="#hexBinary">hexBinary</a> value</td></tr>

    </tbody></table>

 **Result:**matches [*hexBinary*](#nt-hexBinary) **Algorithm:**

<table><tbody><tr><td>Let </td><td>h be the sequence of literals formed by applying
            <a href="#f-hexOctetCanonical">·hexOctetCanonical·</a> to each octet in
            o, in order, and concatenating the results.</td></tr></tbody></table>

 Return *h*.    Auxiliary procedures for canonical mapping of [hexBinary](#hexBinary) **<a id="f-hexOctetCanonical">*hexOctetCanonical*</a>** (*o*) → [*hexOctet*](#nt-hexOctet) Maps a binary octet to a literal matching the [*hexOctet*](#nt-hexOctet) production.  **Arguments:**

<table><tbody>
    <tr><td>o</td><td>: </td><td>a binary octet</td></tr>

    </tbody></table>

 **Result:**matches [*hexOctet*](#nt-hexOctet) **Algorithm:**

<table><tbody><tr><td>Let </td><td>lo be the four low-order bits of o,
            and hi be the four high-order bits.</td></tr></tbody></table>

 Return [*hexDigitCanonical*](#f-hexDigitCanonical)(*hi*) & [*hexDigitCanonical*](#f-hexDigitCanonical)(*lo*).  **<a id="f-hexDigitCanonical">*hexDigitCanonical*</a>** (*d*) → [*hexDigit*](#nt-hexDigit) Maps a four-bit sequence to a hexadecimal digit (a literal matching the [*hexDigit*](#nt-hexDigit) production).  **Arguments:**

<table><tbody>
    <tr><td>d</td><td>: </td><td>a sequence of four binary digits</td></tr>

    </tbody></table>

 **Result:**matches [*hexDigit*](#nt-hexDigit) **Algorithm:** Return

- '`0`' when *d* = 0000,
- '`1`' when *d* = 0001,
- '`2`' when *d* = 0010,
- '`3`' when *d* = 0011,
- ...
- '`E`' when *d* = 1110,
- '`F`' when *d* = 1111.

## <a id="sec-datatypes-and-facets"></a>F Datatypes and Facets

### <a id="app-fundamental-facets"></a>F.1 Fundamental Facets

 The following table shows the values of the fundamental facets for each [built-in](#dt-built-in) datatype.

<table>
        <tbody>
        <tr><th> </th><th>Datatype</th><th>ordered</th><th>bounded</th><th>cardinality</th><th>numeric</th></tr>
        <tr><td rowspan="19"><a href="#dt-primitive">primitive</a></td><td><a href="#string">string</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#boolean">boolean</a></td><td>false</td><td>false</td><td>finite</td><td>false</td></tr>
        <tr><td><a href="#float">float</a></td><td>partial</td><td>true</td><td>finite</td><td>true</td></tr>
        <tr><td><a href="#double">double</a></td><td>partial</td><td>true</td><td>finite</td><td>true</td></tr>
        <tr><td><a href="#decimal">decimal</a></td><td>total</td><td>false</td><td>countably infinite</td><td>true</td></tr>
        <tr><td><a href="#duration">duration</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#dateTime">dateTime</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#time">time</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#date">date</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#gYearMonth">gYearMonth</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#gYear">gYear</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#gMonthDay">gMonthDay</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#gDay">gDay</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#gMonth">gMonth</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#hexBinary">hexBinary</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#base64Binary">base64Binary</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#anyURI">anyURI</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#QName">QName</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#NOTATION">NOTATION</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td colspan="7"></td></tr>
        <tr><td rowspan="28"><a href="#dt-constructed">non-primitive</a></td><td><a href="#normalizedString">normalizedString</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#token">token</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#language">language</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#IDREFS">IDREFS</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#ENTITIES">ENTITIES</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#NMTOKEN">NMTOKEN</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#NMTOKENS">NMTOKENS</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#Name">Name</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#NCName">NCName</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#ID">ID</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#IDREF">IDREF</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#ENTITY">ENTITY</a></td><td>false</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#integer">integer</a></td><td>total</td><td>false</td><td>countably infinite</td><td>true</td></tr>
        <tr><td><a href="#nonPositiveInteger">nonPositiveInteger</a></td><td>total</td><td>false</td><td>countably infinite</td><td>true</td></tr>
        <tr><td><a href="#negativeInteger">negativeInteger</a></td><td>total</td><td>false</td><td>countably infinite</td><td>true</td></tr>
        <tr><td><a href="#long">long</a></td><td>total</td><td>true</td><td>finite</td><td>true</td></tr>
        <tr><td><a href="#int">int</a></td><td>total</td><td>true</td><td>finite</td><td>true</td></tr>
        <tr><td><a href="#short">short</a></td><td>total</td><td>true</td><td>finite</td><td>true</td></tr>
        <tr><td><a href="#byte">byte</a></td><td>total</td><td>true</td><td>finite</td><td>true</td></tr>
        <tr><td><a href="#nonNegativeInteger">nonNegativeInteger</a></td><td>total</td><td>false</td><td>countably infinite</td><td>true</td></tr>
        <tr><td><a href="#unsignedLong">unsignedLong</a></td><td>total</td><td>true</td><td>finite</td><td>true</td></tr>
        <tr><td><a href="#unsignedInt">unsignedInt</a></td><td>total</td><td>true</td><td>finite</td><td>true</td></tr>
        <tr><td><a href="#unsignedShort">unsignedShort</a></td><td>total</td><td>true</td><td>finite</td><td>true</td></tr>
        <tr><td><a href="#unsignedByte">unsignedByte</a></td><td>total</td><td>true</td><td>finite</td><td>true</td></tr>
        <tr><td><a href="#positiveInteger">positiveInteger</a></td><td>total</td><td>false</td><td>countably infinite</td><td>true</td></tr>
        <tr><td><a href="#yearMonthDuration">yearMonthDuration</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#dayTimeDuration">dayTimeDuration</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>
        <tr><td><a href="#dateTimeStamp">dateTimeStamp</a></td><td>partial</td><td>false</td><td>countably infinite</td><td>false</td></tr>

        </tbody>
    </table>

## <a id="regexs"></a>G Regular Expressions

A [regular expression](#dt-regex) *R* is a sequence of characters that denote a set of strings *L*(*R*).  When used to constrain a [lexical space](#dt-lexical-space), a regular expression *R* asserts that only strings in *L*(*R*) are valid [literals](#dt-literal) for values of that type.

**Note:** Unlike some popular regular expression languages (including those defined by Perl and standard Unix utilities), the regular expression language defined here implicitly anchors all regular expressions at the head and tail, as the most common use of regular expressions in [pattern](#dt-pattern) is to match entire [literals](#dt-literal). For example, a datatype [derived](#dt-derived) from [string](#string) such that all values must begin with the character '`A`' (#x41) and end with the character '`Z`' (#x5a) would be defined as follows:

```
<simpleType name='myString'>
 <restriction base='string'>
  <pattern value='A.*Z'/>
 </restriction>
</simpleType>
```

In regular expression languages that are not implicitly anchored at the head and tail, it is customary to write the equivalent regular expression as:

> > `^A.*Z$`

 where '`^`' anchors the pattern at the head and '`$`' anchors at the tail.In those rare cases where an unanchored match is desired, including '`.*`' at the beginning and ending of the regular expression will achieve the desired results.  For example, a datatype [derived](#dt-derived) from string such that all values must contain at least 3 consecutive '`A`' (#x41) characters somewhere within the value could be defined as follows:

```
<simpleType name='myString'>
 <restriction base='string'>
  <pattern value='.*AAA.*'/>
 </restriction>
</simpleType>
```

### <a id="regex-branch"></a>G.1 Regular expressions and branches

<a id="dt-regex">[Definition:]</a>A **regular expression** is composed from zero or more [branches](#dt-branch), separated by '`|`' characters.

<table><tbody><tr><td>Regular Expression</td></tr><tr><td><table><tbody><tr><td><a id="regex"></a><a id="nt-regExp"></a>[64]   </td><td><code>regExp</code></td><td>   ::=   </td><td><code>
    <a href="#nt-branch">branch</a>
    ( '|' <a href="#nt-branch">branch</a> )*
</code></td></tr></tbody></table></td></tr></tbody></table>

<table><thead><tr><th>For all
    <a href="#dt-branch">·branches·</a>
    S, and for all
    <a href="#dt-regex">·regular
        expressions·</a> T, valid
    <a href="#dt-regex">·regular
        expressions·</a> R are:</th><th>Denoting the set of strings
    L(R)
    containing:</th></tr></thead><tbody><tr><td>(empty string)</td><td>just the empty string</td></tr><tr><td>S</td><td>all strings in
    L(S)</td></tr><tr><td>S <code>|</code>T</td><td>all strings in
    L(S)
    and all strings in
    L(T)</td></tr></tbody></table>

<a id="dt-branch">[Definition:]</a>A **branch** consists of zero or more [pieces](#dt-piece), concatenated together.

<table><tbody><tr><td>Branch</td></tr><tr><td><table><tbody><tr><td><a id="branch"></a><a id="nt-branch"></a>[65]   </td><td><code>branch</code></td><td>   ::=   </td><td><code><a href="#nt-piece">piece</a>*</code></td></tr></tbody></table></td></tr></tbody></table>

<table><thead><tr><th>
    For all
    <a href="#dt-piece">·pieces·</a>
    S, and for all
    <a href="#dt-branch">·branches·</a>
    T, valid
    <a href="#dt-branch">·branches·</a>
    R are:
</th><th>Denoting the set of strings
    L(R)
    containing:</th></tr></thead><tbody><tr><td>S</td><td>all strings in
    L(S)</td></tr><tr><td>ST</td><td>all strings
    st
    with s in
    L(S)
    and
    t
    in
    L(T)</td></tr></tbody></table>

### <a id="regex-piece"></a>G.2 Pieces, atoms, quantifiers

<a id="dt-piece">[Definition:]</a> A **piece** is an [atom](#dt-atom), possibly followed by a [quantifier](#dt-quantifier).

<table><tbody><tr><td>Piece</td></tr><tr><td><table><tbody><tr><td><a id="piece"></a><a id="nt-piece"></a>[66]   </td><td><code>piece</code></td><td>   ::=   </td><td><code><a href="#nt-atom">atom</a> <a href="#nt-quantifier">quantifier</a>?</code></td></tr></tbody></table></td></tr></tbody></table>

<table><thead><tr><th>For all
    <a href="#dt-atom">·atoms·</a>
    S and non-negative
    integers
    n,
    m
    such that
    n
    ≤ m, valid
    <a href="#dt-piece">·pieces·</a>
    R are:</th><th>Denoting the set of strings
    L(R)
    containing:</th></tr></thead><tbody><tr><td>S</td><td>all strings in
    L(S)</td></tr><tr><td>S <code>?</code></td><td>the empty string, and all strings in
    L(S)</td></tr><tr><td>S <code>*</code></td><td>all strings in
    L(S <code>?</code>)
    and all strings
    st
    with s in
    L(S <code>*</code>)
    and
    t
    in
    L(S)  
    (all concatenations of zero or more strings from
        L(S) )</td></tr><tr><td>S <code>+</code></td><td>all strings
    st
    with s in
    L(S)
    and
    t
    in
    L(S <code>*</code>)  
    (all concatenations of one or more strings from
        L(S) )</td></tr><tr><td>S <code>{</code>n<code>,</code>m<code>}</code></td><td>all strings
    st
    with s in
    L(S)
    and
    t
    in
    L(S <code>{</code>n−1<code>,</code>m−1<code>}</code>)
    (all
        concatenations
        of at least
        n,
        and at most
        m,
        strings from
        L(S) )</td></tr><tr><td>S <code>{</code>n<code>}</code></td><td>all
    strings in
    L(S<code>{</code>n<code>,</code>n<code>}</code>) 
    (all
        concatenations
        of exactly
        n
        strings from
        L(S) )</td></tr><tr><td>S <code>{</code>n<code>,}</code></td><td>all
    strings in
    L(S<code>{</code>n<code>}</code> S <code>*</code>) 
    (all
        concatenations
        of at least
        n
        strings from
        L(S) )</td></tr><tr><td>S<code> {0,</code>m<code>}</code></td><td>all strings
    st
    with s in
    L(S <code>?</code>)
    and
    t
    in
    L(S <code>{</code>0<code>,</code>m−1<code>}</code>). 
    (all
        concatenations
        of at most
        m
        strings from
        L(S) )</td></tr><tr><td>S <code>{0,0}</code></td><td>only the empty string</td></tr></tbody></table>

**Note:** The regular expression language in the Perl Programming Language [[Perl]](#Perl) does not include a quantifier of the form  *S* `{,`*m*`}` , since it is logically equivalent to *S* `{0,`*m*`}` .  We have, therefore, left this logical possibility out of the regular expression language defined by this specification.

<a id="dt-quantifier">[Definition:]</a>A **quantifier** is one of '`?`', '`*`', or '`+`', or a string of the form  `{`*n*`,`*m*`}`  or  `{`*n*`,}` , which have the meanings defined in the table above.

<table><tbody><tr><td>Quantifier</td></tr><tr><td><table><tbody><tr><td><a id="quant"></a><a id="nt-quantifier"></a>[67]   </td><td><code>quantifier</code></td><td>   ::=   </td><td><code>[?*+] | ( '{' <a href="#nt-quantity">quantity</a> '}' )</code></td></tr></tbody><tbody><tr><td><a id="quantity"></a><a id="nt-quantity"></a>[68]   </td><td><code>quantity</code></td><td>   ::=   </td><td><code><a href="#nt-quantRange">quantRange</a> |
    <a href="#nt-quantMin">quantMin</a> | <a href="#nt-QuantExact">QuantExact</a></code></td></tr></tbody><tbody><tr><td><a id="quantRange"></a><a id="nt-quantRange"></a>[69]   </td><td><code>quantRange</code></td><td>   ::=   </td><td><code><a href="#nt-QuantExact">QuantExact</a> ',' <a href="#nt-QuantExact">QuantExact</a></code></td></tr></tbody><tbody><tr><td><a id="quantMin"></a><a id="nt-quantMin"></a>[70]   </td><td><code>quantMin</code></td><td>   ::=   </td><td><code><a href="#nt-QuantExact">QuantExact</a> ','</code></td></tr></tbody><tbody><tr><td><a id="quantExact"></a><a id="nt-QuantExact"></a>[71]   </td><td><code>QuantExact</code></td><td>   ::=   </td><td><code>[0-9]+</code></td></tr></tbody></table></td></tr></tbody></table>

 <a id="dt-atom">[Definition:]</a> An **atom** is either a [normal character](#dt-normalc), a [character class](#dt-charclass), or a parenthesized [regular expression](#dt-regex).

<table><tbody><tr><td>Atom</td></tr><tr><td><table><tbody><tr><td><a id="atom"></a><a id="nt-atom"></a>[72]   </td><td><code>atom</code></td><td>   ::=   </td><td><code><a href="#nt-NormalChar">NormalChar</a> |
    <a href="#nt-charClass">charClass</a> | ( '('
    <a href="#nt-regExp">regExp</a> ')' )</code></td></tr></tbody></table></td></tr></tbody></table>

<table><thead><tr><th>
    For all
    <a href="#dt-normalc">·normal
        characters·</a>
    c,
    <a href="#dt-charclass">·character
        classes·</a> C, and
    <a href="#dt-regex">·regular
        expressions·</a> S, valid
    <a href="#dt-atom">·atoms·</a>
    R are:
</th><th>Denoting the set of strings
    L(R)
    containing:</th></tr></thead><tbody><tr><td>c</td><td>the single string consisting only of
    c</td></tr><tr><td>C</td><td>all strings in
    L(C)</td></tr><tr><td><code>(</code> S <code>)</code></td><td>L(S)</td></tr></tbody></table>

### <a id="regex-char-metachar"></a>G.3 Characters and metacharacters

<a id="dt-metac">[Definition:]</a>A **metacharacter** is either '`.`', '`\`', '`?`', '`*`', '`+`', '`{`', '`}`', '`(`', '`)`', '`|`', '`[`', or '`]`'.  These characters have special meanings in [regular expressions](#dt-regex), but can be escaped to form [atoms](#dt-atom) that denote the sets of strings containing only themselves, i.e., an escaped **metacharacter** behaves like a [normal character](#dt-normalc).

<a id="dt-normalc">[Definition:]</a>A **normal character** is any XML character that is not a [metacharacter](#dt-metac).  In [regular expressions](#dt-regex), a **normal character** is an [atom](#dt-atom) that denotes the singleton set of strings containing only itself.

<table><tbody><tr><td>Normal Character</td></tr><tr><td><table><tbody><tr><td><a id="char"></a><a id="nt-NormalChar"></a>[73]   </td><td><code>NormalChar</code></td><td>   ::=   </td><td><code>[^.\?*+{}()|#x5B#x5D]</code></td><td>/*  N.B.:  #x5B = '<code>[</code>',
    #x5D = '<code>]</code>'  */</td></tr></tbody></table></td></tr></tbody></table>

### <a id="charcter-classes"></a>G.4 Character Classes

        G.4.1 [Character class expressions](#charclassexps)
         G.4.2 [Character Class Escapes](#cces)
             G.4.2.1 [Single-character escapes](#cces-sce)
             G.4.2.2 [Category escapes](#cces-catesc)
             G.4.2.3 [Block escapes](#cces-blockesc)
             G.4.2.4 [Unrecognized category escapes](#sec-unrecognized-catesc)
             G.4.2.5 [Multi-character escapes](#cces-mce)

<a id="dt-charclass">[Definition:]</a>A **character class** is an [atom](#dt-atom) *R* that identifies a set of characters *C*(*R*).  The set of strings *L*(*R*).  denoted by a character class *R* contains one single-character string "*c*" for each character *c* in *C*(*R*). <a id="anchor11125c"></a>

<table><tbody><tr><td>Character Class</td></tr><tr><td><table><tbody><tr><td><a id="charClass"></a><a id="nt-charClass"></a>[74]   </td><td><code>charClass</code></td><td>   ::=   </td><td><code>
    <a href="#nt-SingleCharEsc">SingleCharEsc</a> |
    <a href="#nt-charClassEsc">charClassEsc</a> |
    <a href="#nt-charClassExpr">charClassExpr</a> |
    <a href="#nt-WildcardEsc">WildcardEsc</a>
</code></td></tr></tbody></table></td></tr></tbody></table>

A character class is either a [single-character escape](#dt-cces1) or a [character class escape](#dt-cces) or a [character class expression](#dt-charexpr) or a [wildcard character](#dt-wcchar).

**Note:** The rules for which characters must be escaped and which can represent themselves are different when inside a [character class expression](#dt-charexpr); some [normal characters](#dt-normalc) must be escaped and some [metacharacters](#dt-metac) need not be.

#### <a id="charclassexps"></a>G.4.1 Character class expressions

<a id="dt-charexpr">[Definition:]</a>A **character class expression** ([*charClassExpr*](#nt-charClassExpr)) is a [character group](#dt-chargroup) surrounded by '`[`' and '`]`' characters.  For all character groups *G*,  `[` *G* `]`  is a valid **character class expression**, identifying the set of characters *C*(`[`*G*`]`) = *C*(*G*).

<table><tbody><tr><td>Character Class Expression</td></tr><tr><td><table><tbody><tr><td><a id="charClassExpr"></a><a id="nt-charClassExpr"></a>[75]   </td><td><code>charClassExpr</code></td><td>   ::=   </td><td><code>'[' <a href="#nt-charGroup">charGroup</a> ']'</code></td></tr></tbody></table></td></tr></tbody></table>

 <a id="dt-chargroup">[Definition:]</a> A **character group** ([*charGroup*](#nt-charGroup)) starts with either a [positive character group](#dt-poschargroup) or a [negative character group](#dt-negchargroup), and is optionally followed by a subtraction operator '`-`' and a further [character class expression](#dt-charexpr).  <a id="dt-ccsub">[Definition:]</a>A [character group](#dt-chargroup) that contains a subtraction operator is referred to as a **character class subtraction**.

<table><tbody><tr><td>Character Group</td></tr><tr><td><table><tbody><tr><td><a id="chargroup"></a><a id="nt-charGroup"></a>[76]   </td><td><code>charGroup</code></td><td>   ::=   </td><td><code>(
    <a href="#nt-posCharGroup">posCharGroup</a> |
    <a href="#nt-negCharGroup">negCharGroup</a>
    )

    ( '-' <a href="#nt-charClassExpr">charClassExpr</a> )?

</code></td></tr></tbody></table></td></tr></tbody></table>

If the first character in a [*charGroup*](#nt-charGroup) is '`^`', this is taken as indicating that the [*charGroup*](#nt-charGroup) starts with a [*negCharGroup*](#nt-negCharGroup).  A [*posCharGroup*](#nt-posCharGroup) can itself start with '`^`' but only when it appears within a [*negCharGroup*](#nt-negCharGroup), that is, when the '`^`' is preceded by another '`^`'.

**Note:** For example, the string '`[^X]`' is ambiguous according the grammar rules, denoting either a character class consisting of a negative character group with '`X`' as a member, or a positive character class with '`X`' and '`^`' as members.  The normative prose rule just given requires that the first interpretation be taken.The string '`[^]`' is unambiguous: the grammar recognizes it as a character class expression containing a positive character group containing just the character '`^`'.  But the grammatical derivation of the string violates the rule just given, so the string '`[^]`' **must not** be accepted as a regular expression.

A '`-`' character is recognized as a subtraction operator (and hence, as terminating the [*posCharGroup*](#nt-posCharGroup) or [*negCharGroup*](#nt-negCharGroup)) if it is immediately followed by a '`[`' character.

 For any [positive character group](#dt-poschargroup) or [negative character group](#dt-negchargroup) *G*, and any [character class expression](#dt-charexpr) *C*,  *G* `-` *C*  is a valid [character group](#dt-chargroup), identifying the set of all characters in *C*(*G*) that are not in *C*(*C*).

<a id="dt-poschargroup">[Definition:]</a>A **positive character group** consists of one or more [character group parts](#dt-cgpart), concatenated together. The set of characters identified by a **positive character group** is the union of all of the sets identified by its constituent [character group parts](#dt-cgpart).

<table><tbody><tr><td>Positive Character Group</td></tr><tr><td><table><tbody><tr><td><a id="poschargroup"></a><a id="nt-posCharGroup"></a>[77]   </td><td><code>posCharGroup</code></td><td>   ::=   </td><td><code>
    (

    <a href="#nt-charGroupPart">charGroupPart</a>

    )+
</code></td></tr></tbody></table></td></tr></tbody></table>

<table><thead><tr><th>
    For all

    <a href="#dt-charrange">·character
        ranges·</a>
    R, all

    <a href="#dt-cces">·character class
        escapes·</a>
    E,
    and all

    <a href="#dt-poschargroup">·positive character groups·</a>
    P, valid

    <a href="#dt-poschargroup">·positive
        charater groups·</a>
    G are:
</th><th>
    Identifying the set of characters
    C(G)
    containing:
</th></tr></thead><tbody><tr><td>R</td><td>all characters in
    C(R)</td></tr><tr><td>E</td><td>all characters in
    C(E)</td></tr><tr><td>RP</td><td>all characters in
    C(R)
    and all characters in
    C(P)</td></tr><tr><td>EP</td><td>all characters in
    C(E)
    and all characters in
    C(P)</td></tr></tbody></table>

<a id="dt-negchargroup">[Definition:]</a>A **negative character group** ([*negCharGroup*](#nt-negCharGroup)) consists of a '`^`' character followed by a [positive character group](#dt-poschargroup). The set of characters identified by a negative character group *C*(`^`*P*) is the set of all characters that are *not* in *C*(*P*).

<table><tbody><tr><td>Negative Character Group</td></tr><tr><td><table><tbody><tr><td><a id="negchargroup"></a><a id="nt-negCharGroup"></a>[78]   </td><td><code>negCharGroup</code></td><td>   ::=   </td><td><code>'^' <a href="#nt-posCharGroup">posCharGroup</a></code></td></tr></tbody></table></td></tr></tbody></table>

<a id="dt-cgpart">[Definition:]</a>A **character group part** ([*charGroupPart*](#nt-charGroupPart)) is any of: a single unescaped character ([*SingleCharNoEsc*](#nt-SingleCharNoEsc)), a single escaped character ([*SingleCharEsc*](#nt-SingleCharEsc)), a character class escape ([*charClassEsc*](#nt-charClassEsc)), or a character range ([*charRange*](#nt-charRange)). <a id="anchor11125a"></a>

<table><tbody><tr><td>Character Group Part</td></tr><tr><td><table><tbody><tr><td><a id="charGroupPart"></a><a id="nt-charGroupPart"></a>[79]   </td><td><code>charGroupPart</code></td><td>   ::=   </td><td><code><a href="#nt-singleChar">singleChar</a>
    | <a href="#nt-charRange">charRange</a>
    |
    <a href="#nt-charClassEsc">charClassEsc</a>

</code></td></tr></tbody><tbody><tr><td><a id="singleChar"></a><a id="nt-singleChar"></a>[80]   </td><td><code>singleChar</code></td><td>   ::=   </td><td><code><a href="#nt-SingleCharEsc">SingleCharEsc</a> | <a href="#nt-SingleCharNoEsc">SingleCharNoEsc</a></code></td></tr></tbody></table></td></tr></tbody></table>

If a [*charGroupPart*](#nt-charGroupPart) starts with a [*singleChar*](#nt-singleChar) and this is immediately followed by a hyphen, then the following rules apply.

1. If the hyphen is immediately followed by '`[`', then the hyphen is not part of the [*charGroupPart*](#nt-charGroupPart): instead, it is recognized as a character-class subtraction operator.
2.  If the hyphen is immediately followed by '`]`', then the hyphen is recognized as a [*singleChar*](#nt-singleChar) and is part of the [*charGroupPart*](#nt-charGroupPart).
3.  If the hyphen is immediately followed by '`-[`', then the hyphen is recognized as a [*singleChar*](#nt-singleChar) and is part of the [*charGroupPart*](#nt-charGroupPart).
4.  Otherwise, the hyphen **must** be immediately followed by some [*singleChar*](#nt-singleChar) other than a hyphen. In this case the hyphen is not part of the [*charGroupPart*](#nt-charGroupPart); instead it is recognized, together with the immediately preceding and following instances of [*singleChar*](#nt-singleChar), as a [*charRange*](#nt-charRange).
5.  If the hyphen is followed by any other character sequence, then the string in which it occurs is not recognized as a regular expression.
 It is an error if either of the two [*singleChar*](#nt-singleChar)s in a [*charRange*](#nt-charRange) is a [*SingleCharNoEsc*](#nt-SingleCharNoEsc) comprising an unescaped hyphen. **Note:** The rule just given resolves what would otherwise be the ambiguous interpretation of some strings, e.g. '`[a-k-z]`'; it also constrains regular expressions in ways not expressed in the grammar. For example, the rule (not the grammar) excludes the string '`[--z]`' from the regular expression language defined here.

<a id="dt-charrange">[Definition:]</a>A **character range** *R* identifies a set of characters *C*(*R*) with UCS code points in a specified range.

<table><tbody><tr><td>Character Range</td></tr><tr><td><table><tbody><tr><td><a id="charrange"></a><a id="nt-charRange"></a>[81]   </td><td><code>charRange</code></td><td>   ::=   </td><td><code>

    <a href="#nt-singleChar">singleChar</a> '-' <a href="#nt-singleChar">singleChar</a>

</code></td></tr></tbody></table></td></tr></tbody></table>

 A [character range](#dt-charrange) in the form  *s*`-`*e*  identifies the set of characters with UCS code points greater than or equal to the code point of *s*, but not greater than the code point of *e*.

<table><tbody><tr><td>Single Unescaped Character</td></tr><tr><td><table><tbody><tr><td><a id="SingleCharNoEsc"></a><a id="nt-SingleCharNoEsc"></a>[82]   </td><td><code>SingleCharNoEsc</code></td><td>   ::=   </td><td><code>[^\#x5B#x5D] </code></td><td>/*  N.B.: 
    #x5B = '<code>[</code>',
    #x5D = '<code>]</code>'  */</td></tr></tbody></table></td></tr></tbody></table>

A single unescaped character ([*SingleCharNoEsc*](#nt-SingleCharNoEsc)) is any character except '`[`' or '`]`'. There are special rules, described earlier, that constrain the use of the characters '`-`' and '`^`' in order to disambiguate the syntax.

A single unescaped character identifies the singleton set of characters containing that character alone.

A single escaped character ([*SingleCharEsc*](#nt-SingleCharEsc)), when used within a character group, identifies the singleton set of characters containing the character denoted by the escape (see [Character Class Escapes (§G.4.2)](#cces)).

#### <a id="cces"></a>G.4.2 Character Class Escapes

<a id="dt-cces">[Definition:]</a>A **character class escape** is a short sequence of characters that identifies a predefined character class.  The valid character class escapes are the [multi-character escapes](#dt-ccesN), and the [category escapes](#dt-ccescat) (including the [block escapes](#dt-ccesblock)).

<table><tbody><tr><td>Character Class Escape</td></tr><tr><td><table><tbody><tr><td><a id="charclassesc"></a><a id="nt-charClassEsc"></a>[83]   </td><td><code>charClassEsc</code></td><td>   ::=   </td><td><code>
    (
    <a href="#nt-MultiCharEsc">MultiCharEsc</a> |
    <a href="#nt-catEsc">catEsc</a> |
    <a href="#nt-complEsc">complEsc</a>
    )
</code></td></tr></tbody></table></td></tr></tbody></table>

##### <a id="cces-sce"></a>G.4.2.1 Single-character escapes

Closely related to the character-class escapes are the single-character escapes. <a id="dt-cces1">[Definition:]</a>A **single-character escape** identifies a set containing only one character—usually because that character is difficult or impossible to write directly into a [regular expression](#dt-regex).

<table><tbody><tr><td>Single Character Escape</td></tr><tr><td><table><tbody><tr><td><a id="singlecharesc"></a><a id="nt-SingleCharEsc"></a>[84]   </td><td><code>SingleCharEsc</code></td><td>   ::=   </td><td><code>'\' [nrt\|.?*+(){}#x2D#x5B#x5D#x5E]</code></td><td>/* N.B.: 
    #x2D = '<code>-</code>', #x5B = '<code>[</code>', #x5D =
    '<code>]</code>', #x5E = '<code>^</code>' */</td></tr></tbody></table></td></tr></tbody></table>

<table><thead><tr><th>The valid
    <a href="#dt-cces1">·single
        character escapes·</a> R are:</th><th>Identifying the set of characters

    containing:</th></tr></thead><tbody><tr><td><code>\n</code></td><td>the newline character (#xA)</td></tr><tr><td><code>\r</code></td><td>the return character (#xD)</td></tr><tr><td><code>\t</code></td><td>the tab character (#x9)</td></tr><tr><td><code>\\</code></td><td><code>\</code></td></tr><tr><td><code>\|</code></td><td><code>|</code></td></tr><tr><td><code>\.</code></td><td><code>.</code></td></tr><tr><td><code>\-</code></td><td><code>-</code></td></tr><tr><td><code>\^</code></td><td><code>^</code></td></tr><tr><td><code>\?</code></td><td><code>?</code></td></tr><tr><td><code>\*</code></td><td><code>*</code></td></tr><tr><td><code>\+</code></td><td><code>+</code></td></tr><tr><td><code>\{</code></td><td><code>{</code></td></tr><tr><td><code>\}</code></td><td><code>}</code></td></tr><tr><td><code>\(</code></td><td><code>(</code></td></tr><tr><td><code>\)</code></td><td><code>)</code></td></tr><tr><td><code>\[</code></td><td><code>[</code></td></tr><tr><td><code>\]</code></td><td><code>]</code></td></tr></tbody></table>

##### <a id="cces-catesc"></a>G.4.2.2 Category escapes

<a id="dt-ccescat">[Definition:]</a> [[Unicode Database]](#UnicodeDB) specifies a number of possible values for the "General Category" property and provides mappings from code points to specific character properties.  The set containing all characters that have property *X* can be identified with a **category escape** `\p{`*X*`}` (using a lower-case 'p').  The complement of this set is specified with the **category escape**  `\P{`*X*`}` (using an upper-case 'P').  For all *X*, if *X* is a recognized character-property code, then `[\P{*X*}]` = `[^\p{*X*}]`.

<table><tbody><tr><td>Category Escape</td></tr><tr><td><table><tbody><tr><td><a id="catesc"></a><a id="nt-catEsc"></a>[85]   </td><td><code>catEsc</code></td><td>   ::=   </td><td><code>'\p{' <a href="#nt-charProp">charProp</a> '}'</code></td></tr></tbody><tbody><tr><td><a id="complesc"></a><a id="nt-complEsc"></a>[86]   </td><td><code>complEsc</code></td><td>   ::=   </td><td><code>'\P{' <a href="#nt-charProp">charProp</a> '}'</code></td></tr></tbody><tbody><tr><td><a id="charprop"></a><a id="nt-charProp"></a>[87]   </td><td><code>charProp</code></td><td>   ::=   </td><td><code><a href="#nt-IsCategory">IsCategory</a> | <a href="#nt-IsBlock">IsBlock</a></code></td></tr></tbody></table></td></tr></tbody></table>

 [[Unicode Database]](#UnicodeDB) is subject to future revision.  For example, the mapping from code points to character properties might be updated. All [minimally conforming](#dt-minimally-conforming) processors [**must**](#dt-must) support the character properties defined in the version of [[Unicode Database]](#UnicodeDB) cited in the normative references ([Normative (§K.1)](#normative-biblio)) or in some later version of the Unicode database.  Implementors are encouraged to support the character properties defined in any later versions. When the implementation supports multiple versions of the Unicode database, and they differ in salient respects (e.g. different properties are assigned to the same character in different versions of the database), then it is [implementation-defined](#key-impl-def) which set of property definitions is used for any given assessment episode.

**Note:** In order to benefit from continuing work on the Unicode database, a conforming implementation might by default use the latest supported version of the character properties. In order to maximize consistency with other implementations of this specification, however, an implementation might choose to provide [user options](#dt-useroption) to specify the use of the version of the database cited in the normative references. The `PropertyAliases.txt` and `PropertyValueAliases.txt` files of the Unicode database may be helpful to implementors in this connection.

 For convenience, the following table lists the values of the "General Category" property in the version of [[Unicode Database]](#UnicodeDB) cited in the normative references ([Normative (§K.1)](#normative-biblio)).  The properties with single-character names are not defined in [[Unicode Database]](#UnicodeDB).  The value of a single-character property is the union of the values of all the two-character properties whose first character is the character in question.  For example, for `N`, the union of `Nd`, `Nl` and `No`.

**Note:** As of this publication the Java regex library does *not* include `Cn` in its definition of `C`, so that definition cannot be used without modification in conformant implementations.

<table><tbody><tr><th>Category</th><th>Property</th><th>Meaning</th></tr><tr><td rowspan="6">Letters</td><td>L</td><td>All Letters</td></tr><tr><td>Lu</td><td>uppercase</td></tr><tr><td>Ll</td><td>lowercase</td></tr><tr><td>Lt</td><td>titlecase</td></tr><tr><td>Lm</td><td>modifier</td></tr><tr><td>Lo</td><td>other</td></tr><tr><td colspan="3"> </td></tr><tr><td rowspan="4">Marks</td><td>M</td><td>All Marks</td></tr><tr><td>Mn</td><td>nonspacing</td></tr><tr><td>Mc</td><td>spacing combining</td></tr><tr><td>Me</td><td>enclosing</td></tr><tr><td colspan="3"> </td></tr><tr><td rowspan="4">Numbers</td><td>N</td><td>All Numbers</td></tr><tr><td>Nd</td><td>decimal digit</td></tr><tr><td>Nl</td><td>letter</td></tr><tr><td>No</td><td>other</td></tr><tr><td colspan="3"> </td></tr><tr><td rowspan="8">Punctuation</td><td>P</td><td>All Punctuation</td></tr><tr><td>Pc</td><td>connector</td></tr><tr><td>Pd</td><td>dash</td></tr><tr><td>Ps</td><td>open</td></tr><tr><td>Pe</td><td>close</td></tr><tr><td>Pi</td><td>initial quote
    (may behave like Ps or Pe depending on usage)</td></tr><tr><td>Pf</td><td>final quote
    (may behave like Ps or Pe depending on usage)</td></tr><tr><td>Po</td><td>other</td></tr><tr><td colspan="3"> </td></tr><tr><td rowspan="4">Separators</td><td>Z</td><td>All Separators</td></tr><tr><td>Zs</td><td>space</td></tr><tr><td>Zl</td><td>line</td></tr><tr><td>Zp</td><td>paragraph</td></tr><tr><td colspan="3"> </td></tr><tr><td rowspan="5">Symbols</td><td>S</td><td>All Symbols</td></tr><tr><td>Sm</td><td>math</td></tr><tr><td>Sc</td><td>currency</td></tr><tr><td>Sk</td><td>modifier</td></tr><tr><td>So</td><td>other</td></tr><tr><td colspan="3"> </td></tr><tr><td rowspan="5">Other</td><td>C</td><td>All Others</td></tr><tr><td>Cc</td><td>control</td></tr><tr><td>Cf</td><td>format</td></tr><tr><td>Co</td><td>private use</td></tr><tr><td>Cn</td><td>not assigned</td></tr></tbody></table>

<table><tbody><tr><td>Categories</td></tr><tr><td><table><tbody><tr><td><a id="cats"></a><a id="nt-IsCategory"></a>[88]   </td><td><code>IsCategory</code></td><td>   ::=   </td><td><code>
    <a href="#nt-Letters">Letters</a> |
    <a href="#nt-Marks">Marks</a> |
    <a href="#nt-Numbers">Numbers</a> |
    <a href="#nt-Punctuation">Punctuation</a> |
    <a href="#nt-Separators">Separators</a> |
    <a href="#nt-Symbols">Symbols</a> |
    <a href="#nt-Others">Others</a>
</code></td></tr></tbody><tbody><tr><td><a id="lets"></a><a id="nt-Letters"></a>[89]   </td><td><code>Letters</code></td><td>   ::=   </td><td><code>'L' [ultmo]?</code></td></tr></tbody><tbody><tr><td><a id="marks"></a><a id="nt-Marks"></a>[90]   </td><td><code>Marks</code></td><td>   ::=   </td><td><code>'M' [nce]?</code></td></tr></tbody><tbody><tr><td><a id="nums"></a><a id="nt-Numbers"></a>[91]   </td><td><code>Numbers</code></td><td>   ::=   </td><td><code>'N' [dlo]?</code></td></tr></tbody><tbody><tr><td><a id="punc"></a><a id="nt-Punctuation"></a>[92]   </td><td><code>Punctuation</code></td><td>   ::=   </td><td><code>'P' [cdseifo]?</code></td></tr></tbody><tbody><tr><td><a id="seps"></a><a id="nt-Separators"></a>[93]   </td><td><code>Separators</code></td><td>   ::=   </td><td><code>'Z' [slp]?</code></td></tr></tbody><tbody><tr><td><a id="syms"></a><a id="nt-Symbols"></a>[94]   </td><td><code>Symbols</code></td><td>   ::=   </td><td><code>'S' [mcko]?</code></td></tr></tbody><tbody><tr><td><a id="others"></a><a id="nt-Others"></a>[95]   </td><td><code>Others</code></td><td>   ::=   </td><td><code>'C' [cfon]?</code></td></tr></tbody></table></td></tr></tbody></table>

**Note:** The properties mentioned above exclude the Cs property.  The Cs property identifies "surrogate" characters, which do not occur at the level of the "character abstraction" that XML instance documents operate on.

##### <a id="cces-blockesc"></a>G.4.2.3 Block escapes

 [[Unicode Database]](#UnicodeDB) groups the code points of the Universal Character Set (UCS) into a number of blocks such as Basic Latin (i.e., ASCII), Latin-1 Supplement, Hangul Jamo, CJK Compatibility, etc.  The block-escape construct allows regular expressions to refer to sets of characters by the name of the block in which they appear, using a [normalized block name](#dt-normalized-block-name).

 <a id="dt-normalized-block-name">[Definition:]</a> For any Unicode block, the **normalized block name** of that block is the string of characters formed by stripping out white space and underbar characters from the block name as given in [[Unicode Database]](#UnicodeDB), while retaining hyphens and preserving case distinctions.

 <a id="dt-ccesblock">[Definition:]</a> A **block escape** expression denotes the set of characters in a given Unicode block. For any Unicode block *B*, with [normalized block name](#dt-normalized-block-name) *X*, the set containing all characters defined in block *B* can be identified with the **block escape** `\p{Is*X*}` (using lower-case 'p'). The complement of this set is denoted by the **block escape** `\P{Is*X*}` (using upper-case 'P'). For all *X*, if *X* is a normalized block name recognized by the processor, then `[\P{Is`*X*`}]` = `[^\p{Is`*X*`}]`.

<table><tbody><tr><td>Block Escape</td></tr><tr><td><table><tbody><tr><td><a id="blockesc"></a><a id="nt-IsBlock"></a>[96]   </td><td><code>IsBlock</code></td><td>   ::=   </td><td><code>'Is' [a-zA-Z0-9#x2D]+</code></td><td>/*  N.B.: 
    #x2D = '<code>-</code>' */</td></tr></tbody></table></td></tr></tbody></table>

[block escape](#dt-ccesblock)`\p{IsBasicLatin}`

**Note:** Current versions of the Unicode database recommend that whenever block names are being matched hyphens, underbars, and white space should be dropped and letters folded to a single case, so both the string '`BasicLatin`' and the string '`-- basic LATIN --`' will match the block name "Basic Latin". The handling of block names in block escapes differs from this behavior in two ways. First, the normalized block names defined in this specification do not suppress hyphens in the Unicode block names and do not level case distinctions. The normalized form of the block name '`Latin-1 Supplement`', for example, is thus '`Latin-1Supplement`', not '`latin1supplement`' or '`LATIN1SUPPLEMENT`'. Second, XSD processors are not required to perform any normalization at all upon the block name as given in the [block escape](#dt-ccesblock), so '`\p{Latin-1Supplement}`' will be recognized as a reference to the Latin-1 Supplement block, but '`\p{Is Latin-1 supplement}`' will not.

 [[Unicode Database]](#UnicodeDB) has been revised since XSD 1.0 was published, and is subject to future revision. In particular, the grouping of code points into blocks has changed, and may change again. All [minimally conforming](#dt-minimally-conforming) processors **must** support the blocks defined in the version of [[Unicode Database]](#UnicodeDB) cited in the normative references ([Normative (§K.1)](#normative-biblio)) or in some later version of the Unicode database. Implementors are encouraged to support the blocks defined in earlier and/or later versions of the Unicode Standard. When the implementation supports multiple versions of the Unicode database, and they differ in salient respects (e.g. different characters are assigned to a given block in different versions of the database), then it is [implementation-defined](#key-impl-def) which set of block definitions is used for any given assessment episode.

In particular, the version of [[Unicode Database]](#UnicodeDB) referenced in XSD 1.0 (namely, Unicode 3.1) contained a number of blocks which have been renamed in later versions of the database Since the older block names may appear in regular expressions within XSD 1.0 schemas, implementors are encouraged to support the superseded block names in XSD 1.1 processors for compatibility, either by default or [at user option](#dt-useroption). At the time this document was prepared, block names from Unicode 3.1 known to have been superseded in this way included:

- #x0370 - #x03FF: Greek
- #x20D0 - #x20FF: CombiningMarksforSymbols
- #xE000 - #xF8FF: PrivateUse
- #xF0000 - #xFFFFD: PrivateUse
- #x100000 - #x10FFFD: PrivateUse

 A tabulation of normalized block names for Unicode 2.0.0 and later is given in [[Unicode block names]](#unicode-escapes).

For the treatment of regular expressions containing unrecognized Unicode block names, see [Unrecognized category escapes (§G.4.2.4)](#sec-unrecognized-catesc).

##### <a id="sec-unrecognized-catesc"></a>G.4.2.4 Unrecognized category escapes

A string of the form "`\p{*S*}`" constitutes a [*catEsc*](#nt-catEsc) (category escape), and similarly a string of the form "`\P{*S*}`" constitutes a [*complEsc*](#nt-complEsc) (category-complement escape) only if the string *S* matches either [*IsCategory*](#nt-IsCategory) or [*IsBlock*](#nt-IsBlock).

**Note:** If an unknown string of characters is used in a category escape instead of a known character category code or a string matching the [*IsBlock*](#nt-IsBlock) production, the resulting string will (normally) not match the [*regExp*](#nt-regExp) production and thus not be a regular expression as defined in this specification. If the non-[*regExp*](#nt-regExp) string occurs where a regular expression is required, the schema document will be in [error](#dt-error).

Any string of hyphens, digits, and Basic Latin characters beginning with '`Is`' will match the non-terminal [*IsBlock*](#nt-IsBlock) and thus be allowed in a regular expression. Most of these strings, however, will not denote any Unicode block. Processors **should** issue a warning if they encounter a regular expression using a block name they do not recognize. Processors **may** [at user option](#dt-useroption) treat unrecognized block names as [errors](#dt-error) in the schema.

**Note:** Treating unrecognized block names as errors increases the likelihood that errors in spelling the block name will be detected and can be helpful in checking the correctness of schema documents. However, it also decreases the portability of schema documents among processors supporting different versions of [[Unicode Database]](#UnicodeDB); it is for this reason that processors are allowed to treat unrecognized block names as errors only when the user has explicitly requested this behavior.

If a string "`Is*X*`" matches the non-terminal [*IsBlock*](#nt-IsBlock) but *X* is not a recognized block name, then the expressions "`\p{Is*X*}`" and "`\P{Is*X*}`" each denote the set of all characters. Processors **may** [at user option](#dt-useroption) treat both "`\p{Is*X*}`" and "`\P{Is*X*}`" as denoting the empty set, instead of the set of all characters.

**Note:** The meaning defined for a block escape with an unrecognized block name makes it synonymous with the regular expression '`.|[\n\r]`'. A processor which does not recognize the block name will thus not enforce the constraint that the characters matched are in, or are not in, the block in question. Any string which satisfies the regular expression as written will be accepted, but not all strings accepted will actually satisfy the expression as written: some strings which do not satisfy the expression as written will also be accepted. So some invalid input will be wrongly identified as valid.If (at [user option](#dt-useroption)) the expressions are treated as denoting the empty set, then the converse is true: any string which fails to satisfy the expression as written will be rejected, but not all strings rejected by the processor will actually have failed to satisfy the expression as written. So some valid input will be wrongly identified as invalid.Which behavior is preferable in concrete circumstances depends on the relative cost of failure to accept valid input (false negatives) and failure to reject invalid input (false positives). It is for this reason that processors are allowed to provide [user options](#dt-useroption) to control the behavior. The principle of being liberal in accepting input (often called Postel's Law) suggests that the default behavior should be to accept strings not known to be invalid, rather than the converse; it is for this reason that block escapes with unknown block names should be treated as matching any character unless the user explicitly requests the alternative behavior.

##### <a id="cces-mce"></a>G.4.2.5 Multi-character escapes

<a id="dt-ccesN">[Definition:]</a>A **multi-character escape** provides a simple way to identify any of a commonly used set of characters:  <a id="dt-wcchar">[Definition:]</a> The **wildcard character** is a metacharacter which matches almost any single character:

<table><tbody><tr><td>Multi-Character Escape</td></tr><tr><td><table><tbody><tr><td><a id="multicharesc"></a><a id="nt-MultiCharEsc"></a>[97]   </td><td><code>MultiCharEsc</code></td><td>   ::=   </td><td><code>'\' [sSiIcCdDwW]</code></td></tr></tbody><tbody><tr><td><a id="wildcardesc"></a><a id="nt-WildcardEsc"></a>[98]   </td><td><code>WildcardEsc</code></td><td>   ::=   </td><td><code>'.'</code></td></tr></tbody></table></td></tr></tbody></table>

<table><thead><tr><th>Character sequence</th><th>Equivalent <a href="#dt-charclass">·character class·</a></th></tr></thead><tbody><tr><td><code>.</code></td><td>[^\n\r]</td></tr><tr><td><code>\s</code></td><td>[#x20\t\n\r]</td></tr><tr><td><code>\S</code></td><td>[^\s]</td></tr><tr><td><code>\i</code></td><td>
    the set of initial name characters, those
    <a href="#dt-match">·matched·</a> by
    <a href="https://www.w3.org/TR/xml11/#NT-NameStartChar">NameStartChar</a> in <a href="#XML">[XML]</a>

</td></tr><tr><td><code>\I</code></td><td>[^\i]</td></tr><tr><td><code>\c</code></td><td>

    the set of name characters, those
    <a href="#dt-match">·matched·</a> by
    <a href="https://www.w3.org/TR/xml11/#NT-NameChar">NameChar</a></td></tr><tr><td><code>\C</code></td><td>[^\c]</td></tr><tr><td><code>\d</code></td><td>\p{Nd}</td></tr><tr><td><code>\D</code></td><td>[^\d]</td></tr><tr><td><code>\w</code></td><td>
    [#x0000-#x10FFFF]-[\p{P}\p{Z}\p{C}]
    (all characters except the set of "punctuation",
    "separator" and "other" characters)
</td></tr><tr><td><code>\W</code></td><td>[^\w]</td></tr></tbody></table>

**Note:** The [regular expression](#dt-regex) language defined here does not attempt to provide a general solution to "regular expressions" over UCS character sequences.  In particular, it does not easily provide for matching sequences of base characters and combining marks. The language is targeted at support of "Level 1" features as defined in [[Unicode Regular Expression Guidelines]](#unicodeRegEx).  It is hoped that future versions of this specification will provide support for "Level 2" features.

## <a id="idef-idep"></a>H Implementation-defined and implementation-dependent features (normative)

### <a id="impl-def"></a>H.1 Implementation-defined features

The following features in this specification are [implementation-defined](#key-impl-def). Any software which claims to conform to this specification (or to the specification of any host language which embeds *XSD 1.1: Datatypes*) **must** describe how these choices have been exercised, in documentation which accompanies any conformance claim.

1. For the datatypes which depend on [[XML]](#XML) or [[Namespaces in XML]](#XMLNS), it is [implementation-defined](#key-impl-def) whether a conforming processor takes the relevant definitions from [[XML]](#XML) and [[Namespaces in XML]](#XMLNS), or from [[XML 1.0]](#XML1.0) and [[Namespaces in XML 1.0]](#XMLNS1.0). Implementations **may** support either the form of these datatypes based on version 1.0 of those specifications, or the form based on version 1.1, or both.
2. For the datatypes with infinite [value spaces](#dt-value-space), it is [implementation-defined](#key-impl-def) whether conforming processors set a limit on the size of the values supported. If such limits are set, they **must** be documented, and the limits **must** be equal to, or exceed, the minimal limits specified in [Partial Implementation of Infinite Datatypes (§5.4)](#partial-implementation). .
3. It is [implementation-defined](#key-impl-def) whether [primitive](#dt-primitive) datatypes other than those defined in this specification are supported.For each [implementation-defined](#key-impl-def) datatype, a [Simple Type Definition](#std) **must** be specified which conforms to the rules given in [Built-in Simple Type Definitions (§4.1.6)](#builtin-stds).  In addition, the following information **must** be provided:

  1. The nature of the datatype's [lexical space](#dt-lexical-space), [value space](#dt-value-space), and [lexical mapping](#dt-lexical-mapping).
  2. The nature of the equality relation; in particular, how to determine whether two values which are not identical are equal.**Note:** There is no requirement that equality be distinct from identity, but it **may** be.
  3. The values of the [fundamental facets](#dt-fundamental-facet).
  4. Which of the [constraining facets](#dt-constraining-facet) defined in this specification are applicable to the datatype (and **may** thus be used in [facet-based restriction](#dt-fb-restriction) from it), and what they mean when applied to it.
  5. If [implementation-defined](#key-impl-def) [constraining facets](#dt-constraining-facet) are supported, which of those [constraining facets](#dt-constraining-facet) are applicable to the datatype, and what they mean when applied to it.
  6. What URI reference (more precisely, what [anyURI](#anyURI) value) is to be used to refer to the datatype, analogous to those provided for the datatypes defined here in section [Built-in Datatypes and Their Definitions (§3)](#built-in-datatypes).**Note:** It is convenient if the URI for a datatype and the [expanded name](https://www.w3.org/TR/2004/REC-xml-names11-20040204/#dt-expname) of its simple type definition are related by a simple mapping, like the URIs given for the [built-in](#dt-built-in) datatypes in [Built-in Datatypes and Their Definitions (§3)](#built-in-datatypes). However, this is not a requirement.
  7. For each [constraining facet](#dt-constraining-facet) given a value for the new [primitive](#dt-primitive), what URI reference (more precisely, what [anyURI](#anyURI) value) is to be used to refer to the usage of that facet on the datatype, analogous to those provided, for the [built-in](#dt-built-in) datatypes, in section [Built-in Datatypes and Their Definitions (§3)](#built-in-datatypes).**Note:** As specified normatively elsewhere, the set of facets given values will at the very least include the [whiteSpace](#f-w) facet.
 The [value space](#dt-value-space) of the [primitive](#dt-primitive) datatype **must** be disjoint from those of the other [primitive](#dt-primitive) datatypes.The [lexical mapping](#dt-lexical-mapping) defined for an [implementation-defined](#key-impl-def) primitive **must** be a total function from the [lexical space](#dt-lexical-space) onto the [value space](#dt-value-space). That is, (1) each [literal](#dt-literal) in the [lexical space](#dt-lexical-space) **must** map to exactly one value, and (2) each value **must** be the image of at least one member of the [lexical space](#dt-lexical-space), and **may** be the image of more than one.For consistency with the [constraining facets](#dt-constraining-facet) defined here, implementors who define new [primitive](#dt-primitive) datatypes **should** allow the [pattern](#dt-pattern) and [enumeration](#dt-enumeration) facets to apply. The implementor **should** specify a [canonical mapping](#dt-canonical-mapping) for the datatype if practicable.
4. It is [implementation-defined](#key-impl-def) whether [constraining facets](#dt-constraining-facet) other than those defined in this specification are supported.For each [implementation-defined](#key-impl-def) facet, the following information **must** be provided:

  1. What properties the facet has, viewed as a schema component.**Note:** For most [implementation-defined](#key-impl-def) facets, the structural pattern used for most [constraining facets](#dt-constraining-facet) defined in this specification is expected to be satisfactory, but other structures **may** be specified.
  2. Whether the facet is a [pre-lexical](#dt-pre-lexical), [lexical](#dt-lexical), or [value-based](#dt-value-based) facet.
  3. Whether restriction of the facet takes the form of replacing a less restrictive facet value with a more restrictive value (as in the [minInclusive](#dt-minInclusive) and most other [constraining facets](#dt-constraining-facet) defined in this specification) or of adding new values to a set of facet values (as for the [pattern](#dt-pattern) facet). In the former case, the information provided **must** also specify how to determine which of two given values is more restrictive (and thus can be used to restrict the other).When an [implementation-defined](#key-impl-def) facet is used in [facet-based restriction](#dt-fb-restriction), the new value **must** be at least as restrictive as the existing value, if any.**Note:** The effect of the preceding paragraph is to ensure that a type derived by [facet-based restriction](#dt-fb-restriction) using an [implementation-defined](#key-impl-def) facet does not allow, or appear to allow, values not present in the [base type](#dt-basetype).
  4. What [primitive](#dt-primitive) datatypes the new [constraining facet](#dt-constraining-facet) applies to, and what it means when applied to them.For a [pre-lexical](#dt-pre-lexical) facet, how to compute the result of applying the facet value to any given [literal](#dt-literal).For a [pre-lexical](#dt-pre-lexical) facet, the order in which it is applied to [literals](#dt-literal), relative to other [pre-lexical](#dt-pre-lexical) facets. For a [lexical](#dt-lexical) facet, how to tell whether any given [literal](#dt-literal) is facet-valid with respect to it.For a [value-based](#dt-value-based) facet, how to tell whether any given value in the relevant [primitive](#dt-primitive) datatypes is facet-valid with respect to it.**Note:** The host language **may** choose to specify that [implementation-defined](#key-impl-def) [constraining facets](#dt-constraining-facet) are applicable to [built-in](#dt-built-in) [primitive](#dt-primitive) datatypes; this information is necessary to make the [implementation-defined](#key-impl-def) facet usable in such host languages.
  5. What URI reference (more precisely, what [anyURI](#anyURI) value) is to be used to refer to the facet, analogous to those provided for the datatypes defined here in section [Built-in Datatypes and Their Definitions (§3)](#built-in-datatypes).
  6. What element is to be used in XSD schema documents to apply the facet in the course of [facet-based restriction](#dt-fb-restriction). A schema document **must** be provided with an element declaration for each [implementation-defined](#key-impl-def) facet; the element declarations **should** specify `xs:facet` as their substitution-group head.**Note:** The elements' [expanded names](https://www.w3.org/TR/2004/REC-xml-names11-20040204/#dt-expname) are used by the condition-inclusion mechanism of [[XSD 1.1 Part 1: Structures]](#structural-schemas) to allow schema authors to test whether a particular facet is supported and adjust the schema document's contents accordingly.
 [Implementation-defined](#key-impl-def) [pre-lexical](#dt-pre-lexical) facets **must not**, when applied to [literals](#dt-literal) which have been whitespace-normalized by the [whiteSpace](#f-w) facet, produce [literals](#dt-literal) which are no longer whitespace-normalized.
5. It is [implementation-defined](#key-impl-def) whether an implementation of this specification supports other versions of the Unicode database [[Unicode Database]](#UnicodeDB) in addition to the version cited normatively in the normative references ([Normative (§K.1)](#normative-biblio)). If an implementation supports additional versions of the Unicode database, it is [implementation-defined](#key-impl-def) which character properties and which block name definitions are used in a given validity assessment. It is [implementation-defined](#key-impl-def) whether an implementation is capable, [at user option](#dt-useroption), of treating unrecognized block names as errors in a schema.It is [implementation-defined](#key-impl-def) whether an implementation is capable, [at user option](#dt-useroption), of treating unrecognized category escapes as denoting the empty set instead of the set of all characters.
**Note:** It follows from the above that each [implementation-defined](#key-impl-def) [primitive](#dt-primitive) datatype and each [implementation-defined](#key-impl-def) constraining facet has an [expanded name](https://www.w3.org/TR/2004/REC-xml-names11-20040204/#dt-expname). These [expanded names](https://www.w3.org/TR/2004/REC-xml-names11-20040204/#dt-expname) are used by the condition-inclusion mechanism of [[XSD 1.1 Part 1: Structures]](#structural-schemas) to allow schema authors to test whether a particular datatype or facet is supported and adjust the schema document's contents accordingly.

### <a id="impl-dep"></a>H.2 Implementation-dependent features

The following features in this specification are [implementation-dependent](#key-impl-dep). Software which claims to conform to this specification (or to the specification of any host language which embeds *XSD 1.1: Datatypes*) **may** describe how these choices have been exercised, in documentation which accompanies any conformance claim.

1. When multiple errors are encountered in type definitions or elsewhere, it is [implementation-dependent](#key-impl-dep) how many of the errors are reported (as long as at least one error is reported), and which, what form the report of errors takes, and how much detail is included.

## <a id="changes"></a>I Changes since version 1.0

### <a id="sec-chdtfacets"></a>I.1 Datatypes and Facets

In order to align this specification with those being prepared by the XSL and XML Query Working Groups, a new datatype named [anyAtomicType](#anyAtomicType) has been introduced; it serves as the base type definition for all [primitive](#dt-primitive) [atomic](#dt-atomic) datatypes.

The treatment of datatypes has been made more precise and explicit; most of these changes affect the section on [Datatype System (§2)](#typesystem). Definitions have been revised thoroughly and technical terms are used more consistently.

The (numeric) equality of values is now distinguished from the identity of the values themselves; this allows [float](#float) and [double](#double) to treat positive and negative zero as distinct values, but nevertheless to treat them as equal for purposes of bounds checking. This allows a better alignment with the expectations of users working with IEEE floating-point binary numbers.

The [{value}](#ff-b-value) of the [bounded](#ff-b) component for ***list*** datatypes is now always ***false***, reflecting the fact that no ordering is prescribed for [list](#dt-list) datatypes, and so they cannot be bounded using the facets defined by this specification.

Units of length have been specified for all datatypes that are permitted the length constraining facet.

The use of the namespace `http://www.w3.org/2001/XMLSchema-datatypes` has been deprecated. The definition of a namespace separate from the main namespace defined by this specification proved not to be necessary or helpful in facilitating the use, by other specifications, of the datatypes defined here, and its use raises a number of difficult unsolved practical questions.

 An [assertions](#f-a) facet has been added, to allow schema authors to associated assertions with simple type definitions, analogous to those allowed by [[XSD 1.1 Part 1: Structures]](#structural-schemas) for complex type definitions.

The discussion of whitespace handling in [whiteSpace (§4.3.6)](#rf-whiteSpace) makes clearer that when the value is **collapse**, [literals](#dt-literal) consisting solely of whitespace characters are reduced to the empty string; the earlier formulation has been misunderstood by some implementors.

 Conforming implementations **may** now support [primitive](#dt-primitive) datatypes and facets in addition to those defined here.

### <a id="sec-chnum"></a>I.2 Numerical Datatypes

As noted above, positive and negative zero, [float](#float) and [double](#double) are now treated as distinct but arithmetically equal values.

The description of the lexical spaces of [unsignedLong](#unsignedLong), [unsignedInt](#unsignedInt), [unsignedShort](#unsignedShort), and [unsignedByte](#unsignedByte) has been revised to agree with the schema for schemas by allowing for the possibility of a leading sign.

The [float](#float) and [double](#double) datatypes now follow IEEE 754 implementation practice more closely; in particular, negative and positive zero are now distinct values, although arithmetically equal. Conversely, NaN is identical but not arithmetically equal to itself.

 The character sequence '`+INF`' has been added to the lexical spaces of [float](#float) and [double](#double).

### <a id="sec-chdt"></a>I.3 Date/time Datatypes

The treatment of [dateTime](#dateTime) and related datatypes has been changed to provide a more explicit account of the value space in terms of seven numeric properties. The most important substantive change is that values now explicitly retain information about the time zone offset indicated in the lexical form; this allows better alignment with the treatment of such values in [[XQuery 1.0 and XPath 2.0 Functions and Operators]](#F_O).

At the suggestion of the [W3C OWL Working Group](https://www.w3.org/2007/OWL/wiki/OWL_Working_Group), a [explicitTimezone](#f-tz) facet has been added to allow date/time datatypes to be restricted by requiring or forbidding an explicit time zone offset from UTC, instead of making it optional. The [dateTimeStamp](#dateTimeStamp) datatype has been defined using this facet.

The treatment of the date/time datatype includes a carefully revised definition of order that ensures that for repeating datatypes ([time](#time), [gDay](#gDay), etc.), timezoned values will be compared as though they are on the same "calendar day" ("local" property values) so that in any given timezone, the days start at the local midnight and end just before local midnight.  Days do not run from 00:00:00Z to 24:00:00Z in timezones other than Z.

The lexical representation '`0000`' for years is recognized and maps to the year 1 BCE; '`-0001`' maps to 2 BCE, etc. This is a change from version 1.0 of this specification, in order to align with established practice (the so-called "astronomical year numbering") and [[ISO 8601]](#ISO8601).

 Algorithms for arithmetic involving [dateTime](#dateTime) and [duration](#duration) values have been provided, and corrections made to the [*timeOnTimeline*](#vp-dt-timeOnTimeline) function.

The treatment of leap seconds is no longer [implementation-defined](#key-impl-def): the date/time types described here do not include leap-second values.

 At the suggestion of the [W3C Internationalization Core Working Group](https://www.w3.org/International/core/), most references to "time zone" have been replaced with references to "time zone offset"; this resolves issue [4642 Terminology: zone offset versus time zone](https://www.w3.org/Bugs/Public/show_bug.cgi?id=4642).

 A number of syntactic and semantic errors in some of the regular expressions given to describe the lexical spaces of the [primitive](#dt-primitive) datatypes (most notably the date/time datatypes) have been corrected.

The lexical mapping for times of the form '`24:00:00`' (with or without a trailing decimal point and zeroes) has been specified explicitly.

### <a id="sec-chother"></a>I.4 Other changes

Support has been added for [[XML]](#XML) version 1.1 and [[Namespaces in XML]](#XMLNS) version 1.1. The datatypes which depend on [[XML]](#XML) and [[Namespaces in XML]](#XMLNS) may now be used with the definitions provided by the 1.1 versions of those specifications, as well as with the definitions in the 1.0 versions. It is [implementation-defined](#key-impl-def) whether software conforming to this specification supports the definitions given in version 1.0, or in version 1.1, of [[XML]](#XML) and [[Namespaces in XML]](#XMLNS).

To reduce confusion and avert a widespread misunderstanding, the normative references to various W3C specifications now state explicitly that while the reference describes the particular edition of a specification current at the time this specification is published, conforming implementations of this specification are not required to ignore later editions of the other specification but instead **may** support later editions, thus allowing users of this specification to benefit from corrections to other specifications on which this one depends.

 The reference to the Unicode Database [[Unicode Database]](#UnicodeDB) has been updated from version 4.1.0 to version 5.1.0, at the suggestion of the [W3C Internationalization Core Working Group](https://www.w3.org/International/core/)

 References to various other specifications have also been updated.

The account of the value space of [duration](#duration) has been changed to specify that values consist only of two numbers (the number of months and the number of seconds) rather than six (years, months, days, hours, minutes, seconds). This allows clearly equivalent durations like P2Y and P24M to have the same value.

Two new totally ordered restrictions of [duration](#duration) have been defined: [yearMonthDuration](#yearMonthDuration), defined in [yearMonthDuration (§3.4.26)](#yearMonthDuration), and [dayTimeDuration](#dayTimeDuration), defined in [dayTimeDuration (§3.4.27)](#dayTimeDuration). This allows better alignment with the treatment of durations in [[XQuery 1.0 and XPath 2.0 Functions and Operators]](#F_O).

The XML representations of the [primitive](#dt-primitive) and [ordinary](#dt-ordinary) built-in datatypes have been moved out of the schema document for schema documents in [Schema for Schema Documents (Datatypes) (normative) (§A)](#schema) and into a different appendix ([Illustrative XML representations for the built-in simple type definitions (§C)](#prim.nxsd)).

Numerous minor corrections have been made in response to comments on earlier working drafts.

The treatment of topics handled both in this specification and in [[XSD 1.1 Part 1: Structures]](#structural-schemas) has been revised to align the two specifications more closely.

Several references to other specifications have been updated to refer to current versions of those specifications, including [[XML]](#XML), [[Namespaces in XML]](#XMLNS), [[RFC 3986]](#RFC3986), [[RFC 3987]](#RFC3987), and [[RFC 3548]](#RFC3548).

Requirements for the datatype-validity of values of type [language](#language) have been clarified.

Explicit definitions have been provided for the lexical and [canonical mappings](#dt-canonical-mapping) of most of the primitive datatypes.

Schema Component Constraint [enumeration facet value required for NOTATION (§3.3.19)](#enumeration-required-notation), which restricts the use of [NOTATION](#NOTATION) to validate [literals](#dt-literal) without first enumerating a set of values, has been clarified.

Some errors in the definition of regular-expression metacharacters have been corrected.

The descriptions of the [pattern](#f-p) and [enumeration](#f-e) facets have been revised to make clearer how values from different derivation steps are combined.

A warning against using the whitespace facet for tokenizing natural-language data has been added on the request of the W3C Internationalization Working Group.

In order to correct an error in version 1 of this specification and of [[XSD 1.1 Part 1: Structures]](#structural-schemas), [unions](#dt-union) are no longer forbidden to be members of other [unions](#dt-union). Descriptions of [union](#dt-union) types have also been changed to reflect the fact that [unions](#dt-union) can be derived by restricting other [unions](#dt-union). The concepts of [transitive membership](#dt-transitivemembership) (the members of all members, recursively) and [basic member](#dt-basicmember) (those datatypes in the transitive membership which are not [unions](#dt-union)) have been introduced and are used.

 The requirements of conformance have been clarified in various ways. A distinction is now made between [implementation-defined](#key-impl-def) and [implementation-dependent](#key-impl-dep) features, and a list of such features is provided in [Implementation-defined and implementation-dependent features (normative) (§H)](#idef-idep). Requirements imposed on host languages which use or incorporate the datatypes defined by this specification are defined.

The definitions of **must**, **must not**, and [error](#dt-error) have been changed to specify that processors **must** detect and report errors in schemas and schema documents (although the quality and level of detail in the error report is not constrained).

 The lexical mapping of the [QName](#QName) datatype, in particular its dependence on the namespace bindings in scope at the place where the [literal](#dt-literal) appears, has been clarified.

 The characterization of [lexical mappings](#dt-lexical-mapping) has been revised to say more clearly when they are functions and when they are not, and when (in the [special](#dt-special) datatypes) there are values in the [value space](#dt-value-space) not mapped to by any members of the [lexical space](#dt-lexical-space).

The nature of equality and identity of lists has been clarified.

 Enumerations, identity constraints, and value constraints now treat both identical values and equal values as being the same for purposes of validation. This affects primitive datatypes in which identity and equality are not the same. Positive and negative zero, for example, are not treated as different for purposes of keys, keyrefs, or uniqueness constraints, and an enumeration which includes either zero will accept either zero.

 The mutual relations of lists and unions have been clarified, in particular the restrictions on what kinds of datatypes **may** appear as the [item type](#dt-itemType) of a list or among the [member types](#dt-memberTypes) of a union.

 Unions with no member types (and thus with empty [value space](#dt-value-space) and [lexical space](#dt-lexical-space)) are now explicitly allowed.

 Cycles in the definitions of [unions](#dt-union) and in the derivation of simple types are now explicitly forbidden.

 A number of minor errors and obscurities have been fixed.

## <a id="normative-glossary"></a>J Glossary (non-normative)

The listing below is for the benefit of readers of a printed version of this document: it collects together all the definitions which appear in the document above.

**[Constraint on Schemas](#dt-cos)**
   Constraints on the schema components themselves, i.e. conditions components [**must**](#dt-must) satisfy to be components at all. Largely to be found in [Datatype components (§4)](#datatype-components).

**[Schema Representation Constraint](#dt-src)**
   Constraints on the representation of schema components in XML.  Some but not all of these are expressed in [Schema for Schema Documents (Datatypes) (normative) (§A)](#schema) and [DTD for Datatype Definitions (non-normative) (§B)](#dtd-for-datatypeDefs).

**[UTC](#dt-utc)**
  **Universal Coordinated Time** (**UTC**) is an adaptation of TAI which closely approximates UT1 by adding [leap-seconds](#dt-leapsec) to selected [UTC](#dt-utc) days.

**[Validation Rule](#dt-cvc)**
   Constraints expressed by schema components which information items [**must**](#dt-must) satisfy to be schema-valid.  Largely to be found in [Datatype components (§4)](#datatype-components).

**[XDM representation](#dt-xdmrep)**
  For any value *V* and any datatype *T*, the **XDM representation of *V* under *T*** is defined recursively as follows. Call the XDM representation *X*. Then

- 1 If *T* = [`xs:anySimpleType`](#dt-anySimpleType) or [`xs:anyAtomicType`](#dt-anyAtomicType) then *X* is *V*, and the [dynamic type](https://www.w3.org/TR/xpath20/#dt-dynamic-type) of *X* is `xs:untypedAtomic`.
- 2 If *T* . [{variety}](#std-variety) = ***atomic***, then let *T2* be the [nearest built-in datatype](#dt-optype) to *T*. If *V* is a member of the [value space](#dt-value-space) of *T2*, then *X* is *V* and the [dynamic type](https://www.w3.org/TR/xpath20/#dt-dynamic-type) of *X* is *T2*. Otherwise (i.e. if *V* is not a member of the [value space](#dt-value-space) of *T2*), *X* is the [XDM representation](#dt-xdmrep) of *V* under *T2* . [{base type definition}](#std-base_type_definition).
- 3 If *T* . [{variety}](#std-variety) = ***list***, then *X* is a sequence of atomic values, each atomic value being the [XDM representation](#dt-xdmrep) of the corresponding item in the list *V* under *T* . [{item type definition}](#std-item_type_definition).
- 4 If *T* . [{variety}](#std-variety) = ***union***, then *X* is the [XDM representation](#dt-xdmrep) of *V* under the [active basic member](#dt-active-basic-member) of *V* when validated against *T*. If there is no [active basic member](#dt-active-basic-member), then *V* has no [XDM representation](#dt-xdmrep) under *T*.

**[absent](#key-null)**
  Throughout this specification, the value *****absent***** is used as a distinguished value to indicate that a given instance of a property "has no value" or "is absent".

**[active basic member](#dt-active-basic-member)**
  If the [active member type](#dt-active-member) is itself a [union](#dt-union), one of *its* members will be *its* [active member type](#dt-active-member), and so on, until finally a [basic (non-union) member](#dt-basicmember) is reached. That [basic member](#dt-basicmember) is the **active basic member** of the union.

**[active member type](#dt-active-member)**
  In a valid instance of any [union](#dt-union), the first of its members in order which accepts the instance as valid is the **active member type**.

**[ancestor](#std-ancestor)**
  The **ancestors** of a [type definition](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html#td) are its [{base type definition}](#std-base_type_definition) and the [ancestors](#std-ancestor) of its [{base type definition}](#std-base_type_definition).

**[anyAtomicType](#dt-anyAtomicType)**
   **anyAtomicType** is a special [restriction](#dt-restriction) of [anySimpleType](#anySimpleType). The [value](#dt-value-space) and [lexical spaces](#dt-lexical-space) of **anyAtomicType** are the unions of the [value](#dt-value-space) and [lexical spaces](#dt-lexical-space) of all the [primitive](#dt-primitive) datatypes, and **anyAtomicType** is their [base type](#dt-basetype).

**[anySimpleType](#dt-anySimpleType)**
   The definition of **anySimpleType** is a special [restriction](#dt-restriction) of ***anyType***.  The [lexical space](#dt-lexical-space) of **anySimpleType** is the set of all sequences of Unicode characters, and its [value space](#dt-value-space) includes all [atomic values](#dt-atomic-value) and all finite-length lists of zero or more [atomic values](#dt-atomic-value).

**[atomic](#dt-atomic)**
  **Atomic** datatypes are those whose [value spaces](#dt-value-space) contain only [atomic values](#dt-atomic-value).  **Atomic** datatypes are [anyAtomicType](#anyAtomicType) and all datatypes [derived](#dt-derived) from it.

**[atomic value](#dt-atomic-value)**
  An **atomic value** is an elementary value, not constructed from simpler values by any user-accessible means defined by this specification.

**[base type](#dt-basetype)**
  Every datatype other than [anySimpleType](#anySimpleType) is associated with another datatype, its **base type**. **Base types** can be [special](#dt-special), [primitive](#dt-primitive), or [ordinary](#dt-ordinary).

**[basic member](#dt-basicmember)**
  Those members of the [transitive membership](#dt-transitivemembership) of a [union](#dt-union) datatype *U* which are themselves not [union](#dt-union) datatypes are the **basic members** of *U*.

**[built-in](#dt-built-in)**
  **Built-in** datatypes are those which are defined in this specification; they can be [special](#dt-special), [primitive](#dt-primitive), or [ordinary](#dt-ordinary) datatypes .

**[canonical mapping](#dt-canonical-mapping)**
  The **canonical mapping** is a prescribed subset of the inverse of a [lexical mapping](#dt-lexical-mapping) which is one-to-one and whose domain (where possible) is the entire range of the [lexical mapping](#dt-lexical-mapping) (the [value space](#dt-value-space)).

**[canonical representation](#dt-canonical-representation)**
  The **canonical representation** of a value in the [value space](#dt-value-space) of a datatype is the [lexical representation](#dt-lexical-representation) associated with that value by the datatype's [canonical mapping](#dt-canonical-mapping)

**[character class subtraction](#dt-ccsub)**
  A [character group](#dt-chargroup) that contains a subtraction operator is referred to as a **character class subtraction**.

**[character group part](#dt-cgpart)**
  A **character group part** ([*charGroupPart*](#nt-charGroupPart)) is any of: a single unescaped character ([*SingleCharNoEsc*](#nt-SingleCharNoEsc)), a single escaped character ([*SingleCharEsc*](#nt-SingleCharEsc)), a character class escape ([*charClassEsc*](#nt-charClassEsc)), or a character range ([*charRange*](#nt-charRange)).

**[constraining facet](#dt-constraining-facet)**
  **Constraining facets** are schema components whose values may be set or changed during [derivation](#dt-derived) (subject to facet-specific controls) to control various aspects of the derived datatype.

**[constructed](#dt-constructed)**
  All [ordinary](#dt-ordinary) datatypes are defined in terms of, or **constructed** from, other datatypes, either by [restricting](#dt-fb-restriction) the [value space](#dt-value-space) or [lexical space](#dt-lexical-space) of a [base type](#dt-basetype) using zero or more [constraining facets](#dt-constraining-facet) or by specifying the new datatype as a [list](#dt-list) of items of some [item type](#dt-itemType), or by defining it as a [union](#dt-union) of some specified sequence of [member types](#dt-memberTypes).

**[datatype](#dt-datatype)**
  In this specification, a **datatype** has three properties:

- A [value space](#dt-value-space), which is a set of values.
- A [lexical space](#dt-lexical-space), which is a set of [literals](#dt-literal) used to denote the values.
- A small collection of *functions, relations, and procedures* associated with the datatype.  Included are equality and (for some datatypes) order relations on the [value space](#dt-value-space), and a [lexical mapping](#dt-lexical-mapping), which is a mapping from the [lexical space](#dt-lexical-space) into the [value space](#dt-value-space).

**[derived](#dt-immediately-derived)**
  A datatype *T* is **immediately derived** from another datatype *X* if and only if *X* is the [base type](#dt-basetype) of *T*.

**[derived](#dt-derived)**
  A datatype *R* is **derived** from another datatype *B* if and only if one of the following is true:

- *B* is the [base type](#dt-basetype) of *R*.
- There is some datatype *X* such that *X* is the [base type](#dt-basetype) of *R*, and *X* is derived from *B*.

**[div](#dt-div)**
  If *m* and *n* are numbers, then *m* **div** *n* is the greatest integer less than or equal to *m* / *n* .

**[error](#dt-error)**
  A failure of a schema or schema document to conform to the rules of this specification. Except as otherwise specified, processors **must** distinguish error-free (conforming) schemas and schema documents from those with errors; if a schema used in type-validation or a schema document used in constructing a schema is in error, processors **must** report the fact; if more than one is in error, it is [implementation-dependent](#key-impl-dep) whether more than one is reported as being in error. If more than one of the constraints given in this specification is violated, it is [implementation-dependent](#key-impl-dep) how many of the violations, and which, are reported. **Note:** Failure of an XML element or attribute to be datatype-valid against a particular datatype in a particular schema is not in itself a failure to conform to this specification and thus, for purposes of this specification, not an error.

**[facet-based restriction](#dt-fb-restriction)**
  A datatype is defined by **facet-based restriction** of another datatype (its [base type](#dt-basetype)), when values for zero or more [constraining facets](#dt-constraining-facet) are specified that serve to constrain its [value space](#dt-value-space) and/or its [lexical space](#dt-lexical-space) to a subset of those of the [base type](#dt-basetype).

**[for compatibility](#dt-compatibility)**
   A feature of this specification included solely to ensure that schemas which use this feature remain compatible with [[XML]](#XML).

**[fundamental facet](#dt-fundamental-facet)**
   Each **fundamental facet** is a schema component that provides a limited piece of information about some aspect of each datatype.

**[implementation-defined](#key-impl-def)**
  Something which **may** vary among conforming implementations, but which **must** be specified by the implementor for each particular implementation, is **implementation-defined**.

**[implementation-dependent](#key-impl-dep)**
  Something which **may** vary among conforming implementations, is not specified by this or any W3C specification, and is not required to be specified by the implementor for any particular implementation, is **implementation-dependent**.

**[incomparable](#dt-incomparable)**
  Two values that are neither equal, less-than, nor greater-than are **incomparable**. Two values that are not [incomparable](#dt-incomparable) are **comparable**.

**[intervening union](#dt-interveningunion)**
  If a datatype *M* is in the [transitive membership](#dt-transitivemembership) of a [union](#dt-union) datatype *U*, but not one of *U*'s [member types](#dt-memberTypes), then a sequence of one or more [union](#dt-union) datatypes necessarily exists, such that the first is one of the [member types](#dt-memberTypes) of *U*, each is one of the [member types](#dt-memberTypes) of its predecessor in the sequence, and *M* is one of the [member types](#dt-memberTypes) of the last in the sequence. The [union](#dt-union) datatypes in this sequence are said to **intervene** between *M* and *U*. When *U* and *M* are given by the context, the datatypes in the sequence are referred to as the **intervening unions**. When *M* is one of the [member types](#dt-memberTypes) of *U*, the set of **intervening unions** is the empty set.

**[item type](#dt-itemType)**
   The [atomic](#dt-atomic) or [union](#dt-union) datatype that participates in the definition of a [list](#dt-list) datatype is the **item type** of that [list](#dt-list) datatype.

**[leap-second](#dt-leapsec)**
  A **leap-second** is an additional second added to the last day of December, June, October, or March, when such an adjustment is deemed necessary by the International Earth Rotation and Reference Systems Service in order to keep [UTC](#dt-utc) within 0.9 seconds of observed astronomical time.  When leap seconds are introduced, the last minute in the day has more than sixty seconds.  In theory leap seconds can also be removed from a day, but this has not yet occurred. (See [[International Earth Rotation Service (IERS)]](#IERS), [[ITU-R TF.460-6]](#itu-r-460-6).) Leap seconds are *not* supported by the types defined here.

**[lexical](#dt-lexical)**
  A constraining facet which directly restricts the [lexical space](#dt-lexical-space) of a datatype is a **lexical** facet.

**[lexical mapping](#dt-lexical-mapping)**
  The **lexical mapping** for a datatype is a prescribed relation which maps from the [lexical space](#dt-lexical-space) of the datatype into its [value space](#dt-value-space).

**[lexical representation](#dt-lexical-representation)**
  The members of the [lexical space](#dt-lexical-space) are **lexical representations** of the values to which they are mapped.

**[lexical space](#dt-lexical-space)**
  The **lexical space** of a datatype is the prescribed set of strings which [the lexical mapping](#dt-lexical-mapping) for that datatype maps to values of that datatype.

**[list](#dt-list)**
  **List** datatypes are those having values each of which consists of a finite-length (possibly empty) sequence of [atomic values](#dt-atomic-value). The values in a list are drawn from some [atomic](#dt-atomic) datatype (or from a [union](#dt-union) of [atomic](#dt-atomic) datatypes), which is the [item type](#dt-itemType) of the **list**.

**[literal](#dt-literal)**
  A sequence of zero or more characters in the Universal Character Set (UCS) which may or may not prove upon inspection to be a member of the [lexical space](#dt-lexical-space) of a given datatype and thus a [lexical representation](#dt-lexical-representation) of a given value in that datatype's [value space](#dt-value-space), is referred to as a **literal**.

**[match](#dt-match)**
   *(Of strings or names:)* Two strings or names being compared must be identical. Characters with multiple possible representations in ISO/IEC 10646 (e.g. characters with both precomposed and base+diacritic forms) match only if they have the same representation in both strings. No case folding is performed. *(Of strings and rules in the grammar:)* A string matches a grammatical production if and only if it belongs to the language generated by that production.

**[may](#dt-may)**
   Schemas, schema documents, and processors are permitted to but need not behave as described.

**[member types](#dt-memberTypes)**
   The datatypes that participate in the definition of a [union](#dt-union) datatype are known as the **member types** of that [union](#dt-union) datatype.

**[minimally conforming](#dt-minimally-conforming)**
  Implementations claiming **minimal conformance** to this specification independent of any host language **must** do **all** of the following:

- 1<a id="gl-support-all-primitives"></a>Support all the [built-in](#dt-built-in) datatypes defined in this specification.
- 2<a id="gl-implement-all-cos"></a>Completely and correctly implement all of the [constraints on schemas](#dt-cos) defined in this specification.
- 3<a id="gl-implement-all-vr"></a>Completely and correctly implement all of the [Validation Rules](#dt-cvc) defined in this specification, when checking the datatype validity of literals against datatypes.

**[mod](#dt-mod)**
  If *m* and *n* are numbers, then *m* **mod** *n* is  *m* − *n* × ( *m* [div](#dt-div) *n*) .

**[must](#dt-must)**
   *(Of schemas and schema documents:)* Schemas and documents are required to behave as described; otherwise they are in [error](#dt-error). *(Of processors:)* Processors are required to behave as described.

**[must not](#dt-mustnot)**
  Schemas, schema documents and processors are forbidden to behave as described; schemas and documents which nevertheless do so are in [error](#dt-error).

**[nearest built-in datatype](#dt-optype)**
  For any datatype *T*, the **nearest built-in datatype** to *T* is the first [built-in](#dt-built-in) datatype encountered in following the chain of links connecting each datatype to its [base type](#dt-basetype). If *T* is a [built-in](#dt-built-in) datatype, then the nearest built-in datatype of *T* is *T* itself; otherwise, it is the nearest built-in datatype of *T*'s [base type](#dt-basetype).

**[normalized block name](#dt-normalized-block-name)**
   For any Unicode block, the **normalized block name** of that block is the string of characters formed by stripping out white space and underbar characters from the block name as given in [[Unicode Database]](#UnicodeDB), while retaining hyphens and preserving case distinctions.

**[optional](#dt-optional)**
  An **optional** property is *permitted* but not *required* to have the distinguished value ***absent***.

**[ordered](#dt-ordered)**
  A [value space](#dt-value-space), and hence a datatype, is said to be **ordered** if some members of the [value space](#dt-value-space) are drawn from a [primitive](#dt-primitive) datatype for which the table in [Fundamental Facets (§F.1)](#app-fundamental-facets) specifies the value ***total*** or ***partial*** for the *ordered* facet.

**[ordinary](#dt-ordinary)**
  **Ordinary** datatypes are all datatypes other than the [special](#dt-special) and [primitive](#dt-primitive) datatypes.

**[owner](#dt-owner)**
  A component may be referred to as the **owner** of its properties, and of the values of those properties.

**[pre-lexical](#dt-pre-lexical)**
  A constraining facet which is used to normalize an initial [literal](#dt-literal) before checking to see whether the resulting character sequence is a member of a datatype's [lexical space](#dt-lexical-space) is a **pre-lexical** facet.

**[primitive](#dt-primitive)**
  **Primitive** datatypes are those datatypes that are not [special](#dt-special) and are not defined in terms of other datatypes; they exist *ab initio*.

**[regular expression](#dt-regex)**
  A **regular expression** is composed from zero or more [branches](#dt-branch), separated by '`|`' characters.

**[restriction](#dt-restriction)**
  A datatype *R* is a **restriction** of another datatype *B* when

**[should](#dt-should)**
  It is recommended that schemas, schema documents, and processors behave as described, but there can be valid reasons for them not to; it is important that the full implications be understood and carefully weighed before adopting behavior at variance with the recommendation.

**[special](#dt-special)**
  The **special** datatypes are [anySimpleType](#anySimpleType) and [anyAtomicType](#anyAtomicType).

**[special value](#dt-specialvalue)**
  A **special value** is an object whose only relevant properties for purposes of this specification are that it is distinct from, and unequal to, any other values (special or otherwise).

**[transitive membership](#dt-transitivemembership)**
  The **transitive membership** of a [union](#dt-union) is the set of its own [member types](#dt-memberTypes), and the [member types](#dt-memberTypes) of its members, and so on. More formally, if *U* is a [union](#dt-union), then (a) its [member types](#dt-memberTypes) are in the transitive membership of *U*, and (b) for any datatypes *T1* and *T2*, if *T1* is in the transitive membership of *U* and *T2* is one of the [member types](#dt-memberTypes) of *T1*, then *T2* is also in the transitive membership of *U*.

**[union](#dt-union)**
  **Union** datatypes are (a) those whose [value spaces](#dt-value-space), [lexical spaces](#dt-lexical-space), and [lexical mappings](#dt-lexical-mapping) are the union of the [value spaces](#dt-value-space), [lexical spaces](#dt-lexical-space), and [lexical mappings](#dt-lexical-mapping) of one or more other datatypes, which are the [member types](#dt-memberTypes) of the union, or (b) those derived by [facet-based restriction](#dt-fb-restriction) of another union datatype.

**[unknown](#dt-unknown-dt)**
  A datatype which is not available for use is said to be **unknown**.

**[unknown](#dt-unknown-f)**
  An [constraining facet](#dt-constraining-facet) which is not supported by the processor in use is **unknown**.

**[user option](#dt-useroption)**
  A choice left under the control of the user of a processor, rather than being fixed for all users or uses of the processor. Statements in this specification that "Processors **may** at user option" behave in a certain way mean that processors **may** provide mechanisms to allow users (i.e. invokers of the processor) to enable or disable the behavior indicated. Processors which do not provide such user-operable controls **must not** behave in the way indicated. Processors which do provide such user-operable controls **must** make it possible for the user to disable the optional behavior. **Note:** The normal expectation is that the default setting for such options will be to disable the optional behavior in question, enabling it only when the user explicitly requests it. This is not, however, a requirement of conformance: if the processor's documentation makes clear that the user can disable the optional behavior, then invoking the processor without requesting that it be disabled can be taken as equivalent to a request that it be enabled. It is required, however, that it in fact be possible for the user to disable the optional behavior. **Note:** Nothing in this specification constrains the manner in which processors allow users to control user options. Command-line options, menu choices in a graphical user interface, environment variables, alternative call patterns in an application programming interface, and other mechanisms may all be taken as providing user options.

**[user-defined](#dt-user-defined)**
   **User-defined** datatypes are those datatypes that are defined by individual schema designers.

**[value space](#dt-value-space)**
  The **value space** *of a datatype* is the set of values for that datatype.

**[value-based](#dt-value-based)**
  A constraining facet which directly restricts the [value space](#dt-value-space) of a datatype is a **value-based** facet.

**[wildcard character](#dt-wcchar)**
   The **wildcard character** is a metacharacter which matches almost any single character:

## <a id="biblio"></a>K References

### <a id="normative-biblio"></a>K.1 Normative

**<a id="ieee754-2008"></a>IEEE 754-2008**
   IEEE. *IEEE Standard for Floating-Point Arithmetic*. 29 August 2008. [http://ieeexplore.ieee.org/servlet/opac?punumber=4610933](http://ieeexplore.ieee.org/servlet/opac?punumber=4610933)

**<a id="XMLNS"></a>Namespaces in XML**
   World Wide Web Consortium.  *Namespaces in XML 1.1 (Second Edition)*, ed. Tim Bray et al. W3C Recommendation 16 August 2006. Available at: [http://www.w3.org/TR/xml-names11/](https://www.w3.org/TR/xml-names11/)  The edition cited is the one current at the date of publication of this specification. Implementations **may** follow the edition cited and/or any later edition(s); it is implementation-defined which. For details of the dependency of this specification on Namespaces in XML 1.1, see [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

**<a id="XMLNS1.0"></a>Namespaces in XML 1.0**
   World Wide Web Consortium.  *Namespaces in XML 1.0 (Third Edition)*, ed. Tim Bray et al. W3C Recommendation 8 December 2009. Available at: [http://www.w3.org/TR/xml-names/](https://www.w3.org/TR/xml-names/)  The edition cited is the one current at the date of publication of this specification. Implementations **may** follow the edition cited and/or any later edition(s); it is implementation-defined which. For details of the dependency of this specification on Namespaces in XML 1.0, see [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

**<a id="RFC3548"></a>RFC 3548**
   S. Josefsson, ed. *RFC 3548: The Base16, Base32, and Base64 Data Encodings*. July 2003.  Available at: [http://www.ietf.org/rfc/rfc3548.txt](http://www.ietf.org/rfc/rfc3548.txt)

**<a id="UnicodeDB"></a>Unicode Database**
   The Unicode Consortium. *Unicode Character Database*. Revision 3.1.0, ed. Mark Davis and Ken Whistler. 2001-02-28. Available at: [http://www.unicode.org/Public/3.1-Update/UnicodeCharacterDatabase-3.1.0.html](http://www.unicode.org/Public/3.1-Update/UnicodeCharacterDatabase-3.1.0.html). For later versions, see [http://www.unicode.org/versions/](http://www.unicode.org/versions/). The edition cited is the one current at the date of publication of XSD 1.0. Implementations **may** follow the edition cited and/or any later edition(s); it is implementation-defined which.

**<a id="XDM"></a>XDM**
   World Wide Web Consortium. *XQuery 1.0 and XPath 2.0 Data Model (XDM) (Second Edition)*, ed. Mary Fernández et al. W3C Recommendation 14 December 2010. Available at: [http://www.w3.org/TR/xpath-datamodel/](https://www.w3.org/TR/xpath-datamodel/).

**<a id="XML"></a>XML**
   World Wide Web Consortium. *Extensible Markup Language (XML) 1.1 (Second Edition)*, ed. Tim Bray et al. W3C Recommendation 16 August 2006, edited in place 29 September 2006. Available at [http://www.w3.org/TR/xml11/](https://www.w3.org/TR/xml11/)  The edition cited is the one current at the date of publication of this specification. Implementations **may** follow the edition cited and/or any later edition(s); it is implementation-defined which. For details of the dependency of this specification on XML 1.1, see [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

**<a id="XML1.0"></a>XML 1.0**
   World Wide Web Consortium. *Extensible Markup Language (XML) 1.0 (Fifth Edition)*, ed. Tim Bray et al. W3C Recommendation 26 November 2008. Available at [http://www.w3.org/TR/xml/](https://www.w3.org/TR/xml/).  The edition cited is the one current at the date of publication of this specification. Implementations **may** follow the edition cited and/or any later edition(s); it is implementation-defined which. For details of the dependency of this specification on XML, see [Dependencies on Other Specifications (§1.3)](#intro-relatedWork).

**<a id="XPATH2"></a>XPath 2.0**
   World Wide Web Consortium. *XML Path Language (XPath) 2.0 (Second Edition)*, ed. Anders Berglund et al. W3C Recommendation 14 December 2010 *(Link errors corrected 3 January 2011)*. Available at: [http://www.w3.org/TR/xpath20/](https://www.w3.org/TR/xpath20/).

**<a id="F_O"></a>XQuery 1.0 and XPath 2.0 Functions and Operators**
   World Wide Web Consortium. *XQuery 1.0 and XPath 2.0 Functions and Operators (Second Edition)*, ed. Ashok Malhotra et al. W3C Recommendation 14 December 2010. Available at: [http://www.w3.org/TR/xpath-functions/](https://www.w3.org/TR/xpath-functions/).

**<a id="structural-schemas"></a>XSD 1.1 Part 1: Structures**
   World Wide Web Consortium. *W3C XML Schema Definition Language (XSD) 1.1 Part 1: Structures*, ed. Shudi (Sandy) Gao 高殊镝, C. M. Sperberg-McQueen, and Henry S. Thompson. W3C Recommendation 5 April 2012. Available at: [http://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html](https://www.w3.org/TR/2012/REC-xmlschema11-1-20120405/structures.html)  The edition cited is the one current at the date of publication of this specification. Implementations **may** follow the edition cited and/or any later edition(s); it is implementation-defined which.

### <a id="non-normative-biblio"></a>K.2 Non-normative

**<a id="BCP47"></a>BCP 47**
   Internet Engineering Task Force (IETF). Best Current Practices 47. 2006. Available at: [http://tools.ietf.org/rfc/bcp/bcp47](http://tools.ietf.org/rfc/bcp/bcp47). Concatenation of *RFC 4646: Tags for Identifying Languages*, ed. A. Phillips and M. Davis, September 2006, [http://www.ietf.org/rfc/bcp/bcp47.txt](http://www.ietf.org/rfc/bcp/bcp47.txt), and *RFC 4647: Matching of Language Tags*, ed. A Phillips and M. Davis, September 2006, [http://www.rfc-editor.org/rfc/bcp/bcp47.txt](http://www.rfc-editor.org/rfc/bcp/bcp47.txt).

**<a id="clinger1990"></a>Clinger, WD (1990)**
   William D Clinger. *How to Read Floating Point Numbers Accurately.* In *Proceedings of Conference on Programming Language Design and Implementation*, pages 92-101. Available at: [ftp://ftp.ccs.neu.edu/pub/people/will/howtoread.ps](ftp://ftp.ccs.neu.edu/pub/people/will/howtoread.ps)

**<a id="html4"></a>HTML 4.01**
   World Wide Web Consortium.  *HTML 4.01 Specification*, ed. Dave Raggett, Arnaud Le Hors, and Ian Jacobs. W3C Recommendation 24 December 1999. Available at: [http://www.w3.org/TR/html401/](https://www.w3.org/TR/html401/)

**<a id="ISO11404"></a>ISO 11404**
   ISO (International Organization for Standardization). *Language-independent Datatypes.* ISO/IEC 11404:2007. See [http://www.iso.org/iso/iso_catalogue/catalogue_tc/catalogue_detail.htm?csnumber=39479](http://www.iso.org/iso/iso_catalogue/catalogue_tc/catalogue_detail.htm?csnumber=39479)

**<a id="ISO8601"></a>ISO 8601**
   ISO (International Organization for Standardization). *Representations of dates and times, 1988-06-15.*

**<a id="ISO8601-2000"></a>ISO 8601:2000 Second Edition**
   ISO (International Organization for Standardization). *Representations of dates and times, second edition, 2000-12-15.*

**<a id="itu-r-460-6"></a>ITU-R TF.460-6**
   International Telecommunication Union (ITU). *Recommendation ITU-R TF.460-6: Standard-frequency and time-signal emissions*. [Geneva: ITU, February 2002.]

**<a id="IERS"></a>International Earth Rotation Service (IERS)**
   International Earth Rotation Service (IERS). See [http://maia.usno.navy.mil](http://maia.usno.navy.mil)

**<a id="LEIRIs"></a>LEIRI**
   *Legacy extended IRIs for XML resource identification*, ed. Henry S. Thompson, Richard Tobin, and Norman Walsh. W3C Working Group Note 3 November 2008 (BNF comment style corrected in place 2009-07-09). See [http://www.w3.org/TR/leiri/](https://www.w3.org/TR/leiri/)

**<a id="Perl"></a>Perl**
   The Perl Programming Language.  See [http://www.perl.org/get.html](http://www.perl.org/get.html)

**<a id="pd-note"></a>Precision Decimal**
   World Wide Web Consortium. *An XSD datatype for IEEE floating-point decimal*, ed. David Peterson and C. M. Sperberg-McQueen. W3C Working Group Note 9 June 2011. Available at [http://www.w3.org/TR/xsd-precisionDecimal/](https://www.w3.org/TR/xsd-precisionDecimal/)

**<a id="RDFSchema"></a>RDF Schema**
   World Wide Web Consortium. *RDF Vocabulary Description Language 1.0: RDF Schema*, ed. Dan Brickley and R. V. Guha. W3C Recommendation 10 February 2004. Available at: [http://www.w3.org/TR/rdf-schema/](https://www.w3.org/TR/rdf-schema/)

**<a id="RFC2045"></a>RFC 2045**
   N. Freed and N. Borenstein. *RFC 2045: Multipurpose Internet Mail Extensions (MIME) Part One: Format of Internet Message Bodies*. 1996.  Available at: [http://www.ietf.org/rfc/rfc2045.txt](http://www.ietf.org/rfc/rfc2045.txt)

**<a id="RFC3066"></a>RFC 3066**
   H. Alvestrand, ed. *RFC 3066: Tags for the Identification of Languages* 1995. Available at: [http://www.ietf.org/rfc/rfc3066.txt](http://www.ietf.org/rfc/rfc3066.txt)

**<a id="RFC3986"></a>RFC 3986**
   T. Berners-Lee, R. Fielding, and L. Masinter, *RFC 3986: Uniform Resource Identifier (URI): Generic Syntax*. January 2005.  Available at: [http://www.ietf.org/rfc/rfc3986.txt](http://www.ietf.org/rfc/rfc3986.txt)

**<a id="RFC3987"></a>RFC 3987**
   M. Duerst and M. Suignard. *RFC 3987: Internationalized Resource Identifiers (IRIs) *. January 2005.  Available at: [http://www.ietf.org/rfc/rfc3987.txt](http://www.ietf.org/rfc/rfc3987.txt)

**<a id="RFC4646"></a>RFC 4646**
   A. Phillips and M. Davis, ed. *RFC 4646: Tags for Identifying Languages* 2006. Available at: [http://www.ietf.org/rfc/rfc4646.txt](http://www.ietf.org/rfc/rfc4646.txt)

**<a id="RFC4647"></a>RFC 4647**
   A. Phillips and M. Davis, ed. *RFC 4647: Matching of Language Tags* 2006. Available at: [http://www.ietf.org/rfc/rfc4647.txt](http://www.ietf.org/rfc/rfc4647.txt)

**<a id="ruby"></a>Ruby**
   World Wide Web Consortium. *Ruby Annotation*, ed. Marcin Sawicki et al. W3C Recommendation 31 May 2001 (Markup errors corrected 25 June 2008). Available at: [http://www.w3.org/TR/ruby/](https://www.w3.org/TR/ruby/)

**<a id="SQL"></a>SQL**
   ISO (International Organization for Standardization).  *ISO/IEC 9075-2:1999, Information technology --- Database languages --- SQL --- Part 2: Foundation (SQL/Foundation)*. [Geneva]: International Organization for Standardization, 1999. See [http://www.iso.org/iso/home.htm](http://www.iso.org/iso/home.htm)

**<a id="ref-timezones"></a>Timezones**
   World Wide Web Consortium. *Working with Time Zones*, ed. Addison Phillips et al. W3C Working Group Note 5 July 2011. Available at [http://www.w3.org/TR/timezone/](https://www.w3.org/TR/timezone/)

**<a id="USNavy"></a>U.S. Naval Observatory Time Service Department**
   *Information about Leap Seconds* Available at: [http://tycho.usno.navy.mil/leapsec.html](http://tycho.usno.navy.mil/leapsec.html)

**<a id="USNavy_leaps"></a>USNO Historical List**
   U.S. Naval Observatory Time Service Department, *Historical list of leap seconds* Available at: [ftp://maia.usno.navy.mil/ser7/tai-utc.dat](ftp://maia.usno.navy.mil/ser7/tai-utc.dat)

**<a id="unicodeRegEx"></a>Unicode Regular Expression Guidelines**
   Mark Davis.  *Unicode Regular Expression Guidelines*, 1988. Available at: [http://www.unicode.org/unicode/reports/tr18/](http://www.unicode.org/reports/tr18/)

**<a id="unicode-escapes"></a>Unicode block names**
   World Wide Web Consortium. *Unicode block names for use in XSD regular expressions*, ed. C. M. Sperberg-McQueen. W3C Working Group Note 9 June 2011. Available at: [http://www.w3.org/TR/xsd-unicode-blocknames/](https://www.w3.org/TR/xsd-unicode-blocknames/)

**<a id="schema-primer"></a>XML Schema Language: Part 0 Primer**
   World Wide Web Consortium. XML Schema Language: Part 0 Primer Second Edition, ed. David C. Fallside and Priscilla Walmsley. W3C Recommendation 28 October 2004. Available at: [http://www.w3.org/TR/xmlschema-0/](https://www.w3.org/TR/xmlschema-0/)

**<a id="schema-requirements"></a>XML Schema Requirements**
   *XML Schema Requirements *, ed. Ashok Malhotra and Murray Maloney. W3C Note 15 February 1999. Available at: [http://www.w3.org/TR/NOTE-xml-schema-req](https://www.w3.org/TR/NOTE-xml-schema-req)

**<a id="XSL"></a>XSL**
   World Wide Web Consortium.  *Extensible Stylesheet Language (XSL)*, ed. Anders Berglund. W3C Recommendation 05 December 2006. Available at:  [http://www.w3.org/TR/xsl11/](https://www.w3.org/TR/xsl11/)

## <a id="acknowledgments"></a>L Acknowledgements (non-normative)

Along with the editors thereof, the following contributed material to the first version of this specification:

> Asir S. Vedamuthu, webMethods, Inc
> Mark Davis, IBM

Co-editor Ashok Malhotra's work on this specification from March 1999 until February 2001 was supported by IBM, and from then until May 2004 by Microsoft.  Since July 2004 his work on this specification has been supported by Oracle Corporation.

The work of Dave Peterson as a co-editor of this specification was supported by IDEAlliance (formerly GCA) through March 2004, and beginning in April 2004 by SGML*Works!*.

The work of C. M. Sperberg-McQueen as a co-editor of this specification was supported by the World Wide Web Consortium through January 2009 and again from June 2010 through May 2011, and beginning in February 2009 by Black Mesa Technologies LLC.

The XML Schema Working Group acknowledges with thanks the members of other W3C Working Groups and industry experts in other forums who have contributed directly or indirectly to the creation of this document and its predecessor.

At the time this document is published, the members in good standing of the XML Schema Working Group are:

- David Ezell, National Association of Convenience Stores (NACS) (*chair*)
- Shudi (Sandy) Gao 高殊镝, IBM
- Mary Holstege, Mark Logic
- Sam Idicula, Oracle Corporation
- Michael Kay, Invited expert
- Jim Melton, Oracle Corporation
- Dave Peterson, Invited expert
- Liam Quin, W3C (*staff contact*)
- C. M. Sperberg-McQueen, invited expert
- Henry S. Thompson, University of Edinburgh
- Kongyi Zhou, Oracle Corporation

The XML Schema Working Group has benefited in its work from the participation and contributions of a number of people who are no longer members of the Working Group in good standing at the time of publication of this Working Draft. Their names are given below. In particular we note with sadness the accidental death of Mario Jeckle shortly before publication of the first Working Draft of XML Schema 1.1. Affiliations given are (among) those current at the time of the individuals' work with the WG.

- Paula Angerstein, Vignette Corporation
- Leonid Arbouzov, Sun Microsystems
- Jim Barnette, Defense Information Systems Agency (DISA)
- David Beech, Oracle Corp.
- Gabe Beged-Dov, Rogue Wave Software
- Laila Benhlima, Ecole Mohammadia d'Ingenieurs Rabat (EMI)
- Doris Bernardini, Defense Information Systems Agency (DISA)
- Paul V. Biron, HL7; later Invited expert
- Don Box, DevelopMentor
- Allen Brown, Microsoft
- Lee Buck, TIBCO Extensibility
- Greg Bumgardner, Rogue Wave Software
- Dean Burson, Lotus Development Corporation
- Charles E. Campbell, Invited expert
- Oriol Carbo, University of Edinburgh
- Wayne Carr, Intel
- Peter Chen, Bootstrap Alliance and LSU
- Tyng-Ruey Chuang, Academia Sinica
- Tony Cincotta, NIST
- David Cleary, Progress Software
- Mike Cokus, MITRE
- Dan Connolly, W3C (*staff contact*)
- Ugo Corda, Xerox
- Roger L. Costello, MITRE
- Joey Coyle, Health Level Seven
- Haavard Danielson, Progress Software
- Josef Dietl, Mozquito Technologies
- Kenneth Dolson, Defense Information Systems Agency (DISA)
- Andrew Eisenberg, Progress Software
- Rob Ellman, Calico Commerce
- Tim Ewald, Developmentor
- Alexander Falk, Altova GmbH
- David Fallside, IBM
- George Feinberg, Object Design
- Dan Fox, Defense Logistics Information Service (DLIS)
- Charles Frankston, Microsoft
- Matthew Fuchs, Commerce One
- Andrew Goodchild, Distributed Systems Technology Centre (DSTC Pty Ltd)
- Xan Gregg, TIBCO Extensibility
- Paul Grosso, Arbortext, Inc
- Martin Gudgin, DevelopMentor
- Ernesto Guerrieri, Inso
- Dave Hollander, Hewlett-Packard Company (*co-chair*)
- Nelson Hung, Corel
- Jane Hunter, Distributed Systems Technology Centre (DSTC Pty Ltd)
- Michael Hyman, Microsoft
- Renato Iannella, Distributed Systems Technology Centre (DSTC Pty Ltd)
- Mario Jeckle, DaimlerChrysler
- Rick Jelliffe, Academia Sinica
- Marcel Jemio, Data Interchange Standards Association
- Simon Johnston, Rational Software
- Kohsuke Kawaguchi, Sun Microsystems
- Dianne Kennedy, Graphic Communications Association
- Janet Koenig, Sun Microsystems
- Setrag Khoshafian, Technology Deployment International (TDI)
- Melanie Kudela, Uniform Code Council
- Ara Kullukian, Technology Deployment International (TDI)
- Andrew Layman, Microsoft
- Dmitry Lenkov, Hewlett-Packard Company
- Bob Lojek, Mozquito Technologies
- John McCarthy, Lawrence Berkeley National Laboratory
- Matthew MacKenzie, XML Global
- Nan Ma, China Electronics Standardization Institute
- Eve Maler, Sun Microsystems
- Ashok Malhotra, IBM, Microsoft, Oracle
- Murray Maloney, Muzmo Communication, acting for Commerce One
- Paolo Marinelli, University of Bologna
- Lisa Martin, IBM
- Noah Mendelsohn, Lotus; IBM; invited expert
- Adrian Michel, Commerce One
- Alex Milowski, Invited expert
- Don Mullen, TIBCO Extensibility
- Murata Makoto, Xerox
- Ravi Murthy, Oracle
- Chris Olds, Wall Data
- Frank Olken, Lawrence Berkeley National Laboratory
- David Orchard, BEA Systems, Inc.
- Paul Pedersen, Mark Logic Corporation
- Shriram Revankar, Xerox
- Mark Reinhold, Sun Microsystems
- Jonathan Robie, Software AG
- Cliff Schmidt, Microsoft
- John C. Schneider, MITRE
- Eric Sedlar, Oracle Corp.
- Lew Shannon, NCR
- Anli Shundi, TIBCO Extensibility
- William Shea, Merrill Lynch
- Jerry L. Smith, Defense Information Systems Agency (DISA)
- John Stanton, Defense Information Systems Agency (DISA)
- Tony Stewart, Rivcom
- Bob Streich, Calico Commerce
- William K. Stumbo, Xerox
- Hoylen Sue, Distributed Systems Technology Centre (DSTC Pty Ltd)
- Ralph Swick, W3C
- John Tebbutt, NIST
- Ross Thompson, Contivo
- Matt Timmermans, Microstar
- Jim Trezzo, Oracle Corp.
- Steph Tryphonas, Microstar
- Scott Tsao, The Boeing Company
- Mark Tucker, Health Level Seven
- Asir S. Vedamuthu, webMethods, Inc
- Fabio Vitali, University of Bologna
- Scott Vorthmann, TIBCO Extensibility
- Priscilla Walmsley, XMLSolutions
- Norm Walsh, Sun Microsystems
- Cherry Washington, Defense Information Systems Agency (DISA)
- Aki Yoshida, SAP AG
- Stefano Zacchiroli, University of Bologna
- Mohamed Zergaoui, Innovimax
