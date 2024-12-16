package main

import (
	"MapReduceExercise/config"
	pb "MapReduceExercise/proto/gen"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"os"
	"sort"
	"sync"
)

var (
	port        = flag.Int("port", 0, "Porta su cui il worker offre il servizio (obbligatorio: numero da 1 a 65535)")
	workersList []config.Worker
	nWorkers    int
)

// Worker Struct Worker che implementa i servizi gRPC
type Worker struct {
	pb.UnimplementedWorkerServiceServer
	ReduceRequest map[string][][]int32
	Mutex         sync.Mutex
	idWorker      int32
}

// Funzione di hashing per partizionare i numeri nella fase di mapping in base al numero di reducers
func hashPartition(key int32, numReducers int32) int32 {
	return key % numReducers
}

// Mapping implementazione della funzione Mapping
func (w *Worker) Mapping(_ context.Context, chunk *pb.Chunk) (*pb.Response, error) {

	// Ordinamento dei numeri ricevuti (ordina in modo crescente)
	sort.Slice(chunk.Numbers, func(i, j int) bool {
		return chunk.Numbers[i] < chunk.Numbers[j]
	})

	// Partizionamento in modo efficiente, idea del Tera-sort, per suddividere il chunk e distribuirlo ai vari reducer
	partitionedChunks := make([][]int32, nWorkers)
	// Assegniamo ciascun numero del chunk al reducer corrispondente tramite la funzione di hash definita
	for _, number := range chunk.Numbers {
		reducerIndex := hashPartition(number, int32(nWorkers))                            // Determiniamo l'indice del reducer tramite la funzione di hash
		partitionedChunks[reducerIndex] = append(partitionedChunks[reducerIndex], number) // Aggiungiamo il numero al "bucket" del reducer corrispondente
	}

	// effettuiamo la richiesta di reducing per ogni reducer
	var wg sync.WaitGroup
	for i, worker := range workersList {
		wg.Add(1) // incrementiamo contatore wg

		go func(i int, worker config.Worker) error {
			defer wg.Done()                                         // decrementiamo il contatore wg quando la goroutine è completata
			address := fmt.Sprintf("%s:%d", worker.IP, worker.Port) // creiamo l'indirizzo del worker da contattare
			partition := partitionedChunks[i]                       // partizione da inviare all i-esimo reducer

			conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials())) // dial
			if err != nil {
				log.Printf("Errore nella connessione al worker %d: %v\n", i, err)
				return err
			}
			defer conn.Close()                                                                                            // chiusura della connessione alla fine della goroutine
			client := pb.NewWorkerServiceClient(conn)                                                                     // creazione del client grpc
			_, err = client.Reducing(context.Background(), &pb.Chunk{Numbers: partition, IdRichiesta: chunk.IdRichiesta}) // effettua la chiamata gRPC di Mapping
			if err != nil {
				log.Printf("Errore nella chiamata Reducing al worker %d: %v\n", i, err)
				return err
			}
			return nil
		}(i, worker)
	}
	wg.Wait() // aspettiamo che tutte le goroutine abbiano completato
	log.Printf("Il mapper %d ha ordinato e partizionato i seguenti numeri: %v\n", w.idWorker, chunk.Numbers)
	return &pb.Response{Ack: true}, nil // Ritorna un ACK positivo
}

// Sort ordina una matrice di righe già ordinate in un unico vettore di int32.
func Sort(matrix [][]int32) []int32 {
	if len(matrix) == 0 {
		return []int32{}
	}
	var result []int32
	indices := make([]int, len(matrix)) // Indici per tracciare la posizione corrente in ogni riga
	for {
		minVal := int32(^uint32(0) >> 1) // Massimo valore int32
		minRow := -1
		// Trova il valore minimo tra i primi elementi disponibili di ogni riga
		for i := 0; i < len(matrix); i++ {
			if indices[i] < len(matrix[i]) && matrix[i][indices[i]] < minVal {
				minVal = matrix[i][indices[i]]
				minRow = i
			}
		}
		// Se non c'è più nessun elemento da processare, esci dal loop
		if minRow == -1 {
			break
		}
		// Aggiungi il valore minimo al risultato e avanza l'indice della riga corrispondente
		result = append(result, minVal)
		indices[minRow]++
	}
	return result
}

