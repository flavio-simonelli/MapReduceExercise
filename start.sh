#!/bin/bash

# la posizione del file .sh
WORK_DIR=$(dirname "$(realpath "$0")")

# Numero di worker da avviare
NUM_WORKERS=4

# Porta base per i worker
BASE_PORT=5000

# Esegui 'go mod tidy' (se necessario)
cd "$WORK_DIR"
go mod tidy

# Funzione per avviare un worker
start_worker() {
    port=$((BASE_PORT + $1))
    cd "$WORK_DIR/worker"
    go run worker -p $port &
}

# Avvia tutti i worker
for i in $(seq 1 $NUM_WORKERS); do
    start_worker $i
done

# Avvia il master
cd "$WORK_DIR/master"
go run master &
