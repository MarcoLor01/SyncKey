package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
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

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	configuration := os.Getenv("CONFIG")

	var filePath string

	if configuration == "1" {
		filePath = "../serversAddr.json"
	} else if configuration == "2" {
		filePath = "../serversAddr.json"
	} else {
		log.Fatalf("Error loading the configuration file: CONFIG is set to '%s'", configuration)
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading file:", err)
		return
	}

	var config Config
	if err1 := json.Unmarshal(fileContent, &config); err1 != nil {
		log.Fatal("Error in JSON decode:", err1)
		return
	}
	//I want a random ID between 1-5 (the number of the servers)

	serverNumber := rand.Intn(len(config.Address) - 1) //The server that I have to contact

	client, err := rpc.Dial("tcp", config.Address[1].Addr)

	defer func(client *rpc.Client) {
		err1 := client.Close()
		if err1 != nil {
			log.Fatal("Error in closing connection")
		}
	}(client)

	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}

	for {
		var actionToDo, key, value string
		if configuration == "1" {

			fmt.Print("Enter action (put/get/delete): ")
			_, err = fmt.Scan(&actionToDo)
			if err != nil {
				log.Fatal(err)
			}

			if actionToDo == "put" || actionToDo == "delete" {
				fmt.Print("Enter key: ")
				_, err = fmt.Scan(&key)
				if err != nil {
					log.Fatal(err)
				}
			}

			if actionToDo == "put" {
				fmt.Print("Enter value: ")
				_, err = fmt.Scan(&value)
				if err != nil {
					log.Fatal(err)
				}
			}
		} else if configuration == "2" {
			actionToDo = os.Getenv("OPERATION")
		}

		if actionToDo == "not specified" {
			log.Printf("When you call the datastore you have to specify the action with -a (action): \n-a put with key and value for a put operation, \n-a get with the key for a get operation, \n-a delete with the key for a delete operation\n")
			os.Exit(-1)
			//--------------------------------------PUT CASE--------------------------------------//
		} else if actionToDo == "put" {

			var n1, n2 string
			if configuration == "1" {
				n1 = key
				n2 = value
			}
			if configuration == "2" {
				n1 = os.Getenv("KEY")
				n2 = os.Getenv("VALUE")
			}

			scalarClock := 0
			log.Printf("Adding element with Key: %s, and Value: %s\n contacting %s", n1, n2, config.Address[serverNumber].Addr) // make([]int, len(config.Address)),

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
			var val string

			time.Sleep(1 * time.Second)

			if result.Done == true {
				val = "Successful"
			} else {
				val = "Failed"
			}
			log.Println(val)
			//--------------------------------------DELETE CASE--------------------------------------//

		} else if actionToDo == "delete" {

			n1 := key
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
			var val1 string

			time.Sleep(1 * time.Second)

			if result.Done == true {
				val1 = "Successful"
			} else {
				val1 = "Failed"
			}
			log.Println(val1)
			//--------------------------------------GET CASE--------------------------------------//
		} else if actionToDo == "get" {

			log.Printf("Synchronous call to RPC server")
			log.Printf("Wait...")
			err = client.Call("Server.GetElement", key, value) //Calling the DeleteElement function
			if err != nil {
				log.Fatal("Error adding the element to db, error: ", err)
			}
			time.Sleep(1 * time.Second)
			if value != "" {
				log.Printf("Element with Key %s, have value: %s", key, value)
			} else {
				log.Printf("No element with Key: %s\n", key)
			}
		} else {

			log.Printf("Uncorrect flag")
			return

		}

		var continueRunning string
		fmt.Print("Do you want to continue? (yes/no): ")
		_, err = fmt.Scan(&continueRunning)
		if err != nil {
			log.Fatal(err)
		}
		if continueRunning != "yes" {
			break
		}
	}

}
