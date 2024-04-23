package serverOperation

//func TestSendToOtherServersCausalConcurrency(t *testing.T) {
//	// Inizializza un oggetto Server
//	s := &Server{} // Assicurati di avere un'istanza valida di Server
//
//	// Numero di goroutine da testare
//	numGoroutines := 10
//
//	// Canale per sincronizzare il completamento delle goroutine
//	wg := sync.WaitGroup{}
//	wg.Add(numGoroutines)
//
//	// Esegui diverse goroutine per testare la funzione sendToOtherServersCausal
//	for i := 0; i < numGoroutines; i++ {
//		go func() {
//			defer wg.Done()
//
//			// Chiama la funzione sendToOtherServersCausal
//			// Passa un messaggio casuale e un riferimento a un oggetto Response
//			// Nota: Assicurati che addresses.Addresses contenga indirizzi validi
//			response := Response{}
//			s.sendToOtherServersCausal(Message{}, &response)
//		}()
//	}
//	// Attendi il completamento di tutte le goroutine
//	wg.Wait()
//}
