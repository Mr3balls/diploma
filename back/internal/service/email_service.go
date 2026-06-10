package service

import (
	"fmt"
	"log"
	"time"

	"esports-backend/internal/pkg/email"
)

// EmailService sends transactional emails.
// All methods silently log and skip when the sender is nil (SMTP not configured).
type EmailService struct {
	sender *email.Sender
}

func NewEmailService(sender *email.Sender) *EmailService {
	return &EmailService{sender: sender}
}

func (s *EmailService) enabled() bool { return s.sender != nil }

func (s *EmailService) send(to, subject, body string) {
	if err := s.sender.Send(to, subject, body); err != nil {
		log.Printf("email send error to %s: %v", to, err)
	}
}

// SendPasswordReset sends a password-reset link to the user.
func (s *EmailService) SendPasswordReset(to, resetURL string) {
	if !s.enabled() {
		return
	}
	subject := "Сброс пароля — ACE Tournament"
	body := layout("Сброс пароля", fmt.Sprintf(`
		<p>Вы запросили сброс пароля. Нажмите кнопку ниже, чтобы создать новый пароль:</p>
		<p style="margin:24px 0;">
			<a href="%s" style="display:inline-block;background:#ff5500;color:#ffffff;padding:12px 28px;border-radius:8px;font-weight:bold;text-decoration:none;">
				Сбросить пароль
			</a>
		</p>
		<p style="color:#90b8ff;font-size:13px;">Ссылка действительна 30 минут. Если вы не запрашивали сброс — проигнорируйте это письмо.</p>
		<p style="color:#4a6fa8;font-size:11px;word-break:break-all;">Или перейдите по ссылке: %s</p>
	`, resetURL, resetURL))
	s.send(to, subject, body)
}

// SendTeamInvite notifies a player they were added to a team.
func (s *EmailService) SendTeamInvite(to, teamName, tournamentTitle string) {
	if !s.enabled() {
		return
	}
	subject := fmt.Sprintf("Вас добавили в команду «%s» — ACE Tournament", teamName)
	body := layout("Приглашение в команду", fmt.Sprintf(`
		<p>Вас добавили в команду <strong>%s</strong> на турнир <strong>%s</strong>.</p>
		<p>Войдите на платформу и подтвердите участие в разделе уведомлений.</p>
	`, teamName, tournamentTitle))
	s.send(to, subject, body)
}

// SendMatchScheduled notifies a player about a scheduled match.
func (s *EmailService) SendMatchScheduled(to, tournamentTitle string, scheduledAt time.Time, location string) {
	if !s.enabled() {
		return
	}
	subject := fmt.Sprintf("Матч назначен — %s", tournamentTitle)
	locationLine := ""
	if location != "" {
		locationLine = fmt.Sprintf(`<p>Место / сервер: <strong>%s</strong></p>`, location)
	}
	body := layout("Матч назначен", fmt.Sprintf(`
		<p>Ваш матч в турнире <strong>%s</strong> назначен.</p>
		<p>Дата и время: <strong>%s</strong></p>
		%s
		<p style="color:#90b8ff;">Войдите на платформу и подтвердите готовность.</p>
	`, tournamentTitle, scheduledAt.Format("02.01.2006 в 15:04"), locationLine))
	s.send(to, subject, body)
}

// SendResultConfirmed notifies team members that their match result was confirmed.
func (s *EmailService) SendResultConfirmed(to, tournamentTitle, winnerTeamName string) {
	if !s.enabled() {
		return
	}
	subject := fmt.Sprintf("Результат матча подтверждён — %s", tournamentTitle)
	body := layout("Результат подтверждён", fmt.Sprintf(`
		<p>Результат матча в турнире <strong>%s</strong> подтверждён.</p>
		<p>Победитель: <strong>%s</strong></p>
		<p style="color:#90b8ff;">Откройте платформу, чтобы посмотреть обновлённую турнирную сетку.</p>
	`, tournamentTitle, winnerTeamName))
	s.send(to, subject, body)
}

// layout wraps content in a minimal branded HTML email.
func layout(title, content string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head><meta charset="UTF-8"><title>%s</title></head>
<body style="margin:0;padding:0;background:#001538;font-family:Arial,sans-serif;color:#e0e8ff;">
  <table width="100%%" cellpadding="0" cellspacing="0">
    <tr><td align="center" style="padding:40px 16px;">
      <table width="600" cellpadding="0" cellspacing="0" style="background:#001f52;border-radius:16px;overflow:hidden;">
        <tr>
          <td style="background:#0a3575;padding:24px 32px;">
            <span style="font-size:22px;font-weight:bold;color:#ffffff;letter-spacing:2px;">ACE</span>
            <span style="font-size:14px;color:#90b8ff;margin-left:8px;">Tournament Platform</span>
          </td>
        </tr>
        <tr>
          <td style="padding:32px;">
            <h2 style="margin:0 0 16px;color:#ffffff;">%s</h2>
            %s
          </td>
        </tr>
        <tr>
          <td style="padding:16px 32px;border-top:1px solid #0a3575;">
            <p style="margin:0;font-size:12px;color:#4a6fa8;">
              Это автоматическое письмо от ACE Tournament Platform. Не отвечайте на него.
            </p>
          </td>
        </tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, title, title, content)
}
