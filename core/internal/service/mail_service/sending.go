package mail_service

import (
	"billionmail-core/internal/service/public"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gogf/gf/util/grand"
	"github.com/gogf/gf/v2/frame/g"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Title       string
	Content     string
	Headers     map[string]string
	Attachments []Attachment
}

// Attachment is a single email attachment with its raw (already decoded) content.
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// SetAttachments sets the attachments for the message.
func (m *Message) SetAttachments(attachments []Attachment) {
	m.Attachments = attachments
}

// HasAttachments reports whether the message carries any attachment.
func (m Message) HasAttachments() bool {
	return len(m.Attachments) > 0
}

// buildMultipartBody builds a multipart/mixed body (HTML part + attachment parts).
// It returns the "Content-Type" header line, the separating blank line, and the
// encoded multipart content, ready to be appended after the common headers.
func (m Message) buildMultipartBody() (string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	// The Content-Type header (with boundary) must precede the body.
	head := fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", mw.Boundary())

	// HTML body part (quoted-printable).
	htmlHeader := textproto.MIMEHeader{}
	htmlHeader.Set("Content-Type", "text/html; charset=utf-8")
	htmlHeader.Set("Content-Transfer-Encoding", "quoted-printable")
	htmlPart, err := mw.CreatePart(htmlHeader)
	if err != nil {
		return "", err
	}
	qp := quotedprintable.NewWriter(htmlPart)
	if _, err = qp.Write([]byte(m.Content)); err != nil {
		return "", err
	}
	if err = qp.Close(); err != nil {
		return "", err
	}

	// Attachment parts (base64).
	for _, att := range m.Attachments {
		contentType := att.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		attHeader := textproto.MIMEHeader{}
		attHeader.Set("Content-Type", contentType+"; "+encodeMIMEName(att.Filename))
		attHeader.Set("Content-Transfer-Encoding", "base64")
		attHeader.Set("Content-Disposition", "attachment; "+encodeMIMEFilename(att.Filename))
		part, err := mw.CreatePart(attHeader)
		if err != nil {
			return "", err
		}
		// Encode as base64 with line wrapping (RFC 2045: 76 chars per line).
		if err = writeBase64Wrapped(part, att.Data); err != nil {
			return "", err
		}
	}

	if err = mw.Close(); err != nil {
		return "", err
	}

	return head + buf.String(), nil
}

// isASCII reports whether s contains only printable ASCII (no bytes >= 0x80,
// no control chars). Such names are safe to place verbatim in a header.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] >= 0x80 {
			return false
		}
	}
	return true
}

// asciiFallbackName produces an ASCII-only placeholder for a filename while
// preserving the extension, used as the RFC 2183 "filename" fallback for
// clients that don't understand RFC 2231's "filename*".
func asciiFallbackName(filename string) string {
	ext := ""
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		e := filename[idx:]
		if isASCII(e) {
			ext = e
		}
	}
	return "attachment" + ext
}

// encodeMIMEName builds the Content-Type "name" parameter, RFC 2047-encoded
// when the filename contains non-ASCII characters.
func encodeMIMEName(filename string) string {
	if isASCII(filename) {
		return fmt.Sprintf("name=%q", filename)
	}
	// RFC 2047 encoded-word (B-encoding), understood by most mail clients.
	return fmt.Sprintf("name=%q", mime.BEncoding.Encode("UTF-8", filename))
}

// encodeMIMEFilename builds the Content-Disposition "filename" parameter.
// For non-ASCII names it emits both an ASCII fallback ("filename=") and an
// RFC 2231 UTF-8 form ("filename*="), which is the modern, correct way to
// carry Unicode filenames through SMTP.
func encodeMIMEFilename(filename string) string {
	if isASCII(filename) {
		return fmt.Sprintf("filename=%q", filename)
	}
	return fmt.Sprintf("filename=%q; filename*=UTF-8''%s",
		asciiFallbackName(filename), rfc2231Escape(filename))
}

// rfc2231Escape percent-encodes a UTF-8 string for use in an RFC 2231
// "filename*" parameter. Only attribute-safe characters are left unescaped.
func rfc2231Escape(s string) string {
	const upperhex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		// RFC 2231 "attribute-char": alphanumerics and a small safe set.
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '!' || c == '#' || c == '$' || c == '&' || c == '+' || c == '-' ||
			c == '.' || c == '^' || c == '_' || c == '`' || c == '|' || c == '~' {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(upperhex[c>>4])
		b.WriteByte(upperhex[c&0x0f])
	}
	return b.String()
}

