# Sistema Distribuito MapReduce per Ordinamento di Numeri  

## 1. Introduzione

Il presente repository implementa un sistema distribuito basato sul paradigma **MapReduce**, con l’obiettivo di elaborare dati in parallelo e distribuire il carico computazionale su più nodi worker.  
Il problema trattato consiste nell’**ordinamento crescente** di un insieme di numeri casuali, sfruttando l’architettura distribuita per ottimizzare le performance in scenari di elaborazione su larga scala.  
Questo sistema è stato realizzato nell’ambito di un esercizio del corso di Sistemi Distribuiti e Cloud Computing.  
La scelta del linguaggio di programmazione per la sua implementazione è **Go**.  
Per la comunicazione tra i vari componenti del sistema è stato adottato il protocollo **gRPC**.


## 2. Panoramica del sistema <br><small>schema di funzionamento</small>

```
        Client
           |
           v
        Master
       /   |   \
      v    v    v
  Worker Worker Worker ...
```

Il sistema è composto da tre principali componenti:  

### 2.1. Client  
- **Descrizione**:  
  Un client minimale (**tiny client**) progettato per generare richieste **rapide** relative al problema dell'ordinamento.  
- **Funzionalità**:  
  - Genera un numero arbitrario di interi casuali a runTime utilizzando come **seed** il timestamp attuale.  
  - Invia le richieste al nodo **Master**, specificando i numeri generati.  

### 2.2. Master  
- **Descrizione**:  
  Il server centrale responsabile della gestione delle richieste provenienti dai client e dell'avvio dell'operazione di computazione distribuita.  
- **Funzionalità**:  
  - Riceve le richieste dal client
  - Divide la lista di numeri in **chunk omogenei** per distribuirli ai nodi **mapper**.  

### 2.3. Worker  
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

### 3.2 Files di configurazione
Come richiesto dall'esercizio la configurazione della struttura del sistema distribuito è statica e non è previsto alcun meccanismo di configurazione dinamica, infatti i worker e il master attivo sono descritti nei seguenti file i configurazione. 
