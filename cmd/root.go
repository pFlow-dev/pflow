package cmd

import (
	"fmt"
	"github.com/pflow-dev/pflow/model/source"
	"github.com/pflow-dev/pflow/service"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"
)

var (
	outputJson bool
	outputSvg  bool
	outputHtml bool

	rootCmd = &cobra.Command{
		Use:   "pflow",
		Short: "simulate petri-net models in the browser",
		Long: `Supports model declaration using Lua or Javascript dialects.
This app will auto-reload models in the web UI whenever watched files change.`,
		Run: Run,
	}
)

func init() {
	rootCmd.ArgAliases = []string{"directory or filepath.[js|lua]"}
	rootCmd.PersistentFlags().BoolVar(&outputJson, "json", false, "output models.json")
	rootCmd.PersistentFlags().BoolVar(&outputSvg, "svg", false, "output image.svg")
	rootCmd.PersistentFlags().BoolVar(&outputHtml, "html", false, "output ./dist")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func Generate(format string) {
	source.LoadModels(service.ModelPath)
	if format == "html" {
		writeHtml()
	} else {
		source.WriteModels(format)
	}
}

func writeHtml() {
	const baseDir = "./dist"
	os.RemoveAll(baseDir)
	os.Mkdir(baseDir, 0777)
	models := path.Join(baseDir, "models.json")
	data := source.ToJson()
	err := os.WriteFile(models, data, 0664)
	if err != nil {
		panic(err)
	}

	os.Mkdir(baseDir, 0777)

	for _, m := range source.Models {
		previewImage, _ := source.ToSvg(m.Cid)

		os.Mkdir(path.Join(baseDir, m.Cid), 0777)
		os.WriteFile(path.Join(baseDir, m.Cid, "image.svg"), previewImage, 0664)
	}

	service.Box.Walk(".", func(filepath string, info fs.FileInfo, err error) error {
		fmt.Println(filepath)
		if info.IsDir() {
			os.Mkdir(path.Join(baseDir, filepath), 0777)
		} else {
			fileData, fileErr := service.Box.Bytes(filepath)
			if fileErr != nil {
				panic(fileErr)
			}
			os.WriteFile(path.Join(baseDir, filepath), fileData, 0664)
		}
		return nil
	})
	//fmt.Printf("wrote %s", data)
	fmt.Printf("wrote %v", baseDir)
}

// open opens the specified URL in the default browser of the user.
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

const url = "http://localhost:8080/p/"

func Serve() {
	source.LoadModels(service.ModelPath)
	m := source.GetFirstModel()
	for {
		if m != nil {
			openBrowser(url + "?run=" + m.Cid)
			break
		} else {
			fmt.Print("no-models sleeping\n")
			time.Sleep(time.Second)
			source.LoadModels(service.ModelPath)
			m = source.GetFirstModel()
		}
	}
	service.Webserver()
}

func Run(cmd *cobra.Command, args []string) {
	_ = cmd
	if len(args) != 1 {
		panic("must provide path to directory or filename")
	} else {
		p, err := filepath.Abs(args[0])
		if err != nil {
			panic(err)
		}
		service.ModelPath = p
		if outputJson {
			Generate("json")
		} else if outputSvg {
			Generate("svg")
		} else if outputHtml {
			Generate("html")
		} else {
			_, err := os.Stat(p)
			if err != nil {
				panic(err)
			}
			Serve()
		}
	}
}
