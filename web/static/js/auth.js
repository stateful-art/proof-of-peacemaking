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

// Wallet connection steps
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

            // Get nonce
            const nonceResponse = await fetch('/auth/nonce?address=' + address);
            const { nonce } = await nonceResponse.json();

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

            if (verifyResponse.ok) {
                // Update UI to show signature success
                updateStepStatus('signStep', 'success', 'Verification complete');
                // Reload page after a short delay
                setTimeout(() => window.location.reload(), 1000);
            } else {
                const data = await verifyResponse.json();
                if (data.error === 'signature does not match address') {
                    // If it's just a signature mismatch, retry verification
                    console.log('Retrying verification...');
                    const retryResponse = await fetch('/auth/verify', {
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
                    
                    if (retryResponse.ok) {
                        updateStepStatus('signStep', 'success', 'Verification complete');
                        setTimeout(() => window.location.reload(), 1000);
                        return;
                    }
                }
                updateStepStatus('signStep', 'error', data.error || 'Verification failed');
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
    } else if (status === 'error') {
        spinner.style.display = 'none';
        messageEl.style.color = 'var(--error-color)';
    }

    if (message) {
        messageEl.textContent = message;
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
            await startWalletConnection();
        });
    }
}); 