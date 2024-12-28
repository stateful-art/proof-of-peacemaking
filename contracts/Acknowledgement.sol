// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract Acknowledgement {
    struct AcknowledgementData {
        uint256 expressionId;    // Reference to the original expression
        address acknowledger;    // Who acknowledged
        address creator;         // Original expression creator
        uint256 timestamp;
        string message;          // Acknowledgment message
        MediaContent content;    // Optional media response
        string ipfsHash;         // Additional documents
    }

    struct MediaContent {
        string textContent;      // IPFS hash for text content
        string audioContent;     // IPFS hash for audio content
        string videoContent;     // IPFS hash for video content
        string imageContent;     // IPFS hash for image content
    }

    // Mapping from acknowledgment ID to AcknowledgementData
    mapping(uint256 => AcknowledgementData) public acknowledgements;
    uint256 public acknowledgementCount;

    // Mapping from expression ID to acknowledgment IDs
    mapping(uint256 => uint256[]) public expressionAcknowledgements;
    
    // Mapping to track if an address has acknowledged an expression
    mapping(uint256 => mapping(address => bool)) public hasAcknowledged;

    // Add proxy operator
    address public proxyOperator;

    // Modifier to allow proxy operator
    modifier onlyProxyOrSender(address user) {
        require(msg.sender == user || msg.sender == proxyOperator, "Not authorized");
        _;
    }

    // Set proxy operator
    function setProxyOperator(address _proxyOperator) external onlyOwner {
        proxyOperator = _proxyOperator;
    }

    // Events
    event AcknowledgementCreated(
        uint256 indexed acknowledgementId,
        uint256 indexed expressionId,
        address indexed acknowledger,
        address creator,
        string message,
        uint256 timestamp
    );

    event MediaContentAdded(
        uint256 indexed acknowledgementId,
        string textContent,
        string audioContent,
        string videoContent,
        string imageContent
    );

    event DocumentAdded(
        uint256 indexed acknowledgementId,
        string ipfsHash
    );

    // Create a new acknowledgement
    function createAcknowledgement(
        uint256 _expressionId,
        address _creator,
        string memory _message
    ) public onlyProxyOrSender(msg.sender) returns (uint256) {
        require(!hasAcknowledged[_expressionId][msg.sender], "Already acknowledged this expression");
        require(msg.sender != _creator, "Cannot acknowledge own expression");

        uint256 acknowledgementId = acknowledgementCount++;
        
        acknowledgements[acknowledgementId] = AcknowledgementData({
            expressionId: _expressionId,
            acknowledger: msg.sender,
            creator: _creator,
            timestamp: block.timestamp,
            message: _message,
            content: MediaContent({
                textContent: "",
                audioContent: "",
                videoContent: "",
                imageContent: ""
            }),
            ipfsHash: ""
        });

        expressionAcknowledgements[_expressionId].push(acknowledgementId);
        hasAcknowledged[_expressionId][msg.sender] = true;

        emit AcknowledgementCreated(
            acknowledgementId,
            _expressionId,
            msg.sender,
            _creator,
            _message,
            block.timestamp
        );

        return acknowledgementId;
    }

    // Add media content to acknowledgement
    function addMediaContent(
        uint256 _acknowledgementId,
        string memory _textContent,
        string memory _audioContent,
        string memory _videoContent,
        string memory _imageContent
    ) public onlyProxyOrSender(msg.sender) {
        require(_acknowledgementId < acknowledgementCount, "Acknowledgement does not exist");
        AcknowledgementData storage ack = acknowledgements[_acknowledgementId];
        require(msg.sender == ack.acknowledger, "Only acknowledger can add content");

        ack.content = MediaContent({
            textContent: _textContent,
            audioContent: _audioContent,
            videoContent: _videoContent,
            imageContent: _imageContent
        });

        emit MediaContentAdded(
            _acknowledgementId,
            _textContent,
            _audioContent,
            _videoContent,
            _imageContent
        );
    }

    // Add document to acknowledgement
    function addDocument(
        uint256 _acknowledgementId, 
        string memory _ipfsHash
    ) public onlyProxyOrSender(msg.sender) {
        require(_acknowledgementId < acknowledgementCount, "Acknowledgement does not exist");
        AcknowledgementData storage ack = acknowledgements[_acknowledgementId];
        require(msg.sender == ack.acknowledger, "Only acknowledger can add documents");

        ack.ipfsHash = _ipfsHash;
        emit DocumentAdded(_acknowledgementId, _ipfsHash);
    }

    // Get acknowledgement details
    function getAcknowledgement(uint256 _acknowledgementId) public view returns (
        uint256 expressionId,
        address acknowledger,
        address creator,
        uint256 timestamp,
        string memory message,
        MediaContent memory content,
        string memory ipfsHash
    ) {
        require(_acknowledgementId < acknowledgementCount, "Acknowledgement does not exist");
        AcknowledgementData storage ack = acknowledgements[_acknowledgementId];
        
        return (
            ack.expressionId,
            ack.acknowledger,
            ack.creator,
            ack.timestamp,
            ack.message,
            ack.content,
            ack.ipfsHash
        );
    }

    // Get all acknowledgements for an expression
    function getExpressionAcknowledgements(uint256 _expressionId) public view returns (uint256[] memory) {
        return expressionAcknowledgements[_expressionId];
    }
} 