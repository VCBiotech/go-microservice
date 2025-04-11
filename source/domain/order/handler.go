package order

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo"
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

func (o *OrderRepo) Create(c echo.Context) {
	var body struct {
		CustomerID uuid.UUID  `json:"customer_id"`
		LineItems  []LineItem `json:"line_items"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
		return
	}

	now := time.Now().UTC()
	order := Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	err := o.Repo.Insert(c.Request().Context(), order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
		return
	}

	res, err := json.Marshal(order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (o *OrderRepo) List(c echo.Context) {
	cursorStr := c.QueryParam("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	const decimal = 10
	const bitSize = 64

	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
		return
	}

	const size = 50
	page := FindAllPage{Offset: cursor, Size: size}
	res, err := o.Repo.FindAll(c.Request().Context(), page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (o *OrderRepo) GetByID(c echo.Context) {
	idParam := c.QueryParam("id")

	const decimal = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
		return
	}

	dbOrder, err := o.Repo.FindById(c.Request().Context(), orderID)
	if errors.Is(err, ErrNotExist) {
		c.JSON(http.StatusNotFound, map[string]string{"Error": err.Error()})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dbOrder)
}

func (o *OrderRepo) UpdateById(c echo.Context) {
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
		return
	}

	idParam := c.QueryParam("id")

	const decimal = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
		return
	}

	dbOrder, err := o.Repo.FindById(c.Request().Context(), orderID)
	if errors.Is(err, ErrNotExist) {
		c.JSON(http.StatusNotFound, map[string]string{"Error": err.Error()})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
		return
	}

	const completedStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now().UTC()

	switch body.Status {
	case shippedStatus:
		if dbOrder.ShippedAt != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"Error": "This order has already been shipped!"})
			return
		}
		dbOrder.ShippedAt = &now
	case completedStatus:
		if dbOrder.ShippedAt == nil {
			c.JSON(http.StatusBadRequest, map[string]string{"Error": "This order hasn't been shipped!"})
			return
		} else if dbOrder.CompletedAt != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"Error": "This order has already been completed!"})
			return
		}
		dbOrder.CompletedAt = &now
	default:
		c.JSON(http.StatusBadRequest, map[string]string{"Error": "Allowed status: ['shipped', 'completed']!"})
		return
	}

	err = o.Repo.Update(c.Request().Context(), dbOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dbOrder)
}

func (o *OrderRepo) DeleteById(c echo.Context) {
	idParam := c.QueryParam("id")

	const decimal = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
		return
	}

	err = o.Repo.DeleteById(c.Request().Context(), orderID)
	if errors.Is(err, ErrNotExist) {
		c.JSON(http.StatusNotFound, map[string]string{"Error": err.Error()})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
