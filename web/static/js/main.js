// Global state
let currentNetwork;
const SUPPORTED_NETWORKS = {
    5: 'Goerli Testnet',  // We'll start with Goerli testnet
    1: 'Ethereum Mainnet'
};

// Initialize the app
async function initApp() {
    if (window.ethereum) {
        // Get current network
        const chainId = await window.ethereum.request({ method: 'eth_chainId' });
        handleNetworkChange(chainId);

        // Listen for network changes
        window.ethereum.on('chainChanged', handleNetworkChange);
    }

    // Add event listeners to buttons
    const createExpressionBtn = document.getElementById('createExpression');
    const viewExpressionsBtn = document.getElementById('viewExpressions');
    const newExpressionBtn = document.getElementById('newExpression');

    if (createExpressionBtn) {
        createExpressionBtn.addEventListener('click', showCreateExpressionModal);
    }

    if (viewExpressionsBtn) {
        viewExpressionsBtn.addEventListener('click', () => {
            window.location.href = '/dashboard';
        });
    }

    if (newExpressionBtn) {
        newExpressionBtn.addEventListener('click', showCreateExpressionModal);
    }
}

// Handle network changes
function handleNetworkChange(chainId) {
    const networkName = SUPPORTED_NETWORKS[parseInt(chainId, 16)];
    currentNetwork = networkName;

    const networkInfo = document.getElementById('networkInfo');
    if (networkInfo) {
        if (networkName) {
            networkInfo.textContent = `Network: ${networkName}`;
            networkInfo.style.color = '#4CAF50';
        } else {
            networkInfo.textContent = 'Please switch to a supported network';
            networkInfo.style.color = '#f44336';
        }
    }
}

// Create expression modal
function showCreateExpressionModal() {
    if (!provider || !signer) {
        alert('Please connect your wallet first');
        return;
    }

    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <h2>Create Expression of Peace</h2>
            <textarea id="expressionText" placeholder="Write your expression of peace..."></textarea>
            <div class="modal-buttons">
                <button class="btn-secondary" onclick="closeModal()">Cancel</button>
                <button class="btn-primary" onclick="submitExpression()">Submit</button>
            </div>
        </div>
    `;

    document.body.appendChild(modal);
    
    // Add modal styles dynamically
    const style = document.createElement('style');
    style.textContent = `
        .modal {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.5);
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .modal-content {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            width: 90%;
            max-width: 600px;
        }
        .modal textarea {
            width: 100%;
            height: 150px;
            margin: 1rem 0;
            padding: 0.5rem;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
        .modal-buttons {
            display: flex;
            justify-content: flex-end;
            gap: 1rem;
        }
    `;
    document.head.appendChild(style);
}

function closeModal() {
    const modal = document.querySelector('.modal');
    if (modal) {
        modal.remove();
    }
}

async function submitExpression() {
    const expressionText = document.getElementById('expressionText').value;
    if (!expressionText.trim()) {
        alert('Please write your expression');
        return;
    }

    try {
        // TODO: Call smart contract to store expression
        console.log('Submitting expression:', expressionText);
        
        // For now, just show success message
        alert('Expression submitted successfully!');
        closeModal();
        
        // Refresh dashboard if we're on it
        if (window.location.pathname === '/dashboard') {
            loadExpressions();
        }
    } catch (error) {
        console.error('Error submitting expression:', error);
        alert('Failed to submit expression');
    }
}

// Load expressions for dashboard
async function loadExpressions() {
    const expressionsList = document.getElementById('expressionsList');
    if (!expressionsList) return;

    // TODO: Load expressions from smart contract
    // For now, show placeholder
    expressionsList.innerHTML = `
        <div class="expression-card">
            <p>No expressions yet</p>
            <button class="btn-primary" onclick="showCreateExpressionModal()">Create First Expression</button>
        </div>
    `;
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', initApp); 