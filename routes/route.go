package routes

import (
	"log"
	"net/http"
	"pitch/controllers"
	"time"
)

// loggingMiddleware log les requêtes HTTP avec protection contre les panics
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC dans middleware: %v", err)
				http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
			}
			log.Printf("%s %s %s %v", r.Method, r.RequestURI, r.RemoteAddr, time.Since(start))
		}()
		next(w, r)
	}
}

// HealthCheck endpoint pour vérifier que l'application fonctionne
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"pitch-ia"}`))
}

// Web configure toutes les routes de l'application
func Web() {
	// Health check (pour Render/Vercel)
	http.HandleFunc("/health", HealthCheck)

	// Page d'accueil (GET)
	http.HandleFunc("/", loggingMiddleware(controllers.Pitch))

	// Traitement du formulaire (POST)
	http.HandleFunc("/analyze-pitch", loggingMiddleware(controllers.AnalyzePitch))
}
