package main

import (
	"github.com/codegangsta/cli"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Global Vars
var demDir string
var debug bool

func main() {
	app := cli.NewApp()
	app.Name = "dem"
	app.Usage = "Docker Environment Management"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "demDir", EnvVar: "DEM_DIR", Usage: "Specify an alternate DEM home directory. Defaults to the \"/var/lib/dem\" directory."},
		cli.BoolFlag{Name: "debug", Usage: "Print debugging infos."},
	}
	app.Commands = []cli.Command{
		{
			Name:  "create",
			Usage: "dem create [<imgset name>]",
			Action: func(c *cli.Context) {
				setGlobalVars(c)
				create(c.Args().First())
			},
		},
		{
			Name:  "remove",
			Usage: "dem remove [<imgset name>]",
			Action: func(c *cli.Context) {
				setGlobalVars(c)
				remove(c.Args().First())
			},
		},
		{
			Name:  "list",
			Usage: "dem list",
			Action: func(c *cli.Context) {
				setGlobalVars(c)
				list()
			},
		},
	}

	app.Run(os.Args)
}

func setGlobalVars(c *cli.Context) {
	demDir = c.GlobalString("demDir")
	debug = c.GlobalBool("debug")

	if demDir == "" {
		demDir = "/var/lib/dem"
	}
}

func create(imgset string) {
	if imgset == "" {
		die("The create command requires that a imgset name is specified.", nil, 127)
	}
	imgsetPath := getImgsetPath(imgset)
	err := os.MkdirAll(imgsetPath, 0700)
	if err != nil {
		die("Unable to create to %s.", err, 127)
	}
	imgset = filepath.Base(imgsetPath)
	Success("Imgset %s create successfully", imgset)
}

func remove(imgset string) {
	if imgset == "" {
		die("The remove command requires that a imgset name is specified.", nil, 127)
	}
}

func use(imgset string) {
	if imgset == "" {
		die("The use command requires that a imgset name is specified.", nil, 127)
	}
	imgsetPath := getImgsetPath(imgset)
	ensureImgsetCreated(imgsetPath)
	reset()
	installSetting(imgsetPath)

	Info("Now using imgset %s", imgset)
}

func list() {
	results := getInstalledImgset()
	current := "test"

	for _, result := range results {
		if current == result {
			Success("->\t%s", result)
		} else {
			Info("\t%s", result)
		}
	}
}

func getInstalledImgset() []string {
	imgsets, _ := filepath.Glob(getImgsetPath("*"))

	var results []string
	for _, imgsetDir := range imgsets {
		imgset := filepath.Base(imgsetDir)

		results = append(results, imgset)
	}

	sort.Strings(results)
	return results
}

func changeDockerDefault(setting, imgsetPath string) {
	fs := []string{"s@#DOCKER_OPTS=\"@DOCKER_OPTS=\"-g", imgsetPath, "@"}

	cmd := exec.Command("sed", "-i", strings.Join(fs, "\\ "), setting)
	err := cmd.Run()
	if err != nil {
		die("Error: %s.", err, 1)
	}
	Debug("Modifying %s", setting)
}

func getImgsetPath(imgset string) string {
	imgsetPath := filepath.Join(demDir, imgset)
	imgsetPath = filepath.ToSlash(imgsetPath)
	return imgsetPath
}

func ensureImgsetCreated(imgsetPath string) {
	err := os.MkdirAll(imgsetPath, 0700)
	if err != nil {
		die("Unable to create to %s.", err, 1)
	}
}

func reset() {
	Debug("here is reset function")
}

func installSetting(imgsetPath string) {
	dockerDefault := getDockerDefault()
	changeDockerDefault(dockerDefault, imgsetPath)

	cmd := exec.Command("service", "docker", "restart")
	err := cmd.Run()
	if err != nil {
		die("Error: %s.", err, 2)
	}
	Debug("Restarting docker service")
}

func getDockerDefault() string {
	return "/etc/default/docker"
}
