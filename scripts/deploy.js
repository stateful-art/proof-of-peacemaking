const { getSelectors, FacetCutAction } = require('./libraries/diamond.js');
const { ethers } = require("hardhat");
const validateEnv = require('./utils/validateEnv');
const updateEnv = require('./utils/updateEnv');

async function deployDiamond() {
    // Validate environment variables
    validateEnv();

    const accounts = await ethers.getSigners();
    const contractOwner = accounts[0];

    // Deploy DiamondCutFacet
    const DiamondCutFacet = await ethers.getContractFactory('DiamondCutFacet');
    const diamondCutFacet = await DiamondCutFacet.deploy();
    await diamondCutFacet.deployed();
    console.log('DiamondCutFacet deployed:', diamondCutFacet.address);

    // Deploy Diamond
    const Diamond = await ethers.getContractFactory('Diamond');
    const diamond = await Diamond.deploy(contractOwner.address, diamondCutFacet.address);
    await diamond.deployed();
    console.log('Diamond deployed:', diamond.address);

    // Deploy DiamondLoupeFacet
    const DiamondLoupeFacet = await ethers.getContractFactory('DiamondLoupeFacet');
    const diamondLoupeFacet = await DiamondLoupeFacet.deploy();
    await diamondLoupeFacet.deployed();
    console.log('DiamondLoupeFacet deployed:', diamondLoupeFacet.address);

    // Deploy facets
    console.log('Deploying facets...');
    
    const FacetNames = [
        'ExpressionFacet',
        'AcknowledgementFacet',
        'POPNFTFacet',
        'PermissionsFacet'
    ];
    
    const cut = [];
    for (const FacetName of FacetNames) {
        const Facet = await ethers.getContractFactory(FacetName);
        const facet = await Facet.deploy();
        await facet.deployed();
        console.log(`${FacetName} deployed: ${facet.address}`);
        
        cut.push({
            facetAddress: facet.address,
            action: FacetCutAction.Add,
            functionSelectors: getSelectors(facet)
        });
    }

    // Upgrade diamond with facets
    console.log('Diamond Cut:', cut);
    const diamondCut = await ethers.getContractAt('IDiamondCut', diamond.address);
    const tx = await diamondCut.diamondCut(cut, ethers.constants.AddressZero, '0x');
    const receipt = await tx.wait();
    if (!receipt.status) {
        throw Error(`Diamond upgrade failed: ${tx.hash}`);
    }
    console.log('Diamond cut complete');

    // Update .env with deployed addresses
    updateEnv({
        diamond: diamond.address,
        diamondCut: diamondCutFacet.address,
        diamondLoupe: diamondLoupeFacet.address,
        expression: expressionFacet.address,
        acknowledgement: acknowledgementFacet.address,
        popnft: popnftFacet.address,
        permissions: permissionsFacet.address
    });

    return diamond.address;
}

async function main() {
    const diamondAddress = await deployDiamond();
    console.log('Completed deployment:', diamondAddress);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    }); 