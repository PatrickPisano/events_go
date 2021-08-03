package services

import (
	"events/pkg/events"
	"events/pkg/storage/postgres"
	"golang.org/x/crypto/bcrypt"
)

func NewUserService(r *postgres.UserStorage) *UserService {
	return &UserService{r}
}

type UserService struct {
	r *postgres.UserStorage
}

func (s *UserService) CreateUser(u *events.User, password string) (int, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return 0, err
	}

	ee, err := s.r.SaveUser(u, string(hash))
	if err != nil {
		return 0, err
	}

	return ee, nil
}

func (s *UserService) EmailMatchPassword(email string, password string) (bool, int, error) {
	id, hashedPassword, err := s.r.UserIDAndPasswordByEmail(email)
	if err != nil {
		return false, 0, err
	}

	bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, 0, nil
	} else if err != nil {
		return false, 0, err
	}

	return true, id, nil
}

func (s *UserService) User(uid int) (*events.User, error) {
	u, err := s.r.User(uid)

	if err != nil {
		return nil, err
	}

	return u, nil
}
