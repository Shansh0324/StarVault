const fs = require('fs');
const path = require('path');
const solc = require('solc');
const { ethers } = require('ethers');
require('dotenv').config({ path: path.resolve(__dirname, '../.env') });

async function main() {
    console.log("Starting deployment...");

    const contractPath = path.resolve(__dirname, 'contracts', 'StarVaultLedger.sol');
    const sourceCode = fs.readFileSync(contractPath, 'utf8');

    // Prepare solc input
    const input = {
        language: 'Solidity',
        sources: {
            'StarVaultLedger.sol': {
                content: sourceCode
            }
        },
        settings: {
            outputSelection: {
                '*': {
                    '*': ['*']
                }
            }
        }
    };

    console.log("Compiling contract...");
    const output = JSON.parse(solc.compile(JSON.stringify(input)));

    if (output.errors) {
        output.errors.forEach(err => {
            console.error(err.formattedMessage);
        });
        if (output.errors.some(e => e.severity === 'error')) {
            process.exit(1);
        }
    }

    const contract = output.contracts['StarVaultLedger.sol']['StarVaultLedger'];
    const abi = contract.abi;
    const bytecode = contract.evm.bytecode.object;

    // Connect to blockchain
    const rpcUrl = process.env.BLOCKCHAIN_RPC_URL || 'https://ethereum-sepolia-rpc.publicnode.com';
    const privateKey = process.env.BLOCKCHAIN_PRIVATE_KEY;

    if (!privateKey) {
        console.error("Missing BLOCKCHAIN_PRIVATE_KEY environment variable.");
        process.exit(1);
    }

    const provider = new ethers.JsonRpcProvider(rpcUrl);
    const wallet = new ethers.Wallet(privateKey, provider);

    console.log(`Deploying from account: ${wallet.address}`);

    const factory = new ethers.ContractFactory(abi, bytecode, wallet);
    
    try {
        const deployedContract = await factory.deploy();
        console.log("Transaction broadcast. Waiting for confirmation...");
        
        await deployedContract.waitForDeployment();
        const address = await deployedContract.getAddress();
        
        console.log(`\n==============================================`);
        console.log(`[OK] StarVaultLedger deployed to: ${address}`);
        console.log(`==============================================\n`);
        console.log(`Please update SMART_CONTRACT_ADDRESS in your .env file with this address.`);
    } catch (err) {
        console.error("Deployment failed:", err);
    }
}

main().catch(console.error);
