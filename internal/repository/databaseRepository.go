package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pervukhinpm/gophermart/internal/config"
	"github.com/pervukhinpm/gophermart/internal/middleware"
	"github.com/pervukhinpm/gophermart/internal/model"
	"go.uber.org/zap"
)

type Repository interface {
	AddOrder(ctx context.Context, order *model.Order) error
	GetOrder(ctx context.Context, id string) (*model.Order, error)
	GetOrders(ctx context.Context, userID string) (*[]model.Order, error)
	UpdateOrder(ctx context.Context, order *model.Order) error

	GetUserByLogin(ctx context.Context, login string) (model.User, error)
	GetUserByID(ctx context.Context, userID string) (model.User, error)
	CreateUser(ctx context.Context, user *model.User) error

	CreateWithdrawal(ctx context.Context, user model.User, withdrawal *Withdrawal) error
	GetWithdrawals(ctx context.Context, userID string) (*[]model.Withdrawal, error)

	Close() error
}

type DatabaseRepository struct {
	db *pgxpool.Pool
}

func NewDatabaseRepository(ctx context.Context, config config.Config) (*DatabaseRepository, error) {
	db, err := pgxpool.New(context.Background(), config.DataBaseURI)
	if err != nil {
		return nil, err
	}

	dbRepository := DatabaseRepository{
		db: db,
	}

	err = dbRepository.createUsersDB(ctx)
	if err != nil {
		return nil, err
	}

	err = dbRepository.createOrdersDB(ctx)
	if err != nil {
		return nil, err
	}

	err = dbRepository.createWithdrawalsDB(ctx)
	if err != nil {
		return nil, err
	}

	return &dbRepository, nil
}

func (dr *DatabaseRepository) createUsersDB(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
		id varchar UNIQUE NOT NULL,
		login varchar UNIQUE NOT NULL,
		password varchar NOT NULL,
		balance int DEFAULT 0,
		withdrawn int DEFAULT 0,
		CONSTRAINT users_pk PRIMARY KEY (id)
		);`
	_, err := dr.db.Exec(ctx, query)
	return err
}

func (dr *DatabaseRepository) createOrdersDB(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS orders (
		id varchar NOT NULL,
		user_id varchar REFERENCES users ON DELETE CASCADE,
		status varchar NOT NULL,
		uploaded_at timestamp with time zone NOT NULL,
		accrual int,
		CONSTRAINT orders_pk PRIMARY KEY (id)
	);`
	_, err := dr.db.Exec(ctx, query)
	return err
}

