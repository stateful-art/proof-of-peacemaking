class Modal {
    constructor(modalId) {
        this.modal = document.getElementById(modalId);
        this.isOpen = false;
        this.setupEventListeners();
    }

    setupEventListeners() {
        // Close on backdrop click
        this.modal.addEventListener('click', (e) => {
            if (e.target === this.modal) {
                this.close();
            }
        });

        // Close on ESC key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isOpen) {
                this.close();
            }
        });

        // Close button
        const closeBtn = this.modal.querySelector('.modal-close');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => this.close());
        }
    }

    open() {
        if (!this.modal) return;
        this.modal.classList.add('active');
        this.isOpen = true;
        document.body.style.overflow = 'hidden'; // Prevent background scrolling
        console.log('Modal opened:', this.modal.id); // Debug log
    }

    close() {
        if (!this.modal) return;
        this.modal.classList.remove('active');
        this.isOpen = false;
        document.body.style.overflow = ''; // Restore scrolling
        console.log('Modal closed:', this.modal.id); // Debug log
    }
}

// Expression creation form handling
class ExpressionForm {
    constructor(formId) {
        this.form = document.getElementById(formId);
        this.mediaRecorder = null;
        this.recordedChunks = [];
        this.isRecording = false;
        this.currentMediaType = null; // Track current media type
        this.hasAudio = false;
        this.hasVideo = false;
        this.hasImage = false;
        
        if (this.form) {
            this.setupEventListeners();
        } else {
            console.error('Expression form not found:', formId);
        }
    }

    setupEventListeners() {
        const audioRecordBtn = document.getElementById('audioRecord');
        const videoRecordBtn = document.getElementById('videoRecord');
        const imageUploadBtn = document.getElementById('imageUpload');
        const imageInput = document.getElementById('imageInput');
        const mediaPreview = document.getElementById('mediaPreview');

        if (audioRecordBtn) {
            audioRecordBtn.addEventListener('click', () => {
                if (!this.isRecording) {
                    this.startRecording('audio');
                } else {
                    this.stopRecording();
                }
            });
        }

        if (videoRecordBtn) {
            videoRecordBtn.addEventListener('click', () => {
                if (!this.isRecording) {
                    this.startRecording('video');
                } else {
                    this.stopRecording();
                }
            });
        }

        if (imageUploadBtn && imageInput) {
            imageUploadBtn.addEventListener('click', () => {
                // Check for both hasImage and hasVideo before allowing upload
                if (this.hasVideo) {
                    console.log('Cannot upload image when video exists');
                    return;
                }
                if (!this.hasImage) {
                    imageInput.click();
                }
            });

            imageInput.addEventListener('change', (e) => {
                this.handleImageUpload(e);
            });
        }

        // Add cancel button functionality
        mediaPreview.addEventListener('click', (e) => {
            if (e.target.classList.contains('cancel-media')) {
                const mediaType = e.target.dataset.type;
                this.cancelMedia(mediaType);
            }
        });

        this.form.addEventListener('submit', (e) => this.handleSubmit(e));
    }

    updateButtonStates() {
        const audioBtn = document.getElementById('audioRecord');
        const videoBtn = document.getElementById('videoRecord');
        const imageBtn = document.getElementById('imageUpload');

        // If video is present, disable everything else
        if (this.hasVideo) {
            if (audioBtn) audioBtn.disabled = true;
            if (videoBtn) videoBtn.disabled = true;
            if (imageBtn) imageBtn.disabled = true;
            return;
        }

        // If audio or image is present, disable video but allow the other
        if (this.hasAudio || this.hasImage) {
            if (videoBtn) videoBtn.disabled = true;
        }

        // Individual media type checks
        if (audioBtn) audioBtn.disabled = this.hasAudio;
        if (videoBtn) videoBtn.disabled = this.hasVideo;
        if (imageBtn) imageBtn.disabled = this.hasImage;

        // Update button states
        if (audioBtn) {
            audioBtn.classList.toggle('recording', this.isRecording && this.currentMediaType === 'audio');
        }
        if (videoBtn) {
            videoBtn.classList.toggle('recording', this.isRecording && this.currentMediaType === 'video');
        }
    }

