class ConversationRoom {
    constructor(config) {
        this.config = config;
        this.room = null;
        this.participants = new Map();
        this.localParticipant = null;
        this.setupUI();
        this.connect();
    }

    async connect() {
        try {
            // Ensure LiveKitClient is loaded
            if (!window.LiveKitClient) {
                throw new Error('LiveKit client not loaded');
            }

            // Format LiveKit host URL
            const livekitUrl = this.config.livekitHost.startsWith('http') 
                ? this.config.livekitHost 
                : `wss://${this.config.livekitHost}`;

            console.log('Connecting to LiveKit host:', livekitUrl);
            
            // Create room instance
            this.room = new LiveKitClient.Room({
                adaptiveStream: true,
                dynacast: true,
                publishDefaults: {
                    simulcast: true,
                    videoSimulcastLayers: [LiveKitClient.VideoPresets.h90, LiveKitClient.VideoPresets.h216],
                },
            });

            // Connect to room
            await this.room.connect(livekitUrl, this.config.token);
            console.log('Connected to room:', this.config.roomName);
            
            this.localParticipant = this.room.localParticipant;
            this.setupLocalParticipant();
            this.setupRoomHandlers();
            
            // Request permissions and create tracks
            await this.enableLocalTracks();
            
        } catch (error) {
            console.error('Failed to connect to room:', error);
            showError('Failed to connect to the room. Please try again.');
        }
    }

    async enableLocalTracks() {
        try {
            const tracks = await LiveKitClient.createLocalTracks({
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
        
        // Add event listeners
        this.audioBtn.addEventListener('click', () => this.toggleAudio());
        this.videoBtn.addEventListener('click', () => this.toggleVideo());
        this.screenShareBtn.addEventListener('click', () => this.toggleScreenShare());
        
        if (this.config.isCreator) {
            const endBtn = document.getElementById('endRoom');
            endBtn.addEventListener('click', () => this.endRoom());
            
            const shareBtn = document.getElementById('shareButton');
            shareBtn.addEventListener('click', () => this.shareRoom());
        } else {
            const leaveBtn = document.getElementById('leaveRoom');
            leaveBtn.addEventListener('click', () => this.leaveRoom());
        }
    }

    setupLocalParticipant() {
        const tile = this.createParticipantTile(this.localParticipant);
        this.participantsGrid.appendChild(tile);
        this.participants.set(this.localParticipant.identity, tile);
        this.updateParticipantCount();
    }

    setupRoomHandlers() {
        this.room.on(LivekitClient.RoomEvent.ParticipantConnected, (participant) => {
            console.log('Participant connected:', participant.identity);
            const tile = this.createParticipantTile(participant);
            this.participantsGrid.appendChild(tile);
            this.participants.set(participant.identity, tile);
            this.updateParticipantCount();
            this.updateGridLayout();
        });

        this.room.on(LivekitClient.RoomEvent.ParticipantDisconnected, (participant) => {
            console.log('Participant disconnected:', participant.identity);
            const tile = this.participants.get(participant.identity);
            if (tile) {
                tile.remove();
                this.participants.delete(participant.identity);
            }
            this.updateParticipantCount();
            this.updateGridLayout();
        });

        this.room.on(LivekitClient.RoomEvent.TrackSubscribed, (track, publication, participant) => {
            const tile = this.participants.get(participant.identity);
            if (tile) {
                if (track.kind === 'video') {
                    const videoElement = tile.querySelector('video');
                    track.attach(videoElement);
                }
            }
        });

        this.room.on(LivekitClient.RoomEvent.TrackUnsubscribed, (track, publication, participant) => {
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

// Initialize room when the page loads
document.addEventListener('DOMContentLoaded', () => {
    new ConversationRoom(window.ROOM_CONFIG);
}); 