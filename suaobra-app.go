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
		e.Router.GET("/query/obras-plus-neighborhood", server.QueryCityNeighborhood)
		e.Router.GET("/query/obras-plus-export", server.QueryObrasPlusExport)
		e.Router.GET("/query/obras-plus-contacts", server.QueryObrasPlusContacts)
		e.Router.GET("/query/cities", server.QueryCities)

		e.Router.GET("/query/dashboard/funnel", server.QueryDashboardFunnel)
		e.Router.GET("/query/dashboard/history", server.QueryDashboardHistory)
		e.Router.GET("/query/dashboard/leads", server.QueryDashboardLeads)
		e.Router.GET("/query/dashboard/users", server.QueryDashboardUsers)

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

		// Team management
		e.Router.GET("/team/members", server.TeamMembers)
		e.Router.POST("/team/invite", server.TeamInvite)
		e.Router.POST("/team/remove-member", server.TeamRemoveMember)
		e.Router.POST("/team/set-manager", server.TeamSetManager)

		// e.Router.PATCH("/patch/crm/stage-lead-adjust", server.PatchStageLeadAdjust)
		// e.Router.PATCH("/patch/crm/stage-lead-add", server.PatchStageLeadAdd)
		// e.Router.PATCH("/patch/crm/stage-lead-remove", server.PatchStageLeadRemove)

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
