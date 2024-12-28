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

async function connectWallet() {
    try {
        await waitForEthers();

        // Check if MetaMask is installed
        if (typeof window.ethereum === 'undefined') {
            alert('Please install MetaMask to use this application');
            return;
        }

        // Request account access
        const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
        console.log('Connected accounts:', accounts);
        
        // Setup ethers provider and signer
        provider = new ethers.BrowserProvider(window.ethereum);
        signer = await provider.getSigner();
        
        // Get and display the connected address
        const address = await signer.getAddress();
        console.log('Connected wallet address:', address);

        // Get and log network information
        const network = await provider.getNetwork();
        console.log('Connected network:', {
            name: network.name,
            chainId: network.chainId,
            ensAddress: network.ensAddress
        });
        
        const balance = await provider.getBalance(address);
        console.log('Wallet balance:', ethers.formatEther(balance), 'ETH');

        const connectButton = document.getElementById('connectWallet');
        if (connectButton) {
            connectButton.textContent = `${address.slice(0, 6)}...${address.slice(-4)}`;
        }
        
        // Update wallet status if on dashboard
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

// Wait for document to load before adding event listeners
document.addEventListener('DOMContentLoaded', () => {
    const connectButton = document.getElementById('connectWallet');
    if (connectButton) {
        connectButton.addEventListener('click', connectWallet);
    }

    // Listen for account changes
    if (window.ethereum) {
        window.ethereum.on('accountsChanged', function (accounts) {
            connectWallet();
        });

        // Listen for chain changes
        window.ethereum.on('chainChanged', function (chainId) {
            // Reload the page on chain change as recommended by MetaMask
            window.location.reload();
        });
    }
}); 