package serverOperation

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"sort"
	"time"
)

var MyId int //ID del server
var addresses ServerInformation

type ServerInformation struct {
	Addresses []ServerAddress `json:"address"`
}

type ServerAddress struct {
	Addr string `json:"addr"`
	Id   int    `json:"id"`
} //Struttura che contiene l'indirizzo e l'id di un server

//Funzione per l'aggiunta in coda dei messaggi, i messaggi vengono ordinati in base al timestamp
//scalare, lo usiamo per l'implementazione della consistenza sequenziale
//In caso di timestamp scalare uguale, viene considerato il timestamp di inserimento in coda,
//Quindi i messaggi aggiunti prima saranno considerati prima all'interno della coda

func (s *ServerSequential) addToQueueSequential(message MessageSequential) {
	s.myQueueMutex.Lock()
	defer s.myQueueMutex.Unlock()
	message.InsertQueueTimestamp = time.Now().UnixNano()

	for i, element := range s.LocalQueue {
		if message.Key == element.Key {
			s.LocalQueue[i] = &message
			s.orderQueue()
			return
		}
	}

	s.LocalQueue = append(s.LocalQueue, &message)
	s.orderQueue()
}

func (s *ServerSequential) orderQueue() {
	sort.Slice(s.LocalQueue, func(i, j int) bool {
		if s.LocalQueue[i].ScalarTimestamp != s.LocalQueue[j].ScalarTimestamp {
			return s.LocalQueue[i].ScalarTimestamp < s.LocalQueue[j].ScalarTimestamp
		}
		return s.LocalQueue[i].InsertQueueTimestamp < s.LocalQueue[j].InsertQueueTimestamp
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

//Funzione per l'eliminazione di un messaggio dalla coda nel caso di consistenza causale

func (s *ServerCausal) removeFromQueueCausal(message MessageCausal) error {
	var isHere bool
	for i, msg := range s.LocalQueue {
		if message.IdUnique == msg.IdUnique {
			s.DataStore[msg.Key] = msg.Value
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			isHere = true
			break
		}
	}
	if isHere != true {
		return fmt.Errorf("message not in queue")
	}
	return nil
}

//Funzione per la rimozione di un messaggio dalla coda nel caso di operazione di Delete nella consistenza causale

func (s *ServerCausal) removeFromQueueDeletingCausal(message MessageCausal) error {
	var isHere bool
	for i, msg := range s.LocalQueue {
		if message.IdUnique == msg.IdUnique {
			delete(s.DataStore, msg.Key)
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			isHere = true
			break
		}
	}
	if isHere != true {
		return fmt.Errorf("message not in queue")
	}
	return nil
}

//Funzione per la rimozione di un messaggio dalla coda nel caso della consistenza sequenziale

func (s *ServerSequential) removeFromQueueSequential(message MessageSequential) {
	for i, msg := range s.LocalQueue {
		if message.Key == msg.Key && message.Value == msg.Value && message.ScalarTimestamp == msg.ScalarTimestamp {
			s.DataStore[msg.Key] = msg.Value
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

//Funzione per la rimozione di un messaggio dalla coda per l'operazione di Delete nel caso della consistenza sequenziale

func (s *ServerSequential) removeFromQueueDeletingSequential(message MessageSequential) {
	for i, msg := range s.LocalQueue {
		if message.Key == msg.Key && message.Value == msg.Value && message.ScalarTimestamp == msg.ScalarTimestamp {
			delete(s.DataStore, msg.Key)
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
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

//Non voglio che l'utente debba aspettare che tutti i server abbiano inserito il messaggio nel datastore,
//questo perché alcuni server potrebbero star aspettando il messaggio successivo per poterlo salvare

func (s *ServerCausal) checkResponses(ch chan bool) bool {
	counter := 0

	for response := range ch {
		fmt.Println("Risultato: ", response)
		if response {
			counter++
		}
	}
	if counter != 0 {
		return true
	} else {
		return false
	}
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

//Funzione che esegue ulteriori controlli ed elimina il primo termine dalla coda locale del server

func (s *ServerSequential) updateQueue(message MessageSequential, reply *ResponseSequential) {
	s.myQueueMutex.Lock()
	if len(s.LocalQueue) != 0 && s.LocalQueue[0].IdUnique == message.IdUnique {
		s.LocalQueue = append(s.LocalQueue[:0], s.LocalQueue[1:]...)
		reply.Done = true
	} else {
		reply.Done = false
	}
	s.myQueueMutex.Unlock()
}

//Funzione per aggiungere un nuovo messaggio al datastore

func (s *ServerSequential) addElementDatastore(message MessageSequential) {
	s.myDatastoreMutex.Lock()
	log.Printf("ESEGUITA DA SERVER %d azione di put, key: %s, value: %s\n", MyId, message.Key, message.Value)
	s.DataStore[message.Key] = message.Value
	s.printDataStore()
	s.myDatastoreMutex.Unlock()
}

//Funzione per la rimozione di un messaggio dal datastore

func (s *ServerSequential) deleteElementDatastore(message MessageSequential) {
	s.myDatastoreMutex.Lock()
	log.Printf("ESEGUITA DA SERVER %d azione di delete, key: %s\n", MyId, message.Key)
	delete(s.DataStore, message.Key)
	s.printDataStore()
	s.myDatastoreMutex.Unlock()
}

//Funzione per creazione di un messaggio di ACK

func (s *ServerSequential) createAckMessage(Message MessageSequential) AckMessage {
	return AckMessage{
		Element:    Message,
		MyServerId: MyId,
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
