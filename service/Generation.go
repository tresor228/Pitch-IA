package service

import (
	"context"
	"fmt"
	"os"
	"pitch/models"
	"regexp"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// GenerationwithAI appelle OpenAI et parse la réponse en PitchResponse.
func GenerationwithAI(input string) *models.PitchResponse {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil
	}

	client := openai.NewClient(apiKey)

	prompt := fmt.Sprintf("Génère un pitch structuré pour la description suivante. Répond exactement au format numéroté ci-dessous (en français) :\n1. [Problème] ...\n2. [Solution] ...\n3. [Marché] ...\n4. [Valeur] ...\n5. [Canaux] ...\n6. [Modèle] ...\nNe rajoute pas de texte hors de ces sections. Description : %s", input)

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
		fmt.Printf("ChatCompletion error: %v\n", err)
		return nil
	}

	if len(resp.Choices) == 0 {
		return nil
	}

	content := resp.Choices[0].Message.Content
	return parseAIResponse(content)
}

// parseAIResponse extrait les sections françaises du texte retourné par l'IA
func parseAIResponse(content string) *models.PitchResponse {
	// mapping par défaut
	result := &models.PitchResponse{}

	// Regex pour capturer les sections numérotées avec ou sans crochets
	// Ex: 1. [Problème] contenu...
	re := regexp.MustCompile(`(?mi)^\s*\d+\.\s*\[?\s*(Probl[eè]me|Solution|March[eé]|Valeur|Canaux|Mod[eè]le)\s*\]?\s*(.*?)(?=\n\s*\d+\.|\z)`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, m := range matches {
		if len(m) >= 3 {
			name := strings.TrimSpace(m[1])
			body := strings.TrimSpace(m[2])
			switch strings.ToLower(name) {
			case "problème", "probleme":
				result.Probleme = body
			case "solution":
				result.Solution = body
			case "marché", "marche":
				result.Marche = body
			case "valeur":
				result.Valeur = body
			case "canaux":
				result.Canaux = body
			case "modèle", "modele":
				result.Modele = body
			}
		}
	}

	// Si regex n'a rien trouvé, tenter une extraction simple par labels suivis de ':'
	if result.Probleme == "" && result.Solution == "" && result.Marche == "" && result.Valeur == "" && result.Canaux == "" && result.Modele == "" {
		// essayer des labels français avec ':'
		labels := []struct{ key, label string }{
			{"Probleme", "Problème:"},
			{"Solution", "Solution:"},
			{"Marche", "Marché:"},
			{"Valeur", "Valeur:"},
			{"Canaux", "Canaux:"},
			{"Modele", "Modèle:"},
		}
		lines := strings.Split(content, "\n")
		current := ""
		buf := map[string][]string{}
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			found := false
			for _, l := range labels {
				if strings.HasPrefix(trimmed, l.label) {
					current = l.key
					rest := strings.TrimSpace(strings.TrimPrefix(trimmed, l.label))
					if rest != "" {
						buf[current] = append(buf[current], rest)
					}
					found = true
					break
				}
			}
			if !found && current != "" {
				buf[current] = append(buf[current], trimmed)
			}
		}
		// assign
		if v, ok := buf["Probleme"]; ok {
			result.Probleme = strings.Join(v, " \n")
		}
		if v, ok := buf["Solution"]; ok {
			result.Solution = strings.Join(v, " \n")
		}
		if v, ok := buf["Marche"]; ok {
			result.Marche = strings.Join(v, " \n")
		}
		if v, ok := buf["Valeur"]; ok {
			result.Valeur = strings.Join(v, " \n")
		}
		if v, ok := buf["Canaux"]; ok {
			result.Canaux = strings.Join(v, " \n")
		}
		if v, ok := buf["Modele"]; ok {
			result.Modele = strings.Join(v, " \n")
		}
	}

	// fallback: si tout vide, mettre le contenu entier dans Probleme
	if result.Probleme == "" && result.Solution == "" && result.Marche == "" && result.Valeur == "" && result.Canaux == "" && result.Modele == "" {
		result.Probleme = strings.TrimSpace(content)
	}

	return result
}

// GeneratePitchResponse retourne un PitchResponse mocké basé sur l'entrée.
// Cette fonction est utilisée par le handler en mode démo (sans OpenAI).
func GeneratePitchResponse(input string) *models.PitchResponse {
	probleme := fmt.Sprintf("Les utilisateurs rencontrent %s, ce qui crée une friction dans le parcours.", input)
	solution := fmt.Sprintf("Nous proposons une solution simple et intuitive basée sur %s, améliorant la conversion.", input)
	marche := "Étudiants urbains 18-30 ans, utilisateurs mobiles cherchant commodité."
	valeur := "Gain de temps, personnalisation et prix attractif."
	canaux := "Réseaux sociaux, partenariats campus, campagnes locales."
	modele := "Freemium + abonnement premium + commissions."

	return &models.PitchResponse{
		Probleme: probleme,
		Solution: solution,
		Marche:   marche,
		Valeur:   valeur,
		Canaux:   canaux,
		Modele:   modele,
	}
}
