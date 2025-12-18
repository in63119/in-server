package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

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

	relayer, err := decrypt("relayer", cfg.Blockchain.PrivateKey.Relayer)
	if err != nil {
		return Accounts{}, err
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

func (c *Client) ReadyRelayer(ctx context.Context, fb *firebase.Client) (*bind.TransactOpts, common.Address, string, error) {
	if c == nil {
		return nil, common.Address{}, "", fmt.Errorf("client is nil")
	}
	accts, err := c.Accounts()
	if err != nil {
		return nil, common.Address{}, "", err
	}

	relayerMap, exists, err := firebase.Read[map[string]types.FirebaseRelayer](ctx, fb, "relayers")
	if err != nil {
		return nil, common.Address{}, "", fmt.Errorf("read firebase relayers: %w", err)
	}
	if !exists {
		return nil, common.Address{}, "", fmt.Errorf("no relayer data found in firebase")
	}

	candidates := []struct {
		pk   string
		addr common.Address
	}{
		{accts.Relayer, addressFromPrivateKey(accts.Relayer)},
		{accts.Relayer2, addressFromPrivateKey(accts.Relayer2)},
		{accts.Relayer3, addressFromPrivateKey(accts.Relayer3)},
	}

	for _, cand := range candidates {
		if cand.pk == "" || (cand.addr == common.Address{}) {
			continue
		}
		addrLower := strings.ToLower(cand.addr.Hex())
		for key, entry := range relayerMap {
			if strings.ToLower(strings.TrimSpace(entry.Address)) == addrLower && entry.Status == types.RelayerStatusReady {
				entry.Status = types.RelayerStatusProcessing
				relayerMap[key] = entry
				if err := firebase.Write(ctx, fb, "relayers", relayerMap); err != nil {
					return nil, common.Address{}, "", fmt.Errorf("update relayer status: %w", err)
				}

				opts, err := c.NewTransactorFromKey(cand.pk)
				if err != nil {
					return nil, common.Address{}, "", err
				}
				return opts, cand.addr, key, nil
			}
		}
	}

	return nil, common.Address{}, "", apperr.Blockchain.ErrNoAvailableRelayer
}

func addressFromPrivateKey(hexKey string) common.Address {
	key, err := parsePrivateKey(hexKey)
	if err != nil {
		return common.Address{}
	}
	return gethcrypto.PubkeyToAddress(key.PublicKey)
}

// AddressFromPrivateKey parses a hex private key and returns its address.
func AddressFromPrivateKey(hexKey string) (common.Address, error) {
	key, err := parsePrivateKey(hexKey)
	if err != nil {
		return common.Address{}, err
	}
	return gethcrypto.PubkeyToAddress(key.PublicKey), nil
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

func (c *Client) RPC() *ethclient.Client {
	if c == nil {
		return nil
	}
	return c.rpc
}

func (c *Client) SendTxByRelayer(ctx context.Context, fb *firebase.Client, contract *bind.BoundContract, method string, args ...any) (*gethtypes.Receipt, error) {
	if contract == nil {
		return nil, fmt.Errorf("contract is nil")
	}

	opts, addr, relayerKey, err := c.ReadyRelayer(ctx, fb)
	if err != nil {
		return nil, err
	}

	markReady := func() {
		if relayerKey == "" || fb == nil {
			return
		}
		relayerMap, exists, err := firebase.Read[map[string]types.FirebaseRelayer](ctx, fb, "relayers")
		if err != nil || !exists {
			return
		}
		entry, ok := relayerMap[relayerKey]
		if !ok {
			return
		}
		entry.Status = types.RelayerStatusReady
		relayerMap[relayerKey] = entry
		_ = firebase.Write(ctx, fb, "relayers", relayerMap)
	}
	defer markReady()

	tx, err := contract.Transact(opts, method, args...)
	if err != nil {
		return nil, err
	}

	receipt, err := bind.WaitMined(ctx, c.RPC(), tx)
	if err != nil {
		return nil, err
	}
	if receipt == nil || receipt.Status != gethtypes.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("tx %s status %v via relayer %s", tx.Hash().Hex(), receiptStatus(receipt), addr.Hex())
	}

	return receipt, nil
}

func receiptStatus(r *gethtypes.Receipt) any {
	if r == nil {
		return nil
	}
	return r.Status
}

type ForwardRequestData struct {
	From      common.Address `abi:"from"`
	To        common.Address `abi:"to"`
	Value     *big.Int       `abi:"value"`
	Gas       *big.Int       `abi:"gas"`
	Deadline  *big.Int       `abi:"deadline"`
	Data      []byte         `abi:"data"`
	Signature []byte         `abi:"signature"`
}

// Excute sends a meta-tx using the InForwarder via a READY relayer (and marks relayer status in Firebase).
// It matches the JS helper name used in the Next.js codebase (typo kept intentionally).
func (c *Client) Excute(ctx context.Context, fb *firebase.Client, recipient types.ContractName, signer *ecdsa.PrivateKey, method string, args ...any) (*gethtypes.Receipt, error) {
	return c.ExecuteMetaTx(ctx, fb, recipient, signer, method, args...)
}

func (c *Client) ExecuteMetaTx(ctx context.Context, fb *firebase.Client, recipient types.ContractName, signer *ecdsa.PrivateKey, method string, args ...any) (*gethtypes.Receipt, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c == nil || c.rpc == nil {
		return nil, fmt.Errorf("client is nil")
	}
	if signer == nil {
		return nil, fmt.Errorf("signer is nil")
	}

	forwarder, forwarderAddr, forwarderABI, err := c.contractWithABI(types.INFORWARDER)
	if err != nil {
		return nil, err
	}

	recipientAddr, recipientABI, err := c.contractABI(recipient)
	if err != nil {
		return nil, err
	}

	calldata, err := recipientABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("pack %s.%s: %w", recipient, method, err)
	}

	fromAddr := gethcrypto.PubkeyToAddress(signer.PublicKey)
	nonce, err := c.forwarderNonce(ctx, forwarderAddr, forwarderABI, fromAddr)
	if err != nil {
		return nil, err
	}
	domain, err := c.forwarderDomain(ctx, forwarderAddr, forwarderABI)
	if err != nil {
		return nil, err
	}

	value := big.NewInt(0)
	deadlineUint := uint64(time.Now().Unix() + 300)
	deadlineBig := new(big.Int).SetUint64(deadlineUint)

	gasLimit, err := c.estimateGas(ctx, fromAddr, recipientAddr, calldata)
	if err != nil {
		gasLimit = 300_000
	}
	reqGas := new(big.Int).SetUint64(gasLimit)

	buildReq := func(gas *big.Int) (ForwardRequestData, error) {
		sig, err := signForwardRequest(signer, domain, fromAddr, recipientAddr, value, gas, nonce, deadlineUint, calldata)
		if err != nil {
			return ForwardRequestData{}, err
		}
		return ForwardRequestData{
			From:      fromAddr,
			To:        recipientAddr,
			Value:     new(big.Int).Set(value),
			Gas:       new(big.Int).Set(gas),
			Deadline:  new(big.Int).Set(deadlineBig),
			Data:      calldata,
			Signature: sig,
		}, nil
	}

	req, err := buildReq(reqGas)
	if err != nil {
		return nil, err
	}

	receipt, err := c.SendTxByRelayer(ctx, fb, forwarder, "execute", req)
	if err == nil {
		return receipt, nil
	}

	// Retry once with bumped gas.
	bumped := new(big.Int).Set(reqGas)
	if bumped.Sign() > 0 {
		bumped.Mul(bumped, big.NewInt(3))
		bumped.Div(bumped, big.NewInt(2))
	}
	if bumped.Cmp(big.NewInt(800_000)) < 0 {
		bumped.SetInt64(800_000)
	}
	retryReq, retryErr := buildReq(bumped)
	if retryErr != nil {
		return nil, err
	}
	receipt2, retrySendErr := c.SendTxByRelayer(ctx, fb, forwarder, "execute", retryReq)
	if retrySendErr == nil {
		return receipt2, nil
	}
	return nil, retrySendErr
}

