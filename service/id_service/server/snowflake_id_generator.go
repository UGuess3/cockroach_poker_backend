package id_server

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ------------------------------------------------
// | timestamp | data center id | service id | id |
// ------------------------------------------------
type SnowflakeIdentity uint64

var (
	// start time of system
	systemStartTime = uint64(time.Date(2024, time.July, 27, 0, 0, 0, 0, time.UTC).UnixMilli())
)

// SnowflakeConfiguration stores configuration of snowflake algorithm.
type SnowflakeConfiguration struct {
	maximumTimestamp    uint64
	maximumDataCenterId uint64
	maximumServiceId    uint64
	maximumId           uint64
	timestampOffset     int
	dataCenterIdOffset  int
	serviceIdOffset     int
	idOffset            int
}

// SnowflakeIdGenerator generator generates id based on
// snowflake algorithm.
type SnowflakeIdGenerator struct {
	config        *SnowflakeConfiguration
	curTimestamp  uint64
	curId         uint64
	dataCenterId  uint64
	serviceId     uint64
	idLock        sync.Mutex
	timestampLock sync.Mutex
}

// NewSnowflakeIdGenerator creates a new SnowflakeIdGenerator
func NewSnowflakeIdGenerator(timestampBitNum int, dataCenterBitNum int,
	serviceIdBitNum int) (*SnowflakeIdGenerator, error) {
	if !checkBitNum(timestampBitNum) ||
		!checkBitNum(dataCenterBitNum) ||
		!checkBitNum(serviceIdBitNum) {
		return nil, fmt.Errorf("illegal number of bits: %d, %d, %d",
			timestampBitNum,
			dataCenterBitNum,
			serviceIdBitNum,
		)
	}

	config := new(SnowflakeConfiguration)

	config.idOffset = 64 - timestampBitNum - dataCenterBitNum - serviceIdBitNum
	config.serviceIdOffset = config.idOffset + serviceIdBitNum
	config.dataCenterIdOffset = config.serviceIdOffset + dataCenterBitNum
	config.timestampOffset = config.dataCenterIdOffset + dataCenterBitNum

	config.maximumTimestamp = uint64(1 << timestampBitNum)
	config.maximumDataCenterId = uint64(1 << dataCenterBitNum)
	config.maximumServiceId = uint64(1 << serviceIdBitNum)
	config.maximumId = uint64(1 << (64 - timestampBitNum - dataCenterBitNum - serviceIdBitNum))

	generator := new(SnowflakeIdGenerator)
	generator.config = config

	generator.idLock = sync.Mutex{}
	generator.timestampLock = sync.Mutex{}
	generator.curTimestamp = uint64(time.Now().UnixMilli())
	generator.curId = uint64(0)
	// TODO: read data center id and service id from config file
	generator.dataCenterId = uint64(0)
	generator.serviceId = uint64(0)
	// END TODO

	return generator, nil
}

// Generate generates a new identiry
func (sig *SnowflakeIdGenerator) Generate() (Identity, error) {
	sig.idLock.Lock()
	defer sig.idLock.Unlock()
	sig.timestampLock.Lock()
	defer sig.timestampLock.Unlock()

	// check clock
	now := uint64(time.Now().UnixMilli())
	if sig.curTimestamp > now {
		return nil, fmt.Errorf("center clock is moving backwards")
	}

	if sig.curTimestamp == now {
		sig.curId++
		// check if current id is overflow
		if sig.curId >= sig.config.maximumId {
			// if overflow
			sig.curId = 0
			// wait for the next millisecond
			for now == sig.curTimestamp {
				sig.timestampLock.Unlock()
				runtime.Gosched()
				sig.timestampLock.Lock()
			}
		}
	} else {
		sig.curTimestamp = now
		sig.curId = 0
	}

	passedTime := sig.curTimestamp - systemStartTime
	if passedTime > sig.config.maximumTimestamp {
		return nil, fmt.Errorf("center clock failure")
	}

	snowflakeIdentity := uint64(0)
	snowflakeIdentity |= (passedTime << sig.config.timestampOffset)
	snowflakeIdentity |= (sig.dataCenterId << sig.config.dataCenterIdOffset)
	snowflakeIdentity |= (sig.serviceId << sig.config.serviceIdOffset)
	snowflakeIdentity |= sig.curId

	return SnowflakeIdentity(snowflakeIdentity), nil
}

func (si SnowflakeIdentity) Value() uint64 {
	return uint64(si)
}

func (si SnowflakeIdentity) String() string {
	return fmt.Sprintf("%d", uint64(si))
}

// checkBitNum returns a bool value represents
// if bitNum in range [0, 64]
func checkBitNum(bitNum int) bool {
	return (0 <= bitNum && bitNum <= 64)
}
