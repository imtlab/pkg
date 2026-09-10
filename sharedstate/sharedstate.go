/*	Package sharedstate provides a synchronizable type with methods used by various handlers and lambda functions
	to safely (synchronously) accumulate progress and un-scoped errors and warnings destined for iconoik Jobs.
	Some handlers pass a pointer to TSharedState into a Producer function in the iconik package,
	hence the need for it to be defined in this separate package to avoid cyclical imports.

*/
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

type TSharedState struct {
					sync.RWMutex	//	embedded struct
	//	exported
	ProgressTotal	uint64
	Progress		uint64
	//	internal
	xErrors			[]error
	xWarnings		[]string
}

//\\//	methods for working with the "Progress" field

/* not needed because only the producer (which is only 1 go routine) will update Progress, so no concurrency.
func (p *TSharedState) SetProgress(progress uint64) {
	p.RLock()
	p.Progress = progress
	p.RUnlock()
}
*/

/*<<<<	COMMENTED OUT just to determine if these two are used anywhere
func (p *TSharedState) IncrementProgress() {
	p.RLock()
	p.Progress++
	p.RUnlock()
}

func (p *TSharedState) GetProgress() (value uint64) {
	p.RLock()
	value = p.Progress
	p.RUnlock()
	return
}
*/

/*	If the only time GetProgress() is called is immediately after calling IncrementProgress()
	Then they should be combined into one within a single RLock/RUnlock block.
*/
func (p *TSharedState) GetIncrementedProgress() (value uint64) {
	p.RLock()
	p.Progress++
	value = p.Progress
	p.RUnlock()
	return
}

//\\//	methods for working with the xErrors field

func (p *TSharedState) AppendError(err error) {
	p.RLock()
	p.xErrors = append(p.xErrors, err)
	p.RUnlock()
}

//	not called from concurrent go routines so RLock and RUnlock not needed
/*
func (p *TSharedState) HasErrors() (value bool) {
//	p.RLock()
	value = (0 != len(p.xErrors))
//	p.RUnlock()
	return
}
*/
func (p *TSharedState) HasErrors() bool {
	return 0 != len(p.xErrors)
}

//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) ErrorCount() (value uint) {
//	p.RLock()
	value = uint(len(p.xErrors))
//	p.RUnlock()
	return
}

//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) Errors() (value []error) {
//	p.RLock()
	value = p.xErrors
//	p.RUnlock()
	return
}

//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) JoinedError() (err error) {
	if count := len(p.xErrors); 0 != count {
//		p.RLock()
		if 1 == count {
			err = p.xErrors[0]
		} else {
			//	combine the errors into one
			err = fmt.Errorf(`%d errors: %s`, count, strings.Join(utils.ErrorsToMessages(p.xErrors), `; `))
		}
//		p.RUnlock()
	}

	return
}

//	JoinedErrorMsg() is more efficient than JoinedError() when the caller only wants err.Error() anyway.
//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) JoinedErrorMsg() (s string) {
	if count := len(p.xErrors); 0 != count {
//		p.RLock()
		if 1 == count {
			s = p.xErrors[0].Error()
		} else {
			//	combine the errors into one
			s = fmt.Sprintf(`%d errors: %s`, count, strings.Join(utils.ErrorsToMessages(p.xErrors), `; `))
		}
//		p.RUnlock()
	}

	return
}

//\\//	methods for working with the xWarnings field

func (p *TSharedState) AppendWarning(s string) {
	p.RLock()
	p.xWarnings = append(p.xWarnings, s)
	p.RUnlock()
}

//	not called from concurrent go routines so RLock and RUnlock not needed
/*
func (p *TSharedState) HasWarnings() (value bool) {
//	p.RLock()
	value = (0 != len(p.xWarnings))
//	p.RUnlock()
	return
}
*/
func (p *TSharedState) HasWarnings() bool {
	return 0 != len(p.xWarnings)
}

//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) WarningCount() (value uint) {
//	p.RLock()
	value = uint(len(p.xWarnings))
//	p.RUnlock()
	return
}

//	not called from concurrent go routines so RLock and RUnlock not needed
func (p *TSharedState) JoinedWarnings() (s string) {
	if count := len(p.xWarnings); 0 != count {
//		p.RLock()
		if 1 == count {
			s = p.xWarnings[0]
		} else {
			//	combine the warnings into one
			s = fmt.Sprintf(`%d warnings: %s`, count, strings.Join(p.xWarnings, `; `))
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
