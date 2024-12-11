package log

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// ILogController is an interface that defines the methods for a user controller.
type ILogController interface {
	GetLogs(c *gin.Context)
	GetAccessLogs(c *gin.Context)
}

// LogController is a struct that implements the ILogController interface.
type LogController struct {
}

// NewLogController is a constructor function that creates a new LogController.
func NewLogController() ILogController {
	return &LogController{}
}

// GetLogs is a function that returns a list of logs.
func (lc *LogController) GetLogs(c *gin.Context) {
	// Open the log file
	filePath := "/tmp/shopi.log"
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open log file"})
		return
	}
	defer file.Close()

	// Set the file headers for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+"shopi-"+time.Now().Format("2006-01-02T15:04:05")+".log")
	c.Header("Content-Type", "application/octet-stream")

	// Serve the static log file from the given path
	c.File(filePath)
}

// GetAccessLogs is a function that returns a list of access logs.
func (lc *LogController) GetAccessLogs(c *gin.Context) {
	// Open the log file
	filePath := "/tmp/shopi-access.log"
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open access log file"})
		return
	}
	defer file.Close()

	// Set the file headers for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+"shopi-access-"+time.Now().Format("2006-01-02T15:04:05")+".log")
	c.Header("Content-Type", "application/octet-stream")

	// Serve the static log file from the given path
	c.File(filePath)
}
