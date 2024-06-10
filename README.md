# <span style="font-size: 2em;">**SyncKey**💾👨‍💻</span>

Implementazione di un Datastore Distribuito, mantenuto consistente tramite algoritmi di Multicast Totalmente e Causalmente Ordinato

**Come usarlo**

Utilizzo in locale

**Windows**

Impostare la variabile CONFIG=1 all'interno del file .env, successivamente lanciare il file batch presente nella directory principale con il comando: ```start.bat 1```

**Linux**

E' necessario prima modificare i permessi del file per renderlo eseguibile con il comando: ```chmod +x start.sh``` e successivamente lanciare il file bash con il comando: ```sh start.sh 1```

**Utilizzo da Docker**

Impostare la variabile CONFIG=2 nel file .env, successivamente seguire le istruzioni in base al proprio SO per lanciare i container, sostituendo 1 con 2 nel lancio del file bash o batch

Per entrare all'interno del client ed eseguire le operazioni di ```put```,```delete``` o ```get```, entrare all'interno del container del Client tramite il comando: ```docker exec -it synckey-client-1 ./clientTests ```.

Per visualizzare i vari Logs dei server e seguirli in tempo reale (attualmente 3 repliche) è possibile eseguire 
```docker logs -f synckey-server1-1```, dove -1 può essere sostituito con l'indice del server di cui vogliamo visualizzarne i logs.

**Utilizzo da AWS**

Lanciare una nuova istanza tramite AWS selezionando un SO (nel caso descritto verrà usato Amazon Linux), accedere ad essa tramite SSH eseguendo: 
```ssh -i /percorso/al/tuo/file.pem ec2-user@<Public IP o Public DNS>``` sostituendo ```/percorso/al/tuo/file.pem``` con il percorso locale alla key .pem e ```Public IP o Public DNS``` con l'indirizzo IP pubblico o il DNS pubblico dell'istanza EC2. 

Installare git tramite: ```sudo yum install git```

Clonare repository tramite: ```git clone https://github.com/MarcoLor01/SyncKey```

Installare Docker-Compose tramite: 
```sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose```
Dopo aver scaricato l'eseguibile, aggiungi i permessi di esecuzione con il seguente comando:
```sudo chmod +x /usr/local/bin/docker-compose```

Runnare demon docker con il comando: ```sudo service docker start```


