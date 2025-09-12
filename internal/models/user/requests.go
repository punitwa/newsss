package user

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// UpdatePreferencesRequest represents a preferences update request
type UpdatePreferencesRequest struct {
	Categories          []string `json:"categories"`
	Sources             []string `json:"sources"`
	NotificationEnabled *bool    `json:"notification_enabled,omitempty"`
	EmailDigest         *bool    `json:"email_digest,omitempty"`
	DigestFrequency     string   `json:"digest_frequency,omitempty"`
	Theme               string   `json:"theme,omitempty"`
	Language            string   `json:"language,omitempty"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// Validation methods

// Validate validates the LoginRequest
func (r *LoginRequest) Validate() error {
	if r.Email == "" {
		return ErrEmptyEmail
	}
	if r.Password == "" {
		return ErrEmptyPassword
	}
	if !isValidEmail(r.Email) {
		return ErrInvalidEmail
	}
	return nil
}

// Validate validates the RegisterRequest
func (r *RegisterRequest) Validate() error {
	if r.Email == "" {
		return ErrEmptyEmail
	}
	if r.Username == "" {
		return ErrEmptyUsername
	}
	if r.Password == "" {
		return ErrEmptyPassword
	}
	if r.FirstName == "" {
		return ErrEmptyFirstName
	}
	if r.LastName == "" {
		return ErrEmptyLastName
	}
	
	if !isValidEmail(r.Email) {
		return ErrInvalidEmail
	}
	if !isValidUsername(r.Username) {
		return ErrInvalidUsername
	}
	if err := validatePassword(r.Password); err != nil {
		return err
	}
	
	return nil
}

// Validate validates the UpdateProfileRequest
func (r *UpdateProfileRequest) Validate() error {
	if r.Username != "" && !isValidUsername(r.Username) {
		return ErrInvalidUsername
	}
	return nil
}

// Validate validates the ChangePasswordRequest
func (r *ChangePasswordRequest) Validate() error {
	if r.CurrentPassword == "" {
		return ErrEmptyPassword
	}
	if r.NewPassword == "" {
		return ErrEmptyPassword
	}
	if err := validatePassword(r.NewPassword); err != nil {
		return err
	}
	if r.CurrentPassword == r.NewPassword {
		return ErrSamePassword
	}
	return nil
}

// Validate validates the UpdatePreferencesRequest
func (r *UpdatePreferencesRequest) Validate() error {
	if r.DigestFrequency != "" {
		validFrequencies := map[string]bool{
			"daily":   true,
			"weekly":  true,
			"monthly": true,
		}
		if !validFrequencies[r.DigestFrequency] {
			return ErrInvalidDigestFrequency
		}
	}
	
	if r.Theme != "" {
		validThemes := map[string]bool{
			"light": true,
			"dark":  true,
			"auto":  true,
		}
		if !validThemes[r.Theme] {
			return ErrInvalidTheme
		}
	}
	
	return nil
}

// ToUser converts RegisterRequest to User
func (r *RegisterRequest) ToUser() *User {
	user := &User{
		Email:     r.Email,
		Username:  r.Username,
		FirstName: r.FirstName,
		LastName:  r.LastName,
		IsActive:  true,
		IsAdmin:   false,
	}
	
	// Set password (will be hashed)
	user.SetPassword(r.Password)
	
	// Set default preferences
	user.Preferences.SetDefaults()
	
	return user
}

// ApplyToUser applies the update request to a user
func (r *UpdateProfileRequest) ApplyToUser(user *User) {
	if r.Username != "" {
		user.Username = r.Username
	}
	if r.FirstName != "" {
		user.FirstName = r.FirstName
	}
	if r.LastName != "" {
		user.LastName = r.LastName
	}
	if r.Avatar != "" {
		user.Avatar = r.Avatar
	}
}

// ApplyToPreferences applies the preferences update to user preferences
func (r *UpdatePreferencesRequest) ApplyToPreferences(prefs *Preferences) {
	if r.Categories != nil {
		prefs.Categories = r.Categories
	}
	if r.Sources != nil {
		prefs.Sources = r.Sources
	}
	if r.NotificationEnabled != nil {
		prefs.NotificationEnabled = *r.NotificationEnabled
	}
	if r.EmailDigest != nil {
		prefs.EmailDigest = *r.EmailDigest
	}
	if r.DigestFrequency != "" {
		prefs.DigestFrequency = r.DigestFrequency
	}
	if r.Theme != "" {
		prefs.Theme = r.Theme
	}
	if r.Language != "" {
		prefs.Language = r.Language
	}
}
