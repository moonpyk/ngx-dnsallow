package config

import (
	"fmt"
	"os/exec"
)

type Nginx struct {
	Path string `yaml:"path"`
}

func (n *Nginx) EnsurePath() string {
	if len(n.Path) == 0 {
		return "/usr/sbin/nginx"
	}

	return n.Path
}

func (n *Nginx) ConfigTest() error {
	err := n.nginxExecute(
		"-t",
	)
	if err != nil {
		return err
	}

	return nil
}

func (n *Nginx) ConfigReload() error {
	err := n.nginxExecute(
		"-s",
		"reload",
	)
	if err != nil {
		return err
	}

	return nil
}

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
