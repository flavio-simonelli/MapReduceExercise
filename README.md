# Sistema Distribuito MapReduce per Ordinamento di Numeri  

## 1. Introduzione

Il presente repository implementa un sistema distribuito basato sul paradigma **MapReduce**, con l’obiettivo di elaborare dati in parallelo e distribuire il carico computazionale su più nodi worker.  
Il problema trattato consiste nell’**ordinamento crescente** di un insieme di numeri casuali, sfruttando l’architettura distribuita per ottimizzare le performance in scenari di elaborazione su larga scala.  
Questo sistema è stato realizzato nell’ambito di un esercizio del corso di Sistemi Distribuiti e Cloud Computing.  
La scelta del linguaggio di programmazione per la sua implementazione è **Go**.  
Per la comunicazione tra i vari componenti del sistema è stato adottato il protocollo **gRPC**.
L'implementazione del sistema è **focalizzata sull'ottimizzazione del costo computazionale**, con particolare attenzione alla gestione efficiente dei dati durante le fasi di mappatura e riduzione. In particolare, nella fase di **partizionamento dei dati** (Chunk → Partizioni), il sistema utilizza una **funzione di partizionamento ispirata al Terasort di Google**, che applica un **algoritmo di hashing** per suddividere il chunk di dati in diverse partizioni da distribuire ai vari reducer. Il processo di partizionamento è basato su una funzione di **hashing** che mappa i numeri del chunk rispetto al numero di **reducer attivi**, garantendo una **distribuzione omogenea** dei dati anche in presenza di **variazioni dinamiche del numero di reducer a runtime**. Questa tecnica consente di ottenere una **partizione bilanciata ed efficiente**, migliorando il carico di lavoro sui reducer.
Inoltre, nella fase di **ordinamento delle partizioni** dei reducer, è stato implementato un **algoritmo di ordinamento delle partizioni ottimizzato**. Questo algoritmo sfrutta il fatto che le partizioni sono già **ordinate localmente**, permettendo di **combinare i dati provenienti dai vari reducer in un singolo array ordinato** in modo particolarmente efficiente. Questo approccio riduce il **costo computazionale** al minimo possibile, con una **complessità temporale di O(N)**. In questo modo, l'elaborazione dei dati risulta particolarmente **scalabile e veloce**, garantendo alte prestazioni anche con un aumento del volume dei dati o del numero di reducer coinvolti nel processo.


## 2. Panoramica del sistema
### 2.1. Rappresentazione grafica

```
        Client
           |
           v
        Master
       /   |   \
      v    v    v
  Worker Worker Worker ...
```
### 2.2. Schema di funzionamento

Il sistema è composto da tre principali componenti:  

### 2.2.1. Client  
- **Descrizione**:  
  Un client minimale (**tiny client**) progettato per generare richieste **rapide** relative al problema dell'ordinamento.  
- **Funzionalità**:  
  - Genera un numero arbitrario (da 1 a 100 numeri) di interi casuali a runTime utilizzando come **seed** il timestamp attuale.  
  - Invia le richieste al nodo **Master**, specificando i numeri generati.  

### 2.2.2. Master  
- **Descrizione**:  
  Il server centrale responsabile della gestione delle richieste provenienti dai client e dell'avvio dell'operazione di computazione distribuita.  
- **Funzionalità**:  
  - Riceve le richieste dal client
  - Divide la lista di numeri in **chunk omogenei** per distribuirli ai nodi **mapper**.  

### 2.2.3. Worker  
- **Descrizione**:  
  Server che , come richiesto nella consegna dell'esercizio, eseguono sia la fase di **mapping** sia quella di **reducing**, rispettando il modello MapReduce. È possibile separare i due ruoli in nodi distinti con minimi interventi architetturali, aumentando ulteriormente la distribuzione del sistema.  
- **Funzionalità**:  
  - **Fase di Mapping**:  
    - Riceve dal master un chunk di dati;
    - Esegue un primo ordinamento locale del chunk;
    - Redistribuisce i numeri del proprio chunk dividendoli in partizioni omogenee per ogni Reducer (che coincidono con gli stessi worker) tramite una funzione di partizionamento dinamico basata su **Terasort**;  
  - **Fase di Reducing**:  
    - Riceve e memorizza all'interno di una matrice  le partizioni ordinate localmente da ogni Mapper
    - Attende la ricezione di tutte le partizioni di ogni mapper prima di iniziare la fase di reducing
    - Combina le partizioni utilizzando un algoritmo di **merge sort ottimizzato** producendo un array ordinato completo, che viene scritto su un file di output dedicato per ogni worker. Questo approccio simula un **file system distribuito**, con output distribuiti su più nodi.  

