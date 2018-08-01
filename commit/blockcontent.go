package commit

import (
	"encoding/hex"

	"github.com/cbergoon/merkletree"
)

type BlockHashContent struct {
	BlockHash string
}

func (b BlockHashContent) CalculateHash() ([]byte, error) {
	bytes, err := hex.DecodeString(b.BlockHash[2:])
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

//Equals tests for equality of two Contents
func (b BlockHashContent) Equals(other merkletree.Content) (bool, error) {
	return (b.BlockHash == other.(BlockHashContent).BlockHash), nil
}
