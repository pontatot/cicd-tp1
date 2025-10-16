package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-web-service/models"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// DB is the global database connection pool
var DB *sql.DB

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of HTTP requests received",
		},
		[]string{"path"},
	)
)

func main() {
	prometheus.MustRegister(requestCount)

	// Get environment variables with defaults
	addr := getEnvOrDefault("CITY_API_ADDR", "127.0.0.1")
	port := getEnvOrDefault("CITY_API_PORT", "2022")

	// Required environment variables
	dbURL := requireEnv("CITY_API_DB_URL")
	dbUser := requireEnv("CITY_API_DB_USER")
	dbPwd := requireEnv("CITY_API_DB_PWD")

	// Initialize database connection
	connStr := fmt.Sprintf("postgresql://%s:%s@%s?sslmode=disable", dbUser, dbPwd, dbURL)
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	// Test database connection
	if err := DB.Ping(); err != nil {
		log.Fatal(err)
	}

	// Setup routes
	http.HandleFunc("/city", CityHandler)
	http.HandleFunc("/_health", HealthHandler)

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Server starting on %s:%s", addr, port)
	log.Fatal(http.ListenAndServe(addr+":"+port, nil))
}

// CityHandler handles HTTP requests for city resources,
// supporting GET (list all cities) and POST (create new city) methods
func CityHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var city models.City
		if err := json.NewDecoder(r.Body).Decode(&city); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := DB.Exec(`
            INSERT INTO city (id, department_code, insee_code, zip_code, name, lat, lon)
            VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			city.ID, city.DepartmentCode, city.InseeCode, city.ZipCode, city.Name, city.Lat, city.Lon)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

	case http.MethodGet:
		rows, err := DB.Query("SELECT id, department_code, insee_code, zip_code, name, lat, lon FROM city")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var cities []models.City
		for rows.Next() {
			var city models.City
			if err := rows.Scan(&city.ID, &city.DepartmentCode, &city.InseeCode, &city.ZipCode, &city.Name, &city.Lat, &city.Lon); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			cities = append(cities, city)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cities)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HealthHandler responds to health check requests
// Returns 200 OK if the service is healthy
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func requireEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}
