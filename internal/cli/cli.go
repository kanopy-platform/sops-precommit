package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	sopsconf "go.mozilla.org/sops/v3/config"
	"go.mozilla.org/sops/v3/decrypt"
)

var SopsNoConfigMatch = errors.New("error loading config: no matching creation rules found")

type RootCommand struct{}

func NewRootCommand() *cobra.Command {
	root := &RootCommand{}
	cmd := &cobra.Command{
		Use:               "sops-precommit",
		PersistentPreRunE: root.persistentPreRunE,
		RunE:              root.runE,
	}
	cmd.PersistentFlags().String("log-level", "info", "Configure log level")
	return cmd
}

type decryption interface {
	File(filepath string, ext string) ([]byte, error)
}

type sopsrules interface {
	IsFileMatchCreationRule(file string) (bool, error)
	HasConf() bool
}

type sopsclient struct {
	ConfPath string
}

func (s *sopsclient) File(filepath string, ext string) ([]byte, error) {
	return decrypt.File(filepath, ext)
}

func (s *sopsclient) IsFileMatchCreationRule(file string) (bool, error) {
	c, err := sopsconf.LoadCreationRuleForFile(s.ConfPath, file, map[string]*string{})
	if err != nil && err.Error() == SopsNoConfigMatch.Error() {
		log.Debugf("File: %s doesn't match any sops config creation_rule regex. Skipping.\n", file)
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if c != nil {
		return true, nil
	}
	return false, nil
}

func (s *sopsclient) HasConf() bool {
	return s.ConfPath != ""
}

func (c *RootCommand) persistentPreRunE(cmd *cobra.Command, args []string) error {
	// bind flags to viper
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("sops_pre_commit")
	viper.AutomaticEnv()
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}
	// set log level
	logLevel, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		return err
	}
	log.SetLevel(logLevel)
	return nil
}

func (c *RootCommand) runE(cmd *cobra.Command, args []string) error {
	log.Debug("debug logging enabled")

	sops := &sopsclient{}

	// List of files in the change set
	files := []string{}

	if len(args) < 1 {
		// Try to parse the change set from a pipe
		tmpFiles, err := parseStdin(os.Stdin)
		if err != nil {
			return err
		}
		files = append(files, tmpFiles...)
	} else {
		// Parse file list from args
		files = append(files, args...)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files: %v", files)
	}

	confPath, err := getSopsConf(".")
	if err != nil {
		return err
	}
	sops.ConfPath = confPath

	filteredFiles, err := getFilteredFiles(sops, files)
	if err != nil {
		return err
	}

	return decryptFiles(sops, filteredFiles)
}

func getFilteredFiles(sops sopsrules, files []string) ([]string, error) {
	filteredFiles := []string{}

	// If we have a sops config we will use it to filter for files we expect to be encrypted in the change set
	if sops.HasConf() {
		for _, f := range files {
			if !fileExists(f) {
				log.Infof("Secret: %s was deleted in this changeset", f)
				continue // skip if the file does not exist.  This means it has been removed from git.
			}

			matchRule, err := sops.IsFileMatchCreationRule(f)
			if err != nil {
				return nil, err
			}

			if matchRule {
				filteredFiles = append(filteredFiles, f)
			}
		}
	} else { // No sops config we parse every file in the change set
		filteredFiles = append(filteredFiles, files...)
	}
	return filteredFiles, nil
}

func getSopsConf(path string) (string, error) {
	// Test for a sops config file, if we don't find one, we will decrypt all input
	confPath, err := sopsconf.FindConfigFile(path)
	if err != nil {
		if err.Error() == "Config file not found" {
			log.Warn("No sops config found in repo, testing all files.")
		} else {
			return "", err
		}
	}
	return confPath, nil
}

func decryptFiles(d decryption, files []string) error {
	var hasError bool

	for _, file := range files {
		_, err := d.File(file, filepath.Ext(file))
		// If we fail to decrypt, note the file and error and process the rest of the change set for other failures
		if err != nil {
			log.Errorf("Error decrypting %s: %s\n", file, err)
			hasError = true
			continue
		}
		// List validated files because it is a better user experience
		log.Infof("File: %s encryption validated", file)
	}

	if hasError {
		return errors.New("failed to validate encryption")
	}

	return nil
}

func fileExists(filename string) bool {
	stat, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !stat.IsDir()
}

func parseStdin(f *os.File) ([]string, error) {
	files := []string{}
	in, err := f.Stat()
	if err != nil {
		return files, err
	}

	// If we are in char device mode we are in a terminal and were handled by arg parsing
	if in.Mode()&os.ModeCharDevice != 0 || in.Size() <= 0 {
		return files, fmt.Errorf("no input or input device")
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			files = append(files, strings.Trim(line, "\""))
		}
	}

	if err := scanner.Err(); err != nil {
		return files, err
	}

	return files, nil
}
