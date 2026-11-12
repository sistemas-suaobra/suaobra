package server

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/models"
	"github.com/spf13/cast"
)

var cityIDPattern = regexp.MustCompile(`^[A-Z]{2}-(?:\*|[A-Z0-9 ]+)$`)

func requirePlatformAdmin(req *Request) error {
	auth := req.AuthRecord()
	if auth == nil || !userIsAdmin(auth) {
		return g.Error("apenas administradores da plataforma podem fazer isso")
	}
	return nil
}

func normalizeCityIDs(raw []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		id := strings.ToUpper(strings.TrimSpace(item))
		id = strings.Join(strings.Fields(id), " ")
		if id == "" || !cityIDPattern.MatchString(id) {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func parseCitiesField(raw any) []string {
	switch value := raw.(type) {
	case []string:
		return normalizeCityIDs(value)
	case []any:
		items := make([]string, 0, len(value))
		for _, item := range value {
			items = append(items, cast.ToString(item))
		}
		return normalizeCityIDs(items)
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || trimmed == "null" {
			return []string{}
		}
		var parsed []string
		if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
			return normalizeCityIDs(parsed)
		}
		return normalizeCityIDs([]string{trimmed})
	case []byte:
		if len(value) == 0 {
			return []string{}
		}
		var parsed []string
		if err := json.Unmarshal(value, &parsed); err == nil {
			return normalizeCityIDs(parsed)
		}
	}
	return []string{}
}

// PlatformTeams lista todas as contas (times) com a abrangência atual.
func PlatformTeams(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	if err := requirePlatformAdmin(&req); err != nil {
		return ErrJSON(403, err)
	}

	sql := `
	select
		t.id,
		t.name,
		t.blocked,
		t.active,
		t.cities,
		(
			select u.email
			from main.user u
			where u.team_id = t.id
			order by u.manager desc, u.created asc
			limit 1
		) as contact_email,
		(
			select count(*)
			from main.user u
			where u.team_id = t.id
		) as member_count,
		(
			select group_concat(u.email, ', ')
			from main.user u
			where u.team_id = t.id and u.manager = 1
		) as manager_emails
	from main.team t
	order by lower(coalesce(nullif(t.name, ''), t.id))
	`

	records, err := req.SqlQueryRecords(sql)
	if err != nil {
		return ErrJSON(500, err, "error listing teams")
	}

	for _, record := range records {
		record["cities"] = parseCitiesField(record["cities"])
		record["blocked"] = cast.ToBool(record["blocked"])
		record["active"] = cast.ToBool(record["active"])
		record["member_count"] = cast.ToInt(record["member_count"])
		record["managers"] = splitEmails(cast.ToString(record["manager_emails"]))
		delete(record, "manager_emails")
	}

	return c.JSON(200, records)
}

func splitEmails(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		email := strings.TrimSpace(part)
		if email == "" {
			continue
		}
		key := strings.ToLower(email)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, email)
	}
	return out
}

// PlatformSetCities grava a abrangência (vários estados/cidades) de um time.
func PlatformSetCities(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	if err := requirePlatformAdmin(&req); err != nil {
		return ErrJSON(403, err)
	}
	if err := req.ValidatePayload("team_id", "cities"); err != nil {
		return ErrJSON(400, err, "error validating payload")
	}

	teamID := strings.TrimSpace(cast.ToString(req.Payload["team_id"]))
	if teamID == "" {
		return ErrJSON(400, g.Error("team_id é necessário"))
	}

	rawCities := req.Payload["cities"]
	var incoming []string
	switch value := rawCities.(type) {
	case []string:
		incoming = value
	case []any:
		for _, item := range value {
			incoming = append(incoming, cast.ToString(item))
		}
	case string:
		if err := json.Unmarshal([]byte(value), &incoming); err != nil {
			return ErrJSON(400, g.Error("cities deve ser uma lista"))
		}
	default:
		return ErrJSON(400, g.Error("cities deve ser uma lista"))
	}

	cities := normalizeCityIDs(incoming)
	encoded, err := json.Marshal(cities)
	if err != nil {
		return ErrJSON(500, err, "error encoding cities")
	}

	team, err := req.Dao().FindRecordById("team", teamID)
	if err != nil {
		return ErrJSON(404, g.Error("time não encontrado"))
	}

	team.Set("cities", string(encoded))
	if err := req.Dao().SaveRecord(team); err != nil {
		return ErrJSON(500, err, "error saving cities")
	}

	return c.JSON(200, g.M(
		"id", team.Id,
		"name", cast.ToString(team.Get("name")),
		"cities", cities,
		"blocked", cast.ToBool(team.Get("blocked")),
		"active", cast.ToBool(team.Get("active")),
	))
}

