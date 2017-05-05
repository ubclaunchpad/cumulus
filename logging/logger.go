package logger

import (
    "os"

    gologging "github.com/op/go-logging"
)

// Set up pretty logger
var (
    Log = gologging.MustGetLogger("example")
    format = gologging.MustStringFormatter(
    	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{color:reset} %{message}`,
    )

    logBackend = gologging.NewLogBackend(os.Stderr, "", 0)

    // For messages written to logBackend we want to add some additional
    // information to the output, including the used log level and the name of
    // the function.
    logBackendFormatter = gologging.NewBackendFormatter(logBackend, format)

    // Only errors and more severe messages should be sent to LogBackend
    logBackendLeveled = gologging.AddModuleLevel(logBackend)
)

// Set up logging level and backend
func Init() {
    logBackendLeveled.SetLevel(gologging.ERROR, "")
    gologging.SetBackend(logBackendLeveled, logBackendFormatter)
}
