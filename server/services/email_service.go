package services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/flarco/g"
	"github.com/pocketbase/pocketbase/models"

	"github.com/suaobra/suaobra-app/server/repositories"
)

type EmailService struct {
	conRepo   *repositories.ConexaoRepo
	emailRepo *repositories.EmailRepo
}

func NewEmailService(
	conRepo *repositories.ConexaoRepo,
	emailRepo *repositories.EmailRepo,
) *EmailService {
	return &EmailService{conRepo: conRepo, emailRepo: emailRepo}
}

// SaveOrUpdateConfig salva ou atualiza a configuração de e-mail.
func (s *EmailService) SaveOrUpdateConfig(teamID string, fields map[string]any) (con *models.Record, email *models.Record, err error) {
	// Busca conexão existente
	exCon, _ := s.conRepo.FindActiveEmailByTeam(teamID)
	if exCon != nil && exCon.Id != "" {
		con = exCon
		exEmail, _ := s.emailRepo.FindByConexao(exCon.Id)
		if exEmail != nil && exEmail.Id != "" {
			// Atualiza existente
			err = s.emailRepo.Update(exEmail, fields)
			if err != nil {
				return nil, nil, err
			}
			return con, exEmail, nil
		}
	}

	// Cria nova conexão
	if con == nil {
		nome := fields["conexoes_email"]
		if nome == nil {
			nome = "E-mail Principal"
		}
		con, err = s.conRepo.CreateEmail(teamID, nome.(string))
		if err != nil {
			return nil, nil, err
		}
	}

	// Cria registro de email
	email, err = s.emailRepo.Create(con.Id, fields)
	if err != nil {
		return con, nil, err
	}

	return con, email, nil
}

// GetConfig retorna a configuração de e-mail do time.
func (s *EmailService) GetConfig(teamID string) (exists bool, con *models.Record, email *models.Record, err error) {
	con, err = s.conRepo.FindActiveEmailByTeam(teamID)
	if err != nil {
		return false, nil, nil, err
	}
	if con == nil || con.Id == "" {
		return false, nil, nil, nil
	}

	email, err = s.emailRepo.FindByConexao(con.Id)
	if err != nil {
		return true, con, nil, err
	}
	if email == nil || email.Id == "" {
		return true, con, nil, nil
	}

	return true, con, email, nil
}

// SendTestEmail envia um e-mail de teste usando a configuração do time.
func (s *EmailService) SendTestEmail(teamID, toEmail, subject, body string) error {
	_, _, email, err := s.GetConfig(teamID)
	if err != nil {
		return g.Error("configuração de e-mail não encontrada: %v", err)
	}
	if email == nil {
		return g.Error("configuração de e-mail não encontrada")
	}

	smtpHost := strings.TrimSpace(email.GetString("smtp_host"))
	smtpPort := int(email.GetInt("smtp_port"))
	smtpUser := strings.TrimSpace(email.GetString("smtp_usuario"))
	smtpPass := strings.TrimSpace(email.GetString("smtp_senha"))
	fromEmail := strings.TrimSpace(email.GetString("remetente_email"))
	fromName := strings.TrimSpace(email.GetString("conexoes_email"))
	replyTo := strings.TrimSpace(email.GetString("reply_to"))
	encryption := strings.TrimSpace(email.GetString("criptografia"))

	if smtpHost == "" || smtpPort == 0 || fromEmail == "" {
		return g.Error("configuração de SMTP incompleta")
	}

	// Monta a mensagem
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", fromName, fromEmail)
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	if replyTo != "" {
		headers["Reply-To"] = replyTo
	}

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)

	// Escolhe o método de envio baseado na criptografia
	switch encryption {
	case "STARTTLS":
		return s.sendViaStartTLS(addr, smtpUser, smtpPass, fromEmail, []string{toEmail}, []byte(message))
	default: // NONE
		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		return smtp.SendMail(addr, auth, fromEmail, []string{toEmail}, []byte(message))
	}
}

func (s *EmailService) sendViaStartTLS(addr, user, pass, from string, to []string, message []byte) error {
	host := strings.Split(addr, ":")[0]

	client, err := smtp.Dial(addr)
	if err != nil {
		return g.Error("erro ao conectar: %v", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	if err = client.StartTLS(tlsConfig); err != nil {
		return g.Error("erro no STARTTLS: %v", err)
	}

	auth := smtp.PlainAuth("", user, pass, host)
	if err = client.Auth(auth); err != nil {
		return g.Error("erro na autenticação: %v", err)
	}

	if err = client.Mail(from); err != nil {
		return g.Error("erro no MAIL FROM: %v", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return g.Error("erro no RCPT TO: %v", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return g.Error("erro ao iniciar DATA: %v", err)
	}

	_, err = w.Write(message)
	if err != nil {
		return g.Error("erro ao escrever mensagem: %v", err)
	}

	err = w.Close()
	if err != nil {
		return g.Error("erro ao fechar DATA: %v", err)
	}

	return client.Quit()
}
