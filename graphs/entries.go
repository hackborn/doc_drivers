package graphs

import (
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

func Entries() map[string]Entry {
	// Clone the map so clients can't modify the original. There
	// is experimental support for this, switch over when it's official.
	m := make(map[string]Entry)
	for k, v := range entries {
		m[k] = v
	}
	return m
}

// ReadEntries reads a collection of entries from an FS.
func ReadEntries(fsys fs.FS, glob string) (map[string]Entry, error) {
	matches, err := fs.Glob(fsys, glob)
	if err != nil {
		return nil, err
	}
	ans := make(map[string]Entry)
	for _, match := range matches {
		name := path.Base(match)
		ext := filepath.Ext(name)
		name = strings.TrimSuffix(name, ext)
		ans[name] = Entry{Graph: NewReadFileFunc(match, fsys)}
	}
	return ans, nil
}

// NewReadFileFunc answers a StringFunc based on reading
// a file from a filesystem.
func NewReadFileFunc(path string, fsys fs.FS) StringFunc {
	return func() (string, error) {
		return readFile(path, fsys)
	}
}

func readFile(path string, fsys fs.FS) (string, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contents, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

var (
	entries = make(map[string]Entry)
)
