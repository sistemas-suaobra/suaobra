package repositories

import (
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type WhatsAppRepo struct {
	dao *daos.Dao
}

func NewWhatsAppRepo(dao *daos.Dao) *WhatsAppRepo {
	return &WhatsAppRepo{dao: dao}
}

func (r *WhatsAppRepo) FindByConexao(conexaoID string) (*models.Record, error) {
	conexaoID = strings.TrimSpace(conexaoID)
	if conexaoID == "" {
		return nil, nil
	}

	wa, _ := r.dao.FindFirstRecordByFilter(
		"conexoes_whatsapp",
		`conexoes ?= {:conexao}`,
		dbx.Params{"conexao": conexaoID},
	)
	if wa != nil && wa.Id != "" {
		return wa, nil
	}

	wa, err := r.dao.FindFirstRecordByFilter(
		"conexoes_whatsapp",
		`conexoes = {:conexao}`,
		dbx.Params{"conexao": conexaoID},
	)
	if err != nil {
		// fallback resiliente: alguns ambientes podem falhar no parser de filter
		// para relation multi. Nesses casos, varremos localmente.
		return r.findByConexaoFallback(conexaoID)
	}
	if wa != nil && wa.Id != "" {
		return wa, nil
	}

	return r.findByConexaoFallback(conexaoID)
}

func (r *WhatsAppRepo) findByConexaoFallback(conexaoID string) (*models.Record, error) {
	records, err := r.FindAll()
	if err != nil {
		return nil, err
	}

	for _, rec := range records {
		for _, id := range rec.GetStringSlice("conexoes") {
			if strings.TrimSpace(id) == conexaoID {
				return rec, nil
			}
		}
	}

	return nil, nil
}

func (r *WhatsAppRepo) FindAll() ([]*models.Record, error) {
	return r.dao.FindRecordsByFilter("conexoes_whatsapp", "id != ''", "", 0, 0, nil)
}

func (r *WhatsAppRepo) Delete(rec *models.Record) error {
	if rec == nil || rec.Id == "" {
		return nil
	}
	return r.dao.DeleteRecord(rec)
}

func (r *WhatsAppRepo) Create(conexaoID string, fields map[string]any) (*models.Record, error) {
	col, err := r.dao.FindCollectionByNameOrId("conexoes_whatsapp")
	if err != nil {
		return nil, err
	}

	wa := models.NewRecord(col)
	wa.Set("conexoes", []string{conexaoID}) // relation multi

	for k, v := range fields {
		wa.Set(k, v)
	}

	if wa.GetDateTime("ultimo_qr_em").Time().IsZero() {
		wa.Set("ultimo_qr_em", time.Now().UTC())
	}

	if err := r.dao.SaveRecord(wa); err != nil {
		return nil, err
	}
	return wa, nil
}

func (r *WhatsAppRepo) TouchUltimoQR(wa *models.Record) {
	wa.Set("ultimo_qr_em", time.Now().UTC())
	_ = r.dao.SaveRecord(wa)
}

// UpdateConnected marca a conexão como ativa: grava device_jid e conectado_em.
func (r *WhatsAppRepo) UpdateConnected(wa *models.Record, jid string) error {
	wa.Set("device_jid", jid)
	wa.Set("conectado_em", time.Now().UTC())
	return r.dao.SaveRecord(wa)
}

// ClearConnected limpa device_jid, conectado_em e ultimo_qr_em (desconexão total).
func (r *WhatsAppRepo) ClearConnected(wa *models.Record) error {
	wa.Set("device_jid", "")
	wa.Set("conectado_em", "")
	wa.Set("ultimo_qr_em", "")
	return r.dao.SaveRecord(wa)
}

// UpdateWebhookAndEvents persiste webhook/events após sync no wuzapi.
func (r *WhatsAppRepo) UpdateWebhookAndEvents(wa *models.Record, webhook, events string) error {
	if wa == nil || wa.Id == "" {
		return nil
	}
	wa.Set("webhook", webhook)
	wa.Set("events", events)
	return r.dao.SaveRecord(wa)
}

// FindByToken localiza um registro de conexoes_whatsapp pelo token do usuário (numero_e164).
func (r *WhatsAppRepo) FindByToken(token string) (*models.Record, error) {
	wa, err := r.dao.FindFirstRecordByFilter(
		"conexoes_whatsapp",
		`numero_e164 = {:token}`,
		dbx.Params{"token": token},
	)
	if err != nil {
		return nil, err
	}
	return wa, nil
}

// FindByInstanciaID localiza um registro de conexoes_whatsapp pelo instancia_id (userID do WUZAPI).
func (r *WhatsAppRepo) FindByInstanciaID(instanciaID string) (*models.Record, error) {
	wa, err := r.dao.FindFirstRecordByFilter(
		"conexoes_whatsapp",
		`instancia_id = {:instancia_id}`,
		dbx.Params{"instancia_id": instanciaID},
	)
	if err != nil {
		return nil, err
	}
	return wa, nil
}
