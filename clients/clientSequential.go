package main

import (
	"fmt"
	"main/clientCommon"
	"net/rpc"
	"sync"
)

func handleActionWithChannel(consistency int, key string, value string, config clientCommon.Config, serverId int, client *rpc.Client, processId int, actionId int, done chan<- bool, errCh chan<- error) {
	err := clientCommon.HandlePutAction(consistency, key, value, config, serverId, client, processId, actionId)
	if err != nil {
		errCh <- fmt.Errorf("error in HandlePutAction %d: %w", actionId, err)
	}
	done <- true
}

func funcServer1(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()
	fmt.Printf("Write W(x)a su processo 0\n")

	actionDone := make(chan bool, 4)
	defer close(actionDone)
	go handleActionWithChannel(consistency, "1", "c", config, 0, client, 1, 1, actionDone, errCh)
	go handleActionWithChannel(consistency, "4", "df", config, 0, client, 1, 2, actionDone, errCh)
	go handleActionWithChannel(consistency, "7", "df", config, 0, client, 1, 3, actionDone, errCh)
	go handleActionWithChannel(consistency, "LastValue3", "df", config, 0, client, 1, 4, actionDone, errCh)

	for i := 0; i < 4; i++ {
		<-actionDone
	}
}

func funcServer2(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()
	fmt.Printf("Write W(x)b su processo 1\n")

	actionDone := make(chan bool, 4)
	defer close(actionDone)

	go handleActionWithChannel(consistency, "2", "b", config, 1, client, 2, 1, actionDone, errCh)
	go handleActionWithChannel(consistency, "5", "sa", config, 1, client, 2, 2, actionDone, errCh)
	go handleActionWithChannel(consistency, "8", "df", config, 1, client, 2, 3, actionDone, errCh)
	go handleActionWithChannel(consistency, "LastValue2", "sa", config, 1, client, 2, 4, actionDone, errCh)

	for i := 0; i < 4; i++ {
		<-actionDone
	}
}

func funcServer3(client *rpc.Client, config clientCommon.Config, consistency int, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()
	fmt.Printf("R(x) su processo 2 = b //R(x) su processo 3 = b\n")

	actionDone := make(chan bool, 4)
	defer close(actionDone)

	go handleActionWithChannel(consistency, "3", "aaa", config, 2, client, 3, 1, actionDone, errCh)
	go handleActionWithChannel(consistency, "6", "df", config, 2, client, 3, 2, actionDone, errCh)
	go handleActionWithChannel(consistency, "9", "df", config, 2, client, 3, 3, actionDone, errCh)
	go handleActionWithChannel(consistency, "LastValue1", "df", config, 2, client, 3, 4, actionDone, errCh)

	for i := 0; i < 4; i++ {
		<-actionDone
	}
}

func TestFirst() error {
	configuration := clientCommon.LoadEnvironment()

	config, err := clientCommon.LoadConfig(configuration)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	client1, err := clientCommon.CreateClient(config, 0) // Sto contattando il server 0
	if err != nil {
		return fmt.Errorf("error creating client 1: %w", err)
	}

	client2, err := clientCommon.CreateClient(config, 1) // Sto contattando il server 1
	if err != nil {
		return fmt.Errorf("error creating client 2: %w", err)
	}

	client3, err := clientCommon.CreateClient(config, 2) // Sto contattando il server 2
	if err != nil {
		return fmt.Errorf("error creating client 3: %w", err)
	}

	errCh := make(chan error)
	defer close(errCh)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		for err := range errCh {
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	fmt.Println("Eseguo goroutine")
	go funcServer1(client1, config, 1, &wg, errCh)
	go funcServer2(client2, config, 1, &wg, errCh)
	go funcServer3(client3, config, 1, &wg, errCh)

	wg.Wait()

	return nil
}

func main() {
	if err := TestFirst(); err != nil {
		fmt.Println("TestFirst error:", err)
	}
}
