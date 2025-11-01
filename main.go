package main

import (
	"log"
	"net/http"
	"os"
	"pitch/routes"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Fichier .env non trouvé, utilisation des variables système")
	}

	routes.Web()

	// Lire le port depuis la variable d'environnement PORT, fallback sur 8082
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Server démarré sur :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
