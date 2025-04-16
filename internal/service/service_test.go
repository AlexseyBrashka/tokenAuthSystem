package service_test

import (
	AuthService "AuthService/internal/service"
	suite "AuthService/tests/suite"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

var ipDataSet = []string{
	"179.114.152.18",
	"148.136.154.50",
	"37.120.200.237",
	"81.73.80.179",
	"92.77.50.9",
	"89.72.231.230",
	"23.70.67.28",
	"216.204.202.125",
	"79.97.104.104",
	"96.10.147.251",
	"51.229.27.177",
	"120.214.15.80",
	"155.18.201.110",
	"91.6.64.161",
	"178.222.149.11"}

type req struct {
	userID  uuid.UUID
	ip      string
	tokenID uuid.UUID
}

func TestTokenPair(t *testing.T) {
	authService := suite.NewAuthService(t)
	for i := 0; i < len(ipDataSet); i++ {
		tokens, _ := authService.CreateTokenPair(context.Background(), uuid.New(), ipDataSet[i])
		assert.IsType(t, AuthService.TokenPair{}, *tokens)
	}
	for i := 0; i < len(ipDataSet); i++ {
		_, err := authService.CreateTokenPair(context.Background(), uuid.New(), "unvalidIP")
		assert.Error(t, err)
	}

}
func TestUseRefresh(t *testing.T) {

	authService := suite.NewAuthService(t)

	// проеряем генерацию
	tokens, err := authService.CreateTokenPair(context.Background(), uuid.New(), ipDataSet[1])
	assert.NoError(t, err)
	assert.IsType(t, AuthService.TokenPair{}, *tokens)

	// обновление через Refresh
	tokens, err = authService.RefreshTokensByRefresh(context.Background(), tokens.RefreshToken, ipDataSet[1])
	assert.IsType(t, AuthService.TokenPair{}, *tokens)

	// обновление через Access
	tokens, err = authService.RefreshTokensByAccess(context.Background(), tokens.AccessToken, ipDataSet[1])
	assert.IsType(t, AuthService.TokenPair{}, *tokens)

	//проверяем на уже использованный Access токен
	oldtoken, err := authService.CreateTokenPair(context.Background(), uuid.New(), ipDataSet[1])
	tokens, err = authService.RefreshTokensByAccess(context.Background(), oldtoken.AccessToken, ipDataSet[1])

	_, err = authService.RefreshTokensByAccess(context.Background(), oldtoken.AccessToken, ipDataSet[1])
	assert.ErrorIs(t, err, AuthService.ErrTokenUsed)

	//проверяем на ввод токена некорректного типа
	tokens, err = authService.CreateTokenPair(context.Background(), uuid.New(), ipDataSet[1])
	assert.NoError(t, err)
	_, err = authService.RefreshTokensByRefresh(context.Background(), tokens.AccessToken, ipDataSet[1])
	assert.ErrorIs(t, err, AuthService.ErrInvalidToken)

	//меняем IP
	tokens, err = authService.CreateTokenPair(context.Background(), uuid.New(), ipDataSet[1])
	assert.NoError(t, err)
	_, err = authService.RefreshTokensByAccess(context.Background(), tokens.AccessToken, ipDataSet[2])
	assert.ErrorIs(t, err, AuthService.ErrIPMismatch)
}
