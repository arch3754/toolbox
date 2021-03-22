package cache

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/arch3754/toolbox/log"
)

type Shard struct {
	query    map[string][]string
	interval time.Duration
	object   interface{}
	f        func() (interface{}, error)
	signal   chan int
	once     sync.Once
}

func (c *Shard) get() (interface{}, error) {
	if c.object != nil {
		return c.object, nil
	}
	if err := c.set(); err != nil {
		return nil, err
	}
	return c.object, nil
}
func (c *Shard) set() error {
	val, err := c.f()
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("not ptr or nil")
	}
	c.object = val
	c.once.Do(c.loop)
	return nil
}
func (c *Shard) loop() {
	go func() {
		ticker := time.NewTicker(c.interval)
		for {
			select {
			case <-ticker.C:
				err := c.set()
				if err != nil {
					log.Rlog.Error("ticker set failed.err:%v", err)
				}
			case <-c.signal:
				ticker.Stop()
				return
			}
		}
	}()
}