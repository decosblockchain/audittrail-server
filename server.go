package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/decosblockchain/audittrail-server/commit"
	"github.com/decosblockchain/audittrail-server/integrity"
	"github.com/decosblockchain/audittrail-server/logging"
	"github.com/decosblockchain/audittrail-server/routes"
	"github.com/decosblockchain/audittrail-server/wallet"
	"github.com/gorilla/mux"
)

func main() {
	logging.Init(os.Stdout, os.Stdout, os.Stdout, os.Stdout)

	wallet.Init()

	ok, err := integrity.CheckIntegrity()
	if err != nil {
		logging.Error.Panic("Error checking integrity", err)
		os.Exit(2)
	}

	if !ok {
		logging.Error.Panic("Integrity check failed")
		os.Exit(3)
	}

	ticker := time.NewTicker(15 * time.Second)
	go func() {
		for range ticker.C {
			go func() {
				err := commit.CommitToMasterChain()
				if err != nil {
					logging.Error.Printf("Error committing to master chain: %s", err.Error())
				}
			}()
		}
	}()

	r := mux.NewRouter()
	r.HandleFunc("/send", routes.SendHandler)

	log.Fatal(http.ListenAndServe(":3000", r))
}
