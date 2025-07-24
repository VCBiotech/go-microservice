package utils

// import (
// 	"strings"

// 	"github.com/clerk/clerk-sdk-go/v2/jwt"
// 	"github.com/clerk/clerk-sdk-go/v2/user"
// 	"github.com/labstack/echo/v4"

// 	"vcbiotech/microservice/telemetry"
// )

// type ClerkConfig struct {
// 	SecretKey string
// }

// func ClerkAuth(c echo.Context) error {
// 	logger := telemetry.SLogger(c.Request().Context())

// 	// Get the session JWT from the Authorization header
// 	sessionToken := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

// 	// Verify the session
// 	claims, err := jwt.Verify(c.Request().Context(), &jwt.VerifyParams{
// 		Token: sessionToken,
// 	})
// 	if err != nil {
// 		logger.Error("Failed verify session", map[string]string{"Error": err.Error()})
// 		return err
// 	}

// 	usr, err := user.Get(c.Request().Context(), claims.Subject)
// 	if err != nil {
// 		logger.Error("Failed to get user", map[string]string{"Error": err.Error()})
// 		return err
// 	}

// 	logger.Info("User authenticated", map[string]string{"user_id": usr.ID})
// 	return nil
// }
