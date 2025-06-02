package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Pitch représente un pitch business généré
type Pitch struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

var (
	// pitches stocke en mémoire les pitches sauvegardés
	pitches = make(map[string]Pitch)

	// pitchesMu protège l'accès concurrent à la map pitches
	pitchesMu sync.RWMutex

	// chemin du fichier de stockage
	storageFile = "pitches.json"
)

// initStorage charge les pitches depuis le fichier de stockage
func initStorage() error {
	pitchesMu.Lock()
	defer pitchesMu.Unlock()

	// Vérifie si le fichier existe
	if _, err := os.Stat(storageFile); os.IsNotExist(err) {
		// Crée un fichier vide si inexistant
		return os.WriteFile(storageFile, []byte("{}"), 0644)
	}

	// Lit le fichier
	data, err := os.ReadFile(storageFile)
	if err != nil {
		return err
	}

	// Désérialise les données
	return json.Unmarshal(data, &pitches)
}

// savePitch sauvegarde un nouveau pitch
func savePitch(pitch Pitch) error {
	pitchesMu.Lock()
	defer pitchesMu.Unlock()

	// Génère un ID unique si non fourni
	if pitch.ID == "" {
		pitch.ID = generateID()
	}

	// Met à jour la date de création
	if pitch.CreatedAt.IsZero() {
		pitch.CreatedAt = time.Now()
	}

	// Ajoute le pitch à la map
	pitches[pitch.ID] = pitch

	// Sauvegarde dans le fichier
	return saveToFile()
}

// saveToFile sauvegarde la map pitches dans le fichier
func saveToFile() error {
	data, err := json.MarshalIndent(pitches, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(storageFile, data, 0644)
}

// generateID génère un ID unique basé sur le timestamp
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getPitch récupère un pitch par son ID (version thread-safe)
func getPitch(id string) (Pitch, bool) {
	pitchesMu.RLock()
	defer pitchesMu.RUnlock()

	pitch, exists := pitches[id]
	return pitch, exists
}
