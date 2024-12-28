// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract Expression {
    struct MediaContent {
        string textContent;      // IPFS hash for text content
        string audioContent;     // IPFS hash for audio content
        string videoContent;     // IPFS hash for video content
        string imageContent;     // IPFS hash for image content
    }

    struct Acknowledgment {
        address acknowledger;
        uint256 timestamp;
        string message;          // Optional acknowledgment message
        string ipfsHash;         // For any additional documents
    }

    struct PeaceExpression {
        address creator;
        MediaContent content;
        uint256 timestamp;
        mapping(address => Acknowledgment) acknowledgments;
        address[] acknowledgers; // Array to track all acknowledgers
        string ipfsHash;        // For storing additional data/documents
    }

    // Mapping from expression ID to PeaceExpression
    mapping(uint256 => PeaceExpression) public expressions;
    uint256 public expressionCount;

    // Events
    event ExpressionCreated(
        uint256 indexed expressionId,
        address indexed creator,
        string textContent,
        string audioContent,
        string videoContent,
        string imageContent,
        uint256 timestamp
    );

    event ExpressionAcknowledged(
        uint256 indexed expressionId,
        address indexed acknowledger,
        string message,
        uint256 timestamp
    );

    event DocumentAdded(
        uint256 indexed expressionId,
        address indexed adder,
        string ipfsHash
    );

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

    // Create a new expression of peace with multimedia content
    function createExpression(
        string memory _textContent,
        string memory _audioContent,
        string memory _videoContent,
        string memory _imageContent
    ) public onlyProxyOrSender(msg.sender) returns (uint256) {
        uint256 expressionId = expressionCount++;
        
        PeaceExpression storage newExpression = expressions[expressionId];
        newExpression.creator = msg.sender;
        newExpression.timestamp = block.timestamp;
        newExpression.content = MediaContent({
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

    // Acknowledge an expression of peace with optional message
    function acknowledgeExpression(uint256 _expressionId, string memory _message) public {
        require(_expressionId < expressionCount, "Expression does not exist");
        PeaceExpression storage expression = expressions[_expressionId];
        
        require(expression.creator != msg.sender, "Cannot acknowledge own expression");
        require(expression.acknowledgments[msg.sender].acknowledger == address(0), 
                "Already acknowledged by this address");

        Acknowledgment memory newAck = Acknowledgment({
            acknowledger: msg.sender,
            timestamp: block.timestamp,
            message: _message,
            ipfsHash: ""
        });

        expression.acknowledgments[msg.sender] = newAck;
        expression.acknowledgers.push(msg.sender);

        emit ExpressionAcknowledged(_expressionId, msg.sender, _message, block.timestamp);
    }

    // Add IPFS document hash to an acknowledgment
    function addAcknowledgmentDocument(
        uint256 _expressionId, 
        string memory _ipfsHash
    ) public {
        require(_expressionId < expressionCount, "Expression does not exist");
        PeaceExpression storage expression = expressions[_expressionId];
        
        require(expression.acknowledgments[msg.sender].acknowledger == msg.sender, 
                "Must have acknowledged first");

        expression.acknowledgments[msg.sender].ipfsHash = _ipfsHash;
        emit DocumentAdded(_expressionId, msg.sender, _ipfsHash);
    }

    // Get expression details
    function getExpression(uint256 _expressionId) public view returns (
        address creator,
        MediaContent memory content,
        uint256 timestamp,
        address[] memory acknowledgersList,
        string memory ipfsHash
    ) {
        require(_expressionId < expressionCount, "Expression does not exist");
        PeaceExpression storage expression = expressions[_expressionId];
        
        return (
            expression.creator,
            expression.content,
            expression.timestamp,
            expression.acknowledgers,
            expression.ipfsHash
        );
    }

    // Get acknowledgment details
    function getAcknowledgment(uint256 _expressionId, address _acknowledger) public view returns (
        uint256 timestamp,
        string memory message,
        string memory ipfsHash
    ) {
        require(_expressionId < expressionCount, "Expression does not exist");
        PeaceExpression storage expression = expressions[_expressionId];
        Acknowledgment memory ack = expression.acknowledgments[_acknowledger];
        
        require(ack.acknowledger != address(0), "No acknowledgment from this address");
        
        return (
            ack.timestamp,
            ack.message,
            ack.ipfsHash
        );
    }

    // Get all acknowledgments for an expression
    function getAllAcknowledgments(uint256 _expressionId) public view returns (
        address[] memory acknowledgers,
        uint256[] memory timestamps,
        string[] memory messages
    ) {
        require(_expressionId < expressionCount, "Expression does not exist");
        PeaceExpression storage expression = expressions[_expressionId];
        
        uint256 count = expression.acknowledgers.length;
        timestamps = new uint256[](count);
        messages = new string[](count);
        
        for (uint256 i = 0; i < count; i++) {
            address acknowledger = expression.acknowledgers[i];
            Acknowledgment memory ack = expression.acknowledgments[acknowledger];
            timestamps[i] = ack.timestamp;
            messages[i] = ack.message;
        }
        
        return (expression.acknowledgers, timestamps, messages);
    }

    // Get expressions by creator
    function getExpressionsByCreator(address _creator) public view returns (uint256[] memory) {
        uint256[] memory result = new uint256[](expressionCount);
        uint256 count = 0;
        
        for (uint256 i = 0; i < expressionCount; i++) {
            if (expressions[i].creator == _creator) {
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