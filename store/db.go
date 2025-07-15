package store

import (
	_ "embed"
	"os"
	"strings"
	"time"

	"github.com/flarco/dbio/database"
	"github.com/flarco/g"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/spf13/cast"
)

//go:embed load.core.data.sql
var loadDataSQL string

var MainDao *daos.Dao
var MainDB database.Connection

func SetPocketBaseDB(app *pocketbase.PocketBase) (err error) {
	MainDao = app.Dao()

	if MainDB != nil {
		MainDB.Close()
	}
	url := g.F("sqlite://%s/data.db?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL", app.DataDir())
	MainDB, err = database.NewConn(url)
	if err != nil {
		return g.Error(err, "could not init main db")
	}

	err = AttachCoreDb()
	if err != nil {
		return g.Error(err, "could not attach core.db")
	}

	// err = LoadCoreData()
	// g.LogFatal(err, "could not init main db")

	err = loadCities()
	g.LogFatal(err, "could not init main db")

	return
}

func AttachCoreDb() error {
	// check if attached first.
	_, err := MainDB.Exec("select 1 from core.core_obras_plus limit 1")
	if err != nil {
		g.Warn("attaching core")
		_, err = MainDB.Exec("attach database 'core.db' as 'core'")
	}
	return err
}

func MainDaoDB() dbx.Builder {
	return MainDao.DB()
}

func MainDbNewQuery(sql string) *dbx.Query {
	query := MainDaoDB().NewQuery(sql)
	// if g.IsDebugLow() {
	// 	g.Debug(query.SQL())
	// }
	return query
}

func ID(table, oldID string) string {
	tablePrefix := strings.TrimRight(table, "s")

	switch tablePrefix {
	case "lead_contact":
		tablePrefix = "cont"
	case "lead_activity":
		tablePrefix = "actv"
	case "list_stage":
		tablePrefix = "stg"
	case "list_lead":
		tablePrefix = "lst_lead"
	}

	if strings.HasPrefix(oldID, tablePrefix) {
		return oldID
	}

	return g.F("%s_%s", tablePrefix, g.RandString(g.AplhanumericRunes, 15))
}

func loadCoreData() error {

	var timestamp int64
	if strings.ToLower(os.Getenv("ENV")) == "development" {
		b, _ := os.ReadFile("store/core_data_timestamp")
		timestamp = cast.ToInt64(string(b))
		stat, err := os.Stat("./data/core/core.db")
		if err == nil && stat.ModTime().Unix() <= timestamp {
			return nil
		}
	}

	g.Info("loading core data")
	_, err := MainDB.ExecMulti(loadDataSQL)
	if err != nil {
		return g.Error(err, "could not load core data")
	}

	if strings.ToLower(os.Getenv("ENV")) == "development" {
		timestamp = time.Now().Unix()
		os.WriteFile("store/core_data_timestamp", []byte(cast.ToString(timestamp)), 0777)
	}

	return nil
}

func loadCities() error {
	g.Info("Loading cities")
	data, err := MainDB.Query(`
		select city
		from core.core_obras_plus
		group by 1
		-- having count(1) >= 1000`)
	g.LogFatal(err, "could not load duckbdb cities")

	for _, row := range data.Rows {
		ObrasCities[cast.ToString(row[0])] = struct{}{}
	}

	return nil
}

func BindSQL(sql string, payload map[string]any) string {
	for k, v := range payload {
		switch val := v.(type) {
		case int, int64, float32, float64:
			sql = strings.ReplaceAll(sql, g.F("{:%s}", k), cast.ToString(val))
		case bool:
			sql = strings.ReplaceAll(sql, g.F("{:%s}", k), cast.ToString(cast.ToInt(val)))
		case time.Time:
			sql = strings.ReplaceAll(sql, g.F("{:%s}", k), `'`+val.Format(types.DefaultDateLayout)+`'`)
		default:
			sql = strings.ReplaceAll(sql, g.F("{:%s}", k), `'`+cast.ToString(val)+`'`)
		}
	}
	return sql
}
