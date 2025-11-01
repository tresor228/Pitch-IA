package controllers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"pitch/models"
	"pitch/service"
)

// Pitch affiche la page principale (GET /)
func Pitch(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("views/PItch.html")
	if err != nil {
		log.Printf("error parsing template: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
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
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	desc := r.FormValue("project_description")

	tmpl, err := template.ParseFiles("views/PItch.html")
	if err != nil {
		log.Printf("error parsing template: %v", err)
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
		tmpl.Execute(w, data)
		return
	}

	// Appel direct à l'API (sortie du mode demo)
	resp := service.GenerationwithAI(desc)
	if resp == nil {
		// Si la clé API n'est pas présente ou erreur côté service
		data.Error = "Impossible de générer le pitch : API non configurée ou erreur interne. Vérifiez la variable d'environnement OPENAI_API_KEY."
		// Si c'est une requête AJAX, retourner JSON
		accept := r.Header.Get("Accept")
		xreq := r.Header.Get("X-Requested-With")
		if strings.Contains(accept, "application/json") || xreq == "XMLHttpRequest" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": data.Error})
			return
		}

		tmpl.Execute(w, data)
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
