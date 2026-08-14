// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title StarVaultLedger
 * @dev A minimal, gas-efficient smart contract designed solely to emit audit events.
 * It does not store the hashes in contract state, which saves significant gas.
 * The logs act as a tamper-proof cryptographic proof.
 */
contract StarVaultLedger {
    // Event emitted when a new hash is anchored
    event HashAnchored(string eventId, string eventHash, uint256 timestamp);

    /**
     * @dev Anchors a cryptographic hash on the blockchain.
     * @param eventId The unique identifier of the audit event.
     * @param eventHash The SHA-256 hash of the audit payload.
     */
    function anchorHash(string memory eventId, string memory eventHash) public {
        emit HashAnchored(eventId, eventHash, block.timestamp);
    }
}
