package main

import (
	"fmt"
	"log"
	"main/clientCommon"
	"net/rpc"
	"sync"
)

func handlePutWithChannel(consistency int, key string, value string, config clientCommon.Config, serverId int, client *rpc.Client, processId int, actionId int, done chan<- bool, errCh chan<- error) {
	err := clientCommon.HandlePutAction(consistency, key, value, config, serverId, client, processId, actionId)
	if err != nil {
		errCh <- fmt.Errorf("error in HandlePutAction %d: %w", actionId, err)
	}
	done <- true

}

func handleGetWithChannel(consistency int, key string, client *rpc.Client, processId int, actionId int, done chan<- bool, errCh chan<- error) {
	err := clientCommon.HandleGetAction(consistency, key, client, processId, actionId)
	if err != nil {
		errCh <- fmt.Errorf("error in HandlePutAction %d: %w", actionId, err)
	}
	done <- true

}

func funcServer1MediumSequential(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	actionDone := make(chan bool, 5)
	defer close(actionDone)
	go handleGetWithChannel(consistency, "y", client, 1, 1, actionDone, errCh)
	go handleGetWithChannel(consistency, "y", client, 1, 2, actionDone, errCh)
	go handlePutWithChannel(consistency, "x", "b", config, 0, client, 1, 3, actionDone, errCh)
	go handleGetWithChannel(consistency, "x", client, 1, 4, actionDone, errCh)
	go handlePutWithChannel(consistency, "LastValue", "LastValue", config, 0, client, 1, 5, actionDone, errCh)

	for i := 0; i < 5; i++ {
		<-actionDone
	}

}

func funcServer2MediumSequential(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()
	fmt.Printf("Write W(x)b su processo 1\n")

	actionDone := make(chan bool, 4)
	defer close(actionDone)

	go handlePutWithChannel(consistency, "y", "b", config, 1, client, 2, 1, actionDone, errCh)
	go handleGetWithChannel(consistency, "x", client, 2, 2, actionDone, errCh)
	go handleGetWithChannel(consistency, "y", client, 2, 3, actionDone, errCh)
	go handleGetWithChannel(consistency, "x", client, 2, 4, actionDone, errCh)
	go handlePutWithChannel(consistency, "LastValue", "sa", config, 1, client, 2, 5, actionDone, errCh)

	for i := 0; i < 4; i++ {
		<-actionDone
	}
}

func funcServer3MediumSequential(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()
	fmt.Printf("R(x) su processo 2 = b //R(x) su processo 3 = b\n")

	actionDone := make(chan bool, 3)
	defer close(actionDone)

	go handleGetWithChannel(consistency, "x", client, 3, 1, actionDone, errCh)
	go handlePutWithChannel(consistency, "x", "c", config, 2, client, 3, 2, actionDone, errCh)
	go handlePutWithChannel(consistency, "y", "c", config, 2, client, 3, 3, actionDone, errCh)
	go handleGetWithChannel(consistency, "x", client, 3, 4, actionDone, errCh)
	go handlePutWithChannel(consistency, "LastValue", "df", config, 2, client, 3, 5, actionDone, errCh)

	for i := 0; i < 5; i++ {
		<-actionDone
	}
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

func TestSequentialMedium(config clientCommon.Config, client1 *rpc.Client, client2 *rpc.Client, client3 *rpc.Client) error {

	errCh := make(chan error, 3)
	defer close(errCh)

	var wg sync.WaitGroup
	wg.Add(3)

	go funcServer1MediumSequential(client1, config, 1, &wg, errCh)
	go funcServer2MediumSequential(client2, config, 1, &wg, errCh)
	go funcServer3MediumSequential(client3, config, 1, &wg, errCh)

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

func TestSequentialHard(config clientCommon.Config, client1 *rpc.Client, client2 *rpc.Client, client3 *rpc.Client) error {

	errCh := make(chan error, 3)
	defer close(errCh)

	var wg sync.WaitGroup
	wg.Add(3)

	//go funcServer1HardSequential(client1, config, 2, &wg, errCh)
	//go funcServer2HardSequential(client2, config, 2, &wg, errCh)
	//go funcServer3HardSequential(client3, config, 2, &wg, errCh)

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

func TestCausalMedium(config clientCommon.Config, client1 *rpc.Client, client2 *rpc.Client, client3 *rpc.Client) error {

	errCh := make(chan error, 3)
	defer close(errCh)

	var wg sync.WaitGroup
	wg.Add(3)

	//go funcServer1MediumCausal(client1, config, 1, &wg, errCh)
	//go funcServer2MediumCausal(client2, config, 1, &wg, errCh)
	//go funcServer3MediumCausal(client3, config, 1, &wg, errCh)

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

func TestCausalHard(config clientCommon.Config, client1 *rpc.Client, client2 *rpc.Client, client3 *rpc.Client) error {

	errCh := make(chan error, 3)
	defer close(errCh)

	var wg sync.WaitGroup
	wg.Add(3)

	//go funcServer1HardCausal(client1, config, 2, &wg, errCh)
	//go funcServer2HardCausal(client2, config, 2, &wg, errCh)
	//go funcServer3HardCausal(client3, config, 2, &wg, errCh)

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

func main() {
	var testType, testDifficulty string

	fmt.Println("Inserisci il tipo di test (sequenziale o causale):")
	_, err := fmt.Scanln(&testType)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserisci la difficoltà del test (medio o difficile):")
	_, err = fmt.Scanln(&testDifficulty)
	if err != nil {
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	err, config, client1, client2, client3 := Configuration()
	if err != nil {
		log.Fatal(err)
		return
	}
	switch testType {
	case "sequenziale":
		switch testDifficulty {
		case "medio":
			// Esegui il test sequenziale medio
			if err := TestSequentialMedium(config, client1, client2, client3); err != nil {
				fmt.Println("TestSequentialMedium error:", err)
			}
		case "difficile":
			// Esegui il test sequenziale difficile
			if err := TestSequentialHard(config, client1, client2, client3); err != nil {
				fmt.Println("TestSequentialHard error:", err)
			}
		default:
			fmt.Println("Difficoltà del test non riconosciuta.")
		}
	case "causale":
		switch testDifficulty {
		case "medio":
			// Esegui il test causale medio
			if err := TestCausalMedium(config, client1, client2, client3); err != nil {
				fmt.Println("TestCausalMedium error:", err)
			}
		case "difficile":
			// Esegui il test causale difficile
			if err := TestCausalHard(config, client1, client2, client3); err != nil {
				fmt.Println("TestCausalHard error:", err)
			}
		default:
			fmt.Println("Difficoltà del test non riconosciuta.")
		}
	default:
		fmt.Println("Tipo di test non riconosciuto.")
	}
}
