#!/usr/bin/env python3
"""Gera Excel de obras a partir do CSV extraído do core.db (últimos 60 dias)."""

from __future__ import annotations

import argparse
import csv
from datetime import date, datetime, timedelta
from pathlib import Path

from openpyxl import Workbook

HEADERS = [
    "NUMERO_OBRA",
    "ENDERECO",
    "BAIRRO",
    "CIDADE",
    "UF",
    "TAMANHO",
    "UNIDADE",
    "TIPO",
    "ATIVIDADE",
    "ETAPA",
    "DATA_DE_INICIO",
    "PREVISAO_DE_TERMINO",
    "DATA_LISTAGEM",
    "PROPRIETARIO",
    "PROFISSIONAL",
    "PROPRIETARIO_TELEFONE",
    "PROFISSIONAL_TELEFONE",
    "PROPRIETARIO_EMAIL",
    "PROFISSIONAL_EMAIL",
    "PROPRIETARIO_TELEFONES_OUTROS",
    "PROFISSIONAL_TELEFONES_OUTROS",
    "PROPRIETARIO_EMAILS_OUTROS",
    "PROFISSIONAL_EMAILS_OUTROS",
]

def parse_date(value: str | None) -> date | None:
    if not value or not str(value).strip():
        return None
    text = str(value).strip()[:10]
    try:
        return datetime.strptime(text, "%Y-%m-%d").date()
    except ValueError:
        return None


def is_projeto_obra(type_val: str) -> bool:
    return (type_val or "").strip().upper().startswith("1 - PROJETO")


def calculate_obra_stage(start_raw: str, end_raw: str, as_of: date, type_val: str = "") -> str:
    if is_projeto_obra(type_val):
        return ""
    start = parse_date(start_raw)
    end = parse_date(end_raw)
    if start is None or end is None:
        return ""
    if as_of > end:
        return "FINALIZADA"
    total = (end - start).total_seconds()
    elapsed = (as_of - start).total_seconds()
    if elapsed < 0:
        return "INICIO"
    if total <= 0:
        return "FINALIZADA"
    percentage = (elapsed / total) * 100
    if percentage <= 33:
        return "INICIO"
    if percentage <= 66:
        return "ESTRUTURA"
    return "ACABAMENTO"


def fmt_date(value: str) -> str:
    parsed = parse_date(value)
    if parsed is None:
        return (value or "").strip()
    return parsed.strftime("%d/%m/%Y")


def row_from_csv(rec: dict, as_of: date) -> list:
    return [
        rec.get("obra_number") or "",
        rec.get("address") or "",
        rec.get("bairro") or "",
        rec.get("city") or "",
        rec.get("state") or "",
        rec.get("size") or "",
        rec.get("unidade") or "",
        rec.get("type") or "",
        rec.get("activity") or "",
        calculate_obra_stage(rec.get("start_date") or "", rec.get("end_date") or "", as_of, rec.get("type") or ""),
        fmt_date(rec.get("start_date") or ""),
        fmt_date(rec.get("end_date") or ""),
        fmt_date(rec.get("first_listing_date") or ""),
        rec.get("owner") or "",
        rec.get("professional") or "",
        rec.get("owner_first_telefone") or "",
        rec.get("professional_first_telefone") or "",
        rec.get("owner_first_email") or "",
        rec.get("professional_first_email") or "",
        rec.get("owner_telefones") or "",
        rec.get("professional_telefones") or "",
        rec.get("owner_emails") or "",
        rec.get("professional_emails") or "",
    ]


def write_xlsx(csv_path: Path, xlsx_path: Path, as_of: date) -> int:
    wb = Workbook(write_only=True)
    ws = wb.create_sheet("Obras")
    ws.append(HEADERS)

    count = 0
    with csv_path.open(newline="", encoding="utf-8") as fh:
        reader = csv.DictReader(fh)
        for rec in reader:
            ws.append(row_from_csv(rec, as_of))
            count += 1

    xlsx_path.parent.mkdir(parents=True, exist_ok=True)
    wb.save(xlsx_path)
    return count


def main() -> None:
    parser = argparse.ArgumentParser(description="CSV do core.db → Excel de obras (60 dias).")
    parser.add_argument("--csv", required=True, help="CSV gerado pelo export_obras_60d.sql")
    parser.add_argument("--out", required=True, help="Caminho do .xlsx de saída")
    parser.add_argument("--ref-date", default="2026-08-15", help="Data de referência (YYYY-MM-DD)")
    args = parser.parse_args()

    as_of = datetime.strptime(args.ref_date, "%Y-%m-%d").date()
    csv_path = Path(args.csv)
    xlsx_path = Path(args.out)
    n = write_xlsx(csv_path, xlsx_path, as_of)
    window_start = as_of - timedelta(days=60)
    print(f"Linhas: {n}")
    print(f"Janela: {window_start.isoformat()} a {as_of.isoformat()}")
    print(f"Arquivo: {xlsx_path}")


if __name__ == "__main__":
    main()
