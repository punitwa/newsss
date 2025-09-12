package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"news-aggregator/internal/config"
	"news-aggregator/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type UserRepository struct {
	db     *pgxpool.Pool
	logger zerolog.Logger
}

func NewUserRepository(cfg *config.Config, logger zerolog.Logger) (*UserRepository, error) {
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	// Create connection pool
	db, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// Test connection
	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &UserRepository{
		db:     db,
		logger: logger.With().Str("component", "user_repository").Logger(),
	}

	// Initialize database schema
	if err := repo.initSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

func (r *UserRepository) initSchema(ctx context.Context) error {
	r.logger.Info().Msg("Initializing user schema")

	// Create tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email TEXT UNIQUE NOT NULL,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			first_name TEXT,
			last_name TEXT,
			avatar TEXT,
			preferences JSONB DEFAULT '{}',
			is_active BOOLEAN DEFAULT true,
			is_admin BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS bookmarks (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			news_id UUID NOT NULL REFERENCES news(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id, news_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_bookmarks_user_id ON bookmarks(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bookmarks_news_id ON bookmarks(news_id)`,
	}

	for _, query := range queries {
		if _, err := r.db.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	r.logger.Info().Msg("User schema initialized successfully")
	return nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	r.logger.Debug().Str("email", user.Email).Msg("Creating user")

	// Marshal preferences to JSON
	preferencesJSON, err := json.Marshal(user.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	query := `
		INSERT INTO users (email, username, password_hash, first_name, last_name, 
						  avatar, preferences, is_active, is_admin)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		user.Email, user.Username, user.PasswordHash, user.FirstName,
		user.LastName, user.Avatar, preferencesJSON, user.IsActive, user.IsAdmin,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	r.logger.Debug().Str("id", id).Msg("Getting user by ID")

	query := `
		SELECT id, email, username, password_hash, first_name, last_name, avatar,
			   preferences, is_active, is_admin, created_at, updated_at
		FROM users WHERE id = $1
	`

	var user models.User
	var preferencesJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Avatar, &preferencesJSON,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Unmarshal preferences
	if len(preferencesJSON) > 0 {
		if err := json.Unmarshal(preferencesJSON, &user.Preferences); err != nil {
			r.logger.Warn().Err(err).Str("id", id).Msg("Failed to unmarshal preferences")
			user.Preferences = models.Preferences{} // Initialize with empty struct
		}
	}

	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	r.logger.Debug().Str("email", email).Msg("Getting user by email")

	query := `
		SELECT id, email, username, password_hash, first_name, last_name, avatar,
			   preferences, is_active, is_admin, created_at, updated_at
		FROM users WHERE email = $1
	`

	var user models.User
	var preferencesJSON []byte

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Avatar, &preferencesJSON,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Unmarshal preferences
	if len(preferencesJSON) > 0 {
		if err := json.Unmarshal(preferencesJSON, &user.Preferences); err != nil {
			r.logger.Warn().Err(err).Str("email", email).Msg("Failed to unmarshal preferences")
			user.Preferences = models.Preferences{} // Initialize with empty struct
		}
	}

	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	r.logger.Debug().Str("id", user.ID).Msg("Updating user")

	// Marshal preferences to JSON
	preferencesJSON, err := json.Marshal(user.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	query := `
		UPDATE users SET 
			email = $2, username = $3, first_name = $4, last_name = $5,
			avatar = $6, preferences = $7, is_active = $8, is_admin = $9,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err = r.db.QueryRow(ctx, query,
		user.ID, user.Email, user.Username, user.FirstName, user.LastName,
		user.Avatar, preferencesJSON, user.IsActive, user.IsAdmin,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	r.logger.Debug().Str("id", id).Msg("Deleting user")

	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) GetUsers(ctx context.Context, page, limit int) ([]models.User, int, error) {
	r.logger.Debug().Int("page", page).Int("limit", limit).Msg("Getting users")

	// Get total count
	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	// Get users with pagination
	offset := (page - 1) * limit
	query := `
		SELECT id, email, username, first_name, last_name, avatar,
			   preferences, is_active, is_admin, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var preferencesJSON []byte

		err := rows.Scan(
			&user.ID, &user.Email, &user.Username, &user.FirstName,
			&user.LastName, &user.Avatar, &preferencesJSON, &user.IsActive,
			&user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user row: %w", err)
		}

		// Unmarshal preferences
		if len(preferencesJSON) > 0 {
			if err := json.Unmarshal(preferencesJSON, &user.Preferences); err != nil {
				r.logger.Warn().Err(err).Str("id", user.ID).Msg("Failed to unmarshal preferences")
				user.Preferences = models.Preferences{} // Initialize with empty struct
			}
		}

		users = append(users, user)
	}

	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("error iterating user rows: %w", rows.Err())
	}

	return users, total, nil
}

func (r *UserRepository) CreateBookmark(ctx context.Context, bookmark *models.Bookmark) error {
	r.logger.Debug().Str("user_id", bookmark.UserID).Str("news_id", bookmark.NewsID).Msg("Creating bookmark")

	query := `
		INSERT INTO bookmarks (user_id, news_id)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query, bookmark.UserID, bookmark.NewsID).Scan(
		&bookmark.ID, &bookmark.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create bookmark: %w", err)
	}

	return nil
}

func (r *UserRepository) GetBookmarks(ctx context.Context, userID string, page, limit int) ([]models.Bookmark, int, error) {
	r.logger.Debug().Str("user_id", userID).Int("page", page).Int("limit", limit).Msg("Getting bookmarks")

	// Get total count
	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM bookmarks WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get bookmark count: %w", err)
	}

	// Get bookmarks with news details
	offset := (page - 1) * limit
	query := `
		SELECT b.id, b.user_id, b.news_id, b.created_at,
			   n.title, n.summary, n.url, n.image_url, n.author, n.source,
			   n.category, n.published_at
		FROM bookmarks b
		JOIN news n ON b.news_id = n.id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []models.Bookmark
	for rows.Next() {
		var bookmark models.Bookmark
		var news models.News

		err := rows.Scan(
			&bookmark.ID, &bookmark.UserID, &bookmark.NewsID, &bookmark.CreatedAt,
			&news.Title, &news.Summary, &news.URL, &news.ImageURL, &news.Author,
			&news.Source, &news.Category, &news.PublishedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan bookmark row: %w", err)
		}

		news.ID = bookmark.NewsID
		bookmark.News = &news
		bookmarks = append(bookmarks, bookmark)
	}

	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("error iterating bookmark rows: %w", rows.Err())
	}

	return bookmarks, total, nil
}

func (r *UserRepository) DeleteBookmark(ctx context.Context, userID, bookmarkID string) error {
	r.logger.Debug().Str("user_id", userID).Str("bookmark_id", bookmarkID).Msg("Deleting bookmark")

	query := `DELETE FROM bookmarks WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, bookmarkID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("bookmark not found")
	}

	return nil
}

func (r *UserRepository) DeleteBookmarkByArticle(ctx context.Context, userID, articleID string) error {
	r.logger.Debug().Str("user_id", userID).Str("article_id", articleID).Msg("Deleting bookmark by article")

	query := `DELETE FROM bookmarks WHERE user_id = $1 AND news_id = $2`

	result, err := r.db.Exec(ctx, query, userID, articleID)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("bookmark not found")
	}

	return nil
}

func (r *UserRepository) Close() error {
	r.db.Close()
	return nil
}
