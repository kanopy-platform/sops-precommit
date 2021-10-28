package main

import (

	"os"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	sopsconf "go.mozilla.org/sops/v3/config"
	"go.mozilla.org/sops/v3/decrypt"
)


var SopsNoConfigMatch = errors.New("error loading config: no matching creation rules found")

var usage = "sopstest file.."

var silent bool

func init() {
	flag.BoolVar(&silent, "silent", false, "Suppress output")
	flag.Parse()
}
func main() {

	exitCode := 0
	defer func() { os.Exit(exitCode) } ()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Println(usage)
		exitCode = 1
		return
	}

	// List of files in the change set 
	files := os.Args[1:]

	// Test for a sops config file, if we don't find one, we will decrypt all input
	confPath, err := sopsconf.FindConfigFile(".")
	hasConfig := true
	if err != nil {
		if err.Error() == "Config file not found" {
			log(fmt.Sprintln("No sops config found in repo, testing all files."))
			hasConfig = false
		} else {
			log(fmt.Sprintln(err))
			return
		}
	}

	filteredFiles := []string{}

	// If we have a sops config we will use it to filter for files we expect to be encrypted in the change set
	if hasConfig {
		//sopsConfigs := map[string]*sopsconf.Config{}
		for _, f := range files {
			c, e := sopsconf.LoadCreationRuleForFile(confPath, f, map[string]*string{})
			if e != nil && e.Error() == SopsNoConfigMatch.Error() {
				log(fmt.Sprintf("File: %s doesn't match any sops config creation_rule regex. Skipping.\n", f))
				continue
			}
			if e != nil {
				log(fmt.Sprintln(e))
				return
			}
			if c != nil {
				filteredFiles = append(filteredFiles, f)
			}
		}
	} else { // No sops config we parse every file in the change set
		filteredFiles = append(filteredFiles, files...)
	}

	for _, file := range filteredFiles {
		_, err := decrypt.File(file, filepath.Ext(file))
		// If we fail to decrypt, note the file and error and process the rest of the change set for other failures
		if err != nil {
			log(fmt.Sprintf("Error derypting %s: %s\n", file, err))
			exitCode = 1	
			continue
		}
		// List validated files because it is a better user experience
		if !silent {
		log(fmt.Sprintln("File: ", file," encryption validated"))
		}
	}
}

func log(message string) {
	if !silent {
		fmt.Println(message)
	}
}