// writeBase64Wrapped writes base64-encoded data wrapped at 76 characters per line.
func writeBase64Wrapped(w io.Writer, data []byte) error {
	encoded := base64.StdEncoding.EncodeToString(data)
	const lineLen = 76
	for i := 0; i < len(encoded); i += lineLen {
		end := i + lineLen
		if end > len(encoded) {
			end = len(encoded)
		}
		if _, err := io.WriteString(w, encoded[i:end]+"\r\n"); err != nil {
			return err
		}
	}
	return nil
}

// MailTitle get title of email
func (m Message) MailTitle() string {
	return mime.QEncoding.Encode("UTF-8", m.Title)
}

// MailText get content of email
func (m Message) MailText() string {
	buf := new(bytes.Buffer)
	wt := quotedprintable.NewWriter(buf)
	wt.Write([]byte(m.Content))
	wt.Close()
	return buf.String()
}

// MailHeader get headers of email
func (m Message) MailHeader() string {
	headers := ""
	for key, value := range m.Headers {
		headers += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	return headers
}

// MessageID get Message-ID of email
func (m Message) MessageID() string {
	if id, ok := m.Headers["Message-ID"]; ok {
		return id
	}
	return ""
}

// SetMessageID set Message-ID of email
func (m *Message) SetMessageID(id string) {
	m.SetHeader("Message-ID", id)
}

// SetHeader set header of email
func (m *Message) SetHeader(key, value string) {
	m.Headers[key] = value
}

// SetHeaders set headers of email
func (m *Message) SetHeaders(headers map[string]string) {
	for key, value := range headers {
		m.SetHeader(key, value)
	}
}

// SetRealName set RealName header of email
func (m *Message) SetRealName(realName string) {
	m.SetHeader("RealName", mime.QEncoding.Encode("UTF-8", realName))
}

func NewMessage(title string, content string) Message {
	return Message{
		Title:   title,
		Content: content,
		Headers: make(map[string]string),
	}
}

type EmailSender struct {
	Email     string       `json:"email" v:"required|email"` // email of sender
	UserName  string       `json:"username" v:"required"`
	Host      string       `json:"host" v:"required"`     // SMTP server address
	Port      string       `json:"port" v:"required"`     // SMTP server port
	Password  string       `json:"password" v:"required"` // SMTP password
	SNI       string       `json:"sni"`                   // SNI for TLS
	client    *smtp.Client // persistent SMTP client connection
	mutex     sync.Mutex   // mutex for thread safety
	connected bool         // connection status
}

type customAuth struct {
	Username, Password string
}

func (a *customAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "PLAIN", []byte("\x00" + a.Username + "\x00" + a.Password), nil
}

func (a *customAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, errors.New("unexpected server response")
	}
	return nil, nil
}

func NewEmailSender() *EmailSender {
	e := &EmailSender{
		Host:      "localhost",
		Port:      "25",
		connected: false,
	}

	if public.IsRunningInContainer() {
		e.Host = "postfix"
		// e.Port = "587"
		// e.SNI, _ = public.DockerEnv("BILLIONMAIL_HOSTNAME")
	}

	return e
}

func NewEmailSenderWithLocal(email string) (es *EmailSender, err error) {
	es = NewEmailSender()

	es.Email = email
	es.UserName = email
	es.Password, err = PasswordPlainByEmail(context.Background(), email)

	if err != nil {
		return
	}

	if !es.IsConfigured() {
		err = errors.New("Email Sender not configured")
		return
	}

	return
}

// PasswordPlainByEmail
func PasswordPlainByEmail(ctx context.Context, email string) (string, error) {
	val, err := g.DB().Model("mailbox").Where("username", email).Value("password_encode")
	if err != nil {
		return "", fmt.Errorf("query password failed: %w", err)
	}
	if val.IsEmpty() {
		return "", fmt.Errorf("password not found for %s", email)
	}
	rawHex, err := hex.DecodeString(val.String())
	if err != nil {
		return "", fmt.Errorf("hex decode failed: %w", err)
	}
	plainBytes, err := base64.StdEncoding.DecodeString(string(rawHex))
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}
	return string(plainBytes), nil
}

