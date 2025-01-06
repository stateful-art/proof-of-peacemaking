// Wait for LiveKit to load
function waitForLiveKit() {
    return new Promise((resolve, reject) => {
        if (window.livekit) {
            resolve(window.livekit);
            return;
        }

        let attempts = 0;
        const maxAttempts = 10;
        const interval = setInterval(() => {
            attempts++;
            if (window.livekit) {
                clearInterval(interval);
                resolve(window.livekit);
            } else if (attempts >= maxAttempts) {
                clearInterval(interval);
                reject(new Error('LiveKit client failed to load'));
            }
        }, 500);
    });
}

class ConversationRoom {
    constructor(config) {
        if (!config) {
            throw new Error('Configuration is required');
        }
        this.config = config;
        this.room = null;
        this.participants = new Map();
        this.localParticipant = null;
        this.initialized = false;
        this.setupUI();
        this.initialize();
    }

    async initialize() {
        try {
            // Wait for LiveKit to load
            const livekit = await waitForLiveKit();
            // Connect to room
            await this.connect(livekit);
        } catch (error) {
            console.error('Failed to initialize:', error);
            showError('Failed to initialize the room. Please try again.');
        }
    }

    async connect(livekit) {
        try {
            // Format LiveKit host URL
            const livekitUrl = this.config.livekitHost.replace(/^https?:\/\//, '').replace(/^wss?:\/\//, '');
            const wsUrl = `wss://${livekitUrl}`;
            
            console.log('Connecting to LiveKit host:', wsUrl);
            console.log('Room name:', this.config.roomName);
            console.log('Identity:', this.config.userId);
            
            // Create room instance with identity
            this.room = new livekit.Room({
                adaptiveStream: true,
                dynacast: true,
                name: this.config.roomName,
                identity: this.config.userId,
                publishDefaults: {
                    simulcast: true,
                    videoSimulcastLayers: [livekit.VideoPresets.h90, livekit.VideoPresets.h216],
                }
            });

            // Connect to room with identity in connect options
            await this.room.connect(wsUrl, this.config.token, {
                autoSubscribe: true
            });
            
            console.log('Connected to room:', this.config.roomName);
            
            this.localParticipant = this.room.localParticipant;
            this.setupLocalParticipant();
            this.setupRoomHandlers(livekit);
            
            // Request permissions and create tracks
            await this.enableLocalTracks(livekit);
            
        } catch (error) {
            console.error('Failed to connect to room:', error);
            showError('Failed to connect to the room. Please try again.');
        }
    }

    async enableLocalTracks(livekit) {
        try {
            const tracks = await livekit.createLocalTracks({
                audio: true,
                video: true
            });
            
            await Promise.all(tracks.map(track => this.localParticipant.publishTrack(track)));
            console.log('Local tracks enabled:', tracks);
            
            // Update UI to reflect enabled state
            this.audioBtn.classList.add('active');
            this.videoBtn.classList.add('active');
        } catch (error) {
            console.error('Failed to enable local tracks:', error);
            showError('Failed to enable camera and microphone. Please check permissions.');
        }
    }

    setupUI() {
        // Setup control buttons
        this.audioBtn = document.getElementById('toggleAudio');
        this.videoBtn = document.getElementById('toggleVideo');
        this.screenShareBtn = document.getElementById('toggleScreenShare');
        this.participantCount = document.getElementById('participantCount');
        this.participantsGrid = document.getElementById('participantsGrid');
        
        if (!this.audioBtn || !this.videoBtn || !this.screenShareBtn || !this.participantCount || !this.participantsGrid) {
            throw new Error('Required UI elements not found');
        }

        // Add event listeners
        this.audioBtn.addEventListener('click', async () => {
            if (this.localParticipant) {
                await this.toggleAudio();
            }
        });
        
        this.videoBtn.addEventListener('click', async () => {
            if (this.localParticipant) {
                await this.toggleVideo();
            }
        });
        
        this.screenShareBtn.addEventListener('click', async () => {
            if (this.localParticipant) {
                await this.toggleScreenShare();
            }
        });
        
        if (this.config.isCreator) {
            const endBtn = document.getElementById('endRoom');
            const shareBtn = document.getElementById('shareButton');
            
            if (!endBtn || !shareBtn) {
                throw new Error('Creator UI elements not found');
            }
            
            endBtn.addEventListener('click', () => this.endRoom());
            shareBtn.addEventListener('click', () => this.shareRoom());
        } else {
            const leaveBtn = document.getElementById('leaveRoom');
            if (!leaveBtn) {
                throw new Error('Leave button not found');
            }
            leaveBtn.addEventListener('click', () => this.leaveRoom());
        }
    }

    setupLocalParticipant() {
        const tile = this.createParticipantTile(this.localParticipant);
        this.participantsGrid.appendChild(tile);
        this.participants.set(this.localParticipant.identity, tile);
        this.updateParticipantCount();
    }

    setupRoomHandlers(livekit) {
        this.room.on(livekit.RoomEvent.ParticipantConnected, (participant) => {
            console.log('Participant connected:', participant.identity);
            const tile = this.createParticipantTile(participant);
            this.participantsGrid.appendChild(tile);
            this.participants.set(participant.identity, tile);
            this.updateParticipantCount();
            this.updateGridLayout();
        });

        this.room.on(livekit.RoomEvent.ParticipantDisconnected, (participant) => {
            console.log('Participant disconnected:', participant.identity);
            const tile = this.participants.get(participant.identity);
            if (tile) {
                tile.remove();
                this.participants.delete(participant.identity);
            }
            this.updateParticipantCount();
            this.updateGridLayout();
        });

        this.room.on(livekit.RoomEvent.TrackSubscribed, (track, publication, participant) => {
            const tile = this.participants.get(participant.identity);
            if (tile) {
                if (track.kind === 'video') {
                    const videoElement = tile.querySelector('video');
                    track.attach(videoElement);
                }
            }
        });

        this.room.on(livekit.RoomEvent.TrackUnsubscribed, (track, publication, participant) => {
            track.detach();
        });
    }

    createParticipantTile(participant) {
        const tile = document.createElement('div');
        tile.className = 'participant-tile';
        
        const video = document.createElement('video');
        video.autoplay = true;
        video.playsInline = true;
        tile.appendChild(video);
        
        const info = document.createElement('div');
        info.className = 'participant-info';
        
        const name = document.createElement('span');
        name.className = 'participant-name';
        name.textContent = participant.identity;
        info.appendChild(name);
        
        const controls = document.createElement('div');
        controls.className = 'participant-controls';
        
        const audioIcon = document.createElement('i');
        audioIcon.className = 'fas fa-microphone';
        controls.appendChild(audioIcon);
        
        const videoIcon = document.createElement('i');
        videoIcon.className = 'fas fa-video';
        controls.appendChild(videoIcon);
        
        info.appendChild(controls);
        tile.appendChild(info);
        
        return tile;
    }

    updateGridLayout() {
        const count = this.participants.size;
        this.participantsGrid.className = 'participants-grid';
        
        if (count === 1) {
            this.participantsGrid.classList.add('grid-1');
        } else if (count === 2) {
            this.participantsGrid.classList.add('grid-2');
        } else if (count <= 4) {
            this.participantsGrid.classList.add('grid-3-4');
        } else if (count <= 6) {
            this.participantsGrid.classList.add('grid-5-6');
        } else {
            this.participantsGrid.classList.add('grid-7-9');
        }
    }

    updateParticipantCount() {
        this.participantCount.textContent = this.participants.size;
    }

    async toggleAudio() {
        const enabled = this.localParticipant.isMicrophoneEnabled;
        await this.localParticipant.setMicrophoneEnabled(!enabled);
        this.audioBtn.classList.toggle('muted', !this.localParticipant.isMicrophoneEnabled);
        this.audioBtn.querySelector('i').className = this.localParticipant.isMicrophoneEnabled ? 
            'fas fa-microphone' : 'fas fa-microphone-slash';
    }

    async toggleVideo() {
        const enabled = this.localParticipant.isCameraEnabled;
        await this.localParticipant.setCameraEnabled(!enabled);
        this.videoBtn.classList.toggle('muted', !this.localParticipant.isCameraEnabled);
        this.videoBtn.querySelector('i').className = this.localParticipant.isCameraEnabled ? 
            'fas fa-video' : 'fas fa-video-slash';
    }

    async toggleScreenShare() {
        const enabled = this.localParticipant.isScreenShareEnabled;
        await this.localParticipant.setScreenShareEnabled(!enabled);
        this.screenShareBtn.classList.toggle('active', this.localParticipant.isScreenShareEnabled);
    }

    async endRoom() {
        if (confirm('Are you sure you want to end the room for all participants?')) {
            try {
                await fetch(`/api/conversations/${this.config.conversationId}/end`, {
                    method: 'POST',
                    credentials: 'include'
                });
                window.location.href = '/conversations';
            } catch (error) {
                console.error('Failed to end room:', error);
                showError('Failed to end the room. Please try again.');
            }
        }
    }

    async leaveRoom() {
        await this.room.disconnect();
        window.location.href = '/conversations';
    }

    async shareRoom() {
        const shareUrl = `${window.location.origin}/conversations/${this.config.conversationId}`;
        try {
            await navigator.clipboard.writeText(shareUrl);
            showSuccess('Room link copied to clipboard!');
        } catch (err) {
            console.error('Failed to copy:', err);
            // Fallback
            const input = document.createElement('input');
            input.value = shareUrl;
            document.body.appendChild(input);
            input.select();
            document.execCommand('copy');
            document.body.removeChild(input);
            showSuccess('Room link copied to clipboard!');
        }
    }
}

function showSuccess(message) {
    const toast = document.createElement('div');
    toast.className = 'toast success';
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => {
        toast.remove();
    }, 3000);
}

function showError(message) {
    const toast = document.createElement('div');
    toast.className = 'toast error';
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => {
        toast.remove();
    }, 3000);
}

// Initialize when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    if (!window.LIVEKIT_CONFIG) {
        console.error('LiveKit configuration not found');
        showError('Failed to initialize the room. Please try again.');
        return;
    }

    try {
        new ConversationRoom(window.LIVEKIT_CONFIG);
    } catch (error) {
        console.error('Failed to initialize room:', error);
        showError('Failed to initialize the room. Please try again.');
    }
}); 