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

func MustInject(data interface{}) {
    err := Inject(data)
    if err != nil {
        log.Panic(err)
    }
}

func Inject(data interface{}) error {
	rtype := reflect.TypeOf(data)

	if rtype.Kind() != reflect.Ptr {
		return errors.New("not a pointer")
	}

	name := rtype.Elem().Name()

	InjectWithName(name, data)

	return nil
}

func InjectWithName(name string, data interface{}) error {
	rtype := reflect.TypeOf(data)

	if rtype.Kind() != reflect.Ptr {
		return errors.New("not a pointer")
	}

	initByName(name, data)

	return nil
}

func LoadByName(name string, data interface{}) error {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("not a pointer")
	}

	container.m.RLock()
	service := container.services[name]
	container.m.RUnlock()

	if service != nil {
		rv.Elem().Set(reflect.ValueOf(service).Elem())
	} else {
		initByName(name, data)
	}

	return nil
}

func Load(data interface{}) error {
	rtype := reflect.TypeOf(data)
	if rtype.Kind() != reflect.Ptr {
		return errors.New("not a pointer")
	}

	name := rtype.Elem().Name()

	return LoadByName(name, data)
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

func initByName(name string, data interface{}) {
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
