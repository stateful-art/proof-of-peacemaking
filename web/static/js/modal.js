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
        container.style.position = 'relative'; // Ensure container is relative

        const cancelBtn = document.createElement('button');
        cancelBtn.type = 'button';
        cancelBtn.classList.add('cancel-media');
        cancelBtn.dataset.type = type;
        cancelBtn.innerHTML = '×';
        cancelBtn.title = 'Remove ' + type;
        
        // Position relative to container
        cancelBtn.style.position = 'absolute';
        cancelBtn.style.top = '-10px';
        cancelBtn.style.right = '-10px';
        cancelBtn.style.zIndex = '10000';
        cancelBtn.style.width = '20px';
        cancelBtn.style.height = '20px';
        cancelBtn.style.borderRadius = '50px';
        cancelBtn.style.backgroundColor = 'white';
        cancelBtn.style.color = 'black';
        cancelBtn.style.border = 'none';
        cancelBtn.style.cursor = 'pointer';
        cancelBtn.style.display = 'flex';
        cancelBtn.style.alignItems = 'center';
        cancelBtn.style.justifyContent = 'center';

        container.appendChild(cancelBtn);
        return container;
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
                video: mediaType === 'video'
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
            }
            
            const options = {
                mimeType: mediaType === 'video' ? 'video/webm' : 'audio/webm',
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
                
                const mediaElement = document.createElement(mediaType);
                mediaElement.src = url;
                mediaElement.controls = true;
                mediaElement.autoplay = false;
                mediaElement.classList.add('media-preview');
                
                // Find the container for this media type and replace its contents
                const container = document.querySelector(`[data-media-type="${mediaType}"]`);
                if (container) {
                    // Keep the cancel button
                    const cancelBtn = container.querySelector('.cancel-media');
                    container.innerHTML = '';
                    container.appendChild(cancelBtn);
                    container.appendChild(mediaElement);
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
                button.textContent = button.id.includes('video') ? 'Record Video' : 'Record Audio';
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