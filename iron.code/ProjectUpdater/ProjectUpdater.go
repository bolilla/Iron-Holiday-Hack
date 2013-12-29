//This worker pushes the projects to update from the scheduled queue to the to-do queue to be managed by "Sisyphus"
package main

import (
	"github.com/iron-io/iron_go/config"
	"github.com/iron-io/iron_go/mq"
	"log"
	"sync"
)

const (
	queue_scheduled_projects = "scheduled_projects"
	queue_to_do_projects     = "to_do"
)

func main() {
	cfg := config.Config("iron_mq")
	log.Println("Config:", cfg)
	cfg.ProjectId = "52bc58376d31d10005000031"
	cfg.Token = "qefvrPEPKzEZEPNSZw7vz0VUh-s"

	qFrom := mq.New(queue_scheduled_projects)
	qFrom.Settings = cfg
	info, err := qFrom.Info()
	if err != nil {
		log.Println("Queue info error ("+queue_scheduled_projects+"):", err)
		return
	}
	log.Println("Queue \"", qFrom, "\":", info)

	qTo := mq.New(queue_to_do_projects)
	qTo.Settings = cfg
	info, err = qTo.Info()
	if err != nil {
		log.Println("Queue info error ("+queue_to_do_projects+"):", err)
		return
	}
	log.Println("Queue \"", qTo, "\":", info)

	var wg sync.WaitGroup
	for msg, err := qFrom.Get(); err == nil; msg, err = qFrom.Get() {
		wg.Add(1)
		go push(qTo, msg, &wg)
	}
	log.Println("Waiting for the end of the pushing")
	wg.Wait()
	log.Println("All messages pushed from \"" + queue_scheduled_projects + "\" to \"" + queue_to_do_projects + "\"")
}

func push(q *mq.Queue, msg *mq.Message, wg *sync.WaitGroup) {
	defer wg.Done()
	id, err := q.PushMessage(msg)
	if err == nil {
		log.Println("Message \""+id+"\" pushed: ", msg)
	} else {
		log.Println("Error pushing \"", msg, "\":", err)
	}
}
