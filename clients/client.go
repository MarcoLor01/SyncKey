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

func loadConfig(configuration string) Config {

	filePath := map[string]string{"1": "../serversAddrLocal.json", "2": "../serversAddrDocker.json"}[configuration]
	if filePath == "" {
		log.Fatalf("Error loading the configuration file: CONFIG is set to '%s'", configuration)
	}
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}
	var config Config
	if err = json.Unmarshal(fileContent, &config); err != nil {
		log.Fatal("Error in JSON decode:", err)
	}
	return config
}

func createClient(config Config) (int, *rpc.Client) {

	//Prendo un valore random compreso tra 0 e il numero di server - 1,
	//così da poter contattare un server random tra quelli disponibili
	//e successivamente lo vado a contattare

	serverNumber := rand.Intn(len(config.Address) - 1)

	client, err := rpc.Dial("tcp", config.Address[serverNumber].Addr)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	return serverNumber, client
}

func addElementToDS(configuration string, key string, value string, config Config, serverNumber int, client *rpc.Client) {

	var n1, n2 string
	if configuration == "1" {
		n1 = key
		n2 = value
	}
	if configuration == "2" {
		n1 = os.Getenv("KEY")
		n2 = os.Getenv("VALUE")
	}

	log.Printf("Adding element with Key: %s, and Value: %s contacting %s", n1, n2, config.Address[serverNumber].Addr) // make([]int, len(config.Address)),

	args := serverOperation.Message{Key: n1, Value: n2, VectorTimestamp: make([]int, len(config.Address)), OperationType: 1}
	log.Printf(call)

	result := &serverOperation.Response{
		Done:        false,
		Deliverable: false,
	}

	log.Printf(wait)

	err := client.Call("Server.ChoiceConsistencyPut", args, &result) //Calling the AddElement routine
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
}

func deleteElementFromDS(key string, config Config, client *rpc.Client) {
	n1 := key

	log.Printf("Get element with Key: %s\n", n1)
	args := serverOperation.Message{Key: n1, VectorTimestamp: make([]int, len(config.Address)), OperationType: 2}
	log.Printf(call)

	result := &serverOperation.Response{
		Done: false,
	}

	log.Printf(wait)
	err := client.Call("Server.ChoiceConsistencyDelete", args, &result)
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
}

func getElementFromDS(key string, client *rpc.Client) {

	log.Printf(call)
	log.Printf(wait)
	var returnValue string
	err := client.Call("Server.GetElement", key, &returnValue)
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
}

// Funzione per stabilire l'azione che devo svolgere
func establishActionToDo() string {
	var actionToDo string
	for {
		fmt.Print("Enter action (put/get/delete): ")
		_, err := fmt.Scan(&actionToDo)
		if err != nil {
			log.Fatal("Error in reading the action: ", err)
		}

		switch actionToDo {
		case "put", "delete", "get":
			// Se l'azione è valida, esce dal ciclo interno
			break
		default:
			log.Printf("Uncorrect flag. Please enter a valid action (put/get/delete).")
			continue
		}
		break
	}
	return actionToDo
}

func insertKey() string {
	var key string
	for {
		fmt.Print("Enter key: ")
		_, err := fmt.Scan(&key)
		if err != nil {
			log.Fatal(err)
		}
		if key != "" {
			break
		}
		fmt.Println("Key cannot be an empty string. Please enter a valid key.")
	}
	return key
}

func insertValue() string {
	var value string
	for {
		fmt.Print("Enter value: ")
		_, err := fmt.Scan(&value)
		if err != nil {
			log.Fatal(err)
		}
		if value != "" {
			break
		}
		fmt.Println("Value cannot be an empty string. Please enter a valid value.")
	}
	return value
}

func main() {

	time.Sleep(10 * time.Second) //Attendo che i server siano pronti

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	configuration := os.Getenv("CONFIG") //Carico il valore della variabile d'ambiente CONFIG,
	//che mi dice se sto usando la configurazione locale (CONFIG = 1) o quella docker (CONFIG = 2)

	config := loadConfig(configuration)

	serverNumber, client := createClient(config)

	for {
		var actionToDo, key, value string
		if configuration == "1" {
			actionToDo = establishActionToDo()
			//Devo prima di tutto andare a prendere la key per la ricerca nel datastore per qualsiasi configurazione
			key = insertKey()
			if actionToDo == "put" { //Se l'utente ha scelto di inserire un elemento nel datastore, deve inoltre fornirmi
				//il valore da associare alla Key, che non può essere vuoto
				value = insertValue()
			}
		} else if configuration == "2" {
			actionToDo = os.Getenv("OPERATION")
			fmt.Print("Enter value: ")
			_, err = fmt.Scan(&value)
			if err != nil {
				log.Fatal(err)
			}
		}

		switch actionToDo {
		case "not specified":
			log.Printf("When you call the datastore you have to specify the action with -a (action): \n-a put with key and value for a put operation, \n-a get with the key for a get operation, \n-a delete with the key for a delete operation\n")
			os.Exit(-1)
		case "put":
			addElementToDS(configuration, key, value, config, serverNumber, client)
		case "delete":
			deleteElementFromDS(key, config, client)
		case "get":
			getElementFromDS(key, client)
		default:
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
			err = client.Close()
			if err != nil {
				return
			}
			break
		}
	}
}
