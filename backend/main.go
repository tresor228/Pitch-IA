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

	prompt := fmt.Sprintf(`Tu es un expert en création de startups. Génère un pitch business complet et unique basé exclusivement sur l'idée suivante :

"%s"

Le pitch doit être entièrement original et contenir ces sections :

 [Problème] Décris le problème concret que cette idée résout (50 mots max)
 [Solution] Présente la solution spécifique proposée (50 mots max)
 [Marché] Identifie précisément le public cible (30 mots max)
 [Valeur] Détaille l'avantage unique par rapport aux alternatives (30 mots max)
 [Canaux] Explique comment atteindre les clients (30 mots max)
 [Modèle] Précise comment générer des revenus (30 mots max)

Important : 
- Sois concret et spécifique
- Évite les phrases génériques
- Adapte chaque section à l'idée fournie
- Utilise exclusivement les informations fournies par l'utilisateur`, details.Idea)

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
			Temperature: 0.8, // Augmenté pour plus de créativité
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

	// Regex améliorée pour capturer les sections
	re := regexp.MustCompile(`(?mi)^\d+\.\s*\[(.*?)\]\s*(.*?)(?:\n\d+\.|$)`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			sectionName := strings.TrimSpace(match[1])
			sectionContent := strings.TrimSpace(match[2])
			if section, ok := sections[sectionName]; ok {
				*section = sectionContent
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

	details.Idea = strings.TrimSpace(details.Idea)
	if details.Idea == "" || len(details.Idea) < 15 {
		http.Error(w, "Veuillez fournir une idée plus détaillée (min 15 caractères)", http.StatusBadRequest)
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
	// Chargement du .env
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Fichier .env non trouvé, utilisation des variables système")
	}

	// Configuration des routes
	http.HandleFunc("/api/generate-pitch", pitchHandler)
	http.HandleFunc("/api/examples", examplesHandler)

	// Serveur de fichiers statiques
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	// Configuration du port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Serveur démarré sur le port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
