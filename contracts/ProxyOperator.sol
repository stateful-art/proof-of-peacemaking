// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "./Expression.sol";
import "./Acknowledgement.sol";
import "./POPNFT.sol";

contract ProxyOperator is Ownable {
    using ECDSA for bytes32;

    Expression public expressionContract;
    Acknowledgement public acknowledgementContract;
    POPNFT public nftContract;

    // Nonce mapping to prevent replay attacks
    mapping(address => uint256) public nonces;

    // Mapping to track if an operator is approved
    mapping(address => bool) public approvedOperators;

    event OperatorAdded(address operator);
    event OperatorRemoved(address operator);
    event TransactionExecuted(address user, bytes32 txHash);

    constructor(
        address _expressionContract,
        address _acknowledgementContract,
        address _nftContract
    ) {
        expressionContract = Expression(_expressionContract);
        acknowledgementContract = Acknowledgement(_acknowledgementContract);
        nftContract = POPNFT(_nftContract);
    }

    // Modifier to check if operator is approved
    modifier onlyOperator() {
        require(approvedOperators[msg.sender], "Not an approved operator");
        _;
    }

    // Add or remove operators
    function setOperator(address operator, bool approved) external onlyOwner {
        approvedOperators[operator] = approved;
        if (approved) {
            emit OperatorAdded(operator);
        } else {
            emit OperatorRemoved(operator);
        }
    }

    // Create expression with meta transaction
    function createExpressionFor(
        address user,
        string memory textContent,
        string memory audioContent,
        string memory videoContent,
        string memory imageContent,
        bytes memory signature
    ) external onlyOperator {
        bytes32 hash = keccak256(abi.encodePacked(
            "createExpression",
            user,
            textContent,
            audioContent,
            videoContent,
            imageContent,
            nonces[user]++
        ));
        require(hash.toEthSignedMessageHash().recover(signature) == user, "Invalid signature");

        expressionContract.createExpression(
            textContent,
            audioContent,
            videoContent,
            imageContent
        );

        emit TransactionExecuted(user, hash);
    }

    // Create acknowledgement with meta transaction
    function createAcknowledgementFor(
        address user,
        uint256 expressionId,
        address creator,
        string memory message,
        bytes memory signature
    ) external onlyOperator {
        bytes32 hash = keccak256(abi.encodePacked(
            "createAcknowledgement",
            user,
            expressionId,
            creator,
            message,
            nonces[user]++
        ));
        require(hash.toEthSignedMessageHash().recover(signature) == user, "Invalid signature");

        acknowledgementContract.createAcknowledgement(
            expressionId,
            creator,
            message
        );

        emit TransactionExecuted(user, hash);
    }

    // Mint NFTs with dual signatures
    function mintProofsWithSignatures(
        uint256 expressionId,
        uint256 acknowledgementId,
        address creator,
        address acknowledger,
        string memory expressionIPFS,
        string memory acknowledgementIPFS,
        bytes memory creatorSignature,
        bytes memory acknowledgerSignature
    ) external onlyOperator {
        // Verify creator's signature
        bytes32 creatorHash = keccak256(abi.encodePacked(
            "mintProof",
            expressionId,
            acknowledgementId,
            creator,
            acknowledger,
            expressionIPFS,
            acknowledgementIPFS,
            nonces[creator]++
        ));
        require(creatorHash.toEthSignedMessageHash().recover(creatorSignature) == creator, 
                "Invalid creator signature");

        // Verify acknowledger's signature
        bytes32 acknowledgerHash = keccak256(abi.encodePacked(
            "mintProof",
            expressionId,
            acknowledgementId,
            creator,
            acknowledger,
            expressionIPFS,
            acknowledgementIPFS,
            nonces[acknowledger]++
        ));
        require(acknowledgerHash.toEthSignedMessageHash().recover(acknowledgerSignature) == acknowledger, 
                "Invalid acknowledger signature");

        // Mint NFTs
        nftContract.mintProofs(
            expressionId,
            acknowledgementId,
            creator,
            acknowledger,
            expressionIPFS,
            acknowledgementIPFS
        );

        emit TransactionExecuted(creator, creatorHash);
        emit TransactionExecuted(acknowledger, acknowledgerHash);
    }

    // Get current nonce for a user
    function getNonce(address user) external view returns (uint256) {
        return nonces[user];
    }
} 