package handlers

import (
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type VoiceLibraryHandler struct {
	service *services.VoiceLibraryService
	log     *logger.Logger
}

func NewVoiceLibraryHandler(cfg *config.Config, log *logger.Logger) (*VoiceLibraryHandler, error) {
	service, err := services.NewVoiceLibraryService(cfg, log)
	if err != nil {
		return nil, err
	}

	return &VoiceLibraryHandler{
		service: service,
		log:     log,
	}, nil
}

func (h *VoiceLibraryHandler) List(c *gin.Context) {
	voices, err := h.service.ListSpeakers(c.Request.Context())
	if err != nil {
		h.log.Errorw("Failed to list voice library", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, voices)
}
