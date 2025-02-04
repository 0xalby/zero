package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"
	"zero/config"
	"zero/db"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
)

type API struct {
	addr string
	db   *sql.DB
}

// Creates a new API server instance
func NewAPI(addr string, db *sql.DB) *API {
	return &API{
		addr: addr,
		db:   db,
	}
}

func init() {
	// Loading enviroment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load environment variables!", "err", err)
	}
	log.Info("Loaded .env enviroment variables!")
}

func main() {
	// Initializing
	config.InitJWT(os.Getenv("JWT_SECRET"))
	config.InitLogger()
	// Connecting to the database using the DATABASE_NAME
	database, err := db.Connect("db/" + os.Getenv("DATABASE_NAME"))
	if err != nil {
		log.Fatal("failed to connect to the database", "err", err)
	}
	defer database.Close()
	// Creating a new API server instance using the API_ADDRESS
	server := NewAPI(os.Getenv("API_ADDRESS"), database)
	// Running the new instance
	if err := Run(server, database); err != nil {
		log.Fatal("API exited!", "err", err)
	}
}

// What does the API do?
func Run(server *API, database *sql.DB) error {
	// Creates a new Chi router
	router := chi.NewRouter()
	// Rate limits everything reasonably to granularity rate limit subrouter's endpoints later
	router.Use(httprate.LimitByIP(50, time.Hour/2))
	// Creates a new Chi subrouter
	subrouter := chi.NewRouter()
	// Mounting the subrouter using API_VERSION
	router.Mount("/api/"+"v"+os.Getenv("API_VERSION"), subrouter)
	// Registering API routes
	Register(subrouter, database)
	// Running
	log.Infof("API running on %s", os.Getenv("API_ADDRESS"))
	return http.ListenAndServe(server.addr, router)
}
