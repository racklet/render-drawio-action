package render

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/scylladb/go-set/strset"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func init() {
	// Setup logger. It seems that as logLevel is a pointer it works to write to it in a bit later stage
	log := zap.NewExample().WithOptions(zap.IncreaseLevel(logLevel))
	defer func() { _ = log.Sync() }()
	zap.ReplaceGlobals(log)
}

var (
	// register the log-level flag natively to the stdlib flag package
	logLevel = zap.LevelFlag("log-level", zap.InfoLevel, "What log level to use")
)

// RenderFunc renders the source file to the destination file. Both paths are absolute.
type RenderFunc func(src, dest string) error

type Config struct {
	// Absolute path of root directory, all other paths are relative
	RootDir string
	// Sub-directories of root-dir to recursively search for SrcFormats files
	SubDirs []string
	// Directories within SubDirs to not traverse into
	SkipDirs []string

	// Mapping between destination/target files and their respective sources. dest-file must end with a valid ext
	Files map[string]string

	// What formats source files have, and the list of valid options. File extensions without the dot. Lower-case.
	SrcFormats      []string
	ValidSrcFormats []string
	// What formats destination files have, and the list of valid options. File extensions without the dot. Lower-case.
	DestFormats      []string
	ValidDestFormats []string
}

func (c *Config) Validate() error {
	// Validate RootDir
	if !filepath.IsAbs(c.RootDir) {
		return fmt.Errorf("RootDir must be an absolute path: %q", c.RootDir)
	}
	if !isDir(c.RootDir) {
		return fmt.Errorf("RootDir must be a directory: %q", c.RootDir)
	}

	validDestFormats := strset.New(c.ValidDestFormats...)
	validSrcFormats := strset.New(c.ValidSrcFormats...)

	// Validate the DestFormats slice
	for i := range c.DestFormats {
		// Enforce lower-case
		c.DestFormats[i] = strings.ToLower(c.DestFormats[i])
		if !validDestFormats.Has(c.DestFormats[i]) {
			return fmt.Errorf("DestFormats[%d] %q is not valid: %s", i, c.DestFormats[i], validDestFormats.String())
		}
	}
	// Validate the SrcFormats slice
	for i := range c.SrcFormats {
		// Enforce lower-case
		c.SrcFormats[i] = strings.ToLower(c.SrcFormats[i])
		if !validSrcFormats.Has(c.SrcFormats[i]) {
			return fmt.Errorf("SrcFormats[%d] %q is not valid: %s", i, c.SrcFormats[i], validSrcFormats.String())
		}
	}

	// Validate the SubDirs slice
	for _, subDir := range c.SubDirs {
		if !fs.ValidPath(subDir) {
			return fmt.Errorf("SubDirs contains invalid path: %q", subDir)
		}
		if filepath.IsAbs(subDir) {
			return fmt.Errorf("SubDirs item must be relative: %q", subDir)
		}
		subDirAbs := filepath.Join(c.RootDir, subDir)
		if !isDir(subDirAbs) {
			return fmt.Errorf("SubDirs item must exist: %q", subDirAbs)
		}
	}

	// Validate the Files slice
	for dest, src := range c.Files {
		if len(dest) == 0 || len(src) == 0 {
			delete(c.Files, dest)
			continue
		}
		if !fs.ValidPath(dest) {
			return fmt.Errorf("Files contains invalid dest-path: %q", dest)
		}
		if !fs.ValidPath(src) {
			return fmt.Errorf("Files contains invalid src-path: %q", src)
		}

		if filepath.IsAbs(dest) {
			return fmt.Errorf("Files dest-path item must be relative: %q", dest)
		}
		if filepath.IsAbs(src) {
			return fmt.Errorf("Files src-path item must be relative: %q", src)
		}

		if !validDestFormats.Has(ExtToFormat(filepath.Ext(dest))) && !setAllowsAny(validDestFormats) {
			return fmt.Errorf("Files dest-path %q doesn't have valid format: %s", dest, validDestFormats.String())
		}
		if !validSrcFormats.Has(ExtToFormat(filepath.Ext(src))) && !setAllowsAny(validSrcFormats) {
			return fmt.Errorf("Files src-path %q doesn't have valid format: %s", dest, validSrcFormats.String())
		}

		srcAbs := filepath.Join(c.RootDir, src)
		if !isFile(srcAbs) {
			return fmt.Errorf("Files src-path item must exist: %q", srcAbs)
		}
	}

	return nil
}

