package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"log"

	"AuthService/internal/storage/postgre"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
	ErrTokenUsed    = errors.New("refresh token already used")
	ErrIPMismatch   = errors.New("IP mismatch")
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"uid"`
	IP     string `json:"ip"`
}

type Service struct {
	storage      *postgre.Storage
	jwtKey       []byte
	accessTTL    time.Duration
	refreshTTL   time.Duration
	emailService EmailService
}

type EmailService interface {
	SendIPChangeWarning(userID uuid.UUID, oldIP, newIP string) error
}

func New(storage *postgre.Storage, jwtKey []byte, accessTTL, refreshTTL time.Duration, emailService EmailService) *Service {
	return &Service{
		storage:      storage,
		jwtKey:       jwtKey,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
		emailService: emailService,
	}
}

func (s *Service) CreateTokenPair(ctx context.Context, userID uuid.UUID, ip string) (*TokenPair, error) {
	tokensID := uuid.New()

	accessToken, err := s.createAccessToken(userID, ip, tokensID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.createRefreshToken(ctx, userID, ip, tokensID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) createAccessToken(userID uuid.UUID, ip string, tokenID uuid.UUID) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        tokenID.String(),
		},
		UserID: userID.String(),
		IP:     ip,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(s.jwtKey)
}

func (s *Service) createRefreshToken(ctx context.Context, userID uuid.UUID, ip string, tokenID uuid.UUID) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	refreshStr := base64.URLEncoding.EncodeToString(tokenBytes)

	hash, err := bcrypt.GenerateFromPassword([]byte(refreshStr), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	token := postgre.RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: string(hash),
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(s.refreshTTL),
		IssuedIP:  ip,
	}

	if err := s.storage.SaveRefreshToken(ctx, token); err != nil {
		return "", err
	}

	tokenData := token.ID.String() + ":" + refreshStr
	return base64.URLEncoding.EncodeToString([]byte(tokenData)), nil
}

func (s *Service) RefreshTokensByRefresh(ctx context.Context, refreshToken string, ip string) (*TokenPair, error) {
	tokenData, err := base64.URLEncoding.DecodeString(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	parts := strings.Split(string(tokenData), ":")
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	tokenID, err := uuid.Parse(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	storedToken, err := s.storage.GetRefreshToken(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	if storedToken.IsUsed {
		return nil, ErrTokenUsed
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedToken.TokenHash), []byte(parts[1])); err != nil {
		return nil, ErrInvalidToken
	}

	if ip != storedToken.IssuedIP {
		if err := s.emailService.SendIPChangeWarning(storedToken.UserID, storedToken.IssuedIP, ip); err != nil {
			log.Printf("Failed to send IP change warning: %v", err)
		}
		return nil, ErrIPMismatch
	}

	if err := s.storage.MarkTokenAsUsed(ctx, tokenID); err != nil {
		return nil, err
	}

	return s.CreateTokenPair(ctx, storedToken.UserID, ip)
}

func (s *Service) RefreshTokensByAccess(ctx context.Context, AccessToken string, ip string) (*TokenPair, error) {

	token, err := jwt.ParseWithClaims(AccessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if token, ok := token.Claims.(*Claims); !ok {
		return nil, ErrInvalidToken
	} else {
		claims = *token
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrTokenExpired
	}

	tokenId, err := uuid.Parse(claims.ID)
	if err != nil {
		log.Printf("Failed to parse tokenId to UUID: %v", err)
	}

	tokenUsed, err := s.storage.ChekcRefreshToken(context.Background(), tokenId)
	if err != nil {
		{
			return nil, err
		}
	}
	if tokenUsed {
		return nil, ErrTokenUsed
	}

	userId, err := uuid.Parse(claims.UserID)
	if err != nil {
		log.Printf("Failed to parse UserID to UUID: %v", err)
	}

	if ip != claims.IP {

		if err := s.emailService.SendIPChangeWarning(userId, claims.IP, ip); err != nil {
			log.Printf("Failed to send IP change warning: %v", err)
		}
		return nil, ErrIPMismatch
	}

	RefreshID, err := uuid.Parse(claims.ID)

	if err != nil {
		return nil, err
	}

	if err := s.storage.MarkTokenAsUsed(ctx, RefreshID); err != nil {
		return nil, err
	}

	return s.CreateTokenPair(ctx, userId, ip)
}
