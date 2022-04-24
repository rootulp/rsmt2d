package rsmt2d

import (
	"fmt"
	"math/rand"
	"testing"
)

var (
	encodedDataDump [][]byte
	decodedDataDump [][]byte
)

func BenchmarkEncoding(b *testing.B) {
	b.ReportAllocs()
	shares, shareSize := 128, 512
	// generate some fake data
	data := generateRandomChunkData(shares, shareSize)
	for codecName, codec := range codecs {
		b.Run(
			fmt.Sprintf("Encoding %v shares of size %v using %s", shares, shareSize, codecName),
			func(b *testing.B) {
				for n := 0; n < b.N; n++ {
					encodedData, err := codec.Encode(data)
					if err != nil {
						b.Error(err)
					}
					encodedDataDump = encodedData
				}
			},
		)
	}
}

func generateRandomChunkData(shares, shareSize int) [][]byte {
	out := make([][]byte, shares)
	for i := 0; i < shares; i++ {
		randData := make([]byte, shareSize)
		_, err := rand.Read(randData)
		if err != nil {
			panic(err)
		}
		out[i] = randData
	}
	return out
}

func BenchmarkDecoding(b *testing.B) {
	shares, shareSize := 128, 512
	b.ReportAllocs()
	// generate some fake data
	for codecName, codec := range codecs {
		data := generateMissingData(shares, shareSize, codec)
		b.Run(
			fmt.Sprintf("Decoding %v shares size %v using %v", shares, shareSize, codecName),
			func(b *testing.B) {
				for n := 0; n < b.N; n++ {
					encodedData, err := codec.Decode(data)
					if err != nil {
						b.Error(err)
					}
					encodedDataDump = encodedData
				}
			},
		)
	}
}

func generateMissingData(count, shareSize int, codec Codec) [][]byte {
	randData := generateRandomChunkData(count, shareSize)
	encoded, err := codec.Encode(randData)
	if err != nil {
		panic(err)
	}

	// remove half of the shares randomly
	for i := 0; i < (count / 2); {
		ind := rand.Intn(count)
		if len(encoded[ind]) == 0 {
			continue
		}
		encoded[ind] = nil
		i++
	}

	return encoded
}
