// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./interfaces/IDiamondCut.sol";
import "./interfaces/IDiamondLoupe.sol";
import "./libraries/LibDiamond.sol";

contract Diamond {    
    constructor(address _contractOwner) {
        LibDiamond.setContractOwner(_contractOwner);
    }

    // Find facet for function that is called and execute the
    // function if a facet is found and return any value.
    fallback() external payable {
        LibDiamond.DiamondStorage storage ds = LibDiamond.diamondStorage();
        address facet = ds.selectorToFacet[msg.sig];
        require(facet != address(0), "Diamond: Function does not exist");
        
        assembly {
            calldatacopy(0, 0, calldatasize())
            let result := delegatecall(gas(), facet, 0, calldatasize(), 0, 0)
            returndatacopy(0, 0, returndatasize())
            switch result
                case 0 {revert(0, returndatasize())}
                default {return (0, returndatasize())}
        }
    }

    receive() external payable {}
} 