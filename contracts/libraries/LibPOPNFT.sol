// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/utils/Counters.sol";
import "./LibStorage.sol";

library LibPOPNFT {
    using Counters for Counters.Counter;

    bytes32 constant POPNFT_STORAGE_POSITION = keccak256("diamond.standard.popnft.storage");

    struct POPNFTStorage {
        // Token name
        string name;
        // Token symbol
        string symbol;
        // Mapping from token ID to owner address
        mapping(uint256 => address) owners;
        // Mapping owner address to token count
        mapping(address => uint256) balances;
        // Mapping from token ID to token URI
        mapping(uint256 => string) tokenURIs;
        // Counter for token IDs
        Counters.Counter tokenIdCounter;
        // Contract state
        bool initialized;
        bool paused;
        bool _notEntered;
        
        // POP verification data
        mapping(uint256 => uint256) tokenToExpressionId; // Links token to the expression it verifies
        mapping(uint256 => Acknowledgement[]) tokenToAcknowledgements; // Links token to its acknowledgements
        mapping(address => bool) hasValidPOP; // Tracks if an address has valid proof of personhood
    }

    function getPOPNFTStorage() internal pure returns (POPNFTStorage storage s) {
        bytes32 position = POPNFT_STORAGE_POSITION;
        assembly {
            s.slot := position
        }
    }

    // Helper function to validate POP status
    function validatePOP(address account) internal view returns (bool) {
        POPNFTStorage storage s = getPOPNFTStorage();
        return s.hasValidPOP[account];
    }
} 