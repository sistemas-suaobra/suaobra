import sqlite3
import sys
import unicodedata

def remove_accents(text):
    """Remove acentos e caracteres especiais de uma string."""
    if not text:
        return text
    return ''.join(c for c in unicodedata.normalize('NFD', text) if unicodedata.category(c) != 'Mn')

db_path = 'data/backend/core.db'

conn = sqlite3.connect(db_path)
cursor = conn.cursor()

# Seleciona todas as linhas com bairro e city
cursor.execute("SELECT id, bairro, city FROM core_obras_plus")
rows = cursor.fetchall()
total = len(rows)
print(f"Encontradas {total} linhas para verificar e atualizar.")

updated_count = 0

for idx, row in enumerate(rows, 1):
    obra_id, bairro, city = row

    new_bairro = remove_accents(bairro) if bairro else None
    new_city = remove_accents(city) if city else None

    # Verifica se houve mudança
    if new_bairro != bairro or new_city != city:
        cursor.execute("""
            UPDATE core_obras_plus
            SET bairro = ?, city = ?
            WHERE id = ?
        """, (new_bairro, new_city, obra_id))
        updated_count += 1

    if idx % 100 == 0 or idx == total:
        print(f"Processadas {idx}/{total} linhas... ({updated_count} atualizadas)")
        sys.stdout.flush()

conn.commit()
conn.close()

print(f"Atualização concluída. Total de linhas atualizadas: {updated_count}")
