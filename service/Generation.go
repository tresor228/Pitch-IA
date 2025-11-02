package service

import (
	"context"
	"fmt"
	"os"
	"pitch/models"
	"regexp"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// GenerationwithAI appelle OpenAI et parse la réponse en PitchResponse avec retry.
func GenerationwithAI(input string) *models.PitchResponse {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil
	}

	client := openai.NewClient(apiKey)

	system := "Tu es un assistant spécialisé dans la création de pitchs structurés. Tu dois TOUJOURS répondre dans un format STRICT avec 6 sections numérotées en français. Chaque section doit être sur SA PROPRE LIGNE, commençant par le numéro suivi d'un point, puis le label entre crochets, puis le contenu. EXEMPLE DE FORMAT OBLIGATOIRE:\n\n1. [Problème] Texte du problème ici\n2. [Solution] Texte de la solution ici\n3. [Marché] Texte du marché ici\n4. [Valeur] Texte de la valeur ici\n5. [Canaux] Texte des canaux ici\n6. [Modèle] Texte du modèle ici\n\nIMPORTANT: Ne mets RIEN avant la première section. Ne mets RIEN après la dernière section. Une seule section par ligne. Utilise EXACTEMENT ce format avec les numéros, points, crochets et labels en français."

	prompt := fmt.Sprintf("Génère un pitch structuré pour ce projet en utilisant EXACTEMENT le format ci-dessous (une ligne par section) :\n\n1. [Problème] Décris le problème spécifique que ce projet résout\n2. [Solution] Décris la solution concrète que ce projet apporte\n3. [Marché] Décris le marché cible et l'opportunité\n4. [Valeur] Décris la proposition de valeur unique\n5. [Canaux] Décris les canaux de distribution/acquisition\n6. [Modèle] Décris le modèle économique\n\nDescription du projet : %s\n\nRéponds UNIQUEMENT avec les 6 lignes au format ci-dessus, sans texte avant ou après.", input)

	// Tentative avec retry (max 3 tentatives)
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			time.Sleep(time.Duration(attempt) * time.Second) // Délai progressif
		}

		// Timeout réduit à 25 secondes pour éviter les timeouts Render/Vercel (qui sont souvent à 30s)
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		
		modelName := "gpt-3.5-turbo"
		resp, err := client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model: modelName,
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
				Temperature: 0.7, // Température pour des réponses plus consistantes
				MaxTokens:   1000, // Limiter les tokens pour des réponses plus rapides
			},
		)
		
		if err != nil {
			errStr := err.Error()
			
			// Ne pas retry pour les erreurs d'authentification
			if strings.Contains(errStr, "401") || strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "invalid") || strings.Contains(errStr, "authentication") {
				return nil
			}
			
			// Continuer pour retry si ce n'est pas la dernière tentative
			if attempt < maxRetries {
				continue
			}
			return nil
		}

		if len(resp.Choices) == 0 {
			if attempt < maxRetries {
				continue
			}
			return nil
		}

		content := resp.Choices[0].Message.Content
		if content == "" {
			if attempt < maxRetries {
				continue
			}
			return nil
		}

		parsed := parseAIResponse(content)

		// Compter les sections remplies
		filledCount := 0
		if parsed.Probleme != "" {
			filledCount++
		}
		if parsed.Solution != "" {
			filledCount++
		}
		if parsed.Marche != "" {
			filledCount++
		}
		if parsed.Valeur != "" {
			filledCount++
		}
		if parsed.Canaux != "" {
			filledCount++
		}
		if parsed.Modele != "" {
			filledCount++
		}

		// Si toutes les sections sont vides, c'est un échec de parsing - retry
		if filledCount == 0 {
			if attempt < maxRetries {
				continue
			}
			return nil
		}

		// Si certaines sections restent vides, remplir avec une suggestion minimale basée sur l'entrée
		if parsed.Probleme == "" {
			parsed.Probleme = "Problème à définir basé sur votre description."
		}
		if parsed.Solution == "" {
			parsed.Solution = "Solution à développer selon votre projet."
		}
		if parsed.Marche == "" {
			parsed.Marche = "Marché cible à identifier."
		}
		if parsed.Valeur == "" {
			parsed.Valeur = "Proposition de valeur unique à définir."
		}
		if parsed.Canaux == "" {
			parsed.Canaux = "Canaux de distribution à mettre en place."
		}
		if parsed.Modele == "" {
			parsed.Modele = "Modèle économique : freemium + abonnement premium ou commissions selon le service."
		}

		return parsed
	}

	return nil
}

