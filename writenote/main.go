package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	// "ioutil"
)

// parse directory and read it
func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	flag.Parse()
	fileText := flag.Args()
	notesPath := os.Getenv("NOTES_PATH")
	if len(fileText) != 1 {
		log.Fatal("Please enter text to add to notes")
	}
	writeNoteInFile(notesPath, fileText[len(fileText)-1])
}

func writeNoteInFile(notesPath string, noteText string) {
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
