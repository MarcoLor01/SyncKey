package main

import (
	"fmt"
	"log"
	"main/clientCommon"
	"net/rpc"
	"sync"
	"time"
)

var iteration = 1
var mu sync.Mutex

func handlePutWithChannel(consistency int, key string, value string, config clientCommon.Config, client *rpc.Client, processId int, actionId int, done chan<- bool, errCh chan<- error) {
	err := clientCommon.HandlePutAction(consistency, key, value, config, client, processId, actionId)
	if err != nil {
		errCh <- fmt.Errorf("error in HandlePutAction %d: %w", actionId, err)
	}
	done <- true
}

func handleGetWithChannel(consistency int, key string, config clientCommon.Config, client *rpc.Client, processId int, actionId int, done chan<- bool, errCh chan<- error) {
	err := clientCommon.HandleGetAction(consistency, key, config, client, processId, actionId)
	if err != nil {
		errCh <- fmt.Errorf("error in HandleGetAction %d: %w", actionId, err)
	}
	done <- true
}

func handleDeleteWithChannel(consistency int, key string, config clientCommon.Config, client *rpc.Client, processId int, actionId int, done chan<- bool, errCh chan<- error) {
	err := clientCommon.HandleDeleteAction(consistency, key, config, client, processId, actionId)
	if err != nil {
		errCh <- fmt.Errorf("error in HandleDeleteAction %d: %w", actionId, err)
	}
	done <- true
}

func executeActions(actions []func(), wg *sync.WaitGroup) {
	defer wg.Done()

	actionDone := make(chan bool, len(actions))
	defer close(actionDone)

	for _, action := range actions {
		go func(action func()) {
			action()
			actionDone <- true
		}(action)

		mu.Lock()
		timeSleep := time.Duration(200 * iteration)
		iteration++
		mu.Unlock()
		time.Sleep(timeSleep * time.Millisecond)
	}

	for i := 0; i < len(actions); i++ {
		<-actionDone
	}
}

func funcServer1MediumCausal(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 3)
	actions := []func(){
		func() { handlePutWithChannel(consistency, "x", "a", config, client, 1, 1, actionDone, errCh) },
		func() { handlePutWithChannel(consistency, "x", "b", config, client, 1, 2, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 1, 3, actionDone, errCh) },
	}
	executeActions(actions, wg)
}

func funcServer2MediumCausal(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 3)
	actions := []func(){
		func() { handleGetWithChannel(consistency, "x", config, client, 2, 1, actionDone, errCh) },
		func() { handlePutWithChannel(consistency, "x", "c", config, client, 2, 2, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 2, 3, actionDone, errCh) },
	}
	executeActions(actions, wg)

}

func funcServer3MediumCausal(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 3)
	actions := []func(){
		func() { handleGetWithChannel(consistency, "x", config, client, 3, 1, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 3, 2, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 3, 3, actionDone, errCh) },
	}
	executeActions(actions, wg)
}

//Test difficile

func funcServer1HardCausal(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 4)
	actions := []func(){
		func() { handlePutWithChannel(consistency, "x", "a", config, client, 1, 1, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "y", config, client, 1, 2, actionDone, errCh) },
		func() { handlePutWithChannel(consistency, "x", "b", config, client, 1, 3, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 1, 4, actionDone, errCh) },
	}
	executeActions(actions, wg)
}

func funcServer2HardCausal(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 4)
	actions := []func(){
		func() { handlePutWithChannel(consistency, "y", "b", config, client, 2, 1, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 2, 2, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "y", config, client, 2, 3, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 2, 4, actionDone, errCh) },
	}
	executeActions(actions, wg)

}

func funcServer3HardCausal(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 4)
	actions := []func(){
		func() { handleGetWithChannel(consistency, "x", config, client, 3, 1, actionDone, errCh) },
		func() { handlePutWithChannel(consistency, "x", "c", config, client, 3, 2, actionDone, errCh) },
		func() { handlePutWithChannel(consistency, "y", "c", config, client, 3, 3, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 3, 4, actionDone, errCh) },
	}
	executeActions(actions, wg)
}

func funcServer1MediumSequential(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 5)
	actions := []func(){
		func() { handlePutWithChannel(consistency, "x", "a", config, client, 1, 1, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "y", config, client, 1, 2, actionDone, errCh) },
		func() { handlePutWithChannel(consistency, "x", "b", config, client, 1, 3, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 1, 4, actionDone, errCh) },
		func() {
			handlePutWithChannel(consistency, "LastValue", "LastValue", config, client, 1, 5, actionDone, errCh)
		},
	}
	executeActions(actions, wg)
}

