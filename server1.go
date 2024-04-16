package main

import (
	"log"
	"main/serverOperation"
	"net"
	"net/rpc"
)

func main() {
	myServer := serverOperation.CreateNewConsistentialDataStore() //Create a new Server
	addr := "localhost:" + "1234"                                 //My address
	serverOperation.MyId = 1                                      //My id
	server := rpc.NewServer()
	err := server.Register(myServer) //Registering my server
	serverOperation.InitializeServerList()
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
