import sqlite3
import os
from datetime import datetime

CORE_DB = 'data/backend/core.db'
DATA_DB = 'data/backend/main/data.db'
FINALIZACAO_LIMITE = '2023-01-01'

# Ajuste conforme o nome real da tabela e coluna de data de finalização
tabela_obras = 'core_obras_plus'  # Exemplo, ajuste se necessário
coluna_id_obra = 'id'   # Exemplo, ajuste se necessário
coluna_data_finalizacao = 'end_date'  # Exemplo, ajuste se necessário

# Ajuste conforme o nome real da tabela e coluna de referência à obra em data.db
tabela_referencias = 'lead'  # Exemplo, ajuste se necessário
coluna_obra_ref = 'obra_id'         # Exemplo, ajuste se necessário

def conectar_db(path):
    if not os.path.exists(path):
        print(f"Arquivo de banco de dados não encontrado: {path}")
        exit(1)
    return sqlite3.connect(path)

def buscar_obras_finalizadas(conn):
    cursor = conn.cursor()
    cursor.execute(f"""
        SELECT {coluna_id_obra}, {coluna_data_finalizacao}
        FROM {tabela_obras}
        WHERE {coluna_data_finalizacao} < ?
    """, (FINALIZACAO_LIMITE,))
    return cursor.fetchall()


# Busca leads que referenciam obras removidas
def buscar_leads_remover(conn, obras_removidas):
    cursor = conn.cursor()
    placeholders = ','.join(['?'] * len(obras_removidas))
    cursor.execute(f"SELECT id FROM lead WHERE obra_id IN ({placeholders})", obras_removidas)
    return [row[0] for row in cursor.fetchall()]

# Conta registros relacionados a leads em tabelas dependentes
def contar_dependentes(conn, tabela, campo, leads_remover):
    if not leads_remover:
        return 0
    cursor = conn.cursor()
    placeholders = ','.join(['?'] * len(leads_remover))
    cursor.execute(f"SELECT COUNT(*) FROM {tabela} WHERE {campo} IN ({placeholders})", leads_remover)
    return cursor.fetchone()[0]



def relatorio_impacto(obras_removidas, leads_remover, dependentes):
    print("==== RELATÓRIO DE IMPACTO ====")
    print(f"Obras a serem removidas de {CORE_DB}: {len(obras_removidas)}")
    print(f"Leads a serem removidos de {DATA_DB}: {len(leads_remover)}")
    for tabela, qtd in dependentes.items():
        print(f"Registros em {tabela} a serem removidos: {qtd}")
    print("A limpeza irá remover permanentemente esses registros.")
    print("Deseja continuar? (s/n)")

def remover_obras(conn, obras_removidas):
    cursor = conn.cursor()
    placeholders = ','.join(['?'] * len(obras_removidas))
    cursor.execute(f"DELETE FROM {tabela_obras} WHERE {coluna_id_obra} IN ({placeholders})", obras_removidas)
    conn.commit()


# Remove registros dependentes de leads
def remover_dependentes(conn, tabela, campo, leads_remover):
    if not leads_remover:
        return
    cursor = conn.cursor()
    placeholders = ','.join(['?'] * len(leads_remover))
    cursor.execute(f"DELETE FROM {tabela} WHERE {campo} IN ({placeholders})", leads_remover)
    conn.commit()



def main():
    core_conn = conectar_db(CORE_DB)
    data_conn = conectar_db(DATA_DB)

    obras_finalizadas = buscar_obras_finalizadas(core_conn)
    obras_removidas = [str(row[0]) for row in obras_finalizadas]

    if not obras_removidas:
        print("Nenhuma obra finalizada antes de 01/2023 encontrada. Nada a remover.")
        return

    # Buscar leads a serem removidos
    leads_remover = buscar_leads_remover(data_conn, obras_removidas)

    # Contar dependentes
    dependentes = {
        'lead_contact': contar_dependentes(data_conn, 'lead_contact', 'lead_id', leads_remover),
        'list_lead': contar_dependentes(data_conn, 'list_lead', 'lead_id', leads_remover),
        'lead_activity': contar_dependentes(data_conn, 'lead_activity', 'lead_id', leads_remover),
        'obra_note': contar_dependentes(data_conn, 'obra_note', 'obra_id', obras_removidas)
    }

    relatorio_impacto(obras_removidas, leads_remover, dependentes)

    confirm = input().strip().lower()
    if confirm != 's':
        print("Operação cancelada pelo usuário.")
        return

    remover_obras(core_conn, obras_removidas)
    # Remover dependentes antes dos leads
    remover_dependentes(data_conn, 'lead_contact', 'lead_id', leads_remover)
    remover_dependentes(data_conn, 'list_lead', 'lead_id', leads_remover)
    remover_dependentes(data_conn, 'lead_activity', 'lead_id', leads_remover)
    remover_dependentes(data_conn, 'obra_note', 'obra_id', obras_removidas)
    # Remover leads
    remover_dependentes(data_conn, 'lead', 'id', leads_remover)
    print("Limpeza concluída com sucesso.")

    core_conn.close()
    data_conn.close()

if __name__ == "__main__":
    main()
