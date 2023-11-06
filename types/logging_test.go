package types

import (
	"testing"

	"github.com/aergoio/aergo-lib/log"
	"github.com/aergoio/aergo/v2/internal/enc/base58"
	"github.com/rs/zerolog"
)

func BenchmarkLogMemAllocationCompared(b *testing.B) {
	type fields struct {
		Bytes *[]byte
	}
	logger := log.NewLogger("benchmark.logger")

	for i := 0; i < b.N; i++ {
		sampleBytes := "sample"
		logger.Warn().Int("idx", i).Str("var", sampleBytes).Msg("bench log")
	}
}

func BenchmarkLogMemAllocation(b *testing.B) {
	type fields struct {
		Bytes *[]byte
	}
	logger := log.NewLogger("benchmark.logger")

	for i := 0; i < b.N; i++ {
		sampleBytes := []byte("sample")
		logger.Warn().Int("idx", i).Str("var", EncodeB58(sampleBytes)).Msg("bench log")
	}
}
func BenchmarkLogMemAllocationD(b *testing.B) {
	type fields struct {
		Bytes *[]byte
	}
	logger := log.NewLogger("benchmark.logger")

	for i := 0; i < b.N; i++ {
		sampleBytes := []byte("sample")
		logger.Debug().Int("idx", i).Str("var", EncodeB58(sampleBytes)).Msg("bench log")
	}
}

func BenchmarkLogMemAllocationRun(b *testing.B) {
	type fields struct {
		Bytes *[]byte
	}
	logger := log.NewLogger("benchmark.logger")

	for i := 0; i < b.N; i++ {
		sampleBytes := []byte("sample")
		logger.Warn().Int("idx", i).Stringer("var", LogBase58(sampleBytes)).Msg("bench log")
	}
}

func BenchmarkLogMemAllocationRunD(b *testing.B) {
	type fields struct {
		Bytes *[]byte
	}
	logger := log.NewLogger("benchmark.logger")

	for i := 0; i < b.N; i++ {
		sampleBytes := []byte("sample")
		logger.Debug().Int("idx", i).Stringer("var", LogBase58(sampleBytes)).Msg("bench log")
	}
}

type LogB58Wrapper []byte

func (t LogB58Wrapper) MarshalZerologObject(e *zerolog.Event) {
	e.Str("b58", base58.Encode(t))
}

func BenchmarkLogMemAllocationWrapper(b *testing.B) {
	type fields struct {
		Bytes *[]byte
	}
	logger := log.NewLogger("benchmark.logger")

	for i := 0; i < b.N; i++ {
		sampleBytes := []byte("sample")
		logger.Warn().Int("idx", i).Object("var", LogB58Wrapper(sampleBytes)).Msg("bench log")
	}
}

func BenchmarkLogMemAllocationWrapperD(b *testing.B) {
	type fields struct {
		Bytes *[]byte
	}
	logger := log.NewLogger("benchmark.logger")

	for i := 0; i < b.N; i++ {
		sampleBytes := []byte("sample")
		logger.Debug().Int("idx", i).Object("var", LogB58Wrapper(sampleBytes)).Msg("bench log")
	}
}
