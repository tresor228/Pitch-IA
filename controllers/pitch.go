package controllers

import (
	"html/template"
	"log"
	"net/http"

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

	// Appel au service (mock pour l'instant)
	resp := service.GeneratePitchResponse(desc)
	data.Response = resp

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Render error", http.StatusInternalServerError)
		return
	}
}
