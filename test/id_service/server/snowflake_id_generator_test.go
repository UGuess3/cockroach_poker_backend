package id_server_test

import (
	"testing"

	server "github.com/frealcone/roach/service/id_service/server"
)

// TestNewSnowflakeIdGenerator tests constructor of snowflake
// ID generator.
func TestNewSnowflakeIdGenerator(t *testing.T) {
	_, err := server.NewSnowflakeIdGenerator(42, 5, 5)
	if err != nil {
		t.Fatalf("Failed to create generator: %s", err.Error())
	}
}

// Test method generate of SnowflakeIdGenerators
func TestGenerate(t *testing.T) {
	generator, err := server.NewSnowflakeIdGenerator(42, 5, 5)
	if err != nil {
		t.Fatalf("Failed to create generator: %s", err.Error())
	}

	round := 1000
	identities := make(map[uint64]struct{}, round)

	for i := 0; i < round; i++ {
		identity, err := generator.Generate()
		if err != nil {
			t.Fatalf("Failed to generate id: %v", err)
		}
		t.Logf("Generated identity: %s\n", identity.String())

		id := identity.Value()

		// 0b11111 -> 0x1F
		dataCenterMark := 0x1F
		expectedDataCenterId, err := server.ConfigGetUint64("data_center_id")
		if err != nil || ((id>>17)&uint64(dataCenterMark)) != expectedDataCenterId {
			t.Fatalf("Unexpected data center id: %d, correct id: %d",
				((id >> 17) & uint64(dataCenterMark)),
				expectedDataCenterId)
		}

		// 0b11111 -> 0x1F
		serviceMark := 0x1F
		expectedServiceId, err := server.ConfigGetUint64("service_id")
		if err != nil || ((id>>12)&uint64(serviceMark)) != expectedServiceId {
			t.Fatalf("Unexpected service id: %d, correct id: %d",
				((id >> 12) & uint64(serviceMark)),
				expectedServiceId)
		}

		if _, ok := identities[id]; ok {
			t.Fatalf("Repeated id: %d", id)
		} else {
			identities[id] = struct{}{}
			t.Log(identity.String() + "\n")
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	generator, err := server.NewSnowflakeIdGenerator(42, 5, 5)
	if err != nil {
		b.Fatalf("Failed to create generator: %s", err.Error())
	}

	identities := make(map[string]struct{})

	for i := 0; i < b.N; i++ {
		identity, err := generator.Generate()
		if err != nil {
			b.Fatal(err)
		}
		if _, ok := identities[identity.String()]; ok {
			b.Fatalf("duplicated identity: %s", identity.String())
		}
	}
}
