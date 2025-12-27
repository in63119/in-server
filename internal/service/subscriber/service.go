package subscriber

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"in-server/pkg/abis"
	"in-server/pkg/apperr"
	"in-server/pkg/config"
	"in-server/pkg/eth"
	"in-server/pkg/firebase"
	"in-server/pkg/types"
)

type Service struct {
	cfg config.Config
	eth *eth.Client
	fb  *firebase.Client
}

func New(ctx context.Context, cfg config.Config) (*Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ethClient, err := eth.Dial(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("dial eth client: %w", err)
	}

	fbClient, err := firebase.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("init firebase: %w", err)
	}

	return &Service{
		cfg: cfg,
		eth: ethClient,
		fb:  fbClient,
	}, nil
}

func (s *Service) Count() (int, error) {
	ctx := context.Background()

	if s.eth == nil {
		return 0, fmt.Errorf("eth client is nil")
	}

	adminCode := strings.TrimSpace(s.cfg.Auth.AdminCode)
	if adminCode == "" {
		return 0, apperr.System.ErrMissingAuthAdminCode
	}

	_, ownerAddr, err := s.eth.Wallet(adminCode)
	if err != nil {
		return 0, fmt.Errorf("owner address: %w", err)
	}

	accts, err := s.eth.Accounts()
	if err != nil {
		return 0, fmt.Errorf("load relayer accounts: %w", err)
	}

	relayerAddr, err := eth.AddressFromPrivateKey(accts.Relayer)
	if err != nil {
		return 0, fmt.Errorf("relayer address: %w", err)
	}

	contract, _, err := s.eth.Contract(types.SUBSCRIBERSTORAGE)
	if err != nil {
		return 0, fmt.Errorf("bind subscriber storage: %w", err)
	}

	callOpts := &bind.CallOpts{Context: ctx, From: relayerAddr}

	var emails []string
	out := []any{&emails}
	if err := contract.Call(callOpts, &out, "getSubscriberEmails", ownerAddr); err != nil {
		return 0, apperr.Wrap(err, apperr.Subscriber.ErrGetSubscribers.Code, "getSubscriberEmails", apperr.Subscriber.ErrGetSubscribers.Status)
	}

	return len(emails), nil
}

func (s *Service) Create(address, email string) error {
	if s.eth == nil {
		return fmt.Errorf("eth client is nil")
	}
	if s.fb == nil {
		return fmt.Errorf("firebase client is nil")
	}

	email = strings.TrimSpace(email)
	account := common.HexToAddress(strings.TrimSpace(address))
	if account == (common.Address{}) {
		adminCode := strings.TrimSpace(s.cfg.Auth.AdminCode)
		if adminCode == "" {
			return apperr.System.ErrMissingAuthAdminCode
		}
		_, adminAddr, err := s.eth.Wallet(adminCode)
		if err != nil {
			return apperr.Blockchain.ErrInvalidWallet
		}
		account = adminAddr
	}
	if account == (common.Address{}) || email == "" {
		return apperr.Subscriber.ErrInvalidBody
	}

	contract, contractAddr, err := s.eth.Contract(types.SUBSCRIBERSTORAGE)
	if err != nil {
		return apperr.Wrap(err, apperr.Subscriber.ErrCreate.Code, "bind subscriber storage", apperr.Subscriber.ErrCreate.Status)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	receipt, err := s.eth.SendTxByRelayer(ctx, s.fb, contract, "addSubscriberEmail", account, email)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "subscriberstorage__subscriberemailalreadyexists") {
			return apperr.Subscriber.ErrAlreadyExists
		}
		return apperr.Wrap(err, apperr.Subscriber.ErrCreate.Code, "addSubscriberEmail", apperr.Subscriber.ErrCreate.Status)
	}

	ok, err := hasSubscriberEmailAddedEvent(s.cfg.Env, contractAddr, receipt)
	if err != nil {
		return apperr.Wrap(err, apperr.Subscriber.ErrCreate.Code, "parse receipt logs", apperr.Subscriber.ErrCreate.Status)
	}
	if !ok {
		return apperr.Subscriber.ErrCreate
	}

	return nil
}

func (s *Service) List(limit, offset int) ([]string, error) {
	ctx := context.Background()

	if s.eth == nil {
		return nil, fmt.Errorf("eth client is nil")
	}

	adminCode := strings.TrimSpace(s.cfg.Auth.AdminCode)
	if adminCode == "" {
		return nil, apperr.System.ErrMissingAuthAdminCode
	}

	_, ownerAddr, err := s.eth.Wallet(adminCode)
	if err != nil {
		return nil, fmt.Errorf("owner address: %w", err)
	}

	accts, err := s.eth.Accounts()
	if err != nil {
		return nil, fmt.Errorf("load relayer accounts: %w", err)
	}

	relayerAddr, err := eth.AddressFromPrivateKey(accts.Relayer)
	if err != nil {
		return nil, fmt.Errorf("relayer address: %w", err)
	}

	contract, _, err := s.eth.Contract(types.SUBSCRIBERSTORAGE)
	if err != nil {
		return nil, fmt.Errorf("bind subscriber storage: %w", err)
	}

	callOpts := &bind.CallOpts{Context: ctx, From: relayerAddr}

	var emails []string
	out := []any{&emails}
	if err := contract.Call(callOpts, &out, "getSubscriberEmails", ownerAddr); err != nil {
		return nil, apperr.Wrap(err, apperr.Subscriber.ErrGetSubscribers.Code, "getSubscriberEmails", apperr.Subscriber.ErrGetSubscribers.Status)
	}

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	start := offset
	if start > len(emails) {
		start = len(emails)
	}
	end := start + limit
	if end > len(emails) {
		end = len(emails)
	}

	return emails[start:end], nil
}

func hasSubscriberEmailAddedEvent(env string, contractAddr common.Address, receipt *gethtypes.Receipt) (bool, error) {
	if receipt == nil {
		return false, fmt.Errorf("receipt is nil")
	}

	artifacts, err := abis.Get(env)
	if err != nil {
		return false, fmt.Errorf("load abis: %w", err)
	}
	art, ok := artifacts[types.SUBSCRIBERSTORAGE]
	if !ok {
		return false, fmt.Errorf("subscriber storage abi not found")
	}

	parsed, err := abi.JSON(strings.NewReader(string(art.ABI)))
	if err != nil {
		return false, fmt.Errorf("parse subscriber storage abi: %w", err)
	}
	ev, ok := parsed.Events["SubscriberEmailAdded"]
	if !ok {
		return false, fmt.Errorf("SubscriberEmailAdded event not found")
	}

	for _, lg := range receipt.Logs {
		if lg == nil || len(lg.Topics) == 0 {
			continue
		}
		if lg.Address != contractAddr {
			continue
		}
		if lg.Topics[0] == ev.ID {
			return true, nil
		}
	}
	return false, nil
}
