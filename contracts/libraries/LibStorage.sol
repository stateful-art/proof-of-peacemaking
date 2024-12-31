// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

// LibStorage uses Diamond Storage 
// ~> https://eip2535diamonds.substack.com/p/keep-your-data-right-in-eip2535-diamonds?open=false#%C2%A7diamond-storage
// For different facets we can specify different locations to start storing data, 
// therefore preventing different facets with different state variables from clashing storage locations.
library LibStorage {
    // Storage namespace positions
    bytes32 constant EXPRESSION_STORAGE_POSITION = keccak256("pop.v1.expression.storage");
    bytes32 constant ACKNOWLEDGEMENT_STORAGE_POSITION = keccak256("pop.v1.acknowledgement.storage");
    bytes32 constant NFT_METADATA_STORAGE_POSITION = keccak256("pop.v1.nft.metadata.storage");
    bytes32 constant GAS_COST_STORAGE_POSITION = keccak256("pop.v1.gas.cost.storage");

    // Helper structs (no storage position needed)
    struct MediaContent {
        string textContent;
        string audioContent;
        string videoContent;
        string imageContent;
    }

    struct Expression {
        address creator;
        MediaContent content;
        uint256 timestamp;
        string ipfsHash;
    }

    struct Acknowledgement {
        address acknowledger;
        uint256 timestamp;
        string message;
        MediaContent content;
        string ipfsHash;
        uint256 expressionId;
        address expressionCreator;
        uint256 acknowledgementId;
    }

    struct NFTMetadata {
        address creator;
        address acknowledger;
        uint256 expressionId;
        uint256 acknowledgementId;
        uint256 mintTimestamp;
        string expressionIPFS;
        string acknowledgementIPFS;
    }

    // Diamond Storage structs (each with its own storage position)
    struct ExpressionStorage {
        mapping(uint256 => Expression) expressions;
        uint256 expressionCount;
        // Track acknowledgers per expression
        mapping(uint256 => address[]) expressionAcknowledgers;
    }

    struct AcknowledgementStorage {
        // expressionId => acknowledger => Acknowledgement
        mapping(uint256 => mapping(address => Acknowledgement)) acknowledgements;
        uint256 acknowledgementCount;
    }

    struct NFTMetadataStorage {
        mapping(uint256 => NFTMetadata) tokenMetadata;
        uint256 tokenCount;
    }

    struct GasCostStorage {
        uint256 expressionGasCost;
        uint256 acknowledgementGasCost;
        uint256 nftMintGasCost;
    }

    // Diamond Storage getters
    function expressionStorage() internal pure returns (ExpressionStorage storage es) {
        bytes32 position = EXPRESSION_STORAGE_POSITION;
        assembly {
            es.slot := position
        }
    }

    function acknowledgementStorage() internal pure returns (AcknowledgementStorage storage s) {
        bytes32 position = ACKNOWLEDGEMENT_STORAGE_POSITION;
        assembly {
            s.slot := position
        }
    }

    function nftMetadataStorage() internal pure returns (NFTMetadataStorage storage ns) {
        bytes32 position = NFT_METADATA_STORAGE_POSITION;
        assembly {
            ns.slot := position
        }
    }

    function gasCostStorage() internal pure returns (GasCostStorage storage gs) {
        bytes32 position = GAS_COST_STORAGE_POSITION;
        assembly {
            gs.slot := position
        }
    }
} 