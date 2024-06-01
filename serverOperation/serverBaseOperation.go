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

type ServerInformation struct {
	Addresses []ServerAddress `json:"address"`
}

type ServerBase struct {
	myClientMessage  []common.ClientMessage //Per ogni client voglio memorizzare quanti messaggi mi arrivano e quanti ho processato
	myClientMutex    []sync.Mutex           //Mutex per la sincronizzazione dell'accesso ai messaggi del client
	DataStore        map[string]string      //Il mio datastore
	myDatastoreMutex sync.Mutex             //Mutex per l'accesso al Datastore
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

func GetLength() int {
	return len(addresses.Addresses)
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

func (s *ServerBase) InitializeMessageClients(numberClients int) {

	// Inizializza la slice con la dimensione corretta
	s.myClientMessage = make([]common.ClientMessage, numberClients)
	s.myClientMutex = make([]sync.Mutex, numberClients)

	// Inizializza ogni elemento della slice
	for i := 0; i < numberClients; i++ {
		s.myClientMessage[i] = common.ClientMessage{
			ClientId:            i + 1, // Usa l'indice come client ID
			ActualNumberMessage: 1,     // Imposta il valore iniziale appropriato
			ActualAnswerMessage: 1,     // Imposta il valore iniziale appropriato
		}
	}
}

//Controllo se i messaggi arrivano in ordine rispettando l'ordinamento FIFO, in caso contrario, attendo
//L'arrivo dei/del messaggi/o che lo precede/precedono

func (s *ServerBase) canProcess(message *common.Message, reply *common.Response) error {

	s.myClientMutex[message.IdMessageClient-1].Lock()

	for {
		if message.IdMessage == s.myClientMessage[message.IdMessageClient-1].ActualNumberMessage {
			s.myClientMessage[message.IdMessageClient-1].ActualNumberMessage++

			reply.Done = true
			s.myClientMutex[message.IdMessageClient-1].Unlock()
			return nil
		} else {
			// Rilascia temporaneamente il lockClockMutex prima di dormire
			s.myClientMutex[message.IdMessageClient-1].Unlock()
			time.Sleep(1 * time.Second)
			s.myClientMutex[message.IdMessageClient-1].Lock()
		}
	}
}

func (s *ServerBase) canAnswer(message *common.Message, reply *common.Response) error {
	s.myClientMutex[message.IdMessageClient-1].Lock()
	defer s.myClientMutex[message.IdMessageClient-1].Unlock()
	for {
		if message.IdMessage == s.myClientMessage[message.IdMessageClient-1].ActualAnswerMessage {
			reply.Done = true
			s.myClientMessage[message.IdMessageClient-1].ActualAnswerMessage++
			return nil

		} else {
			log.Println("BLOCCATO QUI MEX: ", message.IdMessage, "DA: ", message.IdMessageClient, " ACTUAL: ", s.myClientMessage[message.IdMessageClient-1].ActualAnswerMessage, " CLIENT: ", message.IdMessage)
			// Rilascia temporaneamente il lockClockMutex prima di dormire
			s.myClientMutex[message.IdMessageClient-1].Unlock()
			time.Sleep(500 * time.Millisecond)
			s.myClientMutex[message.IdMessageClient-1].Lock()
		}
	}
}
