// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "../libraries/LibDiamond.sol";
import "../libraries/LibStorage.sol";
import "../libraries/LibPermissions.sol";

contract ExpressionFacet {
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
        LibStorage.ExpressionStorage storage es = LibStorage.expressionStorage();
        LibStorage.GasCostStorage storage gs = LibStorage.gasCostStorage();
        
        // Check if user needs to pay gas
        if (!LibPermissions.isSubsidized(msg.sender, LibPermissions.EXPRESSION_PERMISSION)) {
            require(msg.value >= gs.expressionGasCost, "Insufficient gas payment");
        }

        uint256 expressionId = es.expressionCount++;
        LibStorage.Expression storage expression = es.expressions[expressionId];
        
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
        LibStorage.ExpressionStorage storage es = LibStorage.expressionStorage();
        require(_expressionId < es.expressionCount, "Expression does not exist");
        LibStorage.Expression storage expression = es.expressions[_expressionId];
        
        return (
            expression.creator,
            expression.content,
            expression.timestamp,
            expression.acknowledgers,
            expression.ipfsHash
        );
    }

    function getExpressionsByCreator(address _creator) external view returns (uint256[] memory) {
        LibStorage.ExpressionStorage storage es = LibStorage.expressionStorage();
        uint256[] memory result = new uint256[](es.expressionCount);
        uint256 count = 0;
        
        for (uint256 i = 0; i < es.expressionCount; i++) {
            if (es.expressions[i].creator == _creator) {
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
} 