package sure // import "davidb.org/x/gosure/sure"

// AttMap defines an interface to general attributes for file nodes.
type AttMap interface {
	GetKind() string

	// Compare these attributes with another node.  If the type of
	// the other node differs, report 'kind' and be done.
	// Otherwise, appends the field names of the attributes that
	// are missing.
	// Compare(other interface{}, atts []string) []string
}

// A directory tree node.
type Tree struct {
	Name     string
	Atts     AttMap
	Children []*Tree
	Files    []*File
}

// A non-directory node (not necessarily a plain file)
type File struct {
	Name string
	Atts AttMap
}
