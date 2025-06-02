document.addEventListener('DOMContentLoaded', function() {
    // Éléments du DOM
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');
    const generateBtn = document.getElementById('generateBtn');
    const projectIdea = document.getElementById('projectIdea');
    const targetMarket = document.getElementById('targetMarket');
    const competitors = document.getElementById('competitors');
    const uniqueAspect = document.getElementById('uniqueAspect');
    const businessModel = document.getElementById('businessModel');
    const resultContainer = document.getElementById('resultContainer');
    const pitchResult = document.getElementById('pitchResult');
    const copyBtn = document.getElementById('copyBtn');
    const saveBtn = document.getElementById('saveBtn');
    const shareBtn = document.getElementById('shareBtn');
    const examplesList = document.getElementById('examplesList');
    const savedList = document.getElementById('savedList');
    const shareModal = document.getElementById('shareModal');
    const closeBtn = document.querySelector('.close-btn');
    const shareEmail = document.getElementById('shareEmail');
    const confirmShareBtn = document.getElementById('confirmShareBtn');
    const shareStatus = document.getElementById('shareStatus');

    let currentPitchId = null;

    // Gestion des onglets
    tabBtns.forEach(btn => {
        btn.addEventListener('click', function() {
            const tabId = this.getAttribute('data-tab');
            
            // Mettre à jour les boutons d'onglet
            tabBtns.forEach(b => b.classList.remove('active'));
            this.classList.add('active');
            
            // Mettre à jour le contenu des onglets
            tabContents.forEach(content => content.classList.remove('active'));
            document.getElementById(tabId).classList.add('active');
            
            // Charger les données si nécessaire
            if (tabId === 'examples') {
                loadExamples();
            } else if (tabId === 'saved') {
                loadSavedPitches();
            }
        });
    });

    // Charger les exemples
    function loadExamples() {
        fetch('/examples')
            .then(response => response.json())
            .then(examples => {
                examplesList.innerHTML = examples.map((example, index) => `
                    <div class="example-card">
                        <h3>Exemple ${index + 1}</h3>
                        <div class="example-content">${example.replace(/\n/g, '<br>')}</div>
                        <button class="use-example-btn" data-example="${index}">Utiliser cet exemple</button>
                    </div>
                `).join('');
                
                // Ajouter les événements aux boutons
                document.querySelectorAll('.use-example-btn').forEach(btn => {
                    btn.addEventListener('click', function() {
                        const index = this.getAttribute('data-example');
                        projectIdea.value = examples[index].split('\n')[0].replace('Problème: ', '');
                        document.querySelector('.tab-btn[data-tab="generator"]').click();
                    });
                });
            })
            .catch(error => console.error('Error loading examples:', error));
    }

    // Charger les pitchs sauvegardés
    function loadSavedPitches() {
        // Dans une vraie application, vous feriez une requête au backend
        // Pour cet exemple, nous affichons simplement un message
        savedList.innerHTML = '<p>Connectez-vous pour voir vos pitchs sauvegardés</p>';
    }

    // Générer un pitch
    generateBtn.addEventListener('click', async function() {
        const idea = projectIdea.value.trim();
        
        if (!idea) {
            alert('Veuillez entrer une idée de projet');
            return;
        }
        
        generateBtn.disabled = true;
        generateBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Génération en cours...';
        
        const projectDetails = {
            idea: idea,
            targetMarket: targetMarket.value.trim(),
            competitors: competitors.value.trim(),
            uniqueAspect: uniqueAspect.value.trim(),
            businessModel: businessModel.value.trim()
        };
        
        try {
            const response = await fetch('/generate-pitch', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(projectDetails)
            });
            
            if (!response.ok) {
                throw new Error('Erreur lors de la génération du pitch');
            }
            
            const data = await response.json();
            pitchResult.innerHTML = data.pitch.replace(/\n/g, '<br>');
            resultContainer.classList.remove('hidden');
            currentPitchId = data.id;
            
            // Faire défiler jusqu'au résultat
            resultContainer.scrollIntoView({ behavior: 'smooth' });
            
        } catch (error) {
            console.error('Error:', error);
            alert('Une erreur est survenue. Veuillez réessayer.');
        } finally {
            generateBtn.disabled = false;
            generateBtn.innerHTML = '<i class="fas fa-magic"></i> Générer le pitch';
        }
    });
    
    // Copier le pitch
    copyBtn.addEventListener('click', function() {
        const text = pitchResult.textContent;
        navigator.clipboard.writeText(text)
            .then(() => {
                copyBtn.innerHTML = '<i class="fas fa-check"></i> Copié!';
                setTimeout(() => {
                    copyBtn.innerHTML = '<i class="far fa-copy"></i> Copier';
                }, 2000);
            })
            .catch(err => {
                console.error('Erreur lors de la copie:', err);
            });
    });
    
    // Sauvegarder le pitch
    saveBtn.addEventListener('click', function() {
        if (!currentPitchId) {
            alert('Générez d\'abord un pitch avant de sauvegarder');
            return;
        }
        
        // Dans une vraie application, vous feriez une requête au backend
        // pour confirmer la sauvegarde
        saveBtn.innerHTML = '<i class="fas fa-check"></i> Sauvegardé!';
        setTimeout(() => {
            saveBtn.innerHTML = '<i class="far fa-save"></i> Sauvegarder';
        }, 2000);
    });
    
    // Partager le pitch
    shareBtn.addEventListener('click', function() {
        if (!currentPitchId) {
            alert('Générez d\'abord un pitch avant de partager');
            return;
        }
        
        shareModal.classList.remove('hidden');
        shareEmail.value = '';
        shareStatus.classList.add('hidden');
    });
    
    // Fermer le modal
    closeBtn.addEventListener('click', function() {
        shareModal.classList.add('hidden');
    });
    
    // Confirmer le partage
    confirmShareBtn.addEventListener('click', function() {
        const email = shareEmail.value.trim();
        
        if (!email || !email.includes('@')) {
            alert('Veuillez entrer une adresse email valide');
            return;
        }
        
        confirmShareBtn.disabled = true;
        confirmShareBtn.textContent = 'Envoi en cours...';
        
        fetch('/share', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                pitch: pitchResult.textContent,
                email: email
            })
        })
        .then(response => response.json())
        .then(data => {
            shareStatus.textContent = data.message;
            shareStatus.classList.remove('hidden');
            shareStatus.classList.add('success');
            
            setTimeout(() => {
                shareModal.classList.add('hidden');
                confirmShareBtn.disabled = false;
                confirmShareBtn.textContent = 'Envoyer';
            }, 2000);
        })
        .catch(error => {
            shareStatus.textContent = 'Erreur lors du partage';
            shareStatus.classList.remove('hidden');
            shareStatus.classList.add('error');
            confirmShareBtn.disabled = false;
            confirmShareBtn.textContent = 'Envoyer';
        });
    });
    
    // Initialiser l'application
    loadExamples();
});