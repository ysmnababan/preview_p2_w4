package handler

import (
	"fmt"
	"net/http"
	"pagi/helper"
	"pagi/models"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func generateTokenPair(email string, user_id uint) (map[string]string, error) {
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	// This is the information which frontend can use
	// The backend can also decode the token and get admin etc.
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = 1
	claims["email"] = email
	claims["user_id"] = user_id
	claims["exp"] = time.Now().Add(time.Minute * 5).Unix()

	// Generate encoded token and send it as response.
	// The signing string should be secret (a generated UUID works too)
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["sub"] = 1
	rtClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	rt, err := refreshToken.SignedString([]byte("secret"))
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"access_token":  t,
		"refresh_token": rt,
	}, nil
}
func (r *Repo) Login(c echo.Context) error {
	var u models.User
	err := c.Bind(&u)
	if err != nil {
		helper.Logging(c).Error("error binding", err)
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	// validate user
	if u.Email == "" || u.Password == "" {
		helper.Logging(c).Error("error or missing param", u)
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	var login models.User
	res := r.DB.Where("email = ?", u.Email).First(&login)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusBadRequest, "No user found")
		}
		helper.Logging(c).Error("err query", res.Error)
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	if login.Password != u.Password {
		return c.JSON(http.StatusBadRequest, "password does not match")
	}

	token, err := generateTokenPair(login.Email, login.UserID)
	if err != nil {
		helper.Logging(c).Error("err generating token", err)
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	return c.JSON(http.StatusOK, token)
}

func (r *Repo) Register(c echo.Context) error {
	var u models.User
	err := c.Bind(&u)
	if err != nil {
		helper.Logging(c).Error("error binding", err)
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	// validate user
	if u.Email == "" || u.Password == "" || u.Username == "" {
		helper.Logging(c).Error("error or missing param", u)
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	if err := r.DB.Create(&u).Error; err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			// Handle unique constraint violation
			helper.Logging(c).Error("duplicate entry", pgErr.Detail)
			return c.JSON(http.StatusBadRequest, "User already exists")
			// fmt.Println("Unique constraint violated:", pgErr.Detail)
		} else {
			// Handle other errors
			helper.Logging(c).Error("err query", err)
			return c.JSON(http.StatusInternalServerError, "Internal Server Error")
		}
	}
	return c.JSON(http.StatusCreated, u)
}

func (r *Repo) RefreshToken(c echo.Context) error {
	// This is the api to refresh tokens
	// Most of the code is taken from the jwt-go package's sample codes
	// https://godoc.org/github.com/dgrijalva/jwt-go#example-Parse--Hmac

	type tokenReqBody struct {
		RefreshToken string `json:"refresh_token"`
		Email        string `json:"email"`
		UserID       uint   `json:"user_id"`
	}
	tokenReq := tokenReqBody{}
	c.Bind(&tokenReq)

	fmt.Println("TOKEN", tokenReq)
	// Parse takes the token string and a function for looking up the key.
	// The latter is especially useful if you use multiple keys for your application.
	// The standard is to use 'kid' in the head of the token to identify
	// which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	fmt.Println("here")
	token, err := jwt.Parse(tokenReq.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte("secret"), nil
	})

	if err != nil || !token.Valid {
		helper.Logging(c).Warning("token invalid", err)
		// response["message"] = "unauthorized"
		return c.JSON(http.StatusUnauthorized, "unauthorized")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println("here2")
		// Get the user record from database or
		// run through your business logic to verify if the user can log in
		// fmt.Println("XXXXX", claims["email"].(string), uint(claims["user_id"].(float64)))
		if int(claims["sub"].(float64)) == 1 {
			newTokenPair, err := generateTokenPair(tokenReq.Email, tokenReq.UserID)
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, newTokenPair)
		}

		return echo.ErrUnauthorized
	}
	fmt.Println("here3")

	return err

}
