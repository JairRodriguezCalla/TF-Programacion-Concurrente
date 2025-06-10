package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"sync"
)

func predictTree1(consumo float64, uso int, grupo int, empresa int) int {
	if uso <= 1 {
		if consumo <= 564.5 {
			if consumo <= 341.5 {
				if consumo <= 0.5 {
					return 20
				} else {
					return 19
				}
			} else {
				if grupo <= 1 {
					return 19
				} else {
					return 19
				}
			}
		} else {
			if empresa <= 15 {
				if grupo <= 1 {
					return 4
				} else {
					return 2
				}
			} else {
				if consumo <= 3703.0 {
					return 22
				} else {
					return 9
				}
			}
		}
	} else {
		if empresa <= 10 {
			if uso <= 2 {
				if grupo <= 1 {
					return 12
				} else {
					return 19
				}
			} else {
				return 13
			}
		} else {
			return 9
		}
	}
}

func predictTree2(consumo float64, uso int, grupo int, empresa int) int {
	if empresa <= 15 {
		if empresa <= 0 {
			if consumo <= 25 {
				if consumo <= 9.5 {
					return 11
				} else {
					return 43
				}
			} else {
				if uso <= 1 {
					return 19
				} else {
					return 12
				}
			}
		} else {
			if uso <= 1 {
				if grupo <= 1 {
					return 2
				} else {
					return 22
				}
			} else {
				return 9
			}
		}
	} else {
		return 4
	}
}

func predictTree3(consumo float64, uso int, grupo int, empresa int) int {
	if consumo <= 30.5 {
		if uso <= 2 {
			if uso <= 1 {
				if grupo <= 0 {
					return 23
				} else {
					return 19
				}
			} else {
				if consumo <= 3.5 {
					return 3
				} else {
					return 1
				}
			}
		} else {
			return 0
		}
	} else {
		if empresa <= 4 {
			return 5
		} else {
			return 20
		}
	}
}

func predict(consumo float64, uso int, grupo int, empresa int) int {
	votes := make(map[int]int)
	result1 := predictTree1(consumo, uso, grupo, empresa)
	result2 := predictTree2(consumo, uso, grupo, empresa)
	result3 := predictTree3(consumo, uso, grupo, empresa)
	votes[result1]++
	votes[result2]++
	votes[result3]++

	maxVotes := 0
	predicted := -1
	for k, v := range votes {
		if v > maxVotes {
			maxVotes = v
			predicted = k
		}
	}
	return predicted
}

func processRow(row []string, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(row) < 8 {
		return
	}
	consumo, _ := strconv.ParseFloat(row[4], 64)
	uso, _ := strconv.Atoi(row[5])
	grupo, _ := strconv.Atoi(row[6])
	empresa, _ := strconv.Atoi(row[7])

	prediccion := predict(consumo, uso, grupo, empresa)
	fmt.Printf("➡️  Registro procesado: consumo=%.2f, uso=%d, grupo=%d, empresa=%d => Predicción: %d\n",
		consumo, uso, grupo, empresa, prediccion)
}

func main() {
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
		go processRow(row, &wg)
	}
	wg.Wait()
	fmt.Println("✅ Procesamiento concurrente finalizado.")
}
