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
            updateWalletButton(data.address);
            
            // Update UI elements
            const navAuthButtons = document.getElementById('navAuthButtons');
            const connectWalletHero = document.getElementById('connectWalletHero');
            const authenticatedButtons = document.getElementById('authenticatedButtons');
            
            if (navAuthButtons) navAuthButtons.style.display = 'flex';
            if (connectWalletHero) connectWalletHero.style.display = 'none';
            if (authenticatedButtons) authenticatedButtons.style.display = 'block';
            
            return true;
        }
        return false;
    } catch (error) {
        console.error('Error checking session:', error);
        return false;
    }
}

// Add this function to check if wallet is already connected
async function checkConnection() {
    // Only check if we have a valid session
    const hasSession = await checkSession();
    if (hasSession) {
        return;
    }

    // If no session, just update the button to show Connect Wallet
    const walletButton = document.getElementById('connectWallet');
    if (walletButton) {
        walletButton.innerHTML = 'Connect Wallet';
        walletButton.className = 'btn-connect';
    }
}

// Update the connectWallet function
async function connectWallet() {
    if (typeof window.ethereum !== 'undefined') {
        try {
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
            }
        } catch (error) {
            console.error('Error connecting wallet:', error);
            isConnected = false;
            currentAddress = null;
        }
    } else {
        alert('Please install MetaMask to use this feature');
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
        const walletButton = document.getElementById('connectWallet');
        if (walletButton) {
            walletButton.innerHTML = 'Connect Wallet';
            walletButton.className = 'btn-connect';
        }

        // Update other UI elements
        const navAuthButtons = document.getElementById('navAuthButtons');
        const connectWalletHero = document.getElementById('connectWalletHero');
        const authenticatedButtons = document.getElementById('authenticatedButtons');
        
        if (navAuthButtons) navAuthButtons.style.display = 'none';
        if (connectWalletHero) connectWalletHero.style.display = 'block';
        if (authenticatedButtons) authenticatedButtons.style.display = 'none';

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
    const shortAddress = `${address.slice(0, 6)}...${address.slice(-4)}`;
    
    walletButton.innerHTML = `
        <span>${shortAddress}</span>
        <div class="dropdown-content">
            <div class="wallet-info">
                <div>Connected Wallet</div>
                <div class="wallet-address">${address}</div>
            </div>
            <button onclick="copyAddress('${address}')">Copy Address</button>
            <button onclick="viewOnExplorer('${address}')">View on Explorer</button>
            <button class="disconnect" onclick="disconnectWallet()">Disconnect</button>
        </div>
    `;
    walletButton.className = 'wallet-button connected';
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
    await checkSession();

    const connectButton = document.getElementById('connectWallet');
    if (connectButton) {
        connectButton.addEventListener('click', async (e) => {
            if (isConnected) {
                toggleDropdown(e);
            } else {
                try {
                    await connectWallet();
                } catch (error) {
                    console.error('Failed to connect wallet:', error);
                }
            }
        });
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