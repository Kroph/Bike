package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"order-service/internal/domain"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
	GetByID(ctx context.Context, id string) (domain.Order, error)
	Update(ctx context.Context, order domain.Order) error
	List(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, int, error)
	GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error)
}

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Order{}, err
	}
	defer tx.Rollback()

	order.ID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = domain.OrderStatusPending

	orderQuery := `
		INSERT INTO orders (id, user_id, status, total, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, status, total, created_at, updated_at
	`
	err = tx.QueryRowContext(
		ctx,
		orderQuery,
		order.ID,
		order.UserID,
		order.Status,
		order.Total,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.Total,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return domain.Order{}, err
	}

	for i := range order.Items {
		order.Items[i].ID = uuid.New().String()
		order.Items[i].OrderID = order.ID

		itemQuery := `
			INSERT INTO order_items (id, order_id, product_id, name, price, quantity)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.ExecContext(
			ctx,
			itemQuery,
			order.Items[i].ID,
			order.Items[i].OrderID,
			order.Items[i].ProductID,
			order.Items[i].Name,
			order.Items[i].Price,
			order.Items[i].Quantity,
		)
		if err != nil {
			return domain.Order{}, err
		}
	}

	if err = tx.Commit(); err != nil {
		return domain.Order{}, err
	}

	return order, nil
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (domain.Order, error) {
	orderQuery := `
		SELECT id, user_id, status, total, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	var order domain.Order
	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.Total,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Order{}, errors.New("order not found")
		}
		return domain.Order{}, err
	}

	itemsQuery := `
		SELECT id, order_id, product_id, name, price, quantity
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return domain.Order{}, err
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Name,
			&item.Price,
			&item.Quantity,
		)
		if err != nil {
			return domain.Order{}, err
		}
		items = append(items, item)
	}

	order.Items = items
	return order, nil
}

func (r *PostgresOrderRepository) Update(ctx context.Context, order domain.Order) error {
	query := `
		UPDATE orders
		SET status = $1, total = $2, updated_at = $3
		WHERE id = $4
	`

	order.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		order.Status,
		order.Total,
		order.UpdatedAt,
		order.ID,
	)

	return err
}

func (r *PostgresOrderRepository) List(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, int, error) {
	baseQuery := `
		SELECT id, user_id, status, total, created_at, updated_at
		FROM orders
		WHERE 1=1
	`

	countQuery := `
		SELECT COUNT(*)
		FROM orders
		WHERE 1=1
	`

	var conditions string
	var args []interface{}
	var argIndex int = 1

	if filter.UserID != "" {
		conditions += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, filter.UserID)
		argIndex++
	}

	if string(filter.Status) != "" {
		conditions += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.FromDate != nil {
		conditions += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, filter.FromDate)
		argIndex++
	}

	if filter.ToDate != nil {
		conditions += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, filter.ToDate)
		argIndex++
	}

	limit := 10
	offset := 0

	if filter.PageSize > 0 {
		limit = filter.PageSize
	}

	if filter.Page > 0 {
		offset = (filter.Page - 1) * limit
	}

	query := baseQuery + conditions + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []domain.Order
	orderMap := make(map[string]*domain.Order)
	var orderIDs []string

	for rows.Next() {
		var order domain.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.Total,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		orders = append(orders, order)
		orderMap[order.ID] = &orders[len(orders)-1]
		orderIDs = append(orderIDs, order.ID)
	}

	var total int
	err = r.db.QueryRowContext(ctx, countQuery+conditions, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if len(orderIDs) == 0 {
		return orders, total, nil
	}

	placeholders := ""
	itemArgs := make([]interface{}, 0, len(orderIDs))

	for i, id := range orderIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		itemArgs = append(itemArgs, id)
	}

	itemsQuery := fmt.Sprintf(`
		SELECT id, order_id, product_id, name, price, quantity
		FROM order_items
		WHERE order_id IN (%s)
	`, placeholders)

	itemRows, err := r.db.QueryContext(ctx, itemsQuery, itemArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var item domain.OrderItem
		err := itemRows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Name,
			&item.Price,
			&item.Quantity,
		)
		if err != nil {
			return nil, 0, err
		}

		if order, ok := orderMap[item.OrderID]; ok {
			order.Items = append(order.Items, item)
		}
	}

	return orders, total, nil
}

func (r *PostgresOrderRepository) GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	filter := domain.OrderFilter{
		UserID: userID,
	}

	orders, _, err := r.List(ctx, filter)
	return orders, err
}
