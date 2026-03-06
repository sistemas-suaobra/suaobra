package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/cron"
	"github.com/suaobra/suaobra-app/server"
	"github.com/suaobra/suaobra-app/store"

	_ "github.com/suaobra/suaobra-app/store/migrations"
)

func init() {
	if strings.ToLower(os.Getenv("ENV")) == "development" {
		os.Setenv("DEBUG", "LOW")
	}
}

func main() {
	config := pocketbase.Config{
		DefaultDataDir: filepath.Join("data", "main"),
	}
	app := pocketbase.NewWithConfig(config)

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Admin UI
		Dir:         filepath.Join(store.Directory, "migrations"),
		Automigrate: true,
	})

	contextMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("app", app)
			return next(c) // proceed with the request chain
		}
	}

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {

		err := store.SetPocketBaseDB(app)
		g.LogFatal(err, "could not attach core.db")

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.Use(contextMiddleware)
		
		e.Router.GET("/query/obras-plus", server.QueryObrasPlus)
		e.Router.GET("/query/leads-plus", server.QueryLeadsPlus)
		e.Router.GET("/query/obras-plus-neighborhood", server.QueryCityNeighborhood)
		e.Router.GET("/query/obras-plus-export", server.QueryObrasPlusExport)
		e.Router.GET("/query/obras-plus-contacts", server.QueryObrasPlusContacts)
		e.Router.GET("/query/cities", server.QueryCities)

		e.Router.GET("/query/dashboard/funnel", server.QueryDashboardFunnel)
		e.Router.GET("/query/dashboard/history", server.QueryDashboardHistory)
		e.Router.GET("/query/dashboard/leads", server.QueryDashboardLeads)
		e.Router.GET("/query/dashboard/users", server.QueryDashboardUsers)
		e.Router.GET("/query/dashboard/campanhas", server.CampanhasDashboard)

		e.Router.GET("/query/crm/leads", server.QueryCrmLeads)
		e.Router.GET("/query/crm/contacts", server.QueryCrmContacts)
		e.Router.GET("/query/crm/search", server.QueryCrmSearch)

		e.Router.GET("/export/:table", server.ExportTable)

		e.Router.PATCH("/patch/lead-toggle", server.PatchLeadToggle)
		e.Router.PATCH("/patch/lead-owner", server.PatchLeadOwner)

		e.Router.POST("/messenger/sync", server.MessengerSync)
		e.Router.POST("/messenger/queue/submit", server.MessengerQueueSubmit)
		e.Router.POST("/messenger/queue/get", server.MessengerQueueGet)
		e.Router.POST("/messenger/existing", server.MessengerExisting)
		e.Router.GET("/messenger/user", server.MessengerGetOrCreateUser)
		e.Router.POST("/messenger/generate-templates", server.MessengerGenerateTemplates)
		e.Router.POST("/messenger/generate-lead-introduction", server.MessengerGenerateLeadIntroduction)

		// ROTAS DO MESTRE-IA:
		e.Router.POST("/conexoes/whatsapp", server.CriarConexaoWhatsapp)
		e.Router.POST("/conexoes/whatsapp/connect", server.ConectarSessaoWhatsapp)
		e.Router.GET("/conexoes/whatsapp/qr", server.ObterQRCodeWhatsapp)
		e.Router.GET("/conexoes/whatsapp/status", server.StatusConexaoWhatsapp)
		e.Router.POST("/conexoes/whatsapp/disconnect", server.DisconnectConexaoWhatsapp)
		e.Router.POST("/conexoes/whatsapp/send-test", server.EnviarMensagemTesteWhatsapp)
		e.Router.GET("/conexoes/whatsapp", server.ObterConexaoWhatsapp)
		e.Router.POST("/conexoes/whatsapp/fix-webhook", server.FixWebhookWhatsapp)

		// Email
		e.Router.POST("/conexoes/email", server.SalvarConexaoEmail)
		e.Router.GET("/conexoes/email", server.ObterConexaoEmail)
		e.Router.POST("/conexoes/email/send-test", server.EnviarEmailTeste)

		// Campanhas
		e.Router.POST("/campanhas/:id/iniciar", server.IniciarCampanha)
		e.Router.GET("/campanhas/:id/status", server.StatusCampanha)
		e.Router.POST("/campanhas/:id/pausar", server.PausarCampanha)
		e.Router.POST("/campanhas/:id/cancelar", server.CancelarCampanha)
		e.Router.POST("/campanhas/:id/enriquecer", server.EnriquecerDestinatarios)
		e.Router.PATCH("/campanha-destinatarios/marcar-enviado", server.MarcarDestinatariosEnviados)
		e.Router.POST("/campanhas/gerar-mensagem-ia", server.GerarMensagemCampanhaIA)

		// Agente IA
		server.RegisterAgenteIARoutes(e.Router)

		// Webhook do wuzapi — recebe eventos de sessão (Connected, Disconnected, etc.)
		e.Router.POST("/webhooks/whatsmeow", server.WebhookWhatsmeow)

		// Team management
		e.Router.GET("/team/members", server.TeamMembers)
		e.Router.POST("/team/invite", server.TeamInvite)
		e.Router.POST("/team/remove-member", server.TeamRemoveMember)
		e.Router.POST("/team/set-manager", server.TeamSetManager)
		
		// Debug endpoint para RudderStack (temporário)
		e.Router.POST("/debug/rudderstack", func(c echo.Context) error {
			token := c.Request().Header.Get("TOKEN")
			if token != "sua-obra-rudderstack" {
				return c.JSON(401, map[string]string{"error": "unauthorized"})
			}
			
			var payload map[string]interface{}
			if err := c.Bind(&payload); err != nil {
				return c.JSON(400, map[string]string{"error": "bind error: " + err.Error()})
			}
			
			// Log do payload recebido
			g.Info("DEBUG: Received payload: %v", payload)
			
			collection, err := app.Dao().FindCollectionByNameOrId("rudderstack")
			if err != nil {
				return c.JSON(400, map[string]string{"error": "collection not found: " + err.Error()})
			}
			
			record := models.NewRecord(collection)
			
			// Setar cada campo manualmente
			for key, value := range payload {
				record.Set(key, value)
			}
			
			// Tentar salvar
			if err := app.Dao().SaveRecord(record); err != nil {
				g.Error(err, "Failed to save record")
				return c.JSON(400, map[string]string{"error": "save error: " + err.Error()})
			}
			
			return c.JSON(200, map[string]interface{}{
				"success": true,
				"id":      record.Id,
				"record":  record,
			})
		})

		// e.Router.PATCH("/patch/crm/stage-lead-adjust", server.PatchStageLeadAdjust)
		// e.Router.PATCH("/patch/crm/stage-lead-add", server.PatchStageLeadAdd)
		// e.Router.PATCH("/patch/crm/stage-lead-remove", server.PatchStageLeadRemove)

		return nil
	})

	// Auto-sync webhook URL in wuzapi for all active WhatsApp connections on startup
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		go server.SyncWebhookURLs(app)
		return nil
	})

	// cron schedule - FIXED EMAIL SPAM BUG: improved SQL query + rate limiting
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		scheduler := cron.New()

		scheduler.Add("notify-reminder", "30 * * * *", func() {
			server.NotifyReminders(app)
		})

		scheduler.Start()

		return nil
	})

	// schedule.IsDue()

	// register hooks
	server.RegisterHooks(app)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
