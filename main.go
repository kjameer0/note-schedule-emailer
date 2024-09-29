package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	claude "github.com/potproject/claude-sdk-go"
	// Import godotenv
)

type Summary struct {
	Day     int
	Content string
	Err     error
}
type Note struct {
	Day     int
	Content string
}
type EmailConfig struct {
	to       []string
	from     string
	password string
	smtpHost string
	smtpPort string
}

const timePeriod = 7

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	// // smtp server configuration.
	subject := "Subject: check out my notes!"
	emailConfig := EmailConfig{
		to:       []string{os.Getenv("EMAIL_RECEIVER")},
		from:     os.Getenv("EMAIL_SENDER"),
		password: os.Getenv("EMAIL_PASSWORD"),
		smtpHost: "smtp.gmail.com",
		smtpPort: "587",
	}
	notesPath := os.Getenv("NOTES_PATH")
	apiKey := os.Getenv("API_KEY")
	client := claude.NewClient(apiKey)
	notes := readLastWeekNotes(notesPath)
	var wg = &sync.WaitGroup{}

	ch := make(chan Summary, len(notes))
	wg.Add(len(notes))
	for i, note := range notes {
		go func(g int, note Note) {
			defer wg.Done()
			summary, summaryErr := summarizeNote(client, note.Content)
			if summaryErr != nil {
				summary = "Summarization failed. Please check the logs for more details."
				writeToLogFile(summaryErr.Error())
			}
			summaryStruct := Summary{
				Day:     g,
				Content: summary,
				Err:     summaryErr,
			}
			ch <- summaryStruct
		}(i, note)
	}
	wg.Wait()
	close(ch)
	summaries := [timePeriod]string{}
	for summary := range ch {
		dayOfWeek := time.Weekday(summary.Day).String()
		summaries[summary.Day] = dayOfWeek + ": \n" + summary.Content
	}
	sendEmail(emailConfig, subject, strings.Join(summaries[:], "\n"))
}

func sendEmail(emailConfig EmailConfig, subject string, message string) {
	auth := smtp.PlainAuth("", emailConfig.from, emailConfig.password, emailConfig.smtpHost)
	fullEmail := []byte(fmt.Sprintf("%v\r\n\r\n%v\r\n", subject, message))
	err := smtp.SendMail(emailConfig.smtpHost+":"+emailConfig.smtpPort, auth, emailConfig.from, emailConfig.to, fullEmail)
	if err != nil {
		writeToLogFile(err.Error())
		os.Exit(1)
		return
	}
	writeToLogFile("Email Sent Successfully!")
}

// request a summary of the note from the client
func summarizeNote(client *claude.Client, content string) (string, error) {
	prompt := "Summarize the text after this colon in about 4 sentences. Pretend you are the writer. Do not say that you are making a summary. If a summary cannot be made just say 'No summary available'. If a summary can be made, make sure you summarize every bullet point. Make sure you make reference to every bullet point in the note"
	if len(content) == 0 {
		return "No summary available", nil
	}
	m := claude.RequestBodyMessages{
		Model:     "claude-3-5-sonnet-20240620",
		MaxTokens: 1000,
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
		writeToLogFile(err.Error())
		return "Something went wrong while summarizing", err
	}
	return (res.Content[0].Text), nil
}
func readLastWeekNotes(notesPath string) []Note {
	localTime := time.Now()
	//go through each file from the specified time period
	notes := []Note{}
	for daysFromToday := timePeriod; daysFromToday > 0; daysFromToday -= 1 {
		day := localTime.AddDate(0, 0, -daysFromToday)
		curFileName := (convertDateToFilePath(day))
		filePath := filepath.Clean(notesPath) + string(filepath.Separator) + curFileName + ".md"
		textBytes, err := os.ReadFile(filePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				notes = append(notes, Note{int(day.Weekday()), ""})
				continue
			} else {
				writeToLogFile(err.Error())
				log.Fatal("Failed to open file specified by path " + filePath)
			}
		}
		notes = append(notes, Note{int(day.Weekday()), string(textBytes)})
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
