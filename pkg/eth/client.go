package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"in-server/pkg/config"
	appcrypto "in-server/pkg/crypto"
	"in-server/pkg/firebase"
	"in-server/pkg/types"
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
	return gethcrypto.HexToECDSA(key)
}

type Accounts struct {
	Owner    string
	Relayer  string
	Relayer2 string
	Relayer3 string
}

func AccountsFromConfig(cfg config.Config) (Accounts, error) {
	salt := strings.TrimSpace(cfg.Auth.Hash)
	if salt == "" {
		return Accounts{}, fmt.Errorf("auth hash is empty")
	}

	decrypt := func(label, value string) (string, error) {
		out, err := appcrypto.Decrypt(value, salt)
		if err != nil {
			return "", fmt.Errorf("decrypt %s: %w", label, err)
		}
		return strings.TrimSpace(out), nil
	}

	owner, err := decrypt("owner", cfg.Blockchain.PrivateKey.Owner)
	if err != nil {
		return Accounts{}, err
	}

	relayer := strings.TrimSpace(cfg.Blockchain.PrivateKey.Relayer)
	if relayer == "" {
		return Accounts{}, fmt.Errorf("relayer private key is empty")
	}

	relayer2, err := decrypt("relayer2", cfg.Blockchain.PrivateKey.Relayer2)
	if err != nil {
		return Accounts{}, err
	}

	relayer3, err := decrypt("relayer3", cfg.Blockchain.PrivateKey.Relayer3)
	if err != nil {
		return Accounts{}, err
	}

	return Accounts{
		Owner:    owner,
		Relayer:  relayer,
		Relayer2: relayer2,
		Relayer3: relayer3,
	}, nil
}

func (c *Client) ReadyRelayer(ctx context.Context, fb *firebase.Client, cfg config.Config) (*bind.TransactOpts, error) {
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}
	accts, err := AccountsFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	relayerMap, exists, err := firebase.Read[map[string]types.FirebaseRelayer](ctx, fb, "relayers")
	if err != nil {
		return nil, fmt.Errorf("read firebase relayers: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("no relayer data found in firebase")
	}

	candidates := []string{accts.Relayer, accts.Relayer2, accts.Relayer3}

	for _, pk := range candidates {
		if pk == "" {
			continue
		}
		addr := strings.ToLower(addressFromPrivateKey(pk).Hex())

		ready := false
		for _, entry := range relayerMap {
			if strings.ToLower(strings.TrimSpace(entry.Address)) == addr && entry.Status == types.RelayerStatusReady {
				ready = true
				break
			}
		}
		if ready {
			return c.NewTransactorFromKey(pk)
		}
	}

	return nil, fmt.Errorf("no available relayer marked Ready")
}

func addressFromPrivateKey(hexKey string) common.Address {
	key, err := parsePrivateKey(hexKey)
	if err != nil {
		return common.Address{}
	}
	return gethcrypto.PubkeyToAddress(key.PublicKey)
}
