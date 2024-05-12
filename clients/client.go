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
	"strings"
	"time"
)

type Config struct {
	Address []struct {
		ID   int    `json:"id"`
		Addr string `json:"addr"`
	} `json:"address"`
}

const wait string = "Wait..."
const call string = "Synchronous call to RPC server"
const datastoreError string = "Error adding the element to datastore, error: "

func main() {

	time.Sleep(10 * time.Second) //Attendo che i server siano pronti

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	configuration := os.Getenv("CONFIG") //Carico il valore della variabile d'ambiente CONFIG,
	//che mi dice se sto usando la configurazione locale (CONFIG = 1) o quella docker (CONFIG = 2)
	filePath := map[string]string{"1": "../serversAddrLocal.json", "2": "../serversAddrDocker.json"}[configuration]
	if filePath == "" {
		log.Fatalf("Error loading the configuration file: CONFIG is set to '%s'", configuration)
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	var config Config
	if err1 := json.Unmarshal(fileContent, &config); err1 != nil {
		log.Fatal("Error in JSON decode:", err1)
	}
	//Prendo un valore random compreso tra 0 e il numero di server - 1,
	//così da poter contattare un server randomicamente

	serverNumber := rand.Intn(len(config.Address) - 1) //Sarà il server che contatterò per tutte le mie
	//prossime richieste RPC

	client, err := rpc.Dial("tcp", config.Address[serverNumber].Addr)

	defer func(client *rpc.Client) {
		err1 := client.Close()
		if err1 != nil {
			log.Fatal("Error in closing connection", err1)
		}
	}(client)

	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}

	for {
		var actionToDo, key, value string
		if configuration == "1" {

			fmt.Print("Enter action (put/get/delete): ")
			_, err := fmt.Scan(&actionToDo)
			if err != nil {
				log.Fatal("Error in reading the action: ", err)
			}

			if actionToDo == "put" || actionToDo == "delete" || actionToDo == "get" { //Se devo inserire o eliminare un elemento dal datastore,
				//devo ricevere la Key dall'utente, che non può essere vuota
				for {
					fmt.Print("Enter key: ")
					_, err = fmt.Scan(&key)
					if err != nil {
						log.Fatal(err)
					}
					if key != "" {
						break
					}
					fmt.Println("Key cannot be an empty string. Please enter a valid key.")
				}
			}

			if actionToDo == "put" { //Se l'utente ha scelto di inserire un elemento nel datastore, deve inoltre fornirmi
				//il valore da associare alla Key, che non può essere vuoto
				for {
					fmt.Print("Enter value: ")
					_, err = fmt.Scan(&value)
					if err != nil {
						log.Fatal(err)
					}
					if value != "" {
						break
					}
					fmt.Println("Value cannot be an empty string. Please enter a valid value.")
				}
			}

		} else if configuration == "2" {
			actionToDo = os.Getenv("OPERATION")
			fmt.Print("Enter value: ")
			_, err = fmt.Scan(&value)
			if err != nil {
				log.Fatal(err)
			}
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

			log.Printf("Adding element with Key: %s, and Value: %s\n contacting %s", n1, n2, config.Address[serverNumber].Addr) // make([]int, len(config.Address)),

			args := serverOperation.Message{Key: n1, Value: n2, VectorTimestamp: make([]int, len(config.Address)), OperationType: 1}
			log.Printf(call)

			result := &serverOperation.Response{
				Done:        false,
				Deliverable: false,
			}

			log.Printf(wait)
			err = client.Call("Server.ChoiceConsistencyPut", args, &result) //Calling the AddElement routine
			if err != nil {
				log.Fatal(datastoreError, err)
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

			log.Printf("Get element with Key: %s\n", n1)
			args := serverOperation.Message{Key: n1, VectorTimestamp: make([]int, len(config.Address)), OperationType: 2}
			log.Printf(call)

			result := &serverOperation.Response{
				Done: false,
			}

			log.Printf(wait)
			err = client.Call("Server.ChoiceConsistencyDelete", args, &result)
			if err != nil {
				log.Fatal(datastoreError, err)
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

			log.Printf(call)
			log.Printf(wait)
			var returnValue string
			err = client.Call("Server.GetElement", key, &returnValue)
			fmt.Println("Get element with Key: ", key)
			fmt.Println("Value: ", returnValue)
			if err != nil {
				log.Fatal(datastoreError, err)
			}
			time.Sleep(1 * time.Second)

			if returnValue != "" {
				log.Printf("Element with Key %s, have value: %s", key, returnValue)
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
		continueRunning = strings.ToLower(continueRunning)
		if continueRunning != "yes" {
			break
		}
	}
}
