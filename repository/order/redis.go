package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jcuevas012/orders-api/model"
	"github.com/redis/go-redis/v9"
)

func orderIdKey(ID uint64) string {
	return fmt.Sprintf("order:%d", ID)
}

var ErrorNotExists = errors.New("order does not exist")

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

type RedisRepo struct {
	Client *redis.Client
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("fail to encode order: %w", err)
	}

	key := orderIdKey(order.OrderID)

	tx := r.Client.TxPipeline()

	res := tx.SetNX(ctx, key, string(data), 0)

	if err := res.Err(); err != nil {
		tx.Discard()
		return fmt.Errorf("fail to set: %w", err)
	}

	if err := tx.SAdd(ctx, "orders", key).Err(); err != nil {
		tx.Discard()
		return fmt.Errorf("fail to add orders set %w", err)
	}

	if _, err := tx.Exec(ctx); err != nil {
		return fmt.Errorf("fail to execute %w", err)
	}

	return nil
}

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIdKey(id)

	value, err := r.Client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrorNotExists
	} else if err != nil {
		return model.Order{}, fmt.Errorf("get order %w", err)
	}

	var order model.Order

	err = json.Unmarshal([]byte(value), &order)

	if err != nil {
		return model.Order{}, fmt.Errorf("failed to decode order json: %w", err)
	}

	return order, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIdKey(id)

	tx := r.Client.TxPipeline()

	err := tx.Del(ctx, key).Err()

	if errors.Is(err, redis.Nil) {
		tx.Discard()
		return ErrorNotExists
	} else if err != nil {
		tx.Discard()
		return fmt.Errorf("get order %w", err)
	}

	if err := tx.SRem(ctx, "orders", key).Err(); err != nil {
		tx.Discard()
		return fmt.Errorf("delete order set %w", err)
	}

	if _, err := tx.Exec(ctx); err != nil {
		return fmt.Errorf("fail to exec %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {

	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order %w", err)
	}

	key := orderIdKey(order.OrderID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()

	if errors.Is(err, redis.Nil) {
		return ErrorNotExists
	} else if err != nil {
		return fmt.Errorf("update order %w", err)
	}

	return nil
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	keys, cursors, err := res.Result()

	if err != nil {
		return FindResult{}, fmt.Errorf("faild to find orders %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()

	if err != nil {
		return FindResult{}, fmt.Errorf("error fetching orders keys: %w", err)
	}

	orders := make([]model.Order, len(xs))

	for i, x := range xs {
		x := x.(string)
		var order model.Order

		err := json.Unmarshal([]byte(x), &order)

		if err != nil {
			return FindResult{}, fmt.Errorf("faild to decode order %w", err)
		}

		orders[i] = order
	}

	return FindResult{
		Orders: orders,
		Cursor: cursors,
	}, nil
}