// PlatformTeamDetail retorna a conta com membros (para o backoffice).
func PlatformTeamDetail(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	if err := requirePlatformAdmin(&req); err != nil {
		return ErrJSON(403, err)
	}

	teamID := strings.TrimSpace(c.QueryParam("team_id"))
	if teamID == "" {
		teamID = strings.TrimSpace(cast.ToString(req.Payload["team_id"]))
	}
	if teamID == "" {
		return ErrJSON(400, g.Error("team_id é necessário"))
	}

	team, err := req.Dao().FindRecordById("team", teamID)
	if err != nil {
		return ErrJSON(404, g.Error("time não encontrado"))
	}

	sql := `
	select
		u.id,
		u.email,
		u.created,
		u.updated,
		u.manager,
		u.properties
	from main.user u
	where u.team_id = '{team_id}'
	order by u.manager desc, lower(u.email) asc
	`
	members, err := req.SqlQueryRecords(g.R(sql, "team_id", teamID))
	if err != nil {
		return ErrJSON(500, err, "error listing members")
	}

	managers := make([]string, 0)
	for _, member := range members {
		if cast.ToBool(member["manager"]) {
			managers = append(managers, cast.ToString(member["email"]))
		}
	}

	return c.JSON(200, g.M(
		"id", team.Id,
		"name", cast.ToString(team.Get("name")),
		"blocked", cast.ToBool(team.Get("blocked")),
		"active", cast.ToBool(team.Get("active")),
		"cities", parseCitiesField(team.Get("cities")),
		"member_count", len(members),
		"managers", managers,
		"members", members,
	))
}

// PlatformSetTeamStatus ativa/desativa ou bloqueia uma conta.
func PlatformSetTeamStatus(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	if err := requirePlatformAdmin(&req); err != nil {
		return ErrJSON(403, err)
	}
	if err := req.ValidatePayload("team_id"); err != nil {
		return ErrJSON(400, err, "error validating payload")
	}

	teamID := strings.TrimSpace(req.Payload.String("team_id"))
	team, err := req.Dao().FindRecordById("team", teamID)
	if err != nil {
		return ErrJSON(404, g.Error("time não encontrado"))
	}

	if _, ok := req.Payload["active"]; ok {
		team.Set("active", req.Payload.Bool("active"))
	}
	if _, ok := req.Payload["blocked"]; ok {
		team.Set("blocked", req.Payload.Bool("blocked"))
	}

	if err := req.Dao().SaveRecord(team); err != nil {
		return ErrJSON(500, err, "error saving team status")
	}

	return c.JSON(200, g.M(
		"id", team.Id,
		"name", cast.ToString(team.Get("name")),
		"active", cast.ToBool(team.Get("active")),
		"blocked", cast.ToBool(team.Get("blocked")),
		"cities", parseCitiesField(team.Get("cities")),
	))
}

// PlatformUpdateTeam atualiza dados básicos da conta (nome).
func PlatformUpdateTeam(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	if err := requirePlatformAdmin(&req); err != nil {
		return ErrJSON(403, err)
	}
	if err := req.ValidatePayload("team_id"); err != nil {
		return ErrJSON(400, err, "error validating payload")
	}

	teamID := strings.TrimSpace(req.Payload.String("team_id"))
	team, err := req.Dao().FindRecordById("team", teamID)
	if err != nil {
		return ErrJSON(404, g.Error("time não encontrado"))
	}

	if name, ok := req.Payload["name"]; ok {
		trimmed := strings.TrimSpace(cast.ToString(name))
		if trimmed == "" {
			return ErrJSON(400, g.Error("nome inválido"))
		}
		if len(trimmed) > 200 {
			trimmed = trimmed[:200]
		}
		team.Set("name", trimmed)
	}

	if err := req.Dao().SaveRecord(team); err != nil {
		return ErrJSON(500, err, "error saving team")
	}

	return c.JSON(200, g.M(
		"id", team.Id,
		"name", cast.ToString(team.Get("name")),
		"blocked", cast.ToBool(team.Get("blocked")),
		"active", cast.ToBool(team.Get("active")),
		"cities", parseCitiesField(team.Get("cities")),
	))
}

