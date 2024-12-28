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
            
            // Setup provider and signer only if we have window.ethereum
            if (window.ethereum) {
                provider = new ethers.BrowserProvider(window.ethereum);
                signer = await provider.getSigner();
            }
            
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
    try {
        // Check session first
        const hasSession = await checkSession();
        if (hasSession) {
            return currentAddress;
        }

        await waitForEthers();

        if (typeof window.ethereum === 'undefined') {
            alert('Please install MetaMask to connect your wallet');
            return;
        }

        // Request account access
        const accounts = await ethereum.request({ method: 'eth_requestAccounts' });
        const address = accounts[0];
        
        try {
            // Get nonce for signing
            const nonceResponse = await fetch('/auth/nonce?address=' + address, {
                method: 'GET',
                credentials: 'include',
            });
            
            if (!nonceResponse.ok) {
                const errorData = await nonceResponse.json();
                throw new Error(errorData.error || 'Failed to get nonce');
            }
            const nonceData = await nonceResponse.json();

            // Create message for signing
            const message = `Sign this message to verify your wallet. Nonce: ${nonceData.nonce}`;
            console.log('Signing message:', message);
            
            // Request signature
            const signature = await ethereum.request({
                method: 'personal_sign',
                params: [message, address],
            });
            console.log('Got signature:', signature);
            
            // Verify signature
            const verifyResponse = await fetch('/auth/verify', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ address, signature }),
                credentials: 'include',
            });
            
            if (!verifyResponse.ok) {
                const errorData = await verifyResponse.json();
                throw new Error(errorData.error || 'Failed to verify signature');
            }

            const verifyData = await verifyResponse.json();
            if (!verifyData.valid) {
                throw new Error('Invalid signature');
            }

            // Update UI
            isConnected = true;
            currentAddress = address;
            updateWalletButton(address);

            // Setup provider and signer
            provider = new ethers.BrowserProvider(window.ethereum);
            signer = await provider.getSigner();

            return address;
        } catch (error) {
            console.error('Authentication error:', error);
            isConnected = false;
            throw error;
        }
    } catch (error) {
        console.error('Wallet connection error:', error);
        isConnected = false;
        throw error;
    }
}

// Add disconnect function
async function disconnectWallet() {
    try {
        await fetch('/auth/logout', {
            method: 'POST',
            credentials: 'include',
        });
    } catch (error) {
        console.error('Error logging out:', error);
    }

    isConnected = false;
    currentAddress = null;
    provider = null;
    signer = null;
    
    const walletButton = document.getElementById('connectWallet');
    walletButton.innerHTML = 'Connect Wallet';
    walletButton.className = 'btn-connect';
    
    const dropdown = document.querySelector('.dropdown-content');
    if (dropdown) {
        dropdown.classList.remove('show');
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
    await checkConnection();

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
            await connectWallet();
        });
    }
}); 