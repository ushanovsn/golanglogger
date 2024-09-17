package golanglogger

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// logger object type
type LoggerSync struct {
	// logger mutex for param interactions
	rmu sync.RWMutex
	// logger parameters
	param logParam
	// logger running flag
	fLogRun bool
	// logger multiwriter
	writer *io.Writer
}

// create logger w params (if file path is "" then no logging into file)
func NewSync(l LoggingLevel, filePath string) Golanglogger {
	var log LoggerSync
	var err error
	log.param = getBaseParam()

	// set received level
	log.param.logLvl = l
	// set received path
	log.param.fileOutPath = filePath
	// indicate logger started
	log.fLogRun = true

	// set out into file if needed
	if log.param.fileOutPath != "" {
		log.param.logFile, err = getFileOut(log.param.fileOutPath)
		if err != nil {
			log.Error("Error while init log-file: " + err.Error())
		}
	}

	// set result writer
	writer := getWriter(log.param.fNoCon, log.param.fStdErr, log.param.logFile)
	log.writer = &writer

	// the greeting line after logger was created
	log.Out("Logger (sync) starting w log level = " + log.param.logLvl.Name())

	return &log
}

// stopping logger
func (log *LoggerSync) StopLog() {
	log.Out("Logger (sync) stopping by command \"STOP\"")
	log.Out("********************************************************************************************\n\n\n")

	log.rmu.Lock()
	log.fLogRun = false
	log.rmu.Unlock()

	// now logger stopped

	if log.param.logFile != nil {
		err := log.param.logFile.Close()
		if err != nil {
			fmt.Printf("Error while log file closing:  \"%s\"\n", err)
		}
	}

	log.param.logFile = nil
}

// set log level
func (log *LoggerSync) SetLevel(l LoggingLevel) {
	log.rmu.Lock()
	log.param.logLvl = l
	log.rmu.Unlock()

	log.Out("Set logging level = " + l.Name())
}

// set log channel buffer size
func (log *LoggerSync) SetBufferSize(bSize int) {
	log.rmu.Lock()
	log.param.lBuf = bSize
	log.rmu.Unlock()

	log.Out("Logger buffer size don't uses for synchronous logger")
}

// set log file control parameters
func (log *LoggerSync) SetFileParam(mbSize int, daySize int) {
	if mbSize < 0 {
		log.Error(fmt.Sprintf("Wrong file size control value: %d", mbSize))
		return
	}
	if daySize < 0 {
		log.Error(fmt.Sprintf("Wrong day size for file control value: %d", daySize))
		return
	}

	log.rmu.Lock()
	log.param.fileMbSize = mbSize
	log.param.fileDaySize = daySize
	log.rmu.Unlock()

	log.Debug(fmt.Sprintf("Logger set new file parameters: %d Mb size, %d days duration", mbSize, daySize))
}

// set log out writers parameters (con - console out flag, stdErr - standart error out flag)
func (log *LoggerSync) SetStdOut(con bool, stdErr bool) {

	if !con && !stdErr {
		log.Error("Error parameter for out log. Received STDOUT and STDERR false for both. Skip this.")
		return
	}

	log.Debug("Logger will works with new outs parameters")

	writer := getWriter(!con, stdErr, log.param.logFile)

	log.rmu.Lock()
	log.param.fNoCon = !con
	log.param.fStdErr = stdErr
	log.writer = &writer
	log.rmu.Unlock()
}

// set log name
func (log *LoggerSync) SetName(nameStr string) {
	log.rmu.Lock()
	log.param.name = nameStr
	if nameStr == "" {
		log.param.namePref = ""
	} else {
		log.param.namePref = nameStr + ": "
	}
	log.rmu.Unlock()
}

// get logger parameters for writing messages
func (log *LoggerSync) CurrentOutParams() (con bool, stdErr bool, filePath string) {

	log.rmu.RLock()
	con = !log.param.fNoCon
	stdErr = log.param.fStdErr
	filePath = log.param.fileOutPath
	log.rmu.RUnlock()

	return con, stdErr, filePath
}

// get logger buffer size
func (log *LoggerSync) CurrentBufSize() (size int) {

	log.rmu.RLock()
	size = log.param.lBuf
	log.rmu.RUnlock()

	return size
}

// get logger level
func (log *LoggerSync) CurrentLevel() (lvl LoggingLevel) {

	log.rmu.RLock()
	lvl = log.param.logLvl
	log.rmu.RUnlock()

	return lvl
}

// get logger file control parameters
func (log *LoggerSync) CurrentFileControl() (mbSize int, daySize int) {

	log.rmu.RLock()
	mbSize = log.param.fileMbSize
	daySize = log.param.fileDaySize
	log.rmu.RUnlock()

	return mbSize, daySize
}

// get logger level
func (log *LoggerSync) CurrentName() (nameStr string) {

	log.rmu.RLock()
	nameStr = log.param.name
	log.rmu.RUnlock()

	return nameStr
}

