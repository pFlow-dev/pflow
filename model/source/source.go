package source

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pflow-dev/go-metamodel/metamodel"
	"github.com/pflow-dev/go-metamodel/metamodel/image"
	"github.com/pflow-dev/go-metamodel/metamodel/js"
	"github.com/pflow-dev/go-metamodel/metamodel/lua"
	"github.com/pflow-dev/pflow/codec"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type File struct {
	*metamodel.Model `json:"model"`
	Source           []byte
}

var Models = make(map[string]*File, 0)

func LoadModels(path string) error {
	stat, err := os.Stat(path)
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

// GetModel retrieve model def by cid
func GetModel(cid string) *metamodel.Model {
	for _, v := range Models {
		if v.Cid == cid {
			return v.Model
		}
	}
	return nil
}

// GetFirstModel  get first model in set
func GetFirstModel() *metamodel.Model {
	for _, v := range Models {
		return v.Model
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

func ReloadLua(path string) (src *File) {
	source, err := readFile(path)
	var m *metamodel.Model
	m, err = lua.LoadModel(string(source))
	if err != nil {
		panic(err)
	}
	name := filepath.Base(path)
	src = &File{
		Model:  m,
		Source: source,
	}
	src.Path = name
	src.Cid = codec.ToOid(source).String()
	Models[name] = src
	log.Println("loaded:", name, src.Cid)
	return src
}

func ReloadJs(path string) (src *File) {
	source, err := readFile(path)
	var m *metamodel.Model
	m, err = js.LoadModel(string(source))
	if err != nil {
		panic(err)
	}
	name := filepath.Base(path)
	src = &File{
		Model:  m,
		Source: source,
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

func LoadModel(path string) (src *File) {

	fileExt := getFileExt(path)
	switch fileExt {
	case "lua":
		return ReloadLua(path)
	case "js":
		return ReloadJs(path)
	case "md":
	default:
		//log.Printf("unrecognized file ext for models :%s", fileExt)
	}
	return nil
}

func GetSize(m *metamodel.Model) (width int, height int) {
	var limitX int64 = 0
	var limitY int64 = 0

	for _, p := range m.Places {
		if limitX < p.X {
			limitX = p.X
		}
		if limitY < p.Y {
			limitY = p.Y
		}
	}
	for _, t := range m.Transitions {
		if limitX < t.X {
			limitX = t.X
		}
		if limitY < t.Y {
			limitY = t.Y
		}
	}
	const margin = 40

	if width == 0 {
		width = int(limitX) + margin
	}
	if height == 0 {
		height = int(limitY) + margin
	}
	return width, height
}

func WriteModels(format string) {
	switch format {
	case "json":
		fmt.Printf("%s", ToJson())
	case "svg":
		body, _ := ToSvg("")
		fmt.Printf("%s", body)
	default:
		panic("unsupported output format " + format)
	}
}

func ToSvg(cid string) (body []byte, ok bool) {
	var buf = new(bytes.Buffer)
	w := bufio.NewWriter(buf)
	for _, net := range Models {
		if net.Cid == cid || cid == "" {
			width, height := GetSize(net.Model)
			i := image.NewSvg(w, width, height)
			i.Render(net.Model)
			ok = true
			break // KLUDGE: only one model
		}
	}
	err := w.Flush()
	if err != nil {
		panic(err)
	}
	return buf.Bytes(), ok
}

type modelConfig struct {
	Cid      string `json:"cid"`
	Markdown string `json:"markdown"`
	Source   string `json:"source"`
}

type modelCollection struct {
	Version string                      `json:"version"`
	Models  map[string]*metamodel.Model `json:"models"`
	Config  map[string]modelConfig      `json:"config"`
}

func getConfig(m *File) modelConfig {
	// TODO: read <file>.md from filesystem
	return modelConfig{m.Cid, "# " + m.Schema, string(m.Source)}
}

func ToJson() []byte {
	m := make(map[string]*metamodel.Model, 0)
	c := make(map[string]modelConfig, 0)
	for k, v := range Models {
		m[k] = v.Model
		c[k] = getConfig(v)
	}
	return codec.Marshal(modelCollection{
		Version: "0.1.0",
		Models:  m,
		Config:  c,
	})
}
