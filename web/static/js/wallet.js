// Wait for ethers to be available
function waitForEthers() {
    return new Promise((resolve) => {
        const checkEthers = () => {
            if (window.ethers) {
                resolve();
            } else {
                setTimeout(checkEthers, 100);
            }
        };
        checkEthers();
    });
}

let provider;
let signer;
let isConnected = false;
let currentAddress = null;

// Check if user is already authenticated
async function checkSession() {
    try {
        const response = await fetch('/auth/session', {
            method: 'GET',
            credentials: 'include'
        });
        
        const data = await response.json();
        console.log('[AUTH] Session check response:', data);
        
        if (data.authenticated && data.address) {
            isConnected = true;
            currentAddress = data.address;
            
            // Update UI with user icon and dropdown
            const enterButton = document.getElementById('connectWallet');
            if (enterButton) {
                updateWalletButton(data.address);
            }
            
            // Show nav items
            const navAuthItems = document.querySelector('.nav-auth-items');
            if (navAuthItems) {
                navAuthItems.classList.add('visible');
            }

            // Hide auth modal if it's open
            const authModal = document.getElementById('authModal');
            if (authModal && authModal.classList.contains('active')) {
                authModal.classList.remove('active');
            }
            
            return true;
        }
        
        // If not authenticated, ensure UI shows Enter button
        const enterButton = document.getElementById('connectWallet');
        if (enterButton) {
            enterButton.innerHTML = 'Enter';
            enterButton.className = 'action-button';
            // Don't set onclick here, it will be handled by auth.js
        }
        
        return false;
    } catch (error) {
        console.error('Error checking session:', error);
        return false;
    }
}

// Update the connectWallet function
async function connectWallet() {
    try {
        // Check if already connected
        if (isMetaMaskConnected) {
            updateStepStatus('error', 'Wallet connection already pending. Please check MetaMask.', '#FFFFFF');
            return;
        }

        // Check if MetaMask is installed
        if (!window.ethereum) {
            updateStepStatus('error', 'Please install MetaMask to continue.', '#FFFFFF');
            return;
        }

        // Request account access
        const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
        if (accounts.length === 0) {
            updateStepStatus('error', 'No accounts found. Please connect your wallet.', '#FFFFFF');
            return;
        }

        // Set connection state
        isMetaMaskConnected = true;
        
        // Start the authentication process
        await startWalletConnection();
        
        updateStepStatus('success', 'Wallet connected successfully', '#90EE90');
    } catch (error) {
        console.error('Error connecting wallet:', error);
        isMetaMaskConnected = false;
        updateStepStatus('error', error.message || 'Failed to connect wallet', '#FFFFFF');
    }
}

// Add disconnect function
async function disconnectWallet() {
    try {
        await handleMetaMaskDisconnect();
        updateStepStatus('success', 'Wallet disconnected successfully', '#90EE90');
    } catch (error) {
        console.error('Error disconnecting wallet:', error);
        updateStepStatus('error', error.message || 'Failed to disconnect wallet', '#FFFFFF');
    }
}

// Add function to update wallet button
function updateWalletButton(address) {
    console.log('[AUTH] Updating wallet button for address:', address);
    const walletButton = document.getElementById('connectWallet');
    if (!walletButton) {
        console.log('[AUTH] Wallet button not found');
        return;
    }

    if (!address) {
        // If no address, set up the Enter button
        walletButton.innerHTML = 'Enter';
        walletButton.className = 'action-button';
        walletButton.onclick = openAuthModal;
        return;
    }

    const shortAddress = `${address.slice(0, 6)}...${address.slice(-4)}`;
    console.log('[AUTH] Creating user dropdown with address:', shortAddress);

    // Replace the button with a user dropdown
    const userDropdown = document.createElement('div');
    userDropdown.className = 'user-dropdown';
    userDropdown.innerHTML = `
        <button class="user-icon" id="userIcon">
            <img src="/static/img/user.png" alt="User" width="24" height="24">
        </button>
        <div class="dropdown-content">
            <div class="wallet-info">
                <div class="wallet-address">${shortAddress}</div>
            </div>
            <a href="/account" class="dropdown-item">Settings</a>
            <a href="#" class="dropdown-item disconnect" onclick="disconnectWallet(); return false;">Disconnect</a>
        </div>
    `;

    // Replace the existing button with the new dropdown
    walletButton.parentNode.replaceChild(userDropdown, walletButton);

    // Add click handler for the new user icon
    setupUserIconHandlers(userDropdown);
    
    // Show nav items
    const navAuthItems = document.querySelector('.nav-auth-items');
    if (navAuthItems) {
        navAuthItems.classList.add('visible');
    }
}

