package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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

	Migrate() error
}

type DatabaseRepository struct {
	pool *pgxpool.Pool
}

func NewDatabaseRepository(pool *pgxpool.Pool) *DatabaseRepository {
	return &DatabaseRepository{
		pool: pool,
	}
}

func (dr *DatabaseRepository) Migrate() error {
	return migrate(dr.pool)
}

func (dr *DatabaseRepository) AddOrder(ctx context.Context, order *model.Order) error {
	query := `
	INSERT INTO orders (id, user_id, status, uploaded_at)
	VALUES ($1, $2, $3, $4);`

	result, err := dr.pool.Exec(ctx, query, order.OrderNumber, order.UserID, order.Status, order.ProcessedAt)

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
	err := dr.pool.QueryRow(ctx, query, id).Scan(
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
		WHERE user_id = $1
		ORDER BY uploaded_at DESC;`

	rows, err := dr.pool.Query(ctx, query, userID)
	if err != nil {
		middleware.Log.Error("Error fetching orders for user", zap.String("userID", userID), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var order model.Order
		var intAccrual int
		scanErr := rows.Scan(&order.OrderNumber, &order.UserID, &order.Status, &order.ProcessedAt, &intAccrual)
		if scanErr != nil {
			middleware.Log.Error("Error scanning order row", zap.Error(scanErr))
			return nil, scanErr
		}
		order.Accrual = float64(intAccrual) / 100
		orders = append(orders, order)
	}

	if len(orders) == 0 {
		return nil, ErrNoOrders
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

	tx, err := dr.pool.Begin(ctx)

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
	err := dr.pool.QueryRow(ctx, query, login).Scan(&user.ID, &user.Password, &user.Balance, &user.Withdrawn)
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
	err := dr.pool.QueryRow(ctx, query, userID).Scan(&user.Login, &user.Password, &user.Balance, &user.Withdrawn)
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

	_, err := dr.pool.Exec(ctx, query, user.ID, user.Login, user.Password)

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

	tx, err := dr.pool.Begin(ctx)

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
		WHERE user_id = $1
		ORDER BY processed_at DESC;`

	rows, err := dr.pool.Query(ctx, query, userID)
	if err != nil {
		middleware.Log.Error("Error fetching withdrawals for user", zap.String("userID", userID), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	withdrawals := make([]model.Withdrawal, 0)
	for rows.Next() {
		var withdrawal model.Withdrawal
		err = rows.Scan(&withdrawal.UserID, &withdrawal.OrderID, &withdrawal.ProcessedAt, &withdrawal.Amount)
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
