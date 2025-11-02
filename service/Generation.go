package service

import (
	"context"
	"fmt"
	"log"
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
		log.Printf("OPENAI_API_KEY non définie")
		return nil
	}

	client := openai.NewClient(apiKey)
	log.Printf("Création du client OpenAI")

	system := "Tu es un assistant spécialisé dans la création de pitchs structurés. Tu dois TOUJOURS répondre dans un format STRICT avec 6 sections numérotées en français. Chaque section doit être sur SA PROPRE LIGNE, commençant par le numéro suivi d'un point, puis le label entre crochets, puis le contenu. EXEMPLE DE FORMAT OBLIGATOIRE:\n\n1. [Problème] Texte du problème ici\n2. [Solution] Texte de la solution ici\n3. [Marché] Texte du marché ici\n4. [Valeur] Texte de la valeur ici\n5. [Canaux] Texte des canaux ici\n6. [Modèle] Texte du modèle ici\n\nIMPORTANT: Ne mets RIEN avant la première section. Ne mets RIEN après la dernière section. Une seule section par ligne. Utilise EXACTEMENT ce format avec les numéros, points, crochets et labels en français."

	prompt := fmt.Sprintf("Génère un pitch structuré pour ce projet en utilisant EXACTEMENT le format ci-dessous (une ligne par section) :\n\n1. [Problème] Décris le problème spécifique que ce projet résout\n2. [Solution] Décris la solution concrète que ce projet apporte\n3. [Marché] Décris le marché cible et l'opportunité\n4. [Valeur] Décris la proposition de valeur unique\n5. [Canaux] Décris les canaux de distribution/acquisition\n6. [Modèle] Décris le modèle économique\n\nDescription du projet : %s\n\nRéponds UNIQUEMENT avec les 6 lignes au format ci-dessus, sans texte avant ou après.", input)

	// Tentative avec retry (max 3 tentatives)
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			log.Printf("Tentative %d/%d après échec précédent", attempt, maxRetries)
			time.Sleep(time.Duration(attempt) * time.Second) // Délai progressif
		}

		// Timeout réduit à 25 secondes pour éviter les timeouts Render/Vercel (qui sont souvent à 30s)
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		
		log.Printf("Appel à OpenAI (tentative %d/%d) avec timeout de 25s", attempt, maxRetries)
		resp, err := client.CreateChatCompletion(
			ctx,
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
				Temperature: 0.7, // Température pour des réponses plus consistantes
			},
		)
		
		if err != nil {
			lastErr = err
			log.Printf("ChatCompletion error (tentative %d): %v", attempt, err)
			// Log détaillé pour le débogage
			if err == context.DeadlineExceeded || ctx.Err() == context.DeadlineExceeded {
				log.Printf("⚠️ Timeout: La requête a pris plus de 25 secondes")
			}
			// Log supplémentaire pour les erreurs réseau
			if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "network") {
				log.Printf("⚠️ Erreur réseau lors de l'appel à OpenAI")
			}
			// Log pour les erreurs d'authentification (ne pas retry)
			if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "invalid") {
				log.Printf("⚠️ Erreur d'authentification: vérifiez que OPENAI_API_KEY est valide")
				return nil // Ne pas retry pour les erreurs d'auth
			}
			// Continuer pour retry si ce n'est pas la dernière tentative
			if attempt < maxRetries {
				continue
			}
			return nil
		}

		log.Printf("Réponse OpenAI reçue (tentative %d), %d choix disponibles", attempt, len(resp.Choices))

		if len(resp.Choices) == 0 {
			if attempt < maxRetries {
				log.Printf("Aucun choix dans la réponse, nouvelle tentative...")
				continue
			}
			log.Printf("❌ Aucun choix dans la réponse après %d tentatives", maxRetries)
			return nil
		}

		content := resp.Choices[0].Message.Content
		if content == "" {
			if attempt < maxRetries {
				log.Printf("Contenu vide, nouvelle tentative...")
				continue
			}
			log.Printf("❌ Contenu vide après %d tentatives", maxRetries)
			return nil
		}

		// Fonction min locale pour éviter la dépendance externe
		minLen := 500
		if len(content) < minLen {
			minLen = len(content)
		}
		log.Printf("✅ Contenu reçu d'OpenAI (tentative %d, premiers 500 caractères): %s", attempt, content[:minLen])
		log.Printf("Longueur totale du contenu: %d caractères", len(content))

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

		log.Printf("Résultat du parsing - Sections trouvées: %d/6 (Problème: %t, Solution: %t, Marché: %t, Valeur: %t, Canaux: %t, Modèle: %t)",
			filledCount, parsed.Probleme != "", parsed.Solution != "", parsed.Marche != "",
			parsed.Valeur != "", parsed.Canaux != "", parsed.Modele != "")

		// Si toutes les sections sont vides, c'est un échec de parsing - retry
		if filledCount == 0 {
			log.Printf("⚠️ Parsing échoué: toutes les sections sont vides. Contenu brut:\n%s", content)
			if attempt < maxRetries {
				log.Printf("Nouvelle tentative avec un prompt amélioré...")
				continue
			}
			log.Printf("❌ Parsing complètement échoué après %d tentatives", maxRetries)
			return nil
		}

		// Si au moins une section est remplie, on continue mais on remplit les manquantes
		// Si toutes les sections sont vides après parsing, on retourne nil (pas de fallback)
		if filledCount < 6 {
			log.Printf("⚠️ Seulement %d/6 sections trouvées, remplissage des sections manquantes", filledCount)
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

	// Si on arrive ici, toutes les tentatives ont échoué
	log.Printf("❌ Toutes les tentatives ont échoué. Dernière erreur: %v", lastErr)
	return nil
}

