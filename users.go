package gus

import (
	"database/sql"
	"time"
	"github.com/asaskevich/govalidator"
	"golang.org/x/crypto/bcrypt"
)

const (
	ERR_STRING_EMAIL_CONSTRAINT string = "UNIQUE constraint failed: users.email"
)

var (
	ErrEmailTaken        error = ErrInvalid("That email is taken.")
	ErrEmailInvalid      error = ErrInvalid("'email' invalid.")
	ErrEmailRequired     error = ErrInvalid("'email' required.")
	ErrPasswordRequired  error = ErrInvalid("'password' required.")
	ErrInvalidResetToken error = ErrInvalid("Invalid reset token.")
	ErrPasswordInvalid   error = ErrInvalid(
		"'new_password' must contain: 1 Upper, 1 Lower, 1 Number and 8 Chars",
		"OR any alphanumeric with a minimum of 15 chars.")
	ResetTokenExpirySeconds int64
	ResetTokenExpiryKey     = "RESET_TOKEN_EXPIRY"
)

func NewUsers(db *sql.DB) *Users {
	return &Users{db: db}
}

type Users struct {
	db *sql.DB
}

func NewCreateUserParams() CreateUserParams {
	return CreateUserParams{}
}

type CreateUserParams struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	OrgId     int64 `json:"org_id"`
	Role      Role `json:"role"`
}

func (cp *CreateUserParams) Validate() error {
	if !govalidator.IsEmail(cp.Email) {
		return ErrEmailRequired
	}
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (us *Users) Create(p CreateUserParams) (*User, string, error) {
	stmt, err := us.db.Prepare("INSERT INTO users(email, first_name, last_name, phone, password_hash, org_id, updated, created, deleted, role) values(?,?,?,?,?,?,?,?, ?, ?)")
	if err != nil {
		return nil, "", err
	}
	u := &User{Email: p.Email, FirstName: p.FirstName, LastName: p.LastName, Phone: p.Phone,
		OrgId:    p.OrgId, Created: time.Now(), Updated: time.Now(), Role: p.Role}

	password := RandStringBytesMask(15)
	hash, err := hashPassword(password)
	if err != nil {
		return nil, "", err
	}

	res, err := stmt.Exec(u.Email, u.FirstName, u.LastName, u.Phone, hash, u.OrgId, u.Updated, u.Created, 0, u.Role)
	if err != nil {
		if err.Error() == ERR_STRING_EMAIL_CONSTRAINT {
			return nil, "", ErrEmailTaken
		}
		return nil, "", err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, "", err
	}
	u.Id = id
	return u, password, nil
}

func (us *Users) Get(id int64) (*User, error) {
	stmt, err := us.db.Prepare("SELECT id, email, first_name, last_name, phone, org_id, created, updated, role from users WHERE id =  ? AND deleted = 0 LIMIT 1")
	if err != nil {
		return nil, err
	}
	row := stmt.QueryRow(id)
	var u User
	err = CheckNotFound(row.Scan(&u.Id, &u.Email, &u.FirstName, &u.LastName, &u.Phone, &u.OrgId,
		&u.Created, &u.Updated, &u.Role))
	if err != nil {
		return nil, err
	}
	return &u, err
}

type SignInParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *SignInParams) Validate() error {
	if govalidator.IsNull(s.Password) {
		return ErrPasswordRequired
	}
	if govalidator.IsNull(s.Email) {
		return ErrEmailRequired
	}
	if !govalidator.IsEmail(s.Email) {
		return ErrEmailInvalid
	}
	return nil
}

// GetByEmail returns a user and password hash
func (us *Users) GetByEmail(email string) (*User, string, error) {
	stmt, err := us.db.Prepare("SELECT password_hash, id, email, first_name, last_name, phone, org_id, created, updated, role from users WHERE email = ? AND deleted = 0 LIMIT 1")
	if err != nil {
		return nil, "", err
	}
	row := stmt.QueryRow(email)
	var u User
	var passwordHash string
	err = CheckNotFound(row.Scan(&passwordHash, &u.Id, &u.Email, &u.FirstName, &u.LastName, &u.Phone,
		&u.OrgId, &u.Created, &u.Updated, &u.Role))
	if err != nil {
		return nil, "", err
	}
	return &u, passwordHash, err
}

func (us *Users) Authenticate(p SignInParams) (*User, error) {
	u, hash, err := us.GetByEmail(p.Email)
	if err != nil {
		_, ok := err.(*NotFoundError)
		if ok {
			return nil, ErrNotAuth
		}
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(p.Password))
	if err != nil {
		return nil, ErrNotAuth
	}
	return u, nil
}

