package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Pitch struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

var (
	pitches     = make(map[string]Pitch)
	pitchesMu   sync.RWMutex
	storageFile = "pitches.json"
)

func initStorage() error {
	pitchesMu.Lock()
	defer pitchesMu.Unlock()

	if _, err := os.Stat(storageFile); os.IsNotExist(err) {
		return os.WriteFile(storageFile, []byte("{}"), 0644)
	}

	data, err := os.ReadFile(storageFile)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		data = []byte("{}")
	}

	return json.Unmarshal(data, &pitches)
}

func savePitch(pitch Pitch) error {
	pitchesMu.Lock()
	defer pitchesMu.Unlock()

	if pitch.ID == "" {
		pitch.ID = fmt.Sprintf("pitch_%d", time.Now().UnixNano())
	}
	if pitch.CreatedAt.IsZero() {
		pitch.CreatedAt = time.Now()
	}

	pitches[pitch.ID] = pitch
	return saveToFile()
}

func saveToFile() error {
	data, err := json.MarshalIndent(pitches, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(storageFile, data, 0644)
}

//func getPitch(id string) (Pitch, bool) {
//	pitchesMu.RLock()
//	defer pitchesMu.RUnlock()
//	pitch, exists := pitches[id]
//	return pitch, exists
//}

//func getAllPitches() map[string]Pitch {
//	pitchesMu.RLock()
//	defer pitchesMu.RUnlock()
//	result := make(map[string]Pitch)
//	for k, v := range pitches {
//		result[k] = v
//	}
//	return result
//}

//func deletePitch(id string) error {
//	pitchesMu.Lock()
//	defer pitchesMu.Unlock()
//	if _, exists := pitches[id]; !exists {
//		return fmt.Errorf("pitch non trouvé")
//	}
//	delete(pitches, id)
//	return saveToFile()
//}
