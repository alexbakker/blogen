package commands

import (
	logger "log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	"github.com/alexbakker/blogen/blog"
	"github.com/alexbakker/blogen/config"
	"github.com/spf13/cobra"
)

type genFlags struct {
	OutputDir      string
	CPUProfileFile string
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
	genCmd.Flags().StringVarP(&genCmdFlags.CPUProfileFile, "cpu-profile", "", "", "The location to output a CPU profile recording to")
}

func startGen(cmd *cobra.Command, args []string) {
	if genCmdFlags.CPUProfileFile != "" {
		file, err := os.Create(genCmdFlags.CPUProfileFile)
		if err != nil {
			log.Fatalf("pprof file creation error: %s", err)
		}
		defer file.Close()

		pprof.StartCPUProfile(file)
		defer pprof.StopCPUProfile()
	}

	if genCmdFlags.OutputDir == "" {
		genCmdFlags.OutputDir = filepath.Join(rootCmdFlags.Dir, "public")
	}

	generateBlog(rootCmdFlags.Dir, genCmdFlags.OutputDir, true)
}

func generateBlog(inDir string, outDir string, excludeDrafts bool) {
	log.Printf("generating blog %s", inDir)
	start := time.Now()

	var err error
	if cfg, err = config.Load(inDir); err != nil {
		log.Fatalf("config error: %s", err)
	}
	cfg.Blog.ExcludeDrafts = excludeDrafts

	var logger *logger.Logger
	if rootCmdFlags.Verbose {
		logger = log
	}

	blog, err := blog.New(cfg.Blog, inDir, logger)
	if err != nil {
		log.Fatalf("blog init error: %s", err)
	}

	if err = blog.Generate(outDir); err != nil {
		log.Fatalf("error generating blog: %s", err)
	}

	log.Printf("done! %dms", time.Since(start).Nanoseconds()/int64(time.Millisecond))
}
