document.addEventListener('DOMContentLoaded', function() {
    const profileForm = document.getElementById('profileForm');
    const editProfileBtn = document.getElementById('editProfileBtn');
    const cancelEditBtn = document.getElementById('cancelEditBtn');
    const submitBtn = profileForm.querySelector('button[type="submit"]');
    const formInputs = profileForm.querySelectorAll('input');
    const userMenuBtn = document.getElementById('userMenuBtn');
    const userDropdown = document.getElementById('userDropdown');
    let originalFormData = {};

    // Store original form data
    formInputs.forEach(input => {
        originalFormData[input.name] = input.value;
    });

    // Toggle user menu dropdown
    if (userMenuBtn) {
        userMenuBtn.addEventListener('click', function(e) {
            e.stopPropagation();
            userDropdown.classList.toggle('show');
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', function(e) {
            if (!userMenuBtn.contains(e.target) && !userDropdown.contains(e.target)) {
                userDropdown.classList.remove('show');
            }
        });
    }

    // Enable form editing
    editProfileBtn.addEventListener('click', function() {
        formInputs.forEach(input => {
            input.disabled = false;
        });
        submitBtn.disabled = false;
        cancelEditBtn.disabled = false;
        this.style.display = 'none';
    });

    // Cancel editing
    cancelEditBtn.addEventListener('click', function() {
        formInputs.forEach(input => {
            input.disabled = true;
            input.value = originalFormData[input.name];
            input.classList.remove('error');
        });
        submitBtn.disabled = true;
        this.disabled = true;
        editProfileBtn.style.display = 'block';
        
        // Clear any error messages
        const errorMessages = profileForm.querySelectorAll('.error-message');
        errorMessages.forEach(msg => msg.remove());

        // If there were no edits being made, navigate back
        if (!editProfileBtn.style.display || editProfileBtn.style.display === 'block') {
            handleCancel();
        }
    });

    // Handle form submission
    profileForm.addEventListener('submit', async function(e) {
        e.preventDefault();

        // Clear previous error messages
        const errorMessages = profileForm.querySelectorAll('.error-message');
        errorMessages.forEach(msg => msg.remove());
        formInputs.forEach(input => input.classList.remove('error'));

        const formData = new FormData(this);
        const userData = {};
        formData.forEach((value, key) => {
            userData[key] = value;
        });

        try {
            const response = await fetch('/api/users/profile', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(userData),
            });

            const data = await response.json();

            if (response.ok) {
                // Update original form data
                formInputs.forEach(input => {
                    originalFormData[input.name] = input.value;
                    input.disabled = true;
                });
                submitBtn.disabled = true;
                cancelEditBtn.disabled = true;
                editProfileBtn.style.display = 'block';
            } else {
                // Handle validation errors
                if (data.errors) {
                    Object.keys(data.errors).forEach(field => {
                        const input = profileForm.querySelector(`[name="${field}"]`);
                        if (input) {
                            input.classList.add('error');
                            const errorDiv = document.createElement('div');
                            errorDiv.className = 'error-message';
                            errorDiv.textContent = data.errors[field];
                            input.parentNode.appendChild(errorDiv);
                        }
                    });
                }
            }
        } catch (error) {
            console.error('Error updating profile:', error);
        }
    });

    // Handle wallet connection
    window.startWalletConnection = async function() {
        try {
            if (!window.ethereum) {
                throw new Error('MetaMask is not installed. Please install MetaMask to connect your wallet.');
            }

            // Request account access
            const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
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

            // Request signature
            const provider = new ethers.BrowserProvider(window.ethereum);
            const signer = await provider.getSigner();
            const message = `Welcome to Proof of Peacemaking!\n\nPlease sign this message to verify you own this wallet.\n\nNonce: ${nonce}`;
            const signature = await signer.signMessage(message);

            // Connect wallet to account
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
                // Dispatch wallet error event
                const errorEvent = new CustomEvent('walletError', {
                    detail: { message: data.error || 'Failed to connect wallet' }
                });
                window.dispatchEvent(errorEvent);
            }
        } catch (error) {
            // Dispatch wallet error event
            const errorEvent = new CustomEvent('walletError', {
                detail: { message: error.message || 'Error connecting wallet' }
            });
            window.dispatchEvent(errorEvent);
            console.error('Error connecting wallet:', error);
        }
    };
}); 