func (dr *DatabaseRepository) createWithdrawalsDB(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS withdrawals (
		user_id varchar REFERENCES users ON DELETE CASCADE,
		order_id varchar NOT NULL,
		processed_at timestamp with time zone NOT NULL,
		sum int,
		CONSTRAINT withdrawals_pk PRIMARY KEY (order_id)
		);`
	_, err := dr.db.Exec(ctx, query)
	return err
}

func (dr *DatabaseRepository) AddOrder(ctx context.Context, order *model.Order) error {
	query := `
	INSERT INTO orders (id, user_id, status, uploaded_at)
	VALUES ($1, $2, $3, $4);`

	result, err := dr.db.Exec(ctx, query, order.OrderNumber, order.UserID, order.Status, order.ProcessedAt)

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		middleware.Log.Info("Order already exists ", order.OrderNumber)

		existingOrder, err := dr.GetOrder(ctx, order.OrderNumber)
		if err != nil {
			middleware.Log.Error("Error getting existing order", zap.Error(err))
			return err
		}

		if existingOrder.UserID == order.UserID {
			return ErrOrderAlreadyCreatedByUser
		}

		return ErrOrderAlreadyExist
	}

	if err != nil {
		middleware.Log.Error("Error inserting order", zap.Error(err))
		return err
	}

	return nil
}

func (dr *DatabaseRepository) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	query := `
		SELECT id, user_id, status, uploaded_at, accrual
		FROM orders
		WHERE id = $1;`

	order := model.Order{}
	err := dr.db.QueryRow(ctx, query, id).Scan(
		&order.OrderNumber,
		&order.UserID,
		&order.Status,
		&order.ProcessedAt,
		&order.Accrual,
	)

	if err != nil {
		middleware.Log.Error("Error fetching order", zap.String("orderID", id), zap.Error(err))
		return nil, err
	}
	order.Accrual = order.Accrual / 100
	return &order, nil
}

func (dr *DatabaseRepository) GetOrders(ctx context.Context, userID string) (*[]model.Order, error) {
	query := `
		SELECT id, user_id, status, uploaded_at, accrual
		FROM orders
		WHERE user_id = $1;`

	rows, err := dr.db.Query(ctx, query, userID)
	if err != nil {
		middleware.Log.Error("Error fetching orders for user", zap.String("userID", userID), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		err := rows.Scan(&order.OrderNumber, &order.UserID, &order.Status, &order.ProcessedAt, &order.Accrual)
		if err != nil {
			middleware.Log.Error("Error scanning order row", zap.Error(err))
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNoOrders
			}
			return nil, err
		}
		order.Accrual = order.Accrual / 100
		orders = append(orders, order)
	}

	if len(orders) == 0 {
		return &orders, ErrNoOrders
	}

	return &orders, nil
}

func (dr *DatabaseRepository) UpdateOrder(ctx context.Context, order *model.Order) error {
	query1 := `
	UPDATE orders
	SET status = $1, accrual = $2
	WHERE id = $3;`

	query2 := `
	UPDATE users
	SET balance = balance + $1
	WHERE id = $2;`

	tx, err := dr.db.Begin(ctx)

	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, query1, order.Status, order.Accrual, order.OrderNumber)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, query2, order.Accrual, order.UserID)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (dr *DatabaseRepository) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	query := `
		SELECT id, password, balance, withdrawn
		FROM users
		WHERE login = $1;`

	user := model.User{Login: login}
	err := dr.db.QueryRow(ctx, query, login).Scan(&user.ID, &user.Password, &user.Balance, &user.Withdrawn)
	if err != nil {
		middleware.Log.Error("Error fetching user", zap.String("login", login), zap.Error(err))
		return user, err
	}

	return user, nil
}

func (dr *DatabaseRepository) GetUserByID(ctx context.Context, userID string) (model.User, error) {
	query := `
		SELECT login, password, balance, withdrawn
		FROM users
		WHERE id = $1;`

	user := model.User{ID: userID}
	err := dr.db.QueryRow(ctx, query, userID).Scan(&user.Login, &user.Password, &user.Balance, &user.Withdrawn)
	if err != nil {
		middleware.Log.Error("Error fetching user", zap.String("userID", userID), zap.Error(err))
		return user, err
	}

	return user, nil
}

func (dr *DatabaseRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, login, password)
		VALUES ($1, $2, $3);`

	_, err := dr.db.Exec(ctx, query, user.ID, user.Login, user.Password)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			middleware.Log.Warn("Duplicate user entry", zap.String("userID", user.Login), zap.String("login", user.Login))
			return ErrUserDuplicated
		}
		middleware.Log.Error("Error creating user", zap.String("userID", user.Login), zap.Error(err))
		return err
	}

	return nil
}

func (dr *DatabaseRepository) CreateWithdrawal(ctx context.Context, user model.User, withdrawal *Withdrawal) error {
	query1 := `
	UPDATE users
	SET balance = $1, withdrawn = $2
	WHERE id = $3;`

	query2 := `
	INSERT INTO withdrawals (order_id, user_id, processed_at, sum) 
	VALUES ($1, $2, $3, $4);`

	tx, err := dr.db.Begin(ctx)

	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, query1, user.Balance, user.Withdrawn, user.ID)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, query2, withdrawal.OrderID, user.ID, withdrawal.ProcessedAt, withdrawal.Amount)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (dr *DatabaseRepository) GetWithdrawals(ctx context.Context, userID string) (*[]model.Withdrawal, error) {
	query := `
		SELECT user_id, order_id, processed_at, sum
		FROM withdrawals
		WHERE user_id = $1;`

	rows, err := dr.db.Query(ctx, query, userID)
	if err != nil {
		middleware.Log.Error("Error fetching withdrawals for user", zap.String("userID", userID), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	withdrawals := make([]model.Withdrawal, 0)
	for rows.Next() {
		var withdrawal model.Withdrawal
		err := rows.Scan(&withdrawal.UserID, &withdrawal.OrderID, &withdrawal.ProcessedAt, &withdrawal.Amount)
		if err != nil {
			middleware.Log.Error("Error scanning withdrawal row", zap.Error(err))
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNoWithdrawals
			}
			return nil, err
		}
		withdrawal.Amount = withdrawal.Amount / 100
		withdrawals = append(withdrawals, withdrawal)
	}
	if len(withdrawals) == 0 {
		return &withdrawals, ErrNoWithdrawals
	}
	return &withdrawals, nil
}

func (dr *DatabaseRepository) Close() error {
	dr.db.Close()
	return nil
}
