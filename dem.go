package main

import (
	"bytes"
	"github.com/codegangsta/cli"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
)

// Global Vars
var demDir string
var debug bool
var systemDir string = "/var/lib/docker"

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
			Name:  "use",
			Usage: "dem use [<imgset name>]",
			Action: func(c *cli.Context) {
				setGlobalVars(c)
				use(c.Args().First())
			},
		},
		{
			Name:  "list",
			Usage: "dem list",
			Action: func(c *cli.Context) {
				setGlobalVars(c)
				setDockerSystemPath()
				list()
			},
		},
	}

	app.Run(os.Args)
}

func setGlobalVars(c *cli.Context) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	demDir = c.GlobalString("demDir")
	debug = c.GlobalBool("debug")

	demSettings := []string{dir, ".dem"}
	if _, err := os.Stat(strings.Join(demSettings, "/")); os.IsNotExist(err) {
		Warn("\"/.dem\" is not found.")
		if demDir == "" {
			demDir = "/var/lib/dem"
		}
		cmd := exec.Command("echo", demDir, "|", "tee", strings.Join(demSettings, "/"))
		err = cmd.Run()
		if err != nil {
			die("Error: %s.", err, 1)
		}
	} else {
		cmd := exec.Command("cat", strings.Join(demSettings, "/"))
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			die("Error: %s.", err, 2)
		}
		if err = cmd.Start(); err != nil {
			die("Error: %s.", err, 2)
		}
		outC := make(chan string)
		// copy the output in a separate goroutine so printing can't block indefinitely
		go func() {
			var buf bytes.Buffer
			io.Copy(&buf, stdout)
			outC <- buf.String()
		}()
		temp_demDir := filepath.Base(<-outC)
		demDir = strings.TrimSpace(temp_demDir)
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
	if currentImgset := getCurrentImgset(); currentImgset == imgset {
		die("Cannot remove current imgset.", nil, 127)
	}
	imgsetPath := getImgsetPath(imgset)
	if _, err := os.Stat(imgsetPath); os.IsNotExist(err) {
		Warn("%s is not installed.", imgset)
		return
	}
	err := os.RemoveAll(imgsetPath)
	if err != nil {
		die("Unable to remove Imgset %s located in %s.", err, 2, imgset, imgsetPath)
	}
	Success("Remove Imgset %s successfully.", imgset)
}

func use(imgset string) {
	if imgset == "" {
		die("The use command requires that a imgset name is specified.", nil, 127)
	}
	imgsetPath := getImgsetPath(imgset)
	ensureImgsetCreated(imgsetPath)
	reset(getDockerDefault())
	installSetting(imgsetPath)
	makeItCurrent(imgsetPath)

	Info("Now using imgset %s", imgset)
}

func list() {
	results := getInstalledImgset()
	current := getCurrentImgset()

	for _, result := range results {
		if current == result {
			Success("->\t%s", result)
		} else {
			if result != "current" {
				Info("\t%s", result)
			}
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

func getCurrentImgset() string {
	currentPath := getImgsetPath("current")
	cmd := exec.Command("readlink", currentPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		die("Error: %s.", err, 2)
	}
	if err := cmd.Start(); err != nil {
		die("Error: %s.", err, 2)
	}
	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, stdout)
		outC <- buf.String()
	}()
	current := filepath.Base(<-outC)
	current = strings.TrimSpace(current)

	Debug("current imgset: %s", current)
	return current
}

func changeDockerDefault(setting, imgsetPath string) {
	set(setting)
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

func set(setting string) {
	backupName := []string{setting, ".bak"}
	cmd := exec.Command("cp", setting, strings.Join(backupName, ""))
	err := cmd.Run()
	if err != nil {
		die("Error: %s.", err, 1)
	}
	Debug("Copying %s to %s", setting, strings.Join(backupName, ""))
}

func reset(setting string) {
	backupName := []string{setting, ".bak"}
	cmd := exec.Command("mv", "-f", strings.Join(backupName, ""), setting)
	err := cmd.Run()
	if err != nil {
		die("Error: %s.", err, 1)
	}
	Debug("Resstore %s to %s", strings.Join(backupName, ""), setting)
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

func makeItCurrent(imgsetPath string) {
	currentPath := getImgsetPath("current")
	cmd := exec.Command("ln", "-sfn", imgsetPath, currentPath)
	err := cmd.Run()
	if err != nil {
		die("Error: %s.", err, 2)
	}
	Debug("Making %s to current", imgsetPath)
}

func getDockerDefault() string {
	return "/etc/default/docker"
}

func setDockerSystemPath() {
	systemPath := getImgsetPath("system")
	cmd := exec.Command("ln", "-sfn", systemDir, systemPath)
	err := cmd.Run()
	if err != nil {
		die("Error: %s.", err, 2)
	}
	Debug("Making %s to system", systemDir)
}
