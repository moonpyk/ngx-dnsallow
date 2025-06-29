package model

import (
	"fmt"
	"os/exec"
)

type Nginx struct {
	// Path is the configured path to nginx binary
	// if empty, the app will generate a non-empty path before trying to use it
	Path string `yaml:"path"`
}

// EnsurePath ensure a non-empty path is configured
func (n *Nginx) EnsurePath() string {
	if len(n.Path) == 0 {
		return "/usr/sbin/nginx"
	}

	return n.Path
}

// ConfigTest spawns nginx with '-t' checking for any configuration error
func (n *Nginx) ConfigTest() error {
	return n.nginxExecute(
		"-t",
	)
}

// ConfigReload spawns nginx with '-s reload' trying to reload nginx configuration
func (n *Nginx) ConfigReload() error {
	return n.nginxExecute(
		"-s",
		"reload",
	)
}

// nginxExecute is a utility helper to spawn nginx process and checking returned status code
func (n *Nginx) nginxExecute(args ...string) error {
	command := exec.Command(n.EnsurePath(), args...)

	err := command.Start()
	if err != nil {
		return err
	}

	err = command.Wait()
	if err != nil {
		return err
	}

	code := command.ProcessState.ExitCode()

	if code != 0 {
		return fmt.Errorf("nginx exited with %d", code)
	}

	return nil
}
