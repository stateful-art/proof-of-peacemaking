document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('newsletterForm');
    const emailInput = document.getElementById('newsletterEmail');
    const submitButton = document.getElementById('newsletterSubmit');
    const messageDiv = document.getElementById('newsletterMessage');

    function validateEmail(email) {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email);
    }

    function setLoading(isLoading) {
        if (isLoading) {
            submitButton.innerHTML = '<div class="spinner"></div>';
            submitButton.disabled = true;
        } else {
            submitButton.innerHTML = 'Subscribe';
            submitButton.disabled = false;
        }
    }

    function showMessage(message, isError = false) {
        messageDiv.textContent = message;
        messageDiv.className = 'newsletter-message ' + (isError ? 'error' : 'success');
    }

    form.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const email = emailInput.value.trim();
        
        // Clear previous messages
        messageDiv.textContent = '';
        emailInput.classList.remove('error');
        
        // Validate email
        if (!email) {
            emailInput.classList.add('error');
            showMessage('Please enter your email address', true);
            return;
        }
        
        if (!validateEmail(email)) {
            emailInput.classList.add('error');
            showMessage('Please enter a valid email address', true);
            return;
        }
        
        // Start loading
        setLoading(true);
        
        try {
            const response = await fetch('/join-newsletter', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ email }),
            });
            
            if (response.ok) {
                showMessage('Thank you for subscribing to our newsletter!');
                emailInput.value = ''; // Clear the input
            } else {
                throw new Error('Failed to subscribe');
            }
        } catch (error) {
            showMessage('Sorry, something went wrong. Please try again later.', true);
        } finally {
            setLoading(false);
        }
    });
});