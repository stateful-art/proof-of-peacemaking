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

    error ReentrantCall();

    function name() external pure override returns (string memory) {
        return "Proof of Personhood NFT";
    }

    function symbol() external pure override returns (string memory) {
        return "POPNFT";
    }

    function tokenURI(uint256 tokenId) external view override returns (string memory) {
        require(_exists(tokenId), "POPNFTFacet: URI query for nonexistent token");
        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        return s.tokenURIs[tokenId];
    }

    function balanceOf(address owner) external view override returns (uint256) {
        require(owner != address(0), "POPNFTFacet: balance query for zero address");
        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        return s.balances[owner];
    }

    function ownerOf(uint256 tokenId) external view override returns (address) {
        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        address owner = s.owners[tokenId];
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
        
        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        require(!s.paused, "POPNFTFacet: Contract is paused");
        require(s.balances[to] == 0, "POPNFTFacet: Address already has a POPNFT");
        
        if (!s._notEntered) revert ReentrantCall();
        s._notEntered = false;
        
        LibDiamond.enforceIsContractOwner();
        
        uint256 tokenId = s.tokenIdCounter.current();
        s.tokenIdCounter.increment();
        
        _mint(to, tokenId);
        s.tokenURIs[tokenId] = uri;
        s.tokenToExpressionId[tokenId] = expressionId;
        s.hasValidPOP[to] = true;
        
        // Store metadata in AppStorage
        LibStorage.AppStorage storage appStorage = LibStorage.getStorage();
        appStorage.tokenMetadata[tokenId] = LibStorage.NFTMetadata({
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
        
        s._notEntered = true;
    }

    function _mint(address to, uint256 tokenId) internal {
        require(to != address(0), "POPNFTFacet: mint to zero address");
        require(!_exists(tokenId), "POPNFTFacet: token already minted");

        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        s.balances[to] += 1;
        s.owners[tokenId] = to;

        emit Transfer(address(0), to, tokenId);
    }

    function _exists(uint256 tokenId) internal view returns (bool) {
        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        return s.owners[tokenId] != address(0);
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
        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        require(!s.initialized, "POPNFTFacet: Already initialized");
        
        s.initialized = true;
        s._notEntered = true;
    }

    function supportsInterface(bytes4 interfaceId) external pure returns (bool) {
        return
            interfaceId == type(IERC721).interfaceId ||
            interfaceId == type(IERC721Metadata).interfaceId;
    }

    function pause() external {
        LibDiamond.enforceIsContractOwner();
        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        s.paused = true;
        emit Paused(msg.sender);
    }

    function unpause() external {
        LibDiamond.enforceIsContractOwner();
        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        s.paused = false;
        emit Unpaused(msg.sender);
    }

    function getExpressionId(uint256 tokenId) external view returns (uint256) {
        require(_exists(tokenId), "POPNFTFacet: Query for nonexistent token");
        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        return s.tokenToExpressionId[tokenId];
    }

    function getAcknowledgements(uint256 tokenId) external view returns (LibStorage.Acknowledgement[] memory) {
        require(_exists(tokenId), "POPNFTFacet: Query for nonexistent token");
        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        return s.tokenToAcknowledgements[tokenId];
    }

    function hasValidPOP(address account) external view returns (bool) {
        LibPOPNFT.POPNFTStorage storage s = LibPOPNFT.getPOPNFTStorage();
        return s.hasValidPOP[account];
    }
} 