    cancelMedia(type) {
        console.log('Canceling media:', type);
        
        if (this.isRecording) {
            this.stopRecording();
        }

        switch (type) {
            case 'audio':
                this.hasAudio = false;
                break;
            case 'video':
                this.hasVideo = false;
                break;
            case 'image':
                this.hasImage = false;
                document.getElementById('imageInput').value = '';
                break;
        }

        // Remove the specific media preview
        const mediaContainer = document.querySelector(`[data-media-type="${type}"]`);
        if (mediaContainer) {
            mediaContainer.remove();
        }

        this.updateButtonStates();
    }

    createMediaPreviewContainer(type) {
        const container = document.createElement('div');
        container.classList.add('media-container');
        container.dataset.mediaType = type;

        const cancelBtn = document.createElement('button');
        cancelBtn.type = 'button';
        cancelBtn.classList.add('cancel-media');
        cancelBtn.dataset.type = type;
        cancelBtn.innerHTML = '×';
        cancelBtn.title = 'Remove ' + type;

        container.appendChild(cancelBtn);
        return container;
    }

    createAudioPlayer() {
        const audioPlayer = document.createElement('div');
        audioPlayer.className = 'audio-player';
        audioPlayer.innerHTML = `
            <button type="button" class="play-button">
                <i class="fa-solid fa-play"></i>
            </button>
            <div class="audio-controls">
                <span class="time-info current-time">0:00</span>
                <div class="audio-progress">
                    <div class="progress-bar"></div>
                </div>
                <span class="time-info duration">0:00</span>
            </div>
            <div class="volume-control">
                <i class="fa-solid fa-volume-high volume-icon"></i>
                <div class="volume-slider">
                    <div class="volume-level" style="width: 100%"></div>
                </div>
            </div>
        `;
        return audioPlayer;
    }

