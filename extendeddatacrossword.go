package rsmt2d

import (
	"bytes"
	"fmt"
	"math"
)

const (
	row    = 0
	column = 1
)

// ByzantineRowError is thrown when there is a repaired row does not match the expected row merkle root.
type ByzantineRowError struct {
	RowNumber uint
}

func (e *ByzantineRowError) Error() string {
	return fmt.Sprintf("byzantine row: %d", e.RowNumber)
}

// ByzantineColumnError is thrown when there is a repaired column does not match the expected column merkle root.
type ByzantineColumnError struct {
	ColumnNumber uint
}

func (e *ByzantineColumnError) Error() string {
	return fmt.Sprintf("byzantine column: %d", e.ColumnNumber)
}

// UnrepairableDataSquareError is thrown when there is insufficient chunks to repair the square.
type UnrepairableDataSquareError struct {
}

func (e *UnrepairableDataSquareError) Error() string {
	return "failed to solve data square"
}

// RepairExtendedDataSquare repairs an incomplete extended data square, against its expected row and column merkle roots.
// Missing data chunks should be represented as nil.
func RepairExtendedDataSquare(
	rowRoots [][]byte,
	columnRoots [][]byte,
	data [][]byte,
	codec CodecType,
	treeCreatorFn TreeConstructorFn,
) (*ExtendedDataSquare, error) {
	width := int(math.Ceil(math.Sqrt(float64(len(data)))))
	bitMask := NewBitMatrix(width, len(data))
	var chunkSize int
	for i := range data {
		if data[i] != nil {
			bitMask.SetFlat(i)
			if chunkSize == 0 {
				chunkSize = len(data[i])
			}
		}
	}

	if chunkSize == 0 {
		return nil, &UnrepairableDataSquareError{}
	}

	fillerChunk := bytes.Repeat([]byte{0}, chunkSize)
	for i := range data {
		if data[i] == nil {
			data[i] = make([]byte, chunkSize)
			copy(data[i], fillerChunk)
		}
	}

	eds, err := ImportExtendedDataSquare(data, codec, treeCreatorFn)
	if err != nil {
		return nil, err
	}

	err = eds.prerepairSanityCheck(rowRoots, columnRoots, bitMask)
	if err != nil {
		return nil, err
	}

	err = eds.solveCrossword(rowRoots, columnRoots, bitMask)
	if err != nil {
		return nil, err
	}

	return eds, err
}

func (eds *ExtendedDataSquare) solveCrossword(rowRoots [][]byte, columnRoots [][]byte, bitMask bitMatrix) error {
	// Keep repeating until the square is solved
	solved := false
	for !solved {
		solved = true
		progressMade := false

		// Loop through every row and column, attempt to rebuild each row or column if incomplete
		for i := 0; i < int(eds.width); i++ {
			for mode := range []int{row, column} {

				var isIncomplete bool
				var isExtendedPartIncomplete bool
				if mode == row {
					isIncomplete = !bitMask.RowIsOne(i)
					isExtendedPartIncomplete = !bitMask.RowRangeIsOne(i, int(eds.originalDataWidth), int(eds.width))
				} else if mode == column {
					isIncomplete = !bitMask.ColumnIsOne(i)
					isExtendedPartIncomplete = !bitMask.ColRangeIsOne(i, int(eds.originalDataWidth), int(eds.width))
				}

				if isIncomplete { // row/column incomplete
					// Prepare shares
					var vectorData [][]byte
					shares := make([][]byte, eds.width)
					for j := 0; j < int(eds.width); j++ {
						var rowIdx, colIdx int
						if mode == row {
							rowIdx = i
							colIdx = j
							vectorData = eds.Row(uint(i))
						} else if mode == column {
							rowIdx = j
							colIdx = i
							vectorData = eds.Column(uint(i))
						}
						if bitMask.Get(rowIdx, colIdx) {
							shares[j] = vectorData[j]
						}
					}

					// Attempt rebuild
					rebuiltShares, err := Decode(shares, eds.codec)
					if err != nil { // repair unsuccessful
						solved = false
					} else { // repair successful
						progressMade = true
						// Insert rebuilt shares into square
						for p, s := range rebuiltShares {
							if mode == row {
								eds.setCell(uint(i), uint(p), s)
							} else if mode == column {
								eds.setCell(uint(p), uint(i), s)
							}
						}
						if isExtendedPartIncomplete {
							err2 := rebuildExtendedPart(eds, mode, uint(i))
							if err2 != nil {
								return err2
							}
						}

						err3 := verifyRoots(rowRoots, columnRoots, mode, eds, uint(i))
						if err3 != nil {
							return err3
						}

						// Check that newly completed orthogonal vectors match their new merkle roots
						for j := 0; j < int(eds.width); j++ {
							if mode == row && !bitMask.Get(i, j) {
								if bitMask.ColumnIsOne(j) && !bytes.Equal(eds.ColRoot(uint(j)), columnRoots[j]) {
									return &ByzantineColumnError{uint(j)}
								}
							} else if mode == column && !bitMask.Get(j, i) {
								if bitMask.RowIsOne(j) && !bytes.Equal(eds.RowRoot(uint(j)), rowRoots[j]) {
									return &ByzantineRowError{uint(j)}
								}
							}
						}

						// Set vector mask to true
						if mode == row {
							for j := 0; j < int(eds.width); j++ {
								bitMask.Set(i, j)
							}
						} else if mode == column {
							for j := 0; j < int(eds.width); j++ {
								bitMask.Set(j, i)
							}
						}
					}
				}
			}
		}

		if !progressMade {
			return &UnrepairableDataSquareError{}
		}
	}

	return nil
}

