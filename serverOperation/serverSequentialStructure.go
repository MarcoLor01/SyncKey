package serverOperation

import (
	"main/common"
	"sync"
)

type AckMessage struct {
	Element    common.MessageSequential
	MyServerId int
}

//Interfaccia utilizzata per i metodi del messaggio di ACK

type AckID interface {
	GetIdUnique() string
}

func (a *AckMessage) GetIdUnique() string {
	return a.Element.IdUnique
}

type ServerSequential struct {
	LocalQueue    []*common.MessageSequential //Coda locale
	myQueueMutex  sync.Mutex                  //Mutex per l'accesso alla coda
	MyScalarClock int                         //Clock scalare
	myClockMutex  sync.Mutex                  //Mutex per l'accesso al clock scalare
	BaseServer    ServerBase                  //Cose in comune tra server causale e sequenziale
}

type ClockOperation interface {
	getClock() int
	setClock(int)
	incrementClock()
}

type MutexOperation interface {
	lockClockMutex()
	unlockClockMutex()
	lockQueueMutex()
	unlockQueueMutex()
}

func (s *ServerSequential) getClock() int {
	return s.MyScalarClock
}

func (s *ServerSequential) setClock(value int) {
	s.MyScalarClock = value
}

func (s *ServerSequential) incrementClock() {
	s.MyScalarClock++
}

func (s *ServerSequential) lockClockMutex() {
	s.myClockMutex.Lock()
}

func (s *ServerSequential) unlockClockMutex() {
	s.myClockMutex.Unlock()
}

func (s *ServerSequential) lockQueueMutex() {
	s.myQueueMutex.Lock()
}

func (s *ServerSequential) unlockQueueMutex() {
	s.myQueueMutex.Unlock()
}
