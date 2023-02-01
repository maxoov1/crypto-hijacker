package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	*ethclient.Client
	rpc *rpc.Client
}

func New(ctx context.Context, endpoint string) (*Client, error) {
	rpc, err := rpc.DialContext(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	return &Client{ethclient.NewClient(rpc), rpc}, nil
}

func (c *Client) SubscribeToPendingTransactions(ctx context.Context) (*rpc.ClientSubscription, <-chan common.Hash, error) {
	hashes := make(chan common.Hash)

	subscription, err := c.rpc.EthSubscribe(ctx, hashes, "newPendingTransactions")
	if err != nil {
		return nil, nil, err
	}

	return subscription, hashes, nil
}

func (c *Client) SendNewTransaction(
	ctx context.Context, sender, recipient common.Address, amount *big.Int, privateKey *ecdsa.PrivateKey,
) (*types.Transaction, error) {
	nonce, err := c.PendingNonceAt(ctx, sender)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending nonce: %w", err)
	}

	gasPrice, err := c.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price suggestion: %w", err)
	}

	// value - gas * price
	amount = amount.Sub(amount, gasPrice)

	transaction := types.NewTransaction(
		nonce, recipient, amount, 21000, gasPrice, nil,
	)

	chainID, err := c.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain id: %w", err)
	}

	signedTransaction, err := types.SignTx(
		transaction, types.NewEIP155Signer(chainID), privateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	if err := c.SendTransaction(ctx, signedTransaction); err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTransaction, nil
}