// Complete uses information from the environment to complete the config
func (c *Config) Complete(registerFlags func(*Config)) error {
	log := zap.S()
	log.Debug("Parsing flags")
	// Register the --log-level flag from the stdlib flag
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// Register and parse flags
	registerFlags(c)
	pflag.Parse()

	// Our config should now be fully populated
	log.Infow("Got config", "cfg", c)

	// Validate the config
	log.Debug("Validating config")
	if err := c.Validate(); err != nil {
		return err
	}

	osfs := os.DirFS(c.RootDir)
	log.Infof("Created os.DirFS at %s", c.RootDir)

	skipDirs := strset.New(c.SkipDirs...)
	srcFormats := strset.New(c.SrcFormats...)

	for _, subDir := range c.SubDirs {
		log.Infof("Walking subDir %s", subDir)
		_ = fs.WalkDir(osfs, subDir, func(src string, d fs.DirEntry, err error) error {
			log.Debugf("Visiting path %s", src)

			// Fast path: if the file extension is not right, just continue the recursive search
			srcExt := filepath.Ext(src)
			if !d.IsDir() && !srcFormats.Has(ExtToFormat(srcExt)) {
				log.Debugf("File %s not matched, doesn't have any of extensions %s", src, srcFormats)
				return nil
			}
			// Ignore but log traversing errors
			if err != nil {
				log.Errorf("Observed error %v for path %s, but continuing...", err, src)
				return nil
			}
			// Skip excluded directories (note that the full path is matched)
			// TODO: Maybe we should do a filepath.Clean before?
			if d.IsDir() {
				if skipDirs.Has(src) {
					log.Debugf("Directory %s skipped", d.Name())
					return fs.SkipDir
				}
				// Don't process normal directories
				return nil
			}

			// If the file extension is right, create a dest file for each format
			for _, format := range c.DestFormats {
				// Replace the source file extension with the given dest format extension
				dest := strings.TrimSuffix(src, srcExt) + FormatToExt(format)
				c.Files[dest] = src
			}
			return nil
		})
	}

	return nil
}

func (c *Config) Render(fn RenderFunc) error {
	log := zap.S()

	// Exit if there's nothing to do
	if len(c.Files) == 0 {
		log.Info("Found no files to process")
		return nil
	}

	for dest, src := range c.Files {
		log.Infof("Rendering %s -> %s", src, dest)
		absSrc := filepath.Join(c.RootDir, src)
		absDest := filepath.Join(c.RootDir, dest)

		// Run the actual rendering
		if err := fn(absSrc, absDest); err != nil {
			return err
		}
	}

	return nil
}

func DefaultFlags(cfg *Config) {
	pflag.StringVarP(&cfg.RootDir, "root-dir", "r", cfg.RootDir, "Where the root directory for the files that should be rendered are.")
	pflag.StringSliceVarP(&cfg.SubDirs, "sub-dirs", "d", cfg.SubDirs, "Comma-separated list of sub-directories of --root-dir to recursively search for files to render")
	pflag.StringSliceVarP(&cfg.SkipDirs, "skip-dirs", "s", cfg.SkipDirs, "Comma-separated list of sub-directories of --root-dir to skip when recursively checking for files to convert")

	pflag.StringToStringVarP(&cfg.Files, "files", "f", cfg.Files, fmt.Sprintf("Comma-separated list of files to render, of form 'dest-file=src-file'. The extension for src-file can be any of %s, and for dest-file any of %s", cfg.ValidSrcFormats, cfg.ValidDestFormats))

	pflag.StringSliceVar(&cfg.DestFormats, "formats", cfg.DestFormats, "Comma-separated list of formats to render the files as, for use with --subdirs")
}

func FormatToExt(format string) string { return "." + format }
func ExtToFormat(ext string) string    { return strings.TrimPrefix(ext, ".") }

func isFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func setAllowsAny(set *strset.Set) bool { return set.Has("*") }

func GitHubActionSetOutput(key, val string) {
	zap.S().Infow("Setting Github Action output", key, val)
	fmt.Printf("::set-output name=%s::%s\n", key, val)
}

func GitHubActionSetFilesOutput(key, rootDir string, files []string) error {
	// Only run filepath.Rel if rootDir != ""
	if len(rootDir) != 0 {
		var err error
		for i := range files {
			// Make the file path relative to root dir
			files[i], err = filepath.Rel(rootDir, files[i])
			if err != nil {
				return err
			}
		}
	}
	// Run the generic function
	GitHubActionSetOutput(key, strings.Join(files, " "))
	return nil
}
