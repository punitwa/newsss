package user

import (
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID          string                 `json:"id" db:"id"`
	Email       string                 `json:"email" db:"email"`
	Username    string                 `json:"username" db:"username"`
	PasswordHash string                `json:"-" db:"password_hash"`
	FirstName   string                 `json:"first_name" db:"first_name"`
	LastName    string                 `json:"last_name" db:"last_name"`
	Avatar      string                 `json:"avatar" db:"avatar"`
	Preferences Preferences            `json:"preferences" db:"preferences"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	IsAdmin     bool                   `json:"is_admin" db:"is_admin"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Preferences represents user preferences
type Preferences struct {
	Categories          []string `json:"categories"`
	Sources             []string `json:"sources"`
	NotificationEnabled bool     `json:"notification_enabled"`
	EmailDigest         bool     `json:"email_digest"`
	DigestFrequency     string   `json:"digest_frequency"` // daily, weekly, monthly
	Theme               string   `json:"theme"`            // light, dark, auto
	Language            string   `json:"language"`         // en, es, fr, etc.
}

// Profile represents user profile information
type Profile struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	IsActive  bool   `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Validation methods

// Validate validates the User struct
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrEmptyEmail
	}
	if u.Username == "" {
		return ErrEmptyUsername
	}
	if u.FirstName == "" {
		return ErrEmptyFirstName
	}
	if u.LastName == "" {
		return ErrEmptyLastName
	}
	if !isValidEmail(u.Email) {
		return ErrInvalidEmail
	}
	if !isValidUsername(u.Username) {
		return ErrInvalidUsername
	}
	return nil
}

// Validate validates the Preferences struct
func (p *Preferences) Validate() error {
	validFrequencies := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}
	
	if p.DigestFrequency != "" && !validFrequencies[p.DigestFrequency] {
		return ErrInvalidDigestFrequency
	}
	
	validThemes := map[string]bool{
		"light": true,
		"dark":  true,
		"auto":  true,
	}
	
	if p.Theme != "" && !validThemes[p.Theme] {
		return ErrInvalidTheme
	}
	
	return nil
}

// Helper methods

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword verifies if the provided password matches the user's password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// ToProfile converts User to Profile (public information only)
func (u *User) ToProfile() Profile {
	return Profile{
		ID:        u.ID,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Avatar:    u.Avatar,
		Email:     u.Email,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
	}
}

// SetDefaults sets default preferences for a new user
func (p *Preferences) SetDefaults() {
	if p.DigestFrequency == "" {
		p.DigestFrequency = "daily"
	}
	if p.Theme == "" {
		p.Theme = "light"
	}
	if p.Language == "" {
		p.Language = "en"
	}
	p.NotificationEnabled = true
	p.EmailDigest = true
}

// Private validation functions

func isValidEmail(email string) bool {
	// Simple email validation - in production, use a proper email validation library
	return len(email) > 3 && len(email) < 255 && 
		   email[0] != '@' && email[len(email)-1] != '@' &&
		   countChar(email, '@') == 1
}

func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	
	for _, char := range username {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '-' {
			return false
		}
	}
	
	return true
}

func validatePassword(password string) error {
	if len(password) < 6 {
		return ErrPasswordTooShort
	}
	if len(password) > 128 {
		return ErrPasswordTooLong
	}
	
	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}
	
	if !hasUpper || !hasLower || !hasDigit {
		return ErrPasswordTooWeak
	}
	
	return nil
}

func countChar(s string, char rune) int {
	count := 0
	for _, c := range s {
		if c == char {
			count++
		}
	}
	return count
}
