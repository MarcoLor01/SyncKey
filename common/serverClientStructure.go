package common

//Struttura che contiene gli elementi in comune tra il messaggio per la consistenza causale
//e consistenza sequenziale

type Message struct {
	Key             string
	Value           string
	ServerId        int //Necessito di sapere chi ha inviato il messaggio, questo perché se ho un server che invia un messaggio a un altro server, il server che invia il messaggio non deve aggiornare il suo scalarClock
	OperationType   int //putOperation == 1, deleteOperation == 2}
	IdMessageClient int
	IdMessage       int
}

//Strutture di cui necessito per la consistenza sequenziale

type MessageSequential struct {
	MessageBase     Message
	ScalarTimestamp int
	NumberAck       int    //Solo se number == NumberOfServers il messaggio diventa consegnabile
	IdUnique        string //Id univoco che identifica il messaggio, aggiunto perché il confronto con key e value provocava problemi in caso di messaggi con key equivalente
}

type MessageSequentialOperation interface {
	GetKey() string
	GetMessageBase() Message
	GetTimestamp() int
	SetTimestamp(int)
	GetNumberAck() int
	IncrementNumberAck()
	GetID() string
	SetServerID(int)
	GetServerID() int
	SetID(string)
	GetValue() string
	GetOperationType(int)
}

func (m *MessageSequential) GetOperationType() int {
	return m.MessageBase.OperationType
}

func (m *MessageSequential) GetValue() string {
	return m.MessageBase.Value
}

func (m *MessageSequential) GetServerID() int {
	return m.MessageBase.ServerId
}

func (m *MessageSequential) SetServerID(id int) {
	m.MessageBase.ServerId = id
}

func (m *MessageSequential) GetKey() string {
	return m.MessageBase.Key
}

func (m *MessageSequential) SetTimestamp(value int) {
	m.ScalarTimestamp = value
}

func (m *MessageSequential) GetMessageBase() *Message {
	return &m.MessageBase
}

func (m *MessageSequential) GetTimestamp() int {
	return m.ScalarTimestamp
}

func (m *MessageSequential) GetNumberAck() int {
	return m.NumberAck
}

func (m *MessageSequential) IncrementNumberAck() {
	m.NumberAck++
}

func (m *MessageSequential) GetID() string {
	return m.IdUnique
}

func (m *MessageSequential) SetID(value string) {
	m.IdUnique = value
}

//Strutture di cui necessito per la consistenza causale

type MessageCausal struct {
	MessageBase     Message
	VectorTimestamp []int
	ServerId        int
	IdUnique        string
}

type Response struct {
	Done     bool
	GetValue string
} //Risposta per la consistenza causale

type ClientMessage struct {
	ClientId            int //Id del server che mi invierà i messaggi per tutta la durata dell'invio
	ActualNumberMessage int //Il messaggio che mi aspetto di ricevere
	ActualAnswerMessage int //Messaggio di cui posso inoltrare la richiesta per mantenere ordinamento FIFO
}
