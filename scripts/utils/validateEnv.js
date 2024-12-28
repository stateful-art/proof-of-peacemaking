const requiredEnvVars = [
    'INFURA_API_KEY',
    'PRIVATE_KEY',
    'ETHERSCAN_API_KEY',
    'IPFS_PROJECT_ID',
    'IPFS_PROJECT_SECRET'
];

function validateEnv() {
    const missingVars = requiredEnvVars.filter(varName => !process.env[varName]);
    
    if (missingVars.length > 0) {
        console.error('\nMissing required environment variables:');
        missingVars.forEach(varName => {
            console.error(`- ${varName}`);
        });
        console.error('\nPlease check your .env file\n');
        process.exit(1);
    }

    // Validate private key format
    if (!/^0x[0-9a-fA-F]{64}$/.test(process.env.PRIVATE_KEY)) {
        console.error('\nInvalid PRIVATE_KEY format. Should be a 64-character hex string with 0x prefix\n');
        process.exit(1);
    }

    console.log('Environment variables validated successfully');
}

module.exports = validateEnv; 