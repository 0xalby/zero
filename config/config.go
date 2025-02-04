package config

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/go-chi/jwtauth/v5"
	"github.com/muesli/termenv"
)

var TokenAuth *jwtauth.JWTAuth

// Initializes a new JWT using the JWT_SECRET
func InitJWT(secret string) {
	TokenAuth = jwtauth.New("HS256", []byte(os.Getenv("JWT_SECRET")), nil)
	if TokenAuth == nil {
		log.Fatal("failed to initialize jwt")
		return
	}
}

var Logger *log.Logger

// Initializes a custom logger that can be used as a watchdog because it won't be as verbose as the normal logs
func InitLogger() {
	// Creating a style
	styles := log.DefaultStyles()
	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().SetString("NEW").Padding(0, 1, 0, 1).
		Background(lipgloss.Color("42")).
		Foreground(lipgloss.Color("15"))
	// Opening the file
	file, err := os.OpenFile("bin/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("failed to open the log file", "err", err)
	}
	// Creating the logger
	Logger = log.NewWithOptions(file, log.Options{
		Level: log.InfoLevel,
	})
	Logger.SetStyles(styles)
	Logger.SetColorProfile(termenv.TrueColor) // Handles terminal colors
}

// Gets the logger instance
func GetLogger() *log.Logger {
	return Logger
}
