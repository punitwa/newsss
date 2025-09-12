package services

import (
	"context"
	"fmt"
	"time"

	"news-aggregator/internal/config"
	"news-aggregator/internal/models"
	"news-aggregator/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	config     *config.Config
	logger     zerolog.Logger
	repository *repository.UserRepository
}

func NewUserService(cfg *config.Config, logger zerolog.Logger) (*UserService, error) {
	repo, err := repository.NewUserRepository(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create user repository: %w", err)
	}

	return &UserService{
		config:     cfg,
		logger:     logger,
		repository: repo,
	}, nil
}

func (s *UserService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	s.logger.Debug().Str("email", req.Email).Msg("Registering user")

	// Check if user already exists
	existingUser, _ := s.repository.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash password")
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
		IsAdmin:      false,
		Preferences:  models.Preferences{},
	}

	if err := s.repository.CreateUser(ctx, user); err != nil {
		s.logger.Error().Err(err).Str("email", req.Email).Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Remove password hash from response
	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	s.logger.Debug().Str("email", email).Msg("User login attempt")

	user, err := s.repository.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("User not found")
		return "", nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		s.logger.Warn().Str("email", email).Msg("Inactive user login attempt")
		return "", nil, fmt.Errorf("account is inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("Invalid password")
		return "", nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("Failed to generate JWT")
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Remove password hash from user data
	user.PasswordHash = ""

	return token, user, nil
}

func (s *UserService) generateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"email":    user.Email,
		"username": user.Username,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(s.config.JWT.ExpirationTime).Unix(),
		"iat":      time.Now().Unix(),
		"iss":      s.config.JWT.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.SecretKey))
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	s.logger.Debug().Str("user_id", userID).Msg("Getting user profile")

	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get user profile")
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Remove sensitive information
	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) error {
	s.logger.Debug().Str("user_id", userID).Msg("Updating user profile")

	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update fields
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if err := s.repository.UpdateUser(ctx, user); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user profile")
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (s *UserService) AddBookmark(ctx context.Context, userID, newsID string) error {
	s.logger.Debug().Str("user_id", userID).Str("news_id", newsID).Msg("Adding bookmark")

	bookmark := &models.Bookmark{
		UserID: userID,
		NewsID: newsID,
	}

	if err := s.repository.CreateBookmark(ctx, bookmark); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Str("news_id", newsID).Msg("Failed to add bookmark")
		return fmt.Errorf("failed to add bookmark: %w", err)
	}

	return nil
}

func (s *UserService) GetBookmarks(ctx context.Context, userID string, page, limit int) ([]models.Bookmark, int, error) {
	s.logger.Debug().Str("user_id", userID).Int("page", page).Int("limit", limit).Msg("Getting bookmarks")

	bookmarks, total, err := s.repository.GetBookmarks(ctx, userID, page, limit)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to get bookmarks")
		return nil, 0, fmt.Errorf("failed to get bookmarks: %w", err)
	}

	return bookmarks, total, nil
}

func (s *UserService) RemoveBookmark(ctx context.Context, userID, bookmarkID string) error {
	s.logger.Debug().Str("user_id", userID).Str("bookmark_id", bookmarkID).Msg("Removing bookmark")

	if err := s.repository.DeleteBookmark(ctx, userID, bookmarkID); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Str("bookmark_id", bookmarkID).Msg("Failed to remove bookmark")
		return fmt.Errorf("failed to remove bookmark: %w", err)
	}

	return nil
}

func (s *UserService) RemoveBookmarkByArticle(ctx context.Context, userID, articleID string) error {
	s.logger.Debug().Str("user_id", userID).Str("article_id", articleID).Msg("Removing bookmark by article")

	if err := s.repository.DeleteBookmarkByArticle(ctx, userID, articleID); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Str("article_id", articleID).Msg("Failed to remove bookmark")
		return fmt.Errorf("failed to remove bookmark: %w", err)
	}

	return nil
}

func (s *UserService) UpdatePreferences(ctx context.Context, userID string, req *models.PreferencesRequest) error {
	s.logger.Debug().Str("user_id", userID).Msg("Updating user preferences")

	user, err := s.repository.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update preferences
	user.Preferences.Categories = req.Categories
	user.Preferences.Sources = req.Sources
	if req.NotificationEnabled != nil {
		user.Preferences.NotificationEnabled = *req.NotificationEnabled
	}
	if req.EmailDigest != nil {
		user.Preferences.EmailDigest = *req.EmailDigest
	}
	user.Preferences.DigestFrequency = req.DigestFrequency

	if err := s.repository.UpdateUser(ctx, user); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID).Msg("Failed to update user preferences")
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	return nil
}

func (s *UserService) GetUsers(ctx context.Context, page, limit int) ([]models.User, int, error) {
	s.logger.Debug().Int("page", page).Int("limit", limit).Msg("Getting users")

	users, total, err := s.repository.GetUsers(ctx, page, limit)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get users")
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	// Remove sensitive information
	for i := range users {
		users[i].PasswordHash = ""
	}

	return users, total, nil
}
