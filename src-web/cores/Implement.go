package cores

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// InformationStruct
/*
# Return Information
ID := result.id
CID := result.contestId
PID := result.problem.index
PNAME := result.problem.name
CNAME := result.author.members.[name(Maybe NULL)/handle]
LANG := result.programmingLanguage

fileName := PID-PNAME-CNAME-LANG(CID#ID)
*/
type InformationStruct struct {
	ID    int64
	CID   int64
	PID   string
	PNAME string
	CNAME string
	LANG  string
	//TODO F: 对部分缩写进行重构
}

func saveSourceCodeToFile(sourceCode string, infoCode InformationStruct) {
	sufferName := ".txt"
	abbrLANG := "Other"
	if strings.Contains(infoCode.LANG, "C++") || strings.Contains(infoCode.LANG, "G++") || strings.Contains(infoCode.LANG, "Clang") {
		sufferName = ".cpp"
		abbrLANG = "C++"
	} else if strings.Contains(infoCode.LANG, "Java") {
		sufferName = ".java"
		abbrLANG = "Java"
	} else if strings.Contains(infoCode.LANG, "Python") || strings.Contains(infoCode.LANG, "PyPy") {
		sufferName = ".py"
		abbrLANG = "Python"
	} else if strings.Contains(infoCode.LANG, "GCC") {
		sufferName = ".c"
		abbrLANG = "C"
	}
	invalidChar := "<|>\\/:\"*?"
	fileName := fmt.Sprintf("%s-%s-%s-%s(%d#%d)%s", infoCode.PID, infoCode.PNAME, infoCode.CNAME, abbrLANG, infoCode.CID, infoCode.ID, sufferName)
	if strings.ContainsAny(fileName, invalidChar) {
		invalidRegexp := regexp.MustCompile(`[` + invalidChar + `]`)
		newFileName := invalidRegexp.ReplaceAllString(fileName, "")
		LogServer.WithFields(logrus.Fields{
			"oldFileName": fileName,
			"newFileName": newFileName,
		}).Warnln("The constructed file name may contain illegal characters that are not allowed by the operating system.")
		fileName = newFileName
	}
	err := ioutil.WriteFile("./temp/"+RandomTaskName+"/"+fileName, []byte(sourceCode), 0664)
	if err != nil {
		LogServer.WithFields(logrus.Fields{
			"reason":   err.Error(),
			"fileName": fileName,
		}).Errorln("An error occurred while saving the fetched code to a file.")
	}
}

/*
# Condition
result.author.participantType = "CONTESTANT"
result.verdict = "OK",
*/
func getAllAcceptSubmissionData(signedURL string, manageClient *http.Client) []InformationStruct {
	apiJsonString := getAPIJsonString(signedURL)
	allAcceptSubmissionID := getAllAcceptSubmissionID(apiJsonString)
	if len(allAcceptSubmissionID) == 0 {
		LogServer.WithFields(logrus.Fields{
			"signedURL": signedURL,
		}).Errorln("The list of obtained submission records is empty.")
	}
	var allNeedInformation []InformationStruct
	for index, submissionID := range allAcceptSubmissionID {
		infoForID := gjson.Get(apiJsonString, `result.#(id=`+string(submissionID)+`)#`)
		temp := parseJsonFiles(infoForID)
		LogServer.WithFields(logrus.Fields{
			"CID":   temp.CID,
			"ID":    temp.ID,
			"CNAME": temp.CNAME,
			"LANG":  temp.LANG,
		}).Infoln("Fetching this source...")
		taskLogMapRef, _ := TaskMessageChan.Load(RandomTaskName)
		taskLogMapRef.(chan string) <- "[DEBUG] " + strconv.FormatInt(temp.CID, 10) + "#" + strconv.FormatInt(temp.ID, 10) + "#" + temp.CNAME + "#" + temp.LANG
		var tempSourceCode string
		if temp.CID > 100000 {
			tempSourceCode = fetchSubmissionCode(fmt.Sprintf(`https://codeforces.com/gym/%d/submission/%d`, temp.CID, temp.ID), manageClient)
		} else {
			tempSourceCode = fetchSubmissionCode(fmt.Sprintf(`https://codeforces.com/contest/%d/submission/%d`, temp.CID, temp.ID), manageClient)
		}
		saveSourceCodeToFile(tempSourceCode, temp)
		allNeedInformation = append(allNeedInformation, temp)
		LogServer.WithFields(logrus.Fields{
			"Index":        index,
			"SubmissionID": submissionID,
		}).Infoln("Have fetched this source...")
		//TODO F: 此处使用全局变量来计算，后期需修正
		PROCESS.Store(RandomTaskName, float64(index+1)/float64(len(allAcceptSubmissionID)))
	}
	ZipCompress("./temp/"+RandomTaskName, "./temp/"+RandomTaskName)
	return allNeedInformation
}
