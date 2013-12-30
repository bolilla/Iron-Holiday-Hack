package main

import (
	"github.com/iron-io/iron_go/config"
	"github.com/iron-io/iron_go/mq"
	"log"
)

func main() {
	// Create your configuration for iron_worker
	// Find these value in credentials
	cfg := config.Config("iron_mq")
	log.Println("my config:", cfg)
	cfg.ProjectId = "52bc58376d31d10005000031"
	cfg.Token = "qefvrPEPKzEZEPNSZw7vz0VUh-s"

	q := mq.New("my_queue")
	log.Println("Queue:", q)
	//err := q.Clear()
	q.Settings = cfg
	//if err != nil {
	//	log.Println("Queue clearing error:", err)
	//	return
	//}
	//log.Println("Queue cleared:", q)
	info, err := q.Info()
	if err != nil {
		log.Println("Queue info error:", err)
		return
	}
	log.Println("Queue info:", info)
	id, err := q.PushString("Hello, World 23!")
	if err != nil {
		log.Println("Message Push error:", err)
		return
	}
	log.Println("Message ID:", id)
	// get a single message
	msg, err := q.Get()
	if err != nil {
		log.Println("Message Get error:", err)
		return
	}
	log.Printf("The message says: %q\n", msg.Body)
	msg.Delete()
	//// Capture info for this code
	//codeId := "522d160a91c530531f6f528d"

}
