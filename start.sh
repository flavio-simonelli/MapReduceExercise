#!/bin/bash

# la posizione del file .sh
WORK_DIR=$""

# Numero di worker da avviare
NUM_WORKERS=4

# Porta base per i worker
BASE_PORT=50000

# Esegui 'go mod tidy' (se necessario)
cd "$WORK_DIR" || exit
go mod tidy

# Funzione per avviare un worker
start_worker() {
    port=$((BASE_PORT + $1))
    go run ./worker/worker.go -port $port &
}

# Avvia tutti i worker
# shellcheck disable=SC2004
for i in $(seq 0 $(($NUM_WORKERS-1))); do
    start_worker "$i"
done

# Avvia il master
go run ./master/master.go &
