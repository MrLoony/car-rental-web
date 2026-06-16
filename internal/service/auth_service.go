package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	invalidCredentialsMessage      = "Invalid email or password."
	invalidCredentialsLaterMessage = "Invalid email or password. Please try again later."
)

type AuthService struct {
	adminUserRepo *repository.AdminUserRepository
	loginLimiter  *LoginAttemptLimiter
}

func NewAuthService(adminUserRepo *repository.AdminUserRepository, loginLimiter *LoginAttemptLimiter) *AuthService {
	if loginLimiter == nil {
		loginLimiter = NewLoginAttemptLimiter()
	}

	return &AuthService{
		adminUserRepo: adminUserRepo,
		loginLimiter:  loginLimiter,
	}
}

func (s *AuthService) Authenticate(ctx context.Context, form model.LoginForm) (model.AdminUser, model.LoginForm, error) {
	form = normalizeLoginForm(form)
	if validateLoginForm(&form); form.HasErrors() {
		return model.AdminUser{}, form, nil
	}

	if locked, _ := s.loginLimiter.IsLocked(form.Email); locked {
		addCredentialsErrorMessage(&form, invalidCredentialsLaterMessage)
		return model.AdminUser{}, form, nil
	}

	user, err := s.adminUserRepo.GetAdminUserByEmail(ctx, form.Email)
	if err != nil {
		if errors.Is(err, repository.ErrAdminUserNotFound) {
			s.loginLimiter.RecordFailure(form.Email)
			addCredentialsError(&form)
			return model.AdminUser{}, form, nil
		}

		return model.AdminUser{}, form, fmt.Errorf("get admin user by email: %w", err)
	}

	if !passwordMatchesHash(form.Password, user.PasswordHash) {
		s.loginLimiter.RecordFailure(form.Email)
		addCredentialsError(&form)
		return model.AdminUser{}, form, nil
	}

	s.loginLimiter.RecordSuccess(form.Email)
	return user, form, nil
}

func normalizeLoginForm(form model.LoginForm) model.LoginForm {
	if form.Errors == nil {
		form.Errors = make(map[string]string)
	}

	form.Email = strings.TrimSpace(form.Email)

	return form
}

func validateLoginForm(form *model.LoginForm) {
	if form.Email == "" {
		form.Errors["email"] = "Enter your email address."
	}

	if form.Password == "" {
		form.Errors["password"] = "Enter your password."
	}
}

func addCredentialsError(form *model.LoginForm) {
	addCredentialsErrorMessage(form, invalidCredentialsMessage)
}

func addCredentialsErrorMessage(form *model.LoginForm, message string) {
	form.Errors["credentials"] = message
}

func passwordMatchesHash(password string, passwordHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	return err == nil
}
