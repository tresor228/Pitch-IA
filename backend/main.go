package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
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

// Fonction de génération avec OpenAI
func generateWithAI(details ProjectDetails) (string, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	prompt := fmt.Sprintf(`Génère un pitch business professionnel en français basé sur ces éléments:
	Idée: %s
	Marché cible: %s
	Aspect unique: %s
	Modèle économique: %s
	
	Le pitch doit être structuré avec des sections claires et un ton persuasif. Inclure:
	1. Problème identifié
	2. Solution proposée
	3. Marché cible
	4. Avantage compétitif
	5. Modèle économique
	6. Potentiel de croissance`,
		details.Idea, details.TargetMarket, details.UniqueAspect, details.BusinessModel)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("erreur OpenAI: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}

// Fonction de génération locale (fallback)
func generateLocalPitch(details ProjectDetails) (string, error) {
	pitch := fmt.Sprintf(`**PITCH BUSINESS - %s**

🎯 **PROBLÈME IDENTIFIÉ**
%s

💡 **SOLUTION PROPOSÉE**
%s

👥 **CLIENT CIBLE**
%s

⭐ **AVANTAGE COMPÉTITIF**
%s

💰 **MODÈLE ÉCONOMIQUE**
%s

🚀 **POTENTIEL**
Marché en croissance avec opportunité de différenciation.`,
		getValueOrDefault(details.Idea, "Idée innovante"),
		getValueOrDefault(details.Idea, "Problème non spécifié"),
		getValueOrDefault(details.Idea, "Solution innovante"),
		getValueOrDefault(details.TargetMarket, "Marché large"),
		getValueOrDefault(details.UniqueAspect, "Différenciation claire"),
		getValueOrDefault(details.BusinessModel, "Modèle à définir"))

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

	if strings.TrimSpace(details.Idea) == "" {
		http.Error(w, "L'idée principale est requise", http.StatusBadRequest)
		return
	}

	var pitch string
	if os.Getenv("OPENAI_API_KEY") != "" {
		pitch, err = generateWithAI(details)
		if err != nil {
			log.Printf("Erreur OpenAI, fallback local: %v", err)
			pitch, err = generateLocalPitch(details)
		}
	} else {
		pitch, err = generateLocalPitch(details)
	}

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
		log.Printf("Erreur sauvegarde: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"pitch": pitch,
		"id":    newPitch.ID,
	})
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

	if !strings.Contains(data.Email, "@") {
		http.Error(w, "Adresse email invalide", http.StatusBadRequest)
		return
	}

	log.Printf("Pitch partagé vers: %s", data.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Pitch partagé avec succès",
	})
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}
	http.ServeFile(w, r, "./frontend"+r.URL.Path)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Fichier .env non trouvé, utilisation des variables système")
	}

	err = initStorage()
	if err != nil {
		log.Printf("Erreur initialisation stockage: %v", err)
	}

	http.HandleFunc("/generate-pitch", pitchHandler)
	http.HandleFunc("/examples", examplesHandler)
	http.HandleFunc("/share", shareHandler)
	http.HandleFunc("/", staticHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Serveur démarré sur le port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
