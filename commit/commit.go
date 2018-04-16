package commit

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"sync"

	"github.com/decosblockchain/audittrail-server/config"
	"github.com/decosblockchain/audittrail-server/logging"
	"github.com/decosblockchain/audittrail-server/wallet"

	"github.com/cbergoon/merkletree"
	"github.com/onrik/ethrpc"
)

var committing = false

var committingMutex = &sync.Mutex{}

func readLastCommit() (uint64, error) {
	lastCommit := uint64(0)
	if _, err := os.Stat("data/lastcommit.hex"); os.IsNotExist(err) {
		return lastCommit, nil
	}
	b, err := ioutil.ReadFile("data/lastcommit.hex")
	if err != nil {
		return lastCommit, err
	}

	lastCommit = binary.LittleEndian.Uint64(b)
	return lastCommit, nil
}

func writeLastCommit(lastCommit uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, lastCommit)

	err := ioutil.WriteFile("data/lastcommit.hex", b, 0600)
	return err
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func CommitToMasterChain() error {
	committingMutex.Lock()
	if committing {
		committingMutex.Unlock()
		return nil
	}

	committing = true
	logging.Info.Println("Checking if i need to commit to master chain...")

	committingMutex.Unlock()

	lastCommit, err := readLastCommit()
	if err != nil {
		committingMutex.Lock()
		committing = false
		committingMutex.Unlock()
		return err
	}

	client := ethrpc.New(config.EthNode())
	block, err := client.EthBlockNumber()
	if err != nil {
		committingMutex.Lock()
		committing = false
		committingMutex.Unlock()
		return err
	}
	endBlock := uint64(block) - 30 // Due to possible reorgs, only commit blocks that have 30+ confirmations

	logging.Info.Printf("Currently at block %d, last commit was at %d\n", endBlock+30, lastCommit)
	interval := config.CommitInterval()

	if (endBlock - lastCommit) >= interval {
		endBlock := min(endBlock, lastCommit+interval)
		logging.Info.Println("Committing to master chain")

		blockHashes, err := GetBlockHashes(lastCommit+1, endBlock)
		t, _ := merkletree.NewTree(blockHashes)
		mr := t.MerkleRoot()

		logging.Info.Printf("Merkle root of block segment: %x\n", mr)
		err = commitMerkleRoot(lastCommit+1, endBlock, mr)
		if err != nil {
			committingMutex.Lock()
			committing = false
			committingMutex.Unlock()
			return err
		}

		writeLastCommit(endBlock)

	} else {
		logging.Info.Println("No need to commit to master chain (yet)")
	}

	committingMutex.Lock()
	committing = false
	committingMutex.Unlock()

	return nil
}

func commitMerkleRoot(startBlock, endBlock uint64, merkleRoot []byte) error {
	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, startBlock)
	binary.Write(&buf, binary.BigEndian, endBlock)
	buf.Write(merkleRoot[:])

	tx, err := wallet.CreateNullDataTransaction(buf.Bytes())
	if err != nil {
		return err
	}
	signedTx, err := wallet.SignTransaction(tx)
	if err != nil {
		return err
	}

	hash, err := wallet.SendTransaction(signedTx)
	if err != nil {
		return err
	}

	logging.Info.Printf("Committed to master chain succesfully. TX hash: %s", hash.String())

	return nil
}

func GetBlockHashes(startBlock, endBlock uint64) ([]merkletree.Content, error) {
	client := ethrpc.New(config.EthNode())

	blockHashes := []merkletree.Content{}
	for i := startBlock; i <= endBlock; i++ {
		block, err := client.EthGetBlockByNumber(int(i), false)
		if err != nil {
			committingMutex.Lock()
			committing = false
			committingMutex.Unlock()
			return blockHashes, err
		}
		blockHashes = append(blockHashes, BlockHashContent{BlockHash: block.Hash})
	}
	return blockHashes, nil
}
