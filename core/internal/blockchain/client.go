package blockchain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Minimal ABI just for the anchorHash function
const contractABI = `[{"inputs":[{"internalType":"string","name":"eventId","type":"string"},{"internalType":"string","name":"eventHash","type":"string"}],"name":"anchorHash","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

type Client struct {
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
	contract   common.Address
	chainID    *big.Int
	parsedABI  abi.ABI
}

// NewClient initializes a new blockchain client.
func NewClient(rpcURL, privKeyHex, contractAddress string) (*Client, error) {
	if rpcURL == "" || privKeyHex == "" || contractAddress == "" {
		return nil, errors.New("missing blockchain configuration")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to EVM RPC: %w", err)
	}

	privKeyHex = strings.TrimPrefix(privKeyHex, "0x")
	privateKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &Client{
		client:     client,
		privateKey: privateKey,
		contract:   common.HexToAddress(contractAddress),
		chainID:    chainID,
		parsedABI:  parsedABI,
	}, nil
}

// AnchorHash sends a transaction to anchor the event to the blockchain.
func (c *Client) AnchorHash(ctx context.Context, eventID, eventHash string) (string, error) {
	publicKey := c.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := c.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to suggest gas price: %w", err)
	}

	// Pack the arguments for the anchorHash function
	data, err := c.parsedABI.Pack("anchorHash", eventID, eventHash)
	if err != nil {
		return "", fmt.Errorf("failed to pack arguments: %w", err)
	}

	// Estimate gas
	msg := ethereum.CallMsg{
		From:     fromAddress,
		To:       &c.contract,
		GasPrice: gasPrice,
		Data:     data,
	}
	gasLimit, err := c.client.EstimateGas(ctx, msg)
	if err != nil {
		log.Printf("Blockchain: EstimateGas failed, using fallback: %v", err)
		gasLimit = 300000 // Fallback gas limit
	} else {
		gasLimit = uint64(float64(gasLimit) * 1.2) // Add 20% buffer
	}

	// Create transaction
	tx := types.NewTransaction(nonce, c.contract, big.NewInt(0), gasLimit, gasPrice, data)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Broadcast transaction
	err = c.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}
