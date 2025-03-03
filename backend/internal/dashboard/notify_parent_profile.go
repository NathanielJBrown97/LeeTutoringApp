package dashboard

// import (
// 	"fmt"
// 	"net/smtp"
// )

// // sendAlertEmail sends an email alert to nathaniel@leetutoring.com
// // when a parent's invoice email is updated.
// func sendAlertEmail(parentName, invoiceEmail string) error {
// 	// Configure these values appropriately
// 	from := "nathaniel@leetutoring.com" // your sender email address
// 	password := ""                      // your email password or app-specific password
// 	to := "nathaniel@leetutoring.com"   // alert recipient
// 	smtpHost := "smtp.gmail.com"        // for example, if you're using Gmail
// 	smtpPort := "587"

// 	subject := "Subject: ALERT - Parent Changed Invoice Email Preferences\r\n"
// 	// The body of the email
// 	body := fmt.Sprintf("ALERT - Parent Changed Invoice Email Preferences - %s - Please Bill to %s\r\n", parentName, invoiceEmail)

// 	// Combine headers and body into a message
// 	message := []byte(subject + "\r\n" + body)

// 	// Set up authentication information.
// 	auth := smtp.PlainAuth("", from, password, smtpHost)

// 	// Send the email.
// 	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
// 	return err
// }
