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
            alert('Please install MetaMask to use this application');
            return;
        }

        const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
        console.log('Connected accounts:', accounts);
        
        provider = new ethers.BrowserProvider(window.ethereum);
        signer = await provider.getSigner();
        
        const address = await signer.getAddress();
        console.log('Connected wallet address:', address);
        currentAddress = address;
        isConnected = true;

        const network = await provider.getNetwork();
        console.log('Connected network:', {
            name: network.name,
            chainId: network.chainId,
            ensAddress: network.ensAddress
        });
        
        const balance = await provider.getBalance(address);
        console.log('Wallet balance:', ethers.formatEther(balance), 'ETH');

        updateWalletButton(address);
        
        // Update dashboard if on dashboard page
        const walletAddress = document.getElementById('walletAddress');
        if (walletAddress) {
            walletAddress.textContent = `Connected: ${address}`;
            const networkInfo = document.getElementById('networkInfo');
            if (networkInfo) {
                networkInfo.textContent = `Network: ${network.name}`;
            }
        }

        return address;
    } catch (error) {
        console.error('Error connecting wallet:', error);
        alert('Failed to connect wallet: ' + error.message);
    }
}

// Add disconnect function
async function disconnectWallet() {
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
                await connectWallet();
            }
        });
    }

    // Check if already connected
    checkConnection();

    if (window.ethereum) {
        window.ethereum.on('accountsChanged', function (accounts) {
            if (accounts.length === 0) {
                disconnectWallet();
            } else {
                connectWallet();
            }
        });

        window.ethereum.on('chainChanged', function (chainId) {
            window.location.reload();
        });
    }
}); 