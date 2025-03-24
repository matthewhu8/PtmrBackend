package service

import "github.com/gin-gonic/gin"

func ErrorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func StatusResponse(msg string) gin.H {
	return gin.H{"status": msg}
}
