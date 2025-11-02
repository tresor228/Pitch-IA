package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
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
				return absPath
			}
		}
	}

	return "views/Pitch.html"
}

// Pitch affiche la page principale (GET /)
func Pitch(w http.ResponseWriter, r *http.Request) {
	tmplPath := getTemplatePath()
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
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
		http.Error(w, "Render error", http.StatusInternalServerError)
		return
	}
}

// AnalyzePitch traite le formulaire POST /analyze-pitch
func AnalyzePitch(w http.ResponseWriter, r *http.Request) {
	// Protection contre les panics
	defer func() {
		if err := recover(); err != nil {
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		}
	}()

	// Limiter la taille du body (max 10KB pour la description)
	r.Body = http.MaxBytesReader(w, r.Body, 10240)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	desc := r.FormValue("project_description")

	tmplPath := getTemplatePath()
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
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
			http.Error(w, "Render error", http.StatusInternalServerError)
		}
		return
	}

	// Validation de la longueur
	if len(desc) < 10 {
		data.Error = "La description doit contenir au moins 10 caractères."
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Render error", http.StatusInternalServerError)
		}
		return
	}

	if len(desc) > 2000 {
		data.Error = "La description ne doit pas dépasser 2000 caractères."
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Render error", http.StatusInternalServerError)
		}
		return
	}

	resp := service.GenerationwithAI(desc)
	if resp == nil {
		// Vérifier si la clé API est présente pour donner un message d'erreur plus précis
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			data.Error = "⚠️ La clé API OpenAI n'est pas configurée. Veuillez définir la variable d'environnement OPENAI_API_KEY dans les paramètres de votre service."
		} else {
			// Vérifier le format de la clé (doit commencer par sk-)
			if !strings.HasPrefix(apiKey, "sk-") {
				data.Error = "⚠️ Format de clé API invalide. La clé OpenAI doit commencer par 'sk-'. Vérifiez votre configuration."
			} else {
				data.Error = "⚠️ Impossible de générer le pitch après plusieurs tentatives.\n\nCauses possibles :\n• Problème réseau temporaire\n• Timeout de l'API OpenAI (>25s)\n• Quota/rate limit atteint\n• Service OpenAI temporairement indisponible\n\nVeuillez réessayer dans quelques instants."
			}
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
			http.Error(w, "Render error", http.StatusInternalServerError)
		}
		return
	}

	data.Response = resp

	// Si requête AJAX, renvoyer JSON
	accept := r.Header.Get("Accept")
	xreq := r.Header.Get("X-Requested-With")
	if strings.Contains(accept, "application/json") || xreq == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data.Response)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Render error", http.StatusInternalServerError)
		return
	}
}
