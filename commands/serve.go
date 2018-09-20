package commands

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/alexbakker/blogen/server"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

type serveFlags struct {
	Addr      string
	OutputDir string
}

var (
	serveCmdFlags serveFlags
	serveCmd      = &cobra.Command{
		Use:   "serve",
		Short: "Serve the blog over HTTP on the specified address and port",
		Run:   startServe,
	}
)

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&serveCmdFlags.Addr, "addr", "a", "127.0.0.1:8080", "The TCP port to listen on")
	serveCmd.Flags().StringVarP(&serveCmdFlags.OutputDir, "output", "o", "", "The output directory")
}

func startServe(cmd *cobra.Command, args []string) {
	var err error
	if serveCmdFlags.OutputDir == "" {
		serveCmdFlags.OutputDir, err = ioutil.TempDir(os.TempDir(), "blogen-")
		if err != nil {
			log.Fatalf("error creating tmp dir: %s", err)
		}
		defer os.RemoveAll(serveCmdFlags.OutputDir)
	}

	generateBlog(rootCmdFlags.Dir, serveCmdFlags.OutputDir)

	// start HTTP server
	server := server.New(server.Config{Addr: serveCmdFlags.Addr}, serveCmdFlags.OutputDir)
	go func() {
		log.Printf("starting http server on %s", serveCmdFlags.Addr)
		log.Fatal(http.ListenAndServe(serveCmdFlags.Addr, server))
	}()

	// watch for changes to the source directory
	go func() {
		for {
			if err := watchBlog(); err != nil {
				log.Printf("watcher error: %s", err)
				break
			}
		}
	}()

	// wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
}

func watchBlog() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan error)
	go func() {
		for {
			select {
			case _, ok := <-watcher.Events:
				if ok {
					generateBlog(rootCmdFlags.Dir, serveCmdFlags.OutputDir)
				}
				done <- nil
				return
			case err, ok := <-watcher.Errors:
				if !ok {
					done <- nil
					return
				}
				log.Printf("watcher error: %s", err)
			}
		}
	}()

	// recursively add directories to the watcher
	err = filepath.Walk(rootCmdFlags.Dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// TODO: ignore git directory
			if err = watcher.Add(path); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return <-done
}
