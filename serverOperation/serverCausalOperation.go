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
	s.MyClock[MyId-1] += 1
	message.VectorTimestamp = s.MyClock
	message.ServerId = MyId

	response := &Response{
		Done: false,
	}

	s.sendToOtherServersCausal(message, response)
	reply.Done = true
	return nil
}

func (s *Server) sendToOtherServersCausal(message Message, response *Response) {

	var wg sync.WaitGroup
	wg.Add(len(addresses.Addresses))
	int1 := 0
	for _, address := range addresses.Addresses {

		addr := address

		go func() {
			err := func(addr string, msg Message) error { //
				wg.Done()
				log.Printf("Go routine number %d\n", int1)
				int1++
				for {
					client, err := rpc.Dial("tcp", addr)
					if err != nil {

						return err
					}

					reply := Response{Done: false}

					if err1 := client.Call("Server.SaveElementCausal", msg, &reply); err1 != nil {
						return err1
					}

					if reply.Done == false {
						response.Done = false
					}

					response.Done = true
					return nil
				}
			}(addr.Addr, message)

			if err != nil {
				log.Fatal("Fatal error in sendToOtherSeversCausal operation ", err)
			}
		}()
	}
}

func (s *Server) SaveElementCausal(message Message, reply *Response) error {

	s.myMutex.Lock()
	defer s.myMutex.Unlock()
	fmt.Println(message.VectorTimestamp)
	var wg sync.WaitGroup
	s.addToQueue(message)

	for _, message2 := range s.localQueue {
		wg.Add(len(s.localQueue))
		go func(message2 Message) {
			defer wg.Done()
			var mod bool
			for {
				if MyId == message2.ServerId {
					mod = message2.VectorTimestamp[message2.ServerId-1] == s.MyClock[message2.ServerId-1]
				} else {
					mod = message2.VectorTimestamp[message2.ServerId-1] == s.MyClock[message2.ServerId-1]+1
				}

				if mod {
					for index, ts := range message2.VectorTimestamp {
						if index == message2.ServerId-1 {
							continue // Già controllato
						}
						if ts > s.MyClock[index] {
							reply.Done = false // Non è deliverable
							return
						}
					}
					reply.Done = true
				}

				fmt.Println("The message is deliverable")
				s.removeFromQueue(message2)
				s.printDataStore(s.dataStore)

				for ind, ts := range message2.VectorTimestamp {
					if ts > s.MyClock[ind] {
						s.MyClock[ind] = ts
					}
				}

				fmt.Println("My new timestamp:", s.MyClock)
				return
			}

		}(*message2)

	}
	fmt.Printf("Rilascio lock")
	wg.Wait()
	return nil
}
