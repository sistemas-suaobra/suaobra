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
var resolvedCoreDbPath string

func SetPocketBaseDB(app *pocketbase.PocketBase) (err error) {
	MainDao = app.Dao()
	mainDataDir = resolveMainDataDir(app.DataDir())

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

func resolveMainDataDir(dataDir string) string {
	dataDir = strings.TrimSpace(dataDir)
	if dataDir == "" {
		return dataDir
	}
	if abs, err := filepath.Abs(dataDir); err == nil {
		return abs
	}
	return dataDir
}

// CoreDbPath retorna o core.db resolvido em runtime (ou o default de produção).
func CoreDbPath() string {
	if resolvedCoreDbPath != "" {
		return resolvedCoreDbPath
	}
	if p := strings.TrimSpace(os.Getenv("CORE_DB_PATH")); p != "" {
		return p
	}
	return "/app/data/core.db"
}

func coreDbCandidates() []string {
	if p := strings.TrimSpace(os.Getenv("CORE_DB_PATH")); p != "" {
		return []string{p}
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, 8)
	add := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}

	if mainDataDir != "" {
		dataRoot := filepath.Dir(mainDataDir)
		add(filepath.Join(dataRoot, "core.db"))
		add(filepath.Join(dataRoot, "core", "core.db"))
	}

	add("/app/data/core.db")
	add("/app/data/core/core.db")
	add("./data/core/core.db")

	return out
}

func findCoreDbFile() (string, error) {
	for _, candidate := range coreDbCandidates() {
		st, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if st.Size() < 4096 {
			g.Warn("core.db invalido ou vazio (%d bytes) em %s", st.Size(), candidate)
			continue
		}
		if abs, err := filepath.Abs(candidate); err == nil {
			candidate = abs
		}
		return candidate, nil
	}

	return "", g.Error("core.db nao encontrado nos caminhos: %v", coreDbCandidates())
}

func coreIsAttached() bool {
	if MainDB == nil {
		return false
	}
	_, err := MainDB.Exec("select 1 from core.core_obras_plus limit 1")
	return err == nil
}

func AttachCoreDb() error {
	if MainDB == nil {
		return g.Error("MainDB nao inicializado")
	}

	if coreIsAttached() {
		return nil
	}

	corePath, err := findCoreDbFile()
	if err != nil {
		g.Warn("%v — Obras+ indisponivel ate copiar o ficheiro para o volume", err)
		return nil
	}

	resolvedCoreDbPath = corePath
	attachCandidates := coreAttachSQLCandidates(corePath)

	var lastErr error
	for _, attachPath := range attachCandidates {
		g.Info("tentando attach core.db: file=%s sql=%s", corePath, attachPath)
		if _, err = MainDB.Exec(g.F("attach database '%s' as 'core'", escapeSQLLiteral(attachPath))); err != nil {
			lastErr = err
			g.Warn("attach falhou com sql=%s: %v", attachPath, err)
			continue
		}
		if coreIsAttached() {
			g.Info("core.db anexado com sucesso: file=%s sql=%s", corePath, attachPath)
			if err = loadCities(); err != nil {
				g.Warn("nao foi possivel carregar cidades apos attach do core: %v", err)
			}
			return nil
		}
		lastErr = g.Error("attach executou mas core.core_obras_plus continua indisponivel")
	}

	if lastErr != nil {
		g.Warn("falha ao anexar core.db (%s): %v — Obras+ ficara indisponivel", corePath, lastErr)
	}
	return nil
}

// EnsureCoreReady garante core.db anexado e cache de cidades carregado antes de queries Obras+.
func EnsureCoreReady() {
	if MainDB == nil {
		return
	}
	if !coreIsAttached() {
		_ = AttachCoreDb()
	}
	if len(ObrasCities) == 0 && coreIsAttached() {
		_ = loadCities()
	}
}

// coreAttachSQLCandidates gera paths para ATTACH.
// Deploy antigo usava caminho absoluto (/app/data/core.db); Docker legado montava /app/data/core/core.db.
func coreAttachSQLCandidates(absCorePath string) []string {
	absCorePath = strings.TrimSpace(absCorePath)
	if absCorePath == "" {
		return nil
	}

	absCorePath = filepath.ToSlash(absCorePath)
	if abs, err := filepath.Abs(absCorePath); err == nil {
		absCorePath = filepath.ToSlash(abs)
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, 4)
	add := func(path string) {
		path = filepath.ToSlash(strings.TrimSpace(path))
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}

	// 1) Caminho absoluto — funcionava nos deploys anteriores (433ec67 / c7b25b7).
	add(absCorePath)

	// 2) Relativo ao data.db (PocketBase em /app/data/main/data.db).
	if mainDataDir != "" {
		if rel, err := filepath.Rel(mainDataDir, absCorePath); err == nil {
			add(filepath.ToSlash(rel))
		}
	}

	// 3) Fallbacks conhecidos no volume Docker.
	switch {
	case strings.HasSuffix(absCorePath, "/core/core.db"):
		add("../core/core.db")
	case strings.HasSuffix(absCorePath, "/core.db"):
		add("../core.db")
	}

	return out
}

func escapeSQLLiteral(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func MainDaoDB() dbx.Builder {
	return MainDao.DB()
}

func MainDbNewQuery(sql string) *dbx.Query {
	query := MainDaoDB().NewQuery(sql)
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
	if !coreIsAttached() {
		g.Warn("nao foi possivel carregar cidades: core.db nao anexado")
		return nil
	}

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

	g.Info("Loaded %d cities from core.db", len(ObrasCities))
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
