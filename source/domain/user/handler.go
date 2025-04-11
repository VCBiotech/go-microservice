package user

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/argon2"

	"vcbiotech/microservice/telemetry"
)

type Repo interface {
	Insert(ctx context.Context, user *User) error
	Find(ctx context.Context, id uint64) (*User, error)
	Delete(ctx context.Context, id uint64) error
	Update(ctx context.Context, user *User) error
	FindAll(ctx context.Context, page FindAllPage) (FindResult, error)
}

type UserRepo struct {
	Repo Repo
}

type MsgResponse struct {
	Message string `json:"message"`
}

var ErrNotExist = errors.New("User does not exist")

func (u *UserRepo) Create(c echo.Context) error {
	logger := telemetry.SLogger(c.Request().Context())
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"` // Assume this will be hashed before storage
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("Could not decode body", errMsg)
		return c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
	}

	// Hash the password before storing it
	passwordHash, err := hashPassword(body.Password)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Could not hash password", errMsg)
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	user := User{
		Email:        body.Email,
		PasswordHash: passwordHash,
	}

	err = u.Repo.Insert(c.Request().Context(), &user)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to insert new user", errMsg)
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	res, err := json.Marshal(user)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to marshal new user", errMsg)
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	return c.JSON(http.StatusCreated, res)
}

func (u *UserRepo) List(c echo.Context) error {
	logger := telemetry.SLogger(c.Request().Context())
	cursorStr := c.QueryParam("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	const decimal = 10
	const bitSize = 64

	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("Could not parse cursor", errMsg)
		return c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
	}

	const size = 50
	page := FindAllPage{Offset: cursor, Size: size}
	res, err := u.Repo.FindAll(c.Request().Context(), page)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to find page of Users", errMsg)
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	var response struct {
		Items []User `json:"items"`
		Next  uint64 `json:"next,omitempty"`
	}

	response.Items = res.Users
	response.Next = res.Cursor

	return c.JSON(http.StatusOK, response)
}

func (u *UserRepo) GetByID(c echo.Context) error {
	logger := telemetry.SLogger(c.Request().Context())
	idParam := c.QueryParam("id")
	const decimal = 10
	const bitSize = 64

	userID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("Could not parse id", errMsg)
		return c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
	}

	dbUser, err := u.Repo.Find(c.Request().Context(), userID)
	if errors.Is(err, ErrNotExist) {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("User does not exist", errMsg)
		return c.JSON(http.StatusNotFound, map[string]string{"Error": err.Error()})
	} else if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to find user", errMsg)
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	if err := json.NewEncoder(c.Response().Writer).Encode(dbUser); err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to Marshal user.", errMsg)
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	return nil
}

func (u *UserRepo) UpdateById(c echo.Context) error {
	logger := telemetry.SLogger(c.Request().Context())
	var body struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("Failed to decode body", errMsg)
		return c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
	}

	idParam := c.QueryParam("id")
	const decimal = 10
	const bitSize = 64

	userID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("ID was poorly formatted", errMsg)
		return c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
	}

	dbUser, err := u.Repo.Find(c.Request().Context(), userID)
	if errors.Is(err, ErrNotExist) {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("User could not be found", errMsg)
		return c.JSON(http.StatusNotFound, map[string]string{"Error": err.Error()})
	} else if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to find user", errMsg)
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	// Update the user's email
	if body.Email != "" {
		dbUser.Email = body.Email
		dbUser.UpdatedAt = time.Now().UTC()
	}

	err = u.Repo.Update(c.Request().Context(), dbUser)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to update user", errMsg)
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	if err := json.NewEncoder(c.Response().Writer).Encode(dbUser); err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to marshal user", errMsg)
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	return nil
}

func (u *UserRepo) DeleteById(c echo.Context) error {
	logger := telemetry.SLogger(c.Request().Context())
	idParam := c.QueryParam("id")

	const decimal = 10
	const bitSize = 64

	userID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("Failed to parse id", errMsg)
		return c.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
	}

	err = u.Repo.Delete(c.Request().Context(), userID)
	if errors.Is(err, ErrNotExist) {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("User does not exist", errMsg)
		return c.JSON(http.StatusNotFound, map[string]string{"Error": err.Error()})
	} else if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to delete user", errMsg)
		return c.JSON(http.StatusInternalServerError, map[string]string{"Error": err.Error()})
	}

	return c.JSON(http.StatusNoContent, nil)
}

// hashPassword would be a function to hash the password before storing it in your database
func hashPassword(password string) (string, error) {
	// Generate a Salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Hash the password using Argon2
	// Parameters: memory = 64*1024, iterations = 1, parallelism = 4, key length = 32
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Encode the salt and hash to base64 for storing in PHC format
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	// Return the PHC formatted string: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
	phcFormat := fmt.Sprintf("$argon2id$v=19$m=65536,t=1,p=4$%s$%s", saltEncoded, hashEncoded)

	return phcFormat, nil
}
