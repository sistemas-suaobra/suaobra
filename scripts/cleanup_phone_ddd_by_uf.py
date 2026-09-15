#!/usr/bin/env python3
"""Remove telefones cujo DDD não pertence à UF do contato.

1) Cria tabela de backup dos removidos
2) DELETE dos incompatíveis
3) Recalcula has_owner_phone / has_professional_phone / has_owner_email / has_professional_email
"""

from __future__ import annotations

import argparse
import csv
import sqlite3
import time
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
    # 8+ cobre incompletos (ex.: 489737451) além de 10/11 válidos
    if len(d) >= 8:
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


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--db", required=True)
    parser.add_argument("--uf-ddds", required=True)
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    uf_ddds = load_uf_ddd_map(Path(args.uf_ddds))
    print(f"UFs no mapa: {len(uf_ddds)}")

    conn = sqlite3.connect(args.db, timeout=600)
    conn.execute("PRAGMA journal_mode=WAL")
    conn.execute("PRAGMA synchronous=NORMAL")
    conn.execute("PRAGMA temp_store=FILE")
    conn.row_factory = sqlite3.Row
    cur = conn.cursor()

    t0 = time.time()
    print("Identificando telefones com DDD de outro estado...")

    bad_rowids: list[int] = []
    stats = Counter()

    cur.execute("SELECT rowid, telefone, uf FROM core_obras_plus_phone")
    while True:
        rows = cur.fetchmany(50_000)
        if not rows:
            break
        for row in rows:
            stats["scanned"] += 1
            uf = normalize_text(row["uf"])
            if not uf:
                continue
            expected = uf_ddds.get(uf)
            if not expected:
                continue
            ddd = extract_ddd(row["telefone"])
            if ddd is None:
                continue
            if ddd not in expected:
                bad_rowids.append(int(row["rowid"]))
                stats["bad"] += 1
                stats[f"uf:{uf}"] += 1
        if stats["scanned"] % 1_000_000 == 0:
            print(f"  scanned={stats['scanned']:,} bad={stats['bad']:,}", flush=True)

    print(f"Scanned={stats['scanned']:,} to_delete={stats['bad']:,} ({time.time()-t0:.1f}s)")
    if args.dry_run:
        print("DRY-RUN: nada alterado")
        for k, v in stats.most_common(40):
            if str(k).startswith("uf:"):
                print(f"  {k}: {v:,}")
        conn.close()
        return

    if not bad_rowids:
        print("Nada para apagar")
        conn.close()
        return

    backup_table = "core_obras_plus_phone_ddd_cleanup_bak"
    print(f"Usando backup {backup_table}...")
    cur.execute(
        f"""
        CREATE TABLE IF NOT EXISTS {backup_table} AS
        SELECT p.*, p.rowid AS _orig_rowid
        FROM core_obras_plus_phone p
        WHERE 0
        """
    )

    print("Copiando removidos para backup...")
    chunk = 5_000
    for i in range(0, len(bad_rowids), chunk):
        part = bad_rowids[i : i + chunk]
        placeholders = ",".join("?" * len(part))
        cur.execute(
            f"""
            INSERT INTO {backup_table}
            SELECT p.*, p.rowid AS _orig_rowid
            FROM core_obras_plus_phone p
            WHERE p.rowid IN ({placeholders})
            """,
            part,
        )
        if (i // chunk) % 20 == 0:
            print(f"  backup {min(i + chunk, len(bad_rowids)):,}/{len(bad_rowids):,}", flush=True)
            conn.commit()
    conn.commit()
    bak_count = cur.execute(f"SELECT COUNT(*) FROM {backup_table}").fetchone()[0]
    print(f"Backup rows: {bak_count:,}")

    print("Apagando telefones incompatíveis...")
    deleted = 0
    for i in range(0, len(bad_rowids), chunk):
        part = bad_rowids[i : i + chunk]
        placeholders = ",".join("?" * len(part))
        cur.execute(
            f"DELETE FROM core_obras_plus_phone WHERE rowid IN ({placeholders})",
            part,
        )
        deleted += cur.rowcount
        if (i // chunk) % 20 == 0:
            print(f"  deleted {deleted:,}/{len(bad_rowids):,}", flush=True)
            conn.commit()
    conn.commit()
    print(f"Deleted={deleted:,}")

    print("Recalculando flags has_* ...")
    t1 = time.time()
    cur.execute(
        'SELECT DISTINCT nome FROM core_obras_plus_phone WHERE nome IS NOT NULL AND nome != ""'
    )
    phone_names = {r[0] for r in cur.fetchall()}
    cur.execute(
        'SELECT DISTINCT nome FROM core_obras_plus_email WHERE nome IS NOT NULL AND nome != ""'
    )
    email_names = {r[0] for r in cur.fetchall()}
    print(f"  nomes com telefone: {len(phone_names):,}")
    print(f"  nomes com email: {len(email_names):,}")

    cur.execute(
        """
        SELECT rowid, owner, professional,
               has_owner_phone, has_professional_phone,
               has_owner_email, has_professional_email
        FROM core_obras_plus
        """
    )

    updates = []
    scanned_obras = 0
    changed = 0
    while True:
        rows = cur.fetchmany(20_000)
        if not rows:
            break
        for row in rows:
            scanned_obras += 1
            rid = row[0]
            owner = row[1] or ""
            professional = row[2] or ""
            new_op = 1 if owner in phone_names else 0
            new_pp = 1 if professional in phone_names else 0
            new_oe = 1 if owner in email_names else 0
            new_pe = 1 if professional in email_names else 0
            if (
                new_op != (row[3] or 0)
                or new_pp != (row[4] or 0)
                or new_oe != (row[5] or 0)
                or new_pe != (row[6] or 0)
            ):
                updates.append((new_op, new_pp, new_oe, new_pe, rid))
                changed += 1
        if updates and len(updates) >= 5_000:
            conn.executemany(
                """
                UPDATE core_obras_plus
                SET has_owner_phone=?,
                    has_professional_phone=?,
                    has_owner_email=?,
                    has_professional_email=?
                WHERE rowid=?
                """,
                updates,
            )
            conn.commit()
            updates.clear()
            print(f"  obras scanned={scanned_obras:,} changed={changed:,}", flush=True)

    if updates:
        conn.executemany(
            """
            UPDATE core_obras_plus
            SET has_owner_phone=?,
                has_professional_phone=?,
                has_owner_email=?,
                has_professional_email=?
            WHERE rowid=?
            """,
            updates,
        )
        conn.commit()

    remaining = cur.execute("SELECT COUNT(*) FROM core_obras_plus_phone").fetchone()[0]
    print(f"Flags: scanned={scanned_obras:,} changed={changed:,} ({time.time() - t1:.1f}s)")
    print(f"Phones remaining: {remaining:,}")
    print(f"TOTAL elapsed={time.time() - t0:.1f}s")
    print(f"Backup table: {backup_table}")
    conn.close()


if __name__ == "__main__":
    main()
