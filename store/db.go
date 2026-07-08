package store

import (
	_ "embed"
	"os"
	"path/filepath"
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
var mainDataDir string

func SetPocketBaseDB(app *pocketbase.PocketBase) (err error) {
	MainDao = app.Dao()
	mainDataDir = app.DataDir()

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

	err = loadCities()
	if err != nil {
		return err
	}

	return nil
}

// CoreDbPath caminho do SQLite de obras (volume Dokploy: /app/data/core.db).
func CoreDbPath() string {
	if p := strings.TrimSpace(os.Getenv("CORE_DB_PATH")); p != "" {
		return p
	}
	return "/app/data/core.db"
}

func AttachCoreDb() error {
	path := CoreDbPath()

	if _, err := MainDB.Exec("select 1 from core.core_obras_plus limit 1"); err == nil {
		return nil
	}

	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			g.Warn("core.db nao encontrado em %s — Obras+ indisponivel ate copiar o ficheiro para o volume", path)
			return nil
		}
		return g.Error(err, "stat core.db")
	}
	if st.Size() < 4096 {
		g.Warn("core.db invalido ou vazio (%d bytes) em %s — remova e copie o ficheiro real", st.Size(), path)
		return nil
	}

	// Caminhos absolutos como /app/data/core.db quebram o driver dbio no ATTACH
	// ("invalid uri authority: app"). O path relativo ao DB principal resolve no volume Docker.
	attachPath := coreAttachSQLPath(path)
	g.Warn("attaching core from %s (sql path: %s)", path, attachPath)
	if _, err = MainDB.Exec(g.F("attach database '%s' as 'core'", escapeSQLLiteral(attachPath))); err != nil {
		g.Warn("falha ao anexar core.db (%s): %v — Obras+ ficara indisponivel", path, err)
		return nil
	}

	if _, err = MainDB.Exec("select 1 from core.core_obras_plus limit 1"); err != nil {
		g.Warn("core.db em %s nao contem core_obras_plus: %v", path, err)
		return nil
	}

	if err = loadCities(); err != nil {
		g.Warn("nao foi possivel carregar cidades apos attach do core: %v", err)
	}
	return nil
}

// EnsureCoreReady garante core.db anexado e cache de cidades carregado antes de queries Obras+.
func EnsureCoreReady() {
	if MainDB == nil {
		return
	}
	if _, err := MainDB.Exec("select 1 from core.core_obras_plus limit 1"); err != nil {
		_ = AttachCoreDb()
	}
	if len(ObrasCities) == 0 {
		_ = loadCities()
	}
}

// coreAttachSQLPath converte caminho absoluto para relativo ao data.db no ATTACH.
// Em Docker: data.db fica em /app/data/main/ e core.db em /app/data/core.db → ../core.db
func coreAttachSQLPath(absPath string) string {
	absPath = strings.TrimSpace(absPath)
	if absPath == "" {
		return "core.db"
	}

	abs, err := filepath.Abs(absPath)
	if err != nil {
		return filepath.Base(absPath)
	}

	if mainDataDir != "" {
		if rel, err := filepath.Rel(mainDataDir, abs); err == nil {
			rel = filepath.ToSlash(rel)
			if rel != "" && rel != "." {
				return rel
			}
		}
	}

	return "../core.db"
}

func escapeSQLLiteral(value string) string {
	return strings.ReplaceAll(value, "'", "''")
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
	if err != nil {
		g.Warn("nao foi possivel carregar cidades do core (Obras+): %v", err)
		return nil
	}

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
