package http

import "github.com/gofiber/fiber/v2"

type SuccessResponse struct {
	Data any `json:"data"`
	Meta any `json:"meta,omitempty"`
}

func JSONSuccess(c *fiber.Ctx, data any, meta any) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Data: data,
		Meta: meta,
	})
}

func JSONCreated(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Data: data,
	})
}

type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func JSONBadRequest(c *fiber.Ctx, code, message string, details map[string]any) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{code, message, details})
}

func JSONNotFound(c *fiber.Ctx, code, message string, details map[string]any) error {
	return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{code, message, details})
}

func JSONUnauthorized(c *fiber.Ctx, code, message string, details map[string]any) error {
	return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{code, message, details})
}

func JSONInternalError(c *fiber.Ctx, code, message string, details map[string]any) error {
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{code, message, details})
}

func JSONForbidden(c *fiber.Ctx, code, message string, details map[string]any) error {
	return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{code, message, details})
}

func JSONUnprocessableEntity(c *fiber.Ctx, code, message string, details map[string]any) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(ErrorResponse{code, message, details})
}
