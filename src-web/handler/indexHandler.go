package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func parseErrInfo(err string) string {
	returnStr := ""
	if err == "logErr" {
		returnStr = "Login failed. Please check the account and password are correct!"
	}
	return returnStr
}

func IndexPage(context *gin.Context) {
	context.HTML(http.StatusOK, "index.gohtml", gin.H{
		"title": "Login Page",
		"error": parseErrInfo(context.Query("err")),
	})
}
