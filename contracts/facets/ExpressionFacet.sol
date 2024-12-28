// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "../libraries/LibDiamond.sol";
import "../libraries/LibStorage.sol";
import "../libraries/LibPermissions.sol";

contract ExpressionFacet {
    using LibStorage for LibStorage.AppStorage;

    event ExpressionCreated(
        uint256 indexed expressionId,
        address indexed creator,
        string textContent,
        string audioContent,
        string videoContent,
        string imageContent,
        uint256 timestamp
    );

    function createExpression(
        string memory _textContent,
        string memory _audioContent,
        string memory _videoContent,
        string memory _imageContent
    ) external payable returns (uint256) {
        LibStorage.AppStorage storage s = LibStorage.appStorage();
        
        // Check if user needs to pay gas
        if (!LibPermissions.isSubsidized(msg.sender, LibPermissions.EXPRESSION_PERMISSION)) {
            require(msg.value >= s.expressionGasCost, "Insufficient gas payment");
        }

        uint256 expressionId = s.expressionCount++;
        LibStorage.Expression storage expression = s.expressions[expressionId];
        
        expression.creator = msg.sender;
        expression.timestamp = block.timestamp;
        expression.content = LibStorage.MediaContent({
            textContent: _textContent,
            audioContent: _audioContent,
            videoContent: _videoContent,
            imageContent: _imageContent
        });

        emit ExpressionCreated(
            expressionId,
            msg.sender,
            _textContent,
            _audioContent,
            _videoContent,
            _imageContent,
            block.timestamp
        );

        return expressionId;
    }

    function getExpression(uint256 _expressionId) external view returns (
        address creator,
        LibStorage.MediaContent memory content,
        uint256 timestamp,
        address[] memory acknowledgersList,
        string memory ipfsHash
    ) {
        LibStorage.AppStorage storage s = LibStorage.getStorage();
        require(_expressionId < s.expressionCount, "Expression does not exist");
        LibStorage.Expression storage expression = s.expressions[_expressionId];
        
        return (
            expression.creator,
            expression.content,
            expression.timestamp,
            expression.acknowledgers,
            expression.ipfsHash
        );
    }

    function getExpressionsByCreator(address _creator) external view returns (uint256[] memory) {
        LibStorage.AppStorage storage s = LibStorage.getStorage();
        uint256[] memory result = new uint256[](s.expressionCount);
        uint256 count = 0;
        
        for (uint256 i = 0; i < s.expressionCount; i++) {
            if (s.expressions[i].creator == _creator) {
                result[count] = i;
                count++;
            }
        }
        
        // Resize array to actual count
        assembly {
            mstore(result, count)
        }
        
        return result;
    }

    // Add other expression functions...
} 