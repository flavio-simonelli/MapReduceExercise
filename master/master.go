package main

import (
	"MapReduceExercise/config"
	pb "MapReduceExercise/proto/gen"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"sync"
)

var (
	workersList []config.Worker
	nWorkers    int
)

type Master struct {
	pb.UnimplementedMasterServiceServer
}

func (w *Master) NewRequest(_ context.Context, chunk *pb.Chunk) (*pb.Response, error) {
	// dividiamo la slice per ciascun worker
	baseSize := len(chunk.Numbers) / nWorkers // numeri per ogni worker (parte omogenea)
	rest := len(chunk.Numbers) % nWorkers     // numeri in eccesso che dovranno essere spartiti in modo disomogeneo
	chunks := make([][]int32, nWorkers)
	start := 0
	for i := 0; i < nWorkers; i++ {
		// Calcola la dimensione del chunk per il worker corrente
		end := start + baseSize
		if i < rest {
			end++ // Distribuisci il resto in modo uniforme
		}
		chunks[i] = chunk.Numbers[start:end]
		start = end
	}

	// effettuiamo la richiesta di mapping per ogni worker
	var wg sync.WaitGroup
	for i, worker := range workersList {
		wg.Add(1) // incrementiamo contatore wg

		go func(i int, worker config.Worker) {
			defer wg.Done()                                                                                // decrementiamo il contatore wg quando la goroutine è completata
			address := fmt.Sprintf("%s:%d", worker.IP, worker.Port)                                        // creiamo l'indirizzo del worker da contattare
			chunkToSend := chunks[i]                                                                       // chunk da inviare all'i-esimo worker
			conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials())) // dial
			if err != nil {
				log.Fatalf("Errore nella connessione al worker %d: %v\n", i, err)
			}
			defer conn.Close()                                                                                             // chiusura della connessione alla fine della goroutine
			client := pb.NewWorkerServiceClient(conn)                                                                      // creazione del client grpc
			_, err = client.Mapping(context.Background(), &pb.Chunk{Numbers: chunkToSend, IdRichiesta: chunk.IdRichiesta}) // effettua la chiamata gRPC di Mapping
			if err != nil {
				log.Fatalf("Errore nella chiamata Mapping al worker %d: %v\n", i, err)
			}
		}(i, worker)
	}
	wg.Wait() // aspettiamo che tutte le goroutine abbiano completato
	log.Printf("La richiesta %v è stata eseguita con successo", chunk.IdRichiesta)
	return &pb.Response{Ack: true}, nil
}

func main() {
	// lettura della configurazione del master dal file configMaster.json
	readConfigMaster, err := config.ReadConfigMaster()
	if err != nil {
		log.Fatalf("Errore nella lettura della configurazione del master: %v", err)
	}
	log.Println("lettura file configurazione del master completata")

	// lettura della configurazione dei worker dal file configWorker.json
	readConfigWorker, err := config.ReadConfigWorker()
	if err != nil {
		log.Fatalf("Errore nella lettura della configurazione dei worker: %v", err)
	}
	workersList = readConfigWorker.Workers
	nWorkers = len(workersList)
	log.Println("lettura file configurazione worker completata")

	// Avvia il listener TCP
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", readConfigMaster.Master.IP, readConfigMaster.Master.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("listen completata")

	//Creazione del server gRPC
	grpcServer := grpc.NewServer()

	// Registra il servizio MapReduceService
	pb.RegisterMasterServiceServer(grpcServer, &Master{})

	// Avvia il server gRPC
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("impossibile avviare il master: %v", err)
	}
	log.Printf("Master online")

}
