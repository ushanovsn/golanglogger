package golanglogger

import (
	"os"
	"time"
)



// returning file size and errors
func getFileMbSize(path string) (int, error) {
	f, err := os.Stat(path)
	if err == nil {
		return int(f.Size() / 1048576), nil
	} else {
		return 0, err
	}
}



// log file rotation (write into file must be paused while process it)
func changeFile(logFile *os.File, fileName string, is_rotation bool) (*os.File, error) {

	// close old file
	err := logFile.Close()
	if err != nil {
		return nil, err
	}

	if is_rotation {
		// rename old file
		t := time.Now()
		formattedT := t.Format("2006-01-02T15-04-05.000")
		err = os.Rename(fileName, fileName+"_"+formattedT)
		if err != nil {
			return nil, err
		}
	}

	// new file
	logFile, err = os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return logFile, nil
}

// creating the file
func getFileOut(fPath string) (*os.File, error) {
	logFile, err := os.OpenFile(fPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return logFile, nil
}
