package logging

import (
	"io"
	"log"
	"os"
)

var debugLogger *log.Logger = log.New(os.Stderr, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
var warnLogger *log.Logger = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime)
var errLogger *log.Logger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
var quietLogger *log.Logger = log.New(io.Discard, "", 0)

const DEBUG = 1

var LOG_LEVEL = 0

func IsDebug() bool {
	return LOG_LEVEL == DEBUG
}

func SetLogLevel(level int) {
	LOG_LEVEL = level
}

// Debug returns a logger that will only log in debug mode. Set the
func Debug() *log.Logger {
	if LOG_LEVEL != DEBUG {
		return quietLogger
	}

	return debugLogger
}

func Warn() *log.Logger {
	return warnLogger
}

func Error() *log.Logger {
	return errLogger
}
