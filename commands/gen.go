package commands

import (
	"path"
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
		Short: "Generate the site",
		Run:   startGen,
	}

	cfg *config.Config
)

func init() {
	RootCmd.AddCommand(genCmd)
	genCmd.Flags().StringVarP(&genCmdFlags.OutputDir, "output", "o", "", "The output directory")
}

func startGen(cmd *cobra.Command, args []string) {
	start := time.Now()

	var err error
	if cfg, err = config.Load(rootCmdFlags.Dir); err != nil {
		log.Fatalf("config error: %s", err)
	}

	blog, err := blog.New(cfg.Blog, rootCmdFlags.Dir)
	if err != nil {
		log.Fatalf("site error: %s", err)
	}

	if genCmdFlags.OutputDir == "" {
		genCmdFlags.OutputDir = path.Join(rootCmdFlags.Dir, "public")
	}
	if err = blog.Generate(genCmdFlags.OutputDir); err != nil {
		log.Fatalf("error generating site: %s", err)
	}

	log.Printf("done! %dms", time.Since(start).Nanoseconds()/int64(time.Millisecond))
}
