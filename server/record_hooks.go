package server

import (
	"github.com/flarco/g"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/spf13/cast"
	"github.com/suaobra/suaobra-app/store"
)

func RegisterHooks(app *pocketbase.PocketBase) {
	app.OnModelBeforeCreate().Add(func(e *core.ModelEvent) error {
		e.Model.SetId(store.ID(e.Model.TableName(), e.Model.GetId()))
		return nil
	})

	registerUserHooks(app)
	registerTeamHooks(app)
	registerListHooks(app)
	registerListLeadHooks(app)
	registerRudderstackHooks(app)
}

func registerUserHooks(app *pocketbase.PocketBase) {

	// After CREATE user via API
	app.OnRecordAfterCreateRequest("user").Add(func(e *core.RecordCreateEvent) error {

		// send notification email of user joining
		go func() {
			err := sendNotificationSignupToAdmins(app, e.Record)
			if !g.LogError(err, "could not send signup notification email") {
				g.Info("sent signup notification email for: %s ", e.Record.Email())
			}
		}()

		// create default team
		teamCollection, err := app.Dao().FindCollectionByNameOrId("team")
		if err != nil {
			return g.Error(err, "could not find collection team")
		}

		teamRecord, _ := app.Dao().FindFirstRecordByData("team", "name", e.Record.Get("email"))
		if teamRecord == nil {
			teamRecord = models.NewRecord(teamCollection)
			teamRecord.SetId(store.ID(teamCollection.Name, ""))
			teamRecord.Set("properties", g.Marshal(g.M()))
			teamRecord.Set("entitlements", g.Marshal(g.M()))
			teamRecord.Set("blocked", false)
			teamRecord.Set("cities", "[]")
			teamRecord.Set("name", e.Record.Get("email"))
		}

		if cast.ToString(teamRecord.Get("owner_id")) == "" {
			teamRecord.Set("owner_id", e.Record.GetId())
		}

		teamRecord.Set("active", true)
		if err := app.Dao().SaveRecord(teamRecord); err != nil {
			return g.Error(err, "could not save new team")
		}

		// set the user's team if missing
		if e.Record.Get("team_id") == "" || e.Record.Get("emailVisibility") == false {
			e.Record.Set("team_id", teamRecord.GetId())
			e.Record.Set("emailVisibility", true)
			if err := app.Dao().SaveRecord(e.Record); err != nil {
				return g.Error(err, "could not save user team")
			}
		}

		return nil
	})
}

func registerTeamHooks(app *pocketbase.PocketBase) {

	// After CREATE team
	app.OnModelAfterCreate("team").Add(func(e *core.ModelEvent) error {

		// create default list
		listCollection, err := app.Dao().FindCollectionByNameOrId("list")
		if err != nil {
			return g.Error(err, "could not find collection")
		}

		listRecord := models.NewRecord(listCollection)
		properties := g.M()
		listRecord.SetId(store.ID(listCollection.Name, ""))
		listRecord.Set("team_id", e.Model.GetId())
		listRecord.Set("name", "Lista Padrão")
		listRecord.Set("properties", g.Marshal(properties))

		if err := app.Dao().SaveRecord(listRecord); err != nil {
			return g.Error(err, "could not save new list")
		}

		return nil
	})

}

func registerListHooks(app *pocketbase.PocketBase) {

	// After CREATE list
	app.OnModelAfterCreate("list").Add(func(e *core.ModelEvent) error {

		// create default stages
		stageCollection, err := app.Dao().FindCollectionByNameOrId("list_stage")
		if err != nil {
			return g.Error(err, "could not find collection")
		}

		defaultStages := []string{
			"Prospecção",
			// "Visita",
			"Contato",
			"Proposta",
			// "Negociação",
			"Fechado",
		}
		for i, name := range defaultStages {
			stageRecord := models.NewRecord(stageCollection)
			stageRecord.Set("list_id", e.Model.GetId())
			stageRecord.Set("name", name)
			stageRecord.Set("order", i+1)
			stageRecord.Set("properties", "{}")

			if err := app.Dao().SaveRecord(stageRecord); err != nil {
				return g.Error(err, "could not save new list")
			}
		}

		return nil
	})

}

func registerListLeadHooks(app *pocketbase.PocketBase) {

	// After UPDATE list_lead via API
	app.OnRecordAfterUpdateRequest("list_lead").Add(func(e *core.RecordUpdateEvent) error {
		req := NewRequest(e.HttpContext)

		leadActivityCollection, err := app.Dao().FindCollectionByNameOrId("lead_activity")
		if g.LogError(err) {
			return g.Error(err, "could not find collection")
		}

		// get activity type
		activityType := ""
		properties := g.M()
		if stageID := req.Payload.String("stage_id"); stageID != "" {
			activityType = "history"
			properties["stage_id"] = req.Payload.String("stage_id")
		}

		leadActivityRecord := models.NewRecord(leadActivityCollection)
		leadActivityRecord.Set("team_id", req.TeamID())
		leadActivityRecord.Set("lead_id", e.Record.Get("lead_id"))
		leadActivityRecord.Set("actor_email", req.Payload.String("user_email"))
		leadActivityRecord.Set("type", activityType)
		leadActivityRecord.Set("properties", g.Marshal(properties))

		if err := app.Dao().SaveRecord(leadActivityRecord); g.LogError(err) {
			return g.Error(err, "could not save new list_activity")
		}

		return nil
	})
}

func registerRudderstackHooks(app *pocketbase.PocketBase) {
	
	// Otimização: Processar eventos do RudderStack de forma assíncrona
	// para evitar timeout nas requisições
	app.OnRecordAfterCreateRequest("rudderstack").Add(func(e *core.RecordCreateEvent) error {
		// Processar de forma assíncrona para não bloquear a resposta
		go func() {
			// Aqui você pode adicionar processamento adicional se necessário
			// Por exemplo: enviar para fila, processar analytics, etc.
			// Por enquanto apenas logamos o evento
			g.Debug("Received rudderstack event: type=%s, event=%s, messageId=%s", 
				e.Record.GetString("type"), 
				e.Record.GetString("event"),
				e.Record.GetString("messageId"))
		}()
		
		// Retornar imediatamente sem aguardar processamento
		return nil
	})
}

func getModelRecord(app *pocketbase.PocketBase, e *core.ModelEvent) (*models.Record, error) {
	record, err := app.Dao().FindFirstRecordByData(e.Model.TableName(), "id", e.Model.GetId())
	if err != nil {
		err = g.Error(err, "could not get %s with id %s", e.Model.TableName(), e.Model.GetId())
	}

	return record, err
}
