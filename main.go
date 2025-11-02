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
		port = "8088"
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf(" Erreur lors du démarrage du serveur: %v", err)
	}
}
