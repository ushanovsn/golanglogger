//go:build linux && !windows
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
	lFileSys := f.Sys().(*syscall.Stat_t)
	return int64(lFileSys.Atim.Sec), nil
}