    async startRecording(mediaType) {
        try {
            // Check if we already have this media type
            if ((mediaType === 'audio' && this.hasAudio) || 
                (mediaType === 'video' && this.hasVideo)) {
                console.log(`${mediaType} already exists`);
                return;
            }

            // If trying to record video but we have audio or image, prevent it
            if (mediaType === 'video' && (this.hasAudio || this.hasImage)) {
                console.log('Cannot record video when audio or image exists');
                return;
            }

            // If trying to record audio but we have video, prevent it
            if (mediaType === 'audio' && this.hasVideo) {
                console.log('Cannot record audio when video exists');
                return;
            }

            if (this.isRecording) {
                await this.stopRecording();
            }
            
            console.log(`Starting ${mediaType} recording...`);
            const constraints = {
                audio: true,
                video: mediaType === 'video' ? {
                    width: { ideal: 1280 },
                    height: { ideal: 720 }
                } : false
            };

            const stream = await navigator.mediaDevices.getUserMedia(constraints);
            
            // Create container for the media preview
            const container = this.createMediaPreviewContainer(mediaType);
            document.getElementById('mediaPreview').appendChild(container);
            
            if (mediaType === 'video') {
                const previewVideo = document.createElement('video');
                previewVideo.srcObject = stream;
                previewVideo.autoplay = true;
                previewVideo.controls = true;
                previewVideo.muted = true;
                previewVideo.classList.add('media-preview');
                container.appendChild(previewVideo);
            } else if (mediaType === 'audio') {
                const audio = document.createElement('audio');
                audio.classList.add('audio-element');
                container.appendChild(audio);
                container.appendChild(this.createAudioPlayer());
                this.initializeAudioPlayer(container);
            }
            
            // Try different MIME types in order of preference
            const videoTypes = [
                'video/webm;codecs=vp8,opus',
                'video/webm;codecs=vp9,opus',
                'video/webm;codecs=h264,opus',
                'video/webm'
            ];

            let selectedMimeType = mediaType === 'video' ? 
                videoTypes.find(type => MediaRecorder.isTypeSupported(type)) : 
                'audio/webm;codecs=opus';

            if (!selectedMimeType) {
                console.error('No supported MIME type found');
                selectedMimeType = 'video/webm'; // Fallback to basic webm
            }

            console.log('Using MIME type:', selectedMimeType);
            
            const options = {
                mimeType: selectedMimeType,
                videoBitsPerSecond: 2500000, // 2.5 Mbps
                audioBitsPerSecond: 128000
            };
            
            this.mediaRecorder = new MediaRecorder(stream, options);
            this.recordedChunks = [];
            this.currentMediaType = mediaType;
            
            this.mediaRecorder.ondataavailable = (event) => {
                if (event.data.size > 0) {
                    this.recordedChunks.push(event.data);
                }
            };

            this.mediaRecorder.onstop = () => {
                const blob = new Blob(this.recordedChunks, {
                    type: mediaType === 'video' ? 'video/webm' : 'audio/webm'
                });
                const url = URL.createObjectURL(blob);
                
                // Find the container for this media type
                const container = document.querySelector(`[data-media-type="${mediaType}"]`);
                if (container) {
                    // Keep the cancel button
                    const cancelBtn = container.querySelector('.cancel-media');
                    container.innerHTML = '';
                    container.appendChild(cancelBtn);

                    if (mediaType === 'audio') {
                        const audio = document.createElement('audio');
                        audio.src = url;
                        audio.classList.add('audio-element');
                        container.appendChild(audio);
                        
                        const audioPlayer = this.createAudioPlayer();
                        container.appendChild(audioPlayer);
                        this.initializeAudioPlayer(container);
                    } else {
                        const video = document.createElement('video');
                        video.src = url;
                        video.controls = true;
                        video.autoplay = false;
                        video.classList.add('media-preview');
                        container.appendChild(video);
                    }
                }

                if (mediaType === 'audio') {
                    this.hasAudio = true;
                } else {
                    this.hasVideo = true;
                }
            };

            this.mediaRecorder.start(1000);
            this.isRecording = true;
            
            this.updateButtonStates();
            
        } catch (err) {
            console.error('Recording error:', err);
            document.getElementById('mediaError').textContent = 
                'Error accessing media devices. Please ensure you have given permission.';
        }
    }

    initializeAudioPlayer(container) {
        const audio = container.querySelector('.audio-element');
        const player = container.querySelector('.audio-player');
        const playButton = player.querySelector('.play-button');
        const playIcon = playButton.querySelector('i');
        const progress = player.querySelector('.progress-bar');
        const currentTimeEl = player.querySelector('.current-time');
        const durationEl = player.querySelector('.duration');
        const volumeIcon = player.querySelector('.volume-icon');
        const volumeLevel = player.querySelector('.volume-level');
        const volumeSlider = player.querySelector('.volume-slider');
        const progressBar = player.querySelector('.audio-progress');

        // Format time in MM:SS
        const formatTime = time => {
            if (!isFinite(time)) return '0:00';
            const minutes = Math.floor(time / 60);
            const seconds = Math.floor(time % 60);
            return `${minutes}:${seconds.toString().padStart(2, '0')}`;
        };

        // Update progress bar
        audio.addEventListener('timeupdate', () => {
            const percent = (audio.currentTime / audio.duration) * 100;
            progress.style.width = percent + '%';
            currentTimeEl.textContent = formatTime(audio.currentTime);
        });

        // Set duration when metadata is loaded
        audio.addEventListener('loadedmetadata', () => {
            durationEl.textContent = formatTime(audio.duration);
        });

        // Set initial duration text
        if (audio.duration) {
            durationEl.textContent = formatTime(audio.duration);
        } else {
            durationEl.textContent = '0:00';
        }

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
    }

