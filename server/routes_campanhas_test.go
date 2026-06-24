package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdicionarDestinatariosObrasPlus_OcultarJaContactados_BodyParsing(t *testing.T) {
	body := map[string]any{
		"destinatarios": []map[string]string{
			{"obra_id": "obra_1", "contato_tipo": "OWNER"},
		},
		"ocultar_ja_contactados": true,
	}

	raw, err := json.Marshal(body)
	require.NoError(t, err)

	var parsed struct {
		OcultarJaContactados    bool `json:"ocultar_ja_contactados"`
		OcultarJaContactadosAlt bool `json:"ocultarJaContactados"`
	}
	require.NoError(t, json.Unmarshal(raw, &parsed))

	ocultar := parsed.OcultarJaContactados || parsed.OcultarJaContactadosAlt
	assert.True(t, ocultar)
}

func TestAdicionarDestinatariosObrasPlus_AltJSONKey(t *testing.T) {
	body := map[string]any{
		"destinatarios": []map[string]string{
			{"obra_id": "obra_1", "contato_tipo": "OWNER"},
		},
		"ocultarJaContactados": true,
	}

	raw, err := json.Marshal(body)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/campanhas/x/destinatarios/obras-plus", bytes.NewReader(raw))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var payload struct {
		OcultarJaContactados    bool `json:"ocultar_ja_contactados"`
		OcultarJaContactadosAlt bool `json:"ocultarJaContactados"`
	}
	require.NoError(t, c.Bind(&payload))

	ocultar := payload.OcultarJaContactados || payload.OcultarJaContactadosAlt
	assert.True(t, ocultar)
}
