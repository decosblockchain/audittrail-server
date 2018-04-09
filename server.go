package main

import (
	"log"
	"net/http"
	"os"

	"github.com/decosblockchain/audittrail-server/logging"
	"github.com/decosblockchain/audittrail-server/routes"
	"github.com/gorilla/mux"
)

func main() {
	logging.Init(os.Stdout, os.Stdout, os.Stdout, os.Stdout)

	r := mux.NewRouter()
	r.HandleFunc("/send", routes.SendHandler)

	log.Fatal(http.ListenAndServe(":3000", r))
}
