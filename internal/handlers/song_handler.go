package handlers

import (
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SongHandler struct {
	songService ports.SongService
}

func NewSongHandler(songService ports.SongService) *SongHandler {
	return &SongHandler{
		songService: songService,
	}
}

func (h *SongHandler) AddSong(c *fiber.Ctx) error {
	var input struct {
		URL string `json:"url"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userAddress := c.Locals("userAddress").(string)
	err := h.songService.AddSong(c.Context(), input.URL, userAddress)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Song added successfully",
	})
}

func (h *SongHandler) GetNextSong(c *fiber.Ctx) error {
	song, err := h.songService.GetNextSong(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if song == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No songs in queue",
		})
	}

	return c.JSON(song)
}

func (h *SongHandler) MarkSongAsPlayed(c *fiber.Ctx) error {
	songID := c.Params("id")
	if songID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Song ID is required",
		})
	}

	objectID, err := primitive.ObjectIDFromHex(songID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid song ID format",
		})
	}

	err = h.songService.MarkSongAsPlayed(c.Context(), objectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Song marked as played",
	})
}

func (h *SongHandler) GetCurrentlyPlaying(c *fiber.Ctx) (*domain.Song, error) {
	return h.songService.GetCurrentlyPlaying(c.Context())
}

func (h *SongHandler) GetQueue(c *fiber.Ctx) error {
	songs, err := h.songService.GetQueue(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(songs)
}

func (h *SongHandler) GetArchive(c *fiber.Ctx) error {
	songs, err := h.songService.GetArchive(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(songs)
}
