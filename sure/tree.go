package sure

// Attributes are stored as simple key/value pairs, always as strings.
// The particular attributes are defined by the file readers.
type AttMap map[string]string

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
