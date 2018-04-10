package commit

import (
	"encoding/hex"

	"github.com/cbergoon/merkletree"
)

type BlockHashContent struct {
	BlockHash string
}

func (b BlockHashContent) CalculateHash() []byte {
	bytes, _ := hex.DecodeString(b.BlockHash)
	return bytes
}

//Equals tests for equality of two Contents
func (b BlockHashContent) Equals(other merkletree.Content) bool {
	return b.BlockHash == other.(BlockHashContent).BlockHash
}
