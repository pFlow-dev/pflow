package cmd

import (
	"fmt"
	"github.com/pflow-dev/pflow/service"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var (
	outputJson bool

	rootCmd = &cobra.Command{
		Use:   "pflow",
		Short: "simulate petri-net models in the browser",
		Long: `Supports model declaration using Lua or Javascript dialects.
This app will auto-reloading models in a web GUI while iterating on design.`,
		Run: Run,
	}
)

func init() {
	rootCmd.ArgAliases = []string{"directory or filepath.[js|lua]"}
	rootCmd.PersistentFlags().BoolVar(&outputJson, "json", false, "output models.json")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func Generate(format string) {
	service.LoadModels(service.ModelPath)
	service.WriteModels(format)

}

func Serve() {
	open("http://localhost:8080") // TODO: make port a param
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
		} else {
			stat, err := os.Stat(p)
			if err != nil {
				panic(err)
			}
			if stat.IsDir() {
				fmt.Print("expected path to a js or lua file")
			} else {
				Serve()
			}
		}
	}
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
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
