package main

import (
	"log"
	"main/serverOperation"
	"net"
	"net/rpc"
)

func main() {
	myServer := serverOperation.CreateNewConsistentialDataStore()
	addr := "localhost:" + "2345"
	serverOperation.MyId = 2
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
	log.Printf("RPC server listens on port %d", 2345)
	server.Accept(lis)
}
