#!/usr/bin/env python3
"""Diagnóstico read-only: telefone com DDD de outro estado (UF).

Regra: Americana/SP pode ter DDD 11 (DDD de SP).
Só marca inválido se o DDD não pertence ao conjunto de DDDs da UF.

Não apaga nada — só gera relatório.
"""

from __future__ import annotations

import argparse
import csv
import sqlite3
import unicodedata
from collections import Counter
from pathlib import Path


def normalize_text(value: str | None) -> str:
    text = (value or "").strip().upper()
    text = "".join(
        ch for ch in unicodedata.normalize("NFD", text)
        if unicodedata.category(ch) != "Mn"
    )
    return " ".join(text.split())


def digits_only(value: str | None) -> str:
    return "".join(ch for ch in (value or "") if ch.isdigit())


def extract_ddd(telefone: str | None) -> str | None:
    d = digits_only(telefone)
    if not d:
        return None
    if d.startswith("55") and len(d) >= 12:
        d = d[2:]
    if d.startswith("0") and len(d) >= 11:
        d = d[1:]
    if len(d) in (10, 11):
        return d[:2]
    return None


def load_uf_ddd_map(path: Path) -> dict[str, set[str]]:
    out: dict[str, set[str]] = {}
    with path.open(newline="", encoding="utf-8-sig") as fh:
        for row in csv.DictReader(fh):
            uf = normalize_text(row["uf"])
            ddds = {
                "".join(ch for ch in part if ch.isdigit())
                for part in (row.get("ddds") or "").split(",")
            }
            ddds = {d for d in ddds if len(d) == 2}
            if uf and ddds:
                out[uf] = ddds
    return out


def diagnose_phones(db_path: Path, uf_ddds: dict[str, set[str]], sample_limit: int = 40):
    conn = sqlite3.connect(f"file:{db_path}?mode=ro", uri=True)
    conn.row_factory = sqlite3.Row
    cur = conn.cursor()

    totals = Counter()
    by_uf_invalid = Counter()
    by_pattern = Counter()  # UF:esperado_set->ddd_obtido
    samples: list[dict] = []

    cur.execute("SELECT nome, telefone, cidade, uf FROM core_obras_plus_phone")
    batch = 0
    while True:
        rows = cur.fetchmany(50_000)
        if not rows:
            break
        for row in rows:
            totals["total"] += 1
            uf = normalize_text(row["uf"])
            city = normalize_text(row["cidade"])
            tel = row["telefone"]
            ddd = extract_ddd(tel)

            if not uf:
                totals["sem_uf"] += 1
                continue

            if ddd is None:
                totals["formato_invalido"] += 1
                if len(samples) < sample_limit:
                    samples.append(
                        {
                            "motivo": "formato_invalido",
                            "uf": uf,
                            "cidade": city,
                            "telefone": tel,
                            "ddd": "",
                            "ddds_uf": "",
                            "nome": row["nome"],
                        }
                    )
                continue

            expected = uf_ddds.get(uf)
            if not expected:
                totals["uf_sem_mapa"] += 1
                continue

            if ddd in expected:
                totals["ok"] += 1
            else:
                totals["ddd_outro_estado"] += 1
                by_uf_invalid[uf] += 1
                by_pattern[f"{uf}:{ddd}"] += 1
                if len(samples) < sample_limit:
                    samples.append(
                        {
                            "motivo": "ddd_outro_estado",
                            "uf": uf,
                            "cidade": city,
                            "telefone": tel,
                            "ddd": ddd,
                            "ddds_uf": ",".join(sorted(expected)),
                            "nome": row["nome"],
                        }
                    )
        batch += 1
        if batch % 20 == 0:
            print(f"... processadas {totals['total']:,} linhas", flush=True)

    conn.close()
    return totals, by_uf_invalid, by_pattern, samples


def write_report(
    out_path: Path,
    uf_ddds: dict[str, set[str]],
    totals: Counter,
    by_uf_invalid: Counter,
    by_pattern: Counter,
    samples: list[dict],
) -> None:
    lines: list[str] = []
    lines.append("# Diagnóstico DDD x UF (core.db produção)")
    lines.append("")
    lines.append("Regra: o DDD do telefone deve pertencer aos DDDs do **estado (UF)** do contato.")
    lines.append("Ex.: Americana/SP com DDD 11 = OK. Americana/SP com DDD 61 (DF) = inválido.")
    lines.append("")
    lines.append("## Mapa UF → DDDs")
    for uf in sorted(uf_ddds):
        lines.append(f"- **{uf}**: {', '.join(sorted(uf_ddds[uf]))}")
    lines.append("")
    lines.append("## Totais")
    total = totals["total"] or 1
    for key in ("total", "ok", "ddd_outro_estado", "formato_invalido", "sem_uf", "uf_sem_mapa"):
        n = totals[key]
        lines.append(f"- **{key}**: {n:,} ({100 * n / total:.1f}%)")
    lines.append("")
    lines.append("## Top UF com telefone de outro estado")
    for uf, n in by_uf_invalid.most_common(27):
        lines.append(f"- {uf}: {n:,}")
    lines.append("")
    lines.append("## Top padrões UF + DDD indevido")
    for pattern, n in by_pattern.most_common(40):
        lines.append(f"- {pattern}: {n:,}")
    lines.append("")
    lines.append("## Amostras")
    for s in samples:
        lines.append(
            f"- [{s['motivo']}] {s['uf']}/{s['cidade']} tel={s['telefone']} "
            f"ddd={s['ddd']} ddds_uf=[{s['ddds_uf']}] nome={s['nome']}"
        )
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def main() -> None:
    parser = argparse.ArgumentParser(description="Diagnóstico DDD x UF (read-only)")
    parser.add_argument("--db", required=True)
    parser.add_argument("--uf-ddds", required=True, help="CSV uf,ddds")
    parser.add_argument("--out", required=True)
    args = parser.parse_args()

    uf_ddds = load_uf_ddd_map(Path(args.uf_ddds))
    print(f"UFs no mapa: {len(uf_ddds)}")
    for uf in ("SP", "PR", "DF", "AM"):
        print(f"  {uf}: {sorted(uf_ddds.get(uf, []))}")

    totals, by_uf_invalid, by_pattern, samples = diagnose_phones(Path(args.db), uf_ddds)
    write_report(Path(args.out), uf_ddds, totals, by_uf_invalid, by_pattern, samples)
    print(f"Relatório: {args.out}")
    print(dict(totals))


if __name__ == "__main__":
    main()
