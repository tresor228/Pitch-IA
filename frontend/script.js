class PitchGenerator {
    constructor() {
        this.initializeElements();
        this.attachEventListeners();
    }

    initializeElements() {
        this.generateBtn = document.getElementById('generateBtn');
        this.projectIdea = document.getElementById('projectIdea');
        this.targetMarket = document.getElementById('targetMarket');
        this.uniqueValue = document.getElementById('uniqueValue');
        this.resultContainer = document.getElementById('resultContainer');
        this.pitchResult = document.getElementById('pitchResult');
        this.copyBtn = document.getElementById('copyBtn');
        this.regenerateBtn = document.getElementById('regenerateBtn');
        this.loadingOverlay = document.getElementById('loadingOverlay');
    }

    attachEventListeners() {
        this.generateBtn.addEventListener('click', () => this.generatePitch());
        this.copyBtn.addEventListener('click', () => this.copyPitch());
        this.regenerateBtn.addEventListener('click', () => this.generatePitch());
        this.projectIdea.addEventListener('input', () => this.validateForm());
    }

    validateForm() {
        this.generateBtn.disabled = this.projectIdea.value.trim().length < 10;
    }

    async generatePitch() {
        const idea = this.projectIdea.value.trim();
        
        if (!idea || idea.length < 10) {
            this.showError('Veuillez décrire votre idée (minimum 10 caractères)');
            return;
        }

        try {
            this.setLoading(true);
            const pitch = await this.createPitch({
                idea: idea,
                targetMarket: this.targetMarket.value.trim(),
                uniqueValue: this.uniqueValue.value.trim()
            });
            
            this.displayPitch(pitch);
            this.showSuccess('Pitch généré avec succès !');
        } catch (error) {
            this.showError('Erreur: ' + error.message);
            console.error(error);
        } finally {
            this.setLoading(false);
        }
    }

    async createPitch(data) {
        const response = await fetch('/generate-pitch', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                idea: data.idea,
                targetMarket: data.targetMarket,
                uniqueAspect: data.uniqueValue,
                businessModel: ""
            })
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Erreur serveur');
        }
        
        const result = await response.json();
        return result.pitch;
    }

    displayPitch(pitch) {
        this.pitchResult.innerHTML = pitch.replace(/\n/g, '<br>');
        this.resultContainer.classList.remove('hidden');
        this.resultContainer.scrollIntoView({ behavior: 'smooth' });
    }

    async copyPitch() {
        try {
            await navigator.clipboard.writeText(this.pitchResult.textContent);
            this.copyBtn.innerHTML = '<i class="fas fa-check"></i> Copié !';
            setTimeout(() => {
                this.copyBtn.innerHTML = '<i class="far fa-copy"></i> Copier';
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
        this.removeMessages();
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.innerHTML = `<i class="fas fa-exclamation-triangle"></i> ${message}`;
        this.generateBtn.parentNode.appendChild(errorDiv);
        setTimeout(() => errorDiv.remove(), 5000);
    }

    showSuccess(message) {
        this.removeMessages();
        const successDiv = document.createElement('div');
        successDiv.className = 'success-message';
        successDiv.innerHTML = `<i class="fas fa-check-circle"></i> ${message}`;
        this.resultContainer.appendChild(successDiv);
        setTimeout(() => successDiv.remove(), 3000);
    }

    removeMessages() {
        document.querySelectorAll('.error-message, .success-message').forEach(el => el.remove());
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new PitchGenerator();
});