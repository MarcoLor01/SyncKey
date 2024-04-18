package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"main/serverOperation"
	"math/rand"
	"net/rpc"
	"os"
)

func main() {

	fileContent, err := os.ReadFile("serversAddr.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	var config Config
	if err := json.Unmarshal(fileContent, &config); err != nil {
		fmt.Println("Error in JSON decode:", err)
		return
	}
	//I want a random ID between 1-5 (the number of the servers)

	serverNumber := rand.Intn(len(config.Address) - 1) //The server that I have to contact

	client, err := rpc.Dial("tcp", config.Address[serverNumber].Addr)

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
		fmt.Printf("When you call the datastore you have to specify the action with -a (action): \n-a put with key and value for a put operation, \n-a get with the key for a get operation, \n-a delete with the key for a delete operation\n")
		os.Exit(-1)
		//--------------------------------------PUT CASE--------------------------------------//
	} else if *actionToDo == "put" {
		if len(os.Args) < 5 { //The procedure requests 3 arguments + 2 arguments from flag
			fmt.Printf("No args passed in\n")
			os.Exit(1)
		}
		n1 := os.Args[3] //Key
		n2 := os.Args[4] //Value
		scalarClock := 0
		fmt.Printf("Adding element with Key: %s, and Value: %s\n contacting %s\n", n1, n2, config.Address[serverNumber].Addr)
		args := serverOperation.Message{Key: n1, Value: n2, ScalarTimestamp: scalarClock /*VectorTimestamp: make([]int, serverNumber)*/, OperationType: 1}

		log.Printf("Synchronous call to RPC server")

		result := &serverOperation.Response{
			Done: false,
		}

		err = client.Call("Server.AddElement", args, &result) //Calling the AddElement routine
		if err != nil {
			log.Fatal("Error adding the element to db, error: ", err)
		}
		var value string

		if result.Done == true {
			value = "SUCCESS"
		} else {
			value = "ABORTED"
		}
		fmt.Println(value)
		//--------------------------------------DELETE CASE--------------------------------------//
	} else if *actionToDo == "delete" {
		if len(os.Args) < 3 { //The procedure requests 1 arguments + 2 arguments from flag
			fmt.Printf("No args passed in\n")
			os.Exit(1)
		}
		n1 := os.Args[3] //Key
		scalarClock := 0

		fmt.Printf("Get element with Key: %s\n", n1)
		args := serverOperation.Message{Key: n1, ScalarTimestamp: scalarClock, OperationType: 2}

		log.Printf("Synchronous call to RPC server")

		result := &serverOperation.Response{
			Done: false,
		}

		err = client.Call("Server.AddElement", args, &result) //Calling the DeleteElement function
		if err != nil {
			log.Fatal("Error adding the element to db, error: ", err)
		}
		fmt.Println("2")
		var value string

		if result.Done == true {
			value = "SUCCESS"
		} else {
			value = "ABORTED"
		}
		fmt.Println(value)
		//--------------------------------------GET CASE--------------------------------------//
	} else if *actionToDo == "get" {

		if len(os.Args) < 3 { //I need three parameters, 2 for the flag -a and 1 for the get operation<
			fmt.Printf("No args passsed in\n")
			os.Exit(-1)
		}

		key := os.Args[3]
		value := new(string)
		log.Printf("Synchronous call to RPC server")

		err = client.Call("Server.GetElement", key, value) //Calling the DeleteElement function
		if err != nil {
			log.Fatal("Error adding the element to db, error: ", err)
		}

		if *value != "" {
			fmt.Printf("Element with Key %s, have value: %s", key, *value)
		} else {
			fmt.Printf("No element with Key: %s\n", key)
		}
	} else {
		log.Printf("Uncorrect flag")
		return
	}
}
