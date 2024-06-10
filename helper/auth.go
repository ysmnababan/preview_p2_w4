package helper

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// do auth
		tokenString := c.Request().Header.Get("Auth")
		response := map[string]interface{}{}
		if tokenString == "" {
			Logging(c).Warning("unable to get the token")
			response["message"] = "unautorized"
			return c.JSON(http.StatusUnauthorized, response)
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte("secret"), nil
		})

		if err != nil || !token.Valid {
			Logging(c).Warning("token invalid")
			response["message"] = "unauthorized"
			return c.JSON(http.StatusUnauthorized, response)
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Get the user record from database or
			// run through your business logic to verify if the user can log in
			c.Set("email", claims["email"].(string))
			c.Set("user_id", claims["user_id"].(float64))
		} else {
			Logging(c).Warning("claims invalid")
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}

		return next(c)
	}
}
