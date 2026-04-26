package sort

import "os"

// openFile is in its own file so the //nolint:gosec annotation stays
// on a single, narrow location. User-supplied path is the whole point
// of `sort FILE`.
func openFile(name string) (*os.File, error) {
	return os.Open(name) //nolint:gosec // user-supplied path is the whole point
}
