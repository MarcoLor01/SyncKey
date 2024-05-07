package serverOperation

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
)

func (s *Server) CausalSendElement(message Message, reply *Response) error { //Function that sends the element to the other servers
	s.myMutex.Lock()         //Lock the mutex
	defer s.myMutex.Unlock() //Unlock the mutex when the function ends

	s.incrementMyTimestamp()                //Increment the timestamp
	message.prepareMessage(s.MyClock, MyId) //Prepare the message

	response := s.createResponse()                //Create the response
	s.sendToOtherServersCausal(message, response) //Send the message to the other servers
	reply.Done = response.Done                    //Set the answer for the Client with the response.Done
	return nil
}

func (s *Server) incrementMyTimestamp() { //Function that increments the timestamp of the server
	s.MyClock[MyId-1] += 1
	fmt.Println("Incrementing my timestamp", s.MyClock)
}

func (message *Message) prepareMessage(clock []int, id int) { //Function that prepares the message to be sent
	message.VectorTimestamp = clock
	message.ServerId = id
}

func (s *Server) createResponse() *Response { //Function that creates the response
	return &Response{Done: false}
}

func (s *Server) sendToOtherServersCausal(message Message, response *Response) {
	ch := make(chan bool, len(addresses.Addresses)) //Create a channel with the length of the addresses
	var wg sync.WaitGroup                           //Create a WaitGroup

	for _, address := range addresses.Addresses {
		wg.Add(1)                                          //Add a new goroutine
		go s.sendToSingleServer(address, message, ch, &wg) //Send the message to the server
	}

	wg.Wait() //Wait for all the goroutines to finish
	close(ch) //Close the channel

	response.Done = s.checkResponses(ch) //Check the responses
}

func (s *Server) sendToSingleServer(address ServerAddress, message Message, ch chan bool, wg *sync.WaitGroup) { //Function that sends the message to a single server
	defer wg.Done() //Decrease the WaitGroup counter when the function ends

	//-----TESTING-------//
	//if MyId == 1 && address.Id == 3 {
	//	fmt.Println("I'm not sending to the server 3, sleeping...")
	//	time.Sleep(40 * time.Second)
	//}
	//-----TESTING-------//

	client, err := rpc.Dial("tcp", address.Addr) //Dial the server
	if err != nil {
		log.Fatal("Connection RPC error:", err)
		return
	}
	defer func(client *rpc.Client) {
		err1 := client.Close()
		if err1 != nil {
			log.Fatal("Error in closing connection")
		}

	}(client) //Close the connection when the function ends

	reply := &Response{Done: false} //Create a new response

	if err1 := client.Call("Server.SaveElementCausal", message, reply); err1 != nil { //Call the SaveElementCausal function
		log.Fatal("RPC call error:", err)
		return
	}

	select {
	case ch <- reply.Done: //Send the response to the channel
	default:
		fmt.Println("Channel was closed before it could be sent")
	}
}

func (s *Server) SaveElementCausal(message Message, reply *Response) error {
	s.lockIfNeeded(message.ServerId) //Lock the mutex if the serverId is different from MyId

	fmt.Println(message.VectorTimestamp) //Print the vector timestamp

	s.processMessages(message, reply) //Process the messages

	return nil
}

func (s *Server) lockIfNeeded(serverId int) {
	if serverId != MyId { //If the serverId is different from MyId
		s.myMutex.Lock()
		defer s.myMutex.Unlock()
	}
}

func (s *Server) processMessages(message Message, reply *Response) {
	var wg sync.WaitGroup
	s.addToQueue(message) //Add the message to the queue

	for i := len(s.LocalQueue) - 1; i >= 0; i-- {
		message2 := s.LocalQueue[i]
		fmt.Println("Message in the queue:", message2.VectorTimestamp)
		wg.Add(1)                                          //Add a new goroutine
		go s.checkAndProcessMessage(*message2, reply, &wg) //Check and process the message
	}
	wg.Wait()
}

func (s *Server) checkAndProcessMessage(message Message, reply *Response, wg *sync.WaitGroup) {
	defer wg.Done()

	if s.isMessageDeliverable(message) { //If the message is deliverable

		s.processDeliverableMessage(message, reply) //Process the message
	} else {
		fmt.Printf("The message is not deliverable: %d\n , My clock: %d\n", message.VectorTimestamp, s.MyClock)
	}
}

func (s *Server) isMessageDeliverable(message Message) bool {
	var mod bool //Variable that checks if the message is deliverable

	if MyId == message.ServerId { //If the serverId is equal to MyId
		fmt.Println(message.VectorTimestamp, s.MyClock)
		mod = message.VectorTimestamp[message.ServerId-1] == s.MyClock[message.ServerId-1] //I increased my counter before
	} else {
		mod = message.VectorTimestamp[message.ServerId-1] == s.MyClock[message.ServerId-1]+1
	}

	if mod { //If the message is deliverable
		for index, ts := range message.VectorTimestamp {
			if index == message.ServerId-1 { //If the index is equal to the serverId
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
	s.updateTimestamp(message) //Update the timestamp
	reply.Done = true          //Set the response to true
	fmt.Println("The message is deliverable")

	s.removeFromQueue(message) //Remove the message from the queue
	s.printDataStore()         //Print the data store

	fmt.Println("My actual timestamp:", s.MyClock)
}

func (s *Server) updateTimestamp(message Message) { //Function that updates the timestamp
	for ind, ts := range message.VectorTimestamp {
		if ts > s.MyClock[ind] { //If the timestamp is greater than the server's timestamp
			s.MyClock[ind] = ts
		}
	}
}
