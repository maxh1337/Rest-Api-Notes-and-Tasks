package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"rest-api-notes/internal/config"
	"rest-api-notes/internal/domain/entities"
	"time"
)

type twoFactorService struct {
	sS         SessionService
	cfg        *config.Config
	codeLength int
	codeExpiry time.Duration
}

type TwoFactorService interface {
	MakeRequestToGateway(ctx context.Context, phoneNumber, code string) error
	Generate2FACode() (string, error)
}

func NewTwoFactorService(sS SessionService, cfg *config.Config) TwoFactorService {
	return &twoFactorService{
		sS:         sS,
		cfg:        cfg,
		codeLength: 4,
		codeExpiry: 5 * time.Minute}
}

func (s *twoFactorService) MakeRequestToGateway(ctx context.Context, phoneNumber, code string) error {
	url := "https://gatewayapi.telegram.org/sendVerificationMessage"

	authorizationHeader := fmt.Sprintf("Bearer %s", s.cfg.GATEWAY_API_TOKEN)

	sendCodeBody := entities.SendCodeRequest{
		PhoneNumber: phoneNumber,
		Code:        code,
	}

	jsonData, err := json.Marshal(sendCodeBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorizationHeader)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var telegramResp entities.RequestStatus
	if err := json.Unmarshal(body, &telegramResp); err != nil {
		return err
	}

	if telegramResp.DeliveryStatus.Status == "revoked" {
		return entities.Err2FACodeRevoked
	}

	return nil
}

func (s *twoFactorService) Generate2FACode() (string, error) {
	code := ""
	for i := 0; i < s.codeLength; i++ {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("failed to generate random digit: %w", err)
		}
		code += digit.String()
	}
	return code, nil
}
