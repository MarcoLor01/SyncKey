package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"main/serverOperation"
	"net"
	"net/rpc"
	"os"
)

func main() {

	path := os.Getenv("ENV_PATH") //Prendo il path del file .env, che cambia in base
	//a se sto usando la configurazione locale o quella docker
	err := godotenv.Load(path)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configuration := os.Getenv("CONFIG")

	var addressNumber string
	var address string
	if configuration == "2" { //Se sto usando la configurazione docker
		addressNumber = "8083"
		address = "server2:" + addressNumber
	} else if configuration == "1" {
		addressNumber = "2345"
		address = "localhost:" + addressNumber
	}
	var server *rpc.Server
	server = rpc.NewServer()
	err = serverOperation.InitializeServerList()
	if err != nil {
		log.Fatal(err)
	}
	serverOperation.InitializeAndRegisterServerSequential(server, 2)
	serverOperation.InitializeAndRegisterServerCausal(server, 2)
	fmt.Println("Server registered successfully")
	serverOperation.MyId = 2 //Imposto il mio Id che deve essere univoco per ogni server

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Error while starting RPC server: ", err)
	}
	log.Println("RPC server listens on port ", addressNumber)
	server.Accept(lis)
}
