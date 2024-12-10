package main

import (
	"MapReduceExercise/config"
	pb "MapReduceExercise/proto/gen"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"sync"
	"time"
)

var (
	workersList []config.Worker
	nWorkers    int
)

// Genera una lista di numeri di dimensione arbitraria (Input della richiesta MapReduce)
func generatoreNumeri() []int32 {
	// creazione randomizzatore (seed timestamp)
	rand.NewSource(time.Now().UnixNano())
	// creazione slice di lunghezza arbitraria (1-100)
	numeri := make([]int32, rand.Intn(100)+1)
	// Popolamento della slice con numeri casuali
	for i := 0; i < len(numeri); i++ {
		numeri[i] = int32(rand.Intn(100)) // Numeri casuali tra 0 e 99
	}
	return numeri
}

func main() {

	// lettura della configurazione dei worker dal file config.json
	readConfig, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("Errore nella lettura della configurazione: %v", err)
	}
	workersList = readConfig.Workers
	nWorkers = len(workersList)

	// generiamo i numeri della richiesta
	numeri := generatoreNumeri()
	lenList := len(numeri)
	log.Printf("Sono stati generari %d numeri :\n %v \n", lenList, numeri)

	// dividiamo la slice per ciascun worker
	baseSize := lenList / nWorkers // numeri per ogni worker (parte omogenea)
	rest := lenList % nWorkers     // numeri in eccesso che dovranno essere spartiti in modo disomogeneo
	chunks := make([][]int32, nWorkers)
	start := 0
	for i := 0; i < nWorkers; i++ {
		// Calcola la dimensione del chunk per il worker corrente
		end := start + baseSize
		if i < rest {
			end++ // Distribuisci il resto in modo uniforme
		}
		chunks[i] = numeri[start:end]
		start = end
	}

	// effettuiamo la richiesta di mapping per ogni worker
	var wg sync.WaitGroup
	for i, worker := range workersList {
		wg.Add(1) // incrementiamo contatore wg

		go func(i int, worker config.Worker) {
			defer wg.Done()                                                                                // decrementiamo il contatore wg quando la goroutine è completata
			address := fmt.Sprintf("%s:%d", worker.IP, worker.Port)                                        // creiamo l'indirizzo del worker da contattare
			chunk := chunks[i]                                                                             // chunk da inviare all'i-esimo worker
			conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials())) // dial
			if err != nil {
				log.Fatalf("Errore nella connessione al worker %d: %v\n", i, err)
			}
			defer conn.Close()                                                       // chiusura della connessione alla fine della goroutine
			client := pb.NewMapReduceServiceClient(conn)                             // creazione del client grpc
			_, err = client.Mapping(context.Background(), &pb.Chunk{Numbers: chunk}) // effettua la chiamata gRPC di Mapping
			if err != nil {
				log.Fatalf("Errore nella chiamata Mapping al worker %d: %v\n", i, err)
			}

		}(i, worker)
	}
	wg.Wait() // aspettiamo che tutte le goroutine abbiano completato
	log.Println("La richiesta è stata eseguita con successo")
}
