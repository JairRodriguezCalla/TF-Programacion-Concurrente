package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
)

type Registro struct {
	MesFacturacion  string
	CodEmpresa      string
	Grupo           string
	Uso             string
	PromedioConsumo string
	CodTarifa       string
}

func worker(id int, jobs <-chan Registro, wg *sync.WaitGroup) {
	defer wg.Done()
	for registro := range jobs {
		fmt.Printf("[Worker %d] Mes: %s | Empresa: %s | Consumo: %s kWh | Tarifa: %s\n",
			id, registro.MesFacturacion, registro.CodEmpresa, registro.PromedioConsumo, registro.CodTarifa)
	}
}

func main() {
	file, err := os.Open("../data/facturacion.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	jobs := make(chan Registro, len(records))
	var wg sync.WaitGroup

	// Lanzamos 3 workers concurrentes
	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go worker(w, jobs, &wg)
	}

	for i, row := range records {
		if i == 0 {
			continue // omitir cabecera
		}
		if len(row) < 6 {
			continue // evitar filas incompletas
		}
		jobs <- Registro{
			MesFacturacion:  row[0],
			CodEmpresa:      row[1],
			Grupo:           row[2],
			Uso:             row[3],
			PromedioConsumo: row[4],
			CodTarifa:       row[5],
		}
	}

	close(jobs)
	wg.Wait()
	fmt.Println("Procesamiento concurrente finalizado.")
}
