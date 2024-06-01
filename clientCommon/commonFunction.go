package clientCommon

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"main/common"
	"net/rpc"
	"os"
)

const datastoreError string = "error adding the element to datastore, error: "
const datastoreDeleteError string = "error deleting the element from the datastore, error: "
const datastoreGetError string = "error getting the element from the datastore, error: "

type Config struct {
	Address []struct {
		ID   int    `json:"id"`
		Addr string `json:"addr"`
	} `json:"address"`
}

// Operation Definizione del tipo di enum per le operazioni
type Operation int

// Definizione delle costanti per le operazioni
const (
	Put Operation = iota + 1
	Delete
	Get
)

// Mappa per associare i valori numerici alle stringhe
var operationToString = map[Operation]string{
	Put:    "PUT",
	Delete: "DELETE",
	Get:    "GET",
}

// GetOperationString Funzione per ottenere la rappresentazione testuale dell'operazione

func GetOperationString(op Operation) string {
	if str, ok := operationToString[op]; ok {
		return str
	}
	return "unknown"
}

//Funzione per la gestione delle azioni

//Azione di put

func HandlePutAction(consistency int, key string, value string, config Config, serverNumber int, client *rpc.Client, idMessageClient int, idMessage int) error {
	if consistency == 0 {
		err := addElementToDsCausal(key, value, config, serverNumber, client, idMessageClient, idMessage)
		if err != nil {
			return fmt.Errorf(datastoreError+": %w", err)
		}
	} else {
		err := addElementToDsSequential(key, value, config, serverNumber, client, idMessageClient, idMessage)
		if err != nil {
			return fmt.Errorf(datastoreError+": %w", err)
		}
	}
	return nil
}

//Azione di delete

func HandleDeleteAction(consistency int, key string, config Config, serverNumber int, client *rpc.Client, idMessageClient int, idMessage int) error {
	if consistency == 0 {
		err := deleteElementFromDsCausal(key, config, serverNumber, client, idMessageClient, idMessage)
		if err != nil {
			return fmt.Errorf(datastoreDeleteError+": %w", err)
		}
	} else {
		err := deleteElementFromDsSequential(key, config, serverNumber, client, idMessageClient, idMessage)
		if err != nil {
			return fmt.Errorf(datastoreDeleteError+": %w", err)
		}
	}
	return nil
}

//Azione di get

func HandleGetAction(consistency int, key string, config Config, client *rpc.Client, idMessageClient int, idMessage int) error {
	if consistency == 0 {

		err := getElementFromDsCausal(key, config, client, idMessageClient, idMessage)
		if err != nil {
			return fmt.Errorf(datastoreGetError+": %w", err)
		}
	} else {
		err := getElementFromDsSequential(key, client, idMessageClient, idMessage)
		if err != nil {
			return fmt.Errorf(datastoreGetError+": %w", err)
		}
	}
	return nil
}

//Azione di put con consistenza sequenziale

func addElementToDsSequential(key string, value string, config Config, serverNumber int, client *rpc.Client, idMessageClient int, idMessage int) error {

	var n1, n2 string
	n1 = key
	n2 = value

	//printAdd(n1, n2, config, serverNumber)
	message := common.Message{Key: n1, Value: n2, OperationType: 1, IdMessageClient: idMessageClient, IdMessage: idMessage}
	args := common.MessageSequential{MessageBase: message}

	result := common.Response{Done: false}

	err := client.Call("ServerSequential.SequentialSendElement", args, &result)
	if err != nil {
		return fmt.Errorf(datastoreError+": %w", err)
	}

	if result.Done {
		printPutSuccessful(args.MessageBase.Key, args.MessageBase.Value, message.IdMessageClient)
	} else {
		printPutFail(args.MessageBase.Key, args.MessageBase.Value, message.IdMessageClient)
	}
	return nil
}

//Funzione per aggiungere un elemento con consistenza causale

func addElementToDsCausal(key string, value string, config Config, serverNumber int, client *rpc.Client, idMessageClient int, idMessage int) error {

	var n1, n2 string
	n1 = key
	n2 = value
	message := common.Message{Key: n1, Value: n2, OperationType: 1, IdMessageClient: idMessageClient, IdMessage: idMessage}
	args := common.MessageCausal{MessageBase: message, VectorTimestamp: make([]int, len(config.Address))}

	result := common.Response{Done: false}

	err := client.Call("ServerCausal.CausalSendElement", args, &result) //Calling the CausalSendElement routine
	if err != nil {
		return fmt.Errorf(datastoreError+": %w", err)
	}

	if result.Done {
		printPutSuccessful(args.MessageBase.Key, args.MessageBase.Value, message.IdMessageClient)
	} else {
		printPutFail(args.MessageBase.Key, args.MessageBase.Value, message.IdMessageClient)
	}

	return nil
}

//Funzione per aggiungere un elemento con consistenza sequenziale

func deleteElementFromDsSequential(key string, config Config, serverNumber int, client *rpc.Client, idMessageClient int, idMessage int) error {
	n1 := key

	log.Printf("Get element with Key: %s contacting %s", n1, config.Address[serverNumber].Addr)
	message := common.Message{Key: n1, OperationType: 2, IdMessageClient: idMessageClient, IdMessage: idMessage}
	args := common.MessageSequential{MessageBase: message}
	result := common.Response{Done: false}

	err := client.Call("ServerSequential.SequentialSendElement", args, &result)

	if err != nil {
		return fmt.Errorf(datastoreError+": %w", err)
	}

	if result.Done {
		printKeySuccessful(args.MessageBase.Key, GetOperationString(Operation(message.IdMessageClient)), message.IdMessageClient)
	} else {
		printKeyFail(args.MessageBase.Key, GetOperationString(Operation(message.IdMessageClient)), message.IdMessageClient)
	}
	return nil
}

