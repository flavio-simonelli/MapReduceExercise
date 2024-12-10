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
	"sort"
	"sync"
	"time"
)

var (
	port        = flag.Int("port", 0, "Porta su cui il worker offre il servizio (obbligatorio: numero da 1 a 65535)")
	workersList []config.Worker
	nWorkers    int
)

// Struct Worker che implementa i servizi gRPC
type Worker struct {
	pb.UnimplementedWorkerServiceServer
	ReduceQueue map[ReduceQueueKey][][]int
	Mutex       sync.Mutex
}

type ReduceQueueKey struct {
	Client    string
	Timestamp time.Time
}

// Funzione di hashing per partizionare i numeri nella fase di mapping in base al numero di reducers
func hashPartition(key int32, numReducers int32) int32 {
	return key % numReducers
}

// Mapping: implementazione della funzione Mapping
func (w *Worker) Mapping(ctx context.Context, chunk *pb.Chunk) (*pb.Response, error) {

	// Ordinamento dei numeri ricevuti (ordina in modo crescente)
	sort.Slice(chunk.Numbers, func(i, j int) bool {
		return chunk.Numbers[i] < chunk.Numbers[j]
	})
	log.Printf("Sorted numbers: %v", chunk.Numbers)

	// Partizionamento in modo efficiente, idea del Terasort, per suddividere il chunk e distribuirlo ai vari reducer
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

		go func(i int, worker config.Worker) {
			defer wg.Done()                                         // decrementiamo il contatore wg quando la goroutine è completata
			address := fmt.Sprintf("%s:%d", worker.IP, worker.Port) // creiamo l'indirizzo del worker da contattare
			partition := partitionedChunks[i]                       // partizione da inviare all'i-esimo reducer

			conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials())) // dial
			if err != nil {
				log.Fatalf("Errore nella connessione al worker %d: %v\n", i, err)
			}
			defer conn.Close()                                                            // chiusura della connessione alla fine della goroutine
			client := pb.NewWorkerServiceClient(conn)                                     // creazione del client grpc
			_, err = client.Reducing(context.Background(), &pb.Chunk{Numbers: partition}) // effettua la chiamata gRPC di Mapping
			if err != nil {
				log.Fatalf("Errore nella chiamata Reducing al worker %d: %v\n", i, err)
			}
		}(i, worker)
	}
	wg.Wait()                           // aspettiamo che tutte le goroutine abbiano completato
	return &pb.Response{Ack: true}, nil // Ritorna un ACK positivo
}

// Reducing: implementazione della funzione Reducing
func (w *Worker) Reducing(ctx context.Context, chunk *pb.Chunk) (*pb.Response, error) {
	log.Printf("Reducing called with numbers: %v", chunk.Numbers)

	// Logica di esempio: somma i numeri ricevuti
	sum := 0
	for _, num := range chunk.Numbers {
		sum += int(num)
	}
	log.Printf("Reducing result (sum): %d", sum)

	// Ritorna un ACK positivo
	return &pb.Response{Ack: true}, nil
}

func portsetting() (int, error) {
	// effettuiamo il parsing
	flag.Parse()
	// Validazione della porta
	if *port < 1 || *port > 65535 {
		flag.Usage()
		return 0, fmt.Errorf("La porta deve essere un numero tra 1 e 65535")
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

	// COnfigurazione delle variabili globali per la configurazione dei worker
	configFile, err := config.ReadConfigWorker()
	if err != nil {
		log.Fatalf("Errore nella lettura della configurazione: %v", err)
	}
	workersList = configFile.Workers
	nWorkers = len(workersList)

	// Avvia il listener TCP
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//Creazione del server gRPC
	grpcServer := grpc.NewServer()

	// Registra il servizio MapReduceService
	pb.RegisterWorkerServiceServer(grpcServer, &Worker{})

	// Avvia il server gRPC
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("impossibile avviare il worker: %v", err)
	}
	log.Printf("Worker online")
}
