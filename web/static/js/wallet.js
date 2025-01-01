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
            
            return true;
        }
        
        return false;
    } catch (error) {
        console.error('Error checking session:', error);
        return false;
    }
}

// Update the connectWallet function
async function connectWallet() {
    // First check if MetaMask is installed
    if (typeof window.ethereum === 'undefined') {
        // If no MetaMask, open auth modal instead
        openAuthModal();
        return;
    }

    try {
        const loadingSpinner = document.getElementById('loadingSpinner');
        const connectButton = document.getElementById('connectWallet');
        
        // Show loading spinner and hide connect button
        if (loadingSpinner && connectButton) {
            connectButton.style.display = 'none';
            loadingSpinner.style.display = 'inline-flex';
        }

        // Request account access
        const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
        const address = accounts[0];

        // Get nonce from server
        const nonceResponse = await fetch(`/auth/nonce?address=${address}`);
        const nonceData = await nonceResponse.json();

        // Request signature with the exact same message format as backend
        const message = `Sign this message to verify your wallet. Nonce: ${nonceData.nonce}`;
        const signature = await window.ethereum.request({
            method: 'personal_sign',
            params: [
                message,
                address
            ]
        });

        // Verify signature with server
        const verifyResponse = await fetch('/auth/verify', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                address: address,
                signature: signature
            }),
            credentials: 'include'  // Important: include credentials
        });

        if (verifyResponse.ok) {
            isConnected = true;
            currentAddress = address;
            // Update UI
            updateWalletButton(address);
            // Dispatch wallet connected event
            window.dispatchEvent(new Event('walletConnected'));
            // Ensure the page reloads after everything is set
            await checkSession();  // Double-check session is set
            window.location.reload();
        } else {
            const errorData = await verifyResponse.json();
            console.error('Failed to verify signature:', errorData.error);
            isConnected = false;
            currentAddress = null;
            
            // Hide loading spinner and show connect button on error
            if (loadingSpinner && connectButton) {
                loadingSpinner.style.display = 'none';
                connectButton.style.display = 'inline-block';
            }
        }
    } catch (error) {
        console.error('Error connecting wallet:', error);
        isConnected = false;
        currentAddress = null;
        
        // Hide loading spinner and show connect button on error
        const loadingSpinner = document.getElementById('loadingSpinner');
        const connectButton = document.getElementById('connectWallet');
        if (loadingSpinner && connectButton) {
            loadingSpinner.style.display = 'none';
            connectButton.style.display = 'inline-block';
        }
    }
}

// Add disconnect function
async function disconnectWallet() {
    try {
        isConnected = false;
        currentAddress = null;
        
        // Clear session cookie
        await fetch('/auth/logout', {
            method: 'POST',
            credentials: 'include'
        });

        // Update UI
        const enterButton = document.getElementById('connectWallet');
        if (enterButton) {
            enterButton.innerHTML = 'Enter';
            enterButton.className = 'action-button';
            enterButton.onclick = openAuthModal;
        }

        // Update nav items visibility
        const navAuthItems = document.querySelector('.nav-auth-items');
        if (navAuthItems) {
            navAuthItems.classList.remove('visible');
        }

        // Dispatch wallet disconnected event
        window.dispatchEvent(new Event('walletDisconnected'));
        
        // Reload the page
        window.location.reload();
    } catch (error) {
        console.error('Error disconnecting wallet:', error);
        // Even if there's an error, try to reload to get to a clean state
        window.location.reload();
    }
}

// Add function to update wallet button
function updateWalletButton(address) {
    const walletButton = document.getElementById('connectWallet');
    if (!walletButton) return;

    if (!address) {
        // If no address, set up the Enter button
        walletButton.innerHTML = 'Enter';
        walletButton.className = 'action-button';
        walletButton.onclick = openAuthModal;
        return;
    }

    const shortAddress = `${address.slice(0, 6)}...${address.slice(-4)}`;

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
    // Check session status on page load - only once
    const hasSession = await checkSession();

    // Initialize the Enter button if user is not connected
    if (!hasSession) {
        const enterButton = document.getElementById('connectWallet');
        if (enterButton) {
            enterButton.innerHTML = 'Enter';
            enterButton.className = 'action-button';
            enterButton.onclick = openAuthModal;
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