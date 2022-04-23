package rsmt2d

import (
	"fmt"

	"github.com/klauspost/reedsolomon"
)

var _ Codec = &kpGF8Codec{}

func init() {
	registerCodec(KPGF8, NewKpGF8Codec())
}

type kpGF8Codec struct{}

func NewKpGF8Codec() *kpGF8Codec {
	return &kpGF8Codec{}
}

func (k kpGF8Codec) Encode(data [][]byte) ([][]byte, error) {
	// TODO: make sure we re-use these instead of re-initializing:
	l0 := len(data[0])
	enc, err := reedsolomon.New(len(data), len(data))
	if err != nil {
		return nil, err
	}
	res := make([][]byte, len(data)*2)
	copy(res, data)
	for i := len(data); i < len(res); i++ {
		res[i] = make([]byte, l0)
	}
	if err := enc.Encode(res); err != nil {
		return nil, err
	}
	return res, nil

}

func (k kpGF8Codec) Decode(data [][]byte) ([][]byte, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("expected even length data, got: %v", len(data))
	}
	origLen := len(data) / 2
	// TODO: make sure we re-use these instead of re-initializing:
	enc, err := reedsolomon.New(origLen, origLen)
	if err != nil {
		return nil, err
	}
	if err := enc.Reconstruct(data); err != nil {
		return nil, err
	}
	return data[0:origLen], nil

}

func (k kpGF8Codec) maxChunks() int {
	return 128 * 128
}
