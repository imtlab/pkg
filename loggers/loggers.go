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

func init() {
	/*	This will enable calls to loggers.(Info|Warning|Error).Whatever() to work as if log.Whatever() was called
		until Init() is called.
	*/
	p := log.Default()
	Info	= p
	Warning	= p
	Error	= p
//	Trace	= p
}

func Init(infoWriter io.Writer, warningWriter io.Writer, errorWriter io.Writer/*, traceWriter io.Writer*//*, bUTC bool*/) {
	flags := log.LstdFlags	//	LstdFlags = Ldate | Ltime
/*	if bUTC {
		flags |= log.LUTC
	}
*/
	//func New(out io.Writer, prefix string, flag int) *Logger
	Info	= log.New(infoWriter, `INFO: `, flags)
	Warning	= log.New(warningWriter, `WARNING: `, flags|log.Lshortfile)
	Error	= log.New(errorWriter, `ERROR: `, flags|log.Lshortfile)
//	Trace	= log.New(traceWriter, `TRACE: `, flags|log.Lshortfile)
}
