// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

library LibPermissions {
    bytes32 constant STORAGE_POSITION = keccak256("permissions.storage");
    
    uint8 constant EXPRESSION_PERMISSION = 1;
    uint8 constant ACKNOWLEDGEMENT_PERMISSION = 2;
    uint8 constant NFT_PERMISSION = 4;
    
    struct PermissionStorage {
        // Operator => User => Operation => Is Subsidized
        mapping(address => mapping(address => mapping(uint8 => bool))) operatorSubsidies;
        // Operator => Is Active
        mapping(address => bool) activeOperators;
    }
    
    function permissionStorage() internal pure returns (PermissionStorage storage ps) {
        bytes32 position = STORAGE_POSITION;
        assembly {
            ps.slot := position
        }
    }
    
    function isSubsidized(address user, uint8 operation) internal view returns (bool) {
        PermissionStorage storage ps = permissionStorage();
        return ps.activeOperators[msg.sender] && ps.operatorSubsidies[msg.sender][user][operation];
    }
    
    function setOperatorSubsidy(
        address operator,
        address user,
        uint8 operation,
        bool status
    ) internal {
        PermissionStorage storage ps = permissionStorage();
        ps.operatorSubsidies[operator][user][operation] = status;
    }

    function setOperatorStatus(address operator, bool active) internal {
        PermissionStorage storage ps = permissionStorage();
        ps.activeOperators[operator] = active;
    }
} 