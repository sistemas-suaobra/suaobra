const express = require('express');
const puppeteer = require('puppeteer-core');
const crypto = require('crypto');

// ZenRows connection configuration
const connectionURL = 'wss://browser.zenrows.com?apikey=b67313ec4485fd294ff26be3f995989e4b7ab61b&proxy_country=br';

// Function to scrape RRT data
async function scrapeRrt(rrtNumber) {
    let browser;
    try {
        console.log('Connecting to external browser...');
        
        // Connect to ZenRows browser
        browser = await puppeteer.connect({ browserWSEndpoint: connectionURL });
        
        console.log('Browser connected successfully');
        
        const page = await browser.newPage();
        console.log('New page created');
        
        // Configure page to be more realistic
        await page.setViewport({ width: 1366, height: 768 });
        
        // Set realistic user agent
        await page.setUserAgent('Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36');
        
        // Set additional headers to appear more like a real browser
        await page.setExtraHTTPHeaders({
            'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7',
            'Accept-Language': 'pt-BR,pt;q=0.9,en;q=0.8',
            'Accept-Encoding': 'gzip, deflate, br',
            'Sec-Fetch-Dest': 'document',
            'Sec-Fetch-Mode': 'navigate',
            'Sec-Fetch-Site': 'none',
            'Upgrade-Insecure-Requests': '1',
        });
        
        // Remove webdriver property
        await page.evaluateOnNewDocument(() => {
            Object.defineProperty(navigator, 'webdriver', {
                get: () => undefined,
            });
        });
        
        const url = `https://acesso.caubr.gov.br/autenticidade/rrt?numero=${rrtNumber}`;
        console.log(`Navigating to: ${url}`);
        
        // Navigate with realistic timing
        await page.goto(url, { 
            waitUntil: 'networkidle2',
            timeout: 30000 
        });
        
        console.log('Page loaded, waiting for content...');
        
        // Wait for page to fully load and simulate human behavior
        await page.waitForSelector('.body-title', { timeout: 4000 });

        // Check for Cloudflare challenge
        const cloudflareChallenge = await page.$('.cf-browser-verification, .cf-challenge-running, .cf-checking, #challenge-form');
        if (cloudflareChallenge) {
            console.log('Cloudflare challenge detected, waiting...');
            await page.waitForTimeout(8000); // Wait for challenge to complete
        }
        
        // Simulate some mouse movement (human-like behavior)
        await page.mouse.move(100, 100);
        await page.waitForTimeout(500);
        await page.mouse.move(200, 200);
        
        // Wait a bit more for any dynamic content
        await page.waitForTimeout(3000);
        
        // Extract text content (not HTML) like Python script
        const content = await page.evaluate(() => {
            // Try to get main content first, fallback to body
            const main = document.querySelector('main');
            if (main && main.innerText.trim().length > 100) {
                return main.innerText;
            }
            return document.body.innerText;
        });
        
        console.log('Content retrieved');
        
        return content;
        
    } catch (error) {
        console.error('Error in scrapeRrt:', error);
        throw error;
    } finally {
        if (browser) {
            await browser.close();
            console.log('Browser closed');
        }
    }
}

