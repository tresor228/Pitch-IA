package routes

import (
	"net/http"
	"pitch/controllers"
)

func Web() {
	// Page d'accueil (GET)
	http.HandleFunc("/", controllers.Pitch)

	// Traitement du formulaire (POST)
	http.HandleFunc("/analyze-pitch", controllers.AnalyzePitch)

}
