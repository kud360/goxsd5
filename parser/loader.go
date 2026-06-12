package parser

// The discovery step: load the transitive closure of a schema's
// imports/includes/redefines/overrides through a SchemaResolver, then
// register every document's globals into one registry.
//
// Documents are parsed and pass-1-validated once per resolved URI; a
// schemaDoc *instance* exists once per (URI, effective target namespace),
// because a chameleon include (src-include.2.2) gives the same document a
// different namespace per includer. Cyclic includes/imports terminate on the
// instance cache.
//
// Redefine/override semantics (per the settled scoped-registry design):
// the replacement children register as THE global components; the originals
// they replace are suppressed from global registration (pervasively, through
// the target's own compositions) and kept aside so that a redefine child can
// resolve its mandated self-reference to the original via a per-child scoped
// registry.

import (
	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// symKey identifies a global component within one symbol space.
type symKey struct {
	s space
	q xsd.QName
}

// treeEntry caches one resolution+parse attempt per URI. resolveErr means
// the resolver produced nothing (fatal for include, silently tolerated for
// import per src-import note); parseErr means the document was retrieved but
// is not parseable XML (always an error).
type treeEntry struct {
	root       *xmltree.Node
	resolveErr error
	parseErr   error
}

// replacement is one redefine/override composition, recorded during
// discovery and wired during registration.
type replacement struct {
	kind   string // "redefine" or "override"
	node   *xmltree.Node
	owner  *schemaDoc
	target *schemaDoc
}

type docKey struct{ uri, tns string }

type loader struct {
	resolver SchemaResolver
	errs     *xsd.ErrorList

	trees     map[string]*treeEntry
	validated map[string]bool // pass 1 already reported for this URI
	docs      map[docKey]*schemaDoc
	order     []*schemaDoc
	reps      []*replacement
}

func newLoader(resolver SchemaResolver, errs *xsd.ErrorList) *loader {
	return &loader{
		resolver:  resolver,
		errs:      errs,
		trees:     map[string]*treeEntry{},
		validated: map[string]bool{},
		docs:      map[docKey]*schemaDoc{},
	}
}

// loadRoot loads a top-level schema document (an entry point, not a
// composition target). The returned error reports an unloadable root; all
// schema errors go to l.errs.
func (l *loader) loadRoot(location string) error {
	uri := resolveLocation(location, "")
	te := l.tree(uri, location, "")
	if te.resolveErr != nil {
		return te.resolveErr
	}
	if te.parseErr != nil {
		return te.parseErr
	}
	tns, _ := te.root.Attr("targetNamespace")
	l.instance(te.root, uri, tns, false)
	return nil
}

// tree resolves and parses the document at uri (the already-joined form of
// location against base) exactly once.
func (l *loader) tree(uri, location, base string) *treeEntry {
	if te, ok := l.trees[uri]; ok {
		return te
	}
	te := &treeEntry{}
	l.trees[uri] = te
	rc, err := l.resolver.Resolve(location, base)
	if err != nil {
		te.resolveErr = err
		return te
	}
	defer rc.Close()
	te.root, te.parseErr = xmltree.Parse(rc, uri)
	return te
}

// instance returns the schemaDoc for (uri, effective targetNamespace),
// creating it on first use and recursing into its compositions. chameleon
// marks a document absorbed into the includer's namespace (src-include.2.2);
// unqualified component references inside it are remapped to effTNS.
func (l *loader) instance(root *xmltree.Node, uri, effTNS string, chameleon bool) *schemaDoc {
	key := docKey{uri, effTNS}
	if doc, ok := l.docs[key]; ok {
		return doc
	}
	errs := l.errs
	if l.validated[uri] {
		errs = &xsd.ErrorList{} // second namespace instance: discard duplicate pass-1 reports
	}
	l.validated[uri] = true
	doc := loadDoc(root, uri, errs)
	l.docs[key] = doc
	if doc == nil {
		return nil
	}
	if chameleon {
		doc.chameleonNS = effTNS
		doc.targetNamespace = effTNS
		if doc.defaultAttributes.Local != "" && doc.defaultAttributes.Namespace == "" {
			doc.defaultAttributes.Namespace = effTNS
		}
	}
	doc.importedNS = map[string]bool{}
	for _, comp := range doc.compositions {
		if comp.kind == "import" {
			doc.importedNS[comp.namespace] = true
		}
	}
	l.order = append(l.order, doc)
	for _, comp := range doc.compositions {
		l.compose(doc, comp)
	}
	return doc
}

// compose loads one include/import/redefine/override target of doc.
func (l *loader) compose(doc *schemaDoc, comp composition) {
	pos := comp.node.Pos
	if comp.kind == "import" {
		// spec: src-import.1.1/.1.2 — XSD 1.1 Part 1 §4.2.6.2 (src-import)
		if comp.namespace != "" && comp.namespace == doc.targetNamespace {
			l.errs.Addf(xsd.SpecSrcImport, pos, "import namespace %s must differ from the target namespace; use include", comp.namespace)
			return
		}
		if comp.namespace == "" && doc.targetNamespace == "" {
			l.errs.Addf(xsd.SpecSrcImport, pos, "import without a namespace requires the importing schema to have a target namespace")
			return
		}
		if comp.schemaLocation == "" {
			return // namespace dependency only; satisfied by another root or not at all
		}
		uri := resolveLocation(comp.schemaLocation, doc.uri)
		te := l.tree(uri, comp.schemaLocation, doc.uri)
		if te.resolveErr != nil {
			return // an unresolvable import location is explicitly not an error
		}
		if te.parseErr != nil {
			l.errs.Addf(xsd.SpecSrcImport, pos, "imported document %s: %v", uri, te.parseErr)
			return
		}
		tns, _ := te.root.Attr("targetNamespace")
		if tns != comp.namespace {
			// spec: src-import.3 — XSD 1.1 Part 1 §4.2.6.2 (src-import)
			l.errs.Addf(xsd.SpecSrcImport, pos, "imported document %s targets namespace %q, want %q", uri, tns, comp.namespace)
			return
		}
		l.instance(te.root, uri, tns, false)
		return
	}

	// include / redefine / override
	ref := xsd.SpecSrcInclude
	switch comp.kind {
	case "redefine":
		ref = xsd.SpecSrcRedefine
	case "override":
		ref = xsd.SpecSrcOverride
	}
	if comp.schemaLocation == "" {
		return // missing required attribute, already reported by pass 1
	}
	uri := resolveLocation(comp.schemaLocation, doc.uri)
	te := l.tree(uri, comp.schemaLocation, doc.uri)
	if te.resolveErr != nil {
		l.errs.Addf(ref, pos, "cannot resolve %s document %s: %v", comp.kind, uri, te.resolveErr)
		return
	}
	if te.parseErr != nil {
		l.errs.Addf(ref, pos, "%s document %s: %v", comp.kind, uri, te.parseErr)
		return
	}
	tns, _ := te.root.Attr("targetNamespace")
	var target *schemaDoc
	switch {
	case tns == doc.targetNamespace:
		target = l.instance(te.root, uri, tns, false)
	case tns == "":
		// spec: src-include.2.2 — chameleon: the document is absorbed into
		// the including schema's target namespace.
		target = l.instance(te.root, uri, doc.targetNamespace, true)
	default:
		// spec: src-include.2.1 — XSD 1.1 Part 1 §4.2.3 (src-include)
		l.errs.Addf(ref, pos, "%s document %s targets namespace %q, want %q", comp.kind, uri, tns, doc.targetNamespace)
		return
	}
	if target == nil {
		return
	}
	doc.targets = append(doc.targets, target)
	if comp.kind != "include" {
		l.reps = append(l.reps, &replacement{kind: comp.kind, node: comp.node, owner: doc, target: target})
	}
}

// register enters every loaded document's globals into reg: suppression sets
// first (so originals of redefined/overridden components stay out of the
// global registry), then per-document registration, then the replacement
// children themselves.
func (l *loader) register(reg *registry) {
	for _, rep := range l.reps {
		for _, c := range xsdElems(rep.node, rep.owner) {
			s, ok := globalSpaces[c.Name.Local]
			if !ok {
				continue // annotation
			}
			if name, ok := c.Attr("name"); ok {
				l.suppress(rep.target, symKey{s, xsd.QName{Namespace: rep.owner.targetNamespace, Local: name}}, map[*schemaDoc]bool{})
			}
		}
	}
	for _, doc := range l.order {
		if doc != nil {
			registerDoc(reg, doc, l.errs)
		}
	}
	for _, rep := range l.reps {
		l.registerReplacement(reg, rep)
	}
}

// suppress marks k replaced in doc and, transitively, in every document doc
// composes in (the override/redefine transformation applies through the
// target's own includes and overrides).
func (l *loader) suppress(doc *schemaDoc, k symKey, seen map[*schemaDoc]bool) {
	if doc == nil || seen[doc] {
		return
	}
	seen[doc] = true
	if doc.suppressed == nil {
		doc.suppressed = map[symKey]bool{}
	}
	doc.suppressed[k] = true
	for _, t := range doc.targets {
		l.suppress(t, k, seen)
	}
}

// findOriginal locates the suppressed original declaration for k in target's
// composition closure.
func (l *loader) findOriginal(target *schemaDoc, k symKey, seen map[*schemaDoc]bool) *decl {
	if target == nil || seen[target] {
		return nil
	}
	seen[target] = true
	if d := target.originals[k]; d != nil {
		return d
	}
	for _, t := range target.targets {
		if d := l.findOriginal(t, k, seen); d != nil {
			return d
		}
	}
	return nil
}

// registerReplacement enters the children of one redefine/override into the
// global registry. Override children with no matching component in the
// overridden document are ignored per the override transformation; redefine
// children additionally must redefine something that exists and (for types)
// derive from themselves — the self-reference resolves to the suppressed
// original through a per-child scoped registry.
func (l *loader) registerReplacement(reg *registry, rep *replacement) {
	for _, c := range xsdElems(rep.node, rep.owner) {
		s, ok := globalSpaces[c.Name.Local]
		if !ok {
			continue
		}
		name, ok := c.Attr("name")
		if !ok {
			continue // missing required attribute, reported by pass 1
		}
		q := xsd.QName{Namespace: rep.owner.targetNamespace, Local: name}
		k := symKey{s, q}
		orig := l.findOriginal(rep.target, k, map[*schemaDoc]bool{})
		d := &decl{name: q, pos: c.Pos, node: c, doc: rep.owner}
		switch rep.kind {
		case "override":
			if orig == nil {
				// spec: override transformation — XSD 1.1 Part 1 §4.2.5:
				// children that match nothing in the overridden document do
				// not become components.
				continue
			}
		case "redefine":
			if orig == nil {
				// spec: src-redefine.6/.7 — XSD 1.1 Part 1 §4.2.4: each child
				// must redefine a component of the redefined document.
				l.errs.Addf(xsd.SpecSrcRedefine, c.Pos, "redefined %s %s is not declared in the redefined document", s, q)
			}
			l.checkRedefineSelfBase(c, rep.owner, q)
			// The child resolves its own name to the original; every other
			// reference (here and elsewhere) sees the redefinition.
			pdoc := *rep.owner
			pdoc.scoped = reg.scope()
			if orig != nil {
				pdoc.scoped.add(s, orig)
			}
			d.doc = &pdoc
		}
		reg.register(s, d, l.errs)
	}
}

// checkRedefineSelfBase enforces that a redefining type derives from the
// type it redefines.
// spec: src-redefine.5 — XSD 1.1 Part 1 §4.2.4 (src-redefine)
func (l *loader) checkRedefineSelfBase(c *xmltree.Node, doc *schemaDoc, q xsd.QName) {
	var base xsd.QName
	found := false
	switch c.Name.Local {
	case "simpleType":
		if r := firstChild(c, doc, "restriction"); r != nil {
			base, found = qnameAttr(r, doc, "base")
		}
	case "complexType":
		if cc := firstChild(c, doc, "simpleContent", "complexContent"); cc != nil {
			if r := firstChild(cc, doc, "restriction", "extension"); r != nil {
				base, found = qnameAttr(r, doc, "base")
			}
		}
	default:
		return // group/attributeGroup self-reference occurrence checks deferred
	}
	if !found || base != q {
		l.errs.Addf(xsd.SpecSrcRedefine, c.Pos, "redefined type %s must derive from itself", q)
	}
}