// parseAIResponse extrait les sections françaises du texte retourné par l'IA
func parseAIResponse(content string) *models.PitchResponse {
	result := &models.PitchResponse{}

	if content == "" {
		log.Printf("⚠️  Contenu vide reçu pour parsing")
		return result
	}

	// Diviser le contenu en lignes pour un meilleur contrôle
	lines := strings.Split(content, "\n")
	log.Printf("Nombre de lignes à parser: %d", len(lines))

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
	// Pattern plus permissif pour capturer différents formats
	headerRegex := regexp.MustCompile(`^\s*(\d+)\s*[\.\)]\s*\[?\s*([^\]\:\-–—]+?)\s*\]?\s*[:\-–—]?\s*(.*)$`)

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
	sectionStartRegex := regexp.MustCompile(`(\d+)\s*[\.\)]\s*\[?\s*([^\]\:\-–—]+?)\s*\]?\s*[:\-–—]?\s*`)

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
				if match := sectionStartRegex.FindStringSubmatch(sectionText); match != nil && len(match) >= 3 {
					// Sauvegarder la section précédente
					if currentSection != "" && len(currentContent) > 0 {
						saveCurrentSection(currentSection, currentContent)
					}

					label := strings.TrimSpace(match[2])
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
		} else if match := headerRegex.FindStringSubmatch(trimmed); match != nil {
			// C'est un en-tête de section standard (une section par ligne)
			// Sauvegarder la section précédente si elle existe
			if currentSection != "" && len(currentContent) > 0 {
				saveCurrentSection(currentSection, currentContent)
			}

			// Nouvelle section détectée
			label := strings.TrimSpace(match[2])
			inlineContent := strings.TrimSpace(match[3])

			key := detectKey(label)
			inlinePreview := inlineContent
			if len(inlineContent) > 50 {
				inlinePreview = inlineContent[:50] + "..."
			}
			log.Printf("Section détectée: label='%s' -> key='%s', contenu inline='%s'", label, key, inlinePreview)

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
				if match := headerRegex.FindStringSubmatch(trimmed); match != nil {
					label := strings.TrimSpace(match[2])
					content := strings.TrimSpace(match[3])
					key := detectKey(label)
					log.Printf("Nouvelle section détectée dans ligne: label='%s' -> key='%s'", label, key)
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
						log.Printf("Section trouvée avec regex simple: '%s' -> key='%s'", possibleLabel, key)
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

	// Log du résultat final
	log.Printf("Parsing terminé - Sections trouvées: Probleme=%d chars, Solution=%d chars, Marche=%d chars, Valeur=%d chars, Canaux=%d chars, Modele=%d chars",
		len(result.Probleme), len(result.Solution), len(result.Marche),
		len(result.Valeur), len(result.Canaux), len(result.Modele))

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
