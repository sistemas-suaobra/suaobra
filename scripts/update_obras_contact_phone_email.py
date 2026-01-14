import sqlite3
import sys

db_path = 'data/backend/core.db'

conn = sqlite3.connect(db_path)
cursor = conn.cursor()

# Carrega todos os emails e telefones em sets para lookup rápido
cursor.execute("SELECT nome FROM core_obras_plus_email")
email_records = set((row[0]) for row in cursor.fetchall())

cursor.execute("SELECT nome FROM core_obras_plus_phone")
phone_records = set((row[0]) for row in cursor.fetchall())

# Seleciona linhas com colunas nulas
cursor.execute("""
    SELECT id, owner, professional
    FROM core_obras_plus
    WHERE has_owner_phone = 0
       OR has_owner_email = 0
       OR has_professional_phone = 0
       OR has_professional_email = 0
""")
rows = cursor.fetchall()
total = len(rows)
print(f"Encontradas {total} linhas para atualizar.")

# Usa uma transação para atualizar em lote
for idx, row in enumerate(rows, 1):
    obra_id, owner, professional = row

    has_owner_email = 1 if owner in email_records else 0
    has_professional_email = 1 if professional in email_records else 0
    has_owner_phone = 1 if owner in phone_records else 0
    has_professional_phone = 1 if professional in phone_records else 0

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