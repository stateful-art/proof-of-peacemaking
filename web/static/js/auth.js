// Make openAuthModal available globally
window.openAuthModal = openAuthModal;

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
    const authForms = document.getElementById('authForms');
    const walletSteps = document.getElementById('walletSteps');
    
    console.log('Auth modal element:', modal); // Debug log
    
    if (modal) {
        // Reset modal state
        walletSteps.style.display = 'none';
        authForms.style.display = 'block';
        authForms.classList.add('active');
        modal.classList.add('active');
        
        console.log('Added active class to modal'); // Debug log
        // Show login form by default
        switchAuthMode('login');
    } else {
        console.error('Auth modal element not found!'); // Debug log
    }
}

// Make closeAuthModal available globally
window.closeAuthModal = closeAuthModal;

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
    console.log('Login form submitted');
    
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;

    console.log('Login form data:', { email }); // Don't log password

    if (!email || !password) {
        alert('Please fill in all fields');
        return;
    }

    try {
        console.log('Sending login request...');
        const response = await fetch('/auth/login-email', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
            credentials: 'include'
        });

        console.log('Login response status:', response.status);
        
        if (response.ok) {
            console.log('Login successful, reloading page...');
            window.location.reload();
        } else {
            const data = await response.json();
            console.error('Login failed:', data.error);
            alert(data.error || 'Login failed');
        }
    } catch (error) {
        console.error('Login error:', error);
        alert('Login failed. Please try again.');
    }
}

async function handleEmailRegister(event) {
    event.preventDefault();
    console.log('Register form submitted');
    
    const email = document.getElementById('registerEmail').value;
    const password = document.getElementById('registerPassword').value;
    const username = document.getElementById('registerUsername').value;
    const errorDiv = document.getElementById('registerError');

    // Clear any previous error
    errorDiv.style.display = 'none';
    errorDiv.textContent = '';

    console.log('Register form data:', { email, username }); // Don't log password

    if (!email || !password || !username) {
        errorDiv.textContent = 'Please fill in all fields';
        errorDiv.style.display = 'block';
        return;
    }

    try {
        console.log('Sending registration request...');
        const response = await fetch('/auth/register-email', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                email, 
                password,
                username 
            }),
            credentials: 'include'
        });

        console.log('Registration response status:', response.status);
        
        if (response.ok) {
            console.log('Registration successful, reloading page...');
            window.location.reload();
        } else {
            const data = await response.json();
            console.error('Registration failed:', data.error);
            // Show error in form
            if (data.error === 'Email already registered' || data.error === 'Username already taken') {
                errorDiv.textContent = 'Username or email address already registered';
            } else {
                errorDiv.textContent = data.error || 'Registration failed';
            }
            errorDiv.style.display = 'block';
        }
    } catch (error) {
        console.error('Registration error:', error);
        errorDiv.textContent = 'Registration failed. Please try again.';
        errorDiv.style.display = 'block';
    }
}

// Wallet connection steps
let nonceRequestTimeout = null;
let nonceRequestInProgress = false;
let currentNonce = null;

async function getNonceWithDebounce(address) {
    if (nonceRequestInProgress) {
        console.log('Nonce request already in progress');
        return null;
    }

    // Clear any existing timeout
    if (nonceRequestTimeout) {
        clearTimeout(nonceRequestTimeout);
    }

    // Return a promise that resolves with the nonce
    return new Promise((resolve, reject) => {
        nonceRequestTimeout = setTimeout(async () => {
            try {
                nonceRequestInProgress = true;
                const nonceResponse = await fetch('/auth/nonce?address=' + address);
                if (!nonceResponse.ok) {
                    throw new Error('Failed to get nonce');
                }
                const { nonce } = await nonceResponse.json();
                currentNonce = nonce;
                resolve(nonce);
            } catch (error) {
                console.error('Error getting nonce:', error);
                reject(error);
            } finally {
                nonceRequestInProgress = false;
                nonceRequestTimeout = null;
            }
        }, 100); // 100ms delay to debounce
    });
}

