document.addEventListener('DOMContentLoaded', function() {
    // Create Expression button click handler
    const createExpressionBtn = document.getElementById('createExpressionBtn');
    if (createExpressionBtn) {
        createExpressionBtn.addEventListener('click', () => {
            expressionModal.open();
        });
    }

    // Handle acknowledgement buttons
    const acknowledgeButtons = document.querySelectorAll('.acknowledge-button');
    acknowledgeButtons.forEach(button => {
        button.addEventListener('click', async function() {
            const expressionId = this.dataset.expressionId;
            try {
                const response = await fetch('/api/acknowledgements', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ expressionId }),
                });

                if (response.ok) {
                    const acknowledgement = await response.json();
                    // Toggle the acknowledged state based on the status
                    const isActive = acknowledgement.Status === 'ACTIVE';
                    this.classList.toggle('acknowledged', isActive);
                    const heartIcon = this.querySelector('.heart-icon');
                    heartIcon.classList.toggle('acknowledged', isActive);
                    
                    // Get the current count of active acknowledgements
                    const countResponse = await fetch(`/api/acknowledgements/expression/${expressionId}`);
                    if (countResponse.ok) {
                        const acks = await countResponse.json();
                        const activeCount = acks.filter(ack => ack.Status === 'ACTIVE').length;
                        
                        // Update the count display
                        const countSpan = this.querySelector('.acknowledgement-count');
                        countSpan.textContent = activeCount;

                        // Update has-active-acknowledgements class based on count
                        this.classList.toggle('has-active-acknowledgements', activeCount > 0);
                    }
                } else {
                    const errorData = await response.json();
                    console.error('Failed to acknowledge expression:', errorData.error);
                }
            } catch (error) {
                console.error('Error:', error);
            }
        });
    });
}); 