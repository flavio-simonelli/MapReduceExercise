package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFileWorker = "config/configWorker.json"
const configFileMaster = "config/configMaster.json"

// Worker rappresenta un singolo worker con IP e porta e identificativo univoco
type Worker struct {
	IP   string `json:"ip"`
	Port int32  `json:"port"`
	ID   int32  `json:"id"`
}

// ConfigWorker Config rappresenta la slice di worker configurati nel file JSON
type ConfigWorker struct {
	Workers []Worker `json:"workers"`
}

// ConfigMaster Struttura per rappresentare la configurazione master
type ConfigMaster struct {
	Master struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	} `json:"master"`
}

// ReadConfigWorker Funzione di utilità per leggere e parsare il file di configurazione dei worker
func ReadConfigWorker() (*ConfigWorker, error) {
	// Lettura del file JSON
	file, err := os.ReadFile(configFileWorker)
	if err != nil {
		return nil, fmt.Errorf("errore nella lettura del file: %v", err)
	}

	// Parsing del file JSON nella struttura Config
	var config ConfigWorker
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("errore nel parsing del JSON: %v", err)
	}
	return &config, nil // Restituisce la configurazione letta
}

// ReadConfigMaster Funzione di utilità per leggere e parsare il file di configurazione master
func ReadConfigMaster() (*ConfigMaster, error) {
	// Lettura del file JSON
	file, err := os.ReadFile(configFileMaster)
	if err != nil {
		return nil, fmt.Errorf("errore nella lettura del file: %v", err)
	}

	// Parsing del file JSON nella struttura ConfigMaster
	var config ConfigMaster
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("errore nel parsing del JSON: %v", err)
	}
	return &config, nil // Restituisce la configurazione letta
}
