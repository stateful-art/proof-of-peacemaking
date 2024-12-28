// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "../libraries/LibDiamond.sol";
import "../libraries/LibStorage.sol";
import "../libraries/LibPermissions.sol";

contract AcknowledgementFacet {
    using LibStorage for LibStorage.AppStorage;

    event AcknowledgementCreated(
        uint256 indexed expressionId,
        address indexed acknowledger,
        address indexed creator,
        string message,
        uint256 timestamp
    );

    event MediaContentAdded(
        uint256 indexed expressionId,
        address indexed acknowledger,
        string textContent,
        string audioContent,
        string videoContent,
        string imageContent
    );

    function createAcknowledgement(
        uint256 _expressionId,
        address _creator,
        string memory _message,
        string memory _textContent,
        string memory _audioContent,
        string memory _videoContent,
        string memory _imageContent
    ) external payable{
        LibStorage.AppStorage storage s = LibStorage.appStorage();
        
        // Check if user needs to pay gas
        if (!LibPermissions.isSubsidized(msg.sender, LibPermissions.ACKNOWLEDGEMENT_PERMISSION)) {
            require(msg.value >= s.acknowledgementGasCost, "Insufficient gas payment");
        }

        require(_expressionId < s.expressionCount, "Expression does not exist");
        LibStorage.Expression storage expression = s.expressions[_expressionId];
        
        require(!hasAcknowledged(_expressionId, msg.sender), "Already acknowledged");
        require(msg.sender != _creator, "Cannot acknowledge own expression");

        LibStorage.Acknowledgement storage ack = expression.acknowledgments[msg.sender];
        ack.acknowledger = msg.sender;
        ack.timestamp = block.timestamp;
        ack.message = _message;
        ack.content = LibStorage.MediaContent({
            textContent: _textContent,
            audioContent: _audioContent,
            videoContent: _videoContent,
            imageContent: _imageContent
        });

        expression.acknowledgers.push(msg.sender);

        emit AcknowledgementCreated(
            _expressionId,
            msg.sender,
            _creator,
            _message,
            block.timestamp
        );

        emit MediaContentAdded(
            _expressionId,
            msg.sender,
            _textContent,
            _audioContent,
            _videoContent,
            _imageContent
        );
    }

    function hasAcknowledged(uint256 _expressionId, address _acknowledger) public view returns (bool) {
        LibStorage.AppStorage storage s = LibStorage.appStorage();
        require(_expressionId < s.expressionCount, "Expression does not exist");
        return s.expressions[_expressionId].acknowledgments[_acknowledger].acknowledger == _acknowledger;
    }

    function getAcknowledgement(
        uint256 _expressionId,
        address _acknowledger
    ) external view returns (
        uint256 timestamp,
        string memory message,
        LibStorage.MediaContent memory content,
        string memory ipfsHash
    ) {
        LibStorage.AppStorage storage s = LibStorage.appStorage();
        require(_expressionId < s.expressionCount, "Expression does not exist");
        
        LibStorage.Acknowledgement storage ack = s.expressions[_expressionId].acknowledgments[_acknowledger];
        require(ack.acknowledger == _acknowledger, "No acknowledgment from this address");
        
        return (
            ack.timestamp,
            ack.message,
            ack.content,
            ack.ipfsHash
        );
    }

    function getAcknowledgers(uint256 _expressionId) external view returns (address[] memory) {
        LibStorage.AppStorage storage s = LibStorage.appStorage();
        require(_expressionId < s.expressionCount, "Expression does not exist");
        return s.expressions[_expressionId].acknowledgers;
    }
} 