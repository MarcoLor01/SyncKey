package main

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"main/serverOperation"
	"net"
	"net/rpc"
	"os"
)

func main() {
	var myServer *serverOperation.Server

	var actionToDo *string

	path := os.Getenv("ENV_PATH") //Prendo il path del file .env, che cambia in base
	//a se sto usando la configurazione locale o quella docker
	err1 := godotenv.Load(path)
	if err1 != nil {
		log.Fatal("Error loading .env file")
	}

	configuration := os.Getenv("CONFIG")

	var address string
	if configuration == "2" { //Se sto usando la configurazione docker recupero la
		//modalità in cui il server deve essere eseguito tramite la variabile
		//d'ambiente MOD
		action := os.Getenv("MOD")
		actionToDo = &action
		address = "server1:8082"
	} else if configuration == "1" {
		actionToDo = flag.String("m", "", "action")
		flag.Parse()
		address = "localhost:1234"
	}

	serverOperation.InitializeServerList()

	switch *actionToDo {
	case "seq":
		myServer = serverOperation.CreateNewSequentialDataStore()
	case "caus":
		myServer = serverOperation.CreateNewCausalDataStore()
	default:
		fmt.Println("Specify the modality with -m: seq for sequential consistency, caus for causal consistency")
		os.Exit(-1)
	}

	serverOperation.MyId = 1 //Imposto il mio Id che deve essere univoco per ogni server

	server := rpc.NewServer()
	err := server.Register(myServer)

	if err1 != nil {
		log.Fatal("Format of service SyncKey is not correct: ", err)
	}
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Error while starting RPC server: ", err)
	}
	log.Printf("RPC server listens on port %d", 1234)
	server.Accept(lis)
}
