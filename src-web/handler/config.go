package handler

import (
	"Codeforces-ContestCodeDownload/src-web/cores"
	"Codeforces-ContestCodeDownload/src-web/model"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

// decryptUserData TODO F: 添加加密安全传输信息
func decryptUserData(encryptedInformation model.CodeforcesUserModel, decryptedKey string) model.CodeforcesUserModel {
	return encryptedInformation
}

// CodeforcesUserAuth TODO F: 验证通过后，重定向到/result页面，实时显示抓取情况，并展示进度条，否则重定向到/error页面
func CodeforcesUserAuth(context *gin.Context) {
	cores.MissionInitiated()
	encryptedApiKey := context.PostForm("apiKey")
	encryptedApiSecret := context.PostForm("apiSecret")
	encryptedUsername := context.PostForm("usernameOrEmail")
	encryptedPassword := context.PostForm("password")
	encryptedUserData := model.CodeforcesUserModel{
		ApiKey:    encryptedApiKey,
		ApiSecret: encryptedApiSecret,
		Username:  encryptedUsername,
		Password:  encryptedPassword,
	}
	//TODO F: 添加空值校验和账号密码校验（抓取登陆返回值），暂不对API KEY校验。
	userData := decryptUserData(encryptedUserData, "123")
	//fmt.Println("1112222222")
	if checkLoginStatus(cores.GetCodeforcesHttpClient(userData.Username, userData.Password)) == false {
		//fmt.Println("111222333")
		cores.LogServer.Errorln("Login fail. Please check your username and password.")
		context.Abort()
		return
	}
	cores.LogServer.WithFields(logrus.Fields{
		"ApiKey":   userData.ApiKey,
		"Username": userData.Username,
	}).Info("Have access to user information.")
	contestID := 381185
	result := cores.MissionStart(contestID, userData)
	cores.LogServer.WithFields(logrus.Fields{
		"contestID":  contestID,
		"jsonResult": result,
	}).Info("Source code and record correspondence information has been obtained from codeforces.")
	context.Set("CodeforcesResult", result)
	context.Next()
}

//TODO F: 或许将这一检查写成client, error会更合适。
func checkLoginStatus(client *http.Client, response *http.Response) bool {
	body, _ := ioutil.ReadAll(response.Body)
	//fmt.Println(string(body))
	if strings.Contains(string(body), "Invalid handle/email or password") ||
		strings.Contains(string(body), "Please, confirm email before entering the website.") {
		return false
	}
	return true
}
