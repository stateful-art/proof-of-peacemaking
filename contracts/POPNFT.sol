// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/utils/Counters.sol";
import "@openzeppelin/contracts/utils/Strings.sol";
import "@openzeppelin/contracts/utils/Base64.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

contract POPNFT is ERC721, Ownable {
    using Counters for Counters.Counter;
    using Strings for uint256;
    Counters.Counter private _tokenIds;

    // Add proxy operator
    address public proxyOperator;

    struct NFTMetadata {
        address creator;
        address acknowledger;
        uint256 expressionId;
        uint256 acknowledgementId;
        uint256 mintTimestamp;
        string expressionIPFS;
        string acknowledgementIPFS;
    }

    // Mapping from token ID to metadata
    mapping(uint256 => NFTMetadata) public tokenMetadata;
    
    // Events
    event ProofMinted(
        uint256 indexed expressionId,
        uint256 indexed acknowledgementId,
        address creator,
        address acknowledger,
        uint256[] tokenIds
    );

    constructor() ERC721("Proof of Peacemaking", "POP") {}

    // Set proxy operator
    function setProxyOperator(address _proxyOperator) external onlyOwner {
        proxyOperator = _proxyOperator;
    }

    // Prevent token transfers (soulbound)
    function _beforeTokenTransfer(
        address from,
        address to,
        uint256 tokenId,
        uint256 batchSize
    ) internal override {
        require(from == address(0) || to == address(0), "Token is soulbound");
        super._beforeTokenTransfer(from, to, tokenId, batchSize);
    }

    // Mint NFTs for both parties
    function mintProofs(
        uint256 _expressionId,
        uint256 _acknowledgementId,
        address _creator,
        address _acknowledger,
        string memory _expressionIPFS,
        string memory _acknowledgementIPFS
    ) external {
        require(msg.sender == proxyOperator, "Only proxy can mint");
        require(_creator != _acknowledger, "Creator and acknowledger must be different");

        uint256[] memory tokenIds = new uint256[](2);

        // Mint for creator
        uint256 creatorTokenId = _tokenIds.current();
        _mint(_creator, creatorTokenId);
        tokenIds[0] = creatorTokenId;
        _tokenIds.increment();

        // Mint for acknowledger
        uint256 acknowledgerTokenId = _tokenIds.current();
        _mint(_acknowledger, acknowledgerTokenId);
        tokenIds[1] = acknowledgerTokenId;
        _tokenIds.increment();

        // Store metadata for both tokens
        NFTMetadata memory metadata = NFTMetadata({
            creator: _creator,
            acknowledger: _acknowledger,
            expressionId: _expressionId,
            acknowledgementId: _acknowledgementId,
            mintTimestamp: block.timestamp,
            expressionIPFS: _expressionIPFS,
            acknowledgementIPFS: _acknowledgementIPFS
        });

        tokenMetadata[creatorTokenId] = metadata;
        tokenMetadata[acknowledgerTokenId] = metadata;

        emit ProofMinted(
            _expressionId,
            _acknowledgementId,
            _creator,
            _acknowledger,
            tokenIds
        );
    }

    // Generate token URI with metadata
    function tokenURI(uint256 tokenId) public view virtual override returns (string memory) {
        require(_exists(tokenId), "Token does not exist");
        
        NFTMetadata memory metadata = tokenMetadata[tokenId];
        
        string memory json = Base64.encode(
            bytes(string(
                abi.encodePacked(
                    '{"name": "Proof of Peacemaking #', 
                    tokenId.toString(),
                    '", "description": "This soulbound NFT represents a verified proof of peacemaking between two parties.", ',
                    '"attributes": [',
                        '{"trait_type": "Creator", "value": "', toString(metadata.creator), '"},',
                        '{"trait_type": "Acknowledger", "value": "', toString(metadata.acknowledger), '"},',
                        '{"trait_type": "Expression ID", "value": "', metadata.expressionId.toString(), '"},',
                        '{"trait_type": "Acknowledgment ID", "value": "', metadata.acknowledgementId.toString(), '"},',
                        '{"trait_type": "Mint Date", "value": "', metadata.mintTimestamp.toString(), '"}'
                    '], ',
                    '"properties": {',
                        '"expression": "', metadata.expressionIPFS, '",',
                        '"acknowledgement": "', metadata.acknowledgementIPFS, '"',
                    '}'
                    '}'
                )
            ))
        );

        return string(abi.encodePacked("data:application/json;base64,", json));
    }

    // Helper function to convert address to string
    function toString(address account) internal pure returns (string memory) {
        return Strings.toHexString(uint160(account), 20);
    }
} 