package email

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
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
	googlemail "in-server/pkg/google"
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

	return &Service{cfg: cfg, eth: ethClient, fb: fbClient}, nil
}

func GenerateFourDigitCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", fmt.Errorf("generate pin code: %w", err)
	}
	return fmt.Sprintf("%04d", n.Int64()), nil
}

func (s *Service) ClaimPinCode(ctx context.Context, pinCode, recipientEmail string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.eth == nil {
		return fmt.Errorf("eth client is nil")
	}
	if s.fb == nil {
		return fmt.Errorf("firebase client is nil")
	}

	adminCode := strings.TrimSpace(s.cfg.Auth.AdminCode)
	if adminCode == "" {
		return apperr.System.ErrMissingAuthAdminCode
	}

	_, ownerAddr, err := s.eth.Wallet(adminCode)
	if err != nil {
		return apperr.Blockchain.ErrInvalidWallet
	}

	pinCode = strings.TrimSpace(pinCode)
	if pinCode == "" {
		return apperr.Email.ErrClaimPinCode
	}

	contract, contractAddr, err := s.eth.Contract(types.SUBSCRIBERSTORAGE)
	if err != nil {
		return apperr.Wrap(err, apperr.Email.ErrClaimPinCode.Code, "bind subscriber storage", apperr.Email.ErrClaimPinCode.Status)
	}

	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	receipt, err := s.eth.SendTxByRelayer(ctx, s.fb, contract, "claimPinCode", ownerAddr, pinCode)
	if err != nil {
		return apperr.Wrap(err, apperr.Email.ErrClaimPinCode.Code, "claimPinCode", apperr.Email.ErrClaimPinCode.Status)
	}

	ok, err := hasPinCodeStoredEvent(s.cfg.Env, contractAddr, receipt)
	if err != nil {
		return apperr.Wrap(err, apperr.Email.ErrClaimPinCode.Code, "parse receipt logs", apperr.Email.ErrClaimPinCode.Status)
	}
	if !ok {
		return apperr.Email.ErrClaimPinCode
	}

	return s.sendPinCodeEmail(ctx, recipientEmail, pinCode)
}

func (s *Service) sendPinCodeEmail(ctx context.Context, recipient, pinCode string) error {
	subject := "[IN Labs] 인증 코드"
	encodedSubject := googlemail.EncodeSubject(subject)

	body := strings.Join([]string{
		"<html><body>",
		`<div style="text-align:center; margin-bottom:16px;">`,
		`<img src="https://in-labs.s3.ap-northeast-2.amazonaws.com/images/in.png" alt="IN Labs" style="max-width:160px;height:auto;" />`,
		"</div>",
		"<p>IN Labs 구독자 메일 주소 인증 코드 안내입니다.</p>",
		"<p>----------------------</p>",
		fmt.Sprintf("<p>인증 코드: <strong>%s</strong></p>", pinCode),
		"<p>----------------------</p>",
		"<p>본인이 요청하지 않은 경우 이 메일을 무시하셔도 됩니다.</p>",
		"</body></html>",
	}, "")

	return googlemail.SendEmail(ctx, s.cfg, googlemail.EmailContent{
		Recipient: strings.TrimSpace(recipient),
		Subject:   encodedSubject,
		Body:      body,
	})
}

func (s *Service) VerifyPinCode(ctx context.Context, address, pinCode string) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.eth == nil {
		return false, fmt.Errorf("eth client is nil")
	}

	account := common.HexToAddress(strings.TrimSpace(address))
	if account == (common.Address{}) {
		return false, apperr.Email.ErrVerifyPinCode
	}
	pinCode = strings.TrimSpace(pinCode)
	if pinCode == "" {
		return false, apperr.Email.ErrVerifyPinCode
	}

	contract, _, err := s.eth.Contract(types.SUBSCRIBERSTORAGE)
	if err != nil {
		return false, apperr.Wrap(err, apperr.Email.ErrVerifyPinCode.Code, "bind subscriber storage", apperr.Email.ErrVerifyPinCode.Status)
	}

	accts, err := s.eth.Accounts()
	if err != nil {
		return false, apperr.Wrap(err, apperr.Email.ErrVerifyPinCode.Code, "load relayer accounts", apperr.Email.ErrVerifyPinCode.Status)
	}
	relayerAddr, err := eth.AddressFromPrivateKey(accts.Relayer)
	if err != nil {
		return false, apperr.Wrap(err, apperr.Email.ErrVerifyPinCode.Code, "relayer address", apperr.Email.ErrVerifyPinCode.Status)
	}

	var verified bool
	out := []any{&verified}
	if err := contract.Call(&bind.CallOpts{Context: ctx, From: relayerAddr}, &out, "isPinCodeActive", account, pinCode); err != nil {
		return false, apperr.Wrap(err, apperr.Email.ErrVerifyPinCode.Code, "isPinCodeActive", apperr.Email.ErrVerifyPinCode.Status)
	}

	defer func() {
		clearCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = s.clearExpiredPinCodes(clearCtx, account, pinCode)
	}()

	return verified, nil
}

func (s *Service) clearExpiredPinCodes(ctx context.Context, account common.Address, pinCode string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.eth == nil || s.fb == nil {
		return fmt.Errorf("missing clients")
	}

	contract, _, err := s.eth.Contract(types.SUBSCRIBERSTORAGE)
	if err != nil {
		return err
	}

	_, err = s.eth.SendTxByRelayer(ctx, s.fb, contract, "clearPinCode", account, pinCode)
	return err
}

func hasPinCodeStoredEvent(env string, contractAddr common.Address, receipt *gethtypes.Receipt) (bool, error) {
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
	ev, ok := parsed.Events["PinCodeStored"]
	if !ok {
		return false, fmt.Errorf("PinCodeStored event not found")
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
