package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"order-service/internal/domain"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
	GetByID(ctx context.Context, id string) (domain.Order, error)
	Update(ctx context.Context, order domain.Order) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, int, error)
	GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error)
	// Add missing methods
	ExistsByID(ctx context.Context, id string) (bool, error)
	UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error
	GetOrdersByStatus(ctx context.Context, status domain.OrderStatus) ([]domain.Order, error)
	GetOrdersByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.Order, error)
}

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) OrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	// Validate input
	if err := r.validateOrder(order); err != nil {
		return domain.Order{}, err
	}

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Order{}, errors.New("failed to start transaction")
	}
	defer tx.Rollback()

	// Set order defaults
	order.ID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	if order.Status == "" {
		order.Status = domain.OrderStatusPending
	}

	// Calculate total if not provided
	if order.Total == 0 {
		order.Total = r.calculateTotal(order.Items)
	}

	// Insert order
	orderQuery := `
		INSERT INTO orders (id, user_id, status, total, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, status, total, created_at, updated_at`

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
		return domain.Order{}, errors.New("failed to create order")
	}

	// Insert order items
	for i := range order.Items {
		order.Items[i].ID = uuid.New().String()
		order.Items[i].OrderID = order.ID

		itemQuery := `
			INSERT INTO order_items (id, order_id, product_id, name, price, quantity, 
			                         frame_size, wheel_size, color, bike_type)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

		_, err = tx.ExecContext(
			ctx,
			itemQuery,
			order.Items[i].ID,
			order.Items[i].OrderID,
			order.Items[i].ProductID,
			order.Items[i].Name,
			order.Items[i].Price,
			order.Items[i].Quantity,
			nullString(order.Items[i].FrameSize),
			nullString(order.Items[i].WheelSize),
			nullString(order.Items[i].Color),
			nullString(order.Items[i].BikeType),
		)
		if err != nil {
			return domain.Order{}, errors.New("failed to create order item")
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return domain.Order{}, errors.New("failed to commit transaction")
	}

	return order, nil
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (domain.Order, error) {
	if id == "" {
		return domain.Order{}, errors.New("order ID is required")
	}

	// Get order
	orderQuery := `
		SELECT id, user_id, status, total, created_at, updated_at
		FROM orders
		WHERE id = $1`

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
		return domain.Order{}, errors.New("failed to get order")
	}

	// Get order items
	items, err := r.getOrderItems(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}

	order.Items = items
	return order, nil
}

func (r *PostgresOrderRepository) Update(ctx context.Context, order domain.Order) error {
	if order.ID == "" {
		return errors.New("order ID is required")
	}

	if err := r.validateOrder(order); err != nil {
		return err
	}

	order.UpdatedAt = time.Now()

	query := `
		UPDATE orders
		SET status = $1, total = $2, updated_at = $3
		WHERE id = $4`

	result, err := r.db.ExecContext(
		ctx,
		query,
		order.Status,
		order.Total,
		order.UpdatedAt,
		order.ID,
	)
	if err != nil {
		return errors.New("failed to update order")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to check update result")
	}

	if rowsAffected == 0 {
		return errors.New("order not found")
	}

	return nil
}

func (r *PostgresOrderRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("order ID is required")
	}

	// Start transaction to delete order and items
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.New("failed to start transaction")
	}
	defer tx.Rollback()

	// Delete order items first (due to foreign key constraint)
	_, err = tx.ExecContext(ctx, "DELETE FROM order_items WHERE order_id = $1", id)
	if err != nil {
		return errors.New("failed to delete order items")
	}

	// Delete order
	result, err := tx.ExecContext(ctx, "DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		return errors.New("failed to delete order")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to check delete result")
	}

	if rowsAffected == 0 {
		return errors.New("order not found")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return errors.New("failed to commit transaction")
	}

	return nil
}

func (r *PostgresOrderRepository) List(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, int, error) {
	baseQuery := `
		SELECT id, user_id, status, total, created_at, updated_at
		FROM orders`

	countQuery := `SELECT COUNT(*) FROM orders`

	whereClause, args := r.buildWhereClause(filter)

	if whereClause != "" {
		baseQuery += " WHERE " + whereClause
		countQuery += " WHERE " + whereClause
	}

	// Add pagination
	limit := 10
	offset := 0

	if filter.PageSize > 0 {
		limit = filter.PageSize
	}

	if filter.Page > 0 {
		offset = (filter.Page - 1) * limit
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	// Execute main query
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, errors.New("failed to list orders")
	}
	defer rows.Close()

	var orders []domain.Order
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
			return nil, 0, errors.New("failed to scan order")
		}

		orders = append(orders, order)
		orderIDs = append(orderIDs, order.ID)
	}

	// Get total count
	var total int
	countArgs := args[:len(args)-2] // Remove limit and offset
	err = r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, errors.New("failed to get order count")
	}

	// Load items for all orders
	if len(orderIDs) > 0 {
		itemsMap, err := r.getOrderItemsMap(ctx, orderIDs)
		if err != nil {
			return nil, 0, err
		}

		// Assign items to orders
		for i := range orders {
			orders[i].Items = itemsMap[orders[i].ID]
		}
	}

	return orders, total, nil
}

func (r *PostgresOrderRepository) GetUserOrders(ctx context.Context, userID string) ([]domain.Order, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	filter := domain.OrderFilter{
		UserID:   userID,
		PageSize: 100, // Get up to 100 orders
	}

	orders, _, err := r.List(ctx, filter)
	return orders, err
}

func (r *PostgresOrderRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, errors.New("order ID is required")
	}

	query := `SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, errors.New("failed to check if order exists")
	}

	return exists, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, orderID string, status domain.OrderStatus) error {
	if orderID == "" {
		return errors.New("order ID is required")
	}

	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), orderID)
	if err != nil {
		return errors.New("failed to update order status")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.New("failed to check update result")
	}

	if rowsAffected == 0 {
		return errors.New("order not found")
	}

	return nil
}

