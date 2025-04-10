package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Repo interface {
	Insert(ctx context.Context, order Order) error
	FindById(ctx context.Context, id uint64) (Order, error)
	DeleteById(ctx context.Context, id uint64) error
	Update(ctx context.Context, order Order) error
	FindAll(ctx context.Context, page FindAllPage) (FindResult, error)
}

type OrderRepo struct {
	Repo Repo
}

type MsgResponse struct {
	Message string `json:"message"`
}

var ErrNotExist = errors.New("Order does not exist")

func (o *OrderRepo) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID  `json:"customer_id"`
		LineItems  []LineItem `json:"line_items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	order := Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	err := o.Repo.Insert(r.Context(), order)
	if err != nil {
		fmt.Println("Failed to insert order: %w", order)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(order)
	if err != nil {
		fmt.Println("Failed to insert order: %w", order)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
	w.WriteHeader(http.StatusCreated)
}

func (o *OrderRepo) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	const decimal = 10
	const bitSize = 64

	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const size = 50
	page := FindAllPage{Offset: cursor, Size: size}
	res, err := o.Repo.FindAll(r.Context(), page)
	if err != nil {
		fmt.Println("Failed to find all orders: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []Order `json:"items"`
		Next  uint64  `json:"next,omitempty"`
	}

	response.Items = res.Orders
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Failed to marshal orders from database: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (o *OrderRepo) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const decimal = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbOrder, err := o.Repo.FindById(r.Context(), orderID)
	if errors.Is(err, ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(dbOrder); err != nil {
		fmt.Println("Failed to marshal order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *OrderRepo) UpdateById(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err_msg := MsgResponse{Message: "Refer to API docs for endpoint requirements."}
		json.NewEncoder(w).Encode(err_msg)
		return
	}

	idParam := chi.URLParam(r, "id")

	const decimal = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err_msg := MsgResponse{Message: "OrderID must be included and be an integer."}
		json.NewEncoder(w).Encode(err_msg)
		return
	}

	dbOrder, err := o.Repo.FindById(r.Context(), orderID)
	if errors.Is(err, ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find order: %w", err)
		err_msg := MsgResponse{Message: "Please, try again later!"}
		json.NewEncoder(w).Encode(err_msg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	const completedStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now().UTC()

	switch body.Status {
	case shippedStatus:
		if dbOrder.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			err_msg := MsgResponse{Message: "This order has already been shipped!"}
			json.NewEncoder(w).Encode(err_msg)
			return
		}
		dbOrder.ShippedAt = &now
	case completedStatus:
		if dbOrder.ShippedAt == nil {
			w.WriteHeader(http.StatusBadRequest)
			err_msg := MsgResponse{Message: "This order hasn't been shipped!"}
			json.NewEncoder(w).Encode(err_msg)
			return
		} else if dbOrder.CompletedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			err_msg := MsgResponse{Message: "This order has already been completed!"}
			json.NewEncoder(w).Encode(err_msg)
			return
		}
		dbOrder.CompletedAt = &now
	default:
		w.WriteHeader(http.StatusBadRequest)
		err_msg := MsgResponse{Message: "Allowed status: ['shipped', 'completed']!"}
		json.NewEncoder(w).Encode(err_msg)
		return
	}

	err = o.Repo.Update(r.Context(), dbOrder)
	if err != nil {
		fmt.Println("Failed to update:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(dbOrder); err != nil {
		fmt.Println("Failed to marshal order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *OrderRepo) DeleteById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const decimal = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = o.Repo.DeleteById(r.Context(), orderID)
	if errors.Is(err, ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