// WriteResultToFile Funzione per scrivere i risultati ordinati su file
func WriteResultToFile(numbers []int32, filename string) error {
	nameFile := fmt.Sprintf("output/%s.txt", filename)
	// Creiamo il file o lo apriamo in modalità scrittura (crea un nuovo file se non esiste)
	file, err := os.Create(nameFile)
	if err != nil {
		return fmt.Errorf("errore nella creazione del file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			err = fmt.Errorf("errore durante la chiusura del file: %v", err)
		}
	}(file) // Assicuriamoci di chiudere il file alla fine
	// Scriviamo i numeri ordinati nel file, separandoli da una nuova riga
	for _, num := range numbers {
		_, err := fmt.Fprintln(file, num)
		if err != nil {
			return fmt.Errorf("errore nella scrittura del numero %d nel file: %v", num, err)
		}
	}
	// Operazione riuscita
	return nil
}

// Reducing implementazione della funzione Reducing
func (w *Worker) Reducing(_ context.Context, chunk *pb.Chunk) (*pb.Response, error) {
	w.Mutex.Lock()                                                                                 // blocchiamo la scrittura sulla struttura che mantiene tutte le richieste pendenti
	w.ReduceRequest[chunk.IdRichiesta] = append(w.ReduceRequest[chunk.IdRichiesta], chunk.Numbers) // aggiungiamo la nuova porzione
	// controlliamo che abbiamo ricevuto tutte le porzioni da ogni mapper
	if len(w.ReduceRequest[chunk.IdRichiesta]) == nWorkers { // se sono arrivati tutti i messaggi
		res := Sort(w.ReduceRequest[chunk.IdRichiesta])                 // ordiniamo le porzioni che sono arrivate
		filename := fmt.Sprintf("%s_%d", chunk.IdRichiesta, w.idWorker) // costruiamo il nome del file di output
		err := WriteResultToFile(res, filename)
		if err != nil {
			log.Printf("Errore nella save result: %v", err)
			return &pb.Response{Ack: false}, err
		}
		log.Printf("il reducer %d ha ordinato stampato su file i seguenti numeri %v\n", w.idWorker, res)
		delete(w.ReduceRequest, chunk.IdRichiesta)
	}
	w.Mutex.Unlock()
	return &pb.Response{Ack: true}, nil
}

func portsetting() (int, error) {
	// effettuiamo il parsing
	flag.Parse()
	// Validazione della porta
	if *port < 1 || *port > 65535 {
		flag.Usage()
		return 0, fmt.Errorf("la porta deve essere un numero tra 1 e 65535")
	}
	// restituiamo il valore della porta inserita
	return *port, nil
}

func main() {

	// Leggo la porta da riga di comando
	port, err := portsetting()
	if err != nil {
		log.Fatalf("Errore nella lettura della porta selezionata: %v", err)
	}

	// Configurazione delle variabili globali per la configurazione dei worker
	configFile, err := config.ReadConfigWorker()
	if err != nil {
		log.Fatalf("Errore nella lettura della configurazione: %v", err)
	}
	workersList = configFile.Workers
	nWorkers = len(workersList)
	var id int32 = -1 // Impostiamo un valore di default per id (ad esempio -1, che indica che non è stato trovato)
	for _, w := range workersList {
		if w.Port == int32(port) {
			id = w.ID
			break // Se troviamo il worker, usciamo dal ciclo
		}
	}
	if id == -1 { // Se id non è stato modificato, significa che la porta non è stata trovata
		log.Fatalf("La porta %d non è congrua con il file configworker.json oppure si è impostato -1 come id del worker (per favore cambiare id)", port)
	}

	// Avvia il listener TCP
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//Creazione del server gRPC
	grpcServer := grpc.NewServer()

	// Registra il servizio MapReduceService
	pb.RegisterWorkerServiceServer(grpcServer, &Worker{ReduceRequest: make(map[string][][]int32), idWorker: id})

	// Avvia il server gRPC
	log.Printf("worker %d online", id)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("impossibile avviare il worker: %v", err)
	}
}
