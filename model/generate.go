package model

import (
	"bytes"
	"fmt"
	"github.com/pflow-dev/go-metamodel/metamodel"
	"github.com/pflow-dev/go-metamodel/metamodel/js"
	"github.com/pflow-dev/go-metamodel/metamodel/lua"
	"github.com/pflow-dev/pflow/codec"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Models cache loaded source files
var Models = make(map[string]*sourceFile, 0)

func GenerateGoModels(modelPath string) {
	files, err := ioutil.ReadDir(modelPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		ext := strings.Split(file.Name(), ".")
		fileExt := ext[len(ext)-1]
		fmt.Println(fileExt, file.Name())

		switch fileExt {
		case "lua":
			writeModelToSource(LoadLua(filepath.Join(modelPath, file.Name())))
		case "js":
			writeModelToSource(LoadJs(filepath.Join(modelPath, file.Name())))
		case "go":
			continue
		default:
			panic("unknown filetype: " + fileExt)
		}
	}
}

// add source to go codebase
func writeModelToSource(src *sourceFile) {
	ext := strings.Split(src.Path, ".")
	fileExt := ext[len(ext)-1]
	var outFile string

	switch fileExt {
	case "lua":
		outFile = strings.Replace(src.Path, ".lua", ".go", -1)
	case "js":
		outFile = strings.Replace(src.Path, ".js", ".go", -1)
	default:
		panic("unknown filetype: " + fileExt)
	}
	err := os.WriteFile(
		outFile,
		[]byte(serializedModelTemplate(modelTemplate{sourceFile: src})),
		0644,
	)
	if err != nil {
		panic(err)
	} else {
		log.Println("Wrote: " + outFile)
	}
}

type modelTemplate struct {
	*sourceFile
	Json    string
	Oid     string
	Imports string
}

const importHeader = `package model

`
const modelTpl = "{{.Imports}}\n// {{.Path}} {{.Oid}}\nfunc init(){\n\tModel.Load(`{{.Json}}`)\n}"

func serializedModelTemplate(m modelTemplate) string {
	data := codec.Marshal(m.Model)
	m.Json = string(data)
	m.Oid = codec.ToOid(data).String()
	m.Imports = importHeader

	tmpl, err := template.New("serializedModel").Parse(modelTpl)
	if err != nil {
		panic(err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, m)
	if err != nil {
		panic(err)
	}
	return b.String()
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

func LoadLua(path string) (src *sourceFile) {
	source, err := readFile(path)
	var m *metamodel.Model
	m, err = lua.LoadModel(string(source))
	if err != nil {
		panic(err)
	}
	name := filepath.Base(path)
	src = &sourceFile{
		Path:     path,
		Model:    m,
		Cid:      codec.ToOid(source).String(),
		Modified: time.Now(),
	}
	Models[name] = src
	log.Println("reloaded:", name, src.Cid)
	return src
}

func LoadJs(path string) (src *sourceFile) {
	source, err := readFile(path)
	var m *metamodel.Model
	m, err = js.LoadModel(string(source))
	if err != nil {
		panic(err)
	}
	name := filepath.Base(path)
	src = &sourceFile{
		Path:     path,
		Model:    m,
		Cid:      codec.ToOid(source).String(),
		Modified: time.Now(),
	}
	Models[name] = src
	log.Println("reloaded:", name, src.Cid)
	return src
}

type sourceFile struct {
	Path     string           `json:"path"`
	Modified time.Time        `json:"modified"`
	Cid      string           `json:"cid"`
	Model    *metamodel.Model `json:"model"`
}
