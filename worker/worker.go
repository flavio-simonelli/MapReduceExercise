package main

import (
	"MapReduceExercise/config"
	pb "MapReduceExercise/proto"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"os"
	"sort"
)

var (
	port = flag.Int("port", 0, "Porta su cui il worker offre il servizio (obbligatorio: numero da 1 a 65535)")
)

// Struct Worker che implementa i servizi gRPC
type Worker struct {
	pb.UnimplementedMapReduceServiceServer
}

// Funzione di hashing per partizionare i numeri nella fase di mapping in base al numero di reducers
func hashPartition(key int32, numReducers int32) int32 {
	return key % numReducers
}

// Mapping: implementazione della funzione Mapping
func (w *Worker) Mapping(ctx context.Context, chunk *pb.Chunk) (*pb.Response, error) {
	log.Printf("Mapping called with numbers: %v", chunk.Numbers)

	// Ordinamento dei numeri ricevuti (ordina in modo crescente)
	sort.Slice(chunk.Numbers, func(i, j int) bool {
		return chunk.Numbers[i] < chunk.Numbers[j]
	})
	log.Printf("Sorted numbers: %v", chunk.Numbers)

	//legge la configurazione dei worker dal file di configurazione per sapere ip, porta per ciascun worker e sapere quanti sono
	readConfig, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("Errore nella lettura della configurazione: %v", err)
		return nil, err
	}
	// calcolo il numero di reducer
	numReducers := int32(len(readConfig.Workers))

	// Utilizziamo la funzione di hash per dividere il chunk e distribuirlo ai vari reducer
	// Creiamo una slice di slice per contenere i chunk partizionati per ogni reducer
	partitionedChunks := make([][]int32, numReducers)
	// Iteriamo sui numeri nel chunk e li assegnamo al reducer corretto
	for _, number := range chunk.Numbers {
		// Determiniamo l'indice del reducer tramite la funzione di hash
		reducerIndex := hashPartition(number, numReducers)
		// Aggiungiamo il numero al "bucket" del reducer corrispondente
		partitionedChunks[reducerIndex] = append(partitionedChunks[reducerIndex], number)
	}

	// Invio dei dati partizionati ai vari reducers
	for i, chunk := range partitionedChunks {
		// Recuperiamo l'indirizzo del reducer
		address := fmt.Sprintf("%s:%d", readConfig.Workers[i].IP, readConfig.Workers[i].Port)
		log.Printf("Connettendosi al Reducer %d all'indirizzo %s...\n", i+1, address)

		// Creiamo la connessione gRPC al reducer
		conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Errore nella connessione al Reducer %d: %v\n", i+1, err)
			continue
		}
		defer conn.Close()

		// Creiamo il client gRPC per il reducer
		client := pb.NewMapReduceServiceClient(conn)

		// Inviamo il chunk al reducer tramite gRPC
		response, err := client.Reducing(context.Background(), &pb.Chunk{Numbers: chunk})
		if err != nil {
			log.Printf("Errore nella chiamata Reducing al Reducer %d: %v\n", i+1, err)
			continue
		}

		// Stampa la risposta del reducer
		log.Printf("Risposta del Reducer %d: Ack=%v\n", i+1, response.Ack)
	}

	// Ritorna un ACK positivo
	return &pb.Response{Ack: true}, nil
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
	// Leggi porta dalla riga di comando
	port, error := portsetting()
	if error != nil {
		fmt.Println("Errore:", error)
		os.Exit(1)
	}
	log.Printf("il numero di porta scelto è: %d\n", port)

	// Avvia il listener TCP
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("worker listening at %v", port)

	//Creazione del server gRPC
	grpcServer := grpc.NewServer()

	// Registra il servizio MapReduceService
	pb.RegisterMapReduceServiceServer(grpcServer, &Worker{})
	log.Printf("worker registered successfully")

	// Avvia il server gRPC
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("impossibile avviare il worker: %v", err)
	}
}
