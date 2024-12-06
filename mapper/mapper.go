package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "MapReduceExercise/proto/gen"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 0, "Porta su cui il mapper offre il servizio (obbligatorio: numero da 1 a 65535)")
)

// Server che implementa il servizio MapReduceService
type mapReduceServiceServer struct{}

// Implementa il metodo Mapping del servizio MapReduceService
func (s *mapReduceServiceServer) Mapping(ctx context.Context, chunk *pb.Chunk) (*pb.Response, error) {
	log.Printf("Mapping ricevuto chunk con numeri: %v", chunk.Numbers)

	// Logica di Mapping: ad esempio moltiplichiamo ogni numero per 2
	for i := range chunk.Numbers {
		chunk.Numbers[i] = chunk.Numbers[i] * 2
	}

	// Restituiamo una risposta con ack = true
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
	log.Printf("server listening at %v", port)

	//Creazione del server gRPC
	grpcServer := grpc.NewServer()

	pb.RegisterMapReduceServer(grpcServer, &mapReduceServiceServer{})

	// Avvia il server gRPC
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("impossibile avviare il mapper: %v", err)
	}
}
