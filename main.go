package main
// this file runs the cron job tasks
import (
	"fmt"
	"os"
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
	var cfg Config
	readFile(&cfg)
	readEnv(&cfg)
	fmt.Printf("%+v", cfg.Directory.NotesPath)
	fileText := []byte{71}
	os.WriteFile("test.txt", fileText, 0644)
}
func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}
func readFile(cfg *Config) {
	f, err := os.Open("config.yml")
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
