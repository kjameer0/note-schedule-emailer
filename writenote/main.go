package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	// "ioutil"
)

type Config struct {
	Directory struct {
		NotesPath string `yaml:"notesPath" envconfig:"DIR_NOTES_PATH"`
	} `yaml:"directory"`
}

// parse directory and read it
func main() {
	flag.Parse()
	fileText := flag.Args()
	var cfg Config
	readFile(&cfg)
	readEnv(&cfg)
	fmt.Printf("%+v\n", cfg.Directory.NotesPath)
	fmt.Printf("%v", fileText[len(fileText)-1])
	if len(fileText) == 0 {
		log.Fatal("Please enter text to add to notes")
	}
	writeNoteInFile(&cfg, fileText[len(fileText)-1])
}
func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}
func readFile(cfg *Config) {
	f, err := os.Open("../config.yml")
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processError(err)
	}
}
func readEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		processError(err)
	}
}

func writeNoteInFile(cfg *Config, noteText string) {
	notesPath := cfg.Directory.NotesPath
	localTime := time.Now()
	currentMonth := localTime.Month()
	currentDay := localTime.Day()
	currentYear := localTime.Year()
	fileName := fmt.Sprintf("%v-%02d-%02d", currentYear, int(currentMonth), currentDay)
	filePath := filepath.Clean(notesPath) + string(filepath.Separator) + fileName + ".md"
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f, err = os.Create(filePath)
			if err != nil {
				log.Fatal("Failed to create new file for note.")
			}
		} else {
			log.Fatal("Failed to open file specified by path " + "filePath")
		}
	}
	defer f.Close()
	_, err = f.Write([]byte("\n- " + noteText))
	if err != nil {
		log.Fatal("Failed to write to file specified by path", err)
	}
}
