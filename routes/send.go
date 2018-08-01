package routes

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/decosblockchain/audittrail-server/logging"
)

func SendHandler(w http.ResponseWriter, r *http.Request) {
	logging.Info.Printf("Received request to /send\n")

	if r.Method != "POST" {
		logging.Error.Printf("Invalid HTTP method %s\n", r.Method)
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	incomingBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.Error.Printf("%s\n", err.Error())
	}

	buffer := new(bytes.Buffer)
	buffer.Write([]byte("{\"method\":\"sendrawtransaction\",\"params\":["))
	buffer.Write(incomingBytes)
	buffer.Write([]byte("],\"id\":0,\"jsonrpc\":\"2.0\"}"))

	logging.Info.Printf("Sending to CK RPC:\n%s", string(buffer.Bytes()))

	req, err := http.NewRequest("POST", "http://localhost:8384/", bytes.NewReader(buffer.Bytes()))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logging.Error.Printf("Error calling server: %s\n", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(http.StatusCreated)
}
