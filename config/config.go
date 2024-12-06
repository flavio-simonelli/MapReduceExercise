package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFile = "config/config.json"

// Worker rappresenta un singolo worker con IP e porta.
type Worker struct {
	IP   string `json:"ip"`
	Port int32  `json:"port"`
}

// Config rappresenta la slice di worker configurati nel file JSON
type Config struct {
	Workers []Worker `json:"workers"`
}

// Funzione di utilità per leggere e parsare il file di configurazione dei worker
func ReadConfig() (*Config, error) {
	// Lettura del file JSON
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("errore nella lettura del file: %v", err)
	}

	// Parsing del file JSON nella struttura Config
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("errore nel parsing del JSON: %v", err)
	}

	// Restituisce la configurazione letta
	return &config, nil
}
