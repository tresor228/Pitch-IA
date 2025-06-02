package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
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

type Pitch struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type AIPrompt struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float32 `json:"temperature"`
}

type AIResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// Variables globales
var (
	examplePitches = []string{
		"Problème: Les petits commerçants ont du mal à gérer leur inventaire\nSolution: Une app mobile de gestion d'inventaire simplifiée\nClient cible: Petits commerçants indépendants\nValeur: Gain de temps et réduction des erreurs\nCanaux: Boutique en ligne, réseaux sociaux",
		"Problème: Manque de solutions de livraison rapide en zone rurale\nSolution: Réseau de livreurs locaux à vélo\nClient cible: Commerces ruraux et habitants\nValeur: Livraison en moins de 2h à prix abordable\nCanaux: Partenariats avec commerces, site web",
	}

	// Système de stockage
	pitches   = make(map[string]Pitch)
	pitchesMu sync.Mutex
)

// Fonctions de stockage
func initStorage() {
	// Charger les données existantes si nécessaire
	data, err := os.ReadFile("pitches.json")
	if err == nil {
		json.Unmarshal(data, &pitches)
	}
}

func savePitch(pitch Pitch) error {
	pitchesMu.Lock()
	defer pitchesMu.Unlock()

	pitches[pitch.ID] = pitch

	data, err := json.MarshalIndent(pitches, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("pitches.json", data, 0644)
}

// Fonction principale de génération de pitch
func generatePitch(details ProjectDetails) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	prompt := fmt.Sprintf(`En utilisant la méthode Lean Canvas, génère un pitch clair et persuasif pour cette idée de projet. 

Idée principale: "%s"
Marché cible: "%s"
Concurrents principaux: "%s"
Aspect unique: "%s"
Modèle économique: "%s"

Structure le pitch avec ces sections:
1. Problème identifié (détaillé)
2. Solution proposée (précise)
3. Client cible (segmenté)
4. Proposition de valeur (convaincante)
5. Canaux de distribution
6. Avantage compétitif
7. Modèle économique

Le pitch doit être professionnel, concis et adapté à des investisseurs.`,
		details.Idea, details.TargetMarket, details.Competitors, details.UniqueAspect, details.BusinessModel)

	aiPrompt := AIPrompt{
		Model:       "text-davinci-003",
		Prompt:      prompt,
		MaxTokens:   700,
		Temperature: 0.7,
	}

	requestBody, err := json.Marshal(aiPrompt)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var aiResp AIResponse
	err = json.Unmarshal(body, &aiResp)
	if err != nil {
		return "", err
	}

	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return aiResp.Choices[0].Text, nil
}

// Handlers HTTP
func pitchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var details ProjectDetails
	err := json.NewDecoder(r.Body).Decode(&details)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pitch, err := generatePitch(details)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newPitch := Pitch{
		ID:        fmt.Sprintf("%d", time.Now().Unix()),
		Content:   pitch,
		CreatedAt: time.Now(),
	}
	err = savePitch(newPitch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"pitch": pitch,
		"id":    newPitch.ID,
	})
}

func examplesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(examplePitches)
}

func shareHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Pitch partagé avec succès",
	})
}

// Fonction principale
func main() {
	// Chargement des variables d'environnement
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialisation du stockage
	initStorage()

	// Configuration des routes
	http.HandleFunc("/generate-pitch", pitchHandler)
	http.HandleFunc("/examples", examplesHandler)
	http.HandleFunc("/share", shareHandler)
	http.Handle("/", http.FileServer(http.Dir("../frontend")))

	// Démarrage du serveur
	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