func (c *Client) contractABI(name types.ContractName) (common.Address, abi.ABI, error) {
	artifacts, err := abis.Get(c.cfg.Env)
	if err != nil {
		return common.Address{}, abi.ABI{}, fmt.Errorf("load abis: %w", err)
	}
	art, ok := artifacts[name]
	if !ok {
		return common.Address{}, abi.ABI{}, fmt.Errorf("%w: %s", apperr.Blockchain.ErrContractNotFound, name)
	}
	addr := common.HexToAddress(art.Address)
	if addr == (common.Address{}) {
		return common.Address{}, abi.ABI{}, fmt.Errorf("%w: %s address empty", apperr.Blockchain.ErrContractNotFound, name)
	}
	parsedABI, err := abi.JSON(strings.NewReader(string(art.ABI)))
	if err != nil {
		return common.Address{}, abi.ABI{}, fmt.Errorf("parse abi for %s: %w", name, err)
	}
	return addr, parsedABI, nil
}

func (c *Client) contractWithABI(name types.ContractName) (*bind.BoundContract, common.Address, abi.ABI, error) {
	addr, parsed, err := c.contractABI(name)
	if err != nil {
		return nil, common.Address{}, abi.ABI{}, err
	}
	bound := bind.NewBoundContract(addr, parsed, c.rpc, c.rpc, c.rpc)
	return bound, addr, parsed, nil
}