// parseAIResponse extrait les sections françaises du texte retourné par l'IA
func parseAIResponse(content string) *models.PitchResponse {
	result := &models.PitchResponse{}

	if content == "" {
		return result
	}

	// Diviser le contenu en lignes pour un meilleur contrôle
	lines := strings.Split(content, "\n")

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

	// Regex pour détecter les en-têtes numérotés: "1. [Label]" ou "1. Label" ou "1) Label"
	// Pattern amélioré pour extraire correctement le label entre crochets
	// Groupe 1: numéro, Groupe 2: label entre crochets (optionnel), Groupe 3: label sans crochets (si pas de crochets), Groupe 4: contenu
	headerRegex := regexp.MustCompile(`^\s*(\d+)\s*[\.\)]\s*(?:\[([^\]]+)\]|([^:\-–—\[]+?))\s*[:\-–—]?\s*(.*)$`)

	// Regex alternative plus simple pour détecter juste les numéros suivis de labels
	// On utilise [^\d] au lieu de [^0-9\n] car c'est plus simple pour RE2
	simpleHeaderRegex := regexp.MustCompile(`^\s*(\d+)\s*[\.\)]\s*([^\d]+)$`)

	var currentSection string
	var currentContent []string

	// Fonction pour sauvegarder la section courante
	saveCurrentSection := func(key string, content []string) {
		text := strings.TrimSpace(strings.Join(content, "\n"))
		if text == "" {
			return
		}

		switch key {
		case "probleme":
			if result.Probleme == "" {
				result.Probleme = text
			}
		case "solution":
			if result.Solution == "" {
				result.Solution = text
			}
		case "marche":
			if result.Marche == "" {
				result.Marche = text
			}
		case "valeur":
			if result.Valeur == "" {
				result.Valeur = text
			}
		case "canaux":
			if result.Canaux == "" {
				result.Canaux = text
			}
		case "modele":
			if result.Modele == "" {
				result.Modele = text
			}
		}
	}

	// Regex pour détecter le début d'une section (sans lookahead - compatible RE2)
	sectionStartRegex := regexp.MustCompile(`(\d+)\s*[\.\)]\s*(?:\[([^\]]+)\]|([^:\-–—\[]+?))\s*[:\-–—]?\s*`)

	// Parcourir toutes les lignes
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// D'abord, vérifier si la ligne contient plusieurs sections (format compact)
		// On trouve tous les marqueurs de début de section
		allStarts := sectionStartRegex.FindAllStringIndex(trimmed, -1)
		if len(allStarts) > 1 {
			// Cette ligne contient plusieurs sections, les traiter séparément
			for i, startPos := range allStarts {
				var endPos int
				if i+1 < len(allStarts) {
					endPos = allStarts[i+1][0]
				} else {
					endPos = len(trimmed)
				}

				sectionText := strings.TrimSpace(trimmed[startPos[0]:endPos])
				if match := sectionStartRegex.FindStringSubmatch(sectionText); match != nil && len(match) >= 4 {
					// Sauvegarder la section précédente
					if currentSection != "" && len(currentContent) > 0 {
						saveCurrentSection(currentSection, currentContent)
					}

					// match[1] = numéro, match[2] = label entre crochets, match[3] = label sans crochets
					var label string
					if match[2] != "" {
						label = strings.TrimSpace(match[2]) // Label entre crochets
					} else {
						label = strings.TrimSpace(match[3]) // Label sans crochets
					}
					// Extraire le contenu après le label
					labelEndIdx := sectionStartRegex.FindStringIndex(sectionText)
					var content string
					if labelEndIdx != nil && len(labelEndIdx) >= 2 {
						content = strings.TrimSpace(sectionText[labelEndIdx[1]:])
					} else {
						// Fallback: prendre tout après le label trouvé
						content = strings.TrimSpace(sectionText)
					}
					key := detectKey(label)

					if key != "" {
						currentSection = key
						currentContent = []string{}
						if content != "" {
							currentContent = append(currentContent, content)
						}
					}
				}
			}
		} else if match := headerRegex.FindStringSubmatch(trimmed); match != nil && len(match) >= 5 {
			// C'est un en-tête de section standard (une section par ligne)
			// Sauvegarder la section précédente si elle existe
			if currentSection != "" && len(currentContent) > 0 {
				saveCurrentSection(currentSection, currentContent)
			}

			// Nouvelle section détectée
			// match[1] = numéro, match[2] = label entre crochets, match[3] = label sans crochets, match[4] = contenu
			var label string
			if match[2] != "" {
				label = strings.TrimSpace(match[2]) // Label entre crochets [Problème]
			} else {
				label = strings.TrimSpace(match[3]) // Label sans crochets
			}
			inlineContent := strings.TrimSpace(match[4])

			key := detectKey(label)

			currentSection = key
			currentContent = []string{}

			// Si il y a du contenu inline après le label, l'ajouter
			if inlineContent != "" {
				currentContent = append(currentContent, inlineContent)
			}
		} else if currentSection != "" {
			// Vérifier si cette ligne commence une nouvelle section (cas où le format n'est pas standard)
			if sectionStartRegex.MatchString(trimmed) {
				// Sauvegarder la section précédente
				if len(currentContent) > 0 {
					saveCurrentSection(currentSection, currentContent)
				}

				// Extraire la nouvelle section
				if match := headerRegex.FindStringSubmatch(trimmed); match != nil && len(match) >= 5 {
					var label string
					if match[2] != "" {
						label = strings.TrimSpace(match[2]) // Label entre crochets
					} else {
						label = strings.TrimSpace(match[3]) // Label sans crochets
					}
					content := strings.TrimSpace(match[4])
					key := detectKey(label)
					currentSection = key
					currentContent = []string{}
					if content != "" {
						currentContent = append(currentContent, content)
					}
				}
			} else if trimmed != "" {
				// C'est une ligne de contenu normale de la section courante
				if currentSection != "" {
					currentContent = append(currentContent, trimmed)
				}
			}
		} else {
			// Aucune section active, essayer de trouver le début d'une section avec regex simple
			if simpleMatch := simpleHeaderRegex.FindStringSubmatch(trimmed); simpleMatch != nil && len(simpleMatch) >= 3 {
				restOfLine := strings.TrimSpace(simpleMatch[2])
				// Essayer d'extraire un label du début
				labelParts := strings.Fields(restOfLine)
				if len(labelParts) > 0 {
					possibleLabel := labelParts[0]
					key := detectKey(possibleLabel)
					if key != "" {
						currentSection = key
						currentContent = []string{}
						// Ajouter le reste de la ligne comme contenu
						if len(labelParts) > 1 {
							currentContent = append(currentContent, strings.Join(labelParts[1:], " "))
						}
					}
				}
			}
		}
	}

	// Sauvegarder la dernière section
	if currentSection != "" && len(currentContent) > 0 {
		saveCurrentSection(currentSection, currentContent)
	}

	// Si certaines sections sont encore vides, tenter une extraction simple par labels suivis de ':'
	// Cette partie ne s'exécute que si le parsing principal n'a pas fonctionné
	if result.Probleme == "" || result.Solution == "" || result.Marche == "" || result.Valeur == "" || result.Canaux == "" || result.Modele == "" {
		// essayer des labels français avec ':' (fallback)
		labels := []struct{ key, label string }{
			{"Probleme", "Problème:"},
			{"Solution", "Solution:"},
			{"Marche", "Marché:"},
			{"Valeur", "Valeur:"},
			{"Canaux", "Canaux:"},
			{"Modele", "Modèle:"},
		}
		fallbackLines := strings.Split(content, "\n")
		current := ""
		buf := map[string][]string{}
		for _, line := range fallbackLines {
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
			result.Probleme = strings.Join(v, "\n")
		}
		if v, ok := buf["Solution"]; ok && result.Solution == "" {
			result.Solution = strings.Join(v, "\n")
		}
		if v, ok := buf["Marche"]; ok && result.Marche == "" {
			result.Marche = strings.Join(v, "\n")
		}
		if v, ok := buf["Valeur"]; ok && result.Valeur == "" {
			result.Valeur = strings.Join(v, "\n")
		}
		if v, ok := buf["Canaux"]; ok && result.Canaux == "" {
			result.Canaux = strings.Join(v, "\n")
		}
		if v, ok := buf["Modele"]; ok && result.Modele == "" {
			result.Modele = strings.Join(v, "\n")
		}
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
