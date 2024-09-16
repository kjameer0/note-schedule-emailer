package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"

	// Import godotenv
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	from := os.Getenv("EMAIL_SENDER")
	to := []string{os.Getenv("EMAIL_RECEIVER")}
	password := os.Getenv("EMAIL_PASSWORD")
	readLastWeekNotes()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// smtp server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Message.
	message := []byte("This is a test email message.")

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending email.
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email Sent Successfully!")
}

//i basically need to grab the text from the last week's notes
//then i just have to run this file as a cronjob every week
func readLastWeekNotes() {
	notesPath := os.Getenv("NOTES_PATH")
	
}
