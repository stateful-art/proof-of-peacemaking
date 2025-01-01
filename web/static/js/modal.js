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
        if (this.form) {
            this.setupEventListeners();
        } else {
            console.error('Expression form not found:', formId); // Debug log
        }
    }

    setupEventListeners() {
        // Record buttons
        const audioRecordBtn = document.getElementById('audioRecord');
        const videoRecordBtn = document.getElementById('videoRecord');
        const imageUploadBtn = document.getElementById('imageUpload');
        const imageInput = document.getElementById('imageInput');

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
                imageInput.click();
            });

            imageInput.addEventListener('change', (e) => {
                this.handleImageUpload(e);
            });
        }

        // Form submission
        this.form.addEventListener('submit', (e) => this.handleSubmit(e));
    }

    async startRecording(mediaType) {
        try {
            const constraints = {
                audio: true,
                video: mediaType === 'video'
            };

            const stream = await navigator.mediaDevices.getUserMedia(constraints);
            this.mediaRecorder = new MediaRecorder(stream);
            
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
                mediaElement.classList.add('media-preview');
                
                document.getElementById('mediaPreview').appendChild(mediaElement);
                this.recordedChunks = [];
            };

            this.mediaRecorder.start();
            this.isRecording = true;
            
            const button = document.getElementById(mediaType + 'Record');
            button.classList.add('recording');
            button.textContent = 'Stop Recording';
            
        } catch (err) {
            document.getElementById('mediaError').textContent = 
                'Error accessing media devices. Please ensure you have given permission.';
            console.error('Error:', err);
        }
    }

    stopRecording() {
        if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
            this.mediaRecorder.stop();
            this.mediaRecorder.stream.getTracks().forEach(track => track.stop());
            this.isRecording = false;
            
            document.querySelectorAll('.btn-record').forEach(button => {
                button.classList.remove('recording');
                button.textContent = button.id.includes('video') ? 'Record Video' : 'Record Audio';
            });
        }
    }

    handleImageUpload(e) {
        const file = e.target.files[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = (e) => {
                const img = document.createElement('img');
                img.src = e.target.result;
                img.classList.add('media-preview');
                document.getElementById('mediaPreview').appendChild(img);
            };
            reader.readAsDataURL(file);
        }
    }

    async handleSubmit(e) {
        e.preventDefault();
        
        const formData = new FormData();
        
        // Add text content
        formData.append('textContent', this.form.querySelector('textarea[name="textContent"]').value);
        
        // Add image if exists
        const imageInput = document.getElementById('imageInput');
        if (imageInput.files.length > 0) {
            formData.append('imageContent', imageInput.files[0]);
        }

        // Add audio/video if exists
        if (this.recordedChunks.length > 0) {
            const mediaBlob = new Blob(this.recordedChunks, {
                type: this.mediaRecorder?.mimeType || 'audio/webm'
            });
            formData.append(this.mediaRecorder?.videoBitsPerSecond ? 'videoContent' : 'audioContent', mediaBlob);
        }
        
        try {
            const response = await fetch('/api/expressions', {
                method: 'POST',
                body: formData,
                // Don't set Content-Type header, let browser set it with boundary
            });

            if (!response.ok) {
                const error = await response.text();
                throw new Error(error || 'Failed to create expression');
            }

            // Clear form and close modal
            this.form.reset();
            document.getElementById('mediaPreview').innerHTML = '';
            this.recordedChunks = [];
            expressionModal.close();

            // Refresh the feed
            location.reload();

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