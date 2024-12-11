package fjoin

import (
	"path/filepath"
	"strings"

	"github.com/flytam/filenamify"
)

// Joins any number of path elements into a single path, separating them with an OS specific [Separator]. Each path element additionally undergoes filenamification before they are joined.
//
// In other words: like `filepath.Join()` but with each element filenameified.
//
// Empty elements are ignored. The result is `filepath.Clean`ed.
func Join(parts ...string) (string, error) {

	opts := filenamify.Options{Replacement: "_"}

	var namified []string
	isAbs := filepath.IsAbs(parts[0])

	for _, part := range parts {
		if part == "" {
			continue
		}
		partCl := filepath.Clean(part)
		splitted := strings.Split(partCl, "/")
		for _, elem := range splitted {
			if elem == "" {
				continue
			}
			fixed, err := filenamify.Filenamify(elem, opts)
			if err != nil {
				return "", err
			}
			namified = append(namified, fixed)
		}
	}

	joined := filepath.Join(namified...)
	if isAbs {
		joined = string(filepath.Separator) + joined
	}
	return joined, nil
}
