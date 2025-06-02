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
            this.showError(error.message || 'Erreur lors de la génération du pitch');
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
            const errorData = await response.json().catch(() => ({}));
            let errorMessage = errorData.error || 'Erreur serveur';
            
            if (response.status === 429 || errorMessage.includes('quota')) {
                errorMessage = "Limite de requêtes atteinte. Veuillez réessayer plus tard.";
            } else if (response.status === 503) {
                errorMessage = "Service temporairement indisponible";
            }
            
            throw new Error(errorMessage);
        }
        
        return await response.json();
    }

    displayPitch(pitchData) {
        this.problemResult.textContent = '';
        this.solutionResult.textContent = '';
        this.marketResult.textContent = '';
        this.valueResult.textContent = '';
        this.channelsResult.textContent = '';
        this.businessModelResult.textContent = '';
        
        if (pitchData.pitch) {
            this.pitchResult.innerHTML = this.formatPitchContent(pitchData.pitch);
            
            // Extraction des sections si elles existent
            const sections = this.extractSections(pitchData.pitch);
            if (sections.problem) this.problemResult.textContent = sections.problem;
            if (sections.solution) this.solutionResult.textContent = sections.solution;
            if (sections.market) this.marketResult.textContent = sections.market;
            if (sections.value) this.valueResult.textContent = sections.value;
            if (sections.channels) this.channelsResult.textContent = sections.channels;
            if (sections.model) this.businessModelResult.textContent = sections.model;
        } else {
            this.pitchResult.textContent = "Aucun pitch généré";
        }
        
        this.resultContainer.classList.remove('hidden');
        this.resultContainer.scrollIntoView({ behavior: 'smooth' });
    }

    extractSections(content) {
        const sections = {};
        const regex = /(\d+\.\s*\[(.*?)\]\s*)(.*?)(?=\n\d+\.|$)/g;
        let match;
        
        while ((match = regex.exec(content)) !== null) {
            if (match[2] && match[3]) {
                sections[match[2].toLowerCase()] = match[3].trim();
            }
        }
        
        return {
            problem: sections['problème'] || '',
            solution: sections['solution'] || '',
            market: sections['marché'] || '',
            value: sections['valeur'] || '',
            channels: sections['canaux'] || '',
            model: sections['modèle'] || ''
        };
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