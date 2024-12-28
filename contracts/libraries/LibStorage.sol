// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

library LibStorage {
    bytes32 constant STORAGE_POSITION = keccak256("pop.standard.storage");

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
        mapping(address => Acknowledgement) acknowledgments;
        address[] acknowledgers;
        string ipfsHash;
    }

    struct Acknowledgement {
        address acknowledger;
        uint256 timestamp;
        string message;
        MediaContent content;
        string ipfsHash;
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

    struct AppStorage {
        // Expression storage
        mapping(uint256 => Expression) expressions;
        uint256 expressionCount;

        // NFT storage
        mapping(uint256 => NFTMetadata) tokenMetadata;
        uint256 tokenCount;

        // Gas costs
        uint256 expressionGasCost;
        uint256 acknowledgementGasCost;
        uint256 nftMintGasCost;
    }

    function appStorage() internal pure returns (AppStorage storage ds) {
        bytes32 position = STORAGE_POSITION;
        assembly {
            ds.slot := position
        }
    }

    function getStorage() internal pure returns (AppStorage storage s) {
        bytes32 position = STORAGE_POSITION;
        assembly {
            s.slot := position
        }
    }
} 