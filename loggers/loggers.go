package loggers

import (
	"log"
	"io"
)

var (
	// these package-level variables are exported to any package importing this one
	Info	*log.Logger
	Warning	*log.Logger
	Error	*log.Logger
//	Trace	*log.Logger
)

func Init(infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer/*, traceHandle io.Writer*/) {
	Info	= log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime/*|log.LUTC*/)
	Warning	= log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile/*|log.LUTC*/)
	Error	= log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile/*|log.LUTC*/)
//	Trace	= log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile/*|log.LUTC*/)
}
