package postgres

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
	"github.com/zYoma/gophermart/internal/config"
	"github.com/zYoma/gophermart/internal/integrations/loyalty"
	"github.com/zYoma/gophermart/internal/logger"
	"github.com/zYoma/gophermart/internal/models"
	"github.com/zYoma/gophermart/internal/storage"
)

var ErrCreatePool = errors.New("unable to create connection pool")
var ErrPing = errors.New("checking connection to the database")
var ErrURLNotFound = errors.New("url not found")
var ErrCreateUser = errors.New("crerate user")
var ErrCreateUserBalance = errors.New("crerate user balance")
var ErrCreateTable = errors.New("creating tables")
var ErrConflict = errors.New("url already exist")
var ErrRegisteresOrders = errors.New("select from database")
var ErrOrderAlredyExist = errors.New("alredy exist")
var ErrCreatedByOtherUser = errors.New("order created by other user")
var ErrUpdate = errors.New("update")
var ErrMigrate = errors.New("up migration faield")
var ErrSetDialect = errors.New("set dialect faield")
var ErrScanRows = errors.New("scan rows")
var ErrRows = errors.New("line search error")
var ErrCommit = errors.New("bot commit")
var ErrBeginTransaction = errors.New("begin transaction")
var ErrSelect = errors.New("select data from db")
var ErrOrdersNotFound = errors.New("orders for user not found")
var ErrFewPoints = errors.New("few points for operations")
var ErrWithdrawalsNotFound = errors.New("withdrawals not found")

const MigrationDir = "./internal/storage/migrations"

type Storage struct {
	db    *sql.DB
	pool  *pgxpool.Pool
	mutex sync.Mutex
}

func New(cfg *config.Config) (storage.StorageProvider, error) {
	config, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}

	// Используйте stdlib.RegisterConnConfig для регистрации настроек подключения в database/sql.
	db, err := sql.Open("pgx", stdlib.RegisterConnConfig(config.ConnConfig))
	if err != nil {
		return nil, ErrCreatePool
	}

	dbpool, err := pgxpool.New(context.Background(), cfg.DSN)
	if err != nil {
		return nil, ErrCreatePool
	}
	return &Storage{db: db, pool: dbpool}, nil
}

// для накатывания миграций при старте приложения
func (s *Storage) Init() error {
	if err := goose.SetDialect("postgres"); err != nil {
		return ErrSetDialect
	}

	if err := goose.Up(s.db, MigrationDir); err != nil {
		return ErrMigrate
	}

	return nil
}

