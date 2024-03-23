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

	"github.com/go-chi/chi/v5"
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

func (u *UserRepo) Create(w http.ResponseWriter, r *http.Request) {
	logger := telemetry.SLogger(r.Context())
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"` // Assume this will be hashed before storage
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("Could not decode body", errMsg)
		return
	}

	// Hash the password before storing it
	passwordHash, err := hashPassword(body.Password)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Could not hash password", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user := User{
		Email:        body.Email,
		PasswordHash: passwordHash,
	}

	err = u.Repo.Insert(r.Context(), &user)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to insert new user", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(user)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to marshal new user", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

func (u *UserRepo) List(w http.ResponseWriter, r *http.Request) {
	logger := telemetry.SLogger(r.Context())
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	const decimal = 10
	const bitSize = 64

	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("Could not parse cursor", errMsg)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const size = 50
	page := FindAllPage{Offset: cursor, Size: size}
	res, err := u.Repo.FindAll(r.Context(), page)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to find page of Users", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []User `json:"items"`
		Next  uint64 `json:"next,omitempty"`
	}

	response.Items = res.Users
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to Marshal users from database", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (u *UserRepo) GetByID(w http.ResponseWriter, r *http.Request) {
	logger := telemetry.SLogger(r.Context())
	idParam := chi.URLParam(r, "id")
	const decimal = 10
	const bitSize = 64

	userID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("Could not parse id", errMsg)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbUser, err := u.Repo.Find(r.Context(), userID)
	if errors.Is(err, ErrNotExist) {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("User does not exist", errMsg)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to find user", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dbUser); err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to Marshal user.", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (u *UserRepo) UpdateById(w http.ResponseWriter, r *http.Request) {
	logger := telemetry.SLogger(r.Context())
	var body struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("Failed to decode body", errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Refer to API docs for endpoint requirements."))
		return
	}

	idParam := chi.URLParam(r, "id")
	const decimal = 10
	const bitSize = 64

	userID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("ID was poorly formatted", errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("UserID must be included and be an integer."))
		return
	}

	dbUser, err := u.Repo.Find(r.Context(), userID)
	if errors.Is(err, ErrNotExist) {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("User could not be found", errMsg)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to find user", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update the user's email
	if body.Email != "" {
		dbUser.Email = body.Email
		dbUser.UpdatedAt = time.Now().UTC()
	}

	err = u.Repo.Update(r.Context(), dbUser)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to update user", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dbUser); err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to marshal user", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (u *UserRepo) DeleteById(w http.ResponseWriter, r *http.Request) {
	logger := telemetry.SLogger(r.Context())
	idParam := chi.URLParam(r, "id")

	const decimal = 10
	const bitSize = 64

	userID, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("Failed to parse id", errMsg)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = u.Repo.Delete(r.Context(), userID)
	if errors.Is(err, ErrNotExist) {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Info("User does not exist", errMsg)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Failed to delete user", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
