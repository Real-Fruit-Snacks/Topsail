//go:build !windows

package df

import "golang.org/x/sys/unix"

func statFS(path string) (fsStats, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return fsStats{}, err
	}
	return fsStats{
		totalBytes: uint64(st.Blocks) * uint64(st.Bsize),
		freeBytes:  uint64(st.Bavail) * uint64(st.Bsize),
	}, nil
}
