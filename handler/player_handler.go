package handler

import (
	"fmt"
	"log"
	"net/http"
	"pagi/helper"
	"pagi/models"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Repo struct {
	DB *gorm.DB
}

func (r *Repo) GetPlayer(c echo.Context) error {
	var players []models.Player
	fmt.Printf("%v %d", c.Get("email"), uint(c.Get("user_id").(float64)))
	res := r.DB.Find(&players)
	if res.Error != nil {
		helper.Logging(c).Error("err query", res.Error)
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}
	return c.JSON(http.StatusOK, players)
}

func (r *Repo) CreatePlayer(c echo.Context) error {
	var getP models.Player
	err := c.Bind(&getP)
	if err != nil {
		helper.Logging(c).Error("error binding", err)
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	// validate user
	if getP.TeamName == "" || getP.Ranking <= 0 || getP.Username == "" || getP.Score <= 0 {
		helper.Logging(c).Error("error or missing param", getP)
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	if err := r.DB.Create(&getP).Error; err != nil {
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
	return c.JSON(http.StatusCreated, getP)
}

func (r *Repo) UpdatePlayer(c echo.Context) error {
	log.Println(c.Param("id"))
	id, _ := strconv.Atoi(c.Param("id"))

	var getP models.Player
	err := c.Bind(&getP)
	if err != nil {
		helper.Logging(c).Error("error binding", err)
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}

	// validate user
	if getP.TeamName == "" || getP.Ranking <= 0 || getP.Username == "" || getP.Score <= 0 {
		helper.Logging(c).Error("error or missing param", getP)
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	var UpdateP models.Player
	res := r.DB.First(&UpdateP, id)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusBadRequest, "No player found")
		}
		helper.Logging(c).Error("err query", res.Error)
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}
	UpdateP.Ranking = getP.Ranking
	UpdateP.Score = getP.Score
	UpdateP.TeamName = getP.TeamName
	UpdateP.Username = getP.Username
	// UpdateP.PlayerID = uint(id)

	res = r.DB.Save(&UpdateP)
	if res.Error != nil {
		helper.Logging(c).Error("err query", res.Error)
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}
	return c.JSON(http.StatusCreated, UpdateP)
}