// Close closes the SMTP connection
func (e *EmailSender) Close() {
	_ = e.Disconnect()
}

// Connect establishes an SMTP connection
func (e *EmailSender) Connect() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.connected && e.client != nil {
		// Connection already exists
		return nil
	}

	var err error

	if e.isSecure() {
		err = e.connectWithSSL()
	} else {
		err = e.connectPlain()
	}

	if err != nil {
		e.connected = false
		e.client = nil
		return err
	}

	e.connected = true
	return nil
}

// connectWithSSL establishes a secure SMTP connection
func (e *EmailSender) connectWithSSL() error {
	conn, err := tls.Dial("tcp", net.JoinHostPort(e.Host, e.Port), &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
		ServerName:         e.SNI,
	})
	if err != nil {
		return fmt.Errorf("TLS dial: %w", err)
	}

	client, err := smtp.NewClient(conn, e.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("new SMTP client: %w", err)
	}

	auth := smtp.PlainAuth("", e.UserName, e.Password, e.Host)
	if err = client.Auth(auth); err != nil {
		client.Close()
		return fmt.Errorf("SMTP auth: %w", err)
	}

	e.client = client

	return nil
}

// connectPlain establishes a plain SMTP connection
func (e *EmailSender) connectPlain() error {
	client, err := smtp.Dial(net.JoinHostPort(e.Host, e.Port))
	if err != nil {
		return fmt.Errorf("SMTP dial: %w", err)
	}

	// Check if STARTTLS is needed
	if e.Port == "587" {
		if err = client.StartTLS(&tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
			ServerName:         e.SNI,
		}); err != nil {
			client.Close()
			return fmt.Errorf("SMTP STARTTLS: %w", err)
		}
	}

	var auth smtp.Auth

	if e.Port == "25" {
		auth = &customAuth{e.UserName, e.Password}
	} else {
		auth = smtp.PlainAuth("", e.UserName, e.Password, e.Host)
	}

	if err = client.Auth(auth); err != nil {
		client.Close()
		return fmt.Errorf("SMTP auth: %w", err)
	}

	e.client = client
	return nil
}

// Disconnect closes the SMTP connection
func (e *EmailSender) Disconnect() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if !e.connected || e.client == nil {
		return nil
	}

	err := e.client.Quit()
	if err != nil {
		// Force close connection if Quit fails
		_ = e.client.Close()
		g.Log().Warning(context.Background(), "Error closing SMTP connection: ", err)
	}

	e.connected = false
	e.client = nil
	return err
}

// GenerateMessageID generates a unique Message-ID for email
func (e *EmailSender) GenerateMessageID() string {
	randomBytes := grand.B(16)
	randomID := hex.EncodeToString(randomBytes)
	timestampMillis := time.Now().UnixMilli()

	domain := strings.SplitN(e.Email, "@", 2)
	domainPart := "billionmail"
	if len(domain) > 1 {
		domainPart = domain[1]
	}

	return fmt.Sprintf("<%d.%s@%s>", timestampMillis, randomID, domainPart)
}

// Send sends an email to specified recipients
func (e *EmailSender) Send(message Message, recipients []string) error {
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Make sure we have a connection
	if !e.connected || e.client == nil {
		e.mutex.Unlock()
		if err := e.Connect(); err != nil {
			e.mutex.Lock()
			return fmt.Errorf("failed to connect: %w", err)
		}
		e.mutex.Lock()
	}

	// Try to send the message, with reconnect on failure
	err := e.doSend(message, recipients)
	if err != nil {
		// Connection might be stale, try to reconnect once
		g.Log().Debug(context.Background(), "SMTP send failed, attempting reconnection")

		// Clean up
		e.client.Close()
		e.connected = false
		e.client = nil

		// Try to reconnect
		e.mutex.Unlock()
		if err := e.Connect(); err != nil {
			e.mutex.Lock()
			return fmt.Errorf("reconnect failed: %w", err)
		}
		e.mutex.Lock()

		// Try sending again after reconnect
		return e.doSend(message, recipients)
	}

	return nil
}

