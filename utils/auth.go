package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"sericulture/database"
	"sericulture/model"
	"sericulture/settings"

	"github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JWTClaims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	DeviceID string `json:"deviceId"`
	jwt.RegisteredClaims
}

type userCacheEntry struct {
	User      model.User
	ExpiresAt time.Time
}

type userCache struct {
	mu    sync.RWMutex
	items map[string]userCacheEntry
}

var authUserCache = &userCache{items: map[string]userCacheEntry{}}

func getJWTSecret(password string) string {
	secret := strings.TrimSpace(settings.Config.JWTSecret)
	if secret == "" {
		secret = "sericulture-default-secret"
	}
	if password == "" {
		return secret
	}
	return secret + password
}

func GenerateJWT(user model.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID.Hex(),
		Username: user.Username,
		DeviceID: user.DeviceID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(getJWTSecret(user.Password)))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ParseJWT(tokenString string, password string) (*JWTClaims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, errors.New("missing token")
	}

	claims := &JWTClaims{}
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(getJWTSecret(password)), nil
	})
	if err != nil || !parsedToken.Valid {
		return nil, err
	}

	return claims, nil
}

func ValidateJWT(tokenString string, user model.User) error {
	claims, err := ParseJWT(tokenString, user.Password)
	if err != nil {
		return err
	}

	if claims.UserID != user.ID.Hex() || claims.Username != user.Username {
		return errors.New("unauthorized: password changed or token invalid")
	}

	return nil
}

func (c *userCache) Get(userID string) (model.User, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.items[userID]
	if !ok {
		return model.User{}, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(c.items, userID)
		return model.User{}, false
	}

	return entry.User, true
}

func (c *userCache) Set(userID string, user model.User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[userID] = userCacheEntry{User: user, ExpiresAt: time.Now().Add(time.Hour)}
}

func (c *userCache) Delete(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, userID)
}

func ClearUserCache(userID string) {
	if userID == "" {
		return
	}
	authUserCache.Delete(userID)
}

func JWTMiddleware(c *fiber.Ctx) error {
	if c.Method() == fiber.MethodOptions || c.Path() == "/" || c.Path() == "/login" || strings.HasSuffix(c.Path(), "/ws") {
		return c.Next()
	}

	authHeader := c.Get(fiber.HeaderAuthorization)
	if authHeader == "" {
		return unauthorizedResponse(c)
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return unauthorizedResponse(c)
	}

	tokenString := strings.TrimSpace(parts[1])
	claims := &JWTClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(tokenString, claims); err != nil {
		return unauthorizedResponse(c)
	}

	if claims.UserID == "" {
		return unauthorizedResponse(c)
	}

	objectID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return unauthorizedResponse(c)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	user, found := authUserCache.Get(claims.UserID)
	if !found {
		Warn(c, "User cache miss; fetching from DB", "userId", claims.UserID)
		var freshUser model.User
		if err := database.Users.FindOne(ctx, bson.M{"_id": objectID}).Decode(&freshUser); err != nil {
			return unauthorizedResponse(c)
		}
		user = freshUser
		authUserCache.Set(claims.UserID, user)
	}

	if err := ValidateJWT(tokenString, user); err != nil {
		Warn(c, "JWT validation failed", "userId", claims.UserID, "deviceId", user.DeviceID, "error", err.Error())
		return unauthorizedResponse(c)
	}

	c.Locals("user", user)
	c.Locals("userId", user.ID.Hex())
	c.Locals("deviceId", user.DeviceID)
	c.Locals("deviceID", user.DeviceID)
	return c.Next()
}

func unauthorizedResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse{
		ApiPath:      c.OriginalURL(),
		ErrorCode:    fiber.StatusUnauthorized,
		ErrorMessage: "Unauthorized: token is missing, invalid or expired from password change",
		ErrorTime:    time.Now(),
	})
}
