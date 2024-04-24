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

	err1 := godotenv.Load("../.env")
	if err1 != nil {
		log.Fatalf("Error loading .env file: %v", err1)
	}

	configuration := os.Getenv("CONFIG")

	if configuration == "2" {
		action := os.Getenv("MOD")
		actionToDo = &action
	} else if configuration == "1" {
		actionToDo = flag.String("m", "", "action")
		flag.Parse()
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

	addr := "localhost:" + "2345" //My address
	serverOperation.MyId = 2      //My id
	server := rpc.NewServer()
	err := server.Register(myServer) //Registering my server

	if err != nil {
		log.Fatal("Format of service SyncKey is not correct: ", err)
	}
	lis, err := net.Listen("tcp", addr) //Listening the requests
	if err != nil {
		log.Fatal("Error while starting RPC server: ", err)
	}
	log.Printf("RPC server listens on port %d", 2345)
	server.Accept(lis)
}
