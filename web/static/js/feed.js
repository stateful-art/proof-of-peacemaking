document.addEventListener('DOMContentLoaded', function() {
    // Initialize the expression modal
    const expressionModal = new Modal('createExpressionModal');
    const expressionForm = new ExpressionForm('expressionForm');
    window.expressionModal = expressionModal; // Make it globally available

    // Create Expression button click handler
    const createExpressionBtn = document.getElementById('createExpressionBtn');
    if (createExpressionBtn) {
        createExpressionBtn.addEventListener('click', () => {
            console.log('Opening expression modal...'); // Debug log
            expressionModal.open();
        });
    } else {
        console.error('Create Expression button not found'); // Debug log
    }

    // Handle acknowledgment button clicks
    document.querySelectorAll('.acknowledge-button').forEach(button => {
        button.addEventListener('click', async function(e) {
            e.preventDefault();
            const expressionId = this.dataset.expressionId;
            const heartIcon = this.querySelector('.heart-icon');
            const countSpan = this.querySelector('.acknowledgement-count');
            const currentCount = parseInt(countSpan.textContent);
            
            // Optimistically update UI
            const isCurrentlyAcknowledged = button.classList.contains('acknowledged');
            const newCount = isCurrentlyAcknowledged ? currentCount - 1 : currentCount + 1;
            
            // Update UI immediately
            button.classList.toggle('acknowledged');
            heartIcon.classList.toggle('acknowledged');
            countSpan.textContent = newCount;
            
            try {
                const response = await fetch('/api/acknowledgements', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ expressionId })
                });

                if (!response.ok) {
                    // If request fails, revert the optimistic updates
                    button.classList.toggle('acknowledged');
                    heartIcon.classList.toggle('acknowledged');
                    countSpan.textContent = currentCount;
                    throw new Error('Failed to acknowledge expression');
                }

                // No need to update UI here since we already did it optimistically
                const result = await response.json();
                console.log('Acknowledgment updated:', result.status);
            } catch (error) {
                console.error('Error acknowledging expression:', error);
                // Show a toast or notification if you want to inform the user of the error
            }
        });
    });

    // Format time in MM:SS
    const formatTime = time => {
        if (!isFinite(time)) return '0:00';
        const minutes = Math.floor(time / 60);
        const seconds = Math.floor(time % 60);
        return `${minutes}:${seconds.toString().padStart(2, '0')}`;
    };

    // Initialize audio players
    document.querySelectorAll('.audio-player').forEach(player => {
        const audio = player.parentElement.querySelector('.audio-element');
        const playButton = player.querySelector('.play-button');
        const playIcon = playButton.querySelector('i');
        const progress = player.querySelector('.progress-bar');
        const currentTimeEl = player.querySelector('.current-time');
        const durationEl = player.querySelector('.duration');
        const volumeIcon = player.querySelector('.volume-icon');
        const volumeLevel = player.querySelector('.volume-level');
        const volumeSlider = player.querySelector('.volume-slider');
        const progressBar = player.querySelector('.audio-progress');

        // Set initial duration text
        if (audio.duration) {
            durationEl.textContent = formatTime(audio.duration);
        } else {
            audio.addEventListener('loadedmetadata', () => {
                durationEl.textContent = formatTime(audio.duration);
            });
            durationEl.textContent = '0:00';
        }

        // Update progress bar
        audio.addEventListener('timeupdate', () => {
            const percent = (audio.currentTime / audio.duration) * 100;
            progress.style.width = percent + '%';
            currentTimeEl.textContent = formatTime(audio.currentTime);
        });

        // Play/Pause
        playButton.addEventListener('click', () => {
            if (audio.paused) {
                audio.play();
                playIcon.classList.replace('fa-play', 'fa-pause');
            } else {
                audio.pause();
                playIcon.classList.replace('fa-pause', 'fa-play');
            }
        });

        // Click on progress bar
        progressBar.addEventListener('click', e => {
            const rect = progressBar.getBoundingClientRect();
            const percent = (e.clientX - rect.left) / rect.width;
            audio.currentTime = percent * audio.duration;
        });

        // Volume control
        volumeSlider.addEventListener('click', e => {
            const rect = volumeSlider.getBoundingClientRect();
            const percent = (e.clientX - rect.left) / rect.width;
            audio.volume = percent;
            volumeLevel.style.width = (percent * 100) + '%';
            updateVolumeIcon(percent);
        });

        // Update volume icon based on level
        const updateVolumeIcon = (volume) => {
            volumeIcon.className = 'fa-solid volume-icon ' + 
                (volume === 0 ? 'fa-volume-xmark' :
                 volume < 0.5 ? 'fa-volume-low' : 
                 'fa-volume-high');
        };

        // Toggle mute on volume icon click
        volumeIcon.addEventListener('click', () => {
            audio.muted = !audio.muted;
            if (audio.muted) {
                volumeLevel.style.width = '0%';
                volumeIcon.className = 'fa-solid fa-volume-xmark volume-icon';
            } else {
                volumeLevel.style.width = (audio.volume * 100) + '%';
                updateVolumeIcon(audio.volume);
            }
        });

        // Reset when audio ends
        audio.addEventListener('ended', () => {
            playIcon.classList.replace('fa-pause', 'fa-play');
            progress.style.width = '0%';
            currentTimeEl.textContent = '0:00';
        });
    });
}); 