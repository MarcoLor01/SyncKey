package serverOperation

import (
	"encoding/json"
	"log"
	"os"
	"sort"
)

//Two Function for ordering my server queue: the first for the causal consistency and the other for the sequential consistency

func (s *Server) addToQueue(message Message) {
	var insertIndex int

	// Trova l'indice in cui inserire il messaggio
	for i, element := range s.localQueue {
		if message.Key == element.Key {
			s.localQueue[i] = &message // Aggiungi il messaggio alla coda
			return
		}
		if message.ScalarTimestamp >= element.ScalarTimestamp {
			insertIndex = i + 1
		} else {
			break
		}
	}

	// Inserisci il messaggio nell'indice trovato
	s.localQueue = append(s.localQueue[:insertIndex], append([]*Message{&message}, s.localQueue[insertIndex:]...)...)

	// Ordina la coda in base al timestamp scalare
	sort.Slice(s.localQueue, func(i, j int) bool {
		return s.localQueue[i].ScalarTimestamp < s.localQueue[j].ScalarTimestamp
	})

}

//Initialize the list of the server in the configuration file

func InitializeServerList() {
	fileContent, err := os.ReadFile("serversAddr.json")
	if err != nil {
		log.Fatal("Error reading configuration file: ", err)
	}

	err = json.Unmarshal(fileContent, &addresses)
	if err != nil {
		log.Fatal("Error unmarshalling file: ", err)
	}
}

func (s *Server) removeFromQueue(message Message) {

	i := 0
	for _, msg := range s.localQueue {

		if message.Key == msg.Key && message.Value == msg.Value && message.ScalarTimestamp == msg.ScalarTimestamp {

			//I'm going to remove this message from my queue
			s.dataStore[msg.Key] = msg.Value                               //Insert message in my DS
			s.localQueue = append(s.localQueue[:i], s.localQueue[i+1:]...) //Remove
			break
		}
		i++
	}

}

func (s *Server) removeFromQueueDeleting(message Message) {
	i := 0
	for _, msg := range s.localQueue {

		if message.Key == msg.Key && message.Value == msg.Value && message.ScalarTimestamp == msg.ScalarTimestamp {

			//I'm going to remove this message from my queue
			delete(s.dataStore, msg.Key)                                   //Insert message in my DS
			s.localQueue = append(s.localQueue[:i], s.localQueue[i+1:]...) //Remove
			break
		}
		i++
	}
}
