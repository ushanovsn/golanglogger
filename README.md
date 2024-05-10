# golanglogger

Logger module (project within of language learning)


<!--Info block-->

## Use logger:
Logger created as goroutine and will works in async mode.  Messages for log are transmitted through the channel with no zero buffer size.
File day size control works in another goroutine and check the time for period 10 sec. It not enough accurate variant, but low count of file accesses (to prevent old logs from getting into the file, a time check is used for each writing).
A new file will be created after the specified duration at about 00:00.

1. At startup, need to create a logger:
```Go
log := golanglogger.New(golanglogger.DebugLvl, "test.log")
```
define logging level and file name for log, if log file name is empty - no loging to file

2. At stop, need to stop a logger for writing out all buffered messages:
```Go
defer log.StopLog()
```

3. After logger creation, it is possible to change the some parameters while program running:
    3.1. Change logging level:
    ```Go
    log.SetLevel(golanglogger.ErrorLvl)
    ```

    3.2. Change buffer size (default value: 20):
    ```Go
    log.SetBufferSize(100)
    ```
    available values at 1 to 1000

    3.3. Set file control parameters (default switched off values: 0, 0):
    ```Go
    log.SetFileParam(10, 1)
    ```
     available zero or any positive values, when zero - is disabling value.
     the value of the file size in MB
     the value of the file day size in days (day fixed in 00:00 for localtime with automatic using timezone)

4. Use writing log functions, for logging with necessary levels:
```Go
log.Out("This write everithing, without level control")
log.OutDebug("This write only when Debug level available")
log.OutInfo("This write only when Info level available and lower")
log.OutWarning("This write only when Warning level available and lower")
log.OutError("This write only when Error level available and lower")
```
"Debug" is minimum level, "Error" is maximum

