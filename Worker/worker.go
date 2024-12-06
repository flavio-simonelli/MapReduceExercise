package main

import (
	pb "MapReduceExercise/proto"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"sort"
)

var (
	port = flag.Int("port", 0, "Porta su cui il Worker offre il servizio (obbligatorio: numero da 1 a 65535)")
)

// Struct Worker che implementa i servizi gRPC
type Worker struct {
	pb.UnimplementedMapReduceServiceServer
}

// Mapping: implementazione della funzione Mapping
func (w *Worker) Mapping(ctx context.Context, chunk *pb.Chunk) (*pb.Response, error) {
	log.Printf("Mapping called with numbers: %v", chunk.Numbers)

	// Ordinamento dei numeri ricevuti
	sort.Ints(chunk.Numbers)
	log.Printf("Sorted numbers: %v", chunk.Numbers)

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
	log.Printf("Worker listening at %v", port)

	//Creazione del server gRPC
	grpcServer := grpc.NewServer()

	// Registra il servizio MapReduceService
	pb.RegisterMapReduceServiceServer(grpcServer, &Worker{})
	log.Printf("Worker registered successfully")

	// Avvia il server gRPC
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("impossibile avviare il Worker: %v", err)
	}
}
