package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

const pageTemplatePath = "./resources/home.tpl.html"

// myfiles is a very fancy data store as you can tell
var myfiles = [...]file{
	{
		AuthorEmail: "lander.visterin@klarrio.com",
		Name:        "Lander Visterin",
		Groups:      []string{"Admin"},
		Location:    "Belgium",
	},
}

func main() {
	ctx := context.Background()

	services, err := createServices(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	ro := httprouter.New()
	ro.GET("/", handleErrors(services.homeHandle))

	const addr = "127.0.0.1:8000"

	srv := http.Server{
		Addr:        addr,
		Handler:     ro,
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}

	log.Println("serving on", "http://"+addr)

	err = srv.ListenAndServe()
	log.Fatalln(err)
}

func (s services) homeHandle(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	t, err := template.ParseFiles(pageTemplatePath)
	if err != nil {
		return err
	}

	user := s.user.get()
	input := page{
		User:  struct2Map(user),
		Files: make([]displayFile, len(myfiles)),
	}

	json.NewEncoder(os.Stdout).Encode(&input)

	for i, fileToCheck := range myfiles {
		authz, err := s.policy.eval(r.Context(), policyRequest{
			Resource: fileToCheck,
			Subject:  user,
		})
		if err != nil {
			return fmt.Errorf("policy evaluation failed with unexpected error %w", err)
		}

		input.Files[i] = displayFile{
			File:  struct2Map(fileToCheck),
			Authz: authz,
		}
	}

	w.Header().Set("Content-Type", "text/html")
	return t.ExecuteTemplate(w, filepath.Base(pageTemplatePath), &input)
}
