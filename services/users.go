package services

import (
	"database/sql"
	"fmt"
	"net/http"
	"zero/types"
	"zero/utils"

	"github.com/charmbracelet/log"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/glebarez/go-sqlite"
)

type UsersService struct {
	DB *sql.DB
}

func (us *UsersService) ComparePasswords(hash string, plain []byte) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), plain)
	return err == nil
}

func (us *UsersService) CreateUser(user *types.User) error {
	rows, err := us.DB.Exec("INSERT INTO users (customer, username, email, password) VALUES (?,?,?,?)",
		user.Customer, user.Username, user.Email, user.Password)
	if err != nil {
		err = utils.HandleSQLiteErrors(err)
		log.Error("failed to insert into the database", "err", err)
		return err
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		log.Error("failed to get affacted rows", "err", err)
		return err
	}
	if affected == 0 {
		log.Error("failed to create user", "err", err)
		return err
	}
	return nil
}

func (us *UsersService) GetUserById(id int) (*types.User, error) {
	rows, err := us.DB.Query("SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		err = utils.HandleSQLiteErrors(err)
		log.Error("failed to run query on the database", "err", err)
		return nil, err
	}
	user := new(types.User)
	for rows.Next() {
		user, err = scanUsersRow(rows)
		if err != nil {
			return nil, err
		}
	}
	if user.Id == 0 {
		err := fmt.Errorf("user non existent")
		log.Warn("user non existent")
		return nil, err
	}
	if user == nil {
		log.Error("internal error", "err", err)
		return nil, err
	}
	return user, nil
}

func (us *UsersService) GetUserByEmail(email string) (*types.User, error) {
	rows, err := us.DB.Query("SELECT * FROM users WHERE email = ?", email)
	if err != nil {
		err = utils.HandleSQLiteErrors(err)
		log.Error("failed to run query on the database", "err", err)
		return nil, err
	}
	user := new(types.User)
	for rows.Next() {
		user, err = scanUsersRow(rows)
		if err != nil {
			return nil, err
		}
	}
	if user == nil || user.Id == 0 {
		err := fmt.Errorf("email not found")
		log.Error("email not found", "err", err)
		return nil, err
	}
	return user, nil
}

func (us *UsersService) GetCustomer(user *types.User) (string, error) {
	rows, err := us.DB.Query("SELECT * FROM users WHERE id = ?", user.Id)
	if err != nil {
		err = utils.HandleSQLiteErrors(err)
		log.Error("failed to run query on the database", "err", err)
		return "", err
	}
	for rows.Next() {
		user, err = scanUsersRow(rows)
		if err != nil {
			return "", err
		}
	}
	return user.Customer, nil
}

func (us *UsersService) UpdateUserName(id int, new, old string) error {
	rows, err := us.DB.Exec("UPDATE users SET username = ? WHERE username = ? AND id = ?", new, old, id)
	if err != nil {
		err = utils.HandleSQLiteErrors(err)
		log.Error("failed to update the database", "err", err)
		return err
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		log.Error("failed to get affacted rows", "err", err)
		return err
	}
	if affected == 0 {
		log.Error("failed to update user name", "err", err)
		return err
	}
	return nil
}

func (us *UsersService) UpdateUserEmail(id int, new, old string) error {
	rows, err := us.DB.Exec("UPDATE users SET email = ? WHERE email = ? AND id = ?", new, old, id)
	if err != nil {
		err = utils.HandleSQLiteErrors(err)
		log.Error("failed to update the database", "err", err)
		return err
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		log.Error("failed to get affacted rows", "err", err)
		return err
	}
	if affected == 0 {
		log.Error("failed to update user email", "err", err)
		return err
	}
	return nil
}

func (us *UsersService) UpdateUserPassword(id int, new string) error {
	rows, err := us.DB.Exec("UPDATE users SET password = ? WHERE id = ?", new, id)
	if err != nil {
		err = utils.HandleSQLiteErrors(err)
		log.Error("failed to update the database", "err", err)
		return err
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		log.Error("failed to get affacted rows", "err", err)
		return err
	}
	if affected == 0 {
		log.Error("failed to update user password", "err", err)
		return err
	}
	return nil
}

func (us *UsersService) DeleteUser(id int) error {
	rows, err := us.DB.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		err = utils.HandleSQLiteErrors(err)
		log.Error("failed to delete user", "err", err)
		return err
	}
	affacted, err := rows.RowsAffected()
	if affacted == 0 {
		log.Error("trying to delete unexistent user")
		return fmt.Errorf("trying to delete unexistent user")
	}
	if err != nil {
		log.Error("failed to get affected rows by delete", "err", err)
		return err
	}
	return nil
}

func (us *UsersService) Verified(id int) (bool, error) {
	rows, err := us.DB.Query("SELECT * FROM users WHERE id = ? AND verified = ?", id, 1)
	if err != nil {
		err = utils.HandleSQLiteErrors(err)
		log.Error("failed to run query on the database", "err", err)
		return false, err
	}
	var user *types.User
	for rows.Next() {
		user, err = scanUsersRow(rows)
		if err != nil {
			return false, err
		}
	}
	if user != nil {
		return true, nil
	}
	return false, nil
}

func GetUserIdFromContext(r *http.Request) (int, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Error("failed to get claims from context", "err", err)
		return 0, err
	}
	id, ok := claims["user_id"].(float64)
	if !ok {
		log.Error("user_id not found in claims or is not a float64")
		return 0, fmt.Errorf("user_id not found in claims or is not a float64")
	}
	return int(id), nil
}

func scanUsersRow(row *sql.Rows) (*types.User, error) {
	user := new(types.User)
	err := row.Scan(
		&user.Id,
		&user.Customer,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Verified,
		&user.Updated,
		&user.Created,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}
