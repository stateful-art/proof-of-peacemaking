package handlers

import (
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

type ProofNFTHandler struct {
	proofNFTService ports.ProofNFTService
}

func NewProofNFTHandler(proofNFTService ports.ProofNFTService) *ProofNFTHandler {
	return &ProofNFTHandler{
		proofNFTService: proofNFTService,
	}
}

func (h *ProofNFTHandler) RequestProof(c *fiber.Ctx) error {
	var body struct {
		ExpressionID      string `json:"expressionId"`
		AcknowledgementID string `json:"acknowledgementId"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	return h.proofNFTService.RequestProof(c.Context(), body.ExpressionID, body.AcknowledgementID)
}

func (h *ProofNFTHandler) ApproveProof(c *fiber.Ctx) error {
	requestID := c.Params("id")
	return h.proofNFTService.ApproveProof(c.Context(), requestID)
}

func (h *ProofNFTHandler) ListUserProofs(c *fiber.Ctx) error {
	userAddress := c.Locals("userAddress").(string)
	proofs, err := h.proofNFTService.ListUserProofs(c.Context(), userAddress)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(proofs)
}
