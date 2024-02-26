package golanglogger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// logger level enums
type LoggingLevel int

const (
	Debug LoggingLevel = iota
	Info
	Warning
	Error
	definitely
)

// logger cmd enums
type loggingCmd int

const (
	cmdIdle loggingCmd = iota
	cmdWrite
	cmdStop
	cmdReloadConf
	cmdChangeFile
)

// struct for logger interaction
type logData struct {
	t   string
	msg string
	cmd loggingCmd
}

// struct for logger parameters
type logParam struct {
	// current log level
	logLvl LoggingLevel
	// disable console out log flag
	fNoCon bool
	// stderr out log flag
	fStdErr bool
	// logger buffer
	lBuf int
	// logger file
	logFile *os.File
	// file name for logging
	fileOutPath string
	// file size
	fileMbSize int
	// file duration
	fileDaySize int
}

// logger object type
type Logger struct {
	// lodders parameters
	param logParam
	// logger channel
	logChan chan logData
	// logger counter
	wg sync.WaitGroup
}

// create logger w params
func New(l LoggingLevel) *Logger {
	var log Logger
	log.param = getBaseParam()
	

	// set received parameters
	log.param.logLvl = l
	log.logChan = make(chan logData, log.param.lBuf)

	// start new logger
	go logger(log.logChan, &log.param, &log.wg)
	log.Out("Logger starting w log level = " + l.Name())

	return &log
}

// initializing base parameters (set parameters what not standart init values)
func getBaseParam() logParam {
	return logParam{logLvl: Error, lBuf: 20}
}

// stopping logger
func (log *Logger) StopLog() {
	writeCmd(log.logChan, cmdStop)
	log.wg.Wait()
}

// returns LoggingLevel const by text value (Error level if wrong name)
func LoggingLevelValue(s string) (LoggingLevel, error) {
	if l, ok := map[string]LoggingLevel{"debug": Debug, "info": Info, "warning": Warning, "error": Error}[strings.ToLower(s)]; ok {
		return l, nil
	}
	return Error, errors.New("Wrong LogLevel value: " + s)
}

// returns LoggingLevel name
func (l LoggingLevel) Name() string {
	return strings.ToUpper([]string{"Debug", "Info", "Warning", "Error"}[int(l)])
}

// set log level
func (log *Logger) SetLevel(l LoggingLevel) {
	log.param.logLvl = l
	writeCmd(log.logChan, cmdReloadConf)
	log.Out("Set logging level = " + l.Name())
}

// set log to file
func (log *Logger) SetFile(fPath string, mbSize int, daySize int) bool {
	logFile, err := os.OpenFile(fPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.OutError(`Error parameter for log file. Wrong parameter: "` + fPath + `". Error: ` + err.Error())
		return false
	}

	log.param.logFile = logFile
	log.param.fileOutPath = fPath
	log.param.fileMbSize = mbSize
	log.param.fileDaySize = daySize

	log.OutDebug("Logger will restart with new file parameters")
	writeCmd(log.logChan, cmdReloadConf)

	return true
}

// set log out writers parameters
func (log *Logger) SetErrOut(con bool, stdErr bool) bool {

	if !con && !stdErr {
		log.OutError("Error parameter for out log. Received STDOUT and STDERR false for both")
		return false
	}

	log.param.fNoCon = !con
	log.param.fStdErr = stdErr

	log.OutDebug("Logger will restart with new outs parameters")
	writeCmd(log.logChan, cmdReloadConf)

	return true
}

// write logging mesage
func writeLog(c chan logData, t time.Time, s string) {
	c <- logData{t.String(), s, cmdWrite}
}

// write logging command
func writeCmd(c chan<- logData, cmd loggingCmd) {
	c <- logData{"", "", cmd}
}

// out Debug string into log
func (log *Logger) OutDebug(msg string) {
	if int(log.param.logLvl) <= int(Debug) {
		writeLog(log.logChan, time.Now(), "[DBG]: "+msg)
	}
}

// out Info string into log
func (log *Logger) OutInfo(msg string) {
	if int(log.param.logLvl) <= int(Info) {
		writeLog(log.logChan, time.Now(), "[INF]: "+msg)
	}
}

// out Warning string into log
func (log *Logger) OutWarning(msg string) {
	if int(log.param.logLvl) <= int(Warning) {
		writeLog(log.logChan, time.Now(), "[WRN]: "+msg)
	}
}

// out Error string into log
func (log *Logger) OutError(msg string) {
	if int(log.param.logLvl) <= int(Error) {
		writeLog(log.logChan, time.Now(), "[ERR]: "+msg)
	}
}

// always out string into log
func (log *Logger) Out(msg string) {
	writeLog(log.logChan, time.Now(), "[MSG]: "+msg)
}

