package common

//Strutture di cui necessito per la consistenza sequenziale

type MessageSequential struct {
	Key                  string
	Value                string
	ScalarTimestamp      int
	ServerId             int    //Necessito di sapere chi ha inviato il messaggio, questo perché se ho un server che invia un messaggio a un altro server, il server che invia il messaggio non deve aggiornare il suo scalarClock
	NumberAck            int    //Solo se number == NumberOfServers il messaggio diventa consegnabile
	OperationType        int    //putOperation == 1, deleteOperation == 2
	InsertQueueTimestamp int64  //Quando è stato aggiunto in coda il messaggio
	IdUnique             string //Id univoco che identifica il messaggio, aggiunto perché il confronto con key e value provocava problemi in caso di messaggi con key equivalente
}

type ResponseSequential struct {
	Done bool
} //Risposta per la consistenza sequenziale

//Strutture di cui necessito per la consistenza causale

type MessageCausal struct {
	Key             string
	Value           string
	VectorTimestamp []int
	ServerId        int
	numberAck       int
	OperationType   int
	IdUnique        string
}

type ResponseCausal struct {
	Done bool
} //Risposta per la consistenza causale