package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Structures de données
type ProjectDetails struct {
	Idea          string `json:"idea"`
	TargetMarket  string `json:"targetMarket"`
	Competitors   string `json:"competitors"`
	UniqueAspect  string `json:"uniqueAspect"`
	BusinessModel string `json:"businessModel"`
}

// Variables globales
var examplePitches = []string{
	"Problème: Les petits commerçants ont du mal à gérer leur inventaire\nSolution: Une app mobile de gestion d'inventaire simplifiée\nClient cible: Petits commerçants indépendants\nValeur: Gain de temps et réduction des erreurs\nCanaux: Boutique en ligne, réseaux sociaux",
	"Problème: Manque de solutions de livraison rapide en zone rurale\nSolution: Réseau de livreurs locaux à vélo\nClient cible: Commerces ruraux et habitants\nValeur: Livraison en moins de 2h à prix abordable\nCanaux: Partenariats avec commerces, site web",
}

// Fonction de génération de pitch (version simplifiée sans API externe)
func generatePitch(details ProjectDetails) (string, error) {
	// Template de pitch basé sur la méthode Lean Canvas
	pitch := fmt.Sprintf(`**PITCH BUSINESS - %s**

🎯 **PROBLÈME IDENTIFIÉ**
Le marché cible (%s) fait face à des défis significatifs que notre solution peut résoudre de manière efficace.

💡 **SOLUTION PROPOSÉE**
%s

Notre approche unique: %s

👥 **CLIENT CIBLE**
Marché cible: %s
Segmentation claire avec des besoins spécifiques identifiés.

💰 **PROPOSITION DE VALEUR**
- Résolution directe du problème identifié
- Avantage compétitif face aux concurrents: %s
- Valeur ajoutée mesurable pour les clients

📈 **CANAUX DE DISTRIBUTION**
- Marketing digital ciblé
- Partenariats stratégiques
- Vente directe et en ligne

🏆 **AVANTAGE COMPÉTITIF**
%s

💵 **MODÈLE ÉCONOMIQUE**
%s

**Prêt à transformer cette vision en réalité !**`,
		details.Idea,
		getValueOrDefault(details.TargetMarket, "Marché cible à définir"),
		details.Idea,
		getValueOrDefault(details.UniqueAspect, "Innovation et approche différenciée"),
		getValueOrDefault(details.TargetMarket, "Segments de marché stratégiques"),
		getValueOrDefault(details.Competitors, "Concurrence traditionnelle"),
		getValueOrDefault(details.UniqueAspect, "Innovation et positionnement unique"),
		getValueOrDefault(details.BusinessModel, "Modèle de revenus à développer"))

	return pitch, nil
}

// Fonction utilitaire pour gérer les valeurs vides
func getValueOrDefault(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

// Middleware CORS
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Handlers HTTP
func pitchHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var details ProjectDetails
	err := json.NewDecoder(r.Body).Decode(&details)
	if err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validation basique
	if strings.TrimSpace(details.Idea) == "" {
		http.Error(w, "L'idée principale est requise", http.StatusBadRequest)
		return
	}

	pitch, err := generatePitch(details)
	if err != nil {
		http.Error(w, "Erreur lors de la génération: "+err.Error(), http.StatusInternalServerError)
		return
	}

	newPitch := Pitch{
		ID:        fmt.Sprintf("%d", time.Now().Unix()),
		Content:   pitch,
		CreatedAt: time.Now(),
	}

	err = savePitch(newPitch)
	if err != nil {
		log.Printf("Erreur lors de la sauvegarde: %v", err)
		// Continue même si la sauvegarde échoue
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"pitch": pitch,
		"id":    newPitch.ID,
	}
	json.NewEncoder(w).Encode(response)
}

func examplesHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(examplePitches)
}

func shareHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Pitch string `json:"pitch"`
		Email string `json:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validation de l'email basique
	if !strings.Contains(data.Email, "@") {
		http.Error(w, "Adresse email invalide", http.StatusBadRequest)
		return
	}

	// Simuler l'envoi d'email (remplacer par vraie logique d'envoi)
	log.Printf("Partage du pitch vers: %s", data.Email)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":  "success",
		"message": "Pitch partagé avec succès vers " + data.Email,
	}
	json.NewEncoder(w).Encode(response)
}

// Handler pour servir les fichiers statiques
func staticHandler(w http.ResponseWriter, r *http.Request) {
	// Rediriger la racine vers index.html
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}

	// Servir les fichiers depuis le dossier frontend
	http.ServeFile(w, r, "./frontend"+r.URL.Path)
}

// Fonction principale
func main() {
	// Chargement des variables d'environnement (optionnel)
	err := godotenv.Load()
	if err != nil {
		log.Println("Fichier .env non trouvé, utilisation des variables d'environnement système")
	}

	// Initialisation du stockage
	err = initStorage()
	if err != nil {
		log.Printf("Erreur lors de l'initialisation du stockage: %v", err)
	}

	// Configuration des routes API
	http.HandleFunc("/generate-pitch", pitchHandler)
	http.HandleFunc("/examples", examplesHandler)
	http.HandleFunc("/share", shareHandler)

	// Route pour les fichiers statiques
	http.HandleFunc("/", staticHandler)

	// Démarrage du serveur
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Printf("Accédez à l'application sur: http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
