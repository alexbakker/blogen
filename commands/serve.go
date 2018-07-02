package commands

import (
	"net/http"

	"github.com/alexbakker/blogen/server"
	"github.com/spf13/cobra"
)

type serveFlags struct {
	Addr string
}

var (
	serveCmdFlags serveFlags
	serveCmd      = &cobra.Command{
		Use:   "serve",
		Short: "Serve the site over HTTP on the specified address and port",
		Run:   startServe,
	}
)

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&serveCmdFlags.Addr, "addr", "a", "127.0.0.1:8081", "The TCP port to listen on")
}

func startServe(cmd *cobra.Command, args []string) {
	server := server.New(server.Config{Addr: serveCmdFlags.Addr}, rootCmdFlags.Dir)

	log.Printf("starting http server on %s", serveCmdFlags.Addr)
	log.Fatal(http.ListenAndServe(serveCmdFlags.Addr, server))
}
