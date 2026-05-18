package router

import (
	"github.com/gin-gonic/gin"
	"ajaib-testing-code/internal/adapters/framework/primary/rest_fiber/transfer"
)

type Router struct {
	engine *gin.Engine
}

type Config struct {
	TransferHandler *transfer.Handler
}

func NewRouter(config Config) *Router {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	v1 := engine.Group("/v1")
	{
		transfers := v1.Group("/transfers")
		{
			transfers.POST("", config.TransferHandler.CreateTransferHandler)
			transfers.GET("", config.TransferHandler.GetListTransferHandler)
			transfers.GET("/:id", config.TransferHandler.GetDetailTransferHandler)
			transfers.PATCH("/:id/status", config.TransferHandler.UpdateTransferStatusHandler)
		}
	}

	return &Router{
		engine: engine,
	}
}

func (r *Router) Run(port string) error {
	return r.engine.Run(":" + port)
}
