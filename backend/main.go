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

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

type ProjectDetails struct {
	Idea          string `json:"idea"`
	TargetMarket  string `json:"targetMarket"`
	UniqueAspect  string `json:"uniqueAspect"`
	BusinessModel string `json:"businessModel"`
}

type GeneratedPitch struct {
	Problem          string `json:"problem"`
	Solution         string `json:"solution"`
	TargetMarket     string `json:"targetMarket"`
	ValueProposition string `json:"valueProposition"`
	Channels         string `json:"channels"`
	BusinessModel    string `json:"businessModel"`
	FullPitch        string `json:"fullPitch"`
}

var examplePitches = []string{
	"Problème: Les petits commerçants ont du mal à gérer leur inventaire\nSolution: Une app mobile de gestion d'inventaire simplifiée\nClient cible: Petits commerçants indépendants\nValeur: Gain de temps et réduction des erreurs\nCanaux: Boutique en ligne, réseaux sociaux",
	"Problème: Manque de solutions de livraison rapide en zone rurale\nSolution: Réseau de livreurs locaux à vélo\nClient cible: Commerces ruraux et habitants\nValeur: Livraison en moins de 2h à prix abordable\nCanaux: Partenariats avec commerces, site web",
}

func generateWithAI(details ProjectDetails) (GeneratedPitch, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	prompt := fmt.Sprintf(`Tu es un expert en création de pitch business. Crée un pitch UNIQUE basé sur cette idée: "%s"

Structure obligatoire:
1. [Problème] (50 mots max) - Décris le problème concret
2. [Solution] (50 mots max) - Solution spécifique proposée
3. [Marché] (30 mots max) - Détaille le public cible
4. [Valeur] (30 mots max) - Avantage unique précis
5. [Canaux] (30 mots max) - Méthodes de distribution concrètes
6. [Modèle] (30 mots max) - Modèle économique spécifique

Évite les généralités. Sois précis et créatif.`, details.Idea)

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
			Temperature: 0.9,
		},
	)

	if err != nil {
		return GeneratedPitch{}, fmt.Errorf("erreur OpenAI: %v", err)
	}

	return parseAIPitchResponse(resp.Choices[0].Message.Content), nil
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

	re := regexp.MustCompile(`(\d+\.\s*\[(.*?)\]\s*)(.*?)(?=\n\d+\.|$)`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if section, ok := sections[match[2]]; ok {
			*section = strings.TrimSpace(match[3])
		}
	}

	return pitch
}

func generateLocalPitch(details ProjectDetails) (string, error) {
	idea := getValueOrDefault(details.Idea, "votre idée")

	return fmt.Sprintf(`Pitch pour: %s

Problème: Les utilisateurs ont besoin de solutions pour "%s"
Solution: Approche innovante combinant technologie et méthodologie
Marché: Public intéressé par %s
Valeur: Solution %s unique et personnalisable
Canaux: Plateforme en ligne avec marketing digital
Modèle: Freemium avec options payantes`,
		idea, idea, idea, idea), nil
}

func getValueOrDefault(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func isGenericPitch(content string) bool {
	genericTerms := []string{"solution innovante", "marché large", "modèle à définir"}
	for _, term := range genericTerms {
		if strings.Contains(content, term) {
			return true
		}
	}
	return false
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

	var pitchContent string
	var pitchErr error

	if os.Getenv("OPENAI_API_KEY") != "" {
		aiPitch, aiErr := generateWithAI(details)
		if aiErr == nil && !isGenericPitch(aiPitch.FullPitch) {
			pitchContent = aiPitch.FullPitch
		} else {
			log.Printf("Fallback local: %v", aiErr)
			pitchContent, pitchErr = generateLocalPitch(details)
		}
	} else {
		pitchContent, pitchErr = generateLocalPitch(details)
	}

	if pitchErr != nil {
		http.Error(w, "Erreur lors de la génération: "+pitchErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"pitch": pitchContent,
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
