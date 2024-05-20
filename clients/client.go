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

func createCasualClient(config clientCommon.Config) (int, *rpc.Client, error) {

	//Prendo un valore random compreso tra 0 e il numero di server - 1,
	//così da poter contattare un server random tra quelli disponibili
	//e successivamente lo vado a contattare

	serverNumber := rand.Intn(len(config.Address) - 1)

	client, err := rpc.Dial("tcp", config.Address[serverNumber].Addr)
	if err != nil {
		return -1, nil, fmt.Errorf("error in dialing: %w", err)
	}
	return serverNumber, client, nil
}

// Funzione per stabilire l'azione che devo svolgere
func establishActionToDo() (string, error) {
	var actionToDo string
	for {
		fmt.Print("Enter action (put/get/delete): ")
		_, err := fmt.Scan(&actionToDo)
		if err != nil {
			return "", fmt.Errorf("error in reading the action: %w", err)
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
	return actionToDo, nil
}

func insertKey() (string, error) {
	var key string
	for {
		fmt.Print("Enter key: ")
		_, err := fmt.Scan(&key)
		if err != nil {
			return "", fmt.Errorf("error in scan function %w", err)
		}
		if key != "" {
			break
		}
		log.Println("Key cannot be an empty string. Please enter a valid key.")
	}
	return key, nil
}

func insertValue() (string, error) {
	var value string
	for {
		fmt.Print("Enter value: ")
		_, err := fmt.Scan(&value)
		if err != nil {
			return "", fmt.Errorf("error in scan function %w", err)
		}
		if value != "" {
			break
		}
		fmt.Println("Value cannot be an empty string. Please enter a valid value.")
	}
	return value, nil
}

func getAction() (string, string, string, error) {
	var value string

	actionToDo, err := establishActionToDo()
	if err != nil {
		return "", "", "", fmt.Errorf("error estabilishing action")
	}
	//Devo prima di tutto andare a prendere la key per la ricerca nel datastore per qualsiasi configurazione
	key, err := insertKey()
	if actionToDo == "put" {
		//Se l'utente ha scelto di inserire un elemento nel datastore, deve inoltre fornirmi
		//il valore da associare alla Key, che non può essere vuoto
		value, err = insertValue()
		if err != nil {
			return "", "", "", fmt.Errorf("error inserting action")
		}
	}
	return actionToDo, key, value, nil
}

func getConsistency() (int, string, error) {
	var consistencyType int
	for {
		fmt.Print("Enter the consistency type (0 for Causal, 1 for Sequential): ")
		_, err := fmt.Scan(&consistencyType)
		if err != nil {
			return -1, "", fmt.Errorf("error reading input: %w", err)
		}

		switch consistencyType {
		case 0:
			return 0, "Causal", nil
		case 1:
			return 1, "Sequential", nil
		default:
			log.Println("Invalid input. Please enter 0 for Causal or 1 for Sequential.")
		}
	}
}

func switchAction(actionToDo, key, value string, consistency int, serverNumber int, config clientCommon.Config, client *rpc.Client) {

	switch actionToDo {
	case "not specified":
		log.Printf("When you call the datastore you have to specify the action with -a (action): \n-a put with key and value for a put operation, \n-a get with the key for a get operation, \n-a delete with the key for a delete operation\n")
		os.Exit(-1)

	case "put":
		err := clientCommon.HandlePutAction(consistency, key, value, config, serverNumber, client)
		if err != nil {
			log.Fatal(err)
		}
	case "delete":
		err := clientCommon.HandleDeleteAction(consistency, key, config, serverNumber, client)
		if err != nil {
			log.Fatal(err)
		}
	case "get":
		err := clientCommon.HandleGetAction(consistency, key, client)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Printf("Uncorrect flag")
		return
	}

}

func main() {

	//time.Sleep(10 * time.Second) //Attendo che i server siano pronti

	consistency, consistencyString, err := getConsistency()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("The client is using server with the consistency: ", consistencyString)
	configuration := clientCommon.LoadEnvironment()

	config, err := clientCommon.LoadConfig(configuration)
	if err != nil {
		log.Fatal(err)
	}
	serverNumber, client, err := createCasualClient(config)
	if err != nil {
		log.Fatal(err)
	}
	//Voglio garantire trasparenza al client riguardo al tipo di consistenza che voglio implementare,
	//eseguo quindi una chiamata RPC a un server per sapere la modalità con cui il server è stato avviato

	for {
		actionToDo, key, value, err := getAction()
		if err != nil {
			log.Fatal(err)
		}
		//In base all'azione che l'utente ha scelto di svolgere, vado a chiamare la funzione corrispondente
		switchAction(actionToDo, key, value, consistency, serverNumber, config, client)
		//Chiedo all'utente se vuole continuare a svolgere operazioni sul datastore

		var continueRunning string
		fmt.Print("Do you want to continue? (yes/no): ")
		_, err = fmt.Scan(&continueRunning)
		if err != nil {
			log.Fatal(err)
		}
		continueRunning = strings.ToLower(continueRunning)
		if continueRunning != "yes" {
			clientCommon.CloseClient(client)
			break
		}
	}
}
