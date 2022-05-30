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
	IncludeDrafts  bool
	CPUProfileFile string
	VersionInfo    string
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
	genCmd.Flags().BoolVarP(&genCmdFlags.IncludeDrafts, "include-drafts", "", false, "Include draft posts")
	genCmd.Flags().StringVarP(&genCmdFlags.CPUProfileFile, "cpu-profile", "", "", "The location to output a CPU profile recording to")
	genCmd.Flags().StringVarP(&genCmdFlags.VersionInfo, "version-info", "", "", "Version info to pass to blog templates (i.e. git hash)")
}

func startGen(cmd *cobra.Command, args []string) {
	if genCmdFlags.OutputDir == "" {
		genCmdFlags.OutputDir = filepath.Join(rootCmdFlags.Dir, "public")
	}

	generateBlog(rootCmdFlags.Dir, &genCmdFlags)
}

func generateBlog(inDir string, flags *genFlags) {
	if genCmdFlags.CPUProfileFile != "" {
		file, err := os.Create(genCmdFlags.CPUProfileFile)
		if err != nil {
			log.Fatalf("pprof file creation error: %s", err)
		}
		defer file.Close()

		pprof.StartCPUProfile(file)
		defer pprof.StopCPUProfile()
	}

	log.Printf("generating blog %s", inDir)
	start := time.Now()

	var err error
	if cfg, err = config.Load(inDir); err != nil {
		log.Fatalf("config error: %s", err)
	}
	cfg.Blog.VersionInfo = flags.VersionInfo
	cfg.Blog.ExcludeDrafts = !flags.IncludeDrafts

	var logger *logger.Logger
	if rootCmdFlags.Verbose {
		logger = log
	}

	blog, err := blog.New(cfg.Blog, inDir, logger)
	if err != nil {
		log.Fatalf("blog init error: %s", err)
	}

	if err = blog.Generate(flags.OutputDir); err != nil {
		log.Fatalf("error generating blog: %s", err)
	}

	log.Printf("done! %dms", time.Since(start).Nanoseconds()/int64(time.Millisecond))
}
