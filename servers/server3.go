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

	path := os.Getenv("ENV_PATH")
	err1 := godotenv.Load(path)
	if err1 != nil {
		log.Fatal("Error loading .env file")
	}

	configuration := os.Getenv("CONFIG")
	var address string
	if configuration == "2" {
		action := os.Getenv("MOD")
		actionToDo = &action
		address = "server3:8084"
	} else if configuration == "1" {
		actionToDo = flag.String("m", "", "action")
		flag.Parse()
		address = "localhost:3456"
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

	serverOperation.MyId = 3 //My id

	server := rpc.NewServer()        //Creating a new server
	err := server.Register(myServer) //Registering my server

	if err != nil {
		log.Fatal("Format of service SyncKey is not correct: ", err)
	}
	lis, err := net.Listen("tcp", address) //Listening the requests
	if err != nil {
		log.Fatal("Error while starting RPC server: ", err)
	}
	log.Printf("RPC server listens on port %d", 3456)
	server.Accept(lis)
}
