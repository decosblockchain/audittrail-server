package routes

import (
	"encoding/hex"
	"io/ioutil"
	"net/http"

	"github.com/decosblockchain/audittrail-server/config"
	"github.com/decosblockchain/audittrail-server/logging"
	"github.com/onrik/ethrpc"
)

func SendHandler(w http.ResponseWriter, r *http.Request) {
	logging.Info.Printf("Received request to /send\n")

	if r.Method != "POST" {
		logging.Error.Printf("Invalid HTTP method %s\n", r.Method)
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	client := ethrpc.New(config.EthNode())

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.Error.Printf("%s\n", err.Error())
	}

	hexTx := "0x" + hex.EncodeToString(bytes)
	logging.Info.Printf("Received TX: %s", hexTx)

	_, err = client.EthSendRawTransaction(hexTx)
	if err != nil {
		logging.Error.Printf("%s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
