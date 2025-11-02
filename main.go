package main

import (
	"log"
	"net/http"
	"os"
	"pitch/routes"

	"github.com/joho/godotenv"
)

func main() {
	// Charger les variables d'environnement depuis .env (optionnel, pour le développement local)
	_ = godotenv.Load(".env")

	// Configurer les routes
	routes.Web()

	// Lire le port depuis la variable d'environnement PORT
	// Render/Vercel définissent automatiquement cette variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
		log.Println("Variable PORT non définie, utilisation du port par défaut: 8082")
	}

	log.Printf(" Serveur démarré sur le port %s", port)
	log.Printf(" Routes configurées: / et /analyze-pitch")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf(" Erreur lors du démarrage du serveur: %v", err)
	}
}
