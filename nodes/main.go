package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"

	"nodes/rf"
)

func main() {
	fmt.Println("🧠 Iniciando procesamiento PC3...")

	// Descomenta uno a la vez según lo que quieras probar:
	runRandomForestPrediction()
	// workers.RunWorkerPool()
	// fan.RunFanOutFanin()
}

// Ejecuta el procesamiento de datos reales con Random Forest
func runRandomForestPrediction() {
	file, err := os.Open("../data/facturacion_encoded.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for i, row := range records {
		if i == 0 {
			continue // saltar cabecera
		}
		wg.Add(1)
		go rf.ProcessRow(row, &wg)
	}
	wg.Wait()

	fmt.Println("✅ Procesamiento concurrente finalizado.")
}
