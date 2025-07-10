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
