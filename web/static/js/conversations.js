// LiveKit client setup
let room;
let localParticipant;

// Wait for LiveKit script to load
function waitForLiveKit() {
    return new Promise((resolve, reject) => {
        if (window.LiveKitClient) {
            resolve(window.LiveKitClient);
        } else {
            const script = document.querySelector('script[src*="livekit-client"]');
            script.addEventListener('load', () => {
                if (window.LiveKitClient) {
                    resolve(window.LiveKitClient);
                } else {
                    reject(new Error('LiveKit failed to load'));
                }
            });
            script.addEventListener('error', () => reject(new Error('Failed to load LiveKit script')));
        }
    });
}

async function initializeLiveKit(token) {
    try {
        const LiveKitClient = await waitForLiveKit();
        
        room = new LiveKitClient.Room({
            // Video codec preference
            adaptiveStream: true,
            dynacast: true,
            publishDefaults: {
                simulcast: true,
                videoSimulcastLayers: [LiveKitClient.VideoPresets.h90, LiveKitClient.VideoPresets.h216],
            },
        });

        room.on(LiveKitClient.RoomEvent.ParticipantConnected, handleParticipantConnected);
        room.on(LiveKitClient.RoomEvent.ParticipantDisconnected, handleParticipantDisconnected);
        room.on(LiveKitClient.RoomEvent.TrackSubscribed, handleTrackSubscribed);
        room.on(LiveKitClient.RoomEvent.TrackUnsubscribed, handleTrackUnsubscribed);

        await room.connect(window.livekitHost, token);
        localParticipant = room.localParticipant;

        console.log('Connected to LiveKit room:', room.name);
    } catch (error) {
        console.error('Failed to connect to LiveKit room:', error);
        throw error;
    }
}

// Participant event handlers
function handleParticipantConnected(participant) {
    console.log('Participant connected:', participant.identity);
    updateParticipantCount();
}

function handleParticipantDisconnected(participant) {
    console.log('Participant disconnected:', participant.identity);
    updateParticipantCount();
}

function handleTrackSubscribed(track, publication, participant) {
    if (track.kind === 'video') {
        const element = track.attach();
        element.classList.add('participant-video');
        document.getElementById('participants').appendChild(element);
    } else if (track.kind === 'audio') {
        const element = track.attach();
        element.style.display = 'none';
        document.body.appendChild(element);
    }
}

function handleTrackUnsubscribed(track, publication, participant) {
    track.detach().forEach(element => element.remove());
}

// UI event handlers
async function joinConversation(conversationId) {
    try {
        const response = await fetch(`/api/conversations/${conversationId}/join`, {
            method: 'POST',
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to get join token');
        }
        
        const { token } = await response.json();
        await initializeLiveKit(token);
        
        // Enable local tracks based on user preferences
        await enableLocalTracks();
        
    } catch (error) {
        console.error('Failed to join conversation:', error);
        alert('Failed to join conversation. Please try again.');
    }
}

async function enableLocalTracks() {
    try {
        // Request permissions and create tracks
        const tracks = await LiveKit.createLocalTracks({
            audio: true,
            video: true
        });
        
        // Publish tracks
        await Promise.all(tracks.map(track => localParticipant.publishTrack(track)));
        
    } catch (error) {
        console.error('Failed to enable local tracks:', error);
    }
}

async function createConversation(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const data = {
        title: formData.get('title'),
        description: formData.get('description'),
        imageUrl: formData.get('imageUrl'),
        startTime: new Date(formData.get('startTime')).toISOString(),
        tags: formData.get('tags').split(',').map(tag => tag.trim())
    };
    
    try {
        const response = await fetch('/api/conversations', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data),
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to create conversation');
        }
        
        const conversation = await response.json();
        // Refresh the conversations list
        await loadConversations();
        
        // Close the create form modal if you have one
        closeCreateForm();
        
    } catch (error) {
        console.error('Failed to create conversation:', error);
        alert('Failed to create conversation. Please try again.');
    }
}