// PlatformUpdateMember edita membro de qualquer equipe (plataforma).
func PlatformUpdateMember(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	if err := requirePlatformAdmin(&req); err != nil {
		return ErrJSON(403, err)
	}
	if err := req.ValidatePayload("team_id", "user_id"); err != nil {
		return ErrJSON(400, err, "error validating payload")
	}

	teamID := strings.TrimSpace(req.Payload.String("team_id"))
	userID := strings.TrimSpace(req.Payload.String("user_id"))
	record, err := req.Dao().FindRecordById("user", userID)
	if err != nil {
		return ErrJSON(404, g.Error("usuário não encontrado"))
	}
	if cast.ToString(record.Get("team_id")) != teamID {
		return ErrJSON(403, g.Error("usuário não pertence a esta equipe"))
	}

	props, err := propertiesMap(record.Get("properties"))
	if err != nil {
		return ErrJSON(500, err, "erro ao ler os dados do usuário")
	}

	if name, ok := req.Payload["name"]; ok {
		props["name"] = strings.TrimSpace(cast.ToString(name))
	}
	if phone, ok := req.Payload["phone"]; ok {
		props["phone"] = strings.TrimSpace(cast.ToString(phone))
	}
	if _, ok := req.Payload["active"]; ok {
		props["active"] = req.Payload.Bool("active")
	}
	record.Set("properties", props)

	if _, ok := req.Payload["manager"]; ok {
		record.Set("manager", req.Payload.Bool("manager"))
	}
	if _, ok := req.Payload["is_manager"]; ok {
		record.Set("manager", req.Payload.Bool("is_manager"))
	}

	if err := req.Dao().SaveRecord(record); err != nil {
		return ErrJSON(500, err, "erro ao atualizar usuário")
	}

	return c.JSON(200, g.M(
		"id", record.Id,
		"email", record.Email(),
		"manager", record.GetBool("manager"),
		"properties", props,
		"message", "Usuário atualizado",
	))
}

// PlatformRemoveMember remove o usuário da equipe (volta para equipe pessoal).
func PlatformRemoveMember(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	if err := requirePlatformAdmin(&req); err != nil {
		return ErrJSON(403, err)
	}
	if err := req.ValidatePayload("team_id", "user_id"); err != nil {
		return ErrJSON(400, err, "error validating payload")
	}

	teamID := strings.TrimSpace(req.Payload.String("team_id"))
	userID := strings.TrimSpace(req.Payload.String("user_id"))
	record, err := req.Dao().FindRecordById("user", userID)
	if err != nil {
		return ErrJSON(404, g.Error("usuário não encontrado"))
	}
	if cast.ToString(record.Get("team_id")) != teamID {
		return ErrJSON(403, g.Error("usuário não pertence a esta equipe"))
	}

	userEmail := cast.ToString(record.Get("email"))
	personalTeam, err := req.Dao().FindFirstRecordByData("team", "name", userEmail)
	if err != nil || personalTeam == nil {
		teamCollection, findErr := req.Dao().FindCollectionByNameOrId("team")
		if findErr != nil {
			return ErrJSON(500, findErr, "erro ao encontrar coleção de equipe")
		}
		personalTeam = models.NewRecord(teamCollection)
		personalTeam.Set("name", userEmail)
		personalTeam.Set("owner_id", userID)
		personalTeam.Set("active", true)
		personalTeam.Set("blocked", false)
		personalTeam.Set("cities", "[]")
		personalTeam.Set("properties", g.Marshal(g.M()))
		personalTeam.Set("entitlements", g.Marshal(g.M()))
		if saveErr := req.Dao().SaveRecord(personalTeam); saveErr != nil {
			return ErrJSON(500, saveErr, "erro ao criar equipe pessoal")
		}
	}

	record.Set("team_id", personalTeam.Id)
	record.Set("manager", false)
	if err := req.Dao().SaveRecord(record); err != nil {
		return ErrJSON(500, err, "erro ao remover usuário da equipe")
	}

	return c.JSON(200, g.M("message", "Usuário removido com sucesso"))
}

// PlatformDeleteTeam desativa a conta e limpa a abrangência (exclusão lógica).
func PlatformDeleteTeam(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	if err := requirePlatformAdmin(&req); err != nil {
		return ErrJSON(403, err)
	}
	if err := req.ValidatePayload("team_id"); err != nil {
		return ErrJSON(400, err, "error validating payload")
	}

	teamID := strings.TrimSpace(req.Payload.String("team_id"))
	team, err := req.Dao().FindRecordById("team", teamID)
	if err != nil {
		return ErrJSON(404, g.Error("time não encontrado"))
	}

	team.Set("active", false)
	team.Set("blocked", true)
	team.Set("cities", "[]")
	if err := req.Dao().SaveRecord(team); err != nil {
		return ErrJSON(500, err, "error deleting team")
	}

	return c.JSON(200, g.M(
		"id", team.Id,
		"deleted", true,
		"active", false,
		"blocked", true,
		"cities", []string{},
	))
}
