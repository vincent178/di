package di

import (
	"errors"
	"log"
	"reflect"
	"sync"
)

type Initializer interface {
	Init() error
}

type Closer interface {
	Close() error
}

type Container struct {
	services map[string]interface{}
	m        sync.RWMutex
}

var container = &Container{
	services: make(map[string]interface{}),
}

func Load(data interface{}) error {
	rtype := reflect.TypeOf(data)
	rv := reflect.ValueOf(data)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("not a pointer")
	}

	name := rtype.Elem().Name()

	container.m.RLock()
	service := container.services[name]
	container.m.RUnlock()

	if service != nil {
		rv.Elem().Set(reflect.ValueOf(service).Elem())
	} else {
		var once sync.Once
		once.Do(func() {
			if initializer, ok := data.(Initializer); ok {
				if err := initializer.Init(); err != nil {
					log.Panic("init error", err)
				}
			}
			container.m.Lock()
			container.services[name] = data
			container.m.Unlock()
		})
	}

	return nil
}

func Close() error {
	for _, service := range container.services {
		if closer, ok := service.(Closer); ok {
			if err := closer.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}
