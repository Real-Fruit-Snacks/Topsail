//go:build windows

package df

import "golang.org/x/sys/windows"

func statFS(path string) (fsStats, error) {
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fsStats{}, err
	}
	err = windows.GetDiskFreeSpaceEx(pathPtr,
		&freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)
	if err != nil {
		return fsStats{}, err
	}
	return fsStats{
		totalBytes: totalNumberOfBytes,
		freeBytes:  freeBytesAvailable,
	}, nil
}
