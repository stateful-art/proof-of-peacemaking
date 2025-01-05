let player;

// Initialize YouTube player
function onYouTubeIframeAPIReady() {
    player = new YT.Player('player', {
        height: '360',
        width: '100%',
        playerVars: {
            'playsinline': 1,
            'controls': 1,
            'autoplay': 0
        }
    });
}

// Extract playlist ID from URL
function getPlaylistId(url) {
    const regex = /[?&]list=([^#\&\?]+)/;
    const match = url.match(regex);
    return match ? match[1] : null;
}

// Load playlist
function loadPlaylist() {
    const urlInput = document.getElementById('playlistUrl');
    const url = urlInput.value.trim();
    const playlistId = getPlaylistId(url);

    if (!playlistId) {
        alert('Please enter a valid YouTube playlist URL');
        return;
    }

    // Show loading state
    document.querySelector('.loading').classList.add('active');

    try {
        // Load the playlist in the player
        player.loadPlaylist({
            list: playlistId,
            listType: 'playlist'
        });

        // Show player section
        document.getElementById('playerSection').classList.add('active');
        
    } catch (error) {
        console.error('Error loading playlist:', error);
        alert('Failed to load playlist. Please try again.');
    } finally {
        document.querySelector('.loading').classList.remove('active');
    }
} 