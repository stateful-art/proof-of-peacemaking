// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/IERC721Metadata.sol";
import "@openzeppelin/contracts/utils/Counters.sol";
import "../libraries/LibDiamond.sol";
import "../libraries/LibPOPNFT.sol";
import "../libraries/LibStorage.sol";

contract POPNFTFacet is IERC721, IERC721Metadata {
    using Counters for Counters.Counter;

    event POPNFTMinted(
        address indexed to, 
        uint256 indexed tokenId, 
        uint256 indexed expressionId,
        uint256 acknowledgementId,
        string uri
    );

    event Paused(address account);
    event Unpaused(address account);

    error ReentrantCall();

    function name() external pure override returns (string memory) {
        return "Proof of Personhood NFT";
    }

    function symbol() external pure override returns (string memory) {
        return "POPNFT";
    }

    function tokenURI(uint256 tokenId) external view override returns (string memory) {
        require(_exists(tokenId), "POPNFTFacet: URI query for nonexistent token");
        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        return ps.tokenURIs[tokenId];
    }

    function balanceOf(address owner) external view override returns (uint256) {
        require(owner != address(0), "POPNFTFacet: balance query for zero address");
        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        return ps.balances[owner];
    }

    function ownerOf(uint256 tokenId) external view override returns (address) {
        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        address owner = ps.owners[tokenId];
        require(owner != address(0), "POPNFTFacet: owner query for nonexistent token");
        return owner;
    }

    function mint(
        address to, 
        uint256 expressionId,
        uint256 acknowledgementId,
        address expressionCreator,
        string memory uri
    ) external {
        require(bytes(uri).length > 0, "POPNFTFacet: URI cannot be empty");
        require(to != address(this), "POPNFTFacet: Cannot mint to contract itself");
        
        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        LibStorage.NFTMetadataStorage storage ns = LibStorage.nftMetadataStorage();
        
        require(!ps.paused, "POPNFTFacet: Contract is paused");
        require(ps.balances[to] == 0, "POPNFTFacet: Address already has a POPNFT");
        
        if (!ps._notEntered) revert ReentrantCall();
        ps._notEntered = false;
        
        LibDiamond.enforceIsContractOwner();
        
        uint256 tokenId = ps.tokenIdCounter.current();
        ps.tokenIdCounter.increment();
        
        _mint(to, tokenId);
        ps.tokenURIs[tokenId] = uri;
        ps.tokenToExpressionId[tokenId] = expressionId;
        ps.hasValidPOP[to] = true;
        
        // Store metadata in NFTMetadataStorage
        ns.tokenMetadata[tokenId] = LibStorage.NFTMetadata({
            creator: expressionCreator,
            acknowledger: to,
            expressionId: expressionId,
            acknowledgementId: acknowledgementId,
            mintTimestamp: block.timestamp,
            expressionIPFS: "",  // These can be populated if needed
            acknowledgementIPFS: ""
        });
        
        // Event emission with acknowledgementId
        emit POPNFTMinted(to, tokenId, expressionId, acknowledgementId, uri);
        
        ps._notEntered = true;
    }

    function _mint(address to, uint256 tokenId) internal {
        require(to != address(0), "POPNFTFacet: mint to zero address");
        require(!_exists(tokenId), "POPNFTFacet: token already minted");

        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        ps.balances[to] += 1;
        ps.owners[tokenId] = to;

        emit Transfer(address(0), to, tokenId);
    }

    function _exists(uint256 tokenId) internal view returns (bool) {
        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        return ps.owners[tokenId] != address(0);
    }

    // Required IERC721 functions
    function approve(address to, uint256 tokenId) external override {
        revert("POPNFTFacet: POPNFT tokens are soulbound and cannot be transferred");
    }

    function getApproved(uint256 tokenId) external view override returns (address) {
        require(_exists(tokenId), "POPNFTFacet: approved query for nonexistent token");
        return address(0);
    }

    function setApprovalForAll(address operator, bool approved) external override {
        revert("POPNFTFacet: POPNFT tokens are soulbound and cannot be transferred");
    }

    function isApprovedForAll(address owner, address operator) external pure override returns (bool) {
        return false;
    }

    function transferFrom(address from, address to, uint256 tokenId) external override {
        revert("POPNFTFacet: POPNFT tokens are soulbound and cannot be transferred");
    }

    function safeTransferFrom(address from, address to, uint256 tokenId) external override {
        revert("POPNFTFacet: POPNFT tokens are soulbound and cannot be transferred");
    }

    function safeTransferFrom(address from, address to, uint256 tokenId, bytes calldata data) external override {
        revert("POPNFTFacet: POPNFT tokens are soulbound and cannot be transferred");
    }

    function initializePOPNFT() external {
        LibDiamond.enforceIsContractOwner();
        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        require(!ps.initialized, "POPNFTFacet: Already initialized");
        
        ps.initialized = true;
        ps._notEntered = true;
    }

    function supportsInterface(bytes4 interfaceId) external pure returns (bool) {
        return
            interfaceId == type(IERC721).interfaceId ||
            interfaceId == type(IERC721Metadata).interfaceId;
    }

    function pause() external {
        LibDiamond.enforceIsContractOwner();
        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        ps.paused = true;
        emit Paused(msg.sender);
    }

    function unpause() external {
        LibDiamond.enforceIsContractOwner();
        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        ps.paused = false;
        emit Unpaused(msg.sender);
    }

    function getExpressionId(uint256 tokenId) external view returns (uint256) {
        require(_exists(tokenId), "POPNFTFacet: Query for nonexistent token");
        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        return ps.tokenToExpressionId[tokenId];
    }

    function getAcknowledgements(uint256 tokenId) external view returns (LibStorage.Acknowledgement[] memory) {
        require(_exists(tokenId), "POPNFTFacet: Query for nonexistent token");
        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        return ps.tokenToAcknowledgements[tokenId];
    }

    function hasValidPOP(address account) external view returns (bool) {
        LibPOPNFT.POPNFTStorage storage ps = LibPOPNFT.popnftStorage();
        return ps.hasValidPOP[account];
    }
} 