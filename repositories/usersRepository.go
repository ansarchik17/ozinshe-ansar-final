package repositories

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"goozinshe/logger"
	"goozinshe/models"
)

type UsersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(conn *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{db: conn}
}

func (r *UsersRepository) FindAll(c context.Context) ([]models.User, error) {
	logger := logger.GetLogger()
	rows, err := r.db.Query(c, "select * from users")
	if err != nil {
		logger.Error("Could not find all users", zap.Error(err))
		return nil, err
	}
	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.PasswordHash)
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
		users = append(users, user)
	}
	if rows.Err() != nil {
		return nil, err
	}
	return users, nil
}

func (r *UsersRepository) Create(c context.Context, user models.User) (int, error) {
	logger := logger.GetLogger()
	var id int
	err := r.db.QueryRow(c, "insert into users(name, email, password_hash) values($1, $2, $3) returning id", user.Name, user.Email, user.PasswordHash).Scan(&id)
	if err != nil {
		logger.Error("Could not insert user", zap.String("db_msg", err.Error()))
		return 0, err
	}
	return id, err
}

func (r *UsersRepository) FindById(c context.Context, id int) (models.User, error) {
	logger := logger.GetLogger()
	row := r.db.QueryRow(c, "select id, name, email, password_hash from users where id = $1", id)
	var user models.User
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		logger.Error("Could not find user by id", zap.String("db_msg", err.Error()))
		return models.User{}, err
	}
	return user, err
}

func (r *UsersRepository) FindByEmail(c context.Context, email string) (models.User, error) {
	logger := logger.GetLogger()
	row := r.db.QueryRow(c, "select id, name, email, password_hash from users where email = $1", email)
	var user models.User
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.PasswordHash)
	if err != nil {
		logger.Error("Could not find user by email", zap.String("db_msg", err.Error()))
		return models.User{}, err
	}
	return user, err
}

func (r *UsersRepository) Update(c context.Context, id int, user models.User) error {
	logger := logger.GetLogger()
	_, err := r.db.Exec(c, "update users set name = $1, email = $2, password_hash = $3 where id = $4", user.Name, user.Email, user.PasswordHash, id)
	if err != nil {
		logger.Error("Could not update user", zap.String("db_msg", err.Error()))
		return err
	}
	return err
}

func (r *UsersRepository) ChangePassword(c context.Context, id int, passwordHash string) error {
	logger := logger.GetLogger()
	_, err := r.db.Exec(c, "update users set password_hash = $1 where id = $2", passwordHash, id)
	if err != nil {
		logger.Error("Could not change password hash", zap.String("db_msg", err.Error()))
		return err
	}
	return err
}

func (r *UsersRepository) Delete(c context.Context, id int) error {
	logger := logger.GetLogger()
	_, err := r.db.Exec(c, "delete from users where id = $1", id)
	if err != nil {
		logger.Error("Could not delete user", zap.String("db_msg", err.Error()))
		return err
	}
	return err
}