// создает пользователя
func (s *Storage) CreateUser(ctx context.Context, login string, password string) error {

	// Начало транзакции
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		logger.Log.Sugar().Errorf("Ошибка при начале транзакции: %s", err)
		return ErrBeginTransaction
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				logger.Log.Sugar().Errorf("Ошибка при откате транзакции: %s", rbErr)
			}
		}
	}()

	_, err = tx.Exec(ctx, `
        INSERT INTO users (login, password) VALUES ($1, $2);
    `, login, password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return ErrConflict
		}
		logger.Log.Sugar().Errorf("Не удалось создать пользователя: %s", err)
		return ErrCreateUser
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO user_balance (user_login, current, withdrawn) VALUES ($1, 0, 0);
    `, login)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return ErrConflict
		}
		logger.Log.Sugar().Errorf("Не удалось создать баланс пользователя: %s", err)
		return ErrCreateUserBalance
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		logger.Log.Sugar().Errorf("Ошибка при фиксации транзакции: %s", commitErr)
		return ErrCommit
	}

	return nil
}

// получает хеш пароля, для авторизации
func (s *Storage) GetPasswordHash(ctx context.Context, login string) (string, error) {

	var userPassword string
	row := s.pool.QueryRow(ctx, `SELECT password FROM users WHERE login = $1;`, login)
	err := row.Scan(&userPassword)
	if err != nil {
		return "", err
	}

	return userPassword, nil
}

// создает заказ
func (s *Storage) CreateOrder(ctx context.Context, number string, login string) error {

	var userLogin string
	var isCreated bool

	row := s.pool.QueryRow(ctx, `
		INSERT INTO orders (number, user_login, status)
		VALUES ($1, $2, $3)
		ON CONFLICT (number, user_login) DO UPDATE
		SET user_login = EXCLUDED.user_login, updated_at = NOW()
        RETURNING user_login, (xmax = 0) AS is_created;
    `, number, login, "PROCESSING")

	err := row.Scan(&userLogin, &isCreated)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return ErrCreatedByOtherUser
		}
		logger.Log.Sugar().Errorf("Не удалось создать заказ: %s", err)
		return ErrCreateUser
	}

	if isCreated {
		return nil
	} else {
		return ErrOrderAlredyExist
	}

}

// получает заказы в статусе REGISTERED или PROCESSING
func (s *Storage) GetRegisteresOrders(ctx context.Context) ([]string, error) {

	var orders []string
	rows, err := s.pool.Query(ctx, `SELECT number FROM orders WHERE status = 'REGISTERED' OR status = 'PROCESSING';`)
	if err != nil {
		logger.Log.Sugar().Errorf("Не удалось выполнить запрос: %s", err)
		return nil, ErrRegisteresOrders
	}
	defer rows.Close()

	for rows.Next() {
		var number string
		if err := rows.Scan(&number); err != nil {
			logger.Log.Sugar().Errorf("Ошибка при сканировании строки: %s", err)
			return nil, ErrScanRows
		}
		orders = append(orders, number)
	}

	if err = rows.Err(); err != nil {
		logger.Log.Sugar().Errorf("Ошибка при итерации по строкам: %s", err)
		return nil, ErrRows
	}

	return orders, nil
}

// в одной транзакции обновляет заказ и начисляет баллы
func (s *Storage) UpdateOrderAndAccrualPoints(ctx context.Context, orderData *loyalty.OrderResponse) error {
	// Начало транзакции
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		logger.Log.Sugar().Errorf("Ошибка при начале транзакции: %s", err)
		return ErrBeginTransaction
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				logger.Log.Sugar().Errorf("Ошибка при откате транзакции: %s", rbErr)
			}
		}
	}()

	var userLogin string

	if orderData.Status == "PROCESSED" {
		// Обновляем заказ и получаем user_login
		err = tx.QueryRow(ctx, `
            UPDATE orders SET status = $1, accrual = $2 WHERE number = $3 RETURNING user_login;
        `, orderData.Status, orderData.Accrual, orderData.Order).Scan(&userLogin)
		if err != nil {
			return ErrUpdate
		}

		// Используем полученный user_login для обновления баланса пользователя
		_, err = tx.Exec(ctx, `
            UPDATE user_balance SET current = current + $1 WHERE user_login = $2;
        `, orderData.Accrual, userLogin)
		if err != nil {
			return ErrUpdate
		}
	} else if orderData.Status == "INVALID" || orderData.Status == "PROCESSING" {
		// Обновляем статус заказа без начисления баллов
		_, err = tx.Exec(ctx, `
            UPDATE orders SET status = $1 WHERE number = $2;
        `, orderData.Status, orderData.Order)
		if err != nil {
			return ErrUpdate
		}
	}
	// Другие статусы не обрабатываем

	if commitErr := tx.Commit(ctx); commitErr != nil {
		logger.Log.Sugar().Errorf("Ошибка при фиксации транзакции: %s", commitErr)
		return commitErr
	}

	return nil
}

// получает заказов пользователя
func (s *Storage) GetUserOrders(ctx context.Context, userLogin string) ([]models.Order, error) {

	var orders []models.Order
	rows, err := s.pool.Query(ctx, `SELECT number, status, accrual, uploaded_at FROM orders WHERE user_login = $1 ORDER BY uploaded_at desc;`, userLogin)
	if err != nil {
		logger.Log.Sugar().Errorf("Не удалось выполнить запрос: %s", err)
		return nil, ErrSelect
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			logger.Log.Sugar().Errorf("Ошибка при сканировании строки: %s", err)
			return nil, ErrScanRows
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		logger.Log.Sugar().Errorf("Ошибка при итерации по строкам: %s", err)
		return nil, ErrRows
	}

	if len(orders) == 0 {
		return nil, ErrOrdersNotFound
	}

	return orders, nil
}

// получает баланс пользователя
func (s *Storage) GetUserBalance(ctx context.Context, userLogin string) (models.Balance, error) {
	var userBalance models.Balance
	row := s.pool.QueryRow(ctx, `SELECT current, withdrawn FROM user_balance WHERE user_login = $1;`, userLogin)

	err := row.Scan(&userBalance.Current, &userBalance.Withdrawn)
	if err != nil {
		// Другая ошибка выполнения запроса
		logger.Log.Sugar().Errorf("Не удалось выполнить запрос: %s", err)
		return models.Balance{}, err
	}

	return userBalance, nil
}

// в рамках транзакции списание баллов с баланса и создании записи об этом
func (s *Storage) Withdrow(ctx context.Context, sum float64, userLogin string, order string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Начало транзакции
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		logger.Log.Sugar().Errorf("Ошибка при начале транзакции: %s", err)
		return ErrBeginTransaction
	}

	defer func() {
		if err != nil {
			logger.Log.Sugar().Errorf("Ошибка: %s", err)
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				logger.Log.Sugar().Errorf("Ошибка при откате транзакции: %s", rbErr)
			}
		}
	}()

	_, err = tx.Exec(ctx, `
		UPDATE user_balance SET current = current - $1, withdrawn = withdrawn + $2 WHERE user_login = $3;
    `, sum, sum, userLogin)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			// Проверка кода ошибки на соответствие коду нарушения ограничения CHECK
			if pgErr.Code == "23514" {
				logger.Log.Sugar().Errorf("невозможно выполнить операцию: недостаточно средств на балансе: %s", err)
				return ErrFewPoints
			}
		}
		return ErrUpdate
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO withdrawals ("order", sum, user_login) VALUES ($1, $2, $3) ;
    `, order, sum, userLogin)
	if err != nil {
		return ErrUpdate
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		logger.Log.Sugar().Errorf("Ошибка при фиксации транзакции: %s", commitErr)
		return commitErr
	}

	return nil
}

// получает инфо о выводах средств
func (s *Storage) GetUserWithdrawals(ctx context.Context, userLogin string) ([]models.Withdrawn, error) {

	var withdrawals []models.Withdrawn
	rows, err := s.pool.Query(ctx, `SELECT "order", sum, proccesed_at FROM withdrawals WHERE user_login = $1 ORDER BY proccesed_at desc;`, userLogin)
	if err != nil {
		logger.Log.Sugar().Errorf("Не удалось выполнить запрос: %s", err)
		return nil, ErrSelect
	}
	defer rows.Close()

	for rows.Next() {
		var withdraw models.Withdrawn
		if err := rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.ProccesedAt); err != nil {
			logger.Log.Sugar().Errorf("Ошибка при сканировании строки: %s", err)
			return nil, ErrScanRows
		}
		withdrawals = append(withdrawals, withdraw)
	}

	if err = rows.Err(); err != nil {
		logger.Log.Sugar().Errorf("Ошибка при итерации по строкам: %s", err)
		return nil, ErrRows
	}

	if len(withdrawals) == 0 {
		return nil, ErrWithdrawalsNotFound
	}

	return withdrawals, nil
}
