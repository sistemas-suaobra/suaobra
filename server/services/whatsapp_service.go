package services

import (
	"strings"
	"time"

	"github.com/flarco/g"
	"github.com/pocketbase/pocketbase/models"

	"github.com/suaobra/suaobra-app/server/clients/wuzapi"
	"github.com/suaobra/suaobra-app/server/repositories"
)

type WhatsAppService struct {
	conRepo *repositories.ConexaoRepo
	waRepo  *repositories.WhatsAppRepo
	wuz     *wuzapi.Client
	tokens  *TokenService
}

func NewWhatsAppService(
	conRepo *repositories.ConexaoRepo,
	waRepo *repositories.WhatsAppRepo,
	wuz *wuzapi.Client,
	tokens *TokenService,
) *WhatsAppService {
	return &WhatsAppService{conRepo: conRepo, waRepo: waRepo, wuz: wuz, tokens: tokens}
}

// Idempotente: se já existir, retorna.
func (s *WhatsAppService) CreateConnection(teamID, userID, name, apiKey string) (already bool, con *models.Record, wa *models.Record, err error) {
	exCon, _ := s.conRepo.FindActiveWhatsappByTeam(teamID)
	if exCon != nil && exCon.Id != "" {
		exWa, _ := s.waRepo.FindByConexao(exCon.Id)
		if exWa != nil && exWa.Id != "" {
			return true, exCon, exWa, nil
		}
		con = exCon
	} else {
		con = nil
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = "USER_" + userID
	}

	if strings.TrimSpace(apiKey) == "" {
		return false, nil, nil, g.Error("WHATSMEOW_APIKEY not set")
	}

	userToken, err := s.tokens.Generate(apiKey)
	if err != nil {
		return false, nil, nil, err
	}

	created, raw, err := s.wuz.CreateAdminUser(name, userToken)
	if err != nil {
		return false, nil, nil, g.Error("error creating wuzapi user: %v", err)
	}

	if con == nil {
		con, err = s.conRepo.CreateWhatsapp(teamID, name)
		if err != nil {
			return false, nil, nil, err
		}
	}

	fields := map[string]any{
		"provider":      "WUZAPI",
		"api_base_url":  s.wuzBaseURL(),
		"numero_e164":   userToken,
		"instancia_id":  created.Data.ID,
		"device_jid":    "",
		"name":         created.Data.Name,
		"webhook":      created.Data.Webhook,
		"events":       created.Data.Events,
		"raw_response": raw,
		"ultimo_qr_em": time.Now().UTC(),
	}

	wa, err = s.waRepo.Create(con.Id, fields)
	if err != nil {
		return false, con, nil, err
	}

	return false, con, wa, nil
}

func (s *WhatsAppService) ConnectSession(userToken string) (map[string]any, error) {
	return s.wuz.SessionConnect(strings.TrimSpace(userToken))
}

func (s *WhatsAppService) GetQR(userToken string) (wuzapi.SessionQRResp, error) {
	parsed, _, err := s.wuz.SessionQR(strings.TrimSpace(userToken))
	return parsed, err
}

func (s *WhatsAppService) getTokenAndRecordByTeam(teamID string) (token string, wa *models.Record, err error) {
	con, err := s.conRepo.FindActiveWhatsappByTeam(teamID)
	if err != nil || con == nil || con.Id == "" {
		return "", nil, g.Error("whatsapp connection not found")
	}

	wa, err = s.waRepo.FindByConexao(con.Id)
	if err != nil || wa == nil || wa.Id == "" {
		return "", nil, g.Error("whatsapp details not found")
	}

	token = strings.TrimSpace(wa.GetString("numero_e164"))
	if token == "" {
		return "", nil, g.Error("numero_e164 token is empty")
	}

	return token, wa, nil
}

func (s *WhatsAppService) ConnectByTeam(teamID string) (map[string]any, error) {
	token, _, err := s.getTokenAndRecordByTeam(teamID)
	if err != nil {
		return nil, err
	}
	return s.ConnectSession(token)
}

func (s *WhatsAppService) GetQRByTeam(teamID string) (code int, qr string, err error) {
	token, wa, err := s.getTokenAndRecordByTeam(teamID)
	if err != nil {
		return 0, "", err
	}

	parsed, err := s.GetQR(token)
	if err != nil {
		return 0, "", err
	}

	// atualiza timestamp do QR
	s.waRepo.TouchUltimoQR(wa)

	return parsed.Code, parsed.Data.QRCode, nil
}

func (s *WhatsAppService) wuzBaseURL() string {
	return s.wuz.BaseURL()
}

