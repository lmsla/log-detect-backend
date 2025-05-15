package services

import (
	"errors"
	"fmt"
	"log"
	"log-detect/global"
	"net/smtp"
	// "log-detect/utils"
	"bytes"
	// "encoding/base64"
	// "io/ioutil"
	"strings"
	"time"
)

func Mail4(receiver, cc, bcc []string, subject string, logname string, removed []string) {
	// Load 環境參數
	user := global.EnvConfig.Email.User
	password := global.EnvConfig.Email.Password
	host := global.EnvConfig.Email.Host
	port := global.EnvConfig.Email.Port

	// body := fmt.Sprintf("%s 日誌，失聯主機:%s", logname, removed)

	// 將 removed 數組轉換為 HTML 表格
	tableRows := ""
	for i, item := range removed {
		tableRows += fmt.Sprintf("<tr><td>%d</td><td>%s</td></tr>", i+1, item)
	}
	table := fmt.Sprintf(`
		<table border="2" style="border-collapse:collapse;table-layout:auto;text-align:left;">
			<tr>
				<th>#</th>
				<th>Host</th>
			</tr>
			%s
		</table>`, tableRows)

	// 組裝 HTML 內容
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>%s</title>
			<style>
				table { 
					border-collapse: collapse; 
					table-layout: auto; 
				}
				th, td { 
					padding: 8px 12px; 
					border: 1px solid #ddd; 
					text-align: left; 
					white-space: nowrap; /* 防止折行 */
				}
			</style>			
		</head>
		<body>
			<p>%s 日誌，失聯主機如下：</p>
			%s
		</body>
		</html>`, subject, logname, table)

	// r := NewRequest(receiver, cc, bcc, subject, body)

	// fmt.Println("\ntest func mailWithNoAuth")
	// ok, mailError := r.SendEmailTest4()
	// if mailError != nil {
	// 	fmt.Println("Fail to send Email")
	// }
	// if ok {
	// 	fmt.Println("Success to send Email")
	// }

	// fmt.Println("\ntest func complex Mail")
	var mail Mail

	if user == "" {
		mail = &SendMail{host: host, port: port}
	} else {
		mail = &SendMail{user: user, password: password, host: host, port: port}
	}
	// fmt.Println("mail",mail)

	message := Message{
		from:        global.EnvConfig.Email.Sender,
		to:          receiver,
		subject:     subject,
		body:        body,
		contentType: "text/html;charset=utf-8",
		// attachment: Attachment{
		//     name:        "test.jpg",
		//     contentType: "image/jpg",
		//     withFile:    true,
		// },
		// attachment: Attachment{
		//     // name:        "/Users/chen/Documents/gitlab/git-out/product/report-backend/pdf/pdf_file/report01 2022-09-07 08:00:00~2022-09-07 16:24:25.pdf",
		//     name: pdfPath,
		//     contentType: "application/octet-stream",
		//     withFile:    true,
		// },
	}

	err := newFunction(mail, message)
	if err != nil {
		fmt.Println("Fail to send Email")
		fmt.Println(err)
	} else {
		fmt.Println("Success to send Email")
	}

	// fmt.Println("\nEmail test finish")

}

func (r *Request) SendEmailTest4() (bool, error) {

	var stringMsg string
	if len(r.to) > 0 {
		stringMsg += fmt.Sprintf("To: %s\r\n", strings.Join(r.to, ";"))
	}
	if len(r.cc) > 0 {
		if len(r.cc[0]) != 0 {
			stringMsg += fmt.Sprintf("Cc: %s\r\n", strings.Join(r.cc, ";"))
		}
	}

	subject := "Subject: " + r.subject + " !\n"
	stringMsg += subject

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	stringMsg += mime

	stringMsg += r.body
	msg := []byte(stringMsg)

	var receivers []string = r.to
	for _, cc := range r.cc {
		if len(cc) != 0 {
			receivers = append(receivers, cc)
		}
	}
	for _, bcc := range r.bcc {
		if len(bcc) != 0 {
			receivers = append(receivers, bcc)
		}
	}

	sendEmailOK := false
	for _, addr := range global.EnvConfig.Email.SMTP {
		sendEmailOK = mailWithNoAuth(addr, r.to, msg)
		if sendEmailOK {
			break
		}
	}
	if !sendEmailOK {
		return false, errors.New("此次 email 寄送失敗")
		// return false, fmt.Println("此次 email 寄送失敗")
	}
	return true, nil
}

func mailWithNoAuth(addr string, to []string, msg []byte) bool {

	env_smtp := addr
	env_sender := global.EnvConfig.Email.Sender
	env_to := to

	c, err := smtp.Dial(env_smtp)
	if err != nil {
		log.Printf("[Error] Failed to send email (No Auth) by Dail, err:%s\n", err)
		return false
	}
	defer c.Quit()

	// Set the sender and recipient.
	err = c.Mail(env_sender)

	if err != nil {
		log.Printf("[Error] Failed to send email (No Auth) by Mail, err:%s\n", err)
		return false
	}

	for _, receiver := range env_to {
		err = c.Rcpt(receiver)
		if err != nil {
			log.Printf("[Error] Failed to send email (No Auth) by Rcpt, err:%s\n", err)
			return false
		}
	}

	// Send the email body.
	wc, _ := c.Data()
	if err != nil {
		log.Printf("[Error] Failed to send email (No Auth) by Data, err:%s\n", err)
		return false
	}

	defer wc.Close()

	_, err = wc.Write(msg)
	if err != nil {
		log.Printf("[Error] Failed to send email (No Auth) by WriteTo, err:%s\n", err)
		return false
	}

	return true
}

func NewRequest(to, cc, bcc []string, subject, body string) *Request {
	return &Request{
		to:          to,
		cc:          cc,
		bcc:         bcc,
		subject:     subject,
		body:        body,
		attachments: make(map[string][]byte),
	}
}

// Request struct
type Request struct {
	// from        string
	to          []string
	cc          []string
	bcc         []string
	subject     string
	body        string
	attachments map[string][]byte
}

type Mail interface {
	Auth()
	Send(message Message) error
}

type SendMail struct {
	user     string
	password string
	host     string
	port     string
	auth     smtp.Auth
}

// type Attachment struct {
// 	name        []string
// 	filepath    []string
// 	contentType string
// 	withFile    bool
// }

type Message struct {
	from        string
	to          []string
	cc          []string
	bcc         []string
	subject     string
	body        string
	contentType string
	// attachment  Attachment
}

func SendEmail(receiver []string, subject string, logname string, removed []string) {

	user := global.EnvConfig.Email.User
	password := global.EnvConfig.Email.Password
	host := global.EnvConfig.Email.Host
	port := global.EnvConfig.Email.Port
	// send := viper.Get("send")
	fmt.Println(user, password, host, port)

	var mail Mail

	if user == "" {
		mail = &SendMail{host: host, port: port}
	} else {
		mail = &SendMail{user: user, password: password, host: host, port: port}
	}
	// fmt.Println("mail",mail)
	body := fmt.Sprintf("%s 日誌，失聯主機:%s", logname, removed)
	message := Message{
		from: global.EnvConfig.Email.Sender,
		// to:  []string{"rabot6201@gmail.com"},
		// cc:  []string{"russell.chen@bimap.co"},
		// bcc: []string{"russell.chen@bimap.co"},
		to: receiver,
		// cc:          cc_list,
		// bcc:         bcc_list,
		subject:     subject,
		body:        body,
		contentType: "text/plain;charset=utf-8",
		// attachment: Attachment{
		//     name:        "test.jpg",
		//     contentType: "image/jpg",
		//     withFile:    true,
		// },
		// attachment: Attachment{
		// 	// name:        "/Users/chen/Documents/gitlab/git-out/product/report-backend/pdf/pdf_file/report01 2022-09-07 08:00:00~2022-09-07 16:24:25.pdf",
		// 	name:        []string{"/Users/chen/Downloads/00個人研究/test/report_files/report01.pdf", "/Users/chen/Downloads/00個人研究/test/report_files/report02.pdf"},
		// 	contentType: "application/octet-stream",
		// 	withFile:    true,
		// },
	}

	// mail.Send(message)
	err := newFunction(mail, message)
	if err != nil {
		fmt.Println("Send mail error!")
		fmt.Println(err)
	} else {
		fmt.Println("Send mail success!")
	}

}

func newFunction(mail Mail, message Message) error {
	err := mail.Send(message)
	return err
}

func (mail *SendMail) Auth() {
	// mail.auth = smtp.PlainAuth("", mail.user, mail.password, mail.host)
	mail.auth = LoginAuth(mail.user, mail.password)
}

func (mail SendMail) Send(message Message) error {

	toAddress := MergeSlice(message.to, message.cc)
	toAddress = MergeSlice(toAddress, message.bcc)

	// mail.Auth()
	buffer := bytes.NewBuffer(nil)
	boundary := "GoBoundary"
	Header := make(map[string]string)
	Header["From"] = message.from
	Header["To"] = strings.Join(message.to, ";")
	Header["Cc"] = strings.Join(message.cc, ";")
	Header["Bcc"] = strings.Join(message.bcc, ";")
	Header["Subject"] = message.subject
	Header["Content-Type"] = "multipart/mixed;boundary=" + boundary
	Header["Mime-Version"] = "1.0"
	Header["Date"] = time.Now().String()
	mail.writeHeader(buffer, Header)

	body := "\r\n--" + boundary + "\r\n"
	body += "Content-Type:" + message.contentType + "\r\n"
	body += "\r\n" + message.body + "\r\n"
	buffer.WriteString(body)
	buffer.WriteString("\r\n--" + boundary + "--")
	// for _, name := range message.attachment.name {
	// 	if message.attachment.withFile {
	// 		attachment := "\r\n--" + boundary + "\r\n"
	// 		attachment += "Content-Transfer-Encoding:base64\r\n"
	// 		attachment += "Content-Disposition:attachment\r\n"
	// 		attachment += "Content-Type:" + message.attachment.contentType + ";name=\"" + mime.BEncoding.Encode("UTF-8", name) + "\"\r\n"
	// 		buffer.WriteString(attachment)
	// 		defer func() {
	// 			if err := recover(); err != nil {
	// 				log.Fatalln(err)
	// 			}
	// 		}()
	// 		// mail.writeFile(buffer, message.attachment.name)
	// 		mail.writeFile(buffer, global.EnvConfig.Files.ReportFile+"/"+name+".pdf")
	// 	}
	// }
	// 決定發信方式
	addr := fmt.Sprintf("%s:%s", mail.host, mail.port)
	from := message.from
	msg := buffer.Bytes()
	switch {
	case global.EnvConfig.Email.Auth:
		var auth smtp.Auth
		if global.EnvConfig.Email.AuthType == "LoginAuth" {
			auth = LoginAuth(mail.user, mail.password)
		} else {
			auth = smtp.PlainAuth("", mail.user, mail.password, mail.host)
		}
		if err := smtp.SendMail(addr, auth, from, toAddress, msg); err != nil {
			return fmt.Errorf("SendMail with Auth failed: %w", err)
		}
		return nil

	case !global.EnvConfig.Email.DisableTLS:
		if err := smtp.SendMail(addr, nil, from, toAddress, msg); err != nil {
			return fmt.Errorf("SendMail with TLS but no Auth failed: %w", err)
		}
		return nil

	case global.EnvConfig.Email.DisableTLS:
		// 模擬 NoAuth + NoTLS
		c, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("dial failed: %w", err)
		}
		defer c.Quit()

		if err = c.Mail(from); err != nil {
			return fmt.Errorf("MAIL FROM failed: %w", err)
		}
		for _, rcpt := range toAddress {
			if err = c.Rcpt(rcpt); err != nil {
				return fmt.Errorf("RCPT TO failed (%s): %w", rcpt, err)
			}
		}

		wc, err := c.Data()
		if err != nil {
			return fmt.Errorf("DATA failed: %w", err)
		}
		defer wc.Close()

		if _, err = wc.Write(msg); err != nil {
			return fmt.Errorf("write message failed: %w", err)
		}
		return nil

	default:
		return errors.New("no valid SMTP auth/TLS configuration found")
	}

	// err := smtp.SendMail(mail.host+":"+mail.port, mail.auth, message.from, to_address, buffer.Bytes())
	// return err
}

func MergeSlice(s1 []string, s2 []string) []string {
	slice := make([]string, len(s1)+len(s2))
	copy(slice, s1)
	copy(slice[len(s1):], s2)
	return slice
}

func (mail SendMail) writeHeader(buffer *bytes.Buffer, Header map[string]string) string {
	header := ""
	for key, value := range Header {
		header += key + ":" + value + "\r\n"
	}
	header += "\r\n"
	buffer.WriteString(header)
	return header
}

// read and write the file to buffer
// func (mail SendMail) writeFile(buffer *bytes.Buffer, fileName string) {
// 	file, err := ioutil.ReadFile(fileName)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	payload := make([]byte, base64.StdEncoding.EncodedLen(len(file)))
// 	base64.StdEncoding.Encode(payload, file)
// 	buffer.WriteString("\r\n")
// 	for index, line := 0, len(payload); index < line; index++ {
// 		buffer.WriteByte(payload[index])
// 		if (index+1)%76 == 0 {
// 			buffer.WriteString("\r\n")
// 		}
// 	}
// }

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// return "LOGIN", []byte{}, nil
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		}
	}
	return nil, nil
}
