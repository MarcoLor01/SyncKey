package serverOperation

import (
	"main/common"
	"sync"
)

type ServerCausal struct {
	LocalQueue   []*common.MessageCausal //Coda locale
	myQueueMutex sync.Mutex              //Mutex per la sincronizzazione dell'accesso in coda
	MyClock      []int                   //Il mio vettore di clock vettoriale
	myClockMutex sync.Mutex              //Mutex per la sincronizzazione dell'accesso al clock
	BaseServer   ServerBase              //Cose in comune tra server causale e sequenziale
}

//Inizializzazione di un server con consistenza causale

func CreateNewCausalDataStore() *ServerCausal {

	return &ServerCausal{
		LocalQueue: make([]*common.MessageCausal, 0),
		MyClock:    make([]int, len(addresses.Addresses)), //Inizializzo il mio vettore di clock vettoriale
		BaseServer: ServerBase{
			DataStore: make(map[string]string),
		},
	}
}

type ClockVectorialOperation interface {
	getClock() int
	setClock(int)
	incrementClock()
}

func (s *ServerCausal) getClock() []int {
	return s.MyClock
}

func (s *ServerCausal) setClock(value []int) {
	s.MyClock = value
}

func (s *ServerCausal) incrementClock() {
	s.MyClock[MyId-1]++
}

func (s *ServerCausal) lockClockMutex() {
	s.myClockMutex.Lock()
}

func (s *ServerCausal) unlockClockMutex() {
	s.myClockMutex.Unlock()
}

func (s *ServerCausal) lockQueueMutex() {
	s.myQueueMutex.Lock()
}

func (s *ServerCausal) unlockQueueMutex() {
	s.myQueueMutex.Unlock()
}
