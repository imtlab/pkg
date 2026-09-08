package sharedstate

import (
	"fmt"
	"strings"
	"sync"

//	"github.com/imtlab/pkg/loggers"
	"github.com/imtlab/pkg/utils"

//	"github.com/imtlab/govdbp/iconikagent/build"
)

//\\//	package-scope constants and variables
/*
const (
)

var (
)
*/

//\\//	type definitions (and attached methods)

/*	TSharedState is used by various handlers and lambda functions to safely (synchronously) accumulate
	progress and un-scoped errors and warnings destined for iconoik Jobs.
	Some handlers pass a pointer to TSharedState into a Producer function in the iconik package,
	hence the need for it to be defined in this separate package to avoid cyclical imports.
*/
type TSharedState struct {
						sync.RWMutex	//	embedded struct
	JobProgressTotal	uint64
	JobProgress			uint64
	errors				[]error
	warnings			[]string
}

/* not needed because only the producer (which is only 1 go routine) will update, so no concurrency.
func (p *TSharedState) SetJobProgress(progress uint64) {
	p.RLock()
	p.JobProgress = progress
	p.RUnlock()
}
*/

func (p *TSharedState) IncrementJobProgress() {
	p.RLock()
	p.JobProgress++
	p.RUnlock()
}

func (p *TSharedState) GetJobProgress() (value uint64) {
	p.RLock()
	value = p.JobProgress
	p.RUnlock()
	return
}

func (p *TSharedState) AppendError(err error) {
	p.RLock()
	p.errors = append(p.errors, err)
	p.RUnlock()
}

//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) HasErrors() (value bool) {
//	p.RLock()
	value = (0 != len(p.errors))
//	p.RUnlock()
	return
}

//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) ErrorCount() (value int) {
//	p.RLock()
	value = len(p.errors)
//	p.RUnlock()
	return
}

//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) Errors() (value []error) {
//	p.RLock()
	value = p.errors
//	p.RUnlock()
	return
}

//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) JoinedError() (err error) {
	if count := len(p.errors); 0 != count {
//		p.RLock()
		if 1 == count {
			err = p.errors[0]
		} else {
			//	combine the errors into one
			err = fmt.Errorf(`%d errors: %s`, count, strings.Join(utils.ErrorsToMessages(p.errors), `; `))
		}
//		p.RUnlock()
	}

	return
}

func (p *TSharedState) AppendWarning(s string) {
	p.RLock()
	p.warnings = append(p.warnings, s)
	p.RUnlock()
}

//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) HasWarnings() (value bool) {
//	p.RLock()
	value = (0 != len(p.warnings))
//	p.RUnlock()
	return
}

//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) JoinedWarnings() (s string) {
	if count := len(p.warnings); 0 != count {
//		p.RLock()
		if 1 == count {
			s = p.warnings[0]
		} else {
			//	combine the warnings into one
			s = fmt.Sprintf(`%d warnings: %s`, count, strings.Join(p.warnings, `; `))
		}
//		p.RUnlock()
	}

	return
}

//\\//	functions
/*
func init() {
	if build.LocalTest {
		loggers.Info.Println(`Executing sharedstate.init()`)
	}
}
*/
