package models

// Struct pour la réponse de l'API
type PitchResponse struct {
	Probleme string
	Solution string
	Marche   string
	Valeur   string
	Canaux   string
	Modele   string
}

// Struct pour le template
type TemplateData struct {
	UserInput string
	Response  *PitchResponse
	Loading   bool
	Error     string
}
