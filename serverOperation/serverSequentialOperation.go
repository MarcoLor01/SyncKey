package serverOperation

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
)

func (s *Server) AddElement(message Message, response *Response) error {
	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	s.MyScalarClock++
	message.ScalarTimestamp = s.MyScalarClock //The timestamp of the message is my clock
	message.ServerId = MyId

	reply := &Response{Done: false, Deliverable: false}
	s.sendToOtherServers(message, reply)
	fmt.Println("Deliverable: ", reply.Deliverable)
	fmt.Println("Done: ", reply.Done)
	response.Deliverable = reply.Deliverable
	response.Done = reply.Done
	return nil
}

func (s *Server) sendToOtherServers(message Message, response *Response) {

	ch := make(chan bool, len(addresses.Addresses))
	var wg sync.WaitGroup
	for _, address := range addresses.Addresses {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			client, err := rpc.Dial("tcp", addr)
			if err != nil {
				log.Fatal("Error in sendToOtherServers function: ", err)
			}
			reply := &Response{Done: false, Deliverable: false}
			if err1 := client.Call("Server.SaveElement", message, reply); err1 != nil {
				log.Fatal("Error in sendToOtherServers function: ", err1)
			}
			select {
			case ch <- reply.Deliverable: //Send the response to the channel
			default:
				fmt.Println("Channel was closed before it could be sent")
			}

			response.Done = reply.Done
		}(address.Addr)
	}
	wg.Wait()
	close(ch)
	response.Deliverable = s.checkResponses(ch)
}

func (s *Server) SaveElement(message Message, result *Response) error {

	s.lockIfNeeded(message.ServerId)
	if message.ServerId != MyId {
		if message.ScalarTimestamp > s.MyScalarClock {
			s.MyScalarClock = message.ScalarTimestamp
		}
		s.MyScalarClock++
	}
	s.addToQueue(message)
	messageAck := AckMessage{Element: message, MyServerId: MyId}
	var wg sync.WaitGroup
	for _, address := range addresses.Addresses {
		fmt.Printf("Adding in SaveElement")
		wg.Add(1)
		ackSentCounter := 0

		go func(addr string) {
			defer wg.Done()
			for {

				client, err := rpc.Dial("tcp", addr)
				if err != nil {
					log.Fatal("Error dialing: ", err)
				}
				reply := &Response{Done: false, Deliverable: false} //Crea una funzione che lo fa
				if err1 := client.Call("Server.SendAck", messageAck, reply); err1 != nil {
					log.Fatalf("Error sending ack at server %d: %v", message.ServerId, err1)
				}
				if reply.Done {
					result.Done = reply.Done
					result.Deliverable = reply.Deliverable
					break
				}
				if !reply.Done {
					ackSentCounter++
					if ackSentCounter == 5 {
						break
					}
				}
			}
		}(address.Addr)
	}
	wg.Wait()
	return nil
}

func (s *Server) SendAck(message AckMessage, reply *Response) error {

	s.lockIfNeeded(message.Element.ServerId)
	for _, myValue := range s.LocalQueue {
		if message.Element.Value == myValue.Value &&
			message.Element.Key == myValue.Key &&
			message.Element.ScalarTimestamp == myValue.ScalarTimestamp {
			fmt.Printf("I receive ACK from server %d\n", message.MyServerId)
			myValue.numberAck++
			if s.LocalQueue[0] == myValue && myValue.numberAck == len(addresses.Addresses) {
				if message.Element.OperationType == 1 {
					s.removeFromQueue(*myValue)
					s.printDataStore()
					reply.Deliverable = true //The element is deliverable
				}
				if message.Element.OperationType == 2 {
					s.removeFromQueueDeleting(*myValue)
					s.printDataStore()
					reply.Deliverable = true
				}
			}
			reply.Done = true //I have received the ack
			return nil
		}
	}
	reply.Done = false
	reply.Deliverable = false
	return nil
}

func (s *Server) GetElement(key string, reply *string) error {
	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	if value, ok := s.DataStore[key]; ok {
		*reply = value
		fmt.Println("Element found")
		return nil
	}
	return nil
}
