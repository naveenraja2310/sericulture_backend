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

func getFirmwareDir() (string, error) {
	configuredDir := filepath.Join(os.TempDir(), "sericulture", "firmware")

	absDir, err := filepath.Abs(configuredDir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", err
	}
	return absDir, nil
}

func UploadFile(c *fiber.Ctx) error {
	//creating a context
	file, err := c.FormFile("document")
	if err != nil {
		return err
	}
	// Sanitize filename to prevent path traversal attacks:
	filename := filepath.Base(file.Filename)
	uploadDir, err := getFirmwareDir()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.SuccessResponse{
			StatusCode:    http.StatusInternalServerError,
			StatusMessage: "error",
			Data:          "unable to access firmware storage: " + err.Error(),
		})
	}
	savePath := filepath.Join(uploadDir, filename)
	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.SuccessResponse{
			StatusCode:    http.StatusInternalServerError,
			StatusMessage: "error",
			Data:          "failed to save firmware file: " + err.Error(),
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

	uploadDir, err := getFirmwareDir()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(model.SuccessResponse{
			StatusCode:    http.StatusInternalServerError,
			StatusMessage: "error",
			Data:          "unable to access firmware storage: " + err.Error(),
		})
	}

	filePath := filepath.Join(uploadDir, filename)
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
