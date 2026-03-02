import sqlite3
import random
import uuid
import datetime
import string
import os

SOURCE_DB = 'data/backend/main/data.db'
TARGET_DB = 'data/main/data.db'

def generate_id():
    # PocketBase normally uses 15 chars for IDs
    return random.choice(string.ascii_lowercase) + ''.join(random.choices(string.ascii_lowercase + string.digits, k=14))

def main():
    if not os.path.exists('data/backend/main/data.db'):
        print(f"Erro: banco de origem não encontrado: {SOURCE_DB}")
        return

    # Certifique-se de que o diretório de destino existe
    os.makedirs(os.path.dirname(TARGET_DB), exist_ok=True)
    
    if os.path.exists(TARGET_DB):
        os.remove(TARGET_DB)

    print(f"Criando: {TARGET_DB}")
    conn_target = sqlite3.connect(TARGET_DB)
    conn_source = sqlite3.connect(SOURCE_DB)
    
    conn_target.row_factory = sqlite3.Row
    conn_source.row_factory = sqlite3.Row

    # 1. Copiar schema
    print("Extraindo schema...", end="")
    schema_query = "SELECT sql FROM sqlite_master WHERE sql IS NOT NULL AND name NOT LIKE 'sqlite_%';"
    schemas = conn_source.execute(schema_query).fetchall()
    
    for s in schemas:
        try:
            conn_target.execute(s['sql'])
        except sqlite3.OperationalError as e:
            # Algumas views podem falhar se as tabelas ainda não existirem
            print(f"  [Notice] Erro não critico ao recriar schema: {e}")
            pass
    print("OK")

    # 2. Copiar as tabelas estáticas essenciais
    static_tables = ['_migrations', '_collections']
    for table in static_tables:
        print(f"Copiando tabela {table}...")
        rows = conn_source.execute(f"SELECT * FROM {table}").fetchall()
        if not rows: continue
        
        columns = rows[0].keys()
        placeholders = ', '.join(['?' for _ in columns])
        insert_sql = f"INSERT INTO {table} ({', '.join(columns)}) VALUES ({placeholders})"
        
        for row in rows:
            conn_target.execute(insert_sql, tuple(row))
            
    conn_target.commit()

    # 3. Gerar Dados Fake

    now = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%fZ")
    
    # 3.1. Um admin falso
    print("Gerando admins falsos...")
    fake_admin_id = generate_id()
    conn_target.execute("""
        INSERT INTO _admins (id, email, tokenKey, passwordHash, created, updated)
        VALUES (?, ?, ?, ?, ?, ?)
    """, (fake_admin_id, 'admin@teste.sandbox', generate_id()+'a', '$2a$10$vI8aWBnW3fID.ZQ4/zo1G.q1lRps.9cGLcZEiGTRVm5Ruleiq8VGu', now, now))
    # Senha hashada do pocketbase para '1234567890' (exemplo, ou algo padrão)

    # 3.2. Equipes falsas
    print("Gerando equipes falsas...")
    teams = []
    for i in range(2):
        team_id = generate_id()
        conn_target.execute("""
            INSERT INTO team (id, name, active, created, updated)
            VALUES (?, ?, ?, ?, ?)
        """, (team_id, f'Equipe Dev {i+1}', True, now, now))
        teams.append(team_id)

    # 3.3. Usuários falsos
    print("Gerando usuários falsos...")
    users = []
    for i in range(5):
        user_id = generate_id()
        team = random.choice(teams)
        email = f"user_{i}@teste.local"
        name = f"Usuario Falso {i}"
        
        conn_target.execute("""
            INSERT INTO user (id, email, username, name, passwordHash, tokenKey, team_id, verified, created, updated)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """, (user_id, email, f"user{i}", name, '$2a$10$vI8aWBnW3fID.ZQ4/zo1G.q1lRps.9cGLcZEiGTRVm5Ruleiq8VGu', generate_id()+'u', team, True, now, now))
        users.append(user_id)
        
    # 3.4. Obras fake (leads mock)
    print("Gerando leads falsos...")
    leads = []
    for i in range(10):
        lead_id = generate_id()
        obra_id = f"obra_{random.randint(1000, 9999)}"
        team_id = random.choice(teams)
        owner_id = random.choice(users)
        conn_target.execute("""
            INSERT INTO lead (id, obra_id, team_id, owner_id, created, updated)
            VALUES (?, ?, ?, ?, ?, ?)
        """, (lead_id, obra_id, team_id, owner_id, now, now))
        leads.append(lead_id)

    # 3.5. List/Stages Fake e list leads
    print("Gerando listas falsas...")
    for team in teams:
        list_id = generate_id()
        conn_target.execute("INSERT INTO list (id, name, team_id, created, updated) VALUES (?, ?, ?, ?, ?)", 
                            (list_id, f"Lista Teste", team, now, now))
        
        stage1 = generate_id()
        stage2 = generate_id()
        conn_target.execute('INSERT INTO list_stage (id, list_id, name, "order", created, updated) VALUES (?, ?, ?, ?, ?, ?)', (stage1, list_id, "Iniciada", 1, now, now))
        conn_target.execute('INSERT INTO list_stage (id, list_id, name, "order", created, updated) VALUES (?, ?, ?, ?, ?, ?)', (stage2, list_id, "Andamento", 2, now, now))

    conn_target.commit()
    conn_target.close()
    conn_source.close()
    
    print(f"Database FAKE criado com sucesso em: {TARGET_DB}")
    
if __name__ == '__main__':
    main()