func (r *PostgresOrderRepository) GetOrdersByStatus(ctx context.Context, status domain.OrderStatus) ([]domain.Order, error) {
	filter := domain.OrderFilter{
		Status:   status,
		PageSize: 100,
	}

	orders, _, err := r.List(ctx, filter)
	return orders, err
}

func (r *PostgresOrderRepository) GetOrdersByDateRange(ctx context.Context, startDate, endDate time.Time) ([]domain.Order, error) {
	filter := domain.OrderFilter{
		FromDate: &startDate,
		ToDate:   &endDate,
		PageSize: 1000,
	}

	orders, _, err := r.List(ctx, filter)
	return orders, err
}

// Helper methods

func (r *PostgresOrderRepository) getOrderItems(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	itemsQuery := `
		SELECT id, order_id, product_id, name, price, quantity, 
		       COALESCE(frame_size, '') as frame_size,
		       COALESCE(wheel_size, '') as wheel_size,
		       COALESCE(color, '') as color,
		       COALESCE(bike_type, '') as bike_type
		FROM order_items
		WHERE order_id = $1
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, itemsQuery, orderID)
	if err != nil {
		return nil, errors.New("failed to get order items")
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
			&item.FrameSize,
			&item.WheelSize,
			&item.Color,
			&item.BikeType,
		)
		if err != nil {
			return nil, errors.New("failed to scan order item")
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *PostgresOrderRepository) getOrderItemsMap(ctx context.Context, orderIDs []string) (map[string][]domain.OrderItem, error) {
	if len(orderIDs) == 0 {
		return make(map[string][]domain.OrderItem), nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(orderIDs))
	args := make([]interface{}, len(orderIDs))
	for i, id := range orderIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	itemsQuery := fmt.Sprintf(`
		SELECT id, order_id, product_id, name, price, quantity,
		       COALESCE(frame_size, '') as frame_size,
		       COALESCE(wheel_size, '') as wheel_size,
		       COALESCE(color, '') as color,
		       COALESCE(bike_type, '') as bike_type
		FROM order_items
		WHERE order_id IN (%s)
		ORDER BY order_id, id`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, itemsQuery, args...)
	if err != nil {
		return nil, errors.New("failed to get order items")
	}
	defer rows.Close()

	itemsMap := make(map[string][]domain.OrderItem)
	for rows.Next() {
		var item domain.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Name,
			&item.Price,
			&item.Quantity,
			&item.FrameSize,
			&item.WheelSize,
			&item.Color,
			&item.BikeType,
		)
		if err != nil {
			return nil, errors.New("failed to scan order item")
		}

		itemsMap[item.OrderID] = append(itemsMap[item.OrderID], item)
	}

	return itemsMap, nil
}

func (r *PostgresOrderRepository) buildWhereClause(filter domain.OrderFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, filter.UserID)
		argIndex++
	}

	if string(filter.Status) != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.FromDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, filter.FromDate)
		argIndex++
	}

	if filter.ToDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, filter.ToDate)
		argIndex++
	}

	return strings.Join(conditions, " AND "), args
}

func (r *PostgresOrderRepository) validateOrder(order domain.Order) error {
	if order.UserID == "" {
		return errors.New("user ID is required")
	}

	if len(order.Items) == 0 {
		return errors.New("order must contain at least one item")
	}

	if order.Total < 0 {
		return errors.New("order total cannot be negative")
	}

	// Validate each item
	for i, item := range order.Items {
		if item.ProductID == "" {
			return fmt.Errorf("product ID is required for item %d", i+1)
		}
		if item.Name == "" {
			return fmt.Errorf("product name is required for item %d", i+1)
		}
		if item.Price < 0 {
			return fmt.Errorf("product price cannot be negative for item %d", i+1)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("product quantity must be positive for item %d", i+1)
		}
	}

	return nil
}

func (r *PostgresOrderRepository) calculateTotal(items []domain.OrderItem) float64 {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
