package commands

import (
	logger "log"
	"path/filepath"
	"time"

	"github.com/alexbakker/blogen/blog"
	"github.com/alexbakker/blogen/config"
	"github.com/spf13/cobra"
)

type genFlags struct {
	OutputDir string
}

var (
	genCmdFlags genFlags
	genCmd      = &cobra.Command{
		Use:   "gen",
		Short: "Generate the blog",
		Run:   startGen,
	}

	cfg *config.Config
)

func init() {
	RootCmd.AddCommand(genCmd)
	genCmd.Flags().StringVarP(&genCmdFlags.OutputDir, "output", "o", "", "The output directory")
}

func startGen(cmd *cobra.Command, args []string) {
	log.Printf("generating blog %s", rootCmdFlags.Dir)
	start := time.Now()

	var err error
	if cfg, err = config.Load(rootCmdFlags.Dir); err != nil {
		log.Fatalf("config error: %s", err)
	}

	var logger *logger.Logger
	if rootCmdFlags.Verbose {
		logger = log
	}

	blog, err := blog.New(cfg.Blog, rootCmdFlags.Dir, logger)
	if err != nil {
		log.Fatalf("blog init error: %s", err)
	}

	if genCmdFlags.OutputDir == "" {
		genCmdFlags.OutputDir = filepath.Join(rootCmdFlags.Dir, "public")
	}
	if err = blog.Generate(genCmdFlags.OutputDir); err != nil {
		log.Fatalf("error generating blog: %s", err)
	}

	log.Printf("done! %dms", time.Since(start).Nanoseconds()/int64(time.Millisecond))
}
