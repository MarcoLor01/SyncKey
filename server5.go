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
	actionToDo := flag.String("m", "not specified", "action") //Defining a flag, with that
	//the user can specify what he wants to do
	flag.Parse()
	serverOperation.InitializeServerList()
	if *actionToDo == "not specified" {
		fmt.Printf("When you call the server you have to specify the modality with -m: \n-a seq for sequential consistency, \n-a caus for causal consistency\n")
		os.Exit(-1)
	} else if *actionToDo == "seq" {
		myServer = serverOperation.CreateNewSequentialDataStore() //Create a new Server
	} else if *actionToDo == "caus" {
		myServer = serverOperation.CreateNewCausalDataStore() //Change
	} else {
		fmt.Println("Error value for the flag")
		return
	}
	addr := "localhost:" + "5678"
	serverOperation.MyId = 5
	server := rpc.NewServer()
	err := server.Register(myServer)
	serverOperation.InitializeServerList()
	if err != nil {
		log.Fatal("Format of service SyncKey is not correct: ", err)
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Error while starting RPC server: ", err)
	}
	log.Printf("RPC server listens on port %d", 5678)
	server.Accept(lis)
}
