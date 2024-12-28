const { expect } = require("chai");
const { ethers } = require("hardhat");
const { getSelectors, FacetCutAction } = require('../libraries/diamond.js');

describe("Diamond", function() {
    let diamond;
    let diamondCutFacet;
    let diamondLoupeFacet;
    let expressionFacet;
    let owner;
    let addr1;
    let addr2;

    beforeEach(async function() {
        [owner, addr1, addr2] = await ethers.getSigners();

        // Deploy DiamondCutFacet
        const DiamondCutFacet = await ethers.getContractFactory("DiamondCutFacet");
        diamondCutFacet = await DiamondCutFacet.deploy();
        await diamondCutFacet.deployed();

        // Deploy Diamond
        const Diamond = await ethers.getContractFactory("Diamond");
        diamond = await Diamond.deploy(owner.address, diamondCutFacet.address);
        await diamond.deployed();

        // Deploy and add facets
        const DiamondLoupeFacet = await ethers.getContractFactory("DiamondLoupeFacet");
        diamondLoupeFacet = await DiamondLoupeFacet.deploy();
        await diamondLoupeFacet.deployed();

        const ExpressionFacet = await ethers.getContractFactory("ExpressionFacet");
        expressionFacet = await ExpressionFacet.deploy();
        await expressionFacet.deployed();

        // Add facets to diamond
        const cut = [
            {
                facetAddress: diamondLoupeFacet.address,
                action: FacetCutAction.Add,
                functionSelectors: getSelectors(diamondLoupeFacet)
            },
            {
                facetAddress: expressionFacet.address,
                action: FacetCutAction.Add,
                functionSelectors: getSelectors(expressionFacet)
            }
        ];

        await diamond.diamondCut(cut, ethers.constants.AddressZero, "0x");
    });

    describe("Deployment", function() {
        it("Should set the right owner", async function() {
            expect(await diamond.owner()).to.equal(owner.address);
        });

        it("Should have all facets", async function() {
            const loupe = await ethers.getContractAt('IDiamondLoupe', diamond.address);
            const facets = await loupe.facets();
            expect(facets.length).to.equal(3); // DiamondCut, DiamondLoupe, Expression
        });
    });

    // Add more test cases for each facet
}); 