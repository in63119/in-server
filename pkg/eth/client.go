package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"in-server/pkg/abis"
	"in-server/pkg/apperr"
	"in-server/pkg/config"
	appcrypto "in-server/pkg/crypto"
	"in-server/pkg/firebase"
	"in-server/pkg/types"
)

type Client struct {
	rpc     *ethclient.Client
	chainID *big.Int
	cfg     config.Config
}

func Dial(ctx context.Context, cfg config.Config) (*Client, error) {
	rpcURL := strings.TrimSpace("https://public-en-kairos.node.kaia.io")
	if rpcURL == "" {
		return nil, apperr.Blockchain.ErrRPCURLMissing
	}

	rpc, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial eth rpc: %w", err)
	}

	chainID, err := rpc.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get chain id: %w", err)
	}

	return &Client{rpc: rpc, chainID: chainID, cfg: cfg}, nil
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

func (c *Client) Accounts() (Accounts, error) {
	return accountsFromConfig(c.cfg)
}

func accountsFromConfig(cfg config.Config) (Accounts, error) {
	salt := strings.TrimSpace(cfg.Auth.Hash)
	if salt == "" {
		return Accounts{}, apperr.System.ErrMissingAuthHash
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

func (c *Client) ReadyRelayer(ctx context.Context, fb *firebase.Client) (*bind.TransactOpts, error) {
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}
	accts, err := c.Accounts()
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

	return nil, apperr.Blockchain.ErrNoAvailableRelayer
}

func addressFromPrivateKey(hexKey string) common.Address {
	key, err := parsePrivateKey(hexKey)
	if err != nil {
		return common.Address{}
	}
	return gethcrypto.PubkeyToAddress(key.PublicKey)
}

func (c *Client) Contract(name types.ContractName) (*bind.BoundContract, common.Address, error) {
	if c == nil || c.rpc == nil {
		return nil, common.Address{}, fmt.Errorf("client is nil")
	}

	artifacts, err := abis.Get(c.cfg.Env)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("load abis: %w", err)
	}

	artifact, ok := artifacts[name]
	if !ok {
		return nil, common.Address{}, fmt.Errorf("%w: %s", apperr.Blockchain.ErrContractNotFound, name)
	}

	addr := common.HexToAddress(artifact.Address)
	if addr == (common.Address{}) {
		return nil, common.Address{}, fmt.Errorf("%w: %s address empty", apperr.Blockchain.ErrContractNotFound, name)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(artifact.ABI)))
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("parse abi for %s: %w", name, err)
	}

	bound := bind.NewBoundContract(addr, parsedABI, c.rpc, c.rpc, c.rpc)
	return bound, addr, nil
}

func (c *Client) Wallet(email string) (*ecdsa.PrivateKey, common.Address, error) {
	salt := strings.TrimSpace(c.cfg.Auth.Hash)
	if salt == "" {
		return nil, common.Address{}, apperr.System.ErrMissingAuthHash
	}

	input := []byte(email + salt)
	digest := gethcrypto.Keccak256(input)

	pk, err := gethcrypto.ToECDSA(digest)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("to ecdsa: %w", err)
	}
	return pk, gethcrypto.PubkeyToAddress(pk.PublicKey), nil
}
