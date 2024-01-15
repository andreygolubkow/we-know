package jobs

import (
	"fmt"
	"github.com/gocraft/work"
	"net/smtp"
)

type Context struct {
	email  string
	userId int64
}

func (c *Context) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	fmt.Println("Starting job: ", job.Name)
	return next()
}

func (c *Context) FindCustomer(job *work.Job, next work.NextMiddlewareFunc) error {
	// If there's a user_id param, set it in the context for future middleware and handlers to use.
	if _, ok := job.Args["user_id"]; ok {
		c.userId = job.ArgInt64("user_id")
		c.email = job.ArgString("email_address")
		if err := job.ArgError(); err != nil {
			return err
		}
	}
	return next()
}

func (c *Context) SendWelcomeEmail(job *work.Job) error {
	// Extract arguments:
	addr := job.ArgString("email_address")
	if err := job.ArgError(); err != nil {
		return err
	}

	fmt.Println("Sending email")
	// Sending email using net/smtp library
	// (Email should be using gmail domain and password should be created using App password)
	from, password := "xyz@gmail.com", "sthkxlpixjferhfd"
	to := []string{
		addr,
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	message := []byte("This is a Welcome email message.")

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println("Email Sent Successfully!")
	return nil
}
