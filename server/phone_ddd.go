package server

import (
	"strings"
	"unicode"
)

// dddsByUF: DDDs oficiais por estado (agregado IBGE/municípios).
// Usado para filtrar contatos no modal do Obras+ pela UF da obra.
var dddsByUF = map[string]map[string]struct{}{
	"AC": setDDD("68"),
	"AL": setDDD("82"),
	"AM": setDDD("92", "97"),
	"AP": setDDD("96"),
	"BA": setDDD("71", "73", "74", "75", "77"),
	"CE": setDDD("85", "88"),
	"DF": setDDD("61"),
	"ES": setDDD("27", "28"),
	"GO": setDDD("61", "62", "64"),
	"MA": setDDD("98", "99"),
	"MG": setDDD("31", "32", "33", "34", "35", "37", "38"),
	"MS": setDDD("67"),
	"MT": setDDD("65", "66"),
	"PA": setDDD("91", "93", "94"),
	"PB": setDDD("83"),
	"PE": setDDD("81", "87"),
	"PI": setDDD("86", "89"),
	"PR": setDDD("41", "42", "43", "44", "45", "46", "47", "49"),
	"RJ": setDDD("21", "22", "24"),
	"RN": setDDD("84"),
	"RO": setDDD("69"),
	"RR": setDDD("95"),
	"RS": setDDD("51", "53", "54", "55"),
	"SC": setDDD("42", "47", "48", "49"),
	"SE": setDDD("79"),
	"SP": setDDD("11", "12", "13", "14", "15", "16", "17", "18", "19"),
	"TO": setDDD("63"),
}

func setDDD(vals ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(vals))
	for _, v := range vals {
		m[v] = struct{}{}
	}
	return m
}

func digitsOnlyPhone(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// extractPhoneDDD returns the 2-digit area code, or "" if unknown.
func extractPhoneDDD(telefone string) string {
	d := digitsOnlyPhone(telefone)
	if d == "" {
		return ""
	}
	if strings.HasPrefix(d, "55") && len(d) >= 12 {
		d = d[2:]
	}
	if strings.HasPrefix(d, "0") && len(d) >= 11 {
		d = d[1:]
	}
	if len(d) >= 8 {
		return d[:2]
	}
	return ""
}

// phoneDDDMatchesUF reports whether the phone DDD belongs to the given UF.
// Empty UF or unparseable phone → false (hide).
func phoneDDDMatchesUF(telefone, uf string) bool {
	uf = strings.ToUpper(strings.TrimSpace(uf))
	allowed, ok := dddsByUF[uf]
	if !ok {
		return true // UF desconhecida: não filtra
	}
	ddd := extractPhoneDDD(telefone)
	if ddd == "" {
		return false
	}
	_, ok = allowed[ddd]
	return ok
}