// doSend performs the actual message sending
func (e *EmailSender) doSend(message Message, recipients []string) error {
	// Add default headers if not present
	if message.Headers == nil {
		message.Headers = make(map[string]string)
	}

	// Add Message-ID if not already set
	if _, exists := message.Headers["Message-ID"]; !exists {
		message.Headers["Message-ID"] = e.GenerateMessageID()
	}

	// Add Date if not already set
	if _, exists := message.Headers["Date"]; !exists {
		message.Headers["Date"] = time.Now().Format(time.RFC1123Z)
	}

	// Note: Content-Type / Content-Transfer-Encoding are emitted per-branch below
	// (single-part vs multipart) and are intentionally NOT stored in message.Headers
	// to avoid duplicate headers via MailHeader().
	delete(message.Headers, "Content-Type")
	delete(message.Headers, "Content-Transfer-Encoding")

	// Default From header
	from := fmt.Sprintf("%s <%s>", strings.Split(e.Email, "@")[0], e.Email)

	if v, exists := message.Headers["RealName"]; exists {
		from = fmt.Sprintf("%s <%s>", v, e.Email)
		delete(message.Headers, "RealName")
	}

	if v, exists := message.Headers["From"]; exists {
		from = v
		delete(message.Headers, "From")
	}

	// Common headers shared by both single-part and multipart messages.
	commonHeaders := fmt.Sprintf("From: %s\r\n", from) +
		fmt.Sprintf("To: %s\r\n", strings.Join(recipients, ",")) +
		fmt.Sprintf("Subject: %s\r\n", message.MailTitle()) +
		"MIME-Version: 1.0\r\n" +
		"X-Mailer: BillionMail\r\n" +
		message.MailHeader()

	var msg []byte

	if message.HasAttachments() {
		// Build a multipart/mixed message: HTML body + attachment parts.
		body, buildErr := message.buildMultipartBody()
		if buildErr != nil {
			return fmt.Errorf("build multipart body: %w", buildErr)
		}
		msg = []byte(commonHeaders + body)
	} else {
		// Single-part message (original behavior): HTML body, quoted-printable.
		headerString := commonHeaders +
			"Content-Type: text/html; charset=utf-8\r\n" +
			"Content-Transfer-Encoding: quoted-printable\r\n" +
			"\r\n" +
			message.MailText() +
			"\r\n"
		msg = []byte(headerString)
	}

	var err error

	defer func() {
		// Reset the connection state if sending fails
		if err != nil {
			if err = e.client.Reset(); err != nil {
				g.Log().Warning(context.Background(), "SMTP reset: %w", err)
			}
		}
	}()

	// Set the sender
	// MAIL FROM
	if err = e.client.Mail(e.Email); err != nil {
		return fmt.Errorf("SMTP mail: %w", err)
	}

	// Set the recipients
	// RCPT TO
	for _, to := range recipients {
		if err = e.client.Rcpt(to); err != nil {
			return fmt.Errorf("SMTP rcpt: %w", err)
		}
	}

	// Get a writer for the message body
	var w io.WriteCloser
	w, err = e.client.Data()
	if err != nil {
		return fmt.Errorf("SMTP data: %w", err)
	}

	// Write the message
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("SMTP write: %w", err)
	}

	// Close the writer
	if err = w.Close(); err != nil {
		return fmt.Errorf("SMTP close writer: %w", err)
	}

	return nil
}

// isSecure check if TLS connection is needed
func (e *EmailSender) isSecure() bool {
	// usually,
	// port 465 is used for SMTP with SSL (implicit TLS)
	// port 587 is used for SMTP with STARTTLS (explicit TLS)
	if e.Port == "465" {
		return true
	} else if e.Port == "587" {
		// We will use STARTTLS if the port is 587
		return false
	}
	return false
}

// IsConfigured check if Email notification is configured
func (e *EmailSender) IsConfigured() bool {
	e.Email = strings.TrimSpace(e.Email)
	e.Host = strings.TrimSpace(e.Host)
	e.Port = strings.TrimSpace(e.Port)
	e.UserName = strings.TrimSpace(e.UserName)
	e.Password = strings.TrimSpace(e.Password)

	if e.UserName == "" || e.Host == "" || e.Port == "" || e.Password == "" || e.Email == "" {
		return false
	}

	if e.Host != "postfix" && !public.IsHost(e.Host) {
		g.Log().Warning(context.Background(), "Host is not a valid host: ", e.Host)
		return false
	}

	if err := g.Validator().Data(e).Run(context.Background()); err != nil {
		g.Log().Warning(context.Background(), "Email sender validation failed: ", err)
		return false
	}

	return true
}
