package main

import (
	"MapReduceExercise/config"
	pb "MapReduceExercise/proto"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"time"
)

// Funzione che restituisce una slice di numeri interi di lunghezza arbitraria input del Master da ordinare
func generatoreNumeri() []int32 {
	// Inizializzazione del generatore di numeri casuali con seed il timestamp
	rand.Seed(time.Now().UnixNano())
	// Creazione di una slice vuota di dimensione randomica da 1 a 100
	numeri := make([]int32, rand.Intn(100)+1)
	// Popolamento della slice con numeri casuali
	for i := 0; i < len(numeri); i++ {
		numeri[i] = int32(rand.Intn(100)) // Numeri casuali tra 0 e 99
	}
	return numeri
}

func main() {
	// lettura della configurazione dei worker
	readConfig, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("Errore nella lettura della configurazione: %v", err)
		return
	}

	fmt.Printf("Numero di worker: %d\n", len(readConfig.Workers))

	// Stampo i worker in una slice
	workers := readConfig.Workers
	fmt.Println("Lista dei workers:")
	for _, worker := range workers {
		fmt.Printf("IP: %s, Port: %d\n", worker.IP, worker.Port)
	}

	numeri := generatoreNumeri()
	// Stampa della slice con numeri generati e la sua lunghezza
	fmt.Printf("Numeri generati: %v\n", numeri)
	fmt.Printf("Lunghezza della slice: %d\n", len(numeri))

	// Calcoliamo quanti numeri assegnare a ciascun worker
	numeroWorkers := len(workers)
	numeroNumeri := len(numeri)
	baseSize := numeroNumeri / numeroWorkers // numeri per ogni worker (parte omogenea)
	resto := numeroNumeri % numeroWorkers    // numeri in eccesso che dovranno essere spartiti in modo disomogeneo

	// Invio dei chunk a ciascun worker
	// Creazione dei chunk da inviare
	var chunks [][]int32
	start := 0
	for i := 0; i < numeroWorkers; i++ {
		// Calcola la dimensione del chunk per il worker corrente
		chunkSize := baseSize
		if i < resto {
			chunkSize++ // Distribuisci il resto in modo uniforme
		}
		end := start + chunkSize
		chunks = append(chunks, numeri[start:end])
		start = end
	}

	// Invio dei chunk a ciascun worker tramite gRPC
	for i, worker := range workers {
		// Configura l'indirizzo del worker
		address := fmt.Sprintf("%s:%d", worker.IP, worker.Port)
		log.Printf("Connettendosi al worker %d all'indirizzo %s...\n", i+1, address)

		// Connessione gRPC
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Errore nella connessione al worker %d: %v\n", i+1, err)
			continue
		}
		defer conn.Close()

		// Crea il client gRPC per il worker
		client := pb.NewMapReduceServiceClient(conn)

		// Prepara il chunk da inviare
		chunk := chunks[i]
		log.Printf("Invio chunk al worker %d: %v\n", i+1, chunk)

		// Effettua la chiamata gRPC Mapping
		response, err := client.Mapping(context.Background(), &pb.Chunk{Numbers: chunk})
		if err != nil {
			log.Printf("Errore nella chiamata Mapping al worker %d: %v\n", i+1, err)
			continue
		}

		// Stampa la risposta del worker
		log.Printf("Risposta del worker %d: Ack=%v\n", i+1, response.Ack)
	}
}