// Function to extract data from HTML content (following Python scraper_rrt.py pattern)
function extractData(text) {
    try {
        // Helper function to extract data using regex patterns
        function extract(pattern, content) {
            const match = content.match(new RegExp(pattern, 'i'));
            if (match && match.length > 1) {
                return match[1].trim();
            }
            return "";
        }

        // Helper function to format dates from DD/MM/YYYY to YYYY-MM-DD
        function formatDate(dateStr) {
            if (!dateStr) return "";
            try {
                const parts = dateStr.split('/');
                if (parts.length === 3) {
                    const day = parts[0].padStart(2, '0');
                    const month = parts[1].padStart(2, '0');
                    const year = parts[2];
                    return `${year}-${month}-${day}`;
                }
            } catch (error) {
                // Return empty string if date parsing fails
            }
            return "";
        }

        const data = {};

        // Extract basic data following Python pattern
        data.obra_number = extract('Número do RRT:\\s*(\\d+)', text);
        if (!data.obra_number) {
            throw new Error("Não foi possível extrair o número da obra.");
        }

        // Generate MD5 hash for ID (same as Python)
        const crypto = require('crypto');
        const md5Hash = crypto.createHash('md5').update(data.obra_number).digest('hex');
        data.id = "obra_" + md5Hash;

        // Extract professional info
        data.professional = extract('Arquiteto\\(a\\) e Urbanista:\\s*(.*)', text);

        // Extract owner with same logic as Python
        const ownerValue = extract('Nome/Razão Social:\\s*(.*)', text);
        if (ownerValue.startsWith("CPF / CNPJ")) {
            data.owner = "";
        } else {
            data.owner = ownerValue;
        }

        // Extract location data
        data.bairro = extract('Bairro:\\s*(.*)', text);
        
        // Extract and format dates
        data.start_date = formatDate(extract('Data de Início:\\s*(.*)', text));
        data.end_date = formatDate(extract('Previsão de Término:\\s*(.*)', text));
        data.first_listing_date = formatDate(extract('Data de Registro:\\s*(.*)', text));

        // Extract city and state
        const cityState = extract('Cidade/UF:\\s*(.*)', text);
        if (cityState) {
            const parts = cityState.split('/');
            if (parts.length === 2) {
                data.city = parts[0].trim();
                data.state = parts[1].trim();
            } else {
                data.city = "";
                data.state = "";
            }
        } else {
            data.city = "";
            data.state = "";
        }

        // Extract activity section (between 'Atividade Subordinada' and 'Pagamento')
        const activitySectionMatch = text.match(/Atividade Subordinada([\s\S]*?)Pagamento/i);
        const activitySection = activitySectionMatch ? activitySectionMatch[1] : "";

        // Extract activity
        const activityMatch = activitySection.match(/(\d+\.\d+\.\d+\s*-\s*[^\d\s-][\s\S]*?)[\s]*[\d]/);
        data.activity = activityMatch ? activityMatch[1].trim() : "";

        // Extract type with complex regex pattern
        const typeMatch = activitySection.match(/([\d\.]+)\s*([\wÇÃÁÉÍÓÚ]+)\s*>\s*(\d+)\s*>\s*([\d\.]+\s*-\s*[\w\sÇÃÁÉÍÓÚ]+)/i);
        if (typeMatch && typeMatch.length > 4) {
            data.type = `${typeMatch[3]} - ${typeMatch[2]}`;
        } else {
            data.type = "";
        }

        // Extract size and unit
        const sizeUnitMatch = activitySection.match(/(\d+\.?\d*)\s*\/\s*([\w\s²]+)/i);
        if (sizeUnitMatch && sizeUnitMatch.length > 2) {
            try {
                const sizeValue = sizeUnitMatch[1].split('.')[0];
                data.size = parseFloat(sizeValue);
                data.unidade = sizeUnitMatch[2].trim();
            } catch (error) {
                data.size = 0.0;
                data.unidade = "";
            }
        } else {
            data.size = 0.0;
            data.unidade = "";
        }

        // Extract address components
        const tipoLogradouro = extract('Tipo de Logradouro:\\s*(.*)', text);
        const logradouro = extract('Logradouro:\\s*(.*)', text);
        const numero = extract('Número/Ano:\\s*(.*)', text);
        const complemento = extract('Complemento:\\s*(.*)', text);

        // Build complete address (same logic as Python)
        const addressParts = [tipoLogradouro, logradouro, numero, complemento, data.bairro]
            .filter(part => part && part.trim());
        
        if (addressParts.length > 0) {
            data.address = `${addressParts.join(', ')}, ${data.city} - ${data.state}`;
        } else {
            data.address = `${data.city} - ${data.state}`;
        }

        return data;

    } catch (error) {
        console.error('Error extracting data:', error);
        return {
            error: 'Failed to extract data',
            message: error.message,
            raw_html: text.substring(0, 1000) // First 1000 chars for debugging
        };
    }
}

const app = express();
const port = 3001;

// Health check endpoint
app.get('/health', (req, res) => {
    res.json({ status: 'ok', message: 'Server is running' });
});

// Test endpoint to verify Puppeteer is working
app.get('/test', async (req, res) => {
    let browser;
    try {
        console.log('Starting external browser test...');
        
        // Connect to ZenRows browser
        browser = await puppeteer.connect({ browserWSEndpoint: connectionURL });
        
        console.log('Browser connected successfully');
        
        const page = await browser.newPage();
        console.log('New page created');
        
        await page.goto('https://example.com', { waitUntil: 'networkidle2' });
        console.log('Page loaded');
        
        const title = await page.title();
        console.log('Page title retrieved:', title);
        
        res.json({ 
            status: 'ok', 
            message: 'External browser test successful',
            title: title,
            browser_type: 'ZenRows External'
        });
        
    } catch (error) {
        console.error('External browser test failed:', error);
        res.status(500).json({ error: error.message });
    } finally {
        if (browser) {
            await browser.close();
            console.log('Browser closed');
        }
    }
});

// Main scraping endpoint
app.get('/scrape/:rrt_number', async (req, res) => {
    const { rrt_number } = req.params;
    
    try {
        console.log(`Scraping RRT: ${rrt_number}`);
        const rawData = await scrapeRrt(rrt_number);
        const data = extractData(rawData);
        res.json(data);
    } catch (e) {
        res.status(500).json({ error: `Failed to scrape data: ${e.message}` });
    }
});

app.listen(port, () => {
    console.log(`API running at http://localhost:${port}`);
});
