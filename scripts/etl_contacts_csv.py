#!/usr/bin/env python3
"""
ETL Script to load contacts from CSV into SQLite database
Populates core_obras_plus_email and core_obras_plus_phone tables
"""

import csv
import sqlite3
import hashlib
import uuid
import unicodedata
import os
import sys
from datetime import datetime
from pathlib import Path

def normalize_text(text):
    """
    Normalize text: remove accents and convert to uppercase
    """
    if not text:
        return ''
    
    # Remove accents
    nfkd_form = unicodedata.normalize('NFKD', str(text))
    text_without_accents = ''.join([c for c in nfkd_form if not unicodedata.combining(c)])
    
    # Convert to uppercase
    return text_without_accents.upper().strip()

def normalize_phone(phone):
    """
    Normalize phone number: remove all non-digit characters
    """
    if not phone:
        return ''
    return ''.join(filter(str.isdigit, str(phone)))

def normalize_email(email):
    """
    Normalize email: convert to lowercase
    """
    if not email:
        return ''
    return str(email).lower().strip()

def md5_hash(text):
    """
    Generate MD5 hash of text
    """
    if not text:
        return ''
    return hashlib.md5(text.encode('utf-8')).hexdigest()

def make_contact_id(telephone=None, email=None):
    """
    Generate contact_id following the TypeScript logic:
    let hash = uuidv4().replaceAll('-', '')
    let type = this.telephone ? 'phone' : this.email ? 'email' : ''
    return `cont_${type}_${hash}`
    """
    hash_value = str(uuid.uuid4()).replace('-', '')
    contact_type = 'phone' if telephone else 'email' if email else ''
    return f"cont_{contact_type}_{hash_value}"

def make_person_id(nome):
    """
    Generate person_id: pers_ + md5(nome)
    """
    return f"pers_{md5_hash(nome)}"

def process_csv_to_database(csv_path, db_path, batch_size=1000):
    """
    Process CSV file and insert data into SQLite database
    """
    if not os.path.exists(csv_path):
        print(f"Error: CSV file not found: {csv_path}")
        return False
    
    if not os.path.exists(db_path):
        print(f"Error: Database file not found: {db_path}")
        return False
    
    print(f"Starting ETL process...")
    print(f"CSV file: {csv_path}")
    print(f"Database: {db_path}")
    
    # Connect to database
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    # Statistics
    total_rows = 0
    phone_inserts = 0
    email_inserts = 0
    errors = 0
    
    # Read CSV and process
    try:
        with open(csv_path, 'r', encoding='utf-8') as csvfile:
            reader = csv.DictReader(csvfile)
            
            phone_batch = []
            email_batch = []
            
            for row in reader:
                total_rows += 1
                
                try:
                    # Extract and normalize data
                    nome = normalize_text(row.get('Nome', ''))
                    telefone = normalize_phone(row.get('Telefone', ''))
                    email = normalize_email(row.get('Email', ''))
                    cidade = normalize_text(row.get('Município', '') or row.get('Cidade', ''))
                    uf = normalize_text(row.get('UF', ''))
                    bairro = normalize_text(row.get('Bairro', ''))
                    
                    if not nome:
                        print(f"Warning: Skipping row {total_rows} - missing name")
                        continue
                    
                    # Generate IDs
                    person_id = make_person_id(nome)
                    
                    # Current timestamp
                    timestamp = int(datetime.now().timestamp())
                    
                    # Source identifier
                    source = 'contatoszap_brasilia_df'
                    
                    # Insert into phone table if telephone exists
                    if telefone:
                        contact_id = make_contact_id(telephone=telefone)
                        phone_batch.append((
                            contact_id,
                            person_id,
                            None,  # company_id
                            nome,
                            telefone,
                            cidade,
                            uf,
                            None,  # poder_aquisitivo
                            source,
                            timestamp
                        ))
                        
                        if len(phone_batch) >= batch_size:
                            cursor.executemany(
                                """INSERT INTO core_obras_plus_phone 
                                (contact_id, person_id, company_id, nome, telefone, cidade, uf, poder_aquisitivo, source, _sling_loaded_at)
                                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)""",
                                phone_batch
                            )
                            phone_inserts += len(phone_batch)
                            conn.commit()
                            print(f"Inserted {phone_inserts} phone records...")
                            phone_batch = []
                    
                    # Insert into email table if email exists
                    if email:
                        contact_id = make_contact_id(email=email)
                        email_batch.append((
                            contact_id,
                            person_id,
                            None,  # company_id
                            nome,
                            email,
                            cidade,
                            uf,
                            None,  # poder_aquisitivo
                            source,
                            timestamp
                        ))
                        
                        if len(email_batch) >= batch_size:
                            cursor.executemany(
                                """INSERT INTO core_obras_plus_email 
                                (contact_id, person_id, company_id, nome, email, cidade, uf, poder_aquisitivo, source, _sling_loaded_at)
                                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)""",
                                email_batch
                            )
                            email_inserts += len(email_batch)
                            conn.commit()
                            print(f"Inserted {email_inserts} email records...")
                            email_batch = []
                    
                    if total_rows % 10000 == 0:
                        print(f"Processed {total_rows} rows...")
                
                except Exception as e:
                    errors += 1
                    print(f"Error processing row {total_rows}: {e}")
                    continue
            
            # Insert remaining batches
            if phone_batch:
                cursor.executemany(
                    """INSERT INTO core_obras_plus_phone 
                    (contact_id, person_id, company_id, nome, telefone, cidade, uf, poder_aquisitivo, source, _sling_loaded_at)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)""",
                    phone_batch
                )
                phone_inserts += len(phone_batch)
                conn.commit()
            
            if email_batch:
                cursor.executemany(
                    """INSERT INTO core_obras_plus_email 
                    (contact_id, person_id, company_id, nome, email, cidade, uf, poder_aquisitivo, source, _sling_loaded_at)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)""",
                    email_batch
                )
                email_inserts += len(email_batch)
                conn.commit()
        
        # Print summary
        print("\n" + "="*60)
        print("ETL Process Completed")
        print("="*60)
        print(f"Total rows processed: {total_rows}")
        print(f"Phone records inserted: {phone_inserts}")
        print(f"Email records inserted: {email_inserts}")
        print(f"Errors: {errors}")
        print("="*60)
        
        return True
        
    except Exception as e:
        print(f"Fatal error: {e}")
        conn.rollback()
        return False
    
    finally:
        conn.close()

def main():
    """
    Main execution function
    """
    # Define paths
    project_root = Path(__file__).parent.parent
    csv_path = project_root / "contatos" / "500.000 contatos_Brasilia_DF - -- contatoszap.com --.csv"
    db_path = project_root / "data" / "backend" / "core.db"
    
    # Allow command line arguments to override paths
    if len(sys.argv) > 1:
        csv_path = Path(sys.argv[1])
    if len(sys.argv) > 2:
        db_path = Path(sys.argv[2])
    
    # Run ETL
    success = process_csv_to_database(str(csv_path), str(db_path))
    
    if success:
        print("\n✓ ETL process completed successfully!")
        sys.exit(0)
    else:
        print("\n✗ ETL process failed!")
        sys.exit(1)

if __name__ == "__main__":
    main()
