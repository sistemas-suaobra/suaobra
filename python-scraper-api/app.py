from flask import Flask, jsonify
from playwright.sync_api import sync_playwright, TimeoutError
import re
import hashlib

app = Flask(__name__)

def extract_data(text):
    def extract(pattern, content):
        match = re.search(pattern, content, re.IGNORECASE)
        if match and len(match.groups()) > 0:
            return match.group(1).strip()
        return ""

    data = {}
    
    # Função para formatar datas
    def format_date(date_str):
        import datetime
        try:
            return datetime.datetime.strptime(date_str, "%d/%m/%Y").strftime("%Y-%m-%d")
        except Exception:
            return ""

    # Extrair dados básicos
    data['obra_number'] = extract(r'Número do RRT:\s*(\d+)', text)
    if not data['obra_number']:
        raise ValueError("Não foi possível extrair o número da obra.")
    md5_hasher = hashlib.md5()
    md5_hasher.update(data['obra_number'].encode('utf-8'))
    md5_hash = md5_hasher.hexdigest()
    data['id'] = "obra_" + md5_hash
    data['professional'] = extract(r'Arquiteto\(a\) e Urbanista:\s*(.*)', text)
    ownerValue = extract(r'Nome/Razão Social:\s*(.*)', text)
    if ownerValue.startswith("CPF / CNPJ"):
        data['owner'] = ""
    else:
        data['owner'] = ownerValue
    data['bairro'] = extract(r'Bairro:\s*(.*)', text)
    data['start_date'] = format_date(extract(r'Data de Início:\s*(.*)', text))
    data['end_date'] = format_date(extract(r'Previsão de Término:\s*(.*)', text))
    data['first_listing_date'] = format_date(extract(r'Data de Registro:\s*(.*)', text))

    # Cidade e Estado
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

    # Atividade e Tipo
    activityStr = extract(r'Atividade Subordinada([\s\S]*?)Pagamento', text)
    data['activity'] = extract(r'(\d+\.\d+\.\d+\s*-\s*[^\d\s-][\s\S]*?)[\s]*[\d]', activityStr)
    reType = re.compile(r'([\d\.]+)\s*([\wÇÃÁÉÍÓÚ]+)\s*>\s*(\d+)\s*>\s*([\d\.]+\s*-\s*[\w\sÇÃÁÉÍÓÚ]+)', re.IGNORECASE)
    matchType = reType.search(activityStr)
    if matchType:
        data['type'] = f"{matchType.group(3)} - {matchType.group(2)}"
    else:
        data['type'] = extract(r'Modalidade:\s*(.*)', text)

    # Tamanho e Unidade
    reSizeUnit = re.compile(r'(\d+\.?\d*)\s*/\s*([\w\s²]+)', re.IGNORECASE)
    matchSizeUnit = reSizeUnit.search(activityStr)
    if matchSizeUnit:
        sizeValue = matchSizeUnit.group(1).split('.')[0]
        try:
            data['size'] = float(sizeValue)
        except Exception:
            data['size'] = 0.0
        data['unidade'] = matchSizeUnit.group(2)
    else:
        data['size'] = 0.0
        data['unidade'] = ""

    # Endereço completo
    tipoLogradouro = extract(r'Tipo de Logradouro:\s*(.*?)[\r\n]', text)
    logradouro = extract(r'(?:\n|^)Logradouro:\s*(.*)', text)
    numero = extract(r'Número/Ano:\s*(.*)', text)
    complemento = extract(r'Complemento:\s*(.*)', text)
    data['address'] = f"{tipoLogradouro} {logradouro}, {numero}, {complemento}, {data['bairro']} - {data['city']}, {data['state']}"

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
