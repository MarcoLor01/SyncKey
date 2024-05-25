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

//Strutture di cui necessito per la consistenza causale

type MessageCausal struct {
	MessageBase     Message
	VectorTimestamp []int
	ServerId        int
	IdUnique        string
}

type Response struct {
	Done bool
} //Risposta per la consistenza causale

type ClientMessage struct {
	ClientId            int //Id del server che mi invierà i messaggi per tutta la durata dell'invio
	ActualNumberMessage int //Il messaggio che mi aspetto di ricevere
}
