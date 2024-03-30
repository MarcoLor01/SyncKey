package main

import (
	"fmt"
	"log"
	"main/serverOperation"
	"net/rpc"
	"os"
)

const NumberOfServer = 5

func main() {
	addr := "localhost:" + "1234"
	client, err := rpc.Dial("tcp", addr)
	defer func(client *rpc.Client) {
		err := client.Close()
		if err != nil {

		}
	}(client)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	defer func(client *rpc.Client) {
		err := client.Close()
		if err != nil {
			log.Fatal("Error in closing connection")
		}
	}(client)

	if len(os.Args) < 3 {
		fmt.Printf("No args passed in\n")
		os.Exit(1)
	}
	n1 := os.Args[1]
	n2 := os.Args[2]
	timestamp := make([]int, NumberOfServer)
	scalarClock := 0
	fmt.Printf("Adding element with Key: %s, and Value: %s\n", n1, n2)
	fmt.Printf("Scalar clock: %d, vectorial clock: %d\n", scalarClock, timestamp)
	args := serverOperation.DbElement{Key: n1, Value: n2, Timestamp: timestamp, ScalarTimestamp: scalarClock}
	log.Printf("Synchronous call to RPC server")
	var returnString string
	err = client.Call("Server.AddElement", args, &returnString)
	if err != nil {
		log.Fatal("Error adding the element to db, error: ", err)
	}
	fmt.Println(returnString)
}
