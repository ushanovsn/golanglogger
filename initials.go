package golanglogger

// Default log level
const initLogLevel = ErrorLvl

// Default logger channel buffer size
const initLogBuffer = 20

// Size buffer limits
const minBufferSize = 1
const maxBufferSize = 1000

// Default time checking file size (in seconds)
const initCheckTime = 60

// amount seconds in one day for calculating
const secondInDay = 60 * 60 * 24

// base format of time for write in log
const timeFormat = "2006-01-02 15:04:05.000"

// initializing base parameters (set parameters what not standart init values)
func getBaseParam() logParam {
	return logParam{logLvl: initLogLevel, lBuf: initLogBuffer, checkFileTime: initCheckTime, logFile: nil}
}
