//This worker reads the projects to push to appengine, reads the information from GitHub, pushes the information to appengine and re-enqueues the project to be updated in the next hour
package main

import (
	"github.com/iron-io/iron_go/config"
	"github.com/iron-io/iron_go/mq"
	"log"
	"sync"
	"time"
)

const (
	queue_to_do_projects = "to_do"
)

func main() {
	cfg := config.Config("iron_mq")
	log.Println("Config:", cfg)
	cfg.ProjectId = "52bc58376d31d10005000031"
	cfg.Token = "qefvrPEPKzEZEPNSZw7vz0VUh-s"

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
	projectInformation, err := getProjectInformation(msg)
	if err != nil {
		log.Println("Error getting information of project:", msg)
	} else {
		err = pushProjectInformation(projectInformation)
		if err != nil {
			log.Println("Error pushing information of project", msg)
		} else {
			msg.Delete()
			msg.Delay = 3600
			_, err = q.PushMessage(msg)
		}
	}
	//id, err := qTo.PushMessage(msg)
	//if err == nil {
	//	log.Println("Message \""+id+"\" pushed: ", msg)
	//	msg.Delay = 3600
	//	_, err = qFrom.PushMessage(msg)
	//	if err != nil {
	//		log.Println("Error copying message \""+id+"\": ", err)
	//	} else {
	//		msg.Delete()
	//		log.Println("Message \"" + id + "\" re-enqueued")
	//	}
	//} else {
	//	log.Println("Error pushing \"", msg, "\":", err)
	//}
}

//Gets the information from GitHub
func getProjectInformation(msg *mq.Message) (result prjInfo, err error) {
	log.Println("Mocking project information retrieval", msg.Body)
	return
}

//Pushes the information to appengine
func pushProjectInformation(projectInformation prjInfo) error {
	log.Println("Mocking project information pushing")
	return nil
}

//Contains all the information elements of a project
type prjInfo struct {
}