func (s *WhatsAppService) SendTestMessage(teamID, phone, body string) (map[string]any, error) {
	startedAt := time.Now()
	phone = strings.TrimSpace(phone)
	body = strings.TrimSpace(body)

	g.Info(
		"WhatsAppService.SendTestMessage: start team=%s phone=%s body_len=%d body=%s",
		teamID,
		maskWAPhone(phone),
		len(body),
		truncateWALog(body, 800),
	)

	token, wa, err := s.getTokenAndRecordByTeam(teamID)
	if err != nil {
		g.Error(err, "WhatsAppService.SendTestMessage: erro ao obter token/record team=%s", teamID)
		return nil, err
	}

	instanciaID := ""
	deviceJID := ""
	if wa != nil {
		instanciaID = strings.TrimSpace(wa.GetString("instancia_id"))
		deviceJID = strings.TrimSpace(wa.GetString("device_jid"))
	}

	g.Info(
		"WhatsAppService.SendTestMessage: resolved team=%s instancia_id=%s device_jid=%s token=%s",
		teamID,
		instanciaID,
		deviceJID,
		maskWAToken(token),
	)

	resp, err := s.wuz.SendTextMessage(token, phone, body)
	if err != nil {
		g.Error(
			err,
			"WhatsAppService.SendTestMessage: falha no WUZAPI team=%s phone=%s instancia_id=%s",
			teamID,
			maskWAPhone(phone),
			instanciaID,
		)
		return nil, err
	}

	g.Info(
		"WhatsAppService.SendTestMessage: success team=%s phone=%s duration=%s response=%s",
		teamID,
		maskWAPhone(phone),
		time.Since(startedAt),
		truncateWALog(g.Marshal(resp), 1500),
	)

	return resp, nil
}

// CheckAndUpdateStatus consulta o wuzapi para saber se o user está connected,
// e atualiza conectado_em + device_jid no registro local se necessário.
// Retorna (connected, jid, error).
func (s *WhatsAppService) CheckAndUpdateStatus(teamID string) (connected bool, jid string, err error) {
	con, err := s.conRepo.FindActiveWhatsappByTeam(teamID)
	if err != nil || con == nil || con.Id == "" {
		return false, "", g.Error("whatsapp connection not found")
	}

	wa, err := s.waRepo.FindByConexao(con.Id)
	if err != nil || wa == nil || wa.Id == "" {
		return false, "", g.Error("whatsapp details not found")
	}

	instanciaID := strings.TrimSpace(wa.GetString("instancia_id"))
	if instanciaID == "" {
		return false, "", g.Error("instancia_id is empty")
	}

	info, err := s.wuz.GetAdminUser(instanciaID)
	if err != nil {
		return false, "", err
	}

	// Verifica se está LOGADO (não apenas connected aos servidores)
	// Connected = conectado ao servidor WhatsApp
	// LoggedIn = escaneou QR e autenticou de fato
	if !info.LoggedIn {
		return false, "", nil
	}

	// Logado! Mas só salva conectado_em se for uma conexão RECENTE.
	// Verifica se ultimo_qr_em é recente (últimos 5 min) — evita detectar sessões antigas.
	conectadoEm := wa.GetDateTime("conectado_em")
	ultimoQrEm := wa.GetDateTime("ultimo_qr_em")
	
	qrRecente := !ultimoQrEm.Time().IsZero() && time.Since(ultimoQrEm.Time()) < 5*time.Minute
	
	if (conectadoEm.Time().IsZero() || wa.GetString("device_jid") != info.JID) && qrRecente {
		if saveErr := s.waRepo.UpdateConnected(wa, info.JID); saveErr != nil {
			g.Error(saveErr, "erro ao salvar conectado_em para instancia=%s", instanciaID)
		} else {
			g.Info("conectado_em salvo para instancia=%s jid=%s (QR recente, LOGADO)", instanciaID, info.JID)
		}
	}

	return true, info.JID, nil
}

func (s *WhatsAppService) Disconnect(teamID string) error {
	token, wa, err := s.getTokenAndRecordByTeam(teamID)
	if err != nil {
		return err
	}

	raw, err := s.wuz.SessionDisconnect(token)
	if err != nil {
		g.Warn("wuzapi disconnect failed team=%s err=%v", teamID, err)
	} else {
		g.Info("wuzapi disconnect success team=%s response=%v", teamID, raw)
	}

	return s.waRepo.ClearConnected(wa)
}

func (s *WhatsAppService) GetByTeam(teamID string) (exists bool, con *models.Record, wa *models.Record, err error) {
	con, err = s.conRepo.FindActiveWhatsappByTeam(teamID)
	if err != nil {
		return false, nil, nil, err
	}
	if con == nil || con.Id == "" {
		return false, nil, nil, nil
	}

	wa, err = s.waRepo.FindByConexao(con.Id)
	if err != nil {
		return true, con, nil, err
	}
	if wa == nil || wa.Id == "" {
		return true, con, nil, nil
	}

	return true, con, wa, nil
}

func truncateWALog(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

func maskWAPhone(phone string) string {
	var digits strings.Builder
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			digits.WriteRune(c)
		}
	}
	v := digits.String()
	if len(v) <= 4 {
		return v
	}
	return strings.Repeat("*", len(v)-4) + v[len(v)-4:]
}

func maskWAToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 6 {
		return token
	}
	return token[:3] + strings.Repeat("*", len(token)-6) + token[len(token)-3:]
}