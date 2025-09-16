# 🤖 Integração com API do Gemini

A funcionalidade de geração automática de texto de apresentação para leads agora utiliza **diretamente a API do Google Gemini**, removendo a dependência do serviço de mensageria intermediário.

## 📋 Configuração

### 1. Obter Chave da API Gemini

1. Acesse: https://aistudio.google.com/app/apikey
2. Faça login com sua conta Google
3. Clique em "Create API Key"
4. Copie a chave gerada

### 2. Configurar Variáveis de Ambiente

**Para Desenvolvimento Local:**
```bash
# Crie um arquivo .env na raiz do projeto
cp .env.example .env

# Edite o arquivo .env e adicione:
GEMINI_API_KEY=sua_chave_aqui
```

**Para Produção:**
```bash
# Adicione no arquivo .env do ambiente de produção:
GEMINI_API_KEY=sua_chave_aqui
```

### 3. Testar Configuração

```bash
# Exporte a variável de ambiente
export GEMINI_API_KEY=sua_chave_aqui

# Execute o script de teste
./test-gemini-api.sh
```

## 🚀 Como Funciona

### Fluxo de Funcionamento:

1. **Usuário edita perfil** → Modifica campos como "Nome da Loja", "Descrição", "Indústria", "Palavras-chave", ou "Data de Fundação"
2. **Clica em "Salvar"** → Sistema detecta mudanças nos campos-chave
3. **Perfil é salvo** → Dados são persistidos no PocketBase
4. **Chamada para Gemini** → API do Gemini é chamada diretamente com prompt otimizado
5. **Texto gerado** → Gemini retorna texto de apresentação personalizado
6. **Armazenamento** → Texto é salvo no campo `properties.lead_introduction_text`
7. **Exibição** → Novo texto aparece automaticamente no perfil do usuário

### Endpoint Utilizado:
```
POST https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key={API_KEY}
```

### Formato do Prompt:
```
Você deve criar um texto de apresentação profissional e informal para prospecção de leads.
Nome da empresa: [nome]
Descrição da empresa: [descrição]
Setor de atuação: [indústria]
Palavras-chave importantes: [keywords]
Data de fundação: [data]

Instruções:
1. Use um tom informal-profissional, pessoal e acolhedor
2. Destaque a empresa como parceira estratégica
3. Crie conexão com o leitor
4. Use entre 3 a 5 frases
5. Seja conciso e impactante
6. Ideal para primeiro contato com leads
7. Não use aspas ou formatação especial
8. Retorne apenas o texto de apresentação, sem explicações adicionais
```

## 📊 Parâmetros do Gemini

- **Modelo**: `gemini-1.5-flash-latest`
- **Temperature**: 0.7 (equilíbrio entre criatividade e consistência)
- **TopK**: 40
- **TopP**: 0.95
- **Max Output Tokens**: 200
- **Custo**: ~$0.00015 por requisição (muito baixo)

## ⚠️ Tratamento de Erros

### Cenários Cobertos:
- ✅ **API Key inválida ou ausente**
- ✅ **Problemas de conectividade**
- ✅ **Rate limiting do Gemini**
- ✅ **Resposta inválida ou vazia**
- ✅ **Falha ao salvar no banco de dados**
- ✅ **Campos obrigatórios ausentes**

### Comportamento em Caso de Erro:
- 🟢 **Perfil sempre é salvo**, mesmo se a geração de texto falhar
- 🟡 **Toast de aviso** informa o usuário sobre problemas
- 🔄 **Texto anterior é mantido** em caso de falha
- 📝 **Logs detalhados** para debugging

## 🔧 Arquivos Modificados

### Backend:
- `server/routes_messenger.go` - Refatorada função `MessengerGenerateLeadIntroduction`
- `suaobra-app.go` - Mantida rota `/messenger/generate-lead-introduction`

### Frontend:
- `frontend/src/store/store.ts` - Adicionado campo `lead_introduction_text`
- `frontend/src/components/NavBar.tsx` - Lógica de detecção de mudanças e exibição

### Configuração:
- `.env.example` - Documentação das variáveis necessárias
- `test-gemini-api.sh` - Script de teste da integração

## 💡 Vantagens da Integração Direta

1. **Simplicidade** - Remove dependência intermediária
2. **Performance** - Chamada direta, sem proxy
3. **Custo** - API do Gemini é muito barata
4. **Confiabilidade** - Menos pontos de falha
5. **Controle** - Total controle sobre prompts e parâmetros
6. **Logs** - Melhor rastreabilidade de problemas

## 🧪 Para Testar

1. Configure a `GEMINI_API_KEY`
2. Execute o script: `./test-gemini-api.sh`
3. Acesse a interface do suaobra-app
4. Edite o perfil da empresa
5. Modifique campos como nome ou descrição
6. Clique em "Salvar"
7. Verifique se o texto de apresentação aparece automaticamente

## 📞 Suporte

Se encontrar problemas:
1. Verifique se a `GEMINI_API_KEY` está correta
2. Execute o script de teste
3. Verifique os logs do backend
4. Confirme se o Google AI Studio está acessível