async function loadConversations(filter = '') {
    try {
        const response = await fetch(`/api/conversations?${filter}`, {
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to load conversations');
        }
        
        const conversations = await response.json();
        renderConversations(conversations);
        
    } catch (error) {
        console.error('Failed to load conversations:', error);
    }
}

function updateParticipantCount() {
    const count = room.participants.size;
    document.getElementById('participantCount').textContent = count;
}

function renderConversations(conversations) {
    const list = document.getElementById('conversationsList');
    const topCreateButton = document.getElementById('topCreateButton');
    list.innerHTML = '';

    console.log('Current user ID:', window.userID);

    if (!conversations || conversations.length === 0) {
        // Show empty state and hide top button
        topCreateButton.style.display = 'none';
        list.innerHTML = `
            <div class="empty-state">
                <h2>No conversations yet</h2>
                <p>Be the first to start a conversation and connect with others!</p>
                <button class="create-btn" onclick="openCreateForm()">
                    <i class="fas fa-plus"></i>
                    Start a Conversation
                </button>
            </div>
        `;
        return;
    }

    // Show top button when there are conversations
    topCreateButton.style.display = 'block';

    conversations.forEach(conversation => {
        console.log('Conversation:', conversation);
        console.log('Creator ID:', conversation.creatorId);
        console.log('Current user ID:', window.userID);
        console.log('Is creator?', conversation.creatorId === window.userID);

        const card = document.createElement('div');
        card.className = 'conversation-card';
        
        const statusClass = conversation.status === 'live' ? 'live' : '';
        const formattedDate = new Date(conversation.startTime).toLocaleString();
        
        // Determine which button to show based on user role and conversation status
        let actionButton = '';
        const isCreator = conversation.creatorId === window.userID;
        
        if (isCreator) {
            if (conversation.status === 'scheduled') {
                actionButton = `
                    <button class="action-btn go-live-btn" onclick="goLive('${conversation.id}')">
                        <i class="fas fa-broadcast-tower"></i> Go Live
                    </button>`;
            } else if (conversation.status === 'live') {
                actionButton = `
                    <button class="action-btn enter-btn" onclick="window.location.href='/conversations/${conversation.id}'">
                        <i class="fas fa-door-open"></i> Enter Room
                    </button>
                    <button class="action-btn end-btn" onclick="endConversation('${conversation.id}')">
                        <i class="fas fa-stop"></i> End
                    </button>`;
            }
        } else if (conversation.status === 'live') {
            actionButton = `
                <button class="action-btn join-btn" onclick="window.location.href='/conversations/${conversation.id}'">
                    <i class="fas fa-sign-in-alt"></i> Join
                </button>`;
        }
        
        card.innerHTML = `
            <img class="conversation-image" src="${conversation.imageUrl || '/static/img/default-conversation.jpg'}" alt="${conversation.title}">
            <div class="conversation-content">
                <div class="conversation-header">
                    <h3 class="conversation-title">${conversation.title}</h3>
                    <span class="conversation-status ${statusClass}">${conversation.status}</span>
                </div>
                <div class="conversation-meta">
                    Created by ${conversation.creatorId} • Starts ${formattedDate}
                </div>
                <p class="conversation-description">${conversation.description}</p>
                <div class="conversation-tags">
                    ${conversation.tags.map(tag => `<span class="conversation-tag">${tag}</span>`).join('')}
                </div>
                <div class="conversation-actions">
                    ${actionButton}
                </div>
            </div>
        `;
        
        list.appendChild(card);
    });
}

async function goLive(conversationId) {
    try {
        // Start the conversation
        const response = await fetch(`/api/conversations/${conversationId}/start`, {
            method: 'POST',
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Failed to start conversation');
        }

        // Get join token
        const tokenResponse = await fetch(`/api/conversations/${conversationId}/join`, {
            method: 'POST',
            credentials: 'include'
        });

        if (!tokenResponse.ok) {
            throw new Error('Failed to get join token');
        }

        const { token } = await tokenResponse.json();

        // Initialize LiveKit with the token
        await initializeLiveKit(token);
        
        // Show the video container
        document.getElementById('videoContainer').style.display = 'block';
        
        // Enable local tracks
        await enableLocalTracks();
        
        // Refresh the conversations list
        await loadConversations();

    } catch (error) {
        console.error('Failed to go live:', error);
        showError('Failed to start the conversation. Please try again.');
    }
}

async function shareConversation(conversationId) {
    const shareUrl = `${window.location.origin}/conversations/${conversationId}`;
    
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

function openCreateForm() {
    const modal = document.getElementById('createFormModal');
    modal.style.display = 'block';
}

function closeCreateForm() {
    const modal = document.getElementById('createFormModal');
    modal.style.display = 'none';
}

// Close modal when clicking outside
window.onclick = function(event) {
    const modal = document.getElementById('createFormModal');
    if (event.target == modal) {
        modal.style.display = 'none';
    }
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    // Remove duplicate form handler since we already have it in the HTML
    const createForm = document.getElementById('createConversationForm');
    if (createForm) {
        createForm.removeEventListener('submit', createConversation);
    }
    
    // Load initial conversations
    loadConversations();
    
    // Set up filter handlers
    const filters = document.querySelectorAll('.filter-tag');
    filters.forEach(filter => {
        filter.addEventListener('click', () => {
            // Remove active class from all filters
            filters.forEach(f => f.classList.remove('active'));
            // Add active class to clicked filter
            filter.classList.add('active');
            
            const status = filter.textContent.toLowerCase();
            loadConversations(status === 'all' ? '' : `status=${status}`);
        });
    });
}); 