package awk

import "os"

// readFileHelper is in its own file so the //nolint:gosec annotation
// covers a single, narrow location. -f PROGFILE is by definition a
// user-supplied path.
func readFileHelper(name string) ([]byte, error) {
	return os.ReadFile(name) //nolint:gosec // -f PROGFILE is user-supplied by design
}
