const { ethers } = require("hardhat");

async function verifyDiamond(diamondAddress) {
    // Get Diamond Loupe interface
    const diamondLoupe = await ethers.getContractAt('IDiamondLoupe', diamondAddress);
    
    // Get all facets
    const facets = await diamondLoupe.facets();
    console.log('\nDeployed facets:');
    for (const facet of facets) {
        console.log(`${facet.facetAddress}: ${facet.functionSelectors.length} functions`);
    }

    // Test a function from each facet
    console.log('\nTesting facet functions:');
    
    // Test Expression facet
    const expressionFacet = await ethers.getContractAt('ExpressionFacet', diamondAddress);
    const expressionCount = await expressionFacet.expressionCount();
    console.log('Expression count:', expressionCount.toString());

    // Test Acknowledgement facet
    const acknowledgementFacet = await ethers.getContractAt('AcknowledgementFacet', diamondAddress);
    // Add relevant test

    // Test POPNFT facet
    const popnftFacet = await ethers.getContractAt('POPNFTFacet', diamondAddress);
    // Add relevant test

    console.log('\nVerification complete');
}

async function main() {
    const diamondAddress = process.env.DIAMOND_ADDRESS;
    if (!diamondAddress) {
        throw new Error('DIAMOND_ADDRESS not set in environment');
    }
    await verifyDiamond(diamondAddress);
}

main()
    .then(() => process.exit(0))
    .catch(error => {
        console.error(error);
        process.exit(1);
    }); 