package serverOperation

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

func (s *Server) PutElement(message Message, reply *bool) error {
	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	s.myClock[MyId]++
	fmt.Printf("Ho incrementato il mio id %d\n", s.myClock[MyId])
	message.VectorTimestamp = s.myClock
	message.ServerId = MyId
	s.sendToServers(message)
	*reply = true
	return nil
}

func (s *Server) sendToServers(message Message) {

	for _, address := range addresses.Addresses {

		go func(addr string, msg Message) { //A thread for every server that I want to contact

			resultChan := make(chan error)

			for {
				client, err := rpc.Dial("tcp", addr)
				if err != nil {
					log.Fatal("Error in dialing: ", err)
				}
				var result bool //Channel that I use for control the ACK state

				err = client.Call("Server.SaveCausalElement", msg, &result) //Calling the RPC request for all the servers
				fmt.Printf("My result: %t\n\n", result)
				if err != nil {
					resultChan <- fmt.Errorf("error during the connection with %v: ", err)
				}
				if result == false { //If it hasn't succeeded, try again.
					continue
				} else { //If the procedure has been successful, exit the for loop.
					return
				}
			}
		}(address.Addr, message)
	}
}

func (s *Server) SaveCausalElement(message Message, reply *bool) error {

	s.myMutex.Lock()
	defer s.myMutex.Unlock()

	*reply = false

	if message.ServerId != MyId { ///I update my clock only if I'm not the sender
		if message.VectorTimestamp[MyId] > s.myClock[MyId] {
			s.myClock[MyId] = message.VectorTimestamp[MyId]
		}
	}
	s.myClock[MyId]++
	s.localQueue = append(s.localQueue, &message)

	//Now I send the ACK in multicast

	done := make(chan error)
	messageAck := AckMessage{Element: message, MyServerId: MyId}
	for _, address := range addresses.Addresses {

		go func(addr string) {

			for {
				client, err := rpc.Dial("tcp", addr)
				if err != nil {
					log.Fatal("Error dialing: ", err)
				}
				var ackResult *bool
				fmt.Printf("This is my id: %d\n", MyId)
				err = client.Call("Server.SendCausalAck", messageAck, &ackResult)
				if err != nil {
					done <- fmt.Errorf("error sending ack at server %d: %v", message.ServerId, err)
				}
				if *ackResult == false { //If the other server doesn't accept the ACK
					time.Sleep(2) //Wait 2 second and retry
					continue      //Retry
				}
				if *ackResult == true {
					break
				}
			}
		}(address.Addr)
	}
	*reply = true
	return nil
}
