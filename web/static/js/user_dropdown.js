document.addEventListener('DOMContentLoaded', function() {
    const dropdownTrigger = document.querySelector('.user-dropdown-trigger');
    const dropdownContainer = document.querySelector('.user-dropdown-container');

    if (dropdownTrigger && dropdownContainer) {
        // Toggle dropdown on click
        dropdownTrigger.addEventListener('click', function(e) {
            e.stopPropagation();
            dropdownContainer.classList.toggle('active');
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', function(e) {
            if (!dropdownContainer.contains(e.target)) {
                dropdownContainer.classList.remove('active');
            }
        });

        // Handle logout
        const logoutLink = dropdownContainer.querySelector('a[href="/auth/logout"]');
        if (logoutLink) {
            logoutLink.addEventListener('click', async function(e) {
                e.preventDefault();
                
                try {
                    const response = await fetch('/auth/logout', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                    });

                    if (response.ok) {
                        window.location.reload();
                    } else {
                        console.error('Logout failed');
                    }
                } catch (error) {
                    console.error('Error during logout:', error);
                }
            });
        }
    }
}); 