package serverOperation

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
)

func (s *Server) CausalSendElement(message Message, reply *Response) error {
	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	s.incrementMyTimestamp()
	message.prepareMessage(s.MyClock, MyId) //Setto il timestamp e il mio id come sender

	response := s.createResponse()
	s.sendToOtherServersCausal(message, response)
	reply.Deliverable = response.Deliverable
	return nil
}

func (s *Server) incrementMyTimestamp() {
	s.MyClock[MyId-1] += 1
	fmt.Println("Incrementing my timestamp...")
	fmt.Println("My actual timestamp: ", s.MyClock)
}

func (message *Message) prepareMessage(clock []int, id int) {
	message.VectorTimestamp = clock
	message.ServerId = id
}

func (s *Server) createResponse() *Response {
	return &Response{Deliverable: false}
}

func (s *Server) sendToOtherServersCausal(message Message, response *Response) {

	ch := make(chan bool, len(addresses.Addresses))
	var wg sync.WaitGroup

	for _, address := range addresses.Addresses {
		wg.Add(1)
		go s.sendToSingleServer(address, message, ch, &wg)
	}

	wg.Wait()
	close(ch)

	response.Deliverable = s.checkResponses(ch) //Check delle risposte
}

func (s *Server) sendToSingleServer(address ServerAddress, message Message, ch chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	client, err := rpc.Dial("tcp", address.Addr)
	if err != nil {
		log.Fatal("Connection RPC error:", err)
		return
	}
	defer func(client *rpc.Client) {
		err1 := client.Close()
		if err1 != nil {
			log.Fatal("Error in closing connection")
		}

	}(client)

	reply := &Response{Deliverable: false}

	if err1 := client.Call("Server.SaveElementCausal", message, reply); err1 != nil {
		log.Fatal("RPC call error:", err)
		return
	}

	select {
	case ch <- reply.Deliverable:
	default:
		log.Fatal("Channel was closed before it could be sent")
	}
}

func (s *Server) SaveElementCausal(message Message, reply *Response) error {
	s.lockIfNeeded(message.ServerId) //Lock del mutex se il serverId è diverso da MyId

	fmt.Println("The timestamp of the message is: ", message.VectorTimestamp)

	s.processMessages(message, reply)

	return nil
}

func (s *Server) lockIfNeeded(serverId int) {
	if serverId != MyId {
		s.myMutex.Lock()
		defer s.myMutex.Unlock()
	}
}

func (s *Server) processMessages(message Message, reply *Response) {
	var wg sync.WaitGroup
	s.addToQueue(message)
	wg.Add(len(s.LocalQueue))

	for i := len(s.LocalQueue) - 1; i >= 0; i-- {
		message2 := s.LocalQueue[i]
		go s.checkAndProcessMessage(*message2, reply, &wg)
	}
	wg.Wait()
}

func (s *Server) checkAndProcessMessage(message Message, reply *Response, wg *sync.WaitGroup) {
	defer wg.Done()

	if s.isMessageDeliverable(message) {
		s.processDeliverableMessage(message, reply)
	} else {
		fmt.Printf("The message is not deliverable: %d\n , My clock: %d\n", message.VectorTimestamp, s.MyClock)
	}
}

func (s *Server) isMessageDeliverable(message Message) bool {
	var mod bool

	//Se il server che ha inviato il messaggio è uguale al server che lo riceve
	//Controllo se il timestamp del messaggio è uguale al timestamp del server
	//In caso contrario controllo se il timestamp del messaggio è uguale al timestamp del server + 1
	//Perché se il messaggio è stato inviato dal server stesso, il timestamp sarà uguale al timestamp del server

	if MyId == message.ServerId {
		mod = message.VectorTimestamp[message.ServerId-1] == s.MyClock[message.ServerId-1]
	} else {
		mod = message.VectorTimestamp[message.ServerId-1] == s.MyClock[message.ServerId-1]+1
	}

	if mod {
		for index, ts := range message.VectorTimestamp {
			if index == message.ServerId-1 {
				continue
			}
			if ts > s.MyClock[index] { //If the timestamp is greater than the server's timestamp
				return false
			}
		}
		return true
	}
	return false
}

func (s *Server) processDeliverableMessage(message Message, reply *Response) {
	s.updateTimestamp(message)
	reply.Deliverable = true
	fmt.Println("The message is deliverable")
	if message.OperationType == 1 {
		s.removeFromQueue(message)
		s.printDataStore()
	}
	if message.OperationType == 2 {
		s.removeFromQueueDeletingCausal(message)
		s.printDataStore()
	}
	fmt.Println("My actual timestamp:", s.MyClock)
}

func (s *Server) updateTimestamp(message Message) {
	for ind, ts := range message.VectorTimestamp {
		if ts > s.MyClock[ind] {
			s.MyClock[ind] = ts
		}
	}
}

func (s *Server) GetElementCausal(key string, reply *string) error {
	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	if value, ok := s.DataStore[key]; ok {
		*reply = value
		fmt.Println("Element found")
		return nil
	}
	return nil
}
