package controller

import (
	"context"
	"net/http"
	"sericulture/database"
	"sericulture/model"
	"sericulture/service"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetAllNotification(c *fiber.Ctx) error {
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

	deviceID := c.Query("device_id")

	offset := (pagenumber - 1) * limit

	//fetch data from DB
	result, count, err := service.GetAllNotification(ctx, int64(limit), int64(offset), deviceID)
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
		Data:          &fiber.Map{"notification": result, "total_count": count},
	})
}
