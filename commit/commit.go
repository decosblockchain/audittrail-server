package commit

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"sync"

	"github.com/decosblockchain/audittrail-server/config"
	"github.com/decosblockchain/audittrail-server/logging"

	"github.com/cbergoon/merkletree"
	"github.com/onrik/ethrpc"
)

var committing = false

var committingMutex = &sync.Mutex{}

func readLastCommit() (uint64, error) {
	lastCommit := uint64(0)
	if _, err := os.Stat("data/nonce.hex"); os.IsNotExist(err) {
		return lastCommit, nil
	}
	b, err := ioutil.ReadFile("data/nonce.hex")
	if err != nil {
		return lastCommit, err
	}

	lastCommit = binary.LittleEndian.Uint64(b)
	return lastCommit, nil
}

func writeLastCommit(lastCommit uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, lastCommit)

	err := ioutil.WriteFile("lastcommit.hex", b, 0600)
	return err
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
		return err
	}

	client := ethrpc.New(config.EthNode())
	block, err := client.EthBlockNumber()
	if err != nil {
		return err
	}
	currentBlock := uint64(block)
	logging.Info.Printf("Currently at block %d, last commit was at %d\n", currentBlock, lastCommit)

	if (currentBlock - lastCommit) >= config.CommitInterval() {
		blockHashes := []merkletree.Content{}
		for i := lastCommit + 1; i <= currentBlock; i++ {
			block, err := client.EthGetBlockByNumber(int(i), false)
			if err != nil {
				return err
			}
			blockHashes = append(blockHashes, BlockHashContent{BlockHash: block.Hash})
		}
		t, _ := merkletree.NewTree(blockHashes)
		mr := t.MerkleRoot()

		logging.Info.Printf("Merkle root of block segment: %x\n", mr)
		writeLastCommit(currentBlock)

	} else {
		logging.Info.Println("No need to commit to master chain (yet)")
	}

	committingMutex.Lock()
	committing = false
	committingMutex.Unlock()

	return nil
}