//Funzione per eliminare un elemento con consistenza causale

func deleteElementFromDsCausal(key string, config Config, serverNumber int, client *rpc.Client, idMessageClient int, idMessage int) error {
	n1 := key

	log.Printf("Get element with Key: %s contacting %s\n", n1, config.Address[serverNumber].Addr)
	message := common.Message{Key: n1, OperationType: 2, IdMessageClient: idMessageClient, IdMessage: idMessage}
	args := common.MessageCausal{MessageBase: message, VectorTimestamp: make([]int, len(config.Address))}
	result := common.Response{Done: false}

	err := client.Call("ServerCausal.CausalSendElement", args, &result)

	if err != nil {
		return fmt.Errorf(datastoreError+": %w", err)
	}

	if result.Done {
		printKeySuccessful(args.MessageBase.Key, GetOperationString(Operation(message.IdMessageClient)), message.IdMessageClient)
	} else {
		printKeyFail(args.MessageBase.Key, GetOperationString(Operation(message.IdMessageClient)), message.IdMessageClient)
	}
	return nil
}

//Funzione per recuperare un elemento in versione causale

func getElementFromDsCausal(key string, config Config, client *rpc.Client, idMessageClient int, idMessage int) error {

	message := common.Message{Key: key, OperationType: 3, IdMessageClient: idMessageClient, IdMessage: idMessage}
	args := common.MessageCausal{MessageBase: message, VectorTimestamp: make([]int, len(config.Address))}
	reply := common.Response{Done: false}
	err := client.Call("ServerCausal.CausalSendElement", args, &reply)

	fmt.Println("Get element with Key: ", key)

	if err != nil {
		return fmt.Errorf(datastoreError+": %w", err)
	}

	if reply.Done != false {
		printKeySuccessful(key, GetOperationString(Operation(message.OperationType)), idMessageClient)
	} else {
		printKeyFail(key, GetOperationString(Operation(message.OperationType)), idMessageClient)
	}
	return nil
}

//Funzione per recuperare un elemento in versione sequenziale

func getElementFromDsSequential(key string, client *rpc.Client, idMessageClient int, idMessage int) error {
	message := common.Message{
		Key:             key,
		IdMessageClient: idMessageClient,
		IdMessage:       idMessage,
		OperationType:   3,
	}
	reply := common.Response{Done: false}
	args := common.MessageSequential{MessageBase: message}
	err := client.Call("ServerSequential.SequentialSendElement", args, &reply)
	if err != nil {
		return fmt.Errorf(datastoreError+": %w", err)
	}
	if reply.Done != false {
		printKeySuccessful(key, GetOperationString(Operation(message.OperationType)), idMessageClient)
	} else {
		printKeyFail(key, GetOperationString(Operation(message.OperationType)), idMessageClient)
	}
	return nil
}

//Funzione per effettuare Dialing verso un server

func CreateClient(config Config, serverNumber int) (*rpc.Client, error) {

	client, err := rpc.Dial("tcp", config.Address[serverNumber].Addr)
	if err != nil {
		return nil, fmt.Errorf("error in dialing: %w", err)
	}
	return client, nil
}

func LoadConfig(configuration string) (Config, error) {

	filePath := map[string]string{"1": "../serversAddrLocal.json", "2": "../serversAddrDocker.json"}[configuration]
	if filePath == "" {
		log.Fatalf("Error loading the configuration file: CONFIG is set to '%s'", configuration)
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading file: %w", err)
	}

	var config Config
	if err = json.Unmarshal(fileContent, &config); err != nil {
		log.Fatal("Error in JSON decode:", err)
	}
	return config, nil
}

//Funzione per caricare le variabili d'ambiente dal file .env

func LoadEnvironment() string {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
	configuration := os.Getenv("CONFIG") //Carico il valore della variabile d'ambiente CONFIG,
	//che mi dice se sto usando la configurazione locale (CONFIG = 1) o quella docker (CONFIG = 2)
	return configuration
}

//Funzione per andare a chiudere la chiamata a un qualsiasi server, passato come input

func CloseClient(client *rpc.Client) {
	err := client.Close()
	if err != nil {
		log.Println("Error closing the client:", err)
	}
}

func printAdd(n1 string, n2 string, config Config, serverNumber int) {
	log.Printf("Adding element with Key: %s, and Value: %s contacting %s", n1, n2, config.Address[serverNumber].Addr)

}

func printPutSuccessful(key, value string, idServer int) {
	log.Println("ESEGUITA OPERAZIONE DI TIPO PUT SU SERVER: ", idServer, "CON: ", key, ":", value)
}

func printKeySuccessful(key string, operationType string, idServer int) {
	log.Println("ESEGUITA OPERAZIONE DI TIPO ", operationType, "SU SERVER: ", idServer, "CON:", key)
}

func printPutFail(key string, value string, idServer int) {
	log.Println("NON ESEGUITA OPERAZIONE DI TIPO PUT SU SERVER: ", idServer, "CON: ", key, ":", value)
}

func printKeyFail(key string, operationType string, idServer int) {
	log.Println("NON ESEGUITA OPERAZIONE DI TIPO ", operationType, "SU SERVER: ", idServer, "CON:", key)
}