// write logging mesage
func (log *LoggerSync) writeLogSync(t time.Time, s string) {
	// generate message
	msg := logData{cmd: cmdWrite, t: t.Format(timeFormat), msg: s}

	if (log.param.fileMbSize > 0 || log.param.fileDaySize > 0) && log.param.logFile != nil {
		// if exists file control conditions - need to check it
		var fileChange bool
		// error holder, when can't write in file
		var procErrorMsg logData
		// error
		var err error

		// blocking another log writing
		log.rmu.Lock()

		// file duration control check first
		if log.param.fileDaySize > 0 {
			// current time zone offset
			_, zOffset := time.Now().In(time.Local).Zone()
			// check file time duration
			fileChange, _ = checkFileTime(zOffset, log.param.fileOutPath, (log.param.fileDaySize * secondInDay))

			// when need to change file what duration is full
			if fileChange {
				// change file directly
				log.param.logFile, err = changeFile(log.param.logFile, log.param.fileOutPath, true)
				if err != nil {
					procErrorMsg.setVal(cmdIdle, fmt.Sprintf("Error while changing log file by time. Err: %s", err.Error()), time.Now().Format(timeFormat))
					// new writer wo file out
					writer := getWriter(log.param.fNoCon, log.param.fStdErr, nil)
					log.writer = &writer
				} else {
					// new writer into new file - set file to main parameters of base logger
					writer := getWriter(log.param.fNoCon, log.param.fStdErr, log.param.logFile)
					log.writer = &writer
				}

				if procErrorMsg.msg != "" {
					err = writeLog(log.writer, log.param.namePref, procErrorMsg)
					if err != nil {
						fmt.Printf("Not logged ERROR when changing file: %s, for MESSAGE: %s", err, procErrorMsg.msg)
					}

					// skip error for use it after
					procErrorMsg.reset()
				}

				// skip flag for use it after
				fileChange = false
			}
		}

		// writing log
		err = writeLog(log.writer, log.param.namePref, msg)
		if err != nil {
			fmt.Printf("Not logged ERROR: %s, when writing MESSAGE: %s", err, msg.msg)
		}

		// check file MB size control enabled after writing new message
		if log.param.fileMbSize > 0 && log.param.logFile != nil {
			f, err := getFileMbSize(log.param.fileOutPath)
			if err != nil {
				procErrorMsg.setVal(cmdIdle, fmt.Sprintf("Error while get file size. Err: %s", err.Error()), time.Now().Format(timeFormat))
			}

			if procErrorMsg.msg != "" {
				// error when check file size
				err = writeLog(log.writer, log.param.namePref, procErrorMsg)
				if err != nil {
					fmt.Printf("Not logged ERROR when get file size: %s, for MESSAGE: %s", err, procErrorMsg.msg)
				}
				procErrorMsg.reset()
			} else {

				// when file is too big - need to change it
				if f >= log.param.fileMbSize {
					fileChange = true
				}

				if fileChange {
					// change file directly
					log.param.logFile, err = changeFile(log.param.logFile, log.param.fileOutPath, true)
					if err != nil {
						procErrorMsg.setVal(cmdIdle, fmt.Sprintf("Error while changing log file by size. Err: %s", err.Error()), time.Now().Format(timeFormat))
						// new writer wo file out
						writer := getWriter(log.param.fNoCon, log.param.fStdErr, nil)
						log.writer = &writer
					} else {
						// new writer into new file - set file to main parameters of base logger
						writer := getWriter(log.param.fNoCon, log.param.fStdErr, log.param.logFile)
						log.writer = &writer
					}

					if procErrorMsg.msg != "" {
						err = writeLog(log.writer, log.param.namePref, procErrorMsg)
						if err != nil {
							fmt.Printf("Not logged ERROR when changing file: %s, for MESSAGE: %s", err, procErrorMsg.msg)
						}
						procErrorMsg.reset()
					}
				}
			}
		}

		// unlock accef to log writing
		log.rmu.Unlock()

	} else {
		// write message without conditions
		log.rmu.Lock()
		err := writeLog(log.writer, log.param.namePref, msg)
		log.rmu.Unlock()

		// write error into console
		if err != nil {
			fmt.Printf("Not logged ERROR: %s, when writing MESSAGE: %s", err, msg.msg)
		}
	}
}

// out Debug string into log
func (log *LoggerSync) Debug(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(DebugLvl) && log.fLogRun {
		log.writeLogSync(time.Now(), "[DBG]: "+msg)
	}
}

// out Info string into log
func (log *LoggerSync) Info(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(InfoLvl) && log.fLogRun {
		log.writeLogSync(time.Now(), "[INF]: "+msg)
	}
}

// out Warning string into log
func (log *LoggerSync) Warn(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(WarningLvl) && log.fLogRun {
		log.writeLogSync(time.Now(), "[WRN]: "+msg)
	}
}

// out Error string into log
func (log *LoggerSync) Error(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(ErrorLvl) && log.fLogRun {
		log.writeLogSync(time.Now(), "[ERR]: "+msg)
	}
}

// always out string into log
func (log *LoggerSync) Out(msg string) {
	if log.fLogRun {
		log.writeLogSync(time.Now(), "[MSG]: "+msg)
	}

}
