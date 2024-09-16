package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	// Import godotenv
)

func main() {
	err := godotenv.Load(".env")
	// from := os.Getenv("EMAIL_SENDER")
	notesPath := os.Getenv("NOTES_PATH")
	readLastWeekNotes(notesPath)
	// to := []string{os.Getenv("EMAIL_RECEIVER")}
	// password := os.Getenv("EMAIL_PASSWORD")
	readLastWeekNotes(notesPath)
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// // smtp server configuration.
	// smtpHost := "smtp.gmail.com"
	// smtpPort := "587"

	// // Message.
	// message := []byte("This is a test email message.")

	// // Authentication.
	// auth := smtp.PlainAuth("", from, password, smtpHost)

	// // Sending email.
	// err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("Email Sent Successfully!")
}

// i basically need to grab the text from the last week's notes
// then i just have to run this file as a cronjob every week
func readLastWeekNotes(notesPath string) {
	localTime := time.Now()
	daysInWeek := 7
	//go through each file from the past 7 days every sunday
	for daysFromToday := daysInWeek; daysFromToday > 0; daysFromToday -= 1 {
		curFileName := (convertDateToFilePath(localTime.AddDate(0, 0, -daysFromToday)))
		filePath := filepath.Clean(notesPath) + string(filepath.Separator) + curFileName + ".md"
		text, err := os.ReadFile(filePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				log.Fatal("Failed to open file specified by path " + filePath)
			}
		}
		//request to a api to create a summary of that text
		//write that text with a day of the week as a title
		//after for loop return all text
		fmt.Println(string(text))
	}
}
func convertDateToFilePath(date time.Time) string {
	currentMonth := date.Month()
	currentDay := date.Day()
	currentYear := date.Year()
	return fmt.Sprintf("%v-%02d-%02d", currentYear, int(currentMonth), currentDay)
}
