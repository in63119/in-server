package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	rpc     *ethclient.Client
	chainID *big.Int
}

func Dial(ctx context.Context, rpcURL string) (*Client, error) {
	rpc, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial eth rpc: %w", err)
	}

	chainID, err := rpc.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get chain id: %w", err)
	}

	return &Client{rpc: rpc, chainID: chainID}, nil
}

func (c *Client) Close() {
	if c == nil || c.rpc == nil {
		return
	}
	c.rpc.Close()
}

func (c *Client) ChainID() *big.Int {
	if c == nil {
		return nil
	}
	return new(big.Int).Set(c.chainID)
}

func (c *Client) NewTransactorFromKey(hexKey string) (*bind.TransactOpts, error) {
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}
	key, err := parsePrivateKey(hexKey)
	if err != nil {
		return nil, err
	}
	return bind.NewKeyedTransactorWithChainID(key, c.chainID)
}

func parsePrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	key := strings.TrimSpace(hexKey)
	key = strings.TrimPrefix(key, "0x")
	if key == "" {
		return nil, fmt.Errorf("private key is empty")
	}
	return crypto.HexToECDSA(key)
}
