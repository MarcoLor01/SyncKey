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

func addElementToDsSequential(key string, value string, config Config, serverNumber int, client *rpc.Client) {

	var n1, n2 string
	n1 = key
	n2 = value

	log.Printf("Adding element with Key: %s, and Value: %s contacting %s", n1, n2, config.Address[serverNumber].Addr)

	args := serverOperation.MessageSequential{Key: n1, Value: n2, OperationType: 1}
	log.Printf(call)

	result := serverOperation.ResponseSequential{Deliverable: false, Done: false}

	log.Printf(wait)

	err := client.Call("ServerSequential.SequentialSendElement", args, &result) //Calling the SequentialSendElement routine
	if err != nil {
		log.Fatal(datastoreError, err)
	}

	var val string

	time.Sleep(500 * time.Millisecond)

	if result.Deliverable {
		val = "Successful"
	} else {
		val = "Failed"
	}
	log.Println(val)
}

func addElementToDsCausal(key string, value string, config Config, serverNumber int, client *rpc.Client) {

	var n1, n2 string
	n1 = key
	n2 = value

	log.Printf("Adding element with Key: %s, and Value: %s contacting %s", n1, n2, config.Address[serverNumber].Addr)

	args := serverOperation.MessageCausal{Key: n1, Value: n2, VectorTimestamp: make([]int, len(config.Address)), OperationType: 1}
	log.Printf(call)

	result := serverOperation.ResponseCausal{Deliverable: false}

	log.Printf(wait)

	err := client.Call("ServerCausal.CausalSendElement", args, &result) //Calling the CausalSendElement routine
	if err != nil {
		log.Fatal(datastoreError, err)
	}

	var val string

	time.Sleep(500 * time.Millisecond)

	if result.Deliverable {
		val = "Successful"
	} else {
		val = "Failed"
	}
	log.Println(val)
}

func deleteElementFromDsSequential(key string, config Config, serverNumber int, client *rpc.Client) {
	n1 := key

	log.Printf("Get element with Key: %s contacting %s", n1, config.Address[serverNumber].Addr)

	args := serverOperation.MessageSequential{Key: n1, OperationType: 2}
	result := serverOperation.ResponseSequential{Deliverable: false, Done: false}

	log.Printf(call)
	log.Printf(wait)

	err := client.Call("ServerSequential.SequentialSendElement", args, &result)

	if err != nil {
		log.Fatal(datastoreError, err)
	}
	var val1 string

	time.Sleep(1 * time.Second)

	if result.Deliverable {
		val1 = "Successful"
	} else {
		val1 = "Failed"
	}
	log.Println(val1)
}

func deleteElementFromDsCausal(key string, config Config, serverNumber int, client *rpc.Client) {
	n1 := key

	log.Printf("Get element with Key: %s contacting %s\n", n1, config.Address[serverNumber].Addr)

	args := serverOperation.MessageCausal{Key: n1, VectorTimestamp: make([]int, len(config.Address)), OperationType: 2}
	result := serverOperation.ResponseCausal{Deliverable: false}

	log.Printf(call)
	log.Printf(wait)

	err := client.Call("ServerCausal.CausalSendElement", args, &result)

	if err != nil {
		log.Fatal(datastoreError, err)
	}
	var val1 string

	time.Sleep(1 * time.Second)

	if result.Deliverable {
		val1 = "Successful"
	} else {
		val1 = "Failed"
	}
	log.Println(val1)
}

func getElementFromDsCausal(key string, client *rpc.Client) {

	log.Printf(call)
	log.Printf(wait)
	var returnValue string
	err := client.Call("ServerCausal.CausalGetElement", key, &returnValue)
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

func getElementFromDsSequential(key string, client *rpc.Client) {

	log.Printf(call)
	log.Printf(wait)
	var returnValue string
	err := client.Call("ServerSequential.SequentialGetElement", key, &returnValue)
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

func loadEnvironment() string {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
	configuration := os.Getenv("CONFIG") //Carico il valore della variabile d'ambiente CONFIG,
	//che mi dice se sto usando la configurazione locale (CONFIG = 1) o quella docker (CONFIG = 2)
	return configuration
}

func getAction() (string, string, string) {
	var actionToDo, key, value string

	actionToDo = establishActionToDo()
	//Devo prima di tutto andare a prendere la key per la ricerca nel datastore per qualsiasi configurazione
	key = insertKey()
	if actionToDo == "put" {
		//Se l'utente ha scelto di inserire un elemento nel datastore, deve inoltre fornirmi
		//il valore da associare alla Key, che non può essere vuoto
		value = insertValue()
	}
	return actionToDo, key, value
}

func getConsistency() (int, string) {
	var consistencyType int
	for {
		fmt.Print("Enter the consistency type (0 for Causal, 1 for Sequential): ")
		_, err := fmt.Scan(&consistencyType)
		if err != nil {
			log.Fatal("Error reading input: ", err)
		}

		switch consistencyType {
		case 0:
			return 0, "Causal"
		case 1:
			return 1, "Sequential"
		default:
			fmt.Println("Invalid input. Please enter 0 for Causal or 1 for Sequential.")
		}
	}
}

func main() {

	//time.Sleep(10 * time.Second) //Attendo che i server siano pronti

	consistency, consistencyString := getConsistency()
	fmt.Println("The client is using server with the consistency: ", consistencyString)
	configuration := loadEnvironment()

	config := loadConfig(configuration)
	serverNumber, client := createClient(config)

	//Voglio garantire trasparenza al client riguardo al tipo di consistenza che voglio implementare,
	//eseguo quindi una chiamata RPC a un server per sapere la modalità con cui il server è stato avviato

	for {
		actionToDo, key, value := getAction()
		//In base all'azione che l'utente ha scelto di svolgere, vado a chiamare la funzione corrispondente
		switch actionToDo {
		case "not specified":
			log.Printf("When you call the datastore you have to specify the action with -a (action): \n-a put with key and value for a put operation, \n-a get with the key for a get operation, \n-a delete with the key for a delete operation\n")
			os.Exit(-1)

		case "put":
			if consistency == 0 {
				addElementToDsCausal(key, value, config, serverNumber, client)
			} else {
				addElementToDsSequential(key, value, config, serverNumber, client)
			}

		case "delete":
			if consistency == 0 {
				deleteElementFromDsCausal(key, config, serverNumber, client)
			} else {
				deleteElementFromDsSequential(key, config, serverNumber, client)
			}

		case "get":
			if consistency == 0 {
				getElementFromDsCausal(key, client)
			} else {
				getElementFromDsSequential(key, client)
			}

		default:
			log.Printf("Uncorrect flag")
			return
		}

		//Chiedo all'utente se vuole continuare a svolgere operazioni sul datastore

		var continueRunning string
		fmt.Print("Do you want to continue? (yes/no): ")
		_, err := fmt.Scan(&continueRunning)
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
