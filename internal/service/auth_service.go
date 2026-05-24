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

const invalidCredentialsMessage = "Invalid email or password."

type AuthService struct {
	adminUserRepo *repository.AdminUserRepository
}

func NewAuthService(adminUserRepo *repository.AdminUserRepository) *AuthService {
	return &AuthService{adminUserRepo: adminUserRepo}
}

func (s *AuthService) Authenticate(ctx context.Context, form model.LoginForm) (model.AdminUser, model.LoginForm, error) {
	form = normalizeLoginForm(form)
	if validateLoginForm(&form); form.HasErrors() {
		return model.AdminUser{}, form, nil
	}

	user, err := s.adminUserRepo.GetAdminUserByEmail(ctx, form.Email)
	if err != nil {
		if errors.Is(err, repository.ErrAdminUserNotFound) {
			addCredentialsError(&form)
			return model.AdminUser{}, form, nil
		}

		return model.AdminUser{}, form, fmt.Errorf("get admin user by email: %w", err)
	}

	if !passwordMatchesHash(form.Password, user.PasswordHash) {
		addCredentialsError(&form)
		return model.AdminUser{}, form, nil
	}

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
		form.Errors["email"] = "Email is required."
	}

	if form.Password == "" {
		form.Errors["password"] = "Password is required."
	}
}

func addCredentialsError(form *model.LoginForm) {
	form.Errors["credentials"] = invalidCredentialsMessage
}

func passwordMatchesHash(password string, passwordHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	return err == nil
}
