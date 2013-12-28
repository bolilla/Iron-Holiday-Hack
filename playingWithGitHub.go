package main

import (
	"errors"
	"fmt"
	"github.com/octokit/go-octokit/octokit"
	"net/url"
)

func main() {
	client := octokit.NewClient(nil)

	repoUrlStr := "repos/octokit/go-octokit"
	repoUrl, err := url.Parse(repoUrlStr)
	if err != nil {
		fmt.Println("Error parsing repository (", repoUrlStr, ") =>", err)
	}
	repository, err := getRepository(client, repoUrl)
	if err != nil {
		fmt.Println("Error Getting repository information", err)
		return
	}
	fmt.Println("Name:", repository.Name)               //USEFUL PIECE OF INFORMATION
	fmt.Println("Description:", repository.Description) //USEFUL PIECE OF INFORMATION

	contributorsUrlStr := "repos/octokit/go-octokit/contributors"
	contrUrl, err := url.Parse(contributorsUrlStr)
	if err != nil {
		fmt.Println("Error parsing contributors URL (", contributorsUrlStr, ") =>", err)
		return
	}
	contributors, result := client.Users(contrUrl).All()
	fmt.Println("contrUrl", contrUrl)
	if result.HasError() {
		fmt.Println("Error Getting contributors of repository (", contributorsUrlStr, ")")
		return
	}
	for _, contributor := range contributors {
		fmt.Printf("%v - %s - %s - %s\n", contributor.ID, contributor.Login, contributor.Name, contributor.AvatarURL)
	}

	//userURL := &octokit.UserURL

	//fmt.Println("Printing GitHub users for the first 1 page")
	//for i := 0; i < 1; i++ {
	//	if userURL == nil {
	//		return
	//	}

	//	url, err := userURL.Expand(nil)
	//	if err != nil {
	//		fmt.Printf("error: %s\n", err)
	//		return
	//	}

	//	users, result := client.Users(url).All()
	//	if result.HasError() {
	//		fmt.Println(result)
	//		return
	//	}

	//	for _, user := range users {
	//		fmt.Printf("%v - %s\n", user.ID, user.Login)
	//	}

	//	userURL = result.NextPage
	//}
}

//Returns the name of the repository
func getRepository(client *octokit.Client, repoUrl *url.URL) (*octokit.Repository, error) {
	if len(repoUrl.String()) == 0 {
		return nil, errors.New("Empty Repository URL")
	}

	repo, result := client.Repositories(repoUrl).One()
	if result.HasError() {
		return nil, errors.New(fmt.Sprintf("Error getting information of repository => %s", result.Err))
	}
	return repo, nil
}
