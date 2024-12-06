package main

import (
	"encoding/json"
	"fmt"
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
	fmt.Printf("Resto: %d\n", resto)
	fmt.Printf("BaseSize: %d\n", baseSize)

	// Invio dei chunk a ciascun worker

}
