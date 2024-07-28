package id_server

import (
	"fmt"
	"time"
)

// IdGeneratorManager holds instances of all kinds of id generator.
type IdGeneratorManager struct {
	snowflake *SnowflakeIdGenerator
}

// Generate generates distributed id number based expilcit
// generating algorithm.
func (igm *IdGeneratorManager) Generate(algorithm string) (Identity, error) {
	var generator IdGenerator

	switch algorithm {
	case "snowflake":
		generator = igm.snowflake
	default:
		return nil, fmt.Errorf("no such generating algorithm")
	}

	return generator.Generate()
}

// Start IdGeneratorManager
func (igm *IdGeneratorManager) Run(sig chan struct{}) error {
	runners := make([]IdGeneratorRunner, 1)
	runners[0] = getSnowflakeRunner(igm.snowflake)

	errs := make(chan error, 1)
	defer close(errs)
	for _, r := range runners {
		go func(r IdGeneratorRunner) {
			err := r()
			if err != nil {
				errs <- err
				sig <- struct{}{}
			}
		}(r)
	}

	<-sig
	if len(errs) > 0 {
		if e, ok := <-errs; ok {
			return e
		}
	}
	return nil
}

func getSnowflakeRunner(generator IdGenerator) IdGeneratorRunner {
	return func() error {
		sfg, ok := generator.(*SnowflakeIdGenerator)
		if !ok {
			return fmt.Errorf("illegal id generator type, need snowflake algorithm")
		}

		ticker := time.NewTicker(time.Millisecond)
		for {
			<-ticker.C
			sfg.timestampLock.Lock()
			sfg.curTimestamp = uint64(time.Now().UnixMilli())
			sfg.timestampLock.Unlock()
		}
	}
}

// NewIdGeneratorManager creates a new manager.
func NewIdGeneratorManager() (*IdGeneratorManager, error) {
	manager := new(IdGeneratorManager)
	var err error

	manager.snowflake, err = NewSnowflakeIdGenerator(42, 5, 5)
	if err != nil {
		return nil, err
	}

	return manager, err
}
