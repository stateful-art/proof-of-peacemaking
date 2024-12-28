document.addEventListener('DOMContentLoaded', () => {
    // Handle create expression button
    const createExpressionBtn = document.getElementById('createExpression');
    if (createExpressionBtn) {
        createExpressionBtn.addEventListener('click', () => {
            // TODO: Implement create expression modal
            console.log('Opening create expression modal...');
        });
    }

    // Handle view expressions button
    const viewExpressionsBtn = document.getElementById('viewExpressions');
    if (viewExpressionsBtn) {
        viewExpressionsBtn.addEventListener('click', () => {
            window.location.href = '/feed';
        });
    }

    // Check if wallet is connected on page load
    if (typeof checkConnection === 'function') {
        checkConnection();
    }
}); 