const { ethers } = require("ethers");
const fs = require('fs');
const path = require('path');

function main() {
    console.log("Generating a fresh Testnet Wallet for StarVault Ledger...");
    const wallet = ethers.Wallet.createRandom();
    
    console.log("\n=========================================================");
    console.log(`Address:     ${wallet.address}`);
    console.log(`Private Key: ${wallet.privateKey}`);
    console.log("=========================================================\n");
    
    console.log("Next Steps:");
    console.log("1. Go to a Sepolia Faucet (e.g. https://sepoliafaucet.com/ or Alchemy Faucet).");
    console.log("2. Enter the Address above to receive free Sepolia ETH.");
    console.log("3. Copy the Private Key and put it in your `.env` as BLOCKCHAIN_PRIVATE_KEY.");
    console.log("4. Once funded, run: npm run deploy");
}

main();
