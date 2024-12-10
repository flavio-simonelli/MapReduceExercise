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
	"time"
)

const client = "client1" // identificativo univoco di un client (può essere un token generato ad ogni connessione per questione di )

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
	// lettura della configurazione del nodo master
	readConfigMaster, err := config.ReadConfigMaster()
	if err != nil {
		log.Fatalf("ReadConfigMaster failed: %v", err)
	}
	address := fmt.Sprintf("%s:%d", readConfigMaster.Master.IP, readConfigMaster.Master.Port) // creiamo l'indirizzo del master da contattare

	//generiamo una sequenza di numeri random da inviare per la richiesta
	numeri := generatoreNumeri()
	lenList := len(numeri)
	log.Printf("Sono stati generari %d numeri :\n %v \n", lenList, numeri)
	//generiamo l'identificativo univoco della richiesta
	id := fmt.Sprintf("%s_%s", client, time.Now().Format("20060102-150405"))

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials())) // dial
	if err != nil {
		log.Fatalf("Errore nella connessione al server: %v\n", err)
	}
	defer conn.Close()
	client := pb.NewMasterServiceClient(conn)                                                     // creazione del client grpc
	_, err = client.NewRequest(context.Background(), &pb.Chunk{Numbers: numeri, IdRichiesta: id}) // effettua la chiamata gRPC di richiesta
	if err != nil {
		log.Fatalf("Errore nella richiesta: %v\n", err)
	}
}
