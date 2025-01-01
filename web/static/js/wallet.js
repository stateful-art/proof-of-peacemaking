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
            // Add click handler for auth modal
            enterButton.onclick = () => {
                const authModal = document.getElementById('authModal');
                if (authModal) {
                    authModal.classList.add('active');
                }
            };
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
        if (window.isMetaMaskConnected) {
            throw new Error('Wallet connection already pending. Please check MetaMask.');
        }

        // Check if MetaMask is installed
        if (!window.ethereum) {
            throw new Error('Please install MetaMask to continue.');
        }

        // Request account access
        const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
        if (accounts.length === 0) {
            throw new Error('No accounts found. Please connect your wallet.');
        }

        const address = accounts[0];

        // Get nonce for the wallet
        const nonceResponse = await fetch('/api/users/wallet-nonce', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ address }),
        });

        if (!nonceResponse.ok) {
            throw new Error('Failed to get nonce for wallet verification');
        }

        const { nonce } = await nonceResponse.json();

        // Set connection state
        window.isMetaMaskConnected = true;

        // Request signature
        const provider = new ethers.providers.Web3Provider(window.ethereum);
        const signer = provider.getSigner();
        const signature = await signer.signMessage(
            `Welcome to Proof of Peacemaking!\n\nPlease sign this message to verify you own this wallet.\n\nNonce: ${nonce}`
        );

        // Verify signature and connect wallet
        const response = await fetch('/api/users/connect-wallet', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                address,
                signature,
                nonce
            }),
        });

        const data = await response.json();
        if (response.ok) {
            window.location.reload();
        } else {
            throw new Error(data.error || 'Failed to connect wallet');
        }
    } catch (error) {
        console.error('Error connecting wallet:', error);
        window.isMetaMaskConnected = false;
        throw error; // Re-throw to be handled by the caller
    }
}

// Add disconnect function
async function disconnectWallet() {
    try {
        // Just call logout endpoint, no need to update wallet status for email auth
        const response = await fetch('/auth/logout', {
            method: 'POST',
            credentials: 'include'
        });

        if (response.ok) {
            window.location.reload();
        } else {
            throw new Error('Failed to logout');
        }
    } catch (error) {
        console.error('Error during logout:', error);
        window.location.reload(); // Reload anyway to ensure clean state
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
        // Add click handler for auth modal
        walletButton.onclick = () => {
            const authModal = document.getElementById('authModal');
            if (authModal) {
                authModal.classList.add('active');
            }
        };
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

// Handle wallet connection for account page
window.startWalletConnection = async function() {
    try {
        await connectWallet();
    } catch (error) {
        // Dispatch wallet error event for the account page
        const errorEvent = new CustomEvent('walletError', {
            detail: { message: error.message || 'Error connecting wallet' }
        });
        window.dispatchEvent(errorEvent);
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
            // Add click handler for auth modal
            enterButton.onclick = () => {
                const authModal = document.getElementById('authModal');
                if (authModal) {
                    authModal.classList.add('active');
                }
            };
        }
    }
}); 