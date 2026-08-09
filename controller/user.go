package controller

import (
	"context"
	"net/http"
	"sericulture/database"
	"sericulture/model"
	"sericulture/service"
	"sericulture/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Login(c *fiber.Ctx) error {
	//creating a context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.ContextTime)*time.Second)
	defer cancel()

	var adminlogin struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&adminlogin); err != nil {
		return c.Status(http.StatusBadRequest).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Failed to parse request body",
			ErrorTime:    time.Now(),
		})
	}

	//fetch data from DB
	user, err := service.Login(ctx, adminlogin.Username, adminlogin.Password)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
			ErrorTime:    time.Now(),
		})
	}

	token, err := utils.GenerateJWT(user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
			ErrorTime:    time.Now(),
		})
	}

	// Return a success model
	return c.Status(http.StatusOK).JSON(model.SuccessResponse{
		StatusCode:    http.StatusOK,
		StatusMessage: "success",
		Data: fiber.Map{
			"user":  user,
			"token": token,
		},
	})
}

func LogOut(c *fiber.Ctx) error {
	//creating a context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.ContextTime)*time.Second)
	defer cancel()

	deviceId := c.Params("id")

	//fetch data from DB
	err := service.LogOut(ctx, deviceId)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
			ErrorTime:    time.Now(),
		})
	}

	// Return a success model
	return c.Status(http.StatusOK).JSON(model.SuccessResponse{
		StatusCode:    http.StatusOK,
		StatusMessage: "success",
		Data:          nil,
	})
}

func CreateUser(c *fiber.Ctx) error {
	//creating a context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.ContextTime)*time.Second)
	defer cancel()

	//parsing a request body
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Failed to parse request body",
			ErrorTime:    time.Now(),
		})
	}

	//saving data in db
	result, err := service.CreateUser(ctx, user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
			ErrorTime:    time.Now(),
		})
	}

	// Return a success model with the created objectid
	return c.Status(http.StatusCreated).JSON(model.SuccessResponse{
		StatusCode:    http.StatusCreated,
		StatusMessage: "success",
		Data:          result,
	})
}

func UpdateUser(c *fiber.Ctx) error {
	//creating a context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.ContextTime)*time.Second)
	defer cancel()

	idParam := utils.StringToObjectID(c.Params("id"))

	//parsing a request body
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "Failed to parse request body",
			ErrorTime:    time.Now(),
		})
	}

	//saving data in db
	result, err := service.UpdateUser(ctx, user, idParam)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
			ErrorTime:    time.Now(),
		})
	}

	// Return a success model with the created objectid
	return c.Status(http.StatusOK).JSON(model.SuccessResponse{
		StatusCode:    http.StatusOK,
		StatusMessage: "success",
		Data:          result,
	})
}

func GetUserById(c *fiber.Ctx) error {
	//creating a context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.ContextTime)*time.Second)
	defer cancel()

	idParam := utils.StringToObjectID(c.Params("id"))

	//fetch data from DB
	result, err := service.GetUserByID(ctx, idParam)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
			ErrorTime:    time.Now(),
		})
	}

	// Return a success model
	return c.Status(http.StatusOK).JSON(model.SuccessResponse{
		StatusCode:    http.StatusOK,
		StatusMessage: "success",
		Data:          result,
	})
}

func DeleteUser(c *fiber.Ctx) error {
	//creating a context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.ContextTime)*time.Second)
	defer cancel()

	idParam := utils.StringToObjectID(c.Params("id"))

	//fetch data from DB
	err := service.DeleteUser(ctx, idParam)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
			ErrorTime:    time.Now(),
		})
	}

	// Return a success model
	return c.Status(http.StatusOK).JSON(model.SuccessResponse{
		StatusCode:    http.StatusOK,
		StatusMessage: "success",
		Data:          nil,
	})
}

func GetAllUsers(c *fiber.Ctx) error {
	//creating a context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(database.ContextTime)*time.Second)
	defer cancel()

	limit, limiterr := strconv.Atoi(c.Query("limit"))
	if limiterr != nil {
		return c.Status(http.StatusBadRequest).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: limiterr.Error(),
			ErrorTime:    time.Now(),
		})
	}

	pagenumber, pagenumbererr := strconv.Atoi(c.Query("page"))
	if pagenumbererr != nil {
		return c.Status(http.StatusBadRequest).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: pagenumbererr.Error(),
			ErrorTime:    time.Now(),
		})
	}

	search := c.Query("search")

	offset := (pagenumber - 1) * limit

	//fetch data from DB
	result, count, err := service.GetAllUsers(ctx, int64(limit), int64(offset), search)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
			ErrorTime:    time.Now(),
		})
	}

	// Return a success model
	return c.Status(http.StatusOK).JSON(model.SuccessResponse{
		StatusCode:    http.StatusOK,
		StatusMessage: "success",
		Data:          &fiber.Map{"users": result, "total_count": count},
	})
}