type forwarderEIP712Domain struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}

func (c *Client) forwarderDomain(ctx context.Context, forwarderAddr common.Address, forwarderABI abi.ABI) (forwarderEIP712Domain, error) {
	if c == nil || c.rpc == nil {
		return forwarderEIP712Domain{}, fmt.Errorf("client is nil")
	}
	if forwarderAddr == (common.Address{}) {
		return forwarderEIP712Domain{}, fmt.Errorf("forwarder address is empty")
	}

	data, err := forwarderABI.Pack("eip712Domain")
	if err != nil {
		return forwarderEIP712Domain{}, fmt.Errorf("pack eip712Domain: %w", err)
	}

	raw, err := c.rpc.CallContract(ctx, ethereum.CallMsg{To: &forwarderAddr, Data: data}, nil)
	if err != nil {
		return forwarderEIP712Domain{}, fmt.Errorf("call eip712Domain: %w", err)
	}

	values, err := forwarderABI.Unpack("eip712Domain", raw)
	if err != nil {
		return forwarderEIP712Domain{}, fmt.Errorf("unpack eip712Domain (len=%d): %w", len(raw), err)
	}
	if len(values) != 7 {
		return forwarderEIP712Domain{}, fmt.Errorf("unexpected eip712Domain outputs: want 7 got %d", len(values))
	}

	var fields [1]byte
	switch v := values[0].(type) {
	case [1]byte:
		fields = v
	case uint8:
		fields[0] = byte(v)
	default:
		return forwarderEIP712Domain{}, fmt.Errorf("unexpected eip712Domain.fields type %T", values[0])
	}

	name, ok := values[1].(string)
	if !ok {
		return forwarderEIP712Domain{}, fmt.Errorf("unexpected eip712Domain.name type %T", values[1])
	}
	version, ok := values[2].(string)
	if !ok {
		return forwarderEIP712Domain{}, fmt.Errorf("unexpected eip712Domain.version type %T", values[2])
	}

	var chainID *big.Int
	switch v := values[3].(type) {
	case *big.Int:
		chainID = v
	case big.Int:
		chainID = new(big.Int).Set(&v)
	default:
		return forwarderEIP712Domain{}, fmt.Errorf("unexpected eip712Domain.chainId type %T", values[3])
	}
	if chainID == nil {
		chainID = big.NewInt(0)
	}

	verifyingContract, ok := values[4].(common.Address)
	if !ok {
		return forwarderEIP712Domain{}, fmt.Errorf("unexpected eip712Domain.verifyingContract type %T", values[4])
	}
	if verifyingContract == (common.Address{}) {
		return forwarderEIP712Domain{}, fmt.Errorf("invalid forwarder domain verifying contract")
	}

	var salt [32]byte
	switch v := values[5].(type) {
	case [32]byte:
		salt = v
	default:
		return forwarderEIP712Domain{}, fmt.Errorf("unexpected eip712Domain.salt type %T", values[5])
	}

	extensions, err := castBigIntSlice(values[6])
	if err != nil {
		return forwarderEIP712Domain{}, fmt.Errorf("decode eip712Domain.extensions: %w", err)
	}

	return forwarderEIP712Domain{
		Fields:            fields,
		Name:              name,
		Version:           version,
		ChainId:           chainID,
		VerifyingContract: verifyingContract,
		Salt:              salt,
		Extensions:        extensions,
	}, nil
}

