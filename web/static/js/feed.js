document.addEventListener('DOMContentLoaded', function() {
    // Initialize the expression modal
    const expressionModal = new Modal('createExpressionModal');
    const expressionForm = new ExpressionForm('expressionForm');
    window.expressionModal = expressionModal; // Make it globally available

    // Create Expression button click handler
    const createExpressionBtn = document.getElementById('createExpressionBtn');
    if (createExpressionBtn) {
        createExpressionBtn.addEventListener('click', () => {
            console.log('Opening expression modal...'); // Debug log
            expressionModal.open();
        });
    } else {
        console.error('Create Expression button not found'); // Debug log
    }

    // Handle acknowledgment button clicks
    document.querySelectorAll('.acknowledge-button').forEach(button => {
        button.addEventListener('click', async function(e) {
            e.preventDefault();
            const expressionId = this.dataset.expressionId;
            const heartIcon = this.querySelector('.heart-icon');
            const countSpan = this.querySelector('.acknowledgement-count');
            const currentCount = parseInt(countSpan.textContent);
            
            // Optimistically update UI
            const isCurrentlyAcknowledged = button.classList.contains('acknowledged');
            const newCount = isCurrentlyAcknowledged ? currentCount - 1 : currentCount + 1;
            
            // Update UI immediately
            button.classList.toggle('acknowledged');
            heartIcon.classList.toggle('acknowledged');
            countSpan.textContent = newCount;
            
            try {
                const response = await fetch('/api/acknowledgements', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ expressionId })
                });

                if (!response.ok) {
                    // If request fails, revert the optimistic updates
                    button.classList.toggle('acknowledged');
                    heartIcon.classList.toggle('acknowledged');
                    countSpan.textContent = currentCount;
                    throw new Error('Failed to acknowledge expression');
                }

                // No need to update UI here since we already did it optimistically
                const result = await response.json();
                console.log('Acknowledgment updated:', result.status);
            } catch (error) {
                console.error('Error acknowledging expression:', error);
                // Show a toast or notification if you want to inform the user of the error
            }
        });
    });
}); 