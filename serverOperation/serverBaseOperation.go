package serverOperation

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"main/common"
	"math/rand"
	"net/rpc"
	"os"
	"sync"
	"time"
)

var MyId int //ID del server
var addresses ServerInformation
var serverReplicas = len(addresses.Addresses)

type ServerInformation struct {
	Addresses []ServerAddress `json:"address"`
}

type ServerBase struct {
	myClientMessage  common.ClientMessage //Messaggi del client
	myClientMutex    sync.Mutex           //Mutex per la sincronizzazione dell'accesso ai messaggi del client
	DataStore        map[string]string    //Il mio datastore
	myDatastoreMutex sync.Mutex           //Mutex per l'accesso al Datastore
}

type ServerAddress struct {
	Addr string `json:"addr"`
	Id   int    `json:"id"`
} //Struttura che contiene l'indirizzo e l'id di un server

//Inizializziamo la lista dei server, funzione chiamata durante la configurazione del server

func InitializeServerList() error {

	config := os.Getenv("CONFIG")

	var filePath string

	if config == "1" {
		filePath = "../serversAddrLocal.json"
	} else if config == "2" {
		filePath = "./serversAddrDocker.json"
	} else {
		log.Fatalf("Error loading the configuration file: CONFIG is set to '%s'", config)
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading configuration file: %w", err)
	}

	err = json.Unmarshal(fileContent, &addresses)
	if err != nil {
		return fmt.Errorf("error unmarshalling file: %w", err)
	}
	return nil
}

//Funzioni per stampare il contenuto del Datastore,
//La prima per la consistenza causale
//La seconda per la consistenza sequenziale

func (s *ServerBase) printDataStore() {
	time.Sleep(3 * time.Millisecond)
	fmt.Printf("\n\n---------------DATASTORE---------------\n")
	for key, value := range s.DataStore {
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}
	fmt.Printf("\n---------------------------------------\n")
}

//Funzione utilizzata per la generazione di un ID univoco per i messaggi

func generateUniqueID() string {
	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()
	ID := currentTimestamp + int64(uniqueID)
	stringID := fmt.Sprintf("%d", ID)
	return stringID
}

//Funzione utilizzata per la chiusura del client al termine dell'utilizzo

func closeClient(client *rpc.Client) {
	err := client.Close()
	if err != nil {
		log.Println("Error closing the client:", err)
	}
}

//Funzione per calcolare un delay casuale, compreso tra 500 ms e 3 secondi

func calculateDelay() int {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Genera un numero casuale compreso tra 0 e 2500 millisecondi.
	randomDelay := r.Intn(2500)
	// Aggiungi il ritardo minimo di 500 millisecondi.
	delay := randomDelay + 500
	return delay
}

func (s *ServerBase) InitializeMessageClient(clientId int) {
	s.myClientMessage = common.ClientMessage{
		ClientId:            clientId, // Inserisci l'ID del server appropriato qui
		ActualNumberMessage: 1,        // Inserisci il numero del messaggio appropriato qui
	}
}

//Controllo se i messaggi arrivano in ordine rispettando l'ordinamento FIFO, in caso contrario, attendo
//L'arrivo dei/del messaggi/o che lo precede/precedono

func (s *ServerBase) canProcess(message *common.Message, reply *common.Response) error {
	for {
		if message.IdMessageClient == s.myClientMessage.ClientId {
			if message.IdMessage == s.myClientMessage.ActualNumberMessage {
				reply.Done = true
				s.myClientMessage.ActualNumberMessage++
				return nil
			} else {
				time.Sleep(1 * time.Second)
			}
		} else {
			return fmt.Errorf("I can't handle message from this Client")
		}
	}
}