// Setup user icon click handlers
function setupUserIconHandlers(userDropdown) {
    const userIcon = userDropdown.querySelector('.user-icon');
    if (!userIcon) return;

    // Remove any existing click handlers
    const newUserIcon = userIcon.cloneNode(true);
    userIcon.parentNode.replaceChild(newUserIcon, userIcon);

    // Add click handler for the user icon
    newUserIcon.addEventListener('click', (e) => {
        e.stopPropagation();
        const dropdown = userDropdown.querySelector('.dropdown-content');
        if (dropdown) {
            dropdown.style.display = dropdown.style.display === 'block' ? 'none' : 'block';
        }
    });

    // Close dropdown when clicking outside
    document.addEventListener('click', (e) => {
        if (!userDropdown.contains(e.target)) {
            const dropdown = userDropdown.querySelector('.dropdown-content');
            if (dropdown) {
                dropdown.style.display = 'none';
            }
        }
    });
}

// Add utility functions
function copyAddress(address) {
    navigator.clipboard.writeText(address);
    alert('Address copied to clipboard!');
}

function viewOnExplorer(address) {
    const explorerUrl = `https://etherscan.io/address/${address}`;
    window.open(explorerUrl, '_blank');
}

// Toggle dropdown
function toggleDropdown(event) {
    if (isConnected) {
        const dropdown = event.currentTarget.querySelector('.dropdown-content');
        dropdown.classList.toggle('show');
        event.stopPropagation();
    }
}

// Close dropdown when clicking outside
window.onclick = function(event) {
    if (!event.target.matches('.wallet-button')) {
        const dropdowns = document.getElementsByClassName('dropdown-content');
        for (const dropdown of dropdowns) {
            if (dropdown.classList.contains('show')) {
                dropdown.classList.remove('show');
            }
        }
    }
}

// Update the event listeners
document.addEventListener('DOMContentLoaded', async () => {
    console.log('Wallet.js DOMContentLoaded event fired');
    
    // Check session status on page load - only once
    const hasSession = await checkSession();

    // Initialize the Enter button if user is not connected
    if (!hasSession) {
        const enterButton = document.getElementById('connectWallet');
        if (enterButton && !enterButton.getAttribute('data-initialized')) {
            enterButton.innerHTML = 'Enter';
            enterButton.className = 'action-button';
            // Mark as initialized to prevent double initialization
            enterButton.setAttribute('data-initialized', 'true');
        }
    } else {
        // User is already connected
        const authModal = document.getElementById('authModal');
        if (authModal) {
            authModal.classList.remove('active');
        }

        // Setup click handlers for existing user icon if present
        const existingDropdown = document.querySelector('.user-dropdown');
        if (existingDropdown) {
            setupUserIconHandlers(existingDropdown);
        }
    }

    // Listen for account changes
    if (window.ethereum) {
        window.ethereum.on('accountsChanged', async function (accounts) {
            if (accounts.length === 0) {
                // User disconnected their wallet
                await disconnectWallet();
            } else {
                // User switched accounts, re-authenticate
                isConnected = false;
                currentAddress = null;
                await connectWallet();
            }
        });

        // Listen for chain changes
        window.ethereum.on('chainChanged', async function() {
            // Handle chain change by re-authenticating
            isConnected = false;
            currentAddress = null;
            await disconnectWallet();
            await connectWallet();
        });
    }
}); 