async function startWalletConnection() {
    // Show wallet steps UI
    const modal = document.getElementById('authModal');
    const authForms = document.getElementById('authForms');
    const walletSteps = document.getElementById('walletSteps');
    
    // Show modal if not already shown
    if (!modal.classList.contains('active')) {
        modal.classList.add('active');
    }
    
    authForms.style.display = 'none';
    walletSteps.style.display = 'block';
    walletSteps.classList.add('active');

    // Activate first step
    const connectStep = document.getElementById('connectStep');
    const signStep = document.getElementById('signStep');
    connectStep.classList.add('active');

    try {
        // Check if MetaMask is installed
        if (typeof window.ethereum === 'undefined') {
            updateStepStatus('connectStep', 'error', 'Please install MetaMask to continue');
            return;
        }

        // Update navbar status
        const walletBtn = document.getElementById('walletBtn');
        if (walletBtn) {
            walletBtn.textContent = 'Connecting...';
        }

        try {
            // Request account access
            const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
            const address = accounts[0];

            // Update UI to show connection success
            updateStepStatus('connectStep', 'success', 'Wallet connected successfully');
            
            // Update navbar status
            if (walletBtn) {
                walletBtn.textContent = 'Confirming...';
            }

            // Activate sign step
            signStep.classList.add('active');
            const signSpinner = signStep.querySelector('.spinner');
            signSpinner.style.display = 'block';

            // Get nonce with debounce
            try {
                const nonce = await getNonceWithDebounce(address);
                if (!nonce) {
                    throw new Error('Failed to get nonce');
                }

                // Create message to sign
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
                    credentials: 'include'
                });

                if (!verifyResponse.ok) {
                    const data = await verifyResponse.json();
                    throw new Error(data.error || 'Verification failed');
                }

                // Handle successful verification
                const verifyData = await verifyResponse.json();
                updateStepStatus('signStep', 'success', 'Verification complete');
                
                // Redirect or reload after a short delay
                setTimeout(() => {
                    if (verifyData.redirect) {
                        window.location.href = verifyData.redirect;
                    } else {
                        window.location.reload();
                    }
                }, 1000);

            } catch (error) {
                console.error('Error in wallet connection flow:', error);
                updateStepStatus('signStep', 'error', error.message || 'Connection failed');
                if (walletBtn) {
                    walletBtn.textContent = 'Enter';
                }
            }

        } catch (error) {
            console.error('Wallet connection error:', error);
            // Reset navbar status
            if (walletBtn) {
                walletBtn.textContent = 'Enter';
            }
            
            if (error.code === 4001) {
                // User rejected request
                updateStepStatus('connectStep', 'error', 'Connection rejected by user');
            } else if (error.code === -32002) {
                // Request already pending
                updateStepStatus('connectStep', 'error', 'Wallet connection already pending. Please check MetaMask.');
            } else {
                // Other errors
                const activeStep = signStep.classList.contains('active') ? 'signStep' : 'connectStep';
                updateStepStatus(activeStep, 'error', 'Failed to connect wallet. Please try again.');
            }
        }
    } catch (error) {
        console.error('Outer wallet connection error:', error);
        // Reset navbar status
        const walletBtn = document.getElementById('walletBtn');
        if (walletBtn) {
            walletBtn.textContent = 'Enter';
        }
    } finally {
        // Always clean up state
        currentNonce = null;
        if (nonceRequestTimeout) {
            clearTimeout(nonceRequestTimeout);
            nonceRequestTimeout = null;
        }
        nonceRequestInProgress = false;
    }
}

function updateStepStatus(stepId, status, message) {
    const step = document.getElementById(stepId);
    const spinner = step.querySelector('.spinner');
    const checkIcon = step.querySelector('.check-icon');
    const messageEl = step.querySelector('p');

    if (status === 'success') {
        spinner.style.display = 'none';
        checkIcon.style.display = 'flex';
        step.classList.add('success');
        messageEl.style.color = '#90EE90'; // Light pistachio green
    } else if (status === 'error') {
        spinner.style.display = 'none';
        messageEl.style.color = '#FFFFFF'; // White for error messages
    }

    if (message) {
        messageEl.textContent = message;
    }
}

// Add at the top of the file
let isMetaMaskConnected = false;

// Add this new function
async function handleMetaMaskDisconnect() {
    console.log('MetaMask disconnected');
    isMetaMaskConnected = false;
    // Clear any pending requests
    if (nonceRequestTimeout) {
        clearTimeout(nonceRequestTimeout);
        nonceRequestTimeout = null;
    }
    nonceRequestInProgress = false;
    currentNonce = null;
    
    // Call our logout endpoint to clean up server-side state
    try {
        await fetch('/auth/logout', {
            method: 'POST',
            credentials: 'include'
        });
    } catch (error) {
        console.error('Error logging out:', error);
    }
    
    // Reload the page to reset UI state
    window.location.reload();
}

// Event listeners
document.addEventListener('DOMContentLoaded', () => {
    console.log('Auth.js DOMContentLoaded event fired');
    
    // Add MetaMask account change listener
    if (window.ethereum) {
        window.ethereum.on('accountsChanged', async (accounts) => {
            if (accounts.length === 0) {
                // User disconnected from MetaMask
                await handleMetaMaskDisconnect();
            }
        });

        // Check initial connection state
        window.ethereum.request({ method: 'eth_accounts' })
            .then(accounts => {
                isMetaMaskConnected = accounts.length > 0;
            })
            .catch(console.error);
    }

    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');
    const metamaskButton = document.querySelector('.btn-metamask');
    const enterButton = document.getElementById('connectWallet');

    console.log('Enter button found:', enterButton);

    // Initialize the Enter button if it exists
    if (enterButton) {
        console.log('Setting up Enter button click handler');
        // Remove any existing click handlers
        const newEnterButton = enterButton.cloneNode(true);
        enterButton.parentNode.replaceChild(newEnterButton, enterButton);
        
        // Add new click handler
        newEnterButton.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            console.log('Enter button clicked');
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
            await startWalletConnection();
        });
    }
}); 