package service

import (
	"fmt"
	"pitch/models"
)

// GeneratePitchResponse retourne un PitchResponse mocké basé sur l'entrée.
// Remarque: plus tard on remplacera cette implémentation par un appel réel à OpenAI.
func GeneratePitchResponse(input string) *models.PitchResponse {
	// Génération simple et lisible pour démonstration
	probleme := fmt.Sprintf("Les utilisateurs rencontrent %s, causant une expérience frustrante.", input)
	solution := fmt.Sprintf("Nous proposons une solution qui automatise et simplifie : %s, avec une interface intuitive.", input)
	marche := "Étudiants urbains, 18-30 ans, utilisateurs mobiles cherchant commodité et bonnes options alimentaires."
	valeur := "Gain de temps, menus healthy personnalisés et coût compétitif."
	canaux := "Campagnes réseaux sociaux, partenariats campus, marketing d'influence et SEO local."
	modele := "Abonnement mensuel + commissions sur commandes; version freemium disponible."

	return &models.PitchResponse{
		Probleme: probleme,
		Solution: solution,
		Marche:   marche,
		Valeur:   valeur,
		Canaux:   canaux,
		Modele:   modele,
	}
}
