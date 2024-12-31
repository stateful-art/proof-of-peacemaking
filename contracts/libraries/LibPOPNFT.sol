// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/utils/Counters.sol";
import "./LibStorage.sol";

library LibPOPNFT {
    using Counters for Counters.Counter;

    bytes32 constant POPNFT_STORAGE_POSITION = keccak256("pop.standard.popnft.storage");

    struct POPNFTStorage {
        // Token name and symbol (immutable)
        string name;
        string symbol;
        
        // Core ERC721 storage
        mapping(uint256 => address) owners;
        mapping(address => uint256) balances;
        mapping(uint256 => string) tokenURIs;
        
        // Token ID management
        Counters.Counter tokenIdCounter;
        
        // Contract state
        bool initialized;
        bool paused;
        bool _notEntered;
        
        // POP verification data
        mapping(uint256 => uint256) tokenToExpressionId;
        mapping(uint256 => LibStorage.Acknowledgement[]) tokenToAcknowledgements;
        mapping(address => bool) hasValidPOP;
    }

    function popnftStorage() internal pure returns (POPNFTStorage storage ps) {
        bytes32 position = POPNFT_STORAGE_POSITION;
        assembly {
            ps.slot := position
        }
    }

    // Helper function to validate POP status
    function validatePOP(address account) internal view returns (bool) {
        POPNFTStorage storage ps = popnftStorage();
        return ps.hasValidPOP[account];
    }
} 