func verifyRoots(rowRoots [][]byte, columnRoots [][]byte, mode int, eds *ExtendedDataSquare, i uint) error {
	// Check that rebuilt vector matches given merkle root
	if mode == row {
		if !bytes.Equal(eds.RowRoot(i), rowRoots[i]) {
			return &ByzantineRowError{i}
		}
	} else if mode == column {
		if !bytes.Equal(eds.ColRoot(i), columnRoots[i]) {
			return &ByzantineColumnError{i}
		}
	}
	return nil
}

func rebuildExtendedPart(eds *ExtendedDataSquare, mode int, rowOrColIdx uint) error {
	var data [][]byte
	if mode == row {
		data = eds.rowSlice(rowOrColIdx, 0, eds.originalDataWidth)
	} else if mode == column {
		data = eds.columnSlice(0, rowOrColIdx, eds.originalDataWidth)
	}
	rebuiltExtendedShares, err := Encode(data, eds.codec)
	if err != nil {
		return err
	}
	for p, s := range rebuiltExtendedShares {
		if mode == row {
			eds.setCell(rowOrColIdx, eds.originalDataWidth+uint(p), s)
		} else if mode == column {
			eds.setCell(eds.originalDataWidth+uint(p), rowOrColIdx, s)
		}
	}

	return nil
}

func (eds *ExtendedDataSquare) prerepairSanityCheck(rowRoots [][]byte, columnRoots [][]byte, bitMask bitMatrix) error {
	var shares [][]byte
	var err error
	for i := uint(0); i < eds.width; i++ {
		rowIsComplete := bitMask.RowIsOne(int(i))
		colIsComplete := bitMask.ColumnIsOne(int(i))

		// if there's no missing data in the this row
		if noMissingData(eds.Row(i)) {
			// ensure that the roots are equal and that rowMask is a vector
			if rowIsComplete && !bytes.Equal(rowRoots[i], eds.RowRoot(i)) {
				return fmt.Errorf("bad root input: row %d expected %v got %v", i, rowRoots[i], eds.RowRoot(i))
			}
		}

		// if there's no missing data in the this col
		if noMissingData(eds.Column(i)) {
			// ensure that the roots are equal and that rowMask is a vector
			if colIsComplete && !bytes.Equal(columnRoots[i], eds.ColRoot(i)) {
				return fmt.Errorf("bad root input: col %d expected %v got %v", i, columnRoots[i], eds.ColRoot(i))
			}
		}

		if rowIsComplete {
			shares, err = Encode(eds.rowSlice(i, 0, eds.originalDataWidth), eds.codec)
			if err != nil {
				return err
			}
			if !bytes.Equal(flattenChunks(shares), flattenChunks(eds.rowSlice(i, eds.originalDataWidth, eds.originalDataWidth))) {
				return &ByzantineRowError{i}
			}
		}

		if colIsComplete {
			shares, err = Encode(eds.columnSlice(0, i, eds.originalDataWidth), eds.codec)
			if err != nil {
				return err
			}
			if !bytes.Equal(flattenChunks(shares), flattenChunks(eds.columnSlice(eds.originalDataWidth, i, eds.originalDataWidth))) {
				return &ByzantineColumnError{i}
			}
		}
	}

	return nil
}

func noMissingData(input [][]byte) bool {
	for _, d := range input {
		if d == nil {
			return false
		}
	}
	return true
}
