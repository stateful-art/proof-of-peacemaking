// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "../libraries/LibDiamond.sol";
import "../libraries/LibPermissions.sol";

contract PermissionsFacet {
    event SubsidyStatusChanged(address operator, address user, uint8 operation, bool status);
    event OperatorStatusChanged(address operator, bool active);

    function setOperatorSubsidies(
        address[] calldata users,
        uint8[] calldata operations,
        bool[] calldata statuses
    ) external {
        LibDiamond.enforceIsContractOwner();
        require(users.length == operations.length && operations.length == statuses.length, "Length mismatch");
        
        for(uint i = 0; i < users.length; i++) {
            LibPermissions.setOperatorSubsidy(msg.sender, users[i], operations[i], statuses[i]);
            emit SubsidyStatusChanged(msg.sender, users[i], operations[i], statuses[i]);
        }
    }

    function setOperatorStatus(address operator, bool active) external {
        LibDiamond.enforceIsContractOwner();
        LibPermissions.setOperatorStatus(operator, active);
        emit OperatorStatusChanged(operator, active);
    }
} 