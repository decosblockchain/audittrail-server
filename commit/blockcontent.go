package commit

import (
	"encoding/hex"

	"github.com/cbergoon/merkletree"
	"github.com/decosblockchain/audittrail-server/logging"
)

type BlockHashContent struct {
	BlockHash string
}

func (b BlockHashContent) CalculateHash() []byte {
	bytes, err := hex.DecodeString(b.BlockHash[2:])
	if err != nil {
		logging.Error.Printf("Error decoding blockhash %s\n", b.BlockHash)
	}

	return bytes
}

//Equals tests for equality of two Contents
func (b BlockHashContent) Equals(other merkletree.Content) bool {
	return b.BlockHash == other.(BlockHashContent).BlockHash
}
