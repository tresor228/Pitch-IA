package models

type ProjectDetails struct {
	Idee          string
	Marche_cible  string
	UniqueAspect  string
	BusinessModel string
}

type GeneratedPitch struct {
	Problem          string
	Solution         string
	Marche_cible     string
	ValueProposition string
	Channels         string
	BusinessModel    string
	FullPitch        string
}
