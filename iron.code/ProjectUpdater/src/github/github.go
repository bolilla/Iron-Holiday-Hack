package github

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

	rate_per_minute = 600
)

var throttle chan time.Time

//func main() {
//	fmt.Println("Result:", GetProjectInformation("iron-io/iron_go"))
//}

//Downloads project information from github and returns it in a pretty struct
func GetProjectInformation(projectNameInput string) (result prjInfo, err error) {
	var initialized bool
	if !initialized {
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
		initialized = true
	}

	projectInfo, err := getProjectInfo("https://api.github.com/repos/" + projectNameInput)
	if err != nil || projectInfo == nil {
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
	result.Name = projectName
	result.Description = projectDescription
	result.Contributors = make([]cntrInfo, 0)
	//fmt.Println("Contributors:")
	for _, contributor := range projectContributors {
		var cont cntrInfo
		var userLogin, userName, userAvatar string
		var userContributions float64
		if contributor[0][CONTRIB_USER_LOGIN] != nil {
			userLogin = contributor[0][CONTRIB_USER_LOGIN].(string)
		}
		if contributor[0][CONTRIB_NUM_CONTRIBS] != nil {
			userContributions = contributor[0][CONTRIB_NUM_CONTRIBS].(float64)
		}
		if contributor[1][CONTRIB_USER_NAME] != nil {
			userName = contributor[1][CONTRIB_USER_NAME].(string)
		}
		if contributor[0][CONTRIB_USER_AVATAR] != nil {
			userAvatar = contributor[0][CONTRIB_USER_AVATAR].(string)
		}
		cont.Login = userLogin
		cont.Contributions = int(userContributions)
		cont.Name = userName
		cont.AvatarUrl = userAvatar
		result.Contributors = append(result.Contributors, cont)
		//fmt.Printf("  %s (%f contribs): %s - %s\n",
		//userLogin, userContributions, userName, userAvatar)
	}
	result.Languages = make([]lngInfo, 0)
	//fmt.Println("Languajes:")
	var totalLines int
	for lang, lines := range projectLanguajes {
		var lng lngInfo
		lng.Name = lang
		lng.Lines = lines
		//fmt.Printf("  %s => %d lines\n", lang, lines)
		totalLines += lines
		result.Languages = append(result.Languages, lng)
	}
	//fmt.Printf("  Total => %d lines\n", totalLines)
	//fmt.Println("Result =>", result)
	return
}

//Contains the relevant information about the project
type prjInfo struct {
	Name         string     `json:"name,omitempty"`
	Description  string     `json:"description,omitempty"`
	Contributors []cntrInfo `json:"contributors,omitempty"`
	Languages    []lngInfo  `json:"languages,omitempty"`
}

//Contains the relevant information about a contributor
type cntrInfo struct {
	Login         string `json:"login,omitempty"`
	Name          string `json:"name,omitempty"`
	AvatarUrl     string `json:"avatarUrl,omitempty"`
	Contributions int    `json:"contributions,omitempty"`
}

//Contains the relevant information about the languajes used in the project
type lngInfo struct {
	Name  string `json:"name,omitempty"`
	Lines int    `json:"lines,omitempty"`
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
		if err != nil || jsonTmp == nil {
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
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth("bolilla", "1702aa1d6465c078cb799847cdd3086d47e58322")
	fmt.Println("Accessing:", url)
	res, err := client.Do(req)
	//res, err := http.Get(url)
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
