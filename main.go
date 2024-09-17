package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
	// Import godotenv
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	// from := os.Getenv("EMAIL_SENDER")
	notesPath := os.Getenv("NOTES_PATH")
	openaiKey := os.Getenv("OPEN_AI_KEY")
	// password := os.Getenv("EMAIL_PASSWORD")
	// to := []string{os.Getenv("EMAIL_RECEIVER")}
	client := openai.NewClient(openaiKey)
	notes := readLastWeekNotes(notesPath, client)
	// summaries := []string{}
	// for _, note := range notes {
	// 	summaries = append(summaries, summarizeNote(client, note))
	// }
	fmt.Println(summarizeNote(client, notes[0]))
	// fmt.Println(notes)
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
func summarizeNote(client *openai.Client, content string) string {
	prompt := "Summarize the text after this colon"
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("%v: %v", prompt, content),
				},
			},
		},
	)

	if err != nil {
		log.Fatalf("ChatCompletion error: %v\n", err)
	}

	return (resp.Choices[0].Message.Content)
}
func readLastWeekNotes(notesPath string, client *openai.Client) []string {
	localTime := time.Now()
	daysInWeek := 7
	//go through each file from the past 7 days every sunday
	//make a and return that slice, which contains the notes from every day in the week
	notes := []string{}
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
		notes = append(notes, string(text))
	}
	return notes
}
func convertDateToFilePath(date time.Time) string {
	currentMonth := date.Month()
	currentDay := date.Day()
	currentYear := date.Year()
	return fmt.Sprintf("%v-%02d-%02d", currentYear, int(currentMonth), currentDay)
}
