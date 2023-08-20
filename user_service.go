package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

// userService provides user info for the currently "logged in" user.
// We don't really have "users" in this demo though and neither do we have login.
// So it's basically just a thing that provides a set of attributes from a live file
type userService struct {
	mu   sync.Mutex
	user user
}

const (
	userInfoPath = "./data/userinfo.yaml"
)

func newUserService(ctx context.Context) (*userService, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	defer watcher.Close()

	u := &userService{}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Has(fsnotify.Write) && filepath.Ext(event.Name) == ".yaml" {
					log.Println("user reload returned:", u.reload())
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("watch error:", err)
			}
		}
	}()

	err = watcher.Add(userInfoPath)
	if err != nil {
		return nil, err
	}

	log.Println("watching for events on", userInfoPath)
	return u, u.reload()
}

func (u *userService) reload() error {
	f, err := os.Open(userInfoPath)
	if err != nil {
		return err
	}
	defer f.Close()

	u.mu.Lock()
	defer u.mu.Unlock()

	y := yaml.NewDecoder(f)
	y.SetStrict(true)
	u.user = user{}
	return y.Decode(&u.user)
}

func (u *userService) get() user {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.user
}
