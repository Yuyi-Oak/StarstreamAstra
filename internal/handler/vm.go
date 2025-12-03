package handler

import (
	"net/http"
	"strconv"

	"Zjmf-kvm/internal/hypervisor"
	"Zjmf-kvm/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterVMHandlers(rg *gin.RouterGroup, db *gorm.DB) {
	hv := hypervisor.NewQEMUHypervisor()
	vmService := service.NewVMService(db, hv)

	rg.GET("/list", func(c *gin.Context) {
		vms, err := vmService.ListVMs()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": vms})
	})

	rg.POST("/create", func(c *gin.Context) {
		var req service.VMCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		vm, err := vmService.CreateVM(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": vm})
	})

	rg.POST("/:id/start", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := vmService.StartVM(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "VM started"})
	})

	rg.POST("/:id/stop", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := vmService.StopVM(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "VM stopped"})
	})

	rg.DELETE("/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := vmService.DeleteVM(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "VM deleted"})
	})

	rg.PATCH("/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var req service.VMCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		cfg := hypervisor.VMConfig{
			Name:     req.Name,
			CPU:      req.CPU,
			MemoryMB: req.MemoryMB,
			DiskGB:   req.DiskGB,
		}
		if err := vmService.ResizeVM(uint(id), cfg); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "VM configuration updated"})
	})
}
