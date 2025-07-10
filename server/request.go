package server

import (
	"context"
	"embed"
	"strings"
	"time"

	"github.com/flarco/dbio/iop"
	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/spf13/cast"
	"github.com/suaobra/suaobra-app/store"
)

//go:embed templates
var templates embed.FS

type Request struct {
	App     *pocketbase.PocketBase
	Payload Payload
	Info    *models.RequestInfo
	Record  *models.Record
	Error   error
	Context *g.Context
	echoCtx echo.Context
}

type Payload map[string]any

func NewRequest(c echo.Context) (req Request) {
	app := c.Get("app").(*pocketbase.PocketBase)
	ctx := g.NewContext(context.Background())
	req = Request{
		App:     app,
		Payload: Payload{},
		Info:    apis.RequestInfo(c),
		echoCtx: c,
		Context: &ctx,
	}

	var userID, teamID, userEmail string
	var userIsManager bool
	if req.AuthRecord() != nil {
		userID = cast.ToString(req.AuthRecord().Get("id"))
		teamID = cast.ToString(req.AuthRecord().Get("team_id"))
		userEmail = cast.ToString(req.AuthRecord().Get("email"))
		userIsManager = req.AuthRecord().GetBool("manager")
	} else if req.Admin() != nil {
		userID = cast.ToString(req.Admin().Id)
		userEmail = cast.ToString(req.Admin().Email)
	} else if c.Get("team_id") != nil {
		teamID = cast.ToString(c.Get("team_id"))
	} else if token := c.Request().Header.Get(echo.HeaderAuthorization); token != "" && token == MessengerAuthorization {
		// allow in
	} else {
		req.Error = g.Error("forbidden")
		return
	}

	m := g.M(
		"user_id", userID,
		"user_email", userEmail,
		"team_id", teamID,
		"is_manager", userIsManager,
	)
	if err := c.Bind(&m); err != nil {
		g.LogError(err)
		req.Error = g.Error(err, "error unmarshaling")
		return
	}

	req.Payload = m

	return req
}

func (r *Request) UserID() string {
	return cast.ToString(r.Payload.Map()["user_id"])
}

func (r *Request) TeamID() string {
	return cast.ToString(r.Payload.Map()["team_id"])
}

func (r *Request) IsManager() bool {
	return cast.ToBool(r.Payload.Map()["is_manager"])
}

func (r *Request) TeamRecord() (record *models.Record, err error) {
	return r.Dao().FindRecordById("team", r.TeamID())
}

func (r *Request) Admin() *models.Admin {
	return r.Info.Admin
}

func (r *Request) AuthRecord() *models.Record {
	return r.Info.AuthRecord
}

func (r *Request) SqlQueryResponse(sql string) (err error) {
	if r.Error != nil {
		return ErrJSON(400, r.Error)
	}

	records, err := r.SqlQueryRecords(sql)
	if err != nil {
		return ErrJSON(500, err, "error querying data")
	}
	return r.echoCtx.JSON(200, records)
}

func (r *Request) SqlQuery(sql string) (data iop.Dataset, err error) {

	// g.Debug(sql)
	// g.Debug(g.Pretty(r.Payload.Map()))
	// records := []dbx.NullStringMap{}
	// err = store.MainDbNewQuery(sql).
	// 	Bind(r.Payload.Map()).
	// 	All(&records)

	// if err != nil {
	// 	return nil, g.Error(err, "error querying data")
	// }

	// Sp := iop.NewStreamProcessor()
	// recs = make([]map[string]any, len(records))
	// for i, rec0 := range records {
	// 	rec := g.M()
	// 	for k, v := range rec0 {
	// 		rec[k] = Sp.ParseString(v.String)
	// 	}
	// 	recs[i] = rec
	// }

	store.AttachCoreDb() // FIXME: weird that core.db seems to detach after a while

	sql = store.BindSQL(sql, r.Payload.Map())
retry:
	data, err = store.MainDB.Query(sql)
	if err != nil {
		if strings.Contains(err.Error(), "no such table: core.core_obras_plus") {
			time.Sleep(100 * time.Millisecond)
			g.Warn("retrying query, got error: no such table: core.core_obras_plus")
			goto retry
		}
		return data, g.Error(err, "error querying data")
	}

	return data, nil
}

