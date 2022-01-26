package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/racklet/render-drawio-action/pkg/render"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("error occurred: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Setup default config
	cfg := &render.Config{
		RootDir:  "/files",
		SubDirs:  []string{"."},
		SkipDirs: []string{".git"},

		Files: map[string]string{},

		SrcFormats:       []string{"drawio"},
		ValidSrcFormats:  []string{"drawio", "*"}, // "*" means that Files' src-file can have any extension
		DestFormats:      []string{"svg"},
		ValidDestFormats: []string{"pdf", "png", "jpg", "svg"},
	}

	// Unconditionally overwrite root path if GH Action workspace is set
	if ghWorkspace := os.Getenv("GITHUB_WORKSPACE"); len(ghWorkspace) != 0 {
		cfg.RootDir = ghWorkspace
	}

	// Validate and complete the config with info from the environment
	if err := cfg.Complete(render.DefaultFlags); err != nil {
		return err
	}

	log := zap.S()

	// Refer to https://github.com/rlespinasse/docker-drawio-desktop-headless/blob/v1.x/scripts/entrypoint.sh
	// and https://github.com/rlespinasse/docker-drawio-desktop-headless/blob/v1.x/scripts/runner.sh to know
	// why things are done in this way.
	displayEnv := os.Getenv("XVFB_DISPLAY")
	if err := os.Setenv("DISPLAY", displayEnv); err != nil {
		return err
	}

	go func() {
		ctx := context.Background()
		output, _, err := render.ShellCommand(ctx, "Xvfb %q %s", displayEnv, os.Getenv("XVFB_OPTIONS")).Run()
		if err != nil && err != context.Canceled {
			log.Errorf("Error executing Xvfb: %v, %q", err, output)
		}
	}()

	// Just make sure Xvfb has a bit of time to start up. We anyway can't get a "ready signal" from it,
	// it just runs in the background, so assume that a second is enough.
	time.Sleep(1 * time.Second)

	// Make outputFiles as large as the set of input files
	outputFiles := make([]string, 0, len(cfg.Files))

	// Render the files using the drawio CLI
	if err := cfg.Render(func(src, dest string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		out, _, err := render.ShellCommand(ctx, "%q -x -t %q -o %q --no-sandbox",
			os.Getenv("DRAWIO_DESKTOP_RUNNER_COMMAND_LINE"), src, dest).Run()
		cancel()
		if err != nil {
			return fmt.Errorf("failed to run drawio for src=%q and dest=%q: %v, output: %s", src, dest, err, out)
		}

		// The output file does not include the root directory prefix
		outputFiles = append(outputFiles, dest)

		return nil
	}); err != nil {
		return err
	}

	// Set the GH Action output
	return render.GitHubActionSetFilesOutput("rendered-files", cfg.RootDir, outputFiles)
}
