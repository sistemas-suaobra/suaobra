package server

import (
	"io"
	"os"
	"time"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/suaobra/suaobra-app/store"
)

func QueryDashboardUsers(c echo.Context) error {
	req := NewRequest(c)
	sqlBytes, _ := templates.ReadFile("templates/dashboard/dashboard_users.sql")
	sql := string(sqlBytes)
	return req.SqlQueryResponse(sql)
}

func QueryDashboardFunnel(c echo.Context) error {
	req := NewRequest(c)
	sqlBytes, _ := templates.ReadFile("templates/dashboard/dashboard_funnel.sql")
	sql := string(sqlBytes)

	// Build where clause
	whereClause := req.BuildWhereClause("l.team_id = {:team_id}", map[string]string{
		"l.owner_id": "user_id",
	})

	// Replace the where_clause placeholder
	sql = g.R(sql, "where_clause", whereClause)

	return req.SqlQueryResponse(sql)
}

func QueryDashboardHistory(c echo.Context) error {
	req := NewRequest(c)
	sqlBytes, _ := templates.ReadFile("templates/dashboard/dashboard_history.sql")
	sql := string(sqlBytes)

	// Build where clauses for both visits and leads
	whereClauseVisits := req.BuildWhereClause("team_id = {:team_id}", map[string]string{
		"owner_id": "user_id",
	})

	whereClauseLeads := req.BuildWhereClause("team_id = {:team_id}", map[string]string{
		"owner_id": "user_id",
	})

	// Replace the where_clause placeholders
	sql = g.R(sql, "where_clause_visits", whereClauseVisits)
	sql = g.R(sql, "where_clause_leads", whereClauseLeads)

	return req.SqlQueryResponse(sql)
}

func dashboardOwnerFilter(req *Request, column string) string {
	userID := req.Payload.String("user_id")
	if userID != "" && userID != "null" && userID != "undefined" {
		return column + " = {:user_id}"
	}
	return "1=1"
}

func queryDashboardList(req *Request, file, ownerColumn string) ([]map[string]any, error) {
	sqlBytes, err := templates.ReadFile(file)
	if err != nil {
		return nil, err
	}
	sql := string(sqlBytes)
	if ownerColumn != "" {
		sql = g.R(sql, "user_filter", dashboardOwnerFilter(req, ownerColumn))
	}
	return req.SqlQueryRecords(sql)
}

func QueryDashboardHome(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error)
	}

	reminders, err := queryDashboardList(&req, "templates/dashboard/dashboard_reminders.sql", "lead.owner_id")
	if err != nil {
		return ErrJSON(500, err, "error querying reminders")
	}
	unassigned, err := queryDashboardList(&req, "templates/dashboard/dashboard_unassigned.sql", "")
	if err != nil {
		return ErrJSON(500, err, "error querying unassigned")
	}
	moves, err := queryDashboardList(&req, "templates/dashboard/dashboard_moves.sql", "lead.owner_id")
	if err != nil {
		return ErrJSON(500, err, "error querying moves")
	}

	return c.JSON(200, g.M(
		"reminders", reminders,
		"unassigned", unassigned,
		"moves", moves,
	))
}

// QueryDashboardBootstrap agrega funnel + home (+ users se manager)
// numa única ida à API — o BFF da Vercel deixa de pagar 3 RTTs BR.
func QueryDashboardBootstrap(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error)
	}

	funnelSQL, err := templates.ReadFile("templates/dashboard/dashboard_funnel.sql")
	if err != nil {
		return ErrJSON(500, err, "error reading funnel sql")
	}
	whereClause := req.BuildWhereClause("l.team_id = {:team_id}", map[string]string{
		"l.owner_id": "user_id",
	})
	funnel, err := req.SqlQueryRecords(g.R(string(funnelSQL), "where_clause", whereClause))
	if err != nil {
		return ErrJSON(500, err, "error querying funnel")
	}

	reminders, err := queryDashboardList(&req, "templates/dashboard/dashboard_reminders.sql", "lead.owner_id")
	if err != nil {
		return ErrJSON(500, err, "error querying reminders")
	}
	unassigned, err := queryDashboardList(&req, "templates/dashboard/dashboard_unassigned.sql", "")
	if err != nil {
		return ErrJSON(500, err, "error querying unassigned")
	}
	moves, err := queryDashboardList(&req, "templates/dashboard/dashboard_moves.sql", "lead.owner_id")
	if err != nil {
		return ErrJSON(500, err, "error querying moves")
	}

	payload := g.M(
		"funnel", funnel,
		"reminders", reminders,
		"unassigned", unassigned,
		"moves", moves,
	)

	if req.IsManager() {
		usersSQL, err := templates.ReadFile("templates/dashboard/dashboard_users.sql")
		if err != nil {
			return ErrJSON(500, err, "error reading users sql")
		}
		users, err := req.SqlQueryRecords(string(usersSQL))
		if err != nil {
			return ErrJSON(500, err, "error querying users")
		}
		payload["users"] = users
	}

	return c.JSON(200, payload)
}

func QueryDashboardLeads(c echo.Context) error {
	req := NewRequest(c)
	sqlBytes, _ := templates.ReadFile("templates/dashboard/dashboard_leads.sql")
	sql := string(sqlBytes)

	// Build where clause
	whereClause := req.BuildWhereClause("l.team_id = {:team_id}", map[string]string{
		"l.owner_id": "user_id",
	})

	// Replace the where_clause placeholder
	sql = g.R(sql, "where_clause", whereClause)

	return req.SqlQueryResponse(sql)
}

func ExportTable(c echo.Context) error {
	// req := NewRequest(c)
	// if req.Admin() == nil {
	// 	return ErrJSON(404, g.Error("not allowed"))
	// }

	table := c.PathParam("table")
	if !g.In(table, "vw_eventos_diarios", "vw_eventos") {
		return ErrJSON(404, g.Error("bad request"))
	}

	ds, err := store.MainDB.StreamRows(g.F("select * from %s", table))
	if err != nil {
		return ErrJSON(500, g.Error("errored while querying"))
	}

	fileName := g.F("%s.%s.csv", table, g.NowFileStr())
	filePath := g.F("/tmp/%s", fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return ErrJSON(500, g.Error(err, "Unable to create temp file"))
	}

	_, err = io.Copy(file, ds.NewCsvReader(0, 0))
	if err != nil {
		return ErrJSON(500, g.Error(err, "Unable to write to temp file"))
	}

	file.Close()

	time.AfterFunc(1*time.Minute, func() { os.Remove(filePath) })
	g.Debug("exporting %s", filePath)

	c.Response().Header().Add("Content-disposition", g.F("attachment; filename=%s", fileName))
	return c.File(filePath)
}
