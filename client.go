package main

import (
	"flag"
	"fmt"
	"log"
	"main/serverOperation"
	"math/rand"
	"net/rpc"
	"os"
	"strings"
)

func main() {

	//I want a random ID between 1-5 (the number of the servers)
	serverId := rand.Intn(4) + 1
	addr := "localhost"
	switch serverId {
	case 1:
		addr = strings.Join([]string{addr, "1234"}, ":")
	case 2:
		addr = strings.Join([]string{addr, "2345"}, ":")
	case 3:
		addr = strings.Join([]string{addr, "3456"}, ":")
	case 4:
		addr = strings.Join([]string{addr, "4567"}, ":")
	case 5:
		addr = strings.Join([]string{addr, "5678"}, ":")
	}

	client, err := rpc.Dial("tcp", addr)
	fmt.Printf(addr)
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

	//--------------------------------------GET CASE--------------------------------------//

	if *actionToDo == "get" {

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
		if value != nil {
			fmt.Printf("Element with Key %s, have value: %s", key, *value)
		} else {
			fmt.Printf("Problem with the get operation")
		}
	}

	if *actionToDo == "test" {

	}
}
