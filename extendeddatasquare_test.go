package rsmt2d

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const ShardSize = 64

var (
	zeros     = bytes.Repeat([]byte{0}, ShardSize)
	ones      = bytes.Repeat([]byte{1}, ShardSize)
	twos      = bytes.Repeat([]byte{2}, ShardSize)
	threes    = bytes.Repeat([]byte{3}, ShardSize)
	fours     = bytes.Repeat([]byte{4}, ShardSize)
	fives     = bytes.Repeat([]byte{5}, ShardSize)
	eights    = bytes.Repeat([]byte{8}, ShardSize)
	elevens   = bytes.Repeat([]byte{11}, ShardSize)
	thirteens = bytes.Repeat([]byte{13}, ShardSize)
	fifteens  = bytes.Repeat([]byte{15}, ShardSize)
)

func TestComputeExtendedDataSquare(t *testing.T) {
	codec := NewLeoRSCodec()

	type testCase struct {
		name string
		data [][]byte
		want [][][]byte
	}
	testCases := []testCase{
		{
			name: "1x1",
			// NOTE: data must contain byte slices that are a multiple of 64
			// bytes.
			// See https://github.com/catid/leopard/blob/22ddc7804998d31c8f1a2617ee720e063b1fa6cd/README.md?plain=1#L27
			// See https://github.com/klauspost/reedsolomon/blob/fd3e6910a7e457563469172968f456ad9b7696b6/README.md?plain=1#L403
			data: [][]byte{ones},
			want: [][][]byte{
				{ones, ones},
				{ones, ones},
			},
		},
		{
			name: "2x2",
			data: [][]byte{
				ones, twos,
				threes, fours,
			},
			want: [][][]byte{
				{ones, twos, zeros, threes},
				{threes, fours, eights, fifteens},
				{twos, elevens, thirteens, fours},
				{zeros, thirteens, fives, eights},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ComputeExtendedDataSquare(tc.data, codec, NewDefaultTree)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, result.squareRow)
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	codec := NewLeoRSCodec()
	result, err := ComputeExtendedDataSquare([][]byte{
		ones, twos,
		threes, fours,
	}, codec, NewDefaultTree)
	if err != nil {
		panic(err)
	}

	edsBytes, err := json.Marshal(result)
	if err != nil {
		t.Errorf("failed to marshal EDS: %v", err)
	}

	var eds ExtendedDataSquare
	err = json.Unmarshal(edsBytes, &eds)
	if err != nil {
		t.Errorf("failed to marshal EDS: %v", err)
	}
	if !reflect.DeepEqual(result.squareRow, eds.squareRow) {
		t.Errorf("eds not equal after json marshal/unmarshal")
	}
}

func TestImmutableRoots(t *testing.T) {
	codec := NewLeoRSCodec()
	result, err := ComputeExtendedDataSquare([][]byte{
		ones, twos,
		threes, fours,
	}, codec, NewDefaultTree)
	if err != nil {
		panic(err)
	}

	mutatedRowRoots, err := result.RowRoots()
	assert.NoError(t, err)

	mutatedRowRoots[0][0]++ // mutate

	rowRoots, err := result.RowRoots()
	assert.NoError(t, err)

	if reflect.DeepEqual(mutatedRowRoots, rowRoots) {
		t.Errorf("Exported EDS RowRoots was mutable")
	}

	mutatedColRoots, err := result.ColRoots()
	assert.NoError(t, err)

	mutatedColRoots[0][0]++ // mutate

	colRoots, err := result.ColRoots()
	assert.NoError(t, err)

	if reflect.DeepEqual(mutatedColRoots, colRoots) {
		t.Errorf("Exported EDS ColRoots was mutable")
	}
}

func TestEDSRowColImmutable(t *testing.T) {
	codec := NewLeoRSCodec()
	result, err := ComputeExtendedDataSquare([][]byte{
		ones, twos,
		threes, fours,
	}, codec, NewDefaultTree)
	if err != nil {
		panic(err)
	}

	row := result.Row(0)
	row[0][0]++
	if reflect.DeepEqual(row, result.Row(0)) {
		t.Errorf("Exported EDS Row was mutable")
	}

	col := result.Col(0)
	col[0][0]++
	if reflect.DeepEqual(col, result.Col(0)) {
		t.Errorf("Exported EDS Col was mutable")
	}
}

// dump acts as a data dump for the benchmarks to stop the compiler from making
// unrealistic optimizations
var dump *ExtendedDataSquare

// BenchmarkExtension benchmarks extending datasquares sizes 4-128 using all
// supported codecs (encoding only)
func BenchmarkExtensionEncoding(b *testing.B) {
	for i := 4; i < 513; i *= 2 {
		for codecName, codec := range codecs {
			if codec.MaxChunks() < i*i {
				// Only test codecs that support this many chunks
				continue
			}

			square := genRandDS(i)
			b.Run(
				fmt.Sprintf("%s %dx%dx%d ODS", codecName, i, i, len(square[0])),
				func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						eds, err := ComputeExtendedDataSquare(square, codec, NewDefaultTree)
						if err != nil {
							b.Error(err)
						}
						dump = eds
					}
				},
			)
		}
	}
}

// BenchmarkExtension benchmarks extending datasquares sizes 4-128 using all
// supported codecs (both encoding and root computation)
func BenchmarkExtensionWithRoots(b *testing.B) {
	for i := 4; i < 513; i *= 2 {
		for codecName, codec := range codecs {
			if codec.MaxChunks() < i*i {
				// Only test codecs that support this many chunks
				continue
			}

			square := genRandDS(i)
			b.Run(
				fmt.Sprintf("%s %dx%dx%d ODS", codecName, i, i, len(square[0])),
				func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						eds, err := ComputeExtendedDataSquare(square, codec, NewDefaultTree)
						if err != nil {
							b.Error(err)
						}
						_, _ = eds.RowRoots()
						_, _ = eds.ColRoots()
						dump = eds
					}
				},
			)
		}
	}
}

// genRandDS make a datasquare of random data, with width describing the number
// of shares on a single side of the ds
func genRandDS(width int) [][]byte {
	var ds [][]byte
	count := width * width
	for i := 0; i < count; i++ {
		share := make([]byte, 256)
		_, err := rand.Read(share)
		if err != nil {
			panic(err)
		}
		ds = append(ds, share)
	}
	return ds
}
