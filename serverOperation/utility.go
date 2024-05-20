package serverOperation

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/rpc"
	"os"
	"reflect"
	"sort"
	"time"
)

//Funzione per l'aggiunta in coda dei messaggi, i messaggi vengono ordinati in base al timestamp
//scalare, lo usiamo per l'implementazione della consistenza sequenziale
//In caso di timestamp scalare uguale, viene considerato il timestamp di inserimento in coda,
//Quindi i messaggi aggiunti prima saranno considerati prima all'interno della coda

func (s *ServerSequential) addToQueueSequential(message MessageSequential) {
	s.myQueueMutex.Lock()
	defer s.myQueueMutex.Unlock()
	message.InsertQueueTimestamp = time.Now().UnixNano()

	for i, element := range s.LocalQueue {
		if message.Key == element.MessageSeq.Key {
			s.LocalQueue[i].MessageSeq = &message
			s.orderQueue()
			return
		}
	}
	msgQueue := &MessageSeqQueue{
		MessageSeq: &message,
		Inserted:   false,
	}
	s.LocalQueue = append(s.LocalQueue, msgQueue)
	s.orderQueue()
}

func (s *ServerSequential) orderQueue() {
	sort.Slice(s.LocalQueue, func(i, j int) bool {
		if s.LocalQueue[i].MessageSeq.ScalarTimestamp != s.LocalQueue[j].MessageSeq.ScalarTimestamp {
			return s.LocalQueue[i].MessageSeq.ScalarTimestamp < s.LocalQueue[j].MessageSeq.ScalarTimestamp
		}
		return s.LocalQueue[i].MessageSeq.InsertQueueTimestamp < s.LocalQueue[j].MessageSeq.InsertQueueTimestamp
	})
}

//Funzione per aggiungere il messaggio alla coda nel caso della consistenza causale

func (s *ServerCausal) addToQueueCausal(message MessageCausal) {
	s.LocalQueue = append(s.LocalQueue, &message)
}

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

func (s *ServerCausal) removeFromQueueCausal(message MessageCausal) {
	for i, msg := range s.LocalQueue {
		if message.Key == msg.Key && message.Value == msg.Value && reflect.DeepEqual(message.VectorTimestamp, msg.VectorTimestamp) == true {
			s.DataStore[msg.Key] = msg.Value
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

//Funzione per la rimozione di un messaggio dalla coda nel caso di operazione di Delete nella consistenza causale

func (s *ServerCausal) removeFromQueueDeletingCausal(message MessageCausal) {
	for i, msg := range s.LocalQueue {
		if message.Key == msg.Key && message.Value == msg.Value && reflect.DeepEqual(message.VectorTimestamp, msg.VectorTimestamp) == true {
			delete(s.DataStore, msg.Key)
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

//Funzione per la rimozione di un messaggio dalla coda nel caso della consistenza sequenziale

func (s *ServerSequential) removeFromQueueSequential(message MessageSequential) {
	for i, msg := range s.LocalQueue {
		if message.Key == msg.MessageSeq.Key && message.Value == msg.MessageSeq.Value && message.ScalarTimestamp == msg.MessageSeq.ScalarTimestamp {
			s.DataStore[msg.MessageSeq.Key] = msg.MessageSeq.Value
			s.LocalQueue = append(s.LocalQueue[:i], s.LocalQueue[i+1:]...)
			break
		}
	}
}

//Funzione per la rimozione di un messaggio dalla coda per l'operazione di Delete nel caso della consistenza sequenziale

func (s *ServerSequential) removeFromQueueDeletingSequential(message MessageSequential) {
	for i, msg := range s.LocalQueue {
		if message.Key == msg.MessageSeq.Key && message.Value == msg.MessageSeq.Value && message.ScalarTimestamp == msg.MessageSeq.ScalarTimestamp {
			delete(s.DataStore, msg.MessageSeq.Key)
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
	if len(s.LocalQueue) != 0 && s.LocalQueue[0].MessageSeq.IdUnique == message.IdUnique {
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
