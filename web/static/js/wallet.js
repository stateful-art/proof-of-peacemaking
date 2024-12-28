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

// Add this function to check if wallet is already connected
async function checkConnection() {
    if (window.ethereum) {
        const accounts = await window.ethereum.request({ method: 'eth_accounts' });
        if (accounts.length > 0) {
            await connectWallet();
        }
    }
}

// Update the connectWallet function
async function connectWallet() {
    try {
        await waitForEthers();

        if (typeof window.ethereum === 'undefined') {
            console.error('MetaMask not installed');
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

            // Update dashboard if on dashboard page
            const walletAddress = document.getElementById('walletAddress');
            if (walletAddress) {
                walletAddress.textContent = `Connected: ${address}`;
                const networkInfo = document.getElementById('networkInfo');
                if (networkInfo) {
                    const network = await provider.getNetwork();
                    networkInfo.textContent = `Network: ${network.name}`;
                }
            }

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

async function registerUser(address, email) {
    try {
        const response = await fetch('/auth/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ address, email }),
            credentials: 'include',
        });
        
        if (!response.ok) {
            throw new Error('Failed to register user');
        }
        
        const data = await response.json();
        return data.user;
    } catch (error) {
        console.error('Error registering user:', error);
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

    // Update dashboard if on dashboard page
    const walletAddress = document.getElementById('walletAddress');
    if (walletAddress) {
        walletAddress.textContent = 'Not connected';
        const networkInfo = document.getElementById('networkInfo');
        if (networkInfo) {
            networkInfo.textContent = 'Please connect your wallet';
        }
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
document.addEventListener('DOMContentLoaded', () => {
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
        window.ethereum.on('accountsChanged', function (accounts) {
            if (accounts.length === 0) {
                disconnectWallet();
            } else if (isConnected) {
                // Only reconnect if we were previously connected
                connectWallet();
            }
        });

        window.ethereum.on('chainChanged', function (chainId) {
            if (isConnected) {
                window.location.reload();
            }
        });
    }
}); 