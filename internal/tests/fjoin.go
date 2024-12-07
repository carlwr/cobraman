package tests

import (
	"path/filepath"
	"strings"

	"github.com/flytam/filenamify"
)

// Like `filepath.Join()`, but additionally filenamifies each individual path component.
func FilenamifyJoin(parts ...string) (string, error) {

	opts := filenamify.Options{Replacement: "_"}

	var fixeds []string
	isAbs := filepath.IsAbs(parts[0])

	for _, part := range parts {
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
			fixeds = append(fixeds, fixed)
		}
	}

	joined := filepath.Join(fixeds...)
	if isAbs {
		joined = string(filepath.Separator) + joined
	}
	return joined, nil
}
