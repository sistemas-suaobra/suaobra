from flask import Flask, jsonify
from playwright.sync_api import sync_playwright, TimeoutError
import re

app = Flask(__name__)

def extract_data(text):
    def extract(pattern, content):
        match = re.search(pattern, content, re.IGNORECASE)
        if match and len(match.groups()) > 0:
            return match.group(1).strip()
        return ""

    data = {}
    
    # Extrair dados básicos
    data['obra_number'] = extract(r'Número do RRT:\s*(.*)', text)
    data['professional'] = extract(r'Arquiteto\(a\) e Urbanista:\s*(.*)', text)
    data['owner'] = extract(r'Nome/Razão Social:\s*(.*)', text) or extract(r'CPF / CNPJ:\s*(.*)', text)
    data['start_date'] = extract(r'Data de Início:\s*(.*)', text)
    data['end_date'] = extract(r'Previsão de Término:\s*(.*)', text)
    data['bairro'] = extract(r'Bairro:\s*(.*)', text)
    data['type'] = extract(r'Modalidade:\s*(.*)', text)
    data['first_listing_date'] = extract(r'Data de Registro:\s*(.*)', text)

    # Extrair cidade e estado
    city_state = extract(r'Cidade/UF:\s*(.*)', text)
    if city_state:
        parts = city_state.split('/')
        if len(parts) == 2:
            data['city'] = parts[0].strip()
            data['state'] = parts[1].strip()
        else:
            data['city'] = ""
            data['state'] = ""
    else:
        data['city'] = ""
        data['state'] = ""

    # Extrair atividade
    data['activity'] = extract(r'1\.  (.*?)(?:\n|$)', text)

    # Extrair tamanho e unidade
    size_unit = extract(r'EXECUÇÃO DE OBRA[\s\w]*([\d\.,]+\s*/\s*.*)', text)
    if size_unit:
        parts = size_unit.split('/')
        if len(parts) == 2:
            data['unidade'] = parts[1].strip()
            size_str = parts[0].strip().replace('.', '').replace(',', '.')
            try:
                data['size'] = float(size_str)
            except ValueError:
                data['size'] = 0.0
        else:
            data['size'] = 0.0
            data['unidade'] = ""
    else:
        data['size'] = 0.0
        data['unidade'] = ""

    # Extrair endereço
    logradouro = extract(r'Logradouro:\s*(.*)', text)
    numero = extract(r'Número/Ano:\s*(.*)', text)
    complemento = extract(r'Complemento:\s*(.*)', text)
    data['address'] = f"{logradouro}, {numero}, {complemento}".strip(', ')

    return data

def get_document_data(numero_documento: str):
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        try:
            context = browser.new_context(
                viewport={'width': 1280, 'height': 720},
                user_agent='Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36'
            )
            page = context.new_page()
            url = f"https://acesso.caubr.gov.br/autenticidade/rrt?numero={numero_documento}&retificador="

            # Configurar timeouts
            page.set_default_timeout(60000)
            page.set_default_navigation_timeout(60000)

            # Carregar página
            response = page.goto(url, wait_until="domcontentloaded")
            if not response or response.status != 200:
                return None

            # Esperar e extrair conteúdo
            page.wait_for_timeout(2000)
            main_content = page.locator("main")
            if not main_content.is_visible():
                main_content = page.locator("body")

            content = main_content.inner_text()
            if not content or len(content.strip()) < 100:
                return None

            return extract_data(content)

        except TimeoutError as e:
            return {"error": f"Timeout ao acessar RRT: {str(e)}"}
        except Exception as e:
            return {"error": f"Erro ao processar RRT: {str(e)}"}
        finally:
            browser.close()

@app.route('/rrt/<numero>')
def get_rrt(numero):
    try:
        data = get_document_data(numero)
        if data is None:
            return jsonify({"error": "Dados não encontrados"}), 404
        if "error" in data:
            return jsonify(data), 500
        return jsonify(data)
    except Exception as e:
        return jsonify({"error": f"Erro interno: {str(e)}"}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
