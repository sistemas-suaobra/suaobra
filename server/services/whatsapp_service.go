package services

import (
	"strings"
	"time"

	"github.com/flarco/g"
	"github.com/pocketbase/pocketbase/models"

	"github.com/suaobra/suaobra-app/server/clients/wuzapi"
	"github.com/suaobra/suaobra-app/server/config"
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

// Idempotente POR USUÁRIO: se o próprio usuário já tiver a sua conexão, retorna.
// Importante: NÃO usamos o fallback legado aqui — assim, um usuário sem número
// próprio consegue criar o seu, mesmo que o time tenha uma conexão compartilhada.
func (s *WhatsAppService) CreateConnection(teamID, userID, name, apiKey string) (already bool, con *models.Record, wa *models.Record, err error) {
	exCon, _ := s.conRepo.FindActiveWhatsappByOwner(teamID, userID)
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
		con, err = s.conRepo.CreateWhatsapp(teamID, userID, name)
		if err != nil {
			return false, nil, nil, err
		}
	}

	fields := map[string]any{
		"provider":     "WUZAPI",
		"api_base_url": s.wuzBaseURL(),
		"numero_e164":  userToken,
		"instancia_id": created.Data.ID,
		"device_jid":   "",
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

func (s *WhatsAppService) resolveWhatsappConnection(teamID, userID string) (owned bool, con *models.Record, err error) {
	con, _ = s.conRepo.FindActiveWhatsappByOwner(teamID, userID)
	if con != nil && con.Id != "" {
		return true, con, nil
	}

	con = s.findBestTeamWhatsappFallback(teamID)
	if con == nil || con.Id == "" {
		return false, nil, nil
	}
	return false, con, nil
}

// findBestTeamWhatsappFallback escolhe a conexão WhatsApp do time realmente conectada.
// Evita exibir a conexão legada (user_id vazio) obsoleta quando um colega já conectou o dele.
func (s *WhatsAppService) findBestTeamWhatsappFallback(teamID string) *models.Record {
	type candidate struct {
		con *models.Record
		at  time.Time
	}

	var best *candidate

	pick := func(con *models.Record) {
		if con == nil || con.Id == "" {
			return
		}
		wa, err := s.waRepo.FindByConexao(con.Id)
		if err != nil || wa == nil || wa.Id == "" {
			return
		}
		if wa.GetDateTime("conectado_em").Time().IsZero() || strings.TrimSpace(wa.GetString("device_jid")) == "" {
			return
		}
		at := wa.GetDateTime("conectado_em").Time()
		if best == nil || at.After(best.at) {
			best = &candidate{con: con, at: at}
		}
	}

	if legacy, _ := s.conRepo.FindActiveWhatsappLegacy(teamID); legacy != nil {
		pick(legacy)
	}

	if cons, err := s.conRepo.FindAllActiveWhatsappByTeam(teamID); err == nil {
		for _, con := range cons {
			pick(con)
		}
	}

	if best != nil {
		return best.con
	}

	// Sem sessão autenticada no time: mantém legado só para permitir criar/conectar QR.
	if legacy, _ := s.conRepo.FindActiveWhatsappLegacy(teamID); legacy != nil {
		return legacy
	}
	return nil
}

func (s *WhatsAppService) resolveOwnedWhatsappConnection(teamID, userID string) (*models.Record, error) {
	con, _ := s.conRepo.FindActiveWhatsappByOwner(teamID, userID)
	if con == nil || con.Id == "" {
		return nil, g.Error("whatsapp connection not found")
	}
	return con, nil
}

// getTokenAndRecordForManage resolve apenas a conexão PRÓPRIA (connect/QR/disconnect).
func (s *WhatsAppService) getTokenAndRecordForManage(teamID, userID string) (token string, wa *models.Record, err error) {
	con, err := s.resolveOwnedWhatsappConnection(teamID, userID)
	if err != nil {
		return "", nil, err
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

// getTokenAndRecordByOwner resolve a conexão para envio (própria ou conectada do time).
func (s *WhatsAppService) getTokenAndRecordByOwner(teamID, userID string) (token string, wa *models.Record, err error) {
	_, con, err := s.resolveWhatsappConnection(teamID, userID)
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

func (s *WhatsAppService) ConnectByUser(teamID, userID string) (map[string]any, error) {
	token, wa, err := s.getTokenAndRecordForManage(teamID, userID)
	if err != nil {
		return nil, err
	}

	raw, err := s.ConnectSession(token)
	if err != nil {
		return raw, err
	}

	// Wuzapi pode limpar o campo Events ao conectar/reconectar.
	// Garantir que Events=All e webhook estejam configurados após cada conexão.
	instanciaID := ""
	if wa != nil {
		instanciaID = strings.TrimSpace(wa.GetString("instancia_id"))
	}
	if instanciaID != "" {
		go func(iid string) {
			// Pequeno delay para dar tempo ao wuzapi processar a conexão
			time.Sleep(3 * time.Second)

			cfg := config.NewWhatsMeowConfig()
			webhookURL := cfg.WebhookURL()
			if webhookURL == "" {
				return
			}

			if updateErr := s.wuz.UpdateAdminUser(iid, map[string]any{
				"webhook": webhookURL,
				"events":  "All",
			}); updateErr != nil {
				g.Warn("ConnectByUser: falha ao garantir Events=All instancia=%s: %v", iid, updateErr)
			} else {
				g.Info("ConnectByUser: Events=All garantido após conexão instancia=%s", iid)
			}
		}(instanciaID)
	}

	return raw, nil
}

func (s *WhatsAppService) GetQRByUser(teamID, userID string) (code int, qr string, err error) {
	token, wa, err := s.getTokenAndRecordForManage(teamID, userID)
	if err != nil {
		return 0, "", err
	}

	parsed, err := s.GetQR(token)
	if err != nil && isWuzNoSessionErr(err) {
		if _, connectErr := s.ConnectSession(token); connectErr != nil {
			return 0, "", connectErr
		}
		time.Sleep(900 * time.Millisecond)
		parsed, err = s.GetQR(token)
	}
	if err != nil {
		return 0, "", err
	}

	// atualiza timestamp do QR
	s.waRepo.TouchUltimoQR(wa)

	return parsed.Code, parsed.Data.QRCode, nil
}

func isWuzNoSessionErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no session") || strings.Contains(msg, "/session/qr failed")
}

func (s *WhatsAppService) wuzBaseURL() string {
	return s.wuz.BaseURL()
}

// SendFromToken envia uma mensagem usando DIRETAMENTE o token (numero_e164) de
// uma instância específica. Usado quando já sabemos por qual número responder
// (ex.: a IA responde pela mesma instância que recebeu a mensagem).
func (s *WhatsAppService) SendFromToken(token, phone, body string) (map[string]any, error) {
	startedAt := time.Now()
	token = strings.TrimSpace(token)
	phone = strings.TrimSpace(phone)
	body = strings.TrimSpace(body)

	if token == "" {
		return nil, g.Error("numero_e164 token is empty")
	}

	g.Info(
		"WhatsAppService.SendFromToken: start token=%s phone=%s body_len=%d body=%s",
		maskWAToken(token),
		maskWAPhone(phone),
		len(body),
		truncateWALog(body, 800),
	)

	resp, err := s.wuz.SendTextMessage(token, phone, body)
	if err != nil {
		g.Error(
			err,
			"WhatsAppService.SendFromToken: falha no WUZAPI token=%s phone=%s",
			maskWAToken(token),
			maskWAPhone(phone),
		)
		return nil, err
	}

	g.Info(
		"WhatsAppService.SendFromToken: success token=%s phone=%s duration=%s response=%s",
		maskWAToken(token),
		maskWAPhone(phone),
		time.Since(startedAt),
		truncateWALog(g.Marshal(resp), 1500),
	)

	return resp, nil
}

// SendForOwner resolve a conexão do dono (owner; senão a legada do time) e
// envia pela instância dele. Usado pelas campanhas (criado_por) e pela IA quando
// não há token de instância específico.
func (s *WhatsAppService) SendForOwner(teamID, ownerUserID, phone, body string) (map[string]any, error) {
	token, wa, err := s.getTokenAndRecordByOwner(teamID, ownerUserID)
	if err != nil {
		g.Error(err, "WhatsAppService.SendForOwner: erro ao obter token team=%s owner=%s", teamID, ownerUserID)
		return nil, err
	}
	if wa != nil {
		g.Info(
			"WhatsAppService.SendForOwner: resolved team=%s owner=%s instancia_id=%s",
			teamID, ownerUserID, strings.TrimSpace(wa.GetString("instancia_id")),
		)
	}
	return s.SendFromToken(token, phone, body)
}

// SendTestMessage envia a mensagem de teste pela conexão do próprio usuário.
func (s *WhatsAppService) SendTestMessage(teamID, userID, phone, body string) (map[string]any, error) {
	return s.SendForOwner(teamID, userID, phone, body)
}

// CheckAndUpdateStatus consulta o wuzapi para saber se o user está connected,
// e atualiza conectado_em + device_jid no registro local se necessário.
// Retorna (connected, jid, error).
func (s *WhatsAppService) CheckAndUpdateStatus(teamID, userID string) (connected bool, jid string, err error) {
	_, con, err := s.resolveWhatsappConnection(teamID, userID)
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
		// Não está logado de fato. Como o wuzapi respondeu sem erro, isto significa
		// que a sessão realmente caiu/foi deslogada. Auto-curamos o DB local para
		// não ficar mostrando "conectado" de forma desatualizada.
		if !wa.GetDateTime("conectado_em").Time().IsZero() || strings.TrimSpace(wa.GetString("device_jid")) != "" {
			if clearErr := s.waRepo.ClearConnected(wa); clearErr != nil {
				g.Warn("erro ao limpar conexão obsoleta instancia=%s: %v", instanciaID, clearErr)
			} else {
				g.Info("conexão obsoleta limpa para instancia=%s (wuzapi reportou LoggedIn=false)", instanciaID)
			}
		}
		return false, "", nil
	}

	// Logado! Persistimos conectado_em + device_jid SEMPRE que o wuzapi confirmar
	// que a sessão está autenticada. Isto torna o estado durável (a tela continua
	// mostrando "conectado" após dias) e independe do webhook "connected" chegar.
	conectadoEm := wa.GetDateTime("conectado_em")
	if conectadoEm.Time().IsZero() || wa.GetString("device_jid") != info.JID {
		if saveErr := s.waRepo.UpdateConnected(wa, info.JID); saveErr != nil {
			g.Error(saveErr, "erro ao salvar conectado_em para instancia=%s", instanciaID)
		} else {
			g.Info("conectado_em salvo para instancia=%s jid=%s (LOGADO)", instanciaID, info.JID)
		}
	}

	return true, info.JID, nil
}

func (s *WhatsAppService) Disconnect(teamID, userID string) error {
	token, wa, err := s.getTokenAndRecordForManage(teamID, userID)
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

func (s *WhatsAppService) DeleteConnection(teamID, userID string) error {
	con, err := s.resolveOwnedWhatsappConnection(teamID, userID)
	if err != nil {
		return err
	}
	if con == nil || con.Id == "" {
		return nil
	}

	wa, err := s.waRepo.FindByConexao(con.Id)
	if err != nil {
		return err
	}

	if wa != nil && wa.Id != "" {
		instanciaID := strings.TrimSpace(wa.GetString("instancia_id"))
		if instanciaID != "" {
			if err := s.wuz.DeleteAdminUser(instanciaID); err != nil {
				return err
			}
		}

		if err := s.waRepo.Delete(wa); err != nil {
			return err
		}
	}

	s.conRepo.Delete(con)
	return nil
}

// GetByUser resolve a conexão exibida na tela de Conexões.
// owned indica se a conexão pertence ao próprio usuário (true) ou é a
// legada/compartilhada do time emprestada como fallback (false). A UI usa isso
// para permitir que um colega conecte o número dele mesmo estando "emprestado".
func (s *WhatsAppService) GetByUser(teamID, userID string) (exists bool, owned bool, con *models.Record, wa *models.Record, err error) {
	var ok bool
	ok, con, err = s.resolveWhatsappConnection(teamID, userID)
	if err != nil {
		return false, false, nil, nil, err
	}
	owned = ok
	if con == nil || con.Id == "" {
		return false, false, nil, nil, nil
	}

	wa, err = s.waRepo.FindByConexao(con.Id)
	if err != nil {
		return true, owned, con, nil, err
	}
	if wa == nil || wa.Id == "" {
		return true, owned, con, nil, nil
	}

	return true, owned, con, wa, nil
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
