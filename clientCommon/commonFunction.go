package clientCommon

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"main/serverOperation"
	"net/rpc"
	"os"
	"time"
)

const wait string = "Wait..."
const call string = "Synchronous call to RPC server"
const datastoreError string = "Error adding the element to datastore, error: "

type Config struct {
	Address []struct {
		ID   int    `json:"id"`
		Addr string `json:"addr"`
	} `json:"address"`
}

func HandlePutAction(consistency int, key string, value string, config Config, serverNumber int, client *rpc.Client) {
	if consistency == 0 {
		addElementToDsCausal(key, value, config, serverNumber, client)
	} else {
		addElementToDsSequential(key, value, config, serverNumber, client)
	}
}

func HandleDeleteAction(consistency int, key string, config Config, serverNumber int, client *rpc.Client) {
	if consistency == 0 {
		deleteElementFromDsCausal(key, config, serverNumber, client)
	} else {
		deleteElementFromDsSequential(key, config, serverNumber, client)
	}
}

func HandleGetAction(consistency int, key string, client *rpc.Client) {
	if consistency == 0 {
		getElementFromDsCausal(key, client)
	} else {
		getElementFromDsSequential(key, client)
	}
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

func CreateClient(config Config, serverNumber int) *rpc.Client {

	client, err := rpc.Dial("tcp", config.Address[serverNumber].Addr)
	if err != nil {
		log.Fatal("Error in dialing: ", err)
	}
	return client
}

func LoadConfig(configuration string) Config {

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

func LoadEnvironment() string {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
	configuration := os.Getenv("CONFIG") //Carico il valore della variabile d'ambiente CONFIG,
	//che mi dice se sto usando la configurazione locale (CONFIG = 1) o quella docker (CONFIG = 2)
	return configuration
}
