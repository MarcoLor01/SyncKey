package main

import (
	"flag"
	"fmt"
	"log"
	"main/serverOperation"
	"net"
	"net/rpc"
	"os"
)

func main() {
	var myServer *serverOperation.Server
	actionToDo := flag.String("m", "", "action")
	flag.Parse()

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

	addr := "localhost:" + "1234" //My address
	serverOperation.MyId = 1      //My id

	server := rpc.NewServer()        //Creating a new server
	err := server.Register(myServer) //Registering my server

	if err != nil {
		log.Fatal("Format of service SyncKey is not correct: ", err)
	}
	lis, err := net.Listen("tcp", addr) //Listening the requests
	if err != nil {
		log.Fatal("Error while starting RPC server: ", err)
	}
	log.Printf("RPC server listens on port %d", 1234)
	server.Accept(lis)
}
