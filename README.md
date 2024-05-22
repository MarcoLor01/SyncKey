# <span style="font-size: 2em;">**SyncKey**💾👨‍💻</span>

Implementazione di un Datastore Distribuito, mantenuto consistente tramite algoritmi di Multicast Totalmente e Causalmente Ordinato

**Come usarlo**

Utilizzo in locale

**Windows**

Impostare la variabile CONFIG=1 all'interno del file .env, successivamente lanciare il file batch presente nella directory principale con il comando: ```start.bat```

**Linux**

E' necessario prima modificare i permessi del file per renderlo eseguibile con il comando: ```chmod +x script.sh``` e successivamente lanciare il file bash con il comando: ```sh script.sh```

**Utilizzo da Docker**

Impostare la variabile CONFIG=2 nel file .env, successivamente seguire le istruzioni in base al proprio SO per lanciare i container.

Per entrare all'interno del client ed eseguire le operazioni di ```put```,```delete``` o ```get```, entrare all'interno del container del Client tramite il comando: ```docker exec -it synckey-client-1 sh -c 'cd clients && ./client' ```.

Per visualizzare i vari Logs dei server e seguirli in tempo reale (attualmente 3 repliche) è possibile eseguire 
```docker logs -f synckey-server1-1```, dove -1 può essere sostituito con l'indice del server di cui vogliamo visualizzarne i logs.

**Utilizzo da AWS***
