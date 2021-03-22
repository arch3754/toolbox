package cache

import (
	"fmt"
	"reflect"
	"time"
)

type Cache struct {
	shards map[string]*Shard
}

func New() *Cache {
	return &Cache{shards: make(map[string]*Shard)}
}

func (c *Cache) NewShard(key string, interval time.Duration, fargs map[string][]string, f func() (interface{}, error)) {

	var shard = &Shard{
		interval: interval,
		f:        f,
		query:    fargs,
	}
	c.shards[key] = shard
}
func (c *Cache) Close() {
	for _, v := range c.shards {
		v.signal <- 1
	}
}
func (c *Cache) GetMapByKey(key string, mapKey string, value interface{}) error {
	shard, ok := c.shards[key]
	if !ok {
		return fmt.Errorf("not exist %v", key)
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("value is nil or not pointer,%v", rv.Kind())
	}

	data, err := shard.get()
	if err != nil {
		return err
	}

	valuet := reflect.TypeOf(value).Elem()

	mapReflect := reflect.MakeMap(valuet)

	nv := reflect.ValueOf(data).Elem()
	for i := 0; i < nv.Len(); i++ {
		val := nv.Index(i).FieldByName(mapKey)
		mapReflect.SetMapIndex(val, nv.Index(i).Addr())
	}
	rv.Elem().Set(mapReflect)

	return nil
}
func (c *Cache) GetsMapByKey(key string, mapKey string, value interface{}) error {
	shard, ok := c.shards[key]
	if !ok {
		return fmt.Errorf("not exist %v", key)
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("value is nil or not pointer,%v", rv.Kind())
	}

	data, err := shard.get()
	if err != nil {
		return err
	}

	valuet := reflect.TypeOf(value).Elem()
	mapReflect := reflect.MakeMap(valuet)
	var mapSlice = make(map[interface{}][]reflect.Value)
	nv := reflect.ValueOf(data).Elem()
	for i := 0; i < nv.Len(); i++ {
		val := nv.Index(i).FieldByName(mapKey)
		mapSlice[val.Interface()] = append(mapSlice[val.Interface()], nv.Index(i).Addr())
	}

	for k, v := range mapSlice {
		x := reflect.New(reflect.SliceOf(v[0].Type()))
		tmp := reflect.Append(x.Elem(), v...)
		x.Elem().Set(tmp)
		mapReflect.SetMapIndex(reflect.ValueOf(k), x.Elem())
	}
	rv.Elem().Set(mapReflect)
	return nil
}
func (c *Cache) Get(key string, value interface{}) error {
	shard, ok := c.shards[key]
	if !ok {
		return fmt.Errorf("not exist %v", key)
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("value is nil or not pointer")
	}

	val, err := shard.get()
	if err != nil {
		return err
	}
	tf := reflect.TypeOf(value).Elem()
	nf := reflect.TypeOf(val).Elem()
	if tf.String() != nf.String() {
		return fmt.Errorf("value of type  %v is not assignable to type %v", tf.String(), nf.String())
	}
	rv.Elem().Set(reflect.ValueOf(val).Elem())

	return nil
}
func (c *Cache) Set() error {
	for _, v := range c.shards {
		err := v.set()
		if err != nil {
			return err
		}
	}
	return nil
}