// BuildWhereClause constructs a dynamic WHERE clause from a base filter and optional additional filters
func (r *Request) BuildWhereClause(baseFilter string, additionalFilters map[string]string) string {
	whereClause := baseFilter

	for field, value := range additionalFilters {
		// Handle user_id parameter, which might be null (empty string)
		// or might have a valid value in the payload
		paramValue := r.Payload.String(value)
		if paramValue != "" && paramValue != "null" && paramValue != "undefined" {
			whereClause += g.F(" and %s = {:%s}", field, value)
		}
	}

	return whereClause
}

func (r *Request) SqlQueryRecords(sql string) (recs []map[string]any, err error) {
	data, err := r.SqlQuery(sql)
	return data.RecordsCasted(), err
}

func (r *Request) SqlExecuteResponse(sql string) (err error) {
	if r.Error != nil {
		return ErrJSON(400, r.Error)
	}

	err = r.SqlExecute(sql)
	if err != nil {
		return ErrJSON(500, err, "error querying data")
	}
	return r.echoCtx.JSON(200, g.M())
}

func (r *Request) SqlExecute(sql string) (err error) {

	// g.Debug(sql)
	// g.Debug(g.Pretty(r.Payload.Map()))
	// _, err = store.MainDbNewQuery(sql).
	// 	Bind(r.Payload.Map()).
	// 	Execute()

	// if err != nil {
	// 	return g.Error(err, "error executing query")
	// }

	store.AttachCoreDb() // FIXME: weird that core.db seems to detach after a while

	sql = store.BindSQL(sql, r.Payload.Map())
	_, err = store.MainDB.Exec(sql)
	if err != nil {
		return g.Error(err, "error querying data")
	}

	return nil
}

func (r *Request) Dao() *daos.Dao { return r.App.Dao() }

func (r *Request) Db() dbx.Builder { return r.Dao().DB() }

func (r *Request) NewRecord(tableName string) *models.Record {
	collection, err := r.Dao().FindCollectionByNameOrId(tableName)
	if err != nil {
		g.LogError(err, "could not get collection %s", tableName)
	}
	r.Record = models.NewRecord(collection)
	for k, v := range r.Payload {
		r.Record.Set(k, v)
	}
	return r.Record
}

func (r *Request) MakeRecord(tableName string) *models.Record {
	r.Record = r.NewRecord(tableName)
	return r.Record
}

func (r *Request) CreateRecord(tableName string) (err error) {
	r.Record = r.MakeRecord(tableName)
	if err := r.Dao().SaveRecord(r.Record); err != nil {
		return g.Error(err, "could not create record in %s", tableName)
	}
	return
}

func (r *Request) SaveRecord() (err error) {
	if err := r.Dao().SaveRecord(r.Record); err != nil {
		return g.Error(err, "could not create record in %s", r.Record.TableName())
	}
	return
}

func (r *Request) ValidatePayload(keys ...string) (err error) {
	if r.Error != nil {
		return g.Error(r.Error)
	}

	eG := g.ErrorGroup{}
	for _, key := range keys {
		if _, ok := r.Payload[key]; !ok {
			eG.Add(g.Error("did not provide key %s", key))
		}
	}
	return eG.Err()
}

func (p Payload) Map() map[string]any {
	return p
}

func (p Payload) String(key string) string {
	return cast.ToString(p[key])
}

func (p Payload) Int(key string) int {
	return cast.ToInt(p[key])
}

func (p Payload) Bool(key string) bool {
	return cast.ToBool(p[key])
}

func (p Payload) Timestamp(key string) time.Time {
	return cast.ToTime(p[key])
}

// ErrJSON returns to the echo.Context as JSON formatted
func ErrJSON(HTTPStatus int, err error, args ...interface{}) *apis.ApiError {
	msg := g.ArgsErrMsg(args...)
	g.LogError(err)
	if msg == "" {
		msg = g.ErrMsg(err)
	} else if g.ErrMsg(err) != "" {
		msg = g.F("%s [%s]", msg, g.ErrMsg(err))
	}
	return apis.NewApiError(HTTPStatus, err.Error(), g.M("error", msg))
}
