package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

type YouTubeHandler struct{}

type PlaylistItem struct {
	VideoId      string `json:"videoId"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Thumbnail    string `json:"thumbnail"`
	ChannelTitle string `json:"channelTitle"`
}

type PlaylistResponse struct {
	Items []PlaylistItem `json:"items"`
}

func NewYouTubeHandler() *YouTubeHandler {
	return &YouTubeHandler{}
}

func (h *YouTubeHandler) GetPlaylist(c *fiber.Ctx) error {
	playlistId := c.Query("id")
	if playlistId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Playlist ID is required",
		})
	}

	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "YouTube API key not configured",
		})
	}

	// Fetch playlist items from YouTube API
	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=50&playlistId=%s&key=%s",
		playlistId,
		apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch playlist",
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read response",
		})
	}

	// Parse YouTube API response
	var ytResponse struct {
		Items []struct {
			Snippet struct {
				Title        string `json:"title"`
				Description  string `json:"description"`
				ChannelTitle string `json:"channelTitle"`
				Thumbnails   struct {
					High struct {
						Url string `json:"url"`
					} `json:"high"`
				} `json:"thumbnails"`
				ResourceId struct {
					VideoId string `json:"videoId"`
				} `json:"resourceId"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &ytResponse); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse response",
		})
	}

	// Convert to our response format
	var response PlaylistResponse
	for _, item := range ytResponse.Items {
		response.Items = append(response.Items, PlaylistItem{
			VideoId:      item.Snippet.ResourceId.VideoId,
			Title:        item.Snippet.Title,
			Description:  item.Snippet.Description,
			Thumbnail:    item.Snippet.Thumbnails.High.Url,
			ChannelTitle: item.Snippet.ChannelTitle,
		})
	}

	return c.JSON(response)
}
