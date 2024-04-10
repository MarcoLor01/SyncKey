package main

import (
	"flag"
	"fmt"
	"log"
	"main/serverOperation"
	"net/rpc"
	"os"
)

func main() {
	addr := "localhost:" + "1234" //Address of first server
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

	actionToDo := flag.String("a", "not specified", "action") //Defining a flag, with that
	//the user can specify what he wants to do
	flag.Parse()

	if *actionToDo == "not specified" {
		fmt.Printf("When you call the datastore you have to specify the action with -a (action): \n-a put with key and value for a put operation, \n-a get with the key for a get operation, \n-a delete with the key for a delete operation")
		os.Exit(-1)
	}

	//--------------------------------------PUT CASE--------------------------------------//
	if *actionToDo == "put" {
		if len(os.Args) < 5 { //The procedure requests 3 arguments + 2 arguments from flag
			fmt.Printf("No args passed in\n")
			os.Exit(1)
		}
		n1 := os.Args[3] //Key
		n2 := os.Args[4] //Value
		scalarClock := 0
		fmt.Printf("Adding element with Key: %s, and Value: %s\n", n1, n2)
		args := serverOperation.Message{Key: n1, Value: n2, ScalarTimestamp: scalarClock}

		log.Printf("Synchronous call to RPC server")

		var result bool

		err = client.Call("Server.AddElement", args, &result) //Calling the AddElement routine
		if err != nil {
			log.Fatal("Error adding the element to db, error: ", err)
		}

		var value string

		if result == true {
			value = "SUCCESS"
		} else {
			value = "ABORTED"
		}
		fmt.Println(value)
	}

	//--------------------------------------DELETE CASE--------------------------------------//

	if *actionToDo == "delete" {
		if len(os.Args) < 3 { //The procedure requests 1 arguments + 2 arguments from flag
			fmt.Printf("No args passed in\n")
			os.Exit(1)
		}
		n1 := os.Args[3] //Key
		scalarClock := 0

		fmt.Printf("Get element with Key: %s\n", n1)
		args := serverOperation.Message{Key: n1, ScalarTimestamp: scalarClock}

		log.Printf("Synchronous call to RPC server")

		var result bool

		err = client.Call("Server.DeleteElement", args, &result) //Calling the DeleteElement function
		if err != nil {
			log.Fatal("Error adding the element to db, error: ", err)
		}

		var value string

		if result == true {
			value = "SUCCESS"
		} else {
			value = "ABORTED"
		}
		fmt.Println(value)
	}
}
