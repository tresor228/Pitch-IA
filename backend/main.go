package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

type ProjectDetails struct {
	Idea          string
	TargetMarket  string
	UniqueAspect  string
	BusinessModel string
}

type GeneratedPitch struct {
	Problem          string
	Solution         string
	TargetMarket     string
	ValueProposition string
	Channels         string
	BusinessModel    string
	FullPitch        string
}

var (
	examplePitches = []string{
		"Problème: Les petits commerçants ont du mal à gérer leur inventaire\nSolution: Une app mobile de gestion d'inventaire simplifiée\nClient cible: Petits commerçants indépendants\nValeur: Gain de temps et réduction des erreurs\nCanaux: Boutique en ligne, réseaux sociaux",
		"Problème: Manque de solutions de livraison rapide en zone rurale\nSolution: Réseau de livreurs locaux à vélo\nClient cible: Commerces ruraux et habitants\nValeur: Livraison en moins de 2h à prix abordable\nCanaux: Partenariats avec commerces, site web",
	}

	pitchCache = make(map[string]GeneratedPitch)
	cacheMutex sync.RWMutex
)

func generateWithAI(details ProjectDetails) (GeneratedPitch, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	prompt := fmt.Sprintf(`En tant qu'expert en création de startups, génère un pitch business complet et unique pour l'idée suivante :

Idée : %s

Le pitch doit contenir ces sections claires et distinctes :

1. [Problème] Décris le problème spécifique que cette idée résout (50 mots max)
2. [Solution] Explique en quoi cette solution est innovante (50 mots max)
3. [Marché] Détaille le public cible précis (âge, profession, besoins) (30 mots max)
4. [Valeur] Quel est l'avantage compétitif unique ? (30 mots max)
5. [Canaux] Comment les clients seront-ils atteints ? (30 mots max)
6. [Modèle] Comment l'argent sera-t-il gagné ? (abonnement, publicité, etc.) (30 mots max)

Sois concret, spécifique et évite les généralités.`, details.Idea)

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
			Temperature: 0.7,
		},
	)

	if err != nil {
		return GeneratedPitch{}, fmt.Errorf("erreur OpenAI: %v", err)
	}

	pitch := parseAIPitchResponse(resp.Choices[0].Message.Content)
	cachePitch(details.Idea, pitch)
	return pitch, nil
}

func parseAIPitchResponse(content string) GeneratedPitch {
	pitch := GeneratedPitch{FullPitch: content}

	sections := map[string]*string{
		"Problème": &pitch.Problem,
		"Solution": &pitch.Solution,
		"Marché":   &pitch.TargetMarket,
		"Valeur":   &pitch.ValueProposition,
		"Canaux":   &pitch.Channels,
		"Modèle":   &pitch.BusinessModel,
	}

	re := regexp.MustCompile(`(?m)^\d+\.\s*\[(.*?)\]\s*(.*?)(?:\n\d+\.|$)`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			if section, ok := sections[match[1]]; ok {
				*section = strings.TrimSpace(match[2])
			}
		}
	}

	return pitch
}

func getCachedPitch(idea string) (GeneratedPitch, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	pitch, exists := pitchCache[idea]
	return pitch, exists
}

func cachePitch(idea string, pitch GeneratedPitch) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	pitchCache[idea] = pitch
}

func enableCORS(w *http.ResponseWriter, r *http.Request) bool {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
		(*w).WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func pitchHandler(w http.ResponseWriter, r *http.Request) {
	if enableCORS(&w, r) {
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

	if strings.TrimSpace(details.Idea) == "" || len(details.Idea) < 10 {
		http.Error(w, "Veuillez fournir une idée plus détaillée (min 10 caractères)", http.StatusBadRequest)
		return
	}

	if cachedPitch, exists := getCachedPitch(details.Idea); exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"pitch": cachedPitch.FullPitch,
		})
		return
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		http.Error(w, "Service OpenAI non configuré", http.StatusServiceUnavailable)
		return
	}

	aiPitch, err := generateWithAI(details)
	if err != nil {
		log.Printf("Erreur OpenAI: %v", err)
		http.Error(w, "Erreur lors de la génération du pitch: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"pitch": aiPitch.FullPitch,
	})
}

func examplesHandler(w http.ResponseWriter, r *http.Request) {
	if enableCORS(&w, r) {
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(examplePitches)
}

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Fichier .env non trouvé, utilisation des variables système")
	}

	http.HandleFunc("/generate-pitch", pitchHandler)
	http.HandleFunc("/examples", examplesHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "../frontend/index.htm")
		} else {
			http.ServeFile(w, r, "../frontend"+r.URL.Path)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Serveur démarré sur le port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
