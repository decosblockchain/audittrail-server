package wallet

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"sync"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/decosblockchain/audittrail-server/config"
	"github.com/decosblockchain/audittrail-server/logging"
)

var keyMutex = &sync.Mutex{}

func getPrivateKey() (*btcec.PrivateKey, error) {
	keyMutex.Lock()

	if _, err := os.Stat("data/keyfile.hex"); os.IsNotExist(err) {
		generatedKey, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			keyMutex.Unlock()
			return nil, err
		}
		err = ioutil.WriteFile("data/keyfile.hex", generatedKey.Serialize(), 0600)
		if err != nil {
			keyMutex.Unlock()
			return nil, err
		}
	}

	privateKeyBytes, err := ioutil.ReadFile("data/keyfile.hex")
	if err != nil {
		keyMutex.Unlock()
		return nil, err
	}

	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)

	keyMutex.Unlock()

	return privateKey, nil
}

func getPubKey() (*btcec.PublicKey, error) {
	priv, err := getPrivateKey()
	if err != nil {
		return nil, err
	}
	return priv.PubKey(), nil
}

func GetAddress() (btcutil.Address, error) {
	pub, err := getPubKey()
	if err != nil {
		return nil, err
	}

	addressPubKey, err := btcutil.NewAddressPubKey(pub.SerializeCompressed(), &chaincfg.Params{PubKeyHashAddrID: config.CoinType()})
	return addressPubKey, err
}

func Init() {
	address, err := GetAddress()
	if err != nil {
		logging.Error.Fatalf("Error initializing wallet: %s", err.Error())
	}
	logging.Info.Printf("My address on the public blockchain is: %s", address.EncodeAddress())
}

func getRpcClient() (*rpcclient.Client, error) {
	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         config.BtcNode(),
		User:         config.BtcRpcUser(),
		Pass:         config.BtcRpcPass(),
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		logging.Error.Fatal(err)
		return nil, err
	}

	return client, nil

}

func GetTxIns(neededAmount uint64) ([]*wire.TxIn, int64, error) {

	leftoverAmount := int64(neededAmount)
	client, err := getRpcClient()
	if err != nil {
		return nil, 0, err
	}
	defer client.Shutdown()

	address, err := GetAddress()
	if err != nil {
		return nil, 0, err
	}

	utxos, err := client.ListUnspentMinMaxAddresses(0, 99999999, []btcutil.Address{address})
	if err != nil {
		return nil, 0, err
	}

	var result []*wire.TxIn
	sort.Slice(utxos, func(i, j int) bool { return utxos[i].Amount < utxos[j].Amount })
	for _, utxo := range utxos {
		logging.Info.Printf("Leftover amount is %d\n", leftoverAmount)

		if leftoverAmount < 0 {
			break
		}

		hash, _ := chainhash.NewHashFromStr(utxo.TxID)
		script, _ := hex.DecodeString(utxo.ScriptPubKey)

		result = append(result, wire.NewTxIn(wire.NewOutPoint(hash, utxo.Vout), script, nil))

		leftoverAmount -= int64(utxo.Amount * 100000000)
	}

	if leftoverAmount > 0 {
		return nil, 0, fmt.Errorf("Insufficient balance")
	}

	return result, 0 - leftoverAmount, nil
}

func CreateNullDataTransaction(payload []byte) (*wire.MsgTx, error) {
	// make the tx
	tx := wire.NewMsgTx(1)

	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN).AddData(payload)
	b, _ := builder.Script()
	tx.AddTxOut(wire.NewTxOut(0, b))

	txins, change, err := GetTxIns(500)
	if err != nil {
		return nil, err
	}
	for _, txi := range txins {
		tx.AddTxIn(txi)
	}

	logging.Info.Printf("Change is %d\n", change)

	if change > 0 {
		// Add change output

		address, err := GetAddress()
		if err != nil {
			return nil, err
		}

		builder := txscript.NewScriptBuilder()
		builder.AddOp(txscript.OP_DUP).AddOp(txscript.OP_HASH160).AddData(address.ScriptAddress()).AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG)
		chgScript, _ := builder.Script()
		tx.AddTxOut(wire.NewTxOut(change, chgScript))
	}

	return tx, nil
}

func SignTransaction(tx *wire.MsgTx) (*wire.MsgTx, error) {

	priv, err := getPrivateKey()
	if err != nil {
		return nil, err
	}

	for i := range tx.TxIn {
		sigScript, err := txscript.SignatureScript(tx, i, tx.TxIn[i].SignatureScript, txscript.SigHashAll, priv, true)
		if err != nil {
			return nil, err
		}
		tx.TxIn[i].SignatureScript = sigScript
	}

	return tx, nil
}

func SendTransaction(tx *wire.MsgTx) (*chainhash.Hash, error) {
	nullHash, _ := chainhash.NewHashFromStr("")

	client, err := getRpcClient()
	if err != nil {
		return nullHash, err
	}
	defer client.Shutdown()

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	tx.Serialize(w)
	w.Flush()
	logging.Info.Printf("Sending commit transaction:\n%x\n", buf.Bytes())

	hash, err := client.SendRawTransaction(tx, false)
	if err != nil {
		return nullHash, err
	}

	return hash, nil
}
