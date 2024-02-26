//go:build windows && !linux
package golanglogger

import (
	"os"
	"syscall"
)


// returning dateTime of creation file in seconds
func getDateTimeFile(fPath string) (int64, error) {
	f, err := os.Stat(fPath)

	if err != nil {
		return 0, err
	}

	// return datetime in seconds
	wFileSys := f.Sys().(*syscall.Win32FileAttributeData)
	return (wFileSys.CreationTime.Nanoseconds() / 1e9), nil
}
