package main

import (
	"log"
	"os"
	"strconv"

	"github.com/box1bs/pur/pur_api/pkg/database"
	"github.com/box1bs/pur/pur_api/pkg/handlers"
)

func main() {
	internalVars, err := handlers.ExtractVars([]string{"DSN", "PORT", "SUMMARY_API_ADDR", "MAX_CONCURRENCY", "ACCURACY"})
	if err != nil {
		log.Printf("failed extract env params: %v", err)
	}
	
	storage, err := database.ConnectToDB(internalVars[0])
	if err != nil {
		log.Printf("failed connect to db: %v\n", err)
		os.Exit(1)
	}

	accur, err := strconv.Atoi(internalVars[4])
	if err != nil {
		log.Printf("error convert accuracy: %v", err)
		return
	}

	oncurrencyControl, err := strconv.Atoi(internalVars[3])
	if err != nil {
		log.Printf("error convert concurrencyControl: %v", err)
		return
	}

	server := handlers.NewAPIServer(internalVars[1], storage, accur, oncurrencyControl, internalVars[2])
	server.Run()
}