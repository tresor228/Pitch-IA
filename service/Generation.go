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

// GenerationwithAI appelle OpenAI et parse la réponse en PitchResponse.
func GenerationwithAI(input string) *models.PitchResponse {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Printf("OPENAI_API_KEY non définie")
		return nil
	}

	client := openai.NewClient(apiKey)
	log.Printf("Création du client OpenAI")

	system := "Tu es un assistant qui doit retourner STRICTEMENT un pitch structuré en français avec 6 sections numérotées. Chaque section doit être sur une ligne séparée avec son numéro, suivi du label entre crochets, puis le contenu. Format exact:\n1. [Problème] Texte du problème\n2. [Solution] Texte de la solution\n3. [Marché] Texte du marché\n4. [Valeur] Texte de la valeur\n5. [Canaux] Texte des canaux\n6. [Modèle] Texte du modèle\nNe mélange pas les sections. Chaque section sur sa propre ligne numérotée."

	prompt := fmt.Sprintf("Génère un pitch structuré pour la description suivante. Répond UNIQUEMENT au format ci-dessous, une section par ligne :\n\n1. [Problème] Décris le problème que résout ce projet\n2. [Solution] Décris la solution proposée\n3. [Marché] Décris le marché cible\n4. [Valeur] Décris la proposition de valeur unique\n5. [Canaux] Décris les canaux de distribution\n6. [Modèle] Décris le modèle économique\n\nDescription du projet : %s", input)

	// Timeout réduit à 45 secondes pour éviter les timeouts Render (qui sont souvent à 30s)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	log.Printf("Appel à OpenAI avec timeout de 45s")
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
		},
	)
	if err != nil {
		log.Printf("ChatCompletion error: %v", err)
		// Ne pas retourner nil directement, retourner une réponse vide avec erreur
		return nil
	}

	log.Printf("Réponse OpenAI reçue, %d choix disponibles", len(resp.Choices))

	if len(resp.Choices) == 0 {
		return nil
	}

	content := resp.Choices[0].Message.Content
	parsed := parseAIResponse(content)

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

// parseAIResponse extrait les sections françaises du texte retourné par l'IA
func parseAIResponse(content string) *models.PitchResponse {
	result := &models.PitchResponse{}
	
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
	headerRegex := regexp.MustCompile(`^\s*(\d+)\s*[\.\)]\s*\[?\s*([^\]\:\-–—]+?)\s*\]?\s*[:\-–—]?\s*(.*)$`)
	
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
	
	// Regex pour détecter toutes les sections dans une ligne (format compact)
	multiSectionRegex := regexp.MustCompile(`(\d+)\s*[\.\)]\s*\[?\s*([^\]\:\-–—]+?)\s*\]?\s*[:\-–—]?\s*([^\d]*?)(?=\d+\s*[\.\)]\s*\[|$)`)
	
	// Parcourir toutes les lignes
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		
		// D'abord, vérifier si la ligne contient plusieurs sections (format compact)
		if matches := multiSectionRegex.FindAllStringSubmatch(trimmed, -1); len(matches) > 1 {
			// Cette ligne contient plusieurs sections, les traiter séparément
			for _, m := range matches {
				if len(m) >= 4 {
					// Sauvegarder la section précédente
					if currentSection != "" && len(currentContent) > 0 {
						saveCurrentSection(currentSection, currentContent)
					}
					
					label := strings.TrimSpace(m[2])
					content := strings.TrimSpace(m[3])
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
			
			currentSection = detectKey(label)
			currentContent = []string{}
			
			// Si il y a du contenu inline après le label, l'ajouter
			if inlineContent != "" {
				currentContent = append(currentContent, inlineContent)
			}
		} else if currentSection != "" {
			// Vérifier si cette ligne commence une nouvelle section (cas où le format n'est pas standard)
			if newMatch := multiSectionRegex.FindStringSubmatch(trimmed); newMatch != nil && len(newMatch) >= 4 {
				// Sauvegarder la section précédente
				if len(currentContent) > 0 {
					saveCurrentSection(currentSection, currentContent)
				}
				
				label := strings.TrimSpace(newMatch[2])
				content := strings.TrimSpace(newMatch[3])
				currentSection = detectKey(label)
				currentContent = []string{}
				if content != "" {
					currentContent = append(currentContent, content)
				}
			} else if trimmed != "" {
				// C'est une ligne de contenu normale de la section courante
				currentContent = append(currentContent, trimmed)
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
