//This worker reads the projects to push to appengine, reads the information from GitHub, pushes the information to appengine and re-enqueues the project to be updated in the next hour
package main

import (
	"encoding/json"
	"github"
	"github.com/iron-io/iron_go/cache"
	"github.com/iron-io/iron_go/config"
	"github.com/iron-io/iron_go/mq"
	"log"
	"sync"
	"time"
)

const (
	queue_to_do_projects = "to_do"
	project_id           = "52bc58376d31d10005000031"
	authentication_token = "qefvrPEPKzEZEPNSZw7vz0VUh-s"
)

func main() {
	cfg := config.Config("iron_mq")
	log.Println("Config:", cfg)
	cfg.ProjectId = project_id
	cfg.Token = authentication_token

	q := mq.New(queue_to_do_projects)
	q.Settings = cfg
	info, err := q.Info()
	if err != nil {
		log.Println("Queue info error ("+queue_to_do_projects+"):", err)
		return
	}
	log.Println("Queue \"", q, "\":", info)
	sisyphus(q)
}

//Copies the messages in qFrom to qTo and waits one minute before checking again
func sisyphus(q *mq.Queue) {
	for {
		processQueue(q)
		time.Sleep(time.Minute / 2)
	}
}

//processes the messages in the queue
func processQueue(q *mq.Queue) {
	var wg sync.WaitGroup
	log.Println("Checking queue...")
	for msg, err := q.Get(); err == nil; msg, err = q.Get() {
		wg.Add(1)
		go manageProject(q, msg, &wg)
	}
	log.Println("Waiting for the end of the pushing")
	wg.Wait()
	log.Println("All project information updated.")
}

//Manages the information in a project
func manageProject(q *mq.Queue, msg *mq.Message, wg *sync.WaitGroup) {
	defer wg.Done()
	projectInformation, err := github.GetProjectInformation(msg.Body)
	if err != nil {
		log.Println("Error getting information of project:", msg)
	} else {
		projectInformationJson, _ := json.Marshal(projectInformation)
		err = pushProjectInformation(projectInformation.Name, projectInformationJson)
		if err != nil {
			log.Println("Error pushing information of project", msg)
		} else {
			msg.Delete()
			msg.Delay = 3600
			_, err = q.PushMessage(msg)
		}
	}
}

//Pushes the information to appengine
func pushProjectInformation(name string, projectInformation []byte) error {
	cfg := config.Config("iron_cache")
	log.Println("Config:", cfg)
	cfg.ProjectId = project_id
	cfg.Token = authentication_token
	c := cache.New("projects")
	c.Settings = cfg
	return c.Set(name, string(projectInformation))
}
