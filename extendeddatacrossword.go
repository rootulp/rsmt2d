package rsmt2d

import (
	"bytes"
	"errors"
	"fmt"

	"gonum.org/v1/gonum/mat"
)

// ErrUnrepairableDataSquare is thrown when there is insufficient chunks to repair the square.
var ErrUnrepairableDataSquare = errors.New("failed to solve data square")

// ErrByzantineRow is thrown when there is a repaired row does not match the expected row merkle root.
type ErrByzantineRow struct {
	RowNumber uint
}

func (e *ErrByzantineRow) Error() string {
	return fmt.Sprintf("byzantine row: %d", e.RowNumber)
}

// ErrByzantineColumn is thrown when there is a repaired column does not match the expected column merkle root.
type ErrByzantineColumn struct {
	ColumnNumber uint
}

func (e *ErrByzantineColumn) Error() string {
	return fmt.Sprintf("byzantine column: %d", e.ColumnNumber)
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
	matrixData := make([]float64, len(data))
	var chunkSize int
	for i := range data {
		if data[i] == nil {
			matrixData[i] = 0
		} else {
			matrixData[i] = 1
			if chunkSize == 0 {
				chunkSize = len(data[i])
			}
		}
	}

	if chunkSize == 0 {
		return nil, ErrUnrepairableDataSquare
	}

	for i := range data {
		if data[i] == nil {
			// fill with all zeroes
			data[i] = make([]byte, chunkSize)
		}
	}

	eds, err := ImportExtendedDataSquare(data, codec, treeCreatorFn)
	if err != nil {
		return nil, err
	}

	matrix := mat.NewDense(int(eds.width), int(eds.width), matrixData)

	err = eds.prerepairSanityCheck(rowRoots, columnRoots, matrix)
	if err != nil {
		return nil, err
	}

	err = eds.solveCrossword(rowRoots, columnRoots, matrix)
	if err != nil {
		return nil, err
	}

	return eds, err
}

func (eds *ExtendedDataSquare) solveCrossword(rowRoots [][]byte, columnRoots [][]byte, mask *mat.Dense) error {
	// Keep repeating until the square is solved
	solved := false
	for !solved {
		solved = true
		progressMade := false
		// Loop through every row and column, attempt to rebuild each row or column if incomplete
		for i := uint(0); i < eds.width; i++ {
			// row:
			vectorMask := mask.RowView(int(i))
			rowIsIncomplete := !vecIsTrue(vectorMask)
			if rowIsIncomplete {
				shares := eds.prepareShares(eds.Row(i), vectorMask)

				// Attempt rebuild
				rebuiltShares, err := Decode(shares, eds.codec)
				if err != nil { // repair unsuccessful
					solved = false
				} else { // repair successful
					progressMade = true
					err := eds.updateRows(rowRoots, columnRoots, mask, i, rebuiltShares, vectorMask)
					if err != nil {
						return err
					}

				}
			}

			// column:
			vectorMask = mask.ColView(int(i))
			isColumnIncomplete := !vecIsTrue(vectorMask)
			if isColumnIncomplete {
				shares := eds.prepareShares(eds.Column(i), vectorMask)

				// Attempt rebuild
				rebuiltShares, err := Decode(shares, eds.codec)
				if err != nil {
					// repair unsuccessful
					solved = false
				} else {
					progressMade = true
					err := eds.updateCol(rowRoots, columnRoots, mask, i, rebuiltShares, vectorMask)
					if err != nil {
						return err
					}
				}
			}
		}
		if !progressMade {
			return ErrUnrepairableDataSquare
		}
	}

	return nil
}

func (eds *ExtendedDataSquare) updateRows(rowRoots [][]byte, columnRoots [][]byte, mask *mat.Dense, i uint, rebuiltShares [][]byte, vectorMask mat.Vector) error {
	//  Insert rebuilt shares into square
	for p, s := range rebuiltShares {
		eds.setCell(i, uint(p), s)
	}

	// Rebuild extended part if incomplete
	if !vecSliceIsTrue(vectorMask, int(eds.originalDataWidth), int(eds.width)) {
		rebuiltExtendedShares, err := Encode(eds.rowSlice(i, 0, eds.originalDataWidth), eds.codec)
		if err != nil {
			return err
		}
		for p, s := range rebuiltExtendedShares {
			eds.setCell(i, eds.originalDataWidth+uint(p), s)
		}
	}
	// Check that rebuilt vector matches given merkle root
	if !bytes.Equal(eds.RowRoot(i), rowRoots[i]) {
		return &ErrByzantineRow{i}
	}

	// Check that newly completed orthogonal vectors match their new merkle roots
	for j := uint(0); j < eds.width; j++ {
		if vectorMask.AtVec(int(j)) == 0 {
			adjMask := mask.ColView(int(j))
			if vecNumTrue(adjMask) == adjMask.Len()-1 && !bytes.Equal(eds.ColRoot(j), columnRoots[j]) {
				return &ErrByzantineColumn{j}
			}
		}
	}
	// Set vector mask to true
	for j := 0; j < int(eds.width); j++ {
		mask.Set(int(i), j, 1)
	}
	return nil
}