func funcServer2MediumSequential(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 5)
	actions := []func(){
		func() { handlePutWithChannel(consistency, "y", "b", config, client, 2, 1, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 2, 2, actionDone, errCh) },
		func() { handlePutWithChannel(consistency, "y", "a", config, client, 2, 3, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 2, 4, actionDone, errCh) },
		func() {
			handlePutWithChannel(consistency, "LastValue", "LastValue", config, client, 2, 5, actionDone, errCh)
		},
	}

	executeActions(actions, wg)
}

func funcServer3MediumSequential(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 5)
	actions := []func(){
		func() { handlePutWithChannel(consistency, "x", "c", config, client, 3, 1, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 3, 2, actionDone, errCh) },
		func() { handlePutWithChannel(consistency, "y", "c", config, client, 3, 3, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 3, 4, actionDone, errCh) },
		func() {
			handlePutWithChannel(consistency, "LastValue", "LastValue", config, client, 3, 5, actionDone, errCh)
		},
	}

	executeActions(actions, wg)
}

func Configuration() (error, clientCommon.Config, *rpc.Client, *rpc.Client, *rpc.Client) {
	configuration := clientCommon.LoadEnvironment()

	config, err := clientCommon.LoadConfig(configuration)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err), config, nil, nil, nil
	}

	client1, err := clientCommon.CreateClient(config, 0) // Sto contattando il server 0
	if err != nil {
		return fmt.Errorf("error creating client 1: %w", err), config, nil, nil, nil
	}

	client2, err := clientCommon.CreateClient(config, 1) // Sto contattando il server 1
	if err != nil {
		return fmt.Errorf("error creating client 2: %w", err), config, nil, nil, nil
	}

	client3, err := clientCommon.CreateClient(config, 2) // Sto contattando il server 2
	if err != nil {
		return fmt.Errorf("error creating client 3: %w", err), config, nil, nil, nil
	}

	return nil, config, client1, client2, client3
}

func TestMedium(config clientCommon.Config, client1 *rpc.Client, client2 *rpc.Client, client3 *rpc.Client, consistency int) error {
	errCh := make(chan error, 3)
	defer close(errCh)

	var wg sync.WaitGroup
	wg.Add(3)
	if consistency == 1 {

		go funcServer1MediumSequential(client1, config, consistency, &wg, errCh)
		go funcServer2MediumSequential(client2, config, consistency, &wg, errCh)
		go funcServer3MediumSequential(client3, config, consistency, &wg, errCh)

	} else if consistency == 0 {

		go funcServer1MediumCausal(client1, config, consistency, &wg, errCh)
		go funcServer2MediumCausal(client2, config, consistency, &wg, errCh)
		go funcServer3MediumCausal(client3, config, consistency, &wg, errCh)

	} else {
		log.Fatal("wrong value for consistency")
	}

	wg.Wait()

	go func() {
		for err := range errCh {
			if err != nil {
				fmt.Println(err)
			}
		}
	}()
	return nil
}

func TestHard(config clientCommon.Config, client1 *rpc.Client, client2 *rpc.Client, client3 *rpc.Client, consistency int) error {
	errCh := make(chan error, 3)
	defer close(errCh)

	var wg sync.WaitGroup
	wg.Add(3)

	if consistency == 1 {

		go funcServer1HardSequential(client1, config, consistency, &wg, errCh)
		go funcServer2HardSequential(client2, config, consistency, &wg, errCh)
		go funcServer3HardSequential(client3, config, consistency, &wg, errCh)

	} else if consistency == 0 {

		go funcServer1HardCausal(client1, config, consistency, &wg, errCh)
		go funcServer2HardCausal(client2, config, consistency, &wg, errCh)
		go funcServer3HardCausal(client3, config, consistency, &wg, errCh)

	} else {
		log.Fatal("wrong value for consistency")
	}
	wg.Wait()

	go func() {
		for err := range errCh {
			if err != nil {
				fmt.Println(err)
			}
		}
	}()
	return nil
}

