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
                
                // Validation en temps réel
                this.projectIdea.addEventListener('input', () => this.validateForm());
            }

            validateForm() {
                const isValid = this.projectIdea.value.trim().length > 10;
                this.generateBtn.disabled = !isValid;
            }

            async generatePitch() {
                const idea = this.projectIdea.value.trim();
                
                if (!idea || idea.length < 10) {
                    this.showError('Veuillez décrire votre idée plus en détail (minimum 10 caractères)');
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
                    this.showError('Une erreur est survenue. Veuillez réessayer.');
                    console.error('Erreur:', error);
                } finally {
                    this.setLoading(false);
                }
            }

            async createPitch(data) {
                // Simulation d'un appel API avec génération de pitch local
                return new Promise((resolve) => {
                    setTimeout(() => {
                        const pitch = this.generatePitchText(data);
                        resolve(pitch);
                    }, 2000);
                });
            }

            generatePitchText(data) {
                const { idea, targetMarket, uniqueValue } = data;
                
                const templates = [
                    `🎯 **LE PROBLÈME**
${idea.split('.')[0] || idea}

👥 **NOTRE CIBLE**
${targetMarket || 'Entrepreneurs et professionnels'} qui cherchent une solution efficace et accessible.

💡 **NOTRE SOLUTION**
Nous proposons une approche innovante qui résout ce problème de manière simple et efficace.

⭐ **NOTRE AVANTAGE**
${uniqueValue || 'Une solution intuitive et abordable'} qui nous différencie de la concurrence.

🚀 **L'OPPORTUNITÉ**
Le marché est prêt pour une solution comme la nôtre. C'est le moment idéal pour agir.

💰 **LE POTENTIEL**
Avec notre approche, nous visons une croissance rapide et durable sur ce marché en expansion.`,

                    `🔍 **LE DÉFI**
${idea}

🎯 **NOTRE MISSION**
Aider ${targetMarket || 'nos clients'} à surmonter ce défi grâce à notre solution innovante.

✨ **CE QUI NOUS REND UNIQUES**
${uniqueValue || 'Notre approche unique'} nous permet de nous démarquer et d'offrir une valeur exceptionnelle.

📈 **L'IMPACT**
Notre solution transforme la façon dont nos clients abordent ce problème, avec des résultats mesurables.

🌟 **LA VISION**
Nous construisons l'avenir de ce secteur, un client satisfait à la fois.`
                ];

                return templates[Math.floor(Math.random() * templates.length)];
            }

            displayPitch(pitch) {
                this.pitchResult.textContent = pitch;
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
                    this.showError('Impossible de copier. Sélectionnez le texte manuellement.');
                }
            }

            setLoading(isLoading) {
                if (isLoading) {
                    this.loadingOverlay.classList.remove('hidden');
                    this.generateBtn.disabled = true;
                } else {
                    this.loadingOverlay.classList.add('hidden');
                    this.generateBtn.disabled = false;
                }
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

        // Initialisation de l'application
        document.addEventListener('DOMContentLoaded', () => {
            new PitchGenerator();
        });