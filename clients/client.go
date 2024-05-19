package main

import (
	"fmt"
	"log"
	"main/clientCommon"
	"math/rand"
	"net/rpc"
	"os"
	"strings"
)

func createCasualClient(config clientCommon.Config) (int, *rpc.Client) {

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
	configuration := clientCommon.LoadEnvironment()

	config := clientCommon.LoadConfig(configuration)
	serverNumber, client := createCasualClient(config)

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
			clientCommon.HandlePutAction(consistency, key, value, config, serverNumber, client)
		case "delete":
			clientCommon.HandleDeleteAction(consistency, key, config, serverNumber, client)
		case "get":
			clientCommon.HandleGetAction(consistency, key, client)
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