    async stopRecording() {
        console.log('Stopping recording...');
        if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
            console.log('MediaRecorder state before stop:', this.mediaRecorder.state);
            this.mediaRecorder.stop();
            this.mediaRecorder.stream.getTracks().forEach(track => {
                console.log('Stopping track:', track.kind);
                track.stop();
            });
            this.isRecording = false;
            console.log('Recording stopped, chunks:', this.recordedChunks.length);
            
            document.querySelectorAll('.btn-record').forEach(button => {
                button.classList.remove('recording');
            });
        } else {
            console.log('No active MediaRecorder to stop');
        }
    }

    handleImageUpload(e) {
        const file = e.target.files[0];
        if (file) {
            // If we have video, prevent image upload
            if (this.hasVideo) {
                console.log('Cannot upload image when video exists');
                e.target.value = ''; // Clear the file input
                return;
            }

            const reader = new FileReader();
            reader.onload = (e) => {
                // Create container for the image preview
                const container = this.createMediaPreviewContainer('image');
                document.getElementById('mediaPreview').appendChild(container);

                const img = document.createElement('img');
                img.src = e.target.result;
                img.classList.add('media-preview');
                container.appendChild(img);
                
                this.hasImage = true;
                this.updateButtonStates();
            };
            reader.readAsDataURL(file);
        }
    }

    async handleSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData();
        
        // Get text content from contenteditable div
        const textContent = this.form.querySelector('.text-input').innerText.trim();
        formData.append('textContent', textContent);
        
        // Add image if exists
        const imageInput = document.getElementById('imageInput');
        if (imageInput.files.length > 0) {
            formData.append('imageContent', imageInput.files[0]);
            console.log('Added image file:', imageInput.files[0]);
        }

        // Add audio/video if exists
        if (this.recordedChunks.length > 0) {
            console.log('=== Media Recording Debug Info ===');
            console.log('Current media type:', this.currentMediaType);
            console.log('Recorded chunks:', this.recordedChunks.length);
            
            const isVideo = this.currentMediaType === 'video';
            const mediaBlob = new Blob(this.recordedChunks, {
                type: isVideo ? 'video/webm' : 'audio/webm'
            });
            console.log('\n=== Media Blob Info ===');
            console.log('Media blob type:', mediaBlob.type);
            console.log('Media blob size:', mediaBlob.size, 'bytes');
            
            const filename = isVideo ? 'video.webm' : 'audio.webm';
            const file = new File([mediaBlob], filename, { type: mediaBlob.type });
            console.log('\n=== Created File Info ===');
            console.log('File name:', file.name);
            console.log('File type:', file.type);
            console.log('File size:', file.size, 'bytes');
            
            formData.append(isVideo ? 'videoContent' : 'audioContent', file);
        }

        // Log final form data contents
        console.log('\n=== Form Data Contents ===');
        for (let [key, value] of formData.entries()) {
            console.log(`${key}:`, value instanceof File ? 
                `File(name=${value.name}, type=${value.type}, size=${value.size})` : 
                value);
        }
        
        try {
            console.log('\n=== Sending Request ===');
            const response = await fetch('/api/expressions', {
                method: 'POST',
                body: formData,
            });

            if (!response.ok) {
                const error = await response.text();
                throw new Error(error || 'Failed to create expression');
            }

            console.log('Expression created successfully');
            
            // Clear form and close modal
            this.form.reset();
            document.getElementById('mediaPreview').innerHTML = '';
            this.recordedChunks = [];
            expressionModal.close();

            // Add delay before reload to see logs
            console.log('Reloading page in 4 seconds...');
            setTimeout(() => {
                location.reload();
            }, 1000);

        } catch (error) {
            console.error('Error creating expression:', error);
            document.getElementById('mediaError').textContent = 
                'Failed to create expression: ' + error.message;
        }
    }
}

// Initialize and export for use in other files
window.Modal = Modal;
window.ExpressionForm = ExpressionForm; 