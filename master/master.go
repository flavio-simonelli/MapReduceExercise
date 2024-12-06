package main

import (
	pb "MapReduceExercise/proto"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/rand"
	"os"
	"time"
)

// Costante per il nome del file JSON
const configFile = "config/config.json"

// Worker rappresenta un singolo worker con IP e porta.
type Worker struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// Config rappresenta la slice di worker configurati nel file JSON
type Config struct {
	Workers []Worker `json:"workers"`
}

// Funzione che restituisce una slice di numeri interi di lunghezza arbitraria input del Master da ordinare
func generatoreNumeri() []int {
	// Inizializzazione del generatore di numeri casuali con seed il timestamp
	rand.Seed(time.Now().UnixNano())
	// Creazione di una slice vuota di dimensione randomica da 1 a 100
	numeri := make([]int, rand.Intn(100)+1)
	// Popolamento della slice con numeri casuali
	for i := 0; i < len(numeri); i++ {
		numeri[i] = rand.Intn(100) // Numeri casuali tra 0 e 99
	}
	return numeri
}

func main() {
	// Lettura della configurazione dei workers scritti nel file JSON
	file, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Errore nella lettura del file: %v\n", err)
		return
	}
	// Parsing del file JSON nella struttura Config
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		fmt.Printf("Errore nel parsing del JSON: %v\n", err)
		return
	}
	fmt.Printf("Numero di worker: %d\n", len(config.Workers))

	// Stampo i worker in una slice
	workers := config.Workers
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
	var chunks [][]int
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
		log.Printf("Connettendosi al Worker %d all'indirizzo %s...\n", i+1, address)

		// Connessione gRPC
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Errore nella connessione al Worker %d: %v\n", i+1, err)
			continue
		}
		defer conn.Close()

		// Crea il client gRPC per il worker
		client := pb.NewMapReduceServiceClient(conn)

		// Prepara il chunk da inviare
		chunk := chunks[i]
		log.Printf("Invio chunk al Worker %d: %v\n", i+1, chunk)

		// Effettua la chiamata gRPC Mapping
		response, err := client.Mapping(context.Background(), &pb.Chunk{Numbers: toInt32Slice(chunk)})
		if err != nil {
			log.Printf("Errore nella chiamata Mapping al Worker %d: %v\n", i+1, err)
			continue
		}

		// Stampa la risposta del worker
		log.Printf("Risposta del Worker %d: Ack=%v\n", i+1, response.Ack)
	}
}

// Funzione di utilità per convertire []int in []int32
func toInt32Slice(numbers []int) []int32 {
	int32Slice := make([]int32, len(numbers))
	for i, num := range numbers {
		int32Slice[i] = int32(num)
	}
	return int32Slice
}