---

## 3. Dettagli dell'implementazione

### 3.1. Comunicazione fra i componenti
La comunicazione fra i componenti del sistema avviene tramite l'utilizzo del protocollo gRPC definito tramite il file di configurazione     `MapReduceExercise/proto/mapreduce.proto`.
- **Messaggi**:

    - `Chunk`: utilizzato nelle varie chiamate RPC per il passaggio di un array di numeri interi, contiene:
        - un array di numeri interi (numbers);
        - un identificativo di richiesta(idRichiesta), utile per tracciare la risposta nel caso di richieste parallele;
    - `response`: messaggio di risposta, contiene:
    - un valore booleano (ack) per il successo dell'operazione;
    - l'ID della richiesta;

- **Servizio `MasterService`:**
    - `NewRequest`:
        - **Chi chiama**: Il client
        - **Chi riceve**: Il master 
        - **Parametro in input**:
            - **`Chunk`**:
                - `numbers`: L'array di numeri interi da elaborare.
                - `idRichiesta`: L'identificativo univoco della richiesta. in questo caso data da `clientID_timepstamp` utile per tracciare le richieste nel caso di richieste parallele.
        - **Parametro in output**:
            - **`response`**:
                - `ack`: Un valore booleano che indica se la richiesta è stata ricevuta correttamente e senza errori (`true` per successo, `false` per errore).
                - `idRichiesta`: L'ID della richiesta inviata, utile per tracciare e associare la risposta alla richiesta originale nel caso di richieste parallele.

