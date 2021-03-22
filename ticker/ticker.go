package ticker

import (
	"time"
)

type Ticker struct {
	ticker   *time.Ticker
	f        func()
	interval time.Duration
}

func NewTicker(interval time.Duration, f func()) *Ticker {
	return &Ticker{
		ticker:   time.NewTicker(interval),
		f:        f,
		interval: interval,
	}
}
func (t *Ticker) run() {
	for {
		<-t.ticker.C
		go t.f()
	}
}
func (t *Ticker) Start() {
	go t.run()
}
func (t *Ticker) Close() {
	t.ticker.Stop()
}