func (c *Client) forwarderNonce(ctx context.Context, forwarderAddr common.Address, forwarderABI abi.ABI, owner common.Address) (*big.Int, error) {
	if c == nil || c.rpc == nil {
		return nil, fmt.Errorf("client is nil")
	}
	if forwarderAddr == (common.Address{}) {
		return nil, fmt.Errorf("forwarder address is empty")
	}

	data, err := forwarderABI.Pack("nonces", owner)
	if err != nil {
		return nil, fmt.Errorf("pack nonces: %w", err)
	}

	raw, err := c.rpc.CallContract(ctx, ethereum.CallMsg{To: &forwarderAddr, Data: data}, nil)
	if err != nil {
		return nil, fmt.Errorf("call nonces: %w", err)
	}

	values, err := forwarderABI.Unpack("nonces", raw)
	if err != nil {
		return nil, fmt.Errorf("unpack nonces (len=%d): %w", len(raw), err)
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("unexpected nonces outputs: want 1 got %d", len(values))
	}

	switch v := values[0].(type) {
	case *big.Int:
		return v, nil
	case big.Int:
		return new(big.Int).Set(&v), nil
	default:
		return nil, fmt.Errorf("unexpected nonces return type %T", values[0])
	}
}

func castBigIntSlice(v any) ([]*big.Int, error) {
	switch vv := v.(type) {
	case []*big.Int:
		return vv, nil
	case []big.Int:
		out := make([]*big.Int, 0, len(vv))
		for i := range vv {
			out = append(out, new(big.Int).Set(&vv[i]))
		}
		return out, nil
	case []any:
		out := make([]*big.Int, 0, len(vv))
		for _, item := range vv {
			switch n := item.(type) {
			case *big.Int:
				out = append(out, n)
			case big.Int:
				out = append(out, new(big.Int).Set(&n))
			default:
				return nil, fmt.Errorf("unexpected element type %T", item)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unexpected slice type %T", v)
	}
}

func (c *Client) estimateGas(ctx context.Context, from, to common.Address, data []byte) (uint64, error) {
	if c == nil || c.rpc == nil {
		return 0, fmt.Errorf("client is nil")
	}
	msg := ethereum.CallMsg{
		From: from,
		To:   &to,
		Data: data,
	}
	gas, err := c.rpc.EstimateGas(ctx, msg)
	if err != nil {
		return 0, err
	}
	// Add a small buffer.
	return gas + gas/10, nil
}

func signForwardRequest(
	signer *ecdsa.PrivateKey,
	domain forwarderEIP712Domain,
	from common.Address,
	to common.Address,
	value *big.Int,
	gas *big.Int,
	nonce *big.Int,
	deadline uint64,
	data []byte,
) ([]byte, error) {
	if signer == nil {
		return nil, fmt.Errorf("signer is nil")
	}
	if nonce == nil {
		nonce = big.NewInt(0)
	}
	if value == nil {
		value = big.NewInt(0)
	}
	if gas == nil {
		gas = big.NewInt(0)
	}

	chainID := domain.ChainId
	if chainID == nil {
		chainID = big.NewInt(0)
	}
	chainHex := math.NewHexOrDecimal256(chainID.Int64())

	td := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"ForwardRequest": []apitypes.Type{
				{Name: "from", Type: "address"},
				{Name: "to", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "gas", Type: "uint256"},
				{Name: "nonce", Type: "uint256"},
				{Name: "deadline", Type: "uint48"},
				{Name: "data", Type: "bytes"},
			},
		},
		PrimaryType: "ForwardRequest",
		Domain: apitypes.TypedDataDomain{
			Name:              domain.Name,
			Version:           domain.Version,
			ChainId:           chainHex,
			VerifyingContract: domain.VerifyingContract.Hex(),
		},
		Message: apitypes.TypedDataMessage{
			"from":     from.Hex(),
			"to":       to.Hex(),
			"value":    value.String(),
			"gas":      gas.String(),
			"nonce":    nonce.String(),
			"deadline": fmt.Sprintf("%d", deadline),
			"data":     hexutil.Encode(data),
		},
	}

	digest, _, err := apitypes.TypedDataAndHash(td)
	if err != nil {
		return nil, fmt.Errorf("typed data hash: %w", err)
	}

	sig, err := gethcrypto.Sign(digest, signer)
	if err != nil {
		return nil, fmt.Errorf("sign typed data: %w", err)
	}
	if len(sig) == 65 && sig[64] < 27 {
		sig[64] += 27
	}
	return sig, nil
}
