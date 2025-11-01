package routes

import (
	"log"
	"net/http"
	"pitch/controllers"
	"time"
)

// loggingMiddleware log les requêtes HTTP
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		log.Printf("%s %s %s %v", r.Method, r.RequestURI, r.RemoteAddr, time.Since(start))
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
