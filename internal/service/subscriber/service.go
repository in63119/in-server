package subscriber

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"in-server/pkg/apperr"
	"in-server/pkg/config"
	"in-server/pkg/eth"
	"in-server/pkg/types"
)

type Service struct {
	cfg config.Config
	eth *eth.Client
}

func New(ctx context.Context, cfg config.Config) (*Service, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ethClient, err := eth.Dial(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("dial eth client: %w", err)
	}

	return &Service{
		cfg: cfg,
		eth: ethClient,
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

func (s *Service) Create(input any) error {
	return fmt.Errorf("not implemented")
}
