package di

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestService struct {
	init        bool
	closed      bool
	nestService *NestService
}

type NestService struct {
	init   bool
	closed bool
}

func (s *NestService) Init() error {
	s.init = true
	return nil
}

func (s *TestService) Init() error {
	var nestService NestService
	if err := Load(&nestService); err != nil {
		return err
	}
	s.init = true
	s.nestService = &nestService
	return nil
}

func (s *TestService) Close() error {
	s.closed = true
	return nil
}

func TestLoad(t *testing.T) {
	// run with --race flag
	t.Run("detect race condition", func(t *testing.T) {
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()

			var t1 TestService
			Load(&t1)
		}()

		go func() {
			defer wg.Done()
			var t2 NestService
			Load(&t2)
		}()

		wg.Wait()
	})

	t.Run("load service", func(t *testing.T) {
		var ts TestService
		err := Load(&ts)
		assert.NoError(t, err)
		assert.True(t, ts.init)
	})

	t.Run("load nested service", func(t *testing.T) {
		var ts TestService
		err := Load(&ts)
		assert.NoError(t, err)
		assert.NotNil(t, ts.nestService)
		assert.True(t, ts.nestService.init)
	})

	t.Run("load existing service", func(t *testing.T) {
		container.services["TestService"] = &TestService{init: true}
		var ts TestService
		err := Load(&ts)
		assert.NoError(t, err)
		assert.Nil(t, ts.nestService)
	})

	t.Run("only pointer is allowed", func(t *testing.T) {
		var ts TestService
		err := Load(ts)
		assert.Error(t, err)
	})

	t.Run("nil value is not allowed", func(t *testing.T) {
		var ts *TestService
		err := Load(ts)
		assert.Error(t, err)
	})
}

func TestClose(t *testing.T) {
	var ts TestService
	container.services["TestService"] = &ts

	err := Close()
	assert.NoError(t, err)
	assert.True(t, ts.closed)
}
