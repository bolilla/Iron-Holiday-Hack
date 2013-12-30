//This file contains the main operations to manage http requests of the users
package webapp

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

const (
	datastore_name       = "projects"
	project_id           = "52bc58376d31d10005000031"
	authentication_token = "qefvrPEPKzEZEPNSZw7vz0VUh-s"
	cache_name           = "projects"
	queue_name           = "to_do"
)

type project struct {
	Name string
	Date time.Time
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/project", projectInfo)
	http.HandleFunc("/addproject", addProject)
}

// guestbookKey returns the key used for all guestbook entries.
func guestbookKey(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, datastore_name, "default_project_set", 0, nil)
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery(datastore_name).Ancestor(guestbookKey(c)).Order("-Date")
	projects := make([]project, 0, 100)
	if _, err := q.GetAll(c, &projects); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := projectListTemplate.Execute(w, projects); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var projectListTemplate = template.Must(template.New("book").Parse(projectListTemplateHTML))

const projectListTemplateHTML = `
<html>
  <body>
    {{range .}}
	 <p>Take a look at project <a href="/project?name={{.Name}}">{{.Name}}</a></p>
    {{end}}
    <form action="/addproject" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Add project"></div>
    </form>
  </body>
</html>
`

//Adds a project to the set of projects to be warched. It adds the project to appenine datastore as well as to ironMq for further treatment
func addProject(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	g := project{
		Name: r.FormValue("content"),
		Date: time.Now(),
	}
	addProjectToQueue(g.Name, w, r)
	key := datastore.NewIncompleteKey(c, datastore_name, guestbookKey(c))
	_, err := datastore.Put(c, key, &g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

//Sends the project to the list of projects to manage
func addProjectToQueue(name string, w http.ResponseWriter, r *http.Request) {
	reader := strings.NewReader("{\"messages\":[{\"body\":\"" + url.QueryEscape(name) + "\"}]}")
	req, _ := http.NewRequest("POST", "https://mq-aws-us-east-1.iron.io/1/projects/52bc58376d31d10005000031/queues/to_do/messages",
		reader)
	req.Header.Set("Authorization", "OAuth "+authentication_token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip/deflate")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "iron_go/cache 1.0 (Go 1.1.2)")

	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error adding new project to the queue: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Project added to the queue.\n"))
	toWrite, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error writting information retrieved from iron cache: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(toWrite)
}

//Shows the information regarding a project
func projectInfo(w http.ResponseWriter, r *http.Request) {
	projectName := r.FormValue("name")
	req, err := http.NewRequest("GET",
		"https://cache-aws-us-east-1.iron.io/1/projects/52bc58376d31d10005000031/caches/projects/items/"+url.QueryEscape(projectName),
		nil)
	req.Header.Set("Authorization", "OAuth "+authentication_token)
	req.Header.Set("Accept", "application/json")
	//req.Header.Set("Accept-Encoding", "gzip/deflate")
	req.Header.Set("User-Agent", "iron_go/cache 1.0 (Go 1.1.2)")

	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error retrieving information from iron cache: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Showing information of " + projectName + ":\n\n"))
	toWrite, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error writting information retrieved from iron cache: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(toWrite)
}
