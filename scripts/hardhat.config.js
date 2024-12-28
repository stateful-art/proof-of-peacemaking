require("@nomiclabs/hardhat-waffle");
require("@nomiclabs/hardhat-etherscan");
require('dotenv').config();
const validateEnv = require('./utils/validateEnv');

// Validate environment variables
if (process.env.NODE_ENV !== 'test') {
    validateEnv();
}

task("deploy", "Deploys the Diamond and all facets")
  .setAction(async () => {
    const { main } = require("./deploy.js");
    await main();
  });

module.exports = {
  solidity: {
    version: "0.8.20",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      }
    }
  },
  paths: {
    sources: "../contracts",    // Look for contracts in parent directory
    tests: "./test",           // Tests are in scripts/test
    cache: "./cache",          // Keep cache in scripts folder
    artifacts: "./artifacts"   // Keep artifacts in scripts folder
  },
  networks: {
    hardhat: {},
    localhost: {
      url: "http://127.0.0.1:8545"
    },
    // Add other networks as needed
  }
}; 