// logger goroutine
func logger(c chan logData, base_param *logParam, wg *sync.WaitGroup) {

	wg.Add(1)
	defer wg.Done()
	// type for inner errors, what going bypass channel
	type inner_error struct {
		t time.Time
		s string
	}

	// parameters copy values
	var param logParam = *base_param
	// log writer
	var writer io.Writer
	// timer for change file by time
	var tmr *time.Timer
	// stop func flag
	var fStop bool = false
	// change file flag
	var cCancel chan bool
	// current received command
	var cmd loggingCmd = cmdIdle

	// init bufer for inner functions errors
	inner_err := []inner_error{}

	// Now base cycle of logger func
	for !fStop {

		// check file day size control
		if param.fileDaySize > 0 && param.logFile != nil {
			const secondInDay = 60 * 60 * 24
			t, err := getDateTimeFile(param.fileOutPath)

			if err == nil {
				// time in seconds at day starting for file creating time + size log in seconds
				fileChangeDaySec := (t / secondInDay) + (int64(param.fileDaySize) * secondInDay)
				// if time is already gone...
				if time.Now().Unix() >= fileChangeDaySec {
					param.logFile, err = changeFile(param.logFile, param.fileOutPath)
					if err != nil {
						inner_err = append(inner_err, inner_error{time.Now(), "Error while changing file. Err: " + err.Error()})
					}
				}

				// timer to change time from now
				tmr = time.NewTimer(time.Duration(fileChangeDaySec-time.Now().Unix()) * time.Second)

				cCancel = make(chan bool)

				go func(c_t chan<- logData, t *time.Timer, c_c <- chan bool) {
					select {
					case <-t.C:
						writeCmd(c_t, cmdChangeFile)
					case <-c_c:
					}
				}(c, tmr, cCancel)

			} else {
				inner_err = append(inner_err, inner_error{time.Now(), "Error while getting datetime creating of file. Err: " + err.Error()})
			}
		}

		// check file MB size control
		if param.fileMbSize > 0 && param.logFile != nil {
			f, err := getFileMbSize(param.fileOutPath)
			if err == nil {
				if f >= param.fileMbSize {
					param.logFile, err = changeFile(param.logFile, param.fileOutPath)
					if err != nil {
						inner_err = append(inner_err, inner_error{time.Now(), "Error while changing file. Err: " + err.Error()})
					}
				}
			} else {
				inner_err = append(inner_err, inner_error{time.Now(), "Error while getting file MB size. Err: " + err.Error()})
			}
		}

		writer = nil
		// chk disabling console
		if !param.fNoCon {
			writer = io.Writer(os.Stdout)
		}
		// chk stdErr out
		if param.fStdErr && (writer == nil) {
			writer = io.Writer(os.Stderr)
		} else if param.fStdErr && (writer != nil) {
			writer = io.MultiWriter(writer, io.Writer(os.Stderr))
		}

		// writer already set to out now adding one more
		if param.logFile != nil {
			writer = io.MultiWriter(writer, param.logFile)
		}

		// writing errors into log what occurs during initialising
		if len(inner_err) > 0 {
			for _, v := range inner_err {
				_, err := io.WriteString(writer, v.t.String()+" -> "+"[ERR]: "+v.s+"\n")
				if err != nil {
					panic(err)
				}
			}
			inner_err = []inner_error{}
		}

		// getting data from channel cycle
		for msg := range c {
			cmd = msg.cmd

			if cmd != cmdWrite {
				break
			}

			_, err := io.WriteString(writer, msg.t+" -> "+msg.msg+"\n")
			if err != nil {
				panic(err)
			}
		}

		// processing commands
		switch cmd {
		case cmdChangeFile:
			var err error
			param.logFile, err = changeFile(param.logFile, param.fileOutPath)
			if err != nil {
				inner_err = append(inner_err, inner_error{time.Now(), "Error while changing file. Err: " + err.Error()})
			}
		case cmdReloadConf:
			//var tmpFile *os.File = param.logFile
			param = *base_param
			//param.logFile = tmpFile
		case cmdStop:
			fStop = true
		}

		// check timer and stopping it
		if tmr != nil && cCancel != nil {
			select {
			case cCancel <- true:
				close(cCancel)
			default:
			}

			tmr.Stop()
		}
	}
}

// returning file size and errors
func getFileMbSize(path string) (int, error) {
	f, err := os.Stat(path)
	if err == nil {
		return int(f.Size() / 1048576), nil
	} else {
		return 0, err
	}
}

// log file rotation
func changeFile(logFile *os.File, fileName string) (*os.File, error) {

	// close old file
	err := logFile.Close()
	if err != nil {
		return nil, err
	}

	// rename old file
	t := time.Now()
	formattedT := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	err = os.Rename(fileName, fileName+"_"+formattedT)
	if err != nil {
		return nil, err
	}

	// new file
	logFile, err = os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return logFile, nil
}
