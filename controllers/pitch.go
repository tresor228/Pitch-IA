package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pitch/models"
	"pitch/service"
)

// min retourne le minimum entre deux entiers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getTemplatePath retourne le chemin absolu vers le template
func getTemplatePath() string {
	// Chemins possibles (en production, les fichiers sont dans le répertoire de travail)
	paths := []string{
		"views/Pitch.html",
		"./views/Pitch.html",
		filepath.Join("views", "Pitch.html"),
		filepath.Join(".", "views", "Pitch.html"),
	}

	// Essayer chaque chemin
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			absPath, err := filepath.Abs(path)
			if err == nil {
				log.Printf("Template trouvé: %s", absPath)
				return absPath
			}
		}
	}

	// Fallback vers le chemin relatif (sera résolu au runtime)
	log.Printf(" Template non trouvé dans les chemins standards, utilisation du chemin relatif")
	return "views/Pitch.html"
}

// Pitch affiche la page principale (GET /)
func Pitch(w http.ResponseWriter, r *http.Request) {
	tmplPath := getTemplatePath()
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf(" Erreur parsing template from %s: %v", tmplPath, err)
		log.Printf(" Répertoire de travail: %s", func() string {
			wd, _ := os.Getwd()
			return wd
		}())
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}

	data := models.TemplateData{
		UserInput: "",
		Response:  nil,
		Loading:   false,
		Error:     "",
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Render error", http.StatusInternalServerError)
		return
	}
}

// AnalyzePitch traite le formulaire POST /analyze-pitch
func AnalyzePitch(w http.ResponseWriter, r *http.Request) {
	// Protection contre les panics
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC dans AnalyzePitch: %v", err)
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		}
	}()

	// Limiter la taille du body (max 10KB pour la description)
	r.Body = http.MaxBytesReader(w, r.Body, 10240)

	if err := r.ParseForm(); err != nil {
		log.Printf("Erreur ParseForm: %v", err)
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	desc := r.FormValue("project_description")

	tmplPath := getTemplatePath()
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("error parsing template from %s: %v", tmplPath, err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	data := models.TemplateData{
		UserInput: desc,
		Response:  nil,
		Loading:   false,
		Error:     "",
	}

	if desc == "" {
		data.Error = "Veuillez décrire votre projet."
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("error executing template: %v", err)
			http.Error(w, "Render error", http.StatusInternalServerError)
		}
		return
	}

	// Validation de la longueur
	if len(desc) < 10 {
		data.Error = "La description doit contenir au moins 10 caractères."
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("error executing template: %v", err)
			http.Error(w, "Render error", http.StatusInternalServerError)
		}
		return
	}

	if len(desc) > 2000 {
		data.Error = "La description ne doit pas dépasser 2000 caractères."
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("error executing template: %v", err)
			http.Error(w, "Render error", http.StatusInternalServerError)
		}
		return
	}

	// Appel direct à l'API (sortie du mode demo)
	log.Printf("Début génération AI pour: %s", desc[:min(50, len(desc))])
	resp := service.GenerationwithAI(desc)
	if resp == nil {
		log.Printf("Génération AI a échoué ou retourné nil")
		// Vérifier si la clé API est présente pour donner un message d'erreur plus précis
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			data.Error = "⚠️ La clé API OpenAI n'est pas configurée. Veuillez définir la variable d'environnement OPENAI_API_KEY."
			log.Printf("Erreur: OPENAI_API_KEY manquante")
		} else {
			data.Error = "⚠️ Impossible de générer le pitch. Cela peut être dû à un problème réseau, un timeout ou une erreur de l'API OpenAI. Veuillez réessayer dans quelques instants."
			log.Printf("Erreur: API OpenAI a échoué malgré la présence de la clé")
		}
		// Si c'est une requête AJAX, retourner JSON
		accept := r.Header.Get("Accept")
		xreq := r.Header.Get("X-Requested-With")
		if strings.Contains(accept, "application/json") || xreq == "XMLHttpRequest" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": data.Error})
			return
		}

		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("error executing template: %v", err)
			http.Error(w, "Render error", http.StatusInternalServerError)
		}
		return
	}

	data.Response = resp

	// Log for debugging parsing issues
	log.Printf("Parsed response: %+v", resp)

	// Si requête AJAX, renvoyer JSON
	accept := r.Header.Get("Accept")
	xreq := r.Header.Get("X-Requested-With")
	if strings.Contains(accept, "application/json") || xreq == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data.Response)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Render error", http.StatusInternalServerError)
		return
	}
}
