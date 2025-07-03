package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"              // Nueva importación para BSON
	"go.mongodb.org/mongo-driver/mongo"         // Nueva importación para el driver de Mongo
	"go.mongodb.org/mongo-driver/mongo/options" // Nueva importación para opciones de Mongo
)

var (
	ctx                   = context.Background()
	rdb                   *redis.Client
	mongoClient           *mongo.Client     // Cliente de MongoDB
	predictionsCollection *mongo.Collection // Colección para guardar predicciones
)

// -------- helpers ----------
func respond(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	// Habilitar CORS para que el frontend pueda llamar a la API
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
	json.NewEncoder(w).Encode(v)
}

// Estructura para el request de predicción que viene del frontend
type PredictRequest struct {
	Consumo float64 `json:"consumo"`
	Uso     int     `json:"uso"`
	Grupo   int     `json:"grupo"`
	Empresa int     `json:"empresa"`
}

// Estructura para el response de la predicción al frontend
type PredictResponse struct {
	Tarifa  int    `json:"tarifa"`
	Message string `json:"message,omitempty"`
}

// Estructura para guardar la predicción en MongoDB
type PredictionRecord struct {
	ID          string    `bson:"_id,omitempty"` // ID único para el registro
	JobID       string    `bson:"job_id"`
	Consumo     float64   `bson:"consumo"`
	Uso         int       `bson:"uso"`
	Grupo       int       `bson:"grupo"`
	Empresa     int       `bson:"empresa"`
	Tarifa      int       `bson:"tarifa"`
	LatencyNs   int64     `bson:"latency_ns"`
	WorkerID    string    `bson:"worker_id"`
	FinishedAt  int64     `bson:"finished_at"`
	PredictedAt time.Time `bson:"predicted_at"` // Campo para la marca de tiempo de la predicción
}

// -------- handlers ----------
func status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	total, _ := rdb.Get(ctx, "tarifa:stats:total").Int64()
	sumLat, _ := rdb.Get(ctx, "tarifa:stats:lat_total_ns").Int64()
	latCnt, _ := rdb.Get(ctx, "tarifa:stats:lat_count").Int64()

	var prom int64
	if latCnt > 0 {
		prom = sumLat / latCnt
	}

	respond(w, map[string]any{
		"total_jobs":      total,
		"latency_ns_prom": prom,
	})
}

func statsTarifas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	keys, _ := rdb.Keys(ctx, "tarifa:stats:tarifa:*").Result()
	data := map[string]int64{}
	for _, k := range keys {
		v, _ := rdb.Get(ctx, k).Int64()
		t := k[len("tarifa:stats:tarifa:"):]
		data[t] = v
	}
	respond(w, data)
}

func statsWorkers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	keys, _ := rdb.Keys(ctx, "tarifa:stats:worker:*").Result()
	data := map[string]int64{}
	for _, k := range keys {
		v, _ := rdb.Get(ctx, k).Int64()
		id := k[len("tarifa:stats:worker:"):]
		data[id] = v
	}
	respond(w, data)
}

// predictHandler maneja las solicitudes POST de predicción
func predictHandler(w http.ResponseWriter, r *http.Request) {
	// Habilitar CORS preflight para solicitudes POST
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req PredictRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Request inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Generar un Job ID único para rastrear la predicción
	jobID := fmt.Sprintf("pred_%d", time.Now().UnixNano())

	// Preparar el payload del job para Redis (similar a push_job)
	jobPayload, err := json.Marshal(map[string]interface{}{
		"ID":      jobID,
		"Consumo": req.Consumo,
		"Uso":     req.Uso,
		"Grupo":   req.Grupo,
		"Empresa": req.Empresa,
	})
	if err != nil {
		http.Error(w, "Error al serializar job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Enviar el job a la cola de Redis
	if err := rdb.LPush(ctx, "tarifa:jobs", jobPayload).Err(); err != nil {
		http.Error(w, "Error al enviar job a Redis: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("API: Job %s enviado a Redis para predicción.\n", jobID)

	// Esperar el resultado de la predicción desde Redis
	resultKey := "tarifa:result:" + jobID
	timeout := time.After(15 * time.Second)
	var predictedTarifa int
	var latencyNs int64
	var workerID string
	var finishedAt int64
	found := false

	for {
		select {
		case <-timeout:
			http.Error(w, "Timeout esperando resultado de predicción. Los workers pueden estar muy ocupados o la cola vacía.", http.StatusGatewayTimeout)
			return
		default:
			val, err := rdb.HGetAll(ctx, resultKey).Result()
			if err != nil {
				log.Printf("Error al leer resultado de Redis para %s: %v", jobID, err)
			}
			if len(val) > 0 {
				tarifaStr := val["tarifa"]
				t, err := strconv.Atoi(tarifaStr)
				if err != nil {
					log.Printf("Error al convertir tarifa '%s': %v", tarifaStr, err)
					http.Error(w, "Error interno procesando resultado.", http.StatusInternalServerError)
					return
				}
				predictedTarifa = t

				latencyNs, _ = strconv.ParseInt(val["latency_ns"], 10, 64)
				workerID = val["worker_id"]
				finishedAt, _ = strconv.ParseInt(val["finished_at"], 10, 64)

				found = true
				// Limpiar el resultado de Redis para este job después de leerlo
				rdb.Del(ctx, resultKey)
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		if found {
			break
		}
	}

	// === NUEVO: Guardar el registro completo de la predicción en MongoDB ===
	record := PredictionRecord{
		JobID:       jobID,
		Consumo:     req.Consumo,
		Uso:         req.Uso,
		Grupo:       req.Grupo,
		Empresa:     req.Empresa,
		Tarifa:      predictedTarifa,
		LatencyNs:   latencyNs,
		WorkerID:    workerID,
		FinishedAt:  finishedAt,
		PredictedAt: time.Now(), // Marca de tiempo actual de la predicción
	}

	_, err = predictionsCollection.InsertOne(ctx, record)
	if err != nil {
		log.Printf("Error al guardar predicción en MongoDB: %v", err)
		// No es crítico para la respuesta al usuario, pero es importante loggear
	} else {
		log.Printf("API: Predicción %s guardada en MongoDB (Tarifa: %d).\n", jobID, predictedTarifa)
	}
	// === FIN NUEVO ===

	// Responder al frontend con la tarifa predicha
	resp := PredictResponse{
		Tarifa:  predictedTarifa,
		Message: "Predicción exitosa",
	}
	respond(w, resp)
}

func main() {
	// --- Conexión a Redis ---
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	rdb = redis.NewClient(&redis.Options{Addr: redisAddr})

	// --- Conexión a MongoDB ---
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongo:27017" // URI por defecto para el contenedor de Mongo
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	var err error
	mongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Error al conectar a MongoDB: %v", err)
	}

	// Hacer ping para verificar la conexión
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Error al hacer ping a MongoDB: %v", err)
	}
	log.Println("✅ Conectado a MongoDB!")

	// Obtener la colección donde se guardarán las predicciones
	predictionsCollection = mongoClient.Database("tarifas_db").Collection("predicciones")
	log.Println("✅ Colección 'predicciones' lista.")

	// Registro de handlers
	http.HandleFunc("/status", status)
	http.HandleFunc("/stats/tarifas", statsTarifas)
	http.HandleFunc("/stats/workers", statsWorkers)
	http.HandleFunc("/predict", predictHandler) // Handler de predicción

	port := getenvDefault("PORT", "8080")
	log.Printf("🌐 API escuchando en :%s ...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
