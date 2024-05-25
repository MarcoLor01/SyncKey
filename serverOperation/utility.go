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
	"sort"
	"sync"
	"time"
)

var MyId int //ID del server
var addresses ServerInformation

type ServerInformation struct {
	Addresses []ServerAddress `json:"address"`
}

type ServerBase struct {
	myClientMessage common.ClientMessage //Messaggi del client
	myClientMutex   sync.Mutex           //Mutex per la sincronizzazione dell'accesso ai messaggi del client
}

type ServerAddress struct {
	Addr string `json:"addr"`
	Id   int    `json:"id"`
} //Struttura che contiene l'indirizzo e l'id di un server

//Funzione per l'aggiunta in coda dei messaggi, i messaggi vengono ordinati in base al timestamp
//scalare, lo usiamo per l'implementazione della consistenza sequenziale
//In caso di timestamp scalare uguale, viene considerato il timestamp di inserimento in coda,
//Quindi i messaggi aggiunti prima saranno considerati prima all'interno della coda

func (s *ServerSequential) orderQueue() {
	sort.Slice(s.LocalQueue, func(i, j int) bool {
		if s.LocalQueue[i].ScalarTimestamp != s.LocalQueue[j].ScalarTimestamp {
			return s.LocalQueue[i].ScalarTimestamp < s.LocalQueue[j].ScalarTimestamp
		}
		return s.LocalQueue[i].MessageBase.ServerId < s.LocalQueue[j].MessageBase.ServerId
	})
}

//Funzione per aggiungere il messaggio alla coda nel caso della consistenza causale

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

func (s *ServerCausal) printDataStore() {
	time.Sleep(3 * time.Millisecond)
	fmt.Printf("\n\n---------------DATASTORE---------------\n")
	for key, value := range s.DataStore {
		fmt.Printf("Key: %s, Value: %s\n", key, value)
	}
	fmt.Printf("\n---------------------------------------\n")
}

func (s *ServerSequential) printDataStore() {

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

func (s *ServerBase) canProcess(message *common.Message, reply *common.Response) error {
	for {
		if message.IdMessageClient == s.myClientMessage.ClientId {
			fmt.Println("Mi aspettavo numero: ", s.myClientMessage.ActualNumberMessage)
			fmt.Println("E ho ricevuto: ", message.IdMessage)
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
