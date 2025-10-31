package main

import (
	"log"
	"net/http"
	"pitch/routes"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Fichier .env non trouvé, utilisation des variables système")
	}

	routes.Web()

	log.Println("Server démarré sur :9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}
