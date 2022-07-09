package service

import (
	"fmt"
	rice "github.com/GeertJohan/go.rice"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
	"github.com/pflow-dev/go-metamodel/metamodel/image"
	"github.com/pflow-dev/pflow/codec"
	"github.com/pflow-dev/pflow/model/source"
	"github.com/r3labs/sse"
	"log"
	"net/http"
)

var (
	// Box encapsulates static files
	Box *rice.Box

	// ModelPath - a pathname provided by cli arg
	ModelPath   = ""
	EventServer = sse.New()

	StreamId = "models"
	LastCid  = ""
)

func init() {
	EventServer.CreateStream(StreamId)
}

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	vars := mux.Vars(r)

	cid := vars["cid"]

	m := source.GetModel(cid)
	if m == nil {
		i := image.NewSvg(w, 512, 256)
		i.Text(50, 50, fmt.Sprintf(`no matching model for cid: %s`, cid))
		i.End()
		return
	}
	width, height := source.GetSize(m)
	i := image.NewSvg(w, width, height)
	state := m.InitialVector()

	q := r.URL.Query()
	rawState := q.Get("state") // fallback to query param
	if rawState != "" {
		err := codec.Unmarshal([]byte(rawState), &state)
		if err != nil {
			i.Text(50, 50, fmt.Sprintf(`error parsing state-vector: %v`, state))
			i.End()
			return
		}
		if len(state) != 0 && len(state) != len(m.Places) {
			i.Text(50, 50, fmt.Sprintf(`invalid state-vector: %v expected len: %v`, state, len(m.Places)))
			i.End()
			return
		}
	}
	if len(state) == 0 {
		state = m.InitialVector()
	}
	i.Render(m, state)
}

func Webserver() {
	go Service()

	// Create a new Mux and set the handler
	router := mux.NewRouter()
	router.HandleFunc("/models.json", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET models.json")
		// TODO: add search params
		w.Header().Set("Content-Type", "application/json")
		w.Write(source.ToJson())
	})
	router.HandleFunc("/{cid}/image.svg", ImageHandler)
	router.HandleFunc("/sse", EventServer.HTTPHandler)
	router.PathPrefix("/").Handler(http.FileServer(Box.HTTPBox()))

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		panic(err)
	}
}

func OnModify(event fsnotify.Event) {
	m := source.LoadModel(event.Name)
	if m != nil && LastCid != m.Cid {
		LastCid = m.Cid
		EventServer.Publish(StreamId, &sse.Event{Data: []byte(`{ "cid": "` + m.Cid + `" }`)})
	}
}

func Service() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if ok {
					if event.Op == fsnotify.Remove {
						watcher.Remove(event.Name)
						watcher.Add(event.Name) // NOTE: editors like Vim does RENAME+CHMOD+REMOVE on write
					}
					OnModify(event)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Println("error:", err)
					return
				}
			}
		}
	}()

	source.LoadModels(ModelPath)
	err = watcher.Add(ModelPath) // REVIEW: will this work watching a directory?
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
