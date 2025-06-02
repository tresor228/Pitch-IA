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
		emptyData := make(map[string]Pitch)
		data, err := json.MarshalIndent(emptyData, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(storageFile, data, 0644)
	}

	// Lit le fichier
	data, err := os.ReadFile(storageFile)
	if err != nil {
		return err
	}

	// Vérifier si le fichier est vide
	if len(data) == 0 {
		pitches = make(map[string]Pitch)
		return nil
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
	return fmt.Sprintf("pitch_%d", time.Now().UnixNano())
}

// getPitch récupère un pitch par son ID (version thread-safe)
func getPitch(id string) (Pitch, bool) {
	pitchesMu.RLock()
	defer pitchesMu.RUnlock()

	pitch, exists := pitches[id]
	return pitch, exists
}

// getAllPitches récupère tous les pitches (version thread-safe)
func getAllPitches() map[string]Pitch {
	pitchesMu.RLock()
	defer pitchesMu.RUnlock()

	// Créer une copie pour éviter les accès concurrents
	result := make(map[string]Pitch)
	for k, v := range pitches {
		result[k] = v
	}
	return result
}

// deletePitch supprime un pitch par son ID
func deletePitch(id string) error {
	pitchesMu.Lock()
	defer pitchesMu.Unlock()

	if _, exists := pitches[id]; !exists {
		return fmt.Errorf("pitch with ID %s not found", id)
	}

	delete(pitches, id)
	return saveToFile()
}