func funcServer1HardSequential(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 9)

	actions := []func(){
		func() { handlePutWithChannel(consistency, "y", "a", config, client, 1, 1, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 1, 2, actionDone, errCh) },
		func() { handleDeleteWithChannel(consistency, "x", config, client, 1, 3, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 1, 4, actionDone, errCh) },
		func() { handlePutWithChannel(consistency, "y", "b", config, client, 1, 5, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "y", config, client, 1, 6, actionDone, errCh) },
		func() { handleDeleteWithChannel(consistency, "y", config, client, 1, 7, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "y", config, client, 1, 8, actionDone, errCh) },
		func() {
			handlePutWithChannel(consistency, "LastValue", "LastValue", config, client, 1, 9, actionDone, errCh)
		},
	}

	executeActions(actions, wg)
}

func funcServer2HardSequential(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 9)

	actions := []func(){
		func() { handlePutWithChannel(consistency, "x", "a", config, client, 2, 1, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "y", config, client, 2, 2, actionDone, errCh) },
		func() { handleDeleteWithChannel(consistency, "y", config, client, 2, 3, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "y", config, client, 2, 4, actionDone, errCh) },
		func() { handlePutWithChannel(consistency, "x", "a", config, client, 2, 5, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "y", config, client, 2, 6, actionDone, errCh) },
		func() { handleDeleteWithChannel(consistency, "x", config, client, 2, 7, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "y", config, client, 2, 8, actionDone, errCh) },
		func() {
			handlePutWithChannel(consistency, "LastValue", "LastValue", config, client, 2, 9, actionDone, errCh)
		},
	}

	executeActions(actions, wg)
}

func funcServer3HardSequential(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	actionDone := make(chan bool, 9)

	actions := []func(){
		func() { handlePutWithChannel(consistency, "y", "b", config, client, 3, 1, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 3, 2, actionDone, errCh) },
		func() { handleDeleteWithChannel(consistency, "x", config, client, 3, 3, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "y", config, client, 3, 4, actionDone, errCh) },
		func() { handlePutWithChannel(consistency, "y", "a", config, client, 3, 5, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "y", config, client, 3, 6, actionDone, errCh) },
		func() { handleDeleteWithChannel(consistency, "x", config, client, 3, 7, actionDone, errCh) },
		func() { handleGetWithChannel(consistency, "x", config, client, 3, 8, actionDone, errCh) },
		func() {
			handlePutWithChannel(consistency, "LastValue", "LastValue", config, client, 3, 9, actionDone, errCh)
		},
	}

	executeActions(actions, wg)
}

func main() {
	var testType, testDifficulty string

	// Funzione per richiedere il tipo di test con controllo validità
	for {
		fmt.Println("Inserisci il tipo di test (sequenziale o causale): ")
		_, err := fmt.Scan(&testType)
		if err != nil {
			log.Println("Errore di input:", err)
			continue
		}
		if testType == "sequenziale" || testType == "causale" {
			break
		} else {
			fmt.Println("Tipo di test non riconosciuto. Per favore inserisci 'sequenziale' o 'causale'.")
		}
	}

	// Funzione per richiedere la difficoltà del test con controllo validità
	for {
		fmt.Println("Inserisci la difficoltà del test (medio o difficile): ")
		_, err := fmt.Scan(&testDifficulty)
		if err != nil {
			log.Println("Errore di input:", err)
			continue
		}
		if testDifficulty == "medio" || testDifficulty == "difficile" {
			break
		} else {
			fmt.Println("Difficoltà del test non riconosciuta. Per favore inserisci 'medio' o 'difficile'.")
		}
	}

	// Chiamata alla configurazione
	err, config, client1, client2, client3 := Configuration()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Esecuzione dei test in base al tipo e alla difficoltà
	switch testType {
	case "sequenziale":
		switch testDifficulty {
		case "medio":
			// Esegui il test sequenziale medio
			if err := TestMedium(config, client1, client2, client3, 1); err != nil {
				fmt.Println("TestMedium error:", err)
			}
		case "difficile":
			// Esegui il test sequenziale difficile
			if err := TestHard(config, client1, client2, client3, 1); err != nil {
				fmt.Println("TestHard error:", err)
			}
		}
	case "causale":
		switch testDifficulty {
		case "medio":
			// Esegui il test causale medio
			if err := TestMedium(config, client1, client2, client3, 0); err != nil {
				fmt.Println("TestCausalMedium error:", err)
			}
		case "difficile":
			// Esegui il test causale difficile
			if err := TestHard(config, client1, client2, client3, 0); err != nil {
				fmt.Println("TestCausalHard error:", err)
			}
		}
	}
}