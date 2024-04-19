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

	fmt.Println(response.Done)
	reply.Done = response.Done

	return nil
}

func (s *Server) sendToOtherServersCausal(message Message, response *Response) {

	ch := make(chan bool, len(addresses.Addresses))
	var wg sync.WaitGroup
	wg.Add(len(addresses.Addresses))

	for _, address := range addresses.Addresses {
		go func(address ServerAddress, ch chan bool) {
			defer wg.Done()
			client, err := rpc.Dial("tcp", address.Addr)
			if err != nil {
				log.Fatal("Errore durante la connessione RPC:", err)
				return
			}
			defer func(client *rpc.Client) {
				err := client.Close()
				if err != nil {

				}
			}(client)

			reply := &Response{Done: false}
			if err := client.Call("Server.SaveElementCausal", message, reply); err != nil {
				log.Fatal("Errore durante la chiamata RPC:", err)
				return
			}
			select {
			case ch <- reply.Done:
			default:
				fmt.Println("Il canale è stato chiuso prima che potesse essere inviato")
			}
		}(address, ch)
	}

	wg.Wait()
	close(ch)

	for i := range ch {
		if i {
			response.Done = true
			return
		}
	}
	response.Done = false
}

func (s *Server) SaveElementCausal(message Message, reply *Response) error {

	if message.ServerId != MyId {
		s.myMutex.Lock()
		defer s.myMutex.Unlock()
	}

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
							fmt.Println("The message is not deliverable")
							return
						}
					}

					reply.Done = true
					fmt.Println("The message is deliverable")
					s.removeFromQueue(message2)
					s.printDataStore(s.dataStore)

					for ind, ts := range message2.VectorTimestamp {
						if ts > s.MyClock[ind] {
							s.MyClock[ind] = ts
						}
					}
				} else {
					fmt.Println("The message is not deliverable")
				}

				fmt.Println("My actual timestamp:", s.MyClock)
				return
			}

		}(*message2)
	}
	wg.Wait()
	return nil
}
