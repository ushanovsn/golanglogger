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
	rmu *sync.RWMutex
	// logger parameters
	param *logParam
	// logger channel
	cCancel chan struct{}
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
	// init log linking objects
	log.rmu = &sync.RWMutex{}
	// indicate logger started
	log.fLogRun = true

	// set out into file if needed
	if log.param.fileOutPath != "" {
		log.param.logFile, err = getFileOut(log.param.fileOutPath)
		if err != nil {
			log.OutError("Error while init log-file: " + err.Error())
		}
	}

	// the greeting line after logger was created
	log.Out("Logger starting w log level = " + log.param.logLvl.Name())


	return &log
}


// stopping logger
func (log *LoggerSync) StopLog() {
	log.Out("Logger stopping by command \"STOP\"")

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
		log.OutError(fmt.Sprintf("Wrong file size control value: %d", mbSize))
		return
	}
	if daySize < 0 {
		log.OutError(fmt.Sprintf("Wrong day size for file control value: %d", daySize))
		return
	}

	log.rmu.Lock()
	log.param.fileMbSize = mbSize
	log.param.fileDaySize = daySize
	log.rmu.Unlock()

	log.OutDebug(fmt.Sprintf("Logger set new file parameters: %d Mb size, %d days duration", mbSize, daySize))
}

// set log out writers parameters (con - console out flag, stdErr - standart error out flag)
func (log *LoggerSync) SetStdOut(con bool, stdErr bool) {

	if !con && !stdErr {
		log.OutError("Error parameter for out log. Received STDOUT and STDERR false for both. Skip this.")
		return
	}

	log.OutDebug("Logger will works with new outs parameters")

	writer := getWriter(!con, stdErr, log.param.logFile)

	log.rmu.Lock()
	log.param.fNoCon = !con
	log.param.fStdErr = stdErr
	log.writer = &writer
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

// write logging mesage
func writeLogSync(w *io.Writer, t time.Time, s string) {
	msg := createMsg(t, s)
	_, err := io.WriteString(*w, msg)
	if err != nil {
		fmt.Printf("Error while write log string \"%s\"; Error: %s\n", msg, err.Error())
	}
}

// out Debug string into log
func (log *LoggerSync) OutDebug(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(DebugLvl) && log.fLogRun {
		log.rmu.Lock()
		writeLogSync(log.writer, time.Now(), "[DBG]: "+msg)
		log.rmu.Unlock()
	}
}

// out Info string into log
func (log *LoggerSync) OutInfo(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(InfoLvl) && log.fLogRun {
		log.rmu.Lock()
		writeLogSync(log.writer, time.Now(), "[INF]: "+msg)
		log.rmu.Unlock()
	}
}

// out Warning string into log
func (log *LoggerSync) OutWarning(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(WarningLvl) && log.fLogRun {
		log.rmu.Lock()
		writeLogSync(log.writer, time.Now(), "[WRN]: "+msg)
		log.rmu.Unlock()
	}
}

// out Error string into log
func (log *LoggerSync) OutError(msg string) {
	// don't use mutex, the level will  changing rare and it not important if reading old or new value of it
	if int(log.param.logLvl) <= int(ErrorLvl) && log.fLogRun {
		log.rmu.Lock()
		writeLogSync(log.writer, time.Now(), "[ERR]: "+msg)
		log.rmu.Unlock()
	}
}

// always out string into log
func (log *LoggerSync) Out(msg string) {
	if log.fLogRun {
		log.rmu.Lock()
		writeLogSync(log.writer, time.Now(), "[MSG]: "+msg)
		log.rmu.Unlock()
	}

}


// combining message for log
func createMsg(t time.Time, s string) string {
	return  (t.Format("2006-01-02 15:04:05.000") + " -> " + s + "\n")
}