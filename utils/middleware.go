package utils

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware validates the JWT token and extracts user_id
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authoriztion header missing"})
			c.Abort()
			return
		}

		// check heaeder format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// check token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtKey, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// extract user_id from token 
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		c.Set("user_id", uint(userID))

		c.Next()
	}
}

func RefreshToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")[7:] // remove "Bearer "

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid or expired token"})
		return
	}

	// define if it is a desktop
	isDesktopClient := c.Request.Header.Get("X-Client-Type") == "desktop"

	// Set expires life of token
	expiration := time.Hour * 24 // 24 hours for web
	if isDesktopClient {
		expiration *= 30 // 30 days for desktop
	}

	// create new token
	claims["exp"] = time.Now().Add(expiration).Unix()
	newToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtKey))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create new token"})
		return
	}

	c.JSON(200, gin.H{"token": newToken})
}

func ValidateToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")[7:] // remove "Bearer "

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid or expired token"})
		return
	}
}

func PingPong(c *gin.Context) {  // todo move it to some other file
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}