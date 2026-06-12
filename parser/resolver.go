package parser

// SchemaResolver is the seam between schemaLocation hints and document
// bytes. The loader resolves every import/include/redefine/override through
// it, so callers can serve documents from catalogs, archives, or test maps.

import (
	"io"
	"net/url"
	"os"
	"path/filepath"
)

// SchemaResolver locates the schema document named by a schemaLocation.
// location is the (possibly relative) value written in the composition
// element; base is the URI of the document containing the reference.
type SchemaResolver interface {
	Resolve(location, base string) (io.ReadCloser, error)
}

// FileResolver resolves schema locations on the local filesystem, relative
// to the referencing document's directory. It is the default resolver.
type FileResolver struct{}

func (FileResolver) Resolve(location, base string) (io.ReadCloser, error) {
	return os.Open(resolveLocation(location, base))
}

// resolveLocation joins a schemaLocation with the URI of the referencing
// document. Absolute URLs and absolute paths pass through; relative
// references resolve against the base's directory (RFC 3986 when the base
// is a URL, filepath semantics otherwise). The result is also the loader's
// deduplication key, so equal documents reached via equal references load
// once.
func resolveLocation(location, base string) string {
	if u, err := url.Parse(location); err == nil && u.IsAbs() {
		return location
	}
	if filepath.IsAbs(location) || base == "" {
		return location
	}
	if bu, err := url.Parse(base); err == nil && bu.IsAbs() {
		if ref, err := url.Parse(location); err == nil {
			return bu.ResolveReference(ref).String()
		}
	}
	return filepath.Join(filepath.Dir(base), location)
}
