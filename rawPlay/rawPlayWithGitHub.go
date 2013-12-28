package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	PRJ_NAME             = "name"
	PRJ_DESC             = "description"
	PRJ_CONTRIB          = "contributors_url"
	PRJ_LANGUAJES        = "languages_url"
	CONTRIB_USER_URL     = "url"
	CONTRIB_USER_LOGIN   = "login"
	CONTRIB_NUM_CONTRIBS = "contributions"
	CONTRIB_USER_NAME    = "name"
	CONTRIB_USER_AVATAR  = "avatar_url"

	rate_per_minute = 60
)

var throttle chan time.Time

func main() {
	throttle = make(chan time.Time, rate_per_minute)
	tick := time.NewTicker(time.Minute / rate_per_minute)
	defer tick.Stop()

	go func() {
		for ns := range tick.C {
			select {
			case throttle <- ns:
			default:
			}
		}
	}()

	var projectNameInput string = "octokit/go-octokit"
	projectInfo, err := getProjectInfo("https://api.github.com/repos/" + projectNameInput)
	if err != nil {
		return
	}
	projectJSON := projectInfo.(map[string]interface{})
	projectName := projectJSON[PRJ_NAME].(string)
	projectDescription := projectJSON[PRJ_DESC].(string)
	projectContributors, err := getProjectContributorsInfo(projectJSON[PRJ_CONTRIB].(string))
	if err != nil {
		return
	}
	projectLanguajes, err := getProjectLanguajesInfo(projectJSON[PRJ_LANGUAJES].(string))
	if err != nil {
		return
	}
	fmt.Printf("Project:%s (%s)\n", projectName, projectDescription)
	fmt.Println("Contributors:")
	for _, contributor := range projectContributors {
		userLogin := contributor[0][CONTRIB_USER_LOGIN].(string)
		userContributions := contributor[0][CONTRIB_NUM_CONTRIBS].(float64)
		userName := contributor[1][CONTRIB_USER_NAME].(string)
		userAvatar := contributor[0][CONTRIB_USER_AVATAR].(string)
		fmt.Printf("  %s (%f contribs): %s - %s\n",
			userLogin, userContributions, userName, userAvatar)
	}
	fmt.Println("Languajes:")
	var totalLines int
	for lang, lines := range projectLanguajes {
		fmt.Printf("  %s => %d lines\n", lang, lines)
		totalLines += lines
	}
	fmt.Printf("  Total => %d lines\n", totalLines)
}

//Returns the number of lines of code per languaje
func getProjectLanguajesInfo(url string) (result map[string]int, err error) {
	jsonInfo, err := getJson(url)
	if err != nil {
		return
	}
	result = make(map[string]int)
	for lang, lines := range jsonInfo.(map[string]interface{}) {
		result[lang] = int(lines.(float64))
	}
	return
}

//Returns an interface that contains the JSON information in the project or an error if an error has happened
func getProjectInfo(projectUrl string) (result interface{}, err error) {
	return getJson(projectUrl)
}

//Returns an array with the information of the contributors. Per contributor an array is returned. position 0 contains the information about user's contribution to the project. position 1 contains information about the user itself
func getProjectContributorsInfo(projectContributorsUrl string) (result [][]map[string]interface{}, err error) {
	contributors, err := getJson(projectContributorsUrl)
	if err != nil {
		return
	}
	result = make([][]map[string]interface{}, len(contributors.([]interface{})))
	for i, contributorBase := range contributors.([]interface{}) {
		result[i] = make([]map[string]interface{}, 2)
		result[i][0] = contributorBase.(map[string]interface{})
		var jsonTmp interface{}
		jsonTmp, err = getJson(result[i][0][CONTRIB_USER_URL].(string))
		if err != nil {
			return
		}
		result[i][1] = jsonTmp.(map[string]interface{})
	}
	fmt.Println("Number of contributors:", len(result))
	return
}

//Returns the Json items recovered from a URL
func getJson(url string) (result interface{}, err error) {
	<-throttle
	fmt.Println("Accessing:", url)
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("Error accessing information in \""+url+"\":", err)
	} else {
		resultJson, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			fmt.Println("Error getting information from \""+url+"\":", err)
		} else {
			err = json.Unmarshal(resultJson, &result)
		}
	}
	return result, err
}
