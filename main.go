package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/joho/godotenv"
	claude "github.com/potproject/claude-sdk-go"
	// Import godotenv
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	// from := os.Getenv("EMAIL_SENDER")
	notesPath := os.Getenv("NOTES_PATH")
	apiKey := os.Getenv("API_KEY")
	// password := os.Getenv("EMAIL_PASSWORD")
	// to := []string{os.Getenv("EMAIL_RECEIVER")}
	client := claude.NewClient(apiKey)
	notes := readLastWeekNotes(notesPath)
	summaries := []string{}
	var wg = &sync.WaitGroup{}

	ch := make(chan string, len(notes))
	for _, note := range notes {
		wg.Add(1)
		go func() {
			ch <- summarizeNote(client, note[1], note[0])
			wg.Done()
		}()
		summary := <-ch
		summaries = append(summaries, summary)
	}
	wg.Wait()
	fmt.Println(summaries)
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
func summarizeNote(client *claude.Client, content string, dayOfWeek string) string {
	prompt := "Summarize the text after this colon in about 3-4 sentences. If a summary cannot be made just say 'No summary available'."
	m := claude.RequestBodyMessages{
		Model:     "claude-3-opus-20240229",
		MaxTokens: 400,
		Messages: []claude.RequestBodyMessagesMessages{
			{
				Role:    claude.MessagesRoleUser,
				Content: fmt.Sprintf("%v: %v", prompt, content),
			},
		},

	}
	ctx := context.Background()
	res, err := client.CreateMessages(ctx, m)
	if err != nil {
		fmt.Println(err)
		return dayOfWeek + ": No summary available"
	}
	return dayOfWeek + ": " + (res.Content[0].Text)
}
func readLastWeekNotes(notesPath string) [][2]string {
	localTime := time.Now()
	daysInWeek := 7
	//go through each file from the past 7 days every sunday
	//make a and return that slice, which contains the notes from every day in the week
	notes := [][2]string{}
	for daysFromToday := daysInWeek; daysFromToday > 0; daysFromToday -= 1 {
		day := localTime.AddDate(0, 0, -daysFromToday)
		curFileName := (convertDateToFilePath(day))
		filePath := filepath.Clean(notesPath) + string(filepath.Separator) + curFileName + ".md"
		textBytes, err := os.ReadFile(filePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				writeToLogFile(err.Error())
				log.Fatal("Failed to open file specified by path " + filePath)
			}
		}
		notes = append(notes, [2]string{day.Weekday().String(), string(textBytes)})
	}
	return notes
}
func convertDateToFilePath(date time.Time) string {
	currentMonth := date.Month()
	currentDay := date.Day()
	currentYear := date.Year()
	return fmt.Sprintf("%v-%02d-%02d", currentYear, int(currentMonth), currentDay)
}
func writeToLogFile(content string) {
	logFile := "cronjob.log"
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	logger.Printf("%v", content)
}
