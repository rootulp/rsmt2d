package rsmt2d

import (
	"fmt"

	"github.com/klauspost/reedsolomon"
)

var _ Codec = &kpGF8Codec{}

func init() {
	registerCodec(KPGF8, NewKpGF8Codec())
}

type kpGF8Codec struct {
	encCache map[int]reedsolomon.Encoder
}

func NewKpGF8Codec() *kpGF8Codec {
	return &kpGF8Codec{make(map[int]reedsolomon.Encoder)}
}

func (k kpGF8Codec) Encode(data [][]byte) ([][]byte, error) {
	l0 := len(data[0])
	var enc reedsolomon.Encoder
	var err error
	if value, ok := k.encCache[len(data)]; ok {
		enc = value
	} else {
		enc, err = reedsolomon.New(len(data), len(data))
		if err != nil {
			return nil, err
		}
		k.encCache[len(data)] = enc
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
	var enc reedsolomon.Encoder
	var err error
	if value, ok := k.encCache[origLen]; ok {
		enc = value
	} else {
		enc, err = reedsolomon.New(origLen, origLen)
		if err != nil {
			return nil, err
		}
		k.encCache[len(data)] = enc
	}
	if err := enc.Reconstruct(data); err != nil {
		return nil, err
	}
	return data[0:origLen], nil

}

func (k kpGF8Codec) maxChunks() int {
	return 128 * 128
}
