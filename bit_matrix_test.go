package rsmt2d

import (
	"testing"
)

func Test_bitMatrix_ColRangeIsOne(t *testing.T) {
	type initParams struct {
		squareSize int
		bits       int
	}
	type args struct {
		c     int
		start int
		end   int
	}
	tests := []struct {
		name       string
		initParams initParams
		setBits    []int
		args       args
		want       bool
	}{
		{"none set", initParams{4, 16}, nil, args{0, 0, 4}, false},
		{"none set", initParams{4, 16}, nil, args{1, 0, 4}, false},
		{"none set", initParams{4, 16}, nil, args{2, 0, 4}, false},
		{"none set", initParams{4, 16}, nil, args{3, 0, 4}, false},
		{"none set in requested range", initParams{4, 16}, []int{0, 1, 2, 3, 4, 5, 6}, args{0, 0, 4}, false},
		{"none set in requested range", initParams{4, 16}, []int{0, 4}, args{0, 1, 3}, false},
		{"all set in requested range", initParams{4, 16}, []int{0, 4, 8}, args{0, 0, 1}, true},
		{"all set in requested range", initParams{4, 16}, []int{0, 4, 8}, args{0, 0, 2}, true},
		{"all set in requested range", initParams{4, 16}, []int{0, 4, 8, 16}, args{0, 0, 1}, true},
		{"all set in requested range", initParams{4, 16}, []int{0, 4, 8, 16}, args{0, 0, 2}, true},
		{"all set in requested range", initParams{4, 16}, []int{0, 4, 8, 12}, args{0, 0, 3}, true},
		{"all set in requested range", initParams{4, 16}, []int{1, 5, 9, 13}, args{1, 0, 3}, true},
		{"all set in requested range", initParams{4, 16}, []int{2, 6, 10, 14}, args{2, 0, 3}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := newBitMatrix(tt.initParams.squareSize, tt.initParams.bits)
			for _, flatIdx := range tt.setBits {
				bm.SetFlat(flatIdx)
			}
			if got := bm.ColRangeIsOne(tt.args.c, tt.args.start, tt.args.end); got != tt.want {
				t.Errorf("ColRangeIsOne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitMatrix_ColumnIsOne(t *testing.T) {
	type initParams struct {
		squareSize int
		bits       int
	}

	type args struct {
		c int
	}
	tests := []struct {
		name       string
		initParams initParams
		setBits    []int
		args       args
		want       bool
	}{
		{"col_0 == 1", initParams{4, 16}, []int{0, 4, 8, 12}, args{0}, true},
		{"col_1 != 1", initParams{4, 16}, []int{0, 4, 8, 16}, args{1}, false},
		{"col_2 != 1", initParams{4, 16}, []int{0, 4, 8, 16}, args{2}, false},
		{"col_3 != 1", initParams{4, 16}, []int{0, 4, 8, 16}, args{3}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := newBitMatrix(tt.initParams.squareSize, tt.initParams.bits)
			for _, flatIdx := range tt.setBits {
				bm.SetFlat(flatIdx)
			}
			if got := bm.ColumnIsOne(tt.args.c); got != tt.want {
				t.Errorf("ColumnIsOne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitMatrix_Get(t *testing.T) {
	type initParams struct {
		squareSize int
		bits       int
	}
	type args struct {
		row int
		col int
	}
	tests := []struct {
		name       string
		initParams initParams
		args       args
		want       bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := newBitMatrix(tt.initParams.squareSize, tt.initParams.bits)
			if got := bm.Get(tt.args.row, tt.args.col); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitMatrix_NumOnesInCol(t *testing.T) {
	type initParams struct {
		squareSize int
		bits       int
	}
	type args struct {
		c int
	}
	tests := []struct {
		name       string
		initParams initParams
		args       args
		want       int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := newBitMatrix(tt.initParams.squareSize, tt.initParams.bits)
			if got := bm.NumOnesInCol(tt.args.c); got != tt.want {
				t.Errorf("NumOnesInCol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitMatrix_NumOnesInRow(t *testing.T) {
	type initParams struct {
		squareSize int
		bits       int
	}
	type args struct {
		r int
	}
	tests := []struct {
		name       string
		initParams initParams
		args       args
		want       int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := newBitMatrix(tt.initParams.squareSize, tt.initParams.bits)
			if got := bm.NumOnesInRow(tt.args.r); got != tt.want {
				t.Errorf("NumOnesInRow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitMatrix_RowIsOne(t *testing.T) {
	type initParams struct {
		squareSize int
		bits       int
	}
	type args struct {
		r int
	}
	tests := []struct {
		name       string
		initParams initParams
		args       args
		want       bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := newBitMatrix(tt.initParams.squareSize, tt.initParams.bits)
			if got := bm.RowIsOne(tt.args.r); got != tt.want {
				t.Errorf("RowIsOne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitMatrix_RowRangeIsOne(t *testing.T) {
	type initParams struct {
		squareSize int
		bits       int
	}
	type args struct {
		r     int
		start int
		end   int
	}
	tests := []struct {
		name       string
		initParams initParams
		args       args
		want       bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := newBitMatrix(tt.initParams.squareSize, tt.initParams.bits)
			if got := bm.RowRangeIsOne(tt.args.r, tt.args.start, tt.args.end); got != tt.want {
				t.Errorf("RowRangeIsOne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitMatrix_Set(t *testing.T) {
	type initParams struct {
		squareSize int
		bits       int
	}
	type args struct {
		row int
		col int
	}
	tests := []struct {
		name       string
		initParams initParams
		args       args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := newBitMatrix(tt.initParams.squareSize, tt.initParams.bits)
			bm.SetFlat(0)
		})
	}
}

func Test_bitMatrix_SetFlat(t *testing.T) {
	type initParams struct {
		squareSize int
		bits       int
	}
	type args struct {
		i int
	}
	tests := []struct {
		name       string
		initParams initParams
		args       args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := newBitMatrix(tt.initParams.squareSize, tt.initParams.bits)
			bm.SetFlat(0)
		})

	}
}
