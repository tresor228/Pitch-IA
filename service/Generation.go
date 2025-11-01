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

	system := "Tu es un assistant qui doit retourner STRICTEMENT un pitch structuré en français avec 6 sections numérotées. Répond uniquement avec les sections demandées, chaque section commençant par son numéro et son label entre crochets exactement comme ci-dessous. Si tu n'as pas assez d'information, écris une brève suggestion pour cette section (1-2 phrases). Ne rajoute pas de texte hors des sections."

	prompt := fmt.Sprintf("Génère un pitch structuré pour la description suivante. Répond exactement au format numéroté ci-dessous (en français) :\n1. [Problème] ...\n2. [Solution] ...\n3. [Marché] ...\n4. [Valeur] ...\n5. [Canaux] ...\n6. [Modèle] ...\nNe rajoute pas de texte hors de ces sections. Description : %s", input)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: system,
				},
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
	parsed := parseAIResponse(content)

	// Si certaines sections restent vides, remplir avec une suggestion minimale basée sur l'entrée
	if parsed.Solution == "" {
		parsed.Solution = fmt.Sprintf("Suggestion de solution basée sur la description : %s", input)
	}
	if parsed.Valeur == "" {
		parsed.Valeur = fmt.Sprintf("Valeur principale : améliore l'expérience utilisateur et réduit les coûts liés à %s", input)
	}
	if parsed.Modele == "" {
		parsed.Modele = "Modèle proposé : freemium + abonnement premium ou commissions selon le service."
	}

	return parsed
}

// parseAIResponse extrait les sections françaises du texte retourné par l'IA
func parseAIResponse(content string) *models.PitchResponse {
	// mapping par défaut
	result := &models.PitchResponse{}
	// Première approche: parser robuste basé sur détection des en-têtes numérotés
	// On repère les marqueurs "1.", "2)" n'importe où dans le texte (inline ou lignes séparées),
	// puis on extrait le label qui suit (ex: "[Problème]" ou "Problème:") et on prend tout
	// le texte jusqu'à l'en-tête suivant.
	headerAnywhereRegex := regexp.MustCompile(`\d+\s*[\.\)]\s*`)
	headerPos := headerAnywhereRegex.FindAllStringIndex(content, -1)

	// mapping des synonymes vers clés
	synonyms := map[string][]string{
		"probleme": {"problème", "probleme"},
		"solution": {"solution"},
		"marche":   {"marché", "marche"},
		"valeur":   {"valeur", "proposition de valeur", "uvp", "valeur unique"},
		"canaux":   {"canaux"},
		"modele":   {"modèle", "modele", "modèle économique", "business model"},
	}

	// helper: detect canonical key from a label string
	detectKey := func(label string) string {
		l := strings.ToLower(strings.TrimSpace(label))
		// remove surrounding brackets if any
		l = strings.Trim(l, "[]")
		for k, list := range synonyms {
			for _, s := range list {
				if strings.Contains(l, s) {
					return k
				}
			}
		}
		return ""
	}

	// helper: split header line into label and inline rest
	headerLabelRe := regexp.MustCompile(`(?i)^\s*\[?\s*([^\]\:\-–—]+?)\s*\]?\s*[:\-–—]?\s*(.*)$`)

	if len(headerPos) > 0 {
		for i, pos := range headerPos {
			markerEnd := pos[1]
			var end int
			if i+1 < len(headerPos) {
				end = headerPos[i+1][0]
			} else {
				end = len(content)
			}

			// sectionText starts after the numeric marker (e.g. after "1.")
			sectionText := strings.TrimSpace(content[markerEnd:end])

			// headerLabelRe extracts label and inline remainder
			label := sectionText
			inline := ""
			if m := headerLabelRe.FindStringSubmatch(sectionText); len(m) >= 3 {
				label = m[1]
				inline = strings.TrimSpace(m[2])
			}

			key := detectKey(label)

			// body is the rest of the sectionText (inline already contains the start)
			body := strings.TrimSpace(inline)
			// if inline was empty, take whole sectionText
			if body == "" {
				body = strings.TrimSpace(sectionText)
			}

			switch key {
			case "probleme":
				result.Probleme = body
			case "solution":
				result.Solution = body
			case "marche":
				result.Marche = body
			case "valeur":
				result.Valeur = body
			case "canaux":
				result.Canaux = body
			case "modele":
				result.Modele = body
			default:
				// unknown header: ignore
			}
		}
	}

	// Si certaines sections sont encore vides, tenter une extraction simple par labels suivis de ':'
	if result.Probleme == "" || result.Solution == "" || result.Marche == "" || result.Valeur == "" || result.Canaux == "" || result.Modele == "" {
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
		// assign only to missing fields
		if v, ok := buf["Probleme"]; ok && result.Probleme == "" {
			result.Probleme = strings.Join(v, " \n")
		}
		if v, ok := buf["Solution"]; ok && result.Solution == "" {
			result.Solution = strings.Join(v, " \n")
		}
		if v, ok := buf["Marche"]; ok && result.Marche == "" {
			result.Marche = strings.Join(v, " \n")
		}
		if v, ok := buf["Valeur"]; ok && result.Valeur == "" {
			result.Valeur = strings.Join(v, " \n")
		}
		if v, ok := buf["Canaux"]; ok && result.Canaux == "" {
			result.Canaux = strings.Join(v, " \n")
		}
		if v, ok := buf["Modele"]; ok && result.Modele == "" {
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
