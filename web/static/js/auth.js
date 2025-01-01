// Auth Modal Control
async function openAuthModal() {
    console.log('openAuthModal called'); // Debug log
    
    // First check if user is already connected
    try {
        const response = await fetch('/auth/session');
        const data = await response.json();
        
        if (data.authenticated) {
            console.log('User already authenticated, not showing modal'); // Debug log
            return;
        }
    } catch (error) {
        console.error('Error checking session:', error);
    }

    const modal = document.getElementById('authModal');
    console.log('Auth modal element:', modal); // Debug log
    
    if (modal) {
        modal.classList.add('active');
        console.log('Added active class to modal'); // Debug log
        // Show login form by default
        switchAuthMode('login');
    } else {
        console.error('Auth modal element not found!'); // Debug log
    }
}

function closeAuthModal() {
    const modal = document.getElementById('authModal');
    if (modal) {
        modal.classList.remove('active');
    }
}

// Close modal when clicking outside
window.onclick = function(event) {
    const modal = document.getElementById('authModal');
    if (event.target === modal) {
        closeAuthModal();
    }
}

// Switch between login and register forms
function switchAuthMode(mode) {
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');
    const loginTab = document.querySelector('[onclick="switchAuthMode(\'login\')"]');
    const registerTab = document.querySelector('[onclick="switchAuthMode(\'register\')"]');

    if (mode === 'login') {
        loginForm.classList.add('active');
        registerForm.classList.remove('active');
        loginTab.classList.add('active');
        registerTab.classList.remove('active');
    } else {
        loginForm.classList.remove('active');
        registerForm.classList.add('active');
        loginTab.classList.remove('active');
        registerTab.classList.add('active');
    }
}

// Form submission handlers
async function handleEmailLogin(event) {
    event.preventDefault();
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;

    try {
        const response = await fetch('/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });

        if (response.ok) {
            window.location.reload();
        } else {
            const data = await response.json();
            alert(data.error || 'Login failed');
        }
    } catch (error) {
        console.error('Login error:', error);
        alert('Login failed. Please try again.');
    }
}

async function handleEmailRegister(event) {
    event.preventDefault();
    const email = document.getElementById('registerEmail').value;
    const password = document.getElementById('registerPassword').value;

    try {
        const response = await fetch('/auth/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });

        if (response.ok) {
            window.location.reload();
        } else {
            const data = await response.json();
            alert(data.error || 'Registration failed');
        }
    } catch (error) {
        console.error('Registration error:', error);
        alert('Registration failed. Please try again.');
    }
}

// Wallet connection
async function connectWallet() {
    if (typeof window.ethereum === 'undefined') {
        alert('Please install MetaMask to use this feature');
        return;
    }

    try {
        // Request account access
        const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
        const address = accounts[0];

        // Get nonce
        const nonceResponse = await fetch('/auth/nonce?address=' + address);
        const { nonce } = await nonceResponse.json();

        // Create message to sign - EXACT match with backend
        const message = `Sign this message to verify your wallet. Nonce: ${nonce}`;

        // Request signature
        const signature = await window.ethereum.request({
            method: 'personal_sign',
            params: [message, address],
        });

        // Verify signature
        const verifyResponse = await fetch('/auth/verify', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                address,
                signature
            }),
            credentials: 'include'  // Important: include credentials
        });

        if (verifyResponse.ok) {
            window.location.reload();
        } else {
            const data = await verifyResponse.json();
            alert(data.error || 'Wallet verification failed');
        }
    } catch (error) {
        console.error('Wallet connection error:', error);
        alert('Failed to connect wallet. Please try again.');
    }
}

// Event listeners
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOMContentLoaded event fired'); // Debug log
    
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');
    const metamaskButton = document.querySelector('.btn-metamask');
    const enterButton = document.getElementById('connectWallet');

    console.log('Enter button found:', enterButton); // Debug log

    // Initialize the Enter button if it exists
    if (enterButton) {
        console.log('Setting up Enter button click handler'); // Debug log
        enterButton.addEventListener('click', (e) => {
            e.preventDefault();
            console.log('Enter button clicked'); // Debug log
            openAuthModal();
        });
    }

    if (loginForm) {
        loginForm.addEventListener('submit', handleEmailLogin);
    }

    if (registerForm) {
        registerForm.addEventListener('submit', handleEmailRegister);
    }

    if (metamaskButton) {
        metamaskButton.addEventListener('click', async () => {
            // Close the auth modal before starting wallet connection
            closeAuthModal();
            // Start wallet connection
            await connectWallet();
        });
    }
}); 