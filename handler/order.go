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

	"victorcalderon/go-microservice/model"
	"victorcalderon/go-microservice/repository/order"
)

type Order struct {
	Repo *order.RedisRepo
}

type MsgResponse struct {
	Message string `json:"message"`
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

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
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
	page := order.FindAllPage{Offset: cursor, Size: size}
	res, err := o.Repo.FindAll(r.Context(), page)
	if err != nil {
		fmt.Println("Failed to find all orders: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
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

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const decimal = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbOrder, err := o.Repo.FindById(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
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

func (o *Order) UpdateById(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprint("Refer to API docs for endpoint requirements.")))
		return
	}

	idParam := chi.URLParam(r, "id")

	const decimal = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprint("OrderID must be included and be an integer.")))
		return
	}

	dbOrder, err := o.Repo.FindById(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find order: %w", err)
		err_msg := MsgResponse{Message: fmt.Sprint("Please, try again later!")}
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
			err_msg := MsgResponse{Message: fmt.Sprint("This order has already been shipped!")}
			json.NewEncoder(w).Encode(err_msg)
			return
		}
		dbOrder.ShippedAt = &now
	case completedStatus:
		if dbOrder.ShippedAt == nil {
			w.WriteHeader(http.StatusBadRequest)
			err_msg := MsgResponse{Message: fmt.Sprint("This order hasn't been shipped!")}
			json.NewEncoder(w).Encode(err_msg)
			return
		} else if dbOrder.CompletedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			err_msg := MsgResponse{Message: fmt.Sprint("This order has already been completed!")}
			json.NewEncoder(w).Encode(err_msg)
			return
		}
		dbOrder.CompletedAt = &now
	default:
		w.WriteHeader(http.StatusBadRequest)
		err_msg := MsgResponse{Message: fmt.Sprint("Allowed status: ['shipped', 'completed']!")}
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

func (o *Order) DeleteById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const decimal = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = o.Repo.DeleteById(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