type UpdateUserParams struct {
	Id        int64 `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      Role `json:"role"`
}

func (up *UpdateUserParams) Validate() error {
	if up.Email != "" && !govalidator.IsEmail(up.Email) {
		return ErrInvalid("'email' invalid.")
	}
	return nil
}

func (us *Users) Update(u UpdateUserParams) error {
	stmt, err := us.db.Prepare("UPDATE users SET first_name = ?, last_name = ?, email = ?, phone = ?, updated = ? WHERE id = ? AND deleted = 0")
	if err != nil {
		return err
	}
	err = CheckUpdated(stmt.Exec(u.FirstName, u.LastName, u.Email, u.Phone, time.Now(), u.Id))
	if err != nil && err.Error() == ERR_STRING_EMAIL_CONSTRAINT {
		return ErrEmailTaken
	}
	return err
}

func (us *Users) Delete(id int64) error {
	stmt, err := us.db.Prepare("UPDATE users SET deleted = 1, updated = ? WHERE id = ? AND deleted = 0")
	if err != nil {
		return err
	}
	return CheckUpdated(stmt.Exec(time.Now(), id))
}

type ListUsersParams struct {
	OrgId int64 `json:"org_id"`
}

func (pm *ListUsersParams) Validate() error {
	return nil
}

func (us *Users) List(p ListUsersParams) ([]*User, error) {
	stmt, err := us.db.Prepare("SELECT id, email, first_name, last_name, phone, org_id, created, updated, role from users WHERE deleted = 0 ORDER by updated DESC")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []*User{}
	for rows.Next() {
		u := &User{}
		rows.Scan(&u.Id, &u.Email, &u.FirstName, &u.LastName, &u.Phone, &u.OrgId, &u.Created, &u.Updated, &u.Role)
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

type ResetPasswordParams struct {
	Email string `json:"email"`
}

func (p *ResetPasswordParams) Validate() error {
	if govalidator.IsNull(p.Email) {
		return ErrEmailRequired
	}
	if !govalidator.IsEmail(p.Email) {
		return ErrEmailInvalid
	}
	return nil
}

func (us *Users) ResetPassword(p ResetPasswordParams) (string, error) {
	u, _, err := us.GetByEmail(p.Email)
	if err != nil {
		return "", err
	}
	token := RandStringBytesMask(256)
	stmt, err := us.db.Prepare("INSERT into password_resets (user_id, email, reset_token, created, deleted) values (?, ?, ?, ?, ?)")
	_, err = stmt.Exec(u.Id, u.Email, token, time.Now(), 0)
	if err != nil {
		return "", err
	}
	return token, nil
}

type ChangePasswordParams struct {
	Email            string `json:"email"`
	ExistingPassword string `json:"existing_password"`
	NewPassword      string `json:"new_password"`
	ResetToken       string `json:"reset_token"`
}

func (s *ChangePasswordParams) Validate() error {
	if govalidator.IsNull(s.Email) {
		return ErrEmailRequired
	}
	if !govalidator.IsEmail(s.Email) {
		return ErrEmailInvalid
	}
	if govalidator.IsNull(s.ExistingPassword) && govalidator.IsNull(s.ResetToken) {
		return ErrInvalid("'existing_password' or 'reset_token' required.")
	}
	if govalidator.IsNull(s.NewPassword) {
		return ErrInvalid("'new_password' is required.")
	}
	if !ValidatePassword(s.NewPassword) {
		return ErrPasswordInvalid
	}
	return nil
}

func (us *Users) ChangePassword(p ChangePasswordParams) error {
	if p.ExistingPassword != "" {
		_, err := us.Authenticate(SignInParams{Email: p.Email, Password: p.ExistingPassword})
		if err != nil {
			return err
		}
	}
	if p.ResetToken != "" {
		stmt, err := us.db.Prepare(
			"SELECT reset_token FROM password_resets where email = ? and created > ? and deleted = 0 " +
				"ORDER BY created DESC LIMIT 1")
		exp := time.Now().Add(-time.Second * time.Duration(ResetTokenExpirySeconds))
		row := stmt.QueryRow(p.Email, exp)
		var resetToken string
		err = CheckNotFound(row.Scan(&resetToken))
		if err != nil {
			return err
		}
		if resetToken != p.ResetToken {
			return ErrInvalidResetToken
		}
	}
	hash, err := hashPassword(p.NewPassword)
	if err != nil {
		return err
	}
	stmt, err := us.db.Prepare("UPDATE users SET password_hash = ?, updated = ? WHERE email = ? AND deleted = 0")
	err = CheckNotFound(err)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(hash, time.Now(), p.Email)
	return nil
}
