package serverOperation

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
)

func (s *Server) SequentialSendElement(message Message, response *Response) error {
	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	s.MyScalarClock++
	message.ScalarTimestamp = s.MyScalarClock
	message.ServerId = MyId

	reply := &Response{Done: false, Deliverable: false}
	err := s.sendToOtherServers(message, reply)
	if err != nil {
		return fmt.Errorf("SequentialSendElement: error sending to other servers: %v", err)
	}
	response.Deliverable = reply.Deliverable
	fmt.Println("RISULTATO FINALE SECONDO: ", response.Deliverable)
	return nil
}

func (s *Server) sendToOtherServers(message Message, response *Response) error {

	var wg sync.WaitGroup
	for _, address := range addresses.Addresses {
		wg.Add(1)
		go s.sequentialSendToSingleServer(address.Addr, message, response, &wg)
	}
	wg.Wait()

	//Se il numero di server che hanno consegnato il messaggio all'applicazione è uguale al numero di server presenti
	//Allora ritorno true, altrimenti false

	response.Deliverable = response.DeliverableServerNumber == len(addresses.Addresses)
	fmt.Println("RISULTATO FINALE: ", response.Deliverable)
	return nil
}

func (s *Server) sequentialSendToSingleServer(addr string, message Message, response *Response, wg *sync.WaitGroup) {
	defer wg.Done()
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Fatal("Error in sendToSingleServer function: ", err)
	}

	defer func(client *rpc.Client) {
		err1 := client.Close()
		if err1 != nil {
			log.Fatal("Error in closing connection")
		}

	}(client)

	reply := &Response{Done: false, Deliverable: false}
	if err1 := client.Call("Server.SequentialSendAck", message, reply); err1 != nil {
		log.Fatal("Error in sendToSingleServer function: ", err1)
	}

	fmt.Println("Numero di server che hanno consegnato: ", reply.DeliverableServerNumber)
	response.DeliverableServerNumber += reply.DeliverableServerNumber
	//Riceverò la risposta dal server che sarà negativa per i primi 4 ACK e che deve diventare
	//Positiva all'ultimo ACK che sancisce la consegna del messaggio all'applicazione
	response.Done = reply.Done
}

func (s *Server) SequentialSendAck(message Message, result *Response) error {
	s.lockIfNeeded(message.ServerId)
	if message.ServerId != MyId {
		if message.ScalarTimestamp > s.MyScalarClock {
			s.MyScalarClock = message.ScalarTimestamp
		}
		s.MyScalarClock++
	}
	s.addToQueue(message)
	messageAck := AckMessage{Element: message, MyServerId: MyId}
	ch := make(chan bool, len(addresses.Addresses))
	var wg sync.WaitGroup
	for _, address := range addresses.Addresses {
		wg.Add(1)
		go s.sequentialSendAckToSingleServer(address.Addr, messageAck, result, &wg, ch)
	}
	wg.Wait()
	close(ch)

	//Raccolgo risultati delle chiamate, ovvero il numero di server che
	//hanno correttamente consegnato il messaggio all'applicazione

	result.DeliverableServerNumber = s.checkSequentialResponses(ch)
	return nil
}

func (s *Server) sequentialSendAckToSingleServer(addr string, messageAck AckMessage, result *Response, wg *sync.WaitGroup, ch chan bool) {
	defer wg.Done()
	for {
		client, err := rpc.Dial("tcp", addr)
		if err != nil {
			log.Fatal("Error dialing: ", err)
		}
		reply := &Response{Done: false, Deliverable: false}
		if err1 := client.Call("Server.SequentialCheckingAck", messageAck, reply); err1 != nil {
			log.Fatalf("Error sending ack at server %d: %v", messageAck.Element.ServerId, err1)
		}
		if reply.Deliverable {
			//Un server, riceverà tutti i true, sarà il server che invierà l'ultimo
			//ACK necessario per la consegna all'applicazione del messaggio,
			//Avverto quindi il chiamante di ciò settando a true il risultato della mia chiamata nel campo Deliverable
			select {
			case ch <- reply.Deliverable:
			default:
				log.Fatal("Channel was closed before it could be sent")
			}
		}
		if reply.Done {
			result.Done = reply.Done
			break
		}
	}
}

func (s *Server) SequentialCheckingAck(message AckMessage, reply *Response) error {
	s.lockIfNeeded(message.Element.ServerId)
	for _, myValue := range s.LocalQueue {
		if message.Element.Value == myValue.Value &&
			message.Element.Key == myValue.Key &&
			message.Element.ScalarTimestamp == myValue.ScalarTimestamp {
			fmt.Printf("I receive ACK from server %d\n", message.MyServerId)
			myValue.numberAck++
			if s.LocalQueue[0] == myValue && myValue.numberAck == len(addresses.Addresses) {
				s.processAckMessage(message, myValue, reply) //Viene settato il valore di reply.deliverable pari a true
			}
			reply.Done = true //Ho ricevuto questo ACK
			return nil
		}
	}
	reply.Done = false
	reply.Deliverable = false
	return nil
}

func (s *Server) processAckMessage(message AckMessage, myValue *Message, reply *Response) {
	if message.Element.OperationType == 1 {
		s.removeFromQueue(*myValue)
		s.printDataStore()
		reply.Deliverable = true
	}
	if message.Element.OperationType == 2 {
		s.removeFromQueueDeleting(*myValue)
		s.printDataStore()
		reply.Deliverable = true
	}
}

func (s *Server) SequentialGetElement(key string, reply *string) error {
	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	if value, ok := s.DataStore[key]; ok {
		*reply = value
		fmt.Println("Element found")
		return nil
	}
	return nil
}
