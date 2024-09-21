package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/smtp"
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
	// // smtp server configuration.
	subject := "Subject: check out my notes!"
	message := "This is the email body."
	emailConfig := EmailConfig{
		to:       []string{os.Getenv("EMAIL_RECEIVER")},
		from:     os.Getenv("EMAIL_SENDER"),
		password: os.Getenv("EMAIL_PASSWORD"),
		smtpHost: "smtp.gmail.com",
		smtpPort: "587",
	}
	sendEmail(emailConfig, subject, message)
	notesPath := os.Getenv("NOTES_PATH")
	apiKey := os.Getenv("API_KEY")
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

}

type EmailConfig struct {
	to       []string
	from     string
	password string
	smtpHost string
	smtpPort string
}

func sendEmail(emailConfig EmailConfig, subject string, message string) {

	// // Message.
	// // Authentication.
	auth := smtp.PlainAuth("", emailConfig.from, emailConfig.password, emailConfig.smtpHost)
	// // Sending email.
	fullEmail := []byte(fmt.Sprintf("%v\r\n\r\n%v\r\n", subject, message))
	err := smtp.SendMail(emailConfig.smtpHost+":"+emailConfig.smtpPort, auth, emailConfig.from, emailConfig.to, fullEmail)
	if err != nil {
		writeToLogFile(err.Error())
		log.Fatal(err)
		return
	}
	//write to log file
	writeToLogFile("Email Sent Successfully!")
	fmt.Println("Email Sent Successfully!")
}

// request a summary of the note from the client
func summarizeNote(client *claude.Client, content string, dayOfWeek string) string {
	prompt := "Summarize the text after this colon in about 3-4 sentences. If a summary cannot be made just say 'No summary available'."
	m := claude.RequestBodyMessages{
		Model:     "claude-3-5-sonnet-20240620",
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
	logger.Printf("%v: %v", time.Now().Format("2006-01-02 15:04:05"), content)
}
