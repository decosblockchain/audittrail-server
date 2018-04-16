package integrity

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/cbergoon/merkletree"
	"github.com/decosblockchain/audittrail-server/commit"
	"github.com/decosblockchain/audittrail-server/logging"
	"github.com/decosblockchain/audittrail-server/wallet"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func CheckIntegrity() (bool, error) {
	rpc, err := wallet.GetRpcClient()
	if err != nil {
		return false, err
	}

	offset := 0
	myAddress, err := wallet.GetAddress()
	if err != nil {
		return false, err
	}

	myAddressString := myAddress.EncodeAddress()

	logging.Info.Printf("Finding transactions matching my address [%s]\n", myAddressString)

	for true {

		transactions, err := rpc.ListTransactionsCountFromWatchOnly("", 50, offset, true)
		if err != nil {
			return false, err
		}
		if len(transactions) == 0 {
			break
		}

		for _, t := range transactions {
			if t.Address == myAddressString {

				// Found transaction!
				result, err := CheckTransaction(t.TxID)
				if err != nil {
					return false, err
				}
				if !result {
					return false, nil
				}
			}
		}

		if len(transactions) == 50 {
			offset += 50
		} else {
			break
		}
	}

	return true, nil
}

func CheckTransaction(txHash string) (bool, error) {
	logging.Info.Printf("Checking transaction [%s]\n", txHash)

	rpc, err := wallet.GetRpcClient()
	if err != nil {
		return false, err
	}

	hash, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		return false, err
	}

	t, err := rpc.GetRawTransactionVerbose(hash)
	if err != nil {
		return false, err
	}

	for _, vo := range t.Vout {
		logging.Info.Printf("Found VOut Value [%f] type [%s]\n", vo.Value, vo.ScriptPubKey.Type)
		if vo.Value == float64(0) && vo.ScriptPubKey.Type == "nulldata" {
			b, err := hex.DecodeString(vo.ScriptPubKey.Hex)
			if err != nil {
				return false, err
			}
			result, err := CheckCommitTransaction(b)
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil
			}
		}
	}

	// TX does not contain a commit, ignore
	return true, nil
}

func CheckCommitTransaction(scriptPubKey []byte) (bool, error) {
	buf := bytes.NewBuffer(scriptPubKey)

	// First two bytes should be 6A and length (8 byte uint64 (startblock) - 8 byte uint64 (endblock) - 32 byte hash (merkle root) = 32+8+8 = 48 = (hex) 30 )
	if !bytes.Equal(buf.Next(2), []byte{0x6A, 0x30}) {
		// This is not a root commit nulldata, ignore
		return true, nil
	}

	var startBlock, endBlock uint64
	var merkleRoot [32]byte

	err := binary.Read(buf, binary.BigEndian, &startBlock)
	if err != nil {
		// Nulldata is somehow the same size, but does not contain a valid uint64; ignore
		return true, nil
	}
	err = binary.Read(buf, binary.BigEndian, &endBlock)
	if err != nil {
		// Nulldata is somehow the same size, but does not contain a valid uint64; ignore
		return true, nil
	}

	copy(merkleRoot[:], buf.Next(32))

	blockHashes, err := commit.GetBlockHashes(startBlock, endBlock)
	if err != nil {
		return false, err
	}
	t, _ := merkletree.NewTree(blockHashes)
	mr := t.MerkleRoot()

	ok := bytes.Equal(merkleRoot[:], mr)
	logging.Info.Printf("Integrity of blocks [%d - %d] OK\n", startBlock, endBlock)
	return ok, nil
}
