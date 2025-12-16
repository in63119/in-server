package visitor

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"in-server/pkg/apperr"
	"in-server/pkg/config"
	"in-server/pkg/crypto"
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

func (s *Service) Visit(ip, url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if s.eth == nil {
		return fmt.Errorf("eth client is nil")
	}

	adminCode := strings.TrimSpace(s.cfg.Auth.AdminCode)
	if adminCode == "" {
		return apperr.System.ErrMissingAuthAdminCode
	}

	_, ownerAddr, err := s.eth.Wallet(adminCode)
	if err != nil {
		return fmt.Errorf("owner address: %w", err)
	}

	contract, _, err := s.eth.Contract(types.VISITORSTORAGE)
	if err != nil {
		return fmt.Errorf("bind visitor storage: %w", err)
	}

	ipHash := common.HexToHash("0x" + crypto.SHA256(strings.TrimSpace(ip)))
	targetURL := strings.TrimSpace(url)
	if targetURL == "" {
		targetURL = "/"
	}

	if _, err := s.eth.SendTxByRelayer(ctx, s.fb, contract, "addHashedVisitorForToday", ownerAddr, ipHash, targetURL); err != nil {
		return apperr.Wrap(err, apperr.Visitors.ErrAddVisit.Code, "send tx by relayer", apperr.Visitors.ErrAddVisit.Status)
	}

	return nil
}

func (s *Service) HasVisited(ip string) (bool, error) {
	ctx := context.Background()

	if s.eth == nil {
		return false, fmt.Errorf("eth client is nil")
	}

	adminCode := strings.TrimSpace(s.cfg.Auth.AdminCode)
	if adminCode == "" {
		return false, apperr.System.ErrMissingAuthAdminCode
	}

	_, ownerAddr, err := s.eth.Wallet(adminCode)
	if err != nil {
		return false, fmt.Errorf("owner address: %w", err)
	}

	contract, _, err := s.eth.Contract(types.VISITORSTORAGE)
	if err != nil {
		return false, fmt.Errorf("bind visitor storage: %w", err)
	}

	callOpts := &bind.CallOpts{Context: ctx}

	var dayID uint64
	dayOut := []any{&dayID}
	if err := contract.Call(callOpts, &dayOut, "currentDayId"); err != nil {
		return false, apperr.Wrap(err, apperr.Visitors.ErrCheckVisit.Code, "currentDayId", apperr.Visitors.ErrCheckVisit.Status)
	}
	if dayOut[0] == nil {
		return false, fmt.Errorf("unexpected currentDayId result type")
	}

	ipHash := "0x" + crypto.SHA256(strings.TrimSpace(ip))
	hash := common.HexToHash(ipHash)

	visitedOut := []any{new(bool)}
	if err := contract.Call(callOpts, &visitedOut, "hasSeenHash", ownerAddr, dayID, hash); err != nil {
		return false, apperr.Wrap(err, apperr.Visitors.ErrCheckVisit.Code, "hasSeenHash", apperr.Visitors.ErrCheckVisit.Status)
	}

	visitedPtr, ok := visitedOut[0].(*bool)
	if !ok || visitedPtr == nil {
		return false, fmt.Errorf("unexpected hasSeenHash result type")
	}

	return *visitedPtr, nil
}

func (s *Service) Count() (uint64, error) {
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

	contract, _, err := s.eth.Contract(types.VISITORSTORAGE)
	if err != nil {
		return 0, fmt.Errorf("bind visitor storage: %w", err)
	}

	callOpts := &bind.CallOpts{Context: ctx}

	var dayID uint64
	dayOut := []any{&dayID}
	if err := contract.Call(callOpts, &dayOut, "currentDayId"); err != nil {
		return 0, apperr.Wrap(err, apperr.Visitors.ErrVisitCount.Code, "currentDayId", apperr.Visitors.ErrVisitCount.Status)
	}
	if dayOut[0] == nil {
		return 0, fmt.Errorf("unexpected currentDayId result type")
	}

	var total *big.Int
	totalOut := []any{&total}
	if err := contract.Call(callOpts, &totalOut, "totalVisitorsOf", ownerAddr, dayID); err != nil {
		return 0, apperr.Wrap(err, apperr.Visitors.ErrVisitCount.Code, "totalVisitorsOf", apperr.Visitors.ErrVisitCount.Status)
	}

	if total == nil {
		return 0, fmt.Errorf("unexpected totalVisitorsOf result type")
	}

	return total.Uint64(), nil
}
