package server

import (
	"strings"
	"unicode"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/models"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var accentTransformer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

func ReplaceAccents(val string) (string, error) {
	newVal, _, err := transform.String(accentTransformer, val)
	if err != nil {
		return val, g.Error(err, "could not transform while running ReplaceAccents")
	}
	return newVal, nil
}

func QueryCrmLeads(c echo.Context) error {
	req := NewRequest(c)
	if _, ok := req.Payload["list_lead_id"]; !ok {
		req.Payload["list_lead_id"] = "-1" // required in query
	}
	if _, ok := req.Payload["stage_id"]; !ok {
		req.Payload["stage_id"] = "-1" // required in query
	}

	sqlB, _ := templates.ReadFile("templates/crm/crm_leads.sql")
	sql := string(sqlB)

	if filterUserID, ok := req.Payload["filter_user_id"]; ok && filterUserID != "" {
		sql = g.R(sql, "user_filter", "lead.owner_id = {:filter_user_id}")
	} else if req.IsManager() {
		sql = g.R(sql, "user_filter", "1=1")
	} else {
		sql = g.R(sql, "user_filter", "lead.owner_id = {:user_id}")
	}

	return req.SqlQueryResponse(sql)
}

func QueryCrmContacts(c echo.Context) error {
	req := NewRequest(c)
	sql, _ := templates.ReadFile("templates/crm/crm_contacts.sql")
	return req.SqlQueryResponse(string(sql))
}

func QueryCrmSearch(c echo.Context) error {
	req := NewRequest(c)
	sql, _ := templates.ReadFile("templates/crm/crm_search.sql")

	words := strings.Split(req.Payload.String("query"), " ")
	filters := []string{}

	for _, word := range words {
		word = strings.TrimSpace(strings.ToUpper(word))
		if word == "" {
			continue
		}

		word, _ = ReplaceAccents(word)

		filter := g.R(
			"(cop.address like '{expr}' or cop.owner like '{expr}' or cop.professional like '{expr}')",
			"expr",
			"%"+word+"%",
		)
		filters = append(filters, filter)
	}

	sqlStr := g.R(string(sql), "filters", strings.Join(filters, " \n  AND "))

	if req.IsManager() {
		sqlStr = g.R(sqlStr, "user_filter", "1=1")
	} else {
		sqlStr = g.R(sqlStr, "user_filter", "l.owner_id = {:user_id}")
	}

	return req.SqlQueryResponse(sqlStr)
}

func PatchStageLeadAdjust(c echo.Context) error {
	req := NewRequest(c)
	sql, _ := templates.ReadFile("templates/crm/crm_stage_lead_adjust.sql")
	return req.SqlExecuteResponse(string(sql))
}

func PatchStageLeadAdd(c echo.Context) error {
	req := NewRequest(c)
	sql, _ := templates.ReadFile("templates/crm/crm_stage_lead_add.sql")
	return req.SqlExecuteResponse(string(sql))
}

func PatchStageLeadRemove(c echo.Context) error {
	req := NewRequest(c)
	sql, _ := templates.ReadFile("templates/crm/crm_stage_lead_remove.sql")
	return req.SqlExecuteResponse(string(sql))
}

func addLeadToVendaMaisList(req *Request) (err error) {
	app := req.App
	record := req.Record

	// check if exists already in a list
	listLeadRecord, err := app.Dao().FindFirstRecordByData("list_lead", "lead_id", record.GetId())
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return g.Error(err, "could not query existing lead for lead %s", record.GetId())
	} else if listLeadRecord != nil && listLeadRecord.Id != "" {
		return nil // already exists
	}

	// create list lead and assign to default list
	listLeadCollection, err := app.Dao().FindCollectionByNameOrId("list_lead")
	if err != nil {
		return g.Error(err, "could not find collection")
	}

	// get default list for team
	teamID := record.Get("team_id")
	listRecord, err := app.Dao().FindFirstRecordByData("list", "team_id", teamID)
	if err != nil {
		return g.Error(err, "could not find matching default list for team %s", teamID)
	}
	// get first stage list for team
	where := dbx.NewExp(
		`list_id = {:list_id} and "order" = 1`,
		g.M("list_id", listRecord.GetId()),
	)
	stageRecords, err := app.Dao().FindRecordsByExpr("list_stage", where)
	if err != nil {
		return g.Error(err, "could not find matching default stage for team %s", teamID)
	}
	if len(stageRecords) > 1 {
		return g.Error("found many stage for for list %s", listRecord.GetId())
	}

	listLeadRecord = models.NewRecord(listLeadCollection)
	properties := g.M()
	listLeadRecord.Set("lead_id", record.GetId())
	listLeadRecord.Set("list_id", listRecord.GetId())
	listLeadRecord.Set("stage_id", stageRecords[0].GetId())
	listLeadRecord.Set("properties", g.Marshal(properties))

	if err := app.Dao().SaveRecord(listLeadRecord); err != nil {
		return g.Error(err, "could not save new list_lead")
	}

	req.Payload["list_lead_id"] = listLeadRecord.Id

	return nil
}

// PatchLeadOwner updates the owner_id of a lead
func PatchLeadOwner(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}

	// Validate parameters
	if err := req.ValidatePayload("lead_id", "owner_id"); err != nil {
		return ErrJSON(400, err, "missing required parameters")
	}

	leadID := req.Payload.String("lead_id")
	ownerID := req.Payload.String("owner_id")

	_, err := req.Dao().DB().
		Update("lead", g.M("owner_id", ownerID), dbx.NewExp("id = {:id}", dbx.Params{"id": leadID})).
		Execute()
	if err != nil {
		return ErrJSON(500, err, "could not update lead owner")
	}

	// Get the updated user record for response
	ownerRecord, err := req.Dao().FindRecordById("user", ownerID)
	if err != nil {
		// We updated the lead successfully, but couldn't get user details
		return c.JSON(200, g.M(
			"success", true,
			"owner_id", ownerID,
		))
	}

	return c.JSON(200, g.M(
		"success", true,
		"owner_id", ownerID,
		"owner_email", ownerRecord.Get("email"),
		"owner_name", ownerRecord.Get("name"),
	))
}
