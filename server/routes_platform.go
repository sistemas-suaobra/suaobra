package server

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
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
		) as member_count
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
	}

	return c.JSON(200, records)
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
	))
}
