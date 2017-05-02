package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aisola/log"
	"github.com/spf13/viper"
)

type App struct {
	Repo    string
	Service string
	User    string
	Name    string
}

func NewApp(repo string) *App {
	app := &App{Repo: repo}
	parts := strings.Split(repo, "/")
	app.Service = parts[0]
	app.User = parts[1]
	app.Name = parts[2]
	return app
}

func (a *App) CloneURL() string {
	return fmt.Sprintf("git@%s:%s/%s.git", a.Service, a.User, a.Name)
}

func (a *App) PublicDir() string {
	public, err := filepath.Abs(viper.GetString("core.public"))
	if err != nil {
		panic(err)
	}
	return public
}

func (a *App) RepoDir() string {
	repo, err := filepath.Abs(viper.GetString("core.repo"))
	if err != nil {
		panic(err)
	}
	return repo
}

func (a *App) SourceDir() string {
	source, err := filepath.Abs(viper.GetString("core.source"))
	if err != nil {
		panic(err)
	}
	return source
}

func (a *App) compile() error {
	log.Infof("Compiling %s/%s", a.User, a.Name)

	_, err := os.Stat(a.SourceDir())
	if !os.IsNotExist(err) {
		if err := os.RemoveAll(a.SourceDir()); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(a.SourceDir(), 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "--git-dir="+a.RepoDir(), "--work-tree="+a.SourceDir(), "checkout", "-q", "-f", "master")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "clone", viper.GetString("hugo.theme_url"),  filepath.Join(a.SourceDir(), "themes", viper.GetString("hugo.theme")))
	cmd.Dir = a.SourceDir()
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("hugo", "-d="+a.PublicDir(), "-t="+viper.GetString("hugo.theme"))
	cmd.Dir = a.SourceDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (a *App) init() error {
	log.Infof("Initializing %s/%s", a.User, a.Name)
	log.Debugf("Repo: %s", a.RepoDir())
	log.Debugf("Source: %s", a.SourceDir())
	log.Debugf("Public: %s", a.PublicDir())

	_, err := os.Stat(a.RepoDir())
	if err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(a.RepoDir(), 0755); err != nil {
			return err
		}

		cmd := exec.Command("git", "--git-dir="+a.RepoDir(), "init")
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}

		cmd = exec.Command("git", "--git-dir="+a.RepoDir(), "remote", "add", "origin", a.CloneURL())
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (a *App) fetch() error {
	log.Infof("Fetching %s/%s", a.User, a.Name)
	cmd := exec.Command("git", "--git-dir="+a.RepoDir(), "fetch", "-f", "origin", "master:master")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (a *App) Update() error {
	if err := a.init(); err != nil {
		return err
	}

	if err := a.fetch(); err != nil {
		return err
	}

	return a.compile()
}
