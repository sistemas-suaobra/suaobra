import json
import sys
from playwright.sync_api import sync_playwright, TimeoutError

connection_url = "wss://browser.zenrows.com?apikey=b67313ec4485fd294ff26be3f995989e4b7ab61b"

async def get_document_data(numero_documento: str):
    async with sync_playwright() as p:
        browser = await p.chromium.connect_over_cdp(connection_url)  # Mostra o navegador durante a execução
        context = browser.new_context(
            viewport={'width': 1280, 'height': 720},  # Definindo um tamanho de tela adequado
            user_agent='Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36'  # User agent mais comum
        )
        page = await context.new_page()
        
        url = f"https://acesso.caubr.gov.br/autenticidade/rrt?numero={numero_documento}&retificador="

        try:
            print(f"Acessando URL: {url}", file=sys.stderr)
            
            # Configurações mais permissivas para carregamento lento
            page.set_default_timeout(3000)  # 3 segundos
            page.set_default_navigation_timeout(3000)  # 3 segundos

            # Primeiro carrega a página sem esperar por network idle
            response = await page.goto(url, wait_until="domcontentloaded", timeout=3000)
            if not response or response.status != 200:
                print(f"Erro: Status code {response.status if response else 'N/A'} ao acessar {url}", file=sys.stderr)
                return None
                
            # Aguarda um pouco para a página estabilizar
            page.wait_for_timeout(3000)
            
            # Tenta rolar a página para baixo e para cima para forçar o carregamento
            page.mouse.move(Math.random() * 800, Math.random() * 800);  # Rola para baixo
            page.wait_for_timeout(1000)
            page.mouse.click(Math.random() * 1000, Math.random() * 1000);  # Rola para cima
            page.wait_for_timeout(1000)
            
            if not response:
                print(f"Erro: Não foi possível acessar a URL {url}", file=sys.stderr)
                return None
                
            if response.status != 200:
                print(f"Erro: Status code {response.status} ao acessar {url}", file=sys.stderr)
                return None

            print("Página carregada, aguardando elementos...", file=sys.stderr)
            
            # Aguarda a página estar completamente carregada
            page.wait_for_load_state("networkidle")
            
            # Tenta encontrar elementos específicos que indiquem que a página carregou
            for selector in ["text=RRT", "text=Número do RRT", "text=Dados do RRT", "main", "body"]:
                try:
                    element = page.wait_for_selector(selector, timeout=3000)
                    if element:
                        print(f"Elemento encontrado: {selector}", file=sys.stderr)
                        break
                except Exception as e:
                    print(f"Elemento não encontrado: {selector}", file=sys.stderr)
                    continue
            
            if not element:
                print("Erro: Elementos não encontrados na página", file=sys.stderr)
                return None

            print("Elementos encontrados, extraindo conteúdo...", file=sys.stderr)

            main_content = page.locator("main")
            if not main_content.is_visible():
                main_content = page.locator("body")

            content = main_content.inner_text()
            
            if not content or len(content.strip()) < 100:
                print("Erro: Conteúdo extraído está vazio ou muito pequeno", file=sys.stderr)
                return None
                
            print(content) # Print result to stdout for Go to capture
            return content

        except TimeoutError as e:
            # Print error to stderr for Go to capture
            print(f"Ocorreu um erro de Timeout ao processar RRT {numero_documento}: {e}", file=sys.stderr)

        finally:
            await browser.close()
            
        return None

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Uso: python automacao_caubr.py <numero_rrt>", file=sys.stderr)
        sys.exit(1)
    
    numero_do_documento = sys.argv[1]
    get_document_data(numero_do_documento)