- **Servizio `WorkerService`:**
    - `Mapping`:
        - **Chi chiama**: Il Master
        - **Chi riceve**: Il Worker (mapper)
        - **Parametro in input**:
            - **`Chunk`**:
                - `numbers`: Il chunk di numeri interi da mappare.
                - `idRichiesta`: L'ID della richiesta originale, utile per tracciare l'operazione nel caso di richieste parallele.
        - **Parametro in output**:
            - **`response`**:
                - `ack`: Un valore booleano che indica se l'operazione di mappatura è stata completata correttamente (`true` per successo, `false` per errore).
                - `idRichiesta`: L'ID della richiesta originale, utile per tracciare la risposta nel caso di richieste parallele.

    - `Reducing`:
        - **Chi chiama**: Il worker (mapper)
        - **Chi riceve**: Il worker (reducer)
        - **Parametro in input**:
            - **`Chunk`**:
                - `numbers`: La partizione di numeri interi da ridurre (ad esempio, per eseguire un'aggregazione).
                - `idRichiesta`: L'ID della richiesta originale, utile per tracciare la risposta nel caso di richieste parallele.
        - **Parametro in output**:
            - **`response`**:
                - `ack`: Un valore booleano che indica se l'operazione di riduzione è stata completata con successo (`true` per successo, `false` per errore).
                - `idRichiesta`: L'ID della richiesta originale, utile per tracciare la risposta nel caso di richieste parallele.

### 3.2. Files di configurazione
Come richiesto dall'esercizio, la configurazione della struttura del sistema distribuito è statica e non prevede meccanismi di configurazione dinamica. In particolare, i dettagli relativi al worker e al master attivo sono definiti nei file di configurazione `configWorker.json` e `configMaster.json`. 
Tale separazione è stata adottata in quanto si considera un ambiente distribuito in cui il client è distribuito con il proprio file di configurazione, contenente **esclusivamente** le informazioni relative al master. Il client, pertanto, non è a conoscenza della configurazione dei worker.

- **File `configMaster.json`**:
    - **`master`**: Un oggetto che rappresenta il nodo master nel sistema distribuito. Contiene i seguenti campi:
        - **`ip`**: Indirizzo IP del master. In questo caso, è impostato su `"localhost"`, che indica che il master è in esecuzione sulla stessa macchina del client.
        - **`port`**: Porta di connessione del master. In questo esempio, la porta è la `50004`, attraverso cui il client può comunicare con il master.


- **File `configWorker.json`**:
    - **`workers`**: Una lista di oggetti, ciascuno dei quali rappresenta un worker nel sistema distribuito. Ogni oggetto contiene:
        - **`id`**: Identificativo univoco del worker (ad esempio, `1`, `2`, `3`, ecc.).
        - **`ip`**: Indirizzo IP del worker. In questo caso, tutti i worker sono configurati con l'indirizzo `"localhost"`, indicando che si trovano sulla stessa macchina del master.
        - **`port`**: Porta di connessione di ciascun worker. Ogni worker è associato a una porta specifica, che varia da `50000` a `50003`.
    - **`idWorkerNotUse`**: Campo che tiene traccia dell'ID che non può essere utilizzato per la definizione di un worker, in quanto designato per operazioni di controllo. Quando un worker viene avviato, è necessario specificare la porta su cui deve operare. Se la porta fornita non corrisponde a una delle porte dei worker definiti nel file di configurazione, il valore della variabile `ID` viene impostato su `-1`, indicando che la configurazione non è valida, poiché il worker non sarebbe identificato correttamente all'interno del file di configurazione.

I seguenti file di configurazione vengono letti e gestiti tramite il package Go `config/utilsConfiguration.go`:
**Costanti**:
- **`configFileWorker`**: Definisce il percorso del file di configurazione per i worker, situato in `config/configWorker.json`.
- **`configFileMaster`**: Definisce il percorso del file di configurazione per il master, situato in `config/configMaster.json`.

**Strutture**:
- **`Worker`**: Rappresenta un singolo worker nel sistema. La struttura contiene i seguenti campi:
    - **`IP`**: L'indirizzo IP del worker, rappresentato come stringa.
    - **`Port`**: La porta associata al worker, di tipo `int32`.
    - **`ID`**: Identificativo univoco del worker, di tipo `int32`.

- **`ConfigWorker`**: Rappresenta la configurazione dei worker nel sistema. Contiene un campo:
    - **`Workers`**: Una slice di oggetti `Worker`, ognuno dei quali rappresenta un worker configurato nel sistema.

- **`ConfigMaster`**: Rappresenta la configurazione del master. Contiene un oggetto:
    - **`Master`**: Un oggetto che include i seguenti campi:
        - **`IP`**: L'indirizzo IP del master, rappresentato come stringa.
        - **`Port`**: La porta associata al master, di tipo `int`.

**Funzioni**:
- **`ReadConfigWorker`**: Funzione di utilità che legge e esegue il parsing del file di configurazione dei worker. La funzione segue i seguenti passaggi:
    1. Legge il contenuto del file JSON `config/configWorker.json`.
    2. Esegue il parsing del file JSON nella struttura `ConfigWorker`.
    3. Restituisce un oggetto `ConfigWorker` contenente la configurazione letta, o un errore nel caso in cui si verifichino problemi durante la lettura o il parsing.

- **`ReadConfigMaster`**: Funzione di utilità che legge e esegue il parsing del file di configurazione del master. La funzione segue i seguenti passaggi:
    1. Legge il contenuto del file JSON `config/configMaster.json`.
    2. Esegue il parsing del file JSON nella struttura `ConfigMaster`.
    3. Restituisce un oggetto `ConfigMaster` contenente la configurazione letta, o un errore nel caso in cui si verifichino problemi durante la lettura o il parsing.

### 3.3 Client
Il client ha il compito di inviare una richiesta al nodo master, contenente una lista di numeri generati casualmente. Il flusso di esecuzione del client è il seguente:

1. **Generazione della Sequenza di Numeri**: La funzione `generatoreNumeri()` crea una slice di interi di dimensione variabile (compresa tra 1 e 100). I numeri sono generati casualmente, utilizzando il timestamp corrente come seme per il generatore di numeri casuali, garantendo così una distribuzione non prevedibile dei valori.

2. **Lettura della Configurazione del Master**: Il client legge la configurazione del master attraverso la funzione `ReadConfigMaster()`, la quale carica i parametri di configurazione (IP e porta) dal file `configMaster.json`. L'indirizzo del master viene quindi costruito dinamicamente combinando l'IP e la porta letti.

3. **Stabilimento della Connessione gRPC**: Il client stabilisce una connessione al nodo master utilizzando il framework gRPC. 

4. **Invio della Richiesta al Master**: Una volta stabilita la connessione, il client invia una richiesta al master mediante il metodo `NewRequest()` del servizio gRPC, passando come parametro un oggetto `Chunk` che contiene la lista di numeri generati e un identificativo univoco per la richiesta. L'identificativo è formato concatenando l'identificativo del client e un timestamp.

### 3.4. Master 
Il nodo master implementa il server gRPC per la gestione delle richieste di tipo MapReduce e per la distribuzione dei carichi di lavoro ai worker. Il flusso di esecuzione del nodo master è il seguente:

1. **Inizializzazione**: 
   - Il master mantiene una lista di worker configurati, letta dal file `configWorker.json`, e il numero totale di worker (`nWorkers`).
   - La struttura `Master` implementa il servizio gRPC `MasterServiceServer`, che definisce le operazioni disponibili per il master, tra cui la gestione delle richieste di mappatura.

2. **Gestione delle Richieste**: 
   - La funzione `NewRequest` riceve una richiesta di tipo `Chunk`, che contiene un array di numeri e un identificativo univoco (`IdRichiesta`).
   - La funzione divide la lista di numeri ricevuti in più chunk, uno per ciascun worker. La divisione dei numeri è bilanciata, ma eventuali numeri in eccesso vengono distribuiti tra i primi worker.

3. **Distribuzione dei Carichi ai Worker**:
   - Una volta suddivisa la lista in chunk, il master avvia una goroutine per ogni worker, utilizzando la libreria `sync` per sincronizzare l'esecuzione delle goroutine.
   - Per ogni worker, il master crea una connessione gRPC e invia il chunk di dati corrispondente tramite il metodo `Mapping` del servizio gRPC del worker.

4. **Gestione della Connessione e Sincronizzazione**:
   - Il master gestisce la connessione a ciascun worker in modalità concorrente, utilizzando goroutine per eseguire in parallelo le richieste di mappatura.
   - Una volta che tutte le goroutine hanno completato il loro lavoro, il master attende la loro conclusione con il metodo `wg.Wait()` della libreria `sync`, garantendo che tutte le operazioni di mappatura siano completate prima di rispondere al client.

### 3.5. Worker

Il nodo worker implementa il server gRPC che offre servizi di mappatura e riduzione per il framework MapReduce. Ogni worker gestisce una parte del processo di elaborazione distribuita, suddividendo il lavoro tra più processi in parallelo. Il flusso di esecuzione del worker è il seguente:

1. **Configurazione e Inizializzazione**:
   - Il worker legge la configurazione da un file JSON (`configWorker.json`), che include informazioni sui worker e il loro numero.
   - La porta su cui il worker offre il servizio gRPC viene specificata tramite una flag della riga di comando. La porta deve essere valida (compresa tra 1 e 65535).

2. **Gestione delle Richieste di Mappatura**:
   - Il worker implementa il metodo `Mapping`, che riceve un `Chunk` contenente un array di numeri. Questi numeri vengono ordinati in modo crescente e successivamente partizionati utilizzando una funzione di hashing tipica del **Terasort**. L'hashing distribuisce i numeri tra i vari reducers in base al numero totale di workers.
   - I numeri vengono partizionati in `nWorkers` bucket, dove ogni bucket è destinato a un reducer. La funzione di hash (`hashPartition`) determina il "bucket" per ciascun numero.

3. **Chiamate ai Reducer**:
   - Una volta partizionati i numeri, il worker invia ciascuna partizione a un altro worker (un reducer) utilizzando il servizio gRPC. La comunicazione avviene tramite goroutine per garantire l'elaborazione parallela.
   - La funzione `Mapping` termina con il completamento di tutte le chiamate ai reducer.

4. **Gestione delle Richieste di Riduzione**:
   - Il worker implementa il metodo `Reducing`, che riceve i dati da un mapper. I dati vengono memorizzati in una mappa (`ReduceRequest`), dove ogni richiesta è identificata tramite un ID unico.
   - Quando il worker riceve tutti i chunk di dati per una specifica richiesta, esegue l'ordinamento dei numeri ricevuti tramite la funzione `Sort`, che unisce tutte le porzioni di dati in un unico array ordinato utilizzando un merge sort ottimizzato con il fatto che le singole righe della matrice sono già ordinate.
   - Una volta ordinati i dati, il worker scrive i risultati su file (in formato `.txt`) utilizzando la funzione `WriteResultToFile`.

5. **Sincronizzazione e Gestione della Concorrenza**:
   - Il worker utilizza la libreria `sync` per sincronizzare le operazioni concorrenti. In particolare, un **mutex** viene utilizzato per proteggere l'accesso alla struttura dati che memorizza le richieste di riduzione in corso, evitando conflitti tra più goroutine.


7. **Scrittura dei Risultati**:
   - Una volta completato l'ordinamento dei dati, i risultati vengono scritti in un file di output, organizzato per ID di richiesta, e salvato nella directory `output/`. Ogni file contiene una lista di numeri ordinati, separati da una nuova riga.
