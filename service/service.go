package service

import (
	"fmt"
	rice "github.com/GeertJohan/go.rice"
	"github.com/fsnotify/fsnotify"
	"github.com/pflow-dev/go-metamodel/metamodel"
	"github.com/pflow-dev/go-metamodel/metamodel/js"
	"github.com/pflow-dev/go-metamodel/metamodel/lua"
	"github.com/pflow-dev/pflow/codec"
	"github.com/r3labs/sse"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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

func Webserver() {
	go Service()

	// Create a new Mux and set the handler
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(Box.HTTPBox()))
	mux.HandleFunc("/models.json", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET models.json")
		w.Header().Set("Content-Type", "application/json")
		w.Write(ToJson())
	})
	mux.HandleFunc("/sse", EventServer.HTTPHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

type SourceFile struct {
	*metamodel.Model `json:"model"`
}

var Models = make(map[string]*SourceFile, 0)

func WriteModels(format string) {
	switch format {
	case "json":
		fmt.Printf("%s", ToJson())
	default:
		panic("unsupported output format " + format)
	}
}

func ToJson() []byte {
	m := make(map[string]*metamodel.Model, 0)
	for k, v := range Models {
		m[k] = v.Model
	}
	return codec.Marshal(m)
}

func OnModify(event fsnotify.Event) {
	m := LoadModel(event.Name)
	if LastCid != m.Cid {
		LastCid = m.Cid
		EventServer.Publish(StreamId, &sse.Event{Data: []byte(`{ "cid": "` + m.Cid + `" }`)})
	}
}

func LoadModels(path string) error {
	stat, err := os.Stat(ModelPath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			if f.IsDir() {
				LoadModels(filepath.Join(path, f.Name()))
			} else {
				LoadModel(filepath.Join(path, f.Name()))
			}
		}
	} else {
		LoadModel(path)
	}
	return nil
}

func readFile(path string) (source []byte, err error) {
	source, err = os.ReadFile(path)
	for {
		if err != nil {
			time.Sleep(time.Second) // sleep
			source, err = os.ReadFile(path)
			if err != nil {
				log.Println(err)
			}
		} else {
			return source, err
		}
	}
}

func ReloadLua(path string) (src *SourceFile) {
	source, err := readFile(path)
	var m *metamodel.Model
	m, err = lua.LoadModel(string(source))
	if err != nil {
		panic(err)
	}
	name := filepath.Base(path)
	src = &SourceFile{
		Model: m,
	}
	src.Path = name
	src.Cid = codec.ToOid(source).String()
	Models[name] = src
	log.Println("loaded:", name, src.Cid)
	return src
}
func ReloadJs(path string) (src *SourceFile) {
	source, err := readFile(path)
	var m *metamodel.Model
	m, err = js.LoadModel(string(source))
	if err != nil {
		panic(err)
	}
	name := filepath.Base(path)
	src = &SourceFile{
		Model: m,
	}
	src.Path = name
	src.Cid = codec.ToOid(source).String()
	Models[name] = src
	log.Println("loaded:", name, src.Cid)
	return src
}

func getFileExt(path string) string {
	ext := strings.Split(path, ".")
	return ext[len(ext)-1]
}

func LoadModel(path string) (src *SourceFile) {

	fileExt := getFileExt(path)
	switch fileExt {
	case "lua":
		return ReloadLua(path)
	case "js":
		return ReloadJs(path)
	default:
		panic("unrecognized file ext for models " + fileExt)
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
					log.Println(event)
					OnModify(event) // REVIEW: might want catch errors?
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Println("error:", err)
					return
				}
			}
		}
	}()

	LoadModel(ModelPath)
	err = watcher.Add(ModelPath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
