package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/open-policy-agent/opa/compile"
)

const (
	basePath   = "./tools/pap/policy"
	regoFile   = "policy.rego"
	tarFile    = "policy.tar.gz"
	wasmFile   = "policy.wasm"
	entrypoint = "policy/authz"
)

func main() {
	ctx := context.Background()

	err := rebuild(ctx)
	if err != nil {
		log.Fatalln("initial build failed with", err)
	}

	errs := make(chan error)
	go func() {
		// serve OPA compatible bundle API
		const addr = "127.0.0.1:3000"
		log.Println("serving on", addr)
		errs <- http.ListenAndServe(addr, http.FileServer(http.Dir(basePath)))
	}()

	go func() {
		errs <- watch(ctx)
	}()

	log.Fatalln(<-errs) // exit on first error
}

// watch for rego changes and rebuild bundle
func watch(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Has(fsnotify.Write) && filepath.Ext(event.Name) == ".rego" {
					log.Println("policy modified:", event.Name)
					log.Println("rebuild returned:", rebuild(ctx))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("watch error:", err)
			}
		}
	}()

	err = watcher.Add(basePath)
	if err != nil {
		return err
	}

	log.Println("watching for events on", basePath)

	<-ctx.Done()
	return ctx.Err()
}

func rebuild(ctx context.Context) error {
	err := rebuildTar(ctx)
	if err != nil {
		return fmt.Errorf("rebuilding tar: %w", err)
	}

	err = rebuildWasm(ctx)
	if err != nil {
		return fmt.Errorf("rebuilding wasm: %w", err)
	}

	log.Println("rebuilt")

	return nil
}

func rebuildTar(ctx context.Context) error {
	f, err := os.Create(filepath.Join(basePath, tarFile))
	if err != nil {
		return err
	}
	defer f.Close()

	return compile.New().
		WithPaths(basePath).
		WithAsBundle(true).
		WithEntrypoints(entrypoint).
		WithTarget(compile.TargetRego).
		WithOutput(f).
		Build(ctx)
}

func rebuildWasm(ctx context.Context) error {
	f, err := os.Create(filepath.Join(basePath, wasmFile))
	if err != nil {
		return err
	}
	defer f.Close()

	return compile.New().
		WithPaths(basePath).
		WithAsBundle(true).
		WithEntrypoints(entrypoint).
		WithTarget(compile.TargetWasm).
		WithOutput(f).
		Build(ctx)
}
