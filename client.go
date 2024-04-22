package main

import (
	"encoding/json"
	"flag"
	"log"
	"main/serverOperation"
	"math/rand"
	"net/rpc"
	"os"
	"time"
)

type Addrs struct {
	ID   int    `json:"id"`
	Addr string `json:"addr"`
}

type Config struct {
	Address []Addrs `json:"address"`
}

func main() {

	fileContent, err := os.ReadFile("serversAddr.json")
	if err != nil {
		log.Fatal("Error reading file:", err) //Reading the address from the json file
		return
	}

	var config Config
	if err := json.Unmarshal(fileContent, &config); err != nil {
		log.Fatal("Error in JSON decode:", err)
		return
	}
	//I want a random ID between 1-5 (the number of the servers)

	serverNumber := rand.Intn(len(config.Address) - 1) //The server that I have to contact

	client, err := rpc.Dial("tcp", config.Address[serverNumber].Addr)

	defer func(client *rpc.Client) {
		err1 := client.Close()
		if err1 != nil {
			log.Fatal("Error in closing connection")
		}
	}(client)

	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}

	actionToDo := flag.String("a", "not specified", "action") //Define a flag, with that
	//the user can specify what he wants to do
	flag.Parse()

	if *actionToDo == "not specified" {
		log.Printf("When you call the datastore you have to specify the action with -a (action): \n-a put with key and value for a put operation, \n-a get with the key for a get operation, \n-a delete with the key for a delete operation\n")
		os.Exit(-1)
		//--------------------------------------PUT CASE--------------------------------------//
	} else if *actionToDo == "put" {

		if len(os.Args) < 5 { //The procedure requests 3 arguments + 2 arguments from flag
			log.Printf("No args passed in\n")
			os.Exit(1)
		}
		n1 := os.Args[3] //Key
		n2 := os.Args[4] //Value
		scalarClock := 0
		log.Printf("Adding element with Key: %s, and Value: %s\n contacting %s", n1, n2, config.Address[serverNumber].Addr)
		args := serverOperation.Message{Key: n1, Value: n2, ScalarTimestamp: scalarClock, VectorTimestamp: make([]int, len(config.Address)), OperationType: 1}
		log.Printf("Synchronous call to RPC server")

		result := &serverOperation.Response{
			Done: false,
		}

		log.Printf("Wait...")
		err = client.Call("Server.ChoiceConsistency", args, &result) //Calling the AddElement routine
		if err != nil {
			log.Fatal("Error adding the element to db, error: ", err)
		}
		var value string

		time.Sleep(1 * time.Second)

		if result.Done == true {
			value = "Successful"
		} else {
			value = "Failed"
		}
		log.Println(value)
		//--------------------------------------DELETE CASE--------------------------------------//

	} else if *actionToDo == "delete" {
		if len(os.Args) < 3 { //The procedure requests 1 arguments + 2 arguments from flag
			log.Printf("No args passed in\n")
			os.Exit(1)
		}
		n1 := os.Args[3] //Key
		scalarClock := 0

		log.Printf("Get element with Key: %s\n", n1)
		args := serverOperation.Message{Key: n1, ScalarTimestamp: scalarClock, OperationType: 2}

		log.Printf("Synchronous call to RPC server")

		result := &serverOperation.Response{
			Done: false,
		}

		log.Printf("Wait...")
		err = client.Call("Server.AddElement", args, &result) //Calling the DeleteElement function
		if err != nil {
			log.Fatal("Error adding the element to db, error: ", err)
		}

		var value string

		time.Sleep(1 * time.Second)

		if result.Done == true {
			value = "Successful"
		} else {
			value = "Failed"
		}
		log.Println(value)
		//--------------------------------------GET CASE--------------------------------------//
	} else if *actionToDo == "get" {

		if len(os.Args) < 3 { //I need three parameters, 2 for the flag -a and 1 for the get operation<
			log.Printf("No args passsed in\n")
			os.Exit(-1)
		}

		key := os.Args[3]
		value := new(string)
		log.Printf("Synchronous call to RPC server")
		log.Printf("Wait...")
		err = client.Call("Server.GetElement", key, value) //Calling the DeleteElement function
		if err != nil {
			log.Fatal("Error adding the element to db, error: ", err)
		}
		time.Sleep(1 * time.Second)
		if *value != "" {
			log.Printf("Element with Key %s, have value: %s", key, *value)
		} else {
			log.Printf("No element with Key: %s\n", key)
		}
	} else {

		log.Printf("Uncorrect flag")
		return

	}
}
