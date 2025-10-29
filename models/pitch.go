package models

type ProjectDetails struct {
	Idea          string
	TargetMarket  string
	UniqueAspect  string
	BusinessModel string
}

type GeneratedPitch struct {
	Problem          string
	Solution         string
	TargetMarket     string
	ValueProposition string
	Channels         string
	BusinessModel    string
	FullPitch        string
}
