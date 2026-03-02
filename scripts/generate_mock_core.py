import sqlite3
import random
import uuid
import datetime
import string
import os

SOURCE_DB = 'data/backend/core.db'
TARGET_DB = 'data/core/core.db'

def generate_id():
    # PocketBase normally uses 15 chars for IDs
    return random.choice(string.ascii_lowercase) + ''.join(random.choices(string.ascii_lowercase + string.digits, k=14))

def fake_name(prefix="Nome Falso"):
    return f"{prefix} {random.randint(100, 999)}"

def main():
    if not os.path.exists(SOURCE_DB):
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
            # Algumas views podem falhar se as tabelas ainda não existirem, ignora na criação de schema inicial
            print(f"\  [Notice] Erro não critico ao recriar schema: {e}")
            pass
    print("OK")

    # 2. Diferente do Pocketbase principal, o core.db não usa _migrations ou _collections. 
    # Podemos pular a cópia de tabelas estáticas do sistema e ir direto pra gerar os dados mocks
    conn_target.commit()

    # 3. Gerar Dados Fake para core_obras_plus
    now = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%fZ")
    
    print("Gerando dados para core_obras_plus...")
    obras = []
    for i in range(20):
        obra_id = generate_id()
        num = f"{random.randint(2000, 2999)}_{random.randint(100000, 999999)}"
        owner = fake_name("Proprietario_Mock")
        professional = fake_name("Profissional_Mock")
        
        has_owner_phone = random.choice([True, False])
        has_owner_email = random.choice([True, False])
        has_prof_phone = random.choice([True, False])
        has_prof_email = random.choice([True, False])
        
        cidade = random.choice(["Curitiba", "São Paulo", "Rio de Janeiro"])
        uf = random.choice(["PR", "SP", "RJ"])
        bairro = "Centro"
        start_date = "2023-01-01"
        end_date = "2024-12-31"

        conn_target.execute("""
            INSERT INTO core_obras_plus (
                id, obra_number, owner, professional, address, bairro, city, state, 
                start_date, end_date, activity, type, size, unidade, 
                has_owner_phone, has_owner_email, has_professional_phone, has_professional_email,
                first_listing_date, last_listing_date, _sling_loaded_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """, (
            obra_id, num, owner, professional, "Rua Mock, 123", bairro, cidade, uf,
            start_date, end_date, "Residencial", "Construção Nova", random.randint(100, 1000), "m²",
            has_owner_phone, has_owner_email, has_prof_phone, has_prof_email,
            "2023-01-01", "2023-10-01", None
        ))
        
        obras.append({
            "id": obra_id,
            "owner": owner,
            "professional": professional,
            "cidade": cidade,
            "uf": uf
        })

    # 4. Gerar telefones mock
    print("Gerando dados para core_obras_plus_phone e email...")
    for obra in obras:
        # Cadastrar o owner (telefone e email opcionalmente)
        person_id = f"owner_{obra['id']}"
        conn_target.execute("INSERT INTO core_obras_plus_phone (contact_id, person_id, nome, telefone, cidade, uf) VALUES (?, ?, ?, ?, ?, ?)",
                            (generate_id(), person_id, obra['owner'], f"4199999{random.randint(1000, 9999)}", obra['cidade'], obra['uf']))
        
        conn_target.execute("INSERT INTO core_obras_plus_email (contact_id, person_id, nome, email, cidade, uf) VALUES (?, ?, ?, ?, ?, ?)",
                            (generate_id(), person_id, obra['owner'], f"{obra['owner'].replace(' ', '.').lower()}@mailinator.local", obra['cidade'], obra['uf']))
                            
        # Cadastrar o professional
        prof_id = f"prof_{obra['id']}"
        conn_target.execute("INSERT INTO core_obras_plus_phone (contact_id, person_id, nome, telefone, cidade, uf) VALUES (?, ?, ?, ?, ?, ?)",
                            (generate_id(), prof_id, obra['professional'], f"4198888{random.randint(1000, 9999)}", obra['cidade'], obra['uf']))
    
    conn_target.commit()
    conn_target.close()
    conn_source.close()
    
    print(f"Database core FAKE criado com sucesso em: {TARGET_DB}")
    
if __name__ == '__main__':
    main()
