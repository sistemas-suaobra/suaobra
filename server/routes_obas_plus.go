package server

import (
	"bytes"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/flarco/dbio/filesys"
	"github.com/flarco/dbio/iop"
	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/suaobra/suaobra-app/store"
	"golang.org/x/exp/slices"
)

type RecordObrasPlus struct {
	Owner        string `json:"owner"`
	Professional string `json:"professional"`
	Address      string `json:"address"`
	Size         int    `json:"size"`
	ObraNumber   int    `json:"obra_number"`
	PhoneNumbers string `json:"phone_numbers"`
	Type         string `json:"type"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
}

func validateCity(city string) bool {
	_, ok := store.ObrasCities[city]
	return ok
}

func validateFilter(filter string) bool {
	filter = strings.ToLower(filter)
	switch {
	case strings.Contains(filter, "'"):
		return false
	}
	return true
}

func makeFilterCond(filter string) string {
	if filter == "" {
		return "1=1"
	}

	whereArr := []string{
		g.F("owner like '%s'", "%"+filter+"%"),
		g.F("professional like '%s'", "%"+filter+"%"),
		g.F("address like '%s'", "%"+filter+"%"),
		g.F("type like '%s'", "%"+filter+"%"),
	}
	return "(" + strings.Join(whereArr, " or ") + ")"
}

func makeNeighborhoodCond(neighborhoods []string) string {
	if len(neighborhoods) == 0 || (len(neighborhoods) == 1 && neighborhoods[0] == "") {
		return "1=1"
	}

	whereArr := lo.Map(neighborhoods, func(n string, i int) string {
		return g.F("cop.bairro = '%s'", n)
	})
	return "(" + strings.Join(whereArr, " or ") + ")"
}

func makeStatusCond(statuses []string) string {
	const (
		StatusAndamento               string = "em-andamento"
		StatusComTelefone             string = "com-telefone"
		StatusComTelefoneProprietario string = "com-telefone-proprietario"
		StatusComTelefoneProfissional string = "com-telefone-profissional"
		StatusComEmail                string = "com-email"
		StatusContactado              string = "contactado"
		StatusNaoContactado           string = "nao-contactado"
		StatusContactoPendente        string = "contato-pendente"
		StatusComObservacao           string = "com-observacao"
		StatusVisitada                string = "ja-visitada"
		StatusNaoVisitada             string = "nao-visitada"
		StatusFavorita                string = "favorita"
		StatusExcluida                string = "excluida"
	)

	whereArr := []string{}

	if slices.Contains(statuses, StatusAndamento) {
		whereArr = append(whereArr, "(current_date >= start_date and current_date <= end_date )")
	}

	if slices.Contains(statuses, StatusComTelefone) {
		whereArr = append(whereArr, "( nullif(cop.has_owner_phone, '') or nullif(cop.has_professional_phone, '') )")
	}

	if slices.Contains(statuses, StatusComTelefoneProprietario) {
		whereArr = append(whereArr, "( nullif(cop.has_owner_phone, '') )")
	}

	if slices.Contains(statuses, StatusComTelefoneProfissional) {
		whereArr = append(whereArr, "( nullif(cop.has_professional_phone, '') )")
	}

	if slices.Contains(statuses, StatusComEmail) {
		whereArr = append(whereArr, "( nullif(cop.has_owner_email, '') or nullif(cop.has_professional_email, '') )")
	}

	if slices.Contains(statuses, StatusVisitada) {
		whereArr = append(whereArr, "( cop.id in ( select obra_id from main.lead where team_id = '{teamId}' and nullif(visited_at, '') is not null ) )")
	}

	if slices.Contains(statuses, StatusNaoVisitada) {
		whereArr = append(whereArr, "( cop.id not in ( select obra_id from main.lead where team_id = '{teamId}' and nullif(visited_at, '') is not null ) )")
	}

	if slices.Contains(statuses, StatusContactado) {
		whereArr = append(whereArr, "( cop.id in ( select obra_id from main.lead where team_id = '{teamId}' and ( nullif(owner_contacted_at, '') is not null or nullif(owner_contact_pending_at, '') is not null or nullif(professional_contacted_at, '') is not null or nullif(professional_contact_pending_at, '') is not null ) ) )")
	}

	if slices.Contains(statuses, StatusNaoContactado) {
		whereArr = append(whereArr, "( cop.id not in ( select obra_id from main.lead where team_id = '{teamId}' and ( nullif(owner_contact_pending_at, '') is not null or nullif(professional_contact_pending_at, '') is not null ) ) )")
	}

	if slices.Contains(statuses, StatusContactoPendente) {
		whereArr = append(whereArr, "( cop.id in ( select obra_id from main.lead where team_id = '{teamId}' and ( (nullif(owner_contact_pending_at, '') is not null and nullif(owner_contacted_at, '') is null ) or (nullif(professional_contact_pending_at, '') is not null and nullif(professional_contacted_at, '') is null ) ) ) )")
	}

	if slices.Contains(statuses, StatusFavorita) {
		whereArr = append(whereArr, "( cop.id in ( select obra_id from main.lead where team_id = '{teamId}' and nullif(favorited_at, '') is not null ) )")
	}

	if slices.Contains(statuses, StatusComObservacao) {
		whereArr = append(whereArr, "( n.id is not null )")
	}

	if slices.Contains(statuses, StatusExcluida) {
		whereArr = append(whereArr, "( cop.id in ( select obra_id from main.lead where team_id = '{teamId}' and nullif(excluded_at, '') is not null ) )")
	} else {
		// exclude by default
		whereArr = append(whereArr, "( cop.id not in ( select obra_id from main.lead where team_id = '{teamId}' and nullif(excluded_at, '') is not null ) )")
	}

	if len(whereArr) == 0 {
		return "1=1"
	}
	return "(" + strings.Join(whereArr, " and ") + ")"
}

func makeDateFilterCond(startDateFrom, startDateTo, endDateFrom, endDateTo string) string {
	whereArr := []string{}

	// Validação básica de formato de data (YYYY-MM-DD)
	validateDate := func(dateStr string) bool {
		if dateStr == "" {
			return true
		}
		// Verificação simples do formato YYYY-MM-DD
		if len(dateStr) == 10 && dateStr[4] == '-' && dateStr[7] == '-' {
			return true
		}
		return false
	}

	if !validateDate(startDateFrom) || !validateDate(startDateTo) || 
	   !validateDate(endDateFrom) || !validateDate(endDateTo) {
		return "1=1" // Se alguma data for inválida, ignora os filtros
	}

	// Filtros de data de início
	if startDateFrom != "" {
		whereArr = append(whereArr, g.F("cop.start_date >= '%s'", startDateFrom))
	}
	if startDateTo != "" {
		whereArr = append(whereArr, g.F("cop.start_date <= '%s'", startDateTo))
	}

	// Filtros de data de fim
	if endDateFrom != "" {
		whereArr = append(whereArr, g.F("date(julianday(cop.end_date) + 400) >= '%s'", endDateFrom))
	}
	if endDateTo != "" {
		whereArr = append(whereArr, g.F("date(julianday(cop.end_date) + 400) <= '%s'", endDateTo))
	}

	if len(whereArr) == 0 {
		return "1=1"
	}
	
	return "(" + strings.Join(whereArr, " and ") + ")"
}

func QueryCities(c echo.Context) error {
	cities := lo.Keys(store.ObrasCities)
	sort.Strings(cities)
	return c.JSON(200, cities)
}

func QueryObrasPlusContacts(c echo.Context) error {
	req := NewRequest(c)
	nome := c.QueryParam("nome")
	uf := c.QueryParam("uf")
	cidade := c.QueryParam("cidade")

	for _, val := range []string{nome, uf, cidade} {
		if !validateFilter(val) {
			return ErrJSON(404, g.Error("invalid filter"))
		}
	}

	m := g.M(
		"nome", nome,
		"uf", uf,
		"cidade", cidade,
	)
	coreObrasPlusPhoneSQL, _ := templates.ReadFile("templates/core/core_obras_plus_phone.sql")
	dataPhone, err := req.SqlQuery(g.Rm(string(coreObrasPlusPhoneSQL), m))
	if err != nil {
		return ErrJSON(502, err, "error querying phone data")
	}

	coreObrasPlusEmailSQL, _ := templates.ReadFile("templates/core/core_obras_plus_email.sql")
	dataEmail, err := req.SqlQuery(g.Rm(string(coreObrasPlusEmailSQL), m))
	if err != nil {
		return ErrJSON(502, err, "error querying email data")
	}

	records := dataPhone.RecordsCasted(true)
	records = append(records, dataEmail.RecordsCasted(true)...)

	return c.JSON(200, g.M("records", records))

}

func QueryCityNeighborhood(c echo.Context) (err error) {
	req := NewRequest(c)

	city := c.QueryParam("city")
	if found := validateCity(city); !found {
		return ErrJSON(404, g.Error("invalid city %s", city))
	}

	coreObrasPlusNeighborhood, _ := templates.ReadFile("templates/core/core_obras_plus_neighborhood.sql")

	data, err := req.SqlQuery(g.Rm(string(coreObrasPlusNeighborhood), g.M("city", city)))
	if err != nil {
		return ErrJSON(502, err, "error querying data")
	}

	return c.JSON(200, g.M("barrios", data.ColValuesStr(0)))
}

func QueryObrasPlus(c echo.Context) (err error) {
	req := NewRequest(c)

	city := c.QueryParam("city")
	neighborhoods := strings.Split(c.QueryParam("bairro"), "|")
	order := c.QueryParam("order")
	statuses := strings.Split(c.QueryParam("statuses"), ",")
	export := cast.ToBool(c.QueryParam("export"))
	filter := strings.TrimSpace(c.QueryParam("filter"))
	offset := cast.ToInt(c.QueryParam("offset"))
	sizeMin := cast.ToInt(c.QueryParam("sizeMin"))
	sizeMax := cast.ToInt(c.QueryParam("sizeMax"))
	itemPerPage := cast.ToInt(c.QueryParam("itemsPerPage"))
	startDateFrom := strings.TrimSpace(c.QueryParam("startDateFrom"))
	startDateTo := strings.TrimSpace(c.QueryParam("startDateTo"))
	endDateFrom := strings.TrimSpace(c.QueryParam("endDateFrom"))
	endDateTo := strings.TrimSpace(c.QueryParam("endDateTo"))

	user, err := getUser(c, req.UserID())
	if err != nil {
		return ErrJSON(404, g.Error("invalid user id"))
	}

	if sizeMax == 0 {
		sizeMax = 9999999
	}

	if user.ID != "" && user.Team.Export > 0 {
		if itemPerPage == 0 || itemPerPage > user.Team.Export {
			itemPerPage = user.Team.Export
		}
	} else if itemPerPage > 50 {
		itemPerPage = 50
	} else if itemPerPage <= 0 {
		itemPerPage = 10
	}

	allowedOrderStr := []any{
		"first_listing_date-desc,start_date-desc",
		"first_listing_date-asc,start_date-asc",
		"size-desc",
		"size-asc",
	}
	if found := validateCity(city); !found {
		return ErrJSON(404, g.Error("invalid city %s", city))
	} else if !g.In(order, allowedOrderStr...) {
		return ErrJSON(404, g.Error("invalid order"))
	} else if !validateFilter(filter) {
		return ErrJSON(404, g.Error("invalid filter"))
	}

	coreObrasPlusSQL, _ := templates.ReadFile("templates/core/core_obras_plus.sql")
	coreObrasPlusEnrichedSQL, _ := templates.ReadFile("templates/core/core_obras_plus_enriched.sql")

	sql := string(coreObrasPlusSQL)

	if export {
		sql = g.Rm(
			string(coreObrasPlusEnrichedSQL),
			g.M("coreObrasPlusSQL", string(coreObrasPlusSQL)),
		)
		offset = 0
	}

	m := g.M(
		"city", city,
		"order", strings.ReplaceAll(order, "-", " "),
		"sizeMin", sizeMin,
		"sizeMax", sizeMax,
		"neighborhoodCond", makeNeighborhoodCond(neighborhoods),
		"statusCond", g.R(makeStatusCond(statuses), "teamId", user.Team.ID),
		"filterCond", makeFilterCond(filter),
		"dateFilterCond", makeDateFilterCond(startDateFrom, startDateTo, endDateFrom, endDateTo),
		"itemPerPage", itemPerPage,
		"offset", offset,
		"teamId", user.Team.ID,
	)

	data, err := req.SqlQuery(g.Rm(sql, m))
	if err != nil {
		return ErrJSON(502, err, "error querying data")
	}

	if export {
		c.Set("data", data)
		return nil
	}

	records := data.RecordsCasted(true)

	// cast bool
	for i, rec := range records {
		records[i]["has_owner_phone"] = cast.ToBool(rec["has_owner_phone"])
		records[i]["has_owner_email"] = cast.ToBool(rec["has_owner_email"])
		records[i]["has_professional_phone"] = cast.ToBool(rec["has_professional_phone"])
		records[i]["has_professional_email"] = cast.ToBool(rec["has_professional_email"])
	}

	sqlCount, _ := templates.ReadFile("templates/core/core_obras_plus_count.sql")
	data, err = req.SqlQuery(g.Rm(string(sqlCount), m))
	if err != nil {
		return ErrJSON(502, err, "error querying data")
	}
	total := cast.ToInt(data.Rows[0][0])

	return c.JSON(200, g.M("records", records, "total", total))
}

func QueryObrasPlusExport(c echo.Context) error {
	city := c.QueryParam("city")

	if city == "" {
		return ErrJSON(404, g.Error("did not provide city"))
	}

	err := QueryObrasPlus(c)
	if err != nil {
		return err
	}

	data := c.Get("data").(iop.Dataset)

	// sometimes returns 0 rows due to sqlite?
	// check length of rows, delay and try again
	if len(data.Rows) == 0 {
		time.Sleep(2 * time.Second)
		err := QueryObrasPlus(c)
		if err != nil {
			return err
		}
		data = c.Get("data").(iop.Dataset)
	}

	// translate columns
	sCols := []string{"address", "bairro", "city", "state", "size", "owner", "professional", "start_date", "end_date", "owner_first_email", "professional_first_email", "owner_first_telefone", "professional_first_telefone", "owner_emails", "professional_emails", "owner_telefones", "professional_telefones"}
	tCols := []string{"ENDERCO", "BAIRRO", "CIDADE", "UF", "TAMANHO", "PROPRIETARIO", "PROFISSIONAL", "DATA_DE_INICIO", "PREVISAO_DE_TERMINO", "PROPRIETARIO_EMAIL", "PROFISSIONAL_EMAIL", "PROPRIETARIO_TELEFONE", "PROFISSIONAL_TELEFONE", "PROPRIETARIO_EMAILS_OUTROS", "PROFISSIONAL_EMAILS_OUTROS", "PROPRIETARIO_TELEFONES_OUTROS", "PROFISSIONAL_TELEFONES_OUTROS"}

	newData := iop.NewDataset(iop.NewColumnsFromFields(tCols...))
	fm := data.Columns.FieldMap(true)
	for _, row := range data.Rows {
		nRow := make([]any, len(tCols))
		for i, col := range sCols {
			nRow[i] = row[fm[col]]
		}
		newData.Rows = append(newData.Rows, nRow)
	}

	xls := filesys.NewExcel()
	err = xls.WriteSheet("Sheet1", newData.Stream(), "overwrite")
	if err != nil {
		return ErrJSON(502, err, "error writing excel")
	}

	buf := new(bytes.Buffer)
	err = xls.WriteToWriter(buf)
	if err != nil {
		return ErrJSON(502, err, "error writing excel buffer")
	}

	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

func PatchLeadToggle(c echo.Context) error {
	c.Set("team_id", c.QueryParam("team_id"))

	req := NewRequest(c)

	if err := req.ValidatePayload("team_id", "obra_id", "toggle_col"); err != nil {
		return ErrJSON(404, err)
	}

	data, err := LeadToggle(&req)
	if err != nil {
		return err
	}
	return c.JSON(200, data)
}

func LeadToggle(req *Request) (data map[string]any, err error) {

	where := dbx.NewExp(
		"team_id = {:team_id} and obra_id = {:obra_id}",
		req.Payload.Map(),
	)
	records, err := req.Dao().FindRecordsByExpr("lead", where)
	if err != nil {
		return nil, ErrJSON(500, err, "error querying lead")
	}

	if len(records) == 0 {
		// need to insert into lead, fills the record pointer
		err := req.CreateRecord("lead")
		if err != nil {
			return nil, ErrJSON(500, err, "error creating record")
		}

		// set owner_id to user who created record
		req.Record.Set("owner_id", req.UserID())
	} else {
		req.Record = records[0]
	}

	// toggle
	toggleCol := req.Payload.String("toggle_col")
	toggleVal := req.Record.GetDateTime(toggleCol)
	if toggleVal.IsZero() || toggleVal.String() == "" {
		req.Record.Set(toggleCol, time.Now().UTC())
	} else if !g.In(toggleCol, "visited_at", "owner_contact_pending_at", "professional_contact_pending_at", "owner_contacted_at", "professional_contacted_at") {
		req.Record.Set(toggleCol, nil)
	}

	g.Debug("set toggleCol (%s) for %s to %s", toggleCol, req.Record.Get("id"), req.Record.GetDateTime(toggleCol).String())

	if err = req.SaveRecord(); err != nil {
		return nil, ErrJSON(500, err, "error saving record")
	}

	if !req.Record.GetDateTime("favorited_at").IsZero() {
		if err = addLeadToVendaMaisList(req); err != nil {
			return nil, ErrJSON(500, err, "error adding lead to list")
		}
	}

	data = g.M(
		"lead_id", req.Record.Get("id"),
		"list_lead_id", req.Payload["list_lead_id"],
		"visited_at", req.Record.GetDateTime("visited_at").String(),
		"favorited_at", req.Record.GetDateTime("favorited_at").String(),
		"excluded_at", req.Record.GetDateTime("excluded_at").String(),
		"owner_contacted_at", req.Record.GetDateTime("owner_contacted_at").String(),
		"professional_contacted_at", req.Record.GetDateTime("professional_contacted_at").String(),
	)
	return data, nil
}
