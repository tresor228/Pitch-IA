class PitchGenerator {
    constructor() {
        this.initializeElements();
        this.attachEventListeners();
        this.validateForm();
    }

    initializeElements() {
        this.generateBtn = document.getElementById('generateBtn');
        this.projectIdea = document.getElementById('projectIdea');
        this.resultContainer = document.getElementById('resultContainer');
        this.loadingOverlay = document.getElementById('loadingOverlay');
        
        this.problemResult = document.getElementById('problemResult');
        this.solutionResult = document.getElementById('solutionResult');
        this.marketResult = document.getElementById('marketResult');
        this.valueResult = document.getElementById('valueResult');
        this.channelsResult = document.getElementById('channelsResult');
        this.businessModelResult = document.getElementById('businessModelResult');
        this.pitchResult = document.getElementById('pitchResult');
        this.copyBtn = document.getElementById('copyBtn');
    }

    attachEventListeners() {
        this.generateBtn.addEventListener('click', () => this.generatePitch());
        this.copyBtn.addEventListener('click', () => this.copyPitch());
        this.projectIdea.addEventListener('input', () => this.validateForm());
    }
    
    validateForm() {
        const isValid = this.projectIdea.value.trim().length >= 10;
        this.generateBtn.disabled = !isValid;
    }

    async generatePitch() {
        const idea = this.projectIdea.value.trim();
        
        if (!idea || idea.length < 10) {
            this.showError('Veuillez décrire votre idée (minimum 10 caractères)');
            return;
        }

        try {
            this.setLoading(true);
            const pitchData = await this.fetchPitch({ idea: idea });
            this.displayPitch(pitchData);
            this.showSuccess('Pitch généré avec succès !');
        } catch (error) {
            this.showError('Erreur: ' + error.message);
            console.error(error);
        } finally {
            this.setLoading(false);
        }
    }

    async fetchPitch(data) {
        const response = await fetch('/generate-pitch', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data)
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Erreur serveur');
        }
        
        return await response.json();
    }

    displayPitch(pitchData) {
        // Reset all fields
        this.problemResult.textContent = '';
        this.solutionResult.textContent = '';
        this.marketResult.textContent = '';
        this.valueResult.textContent = '';
        this.channelsResult.textContent = '';
        this.businessModelResult.textContent = '';
        
        // Try to extract sections if they exist
        if (pitchData.problem) {
            this.problemResult.textContent = pitchData.problem;
        }
        if (pitchData.solution) {
            this.solutionResult.textContent = pitchData.solution;
        }
        if (pitchData.targetMarket) {
            this.marketResult.textContent = pitchData.targetMarket;
        }
        if (pitchData.valueProposition) {
            this.valueResult.textContent = pitchData.valueProposition;
        }
        if (pitchData.channels) {
            this.channelsResult.textContent = pitchData.channels;
        }
        if (pitchData.businessModel) {
            this.businessModelResult.textContent = pitchData.businessModel;
        }
        
        // Display full pitch with formatting
        if (pitchData.pitch) {
            this.pitchResult.innerHTML = this.formatPitchContent(pitchData.pitch);
        } else {
            this.pitchResult.textContent = "Aucun pitch généré";
        }
        
        this.resultContainer.classList.remove('hidden');
        this.resultContainer.scrollIntoView({ behavior: 'smooth' });
    }

    formatPitchContent(content) {
        return content
            .replace(/\[Problème\]/g, '<h3>Problème</h3>')
            .replace(/\[Solution\]/g, '<h3>Solution</h3>')
            .replace(/\[Marché\]/g, '<h3>Marché</h3>')
            .replace(/\[Valeur\]/g, '<h3>Valeur</h3>')
            .replace(/\[Canaux\]/g, '<h3>Canaux</h3>')
            .replace(/\[Modèle\]/g, '<h3>Modèle</h3>')
            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
            .replace(/\n/g, '<br>');
    }

    async copyPitch() {
        try {
            await navigator.clipboard.writeText(this.pitchResult.textContent);
            this.copyBtn.innerHTML = '<i class="fas fa-check"></i> Copié !';
            setTimeout(() => {
                this.copyBtn.innerHTML = '<i class="far fa-copy"></i> Copier le Pitch';
            }, 2000);
        } catch (error) {
            this.showError('Copie manuelle nécessaire');
        }
    }

    setLoading(isLoading) {
        this.loadingOverlay.classList.toggle('hidden', !isLoading);
        this.generateBtn.disabled = isLoading;
    }

    showError(message) {
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.innerHTML = `<i class="fas fa-exclamation-triangle"></i> ${message}`;
        this.generateBtn.parentNode.appendChild(errorDiv);
        setTimeout(() => errorDiv.remove(), 5000);
    }

    showSuccess(message) {
        const successDiv = document.createElement('div');
        successDiv.className = 'success-message';
        successDiv.innerHTML = `<i class="fas fa-check-circle"></i> ${message}`;
        this.resultContainer.appendChild(successDiv);
        setTimeout(() => successDiv.remove(), 3000);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new PitchGenerator();
});