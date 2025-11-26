import sqlite3
import sys

db_path = 'data/backend/core.db'

conn = sqlite3.connect(db_path)
cursor = conn.cursor()

# Carrega todos os emails e telefones em sets para lookup rápido
cursor.execute("SELECT nome, cidade, uf FROM core_obras_plus_email")
email_records = set((row[0], row[1], row[2]) for row in cursor.fetchall())

cursor.execute("SELECT nome, cidade, uf FROM core_obras_plus_phone")
phone_records = set((row[0], row[1], row[2]) for row in cursor.fetchall())

# Seleciona linhas com colunas nulas
cursor.execute("""
    SELECT id, owner, professional, city, state
    FROM core_obras_plus
    WHERE has_owner_phone IS NULL
       OR has_owner_email IS NULL
       OR has_professional_phone IS NULL
       OR has_professional_email IS NULL
""")
rows = cursor.fetchall()
total = len(rows)
print(f"Encontradas {total} linhas para atualizar.")

# Usa uma transação para atualizar em lote
for idx, row in enumerate(rows, 1):
    obra_id, owner, professional, city, state = row

    has_owner_email = 1 if (owner, city, state) in email_records else 0
    has_professional_email = 1 if (professional, city, state) in email_records else 0
    has_owner_phone = 1 if (owner, city, state) in phone_records else 0
    has_professional_phone = 1 if (professional, city, state) in phone_records else 0

    cursor.execute("""
        UPDATE core_obras_plus
        SET has_owner_email = ?,
            has_professional_email = ?,
            has_owner_phone = ?,
            has_professional_phone = ?
        WHERE id = ?
    """, (has_owner_email, has_professional_email, has_owner_phone, has_professional_phone, obra_id))

    if idx % 10 == 0 or idx == total:
        print(f"Processadas {idx}/{total} linhas...")
        sys.stdout.flush()

conn.commit()
conn.close()