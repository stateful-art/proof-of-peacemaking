const fs = require('fs');
const path = require('path');

function updateEnv(addresses) {
    const envPath = path.join(__dirname, '../../.env');
    let envContent = '';

    // Read existing .env content if it exists
    if (fs.existsSync(envPath)) {
        envContent = fs.readFileSync(envPath, 'utf8');
    }

    // Update or add each address
    const addressUpdates = {
        'DIAMOND_ADDRESS': addresses.diamond,
        'DIAMOND_CUT_FACET': addresses.diamondCut,
        'DIAMOND_LOUPE_FACET': addresses.diamondLoupe,
        'EXPRESSION_FACET': addresses.expression,
        'ACKNOWLEDGEMENT_FACET': addresses.acknowledgement,
        'POPNFT_FACET': addresses.popnft,
        'PERMISSIONS_FACET': addresses.permissions
    };

    Object.entries(addressUpdates).forEach(([key, value]) => {
        if (!value) return;

        const regex = new RegExp(`^${key}=.*$`, 'm');
        const newLine = `${key}=${value}`;

        if (regex.test(envContent)) {
            // Update existing line
            envContent = envContent.replace(regex, newLine);
        } else {
            // Add new line
            envContent += `\n${newLine}`;
        }
    });

    // Write back to .env
    fs.writeFileSync(envPath, envContent.trim() + '\n');
    console.log('\nUpdated .env with deployed contract addresses');
}

module.exports = updateEnv; 