/*	Package throttle provides a synchronizable type with methods used to limit service calls
	within any period of duration seconds.
*/
package throttle

import (
	"log"
	"sync"
	"time"
)

type TThrottle struct {
	sync.Mutex
	limit		int
	duration	time.Duration
	fifo		chan time.Time
}

func (p *TThrottle) Init(numeratorLimit int, denominatorSeconds int) {
	p.limit = numeratorLimit
	p.duration = time.Second * time.Duration(denominatorSeconds)
	p.fifo = make(chan time.Time, numeratorLimit)
}

func (p *TThrottle) GetSleepDuration() (dur time.Duration) {
	if nil == p.fifo {	//	guard against pinheadedness
		log.Fatalln("Programmer forgot to call Init() on the throttle object")
	}
	p.Lock()
	now := time.Now()	//	latch this instant

	if(p.limit == len(p.fifo)) {
		//	remove the oldest instant from p.fifo and compute the instant of the next allowable call
		next := (<-p.fifo).Add(p.duration)

		if(next.After(now)) {
			p.fifo <- next
			dur = next.Sub(now)
		} else {
			p.fifo <- now
		}
	} else {
		p.fifo <- now
	}

	p.Unlock()
	return
}