func (eds *ExtendedDataSquare) updateCol(rowRoots [][]byte, columnRoots [][]byte, mask *mat.Dense, i uint, rebuiltShares [][]byte, vectorMask mat.Vector) error {
	// Insert rebuilt shares into square
	for p, s := range rebuiltShares {
		eds.setCell(uint(p), i, s)
	}

	// Rebuild extended part if incomplete
	if !vecSliceIsTrue(vectorMask, int(eds.originalDataWidth), int(eds.width)) {
		rebuiltExtendedShares, err := Encode(eds.columnSlice(0, i, eds.originalDataWidth), eds.codec)
		if err != nil {
			return err
		}
		for p, s := range rebuiltExtendedShares {
			eds.setCell(eds.originalDataWidth+uint(p), i, s)
		}
	}

	// Check that rebuilt vector matches given merkle root
	if !bytes.Equal(eds.ColRoot(i), columnRoots[i]) {
		return &ErrByzantineColumn{i}
	}

	// Check that newly completed orthogonal vectors match their new merkle roots
	for j := uint(0); j < eds.width; j++ {
		if vectorMask.AtVec(int(j)) == 0 {
			adjMask := mask.RowView(int(j))
			if vecNumTrue(adjMask) == adjMask.Len()-1 && !bytes.Equal(eds.RowRoot(j), rowRoots[j]) {
				return &ErrByzantineRow{j}
			}
		}
	}

	// Set vector mask to true
	for j := 0; j < int(eds.width); j++ {
		mask.Set(j, int(i), 1)
		// Set vector mask to true
		for j := 0; j < int(eds.width); j++ {
			mask.Set(j, int(i), 1)
		}
	}
	return nil
}

func (eds *ExtendedDataSquare) prepareShares(vectorData [][]byte, vectorMask mat.Vector) [][]byte {
	shares := make([][]byte, eds.width)
	for j := uint(0); j < eds.width; j++ {
		if vectorMask.AtVec(int(j)) == 1 {
			shares[j] = vectorData[j]
		}
	}
	return shares
}

func (eds *ExtendedDataSquare) prerepairSanityCheck(rowRoots [][]byte, columnRoots [][]byte, mask *mat.Dense) error {
	var shares [][]byte
	var err error
	for i := uint(0); i < eds.width; i++ {
		rowMask := mask.RowView(int(i))
		columnMask := mask.ColView(int(i))
		rowMaskIsVec := vecIsTrue(rowMask)
		columnMaskIsVec := vecIsTrue(columnMask)

		// if there's no missing data in the this row
		if noMissingData(eds.Row(i)) {
			// ensure that the roots are equal and that rowMask is a vector
			if rowMaskIsVec && !bytes.Equal(rowRoots[i], eds.RowRoot(i)) {
				return fmt.Errorf("bad root input: row %d expected %v got %v", i, rowRoots[i], eds.RowRoot(i))
			}
		}

		// if there's no missing data in the this col
		if noMissingData(eds.Column(i)) {
			// ensure that the roots are equal and that rowMask is a vector
			if columnMaskIsVec && !bytes.Equal(columnRoots[i], eds.ColRoot(i)) {
				return fmt.Errorf("bad root input: col %d expected %v got %v", i, columnRoots[i], eds.ColRoot(i))
			}
		}

		if rowMaskIsVec {
			shares, err = Encode(eds.rowSlice(i, 0, eds.originalDataWidth), eds.codec)
			if err != nil {
				return err
			}
			if !bytes.Equal(flattenChunks(shares), flattenChunks(eds.rowSlice(i, eds.originalDataWidth, eds.originalDataWidth))) {
				return &ErrByzantineRow{i}
			}
		}

		if columnMaskIsVec {
			shares, err = Encode(eds.columnSlice(0, i, eds.originalDataWidth), eds.codec)
			if err != nil {
				return err
			}
			if !bytes.Equal(flattenChunks(shares), flattenChunks(eds.columnSlice(eds.originalDataWidth, i, eds.originalDataWidth))) {
				return &ErrByzantineColumn{i}
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

func vecIsTrue(vec mat.Vector) bool {
	for i := 0; i < vec.Len(); i++ {
		if vec.AtVec(i) == 0 {
			return false
		}
	}

	return true
}

func vecSliceIsTrue(vec mat.Vector, start int, end int) bool {
	for i := start; i < end; i++ {
		if vec.AtVec(i) == 0 {
			return false
		}
	}

	return true
}

func vecNumTrue(vec mat.Vector) int {
	var counter int
	for i := 0; i < vec.Len(); i++ {
		if vec.AtVec(i) == 1 {
			counter++
		}
	}

	return counter
}
