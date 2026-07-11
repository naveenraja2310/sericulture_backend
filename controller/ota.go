package controller

import (
	"net/http"
	"os"
	"path/filepath"
	"sericulture/model"
	"sericulture/mqtt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func UploadFile(c *fiber.Ctx) error {
	//creating a context
	file, err := c.FormFile("document")
	if err != nil {
		return err
	}
	// Sanitize filename to prevent path traversal attacks:
	filename := filepath.Base(file.Filename)
	uploadDir := filepath.Join("ota", "firmware")
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return err
	}
	savePath := filepath.Join(uploadDir, filename)
	if err := c.SaveFile(file, savePath); err != nil {
		// Return a success model with the created objectid
		return c.Status(http.StatusInternalServerError).JSON(model.SuccessResponse{
			StatusCode:    http.StatusInternalServerError,
			StatusMessage: "error",
			Data:          err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(model.SuccessResponse{
		StatusCode:    http.StatusCreated,
		StatusMessage: "success",
		Data:          filename,
	})
}

func GetFile(c *fiber.Ctx) error {
	filename := filepath.Base(c.Params("filename"))
	if filename == "." || filename == "" {
		return c.Status(http.StatusBadRequest).JSON(model.SuccessResponse{
			StatusCode:    http.StatusBadRequest,
			StatusMessage: "error",
			Data:          "invalid filename",
		})
	}

	filePath := filepath.Join("ota", "firmware", filename)
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return c.Status(http.StatusNotFound).JSON(model.SuccessResponse{
				StatusCode:    http.StatusNotFound,
				StatusMessage: "error",
				Data:          "file not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(model.SuccessResponse{
			StatusCode:    http.StatusInternalServerError,
			StatusMessage: "error",
			Data:          err.Error(),
		})
	}

	return c.Download(filePath, filename)
}

type FirmwareUpdateRequest struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	Size    int    `json:"size"`
	SHA256  string `json:"sha256"`
}

func UpdateFirmware(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	if deviceID == "" {
		return c.Status(http.StatusBadRequest).JSON(model.ErrorResponse{
			ApiPath:      c.OriginalURL(),
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "device ID is required",
			ErrorTime:    time.Now(),
		})
	}

	var body FirmwareUpdateRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	payload := fiber.Map{
		"method": "otaUpdate",
		"params": body,
	}

	mqtt.SendCommand(deviceID, payload)

	return c.JSON(fiber.Map{
		"message": "firmware updated for device " + deviceID,
	})
}
