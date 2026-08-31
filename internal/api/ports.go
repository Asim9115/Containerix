package api

import (
	"net/http"
	//"github.com/asim9115/containerix/internal/state"
	"github.com/gin-gonic/gin"
)

func (h *GlobalState) GetPorts(c *gin.Context) {
	ports, err := h.Pipeline.Repo.Ports.GetAll()
	if err != nil {
	c.JSON(http.StatusNotFound, err)
	return
	}
	c.JSON(http.StatusOK, ports)
}
