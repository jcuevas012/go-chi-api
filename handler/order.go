package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcuevas012/orders-api/model"
	"github.com/jcuevas012/orders-api/repository/order"
)

type Order struct {
	Repo *order.RedisRepo
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	order := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	err := o.Repo.Insert(r.Context(), order)

	if err != nil {
		fmt.Println("failed to insert order %w", order)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	const decimal = 10
	const bitSize = 64

	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)

	if err != nil {
		fmt.Println("failed to parse cursor")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const size = 10

	res, err := o.Repo.FindAll(r.Context(), order.FindAllPage{
		Offset: cursor,
		Size:   size,
	})

	if err != nil {
		fmt.Println("failed to find all %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []model.Order `json:"line_items"`
		Next  uint64        `json:"next,omitempty"`
	}

	response.Items = res.Orders
	response.Next = response.Next

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshal %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

}

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	base := 10
	bitSize := 64

	orderId, err := strconv.ParseUint(idParam, base, bitSize)

	if err != nil {
		fmt.Println("failed to parse", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	res, err := o.Repo.FindByID(r.Context(), orderId)

	if errors.Is(err, order.ErrorNotExists) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("fail to find order by id", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(res); err != nil {
		fmt.Println("fail to encode json", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")

	base := 10
	bitSize := 64

	orderID, _ := strconv.ParseUint(idParam, base, bitSize)

	foundOrder, err := o.Repo.FindByID(r.Context(), orderID)

	if errors.Is(err, order.ErrorNotExists) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("fail to find order by id", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	const CompletedStatus = "completed"
	const ShippedStatus = "shipped"

	now := time.Now().UTC()

	switch body.Status {
	case ShippedStatus:
		if foundOrder.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		foundOrder.UpdatedAt = &now
		foundOrder.ShippedAt = &now
	case CompletedStatus:
		if foundOrder.CompletedAt != nil && foundOrder.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		foundOrder.UpdatedAt = &now
		foundOrder.CompletedAt = &now
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = o.Repo.Update(r.Context(), foundOrder)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(foundOrder); err != nil {
		fmt.Println("failed to encode found order")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) DeleteById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("delete order by ID")
}
