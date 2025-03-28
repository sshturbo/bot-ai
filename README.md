# Orbi ai

Um assistente de chat inteligente com suporte a múltiplas plataformas (incluindo Telegram) que utiliza modelos de IA para interação natural com os usuários.

## Estrutura do Projeto

```
├── config/           # Configurações da aplicação
├── database/         # Camada de acesso a dados
├── frontend/         # Interface de usuário web
├── models/           # Definições de modelos de dados
└── services/         # Serviços de integração (Gemini, Telegram, HTTP)
```

## Recursos

- Integração com a API Gemini para processamento de linguagem natural
- Bot do Telegram para interação por mensagens
- Interface web para chat via navegador
- Armazenamento de histórico de mensagens

## Requisitos

- Go 1.21+
- Node.js 18+
- Acesso à API Gemini
- Token de bot do Telegram

## Instalação

### Backend (Go)

1. Clone o repositório
2. Crie um arquivo `.env` na raiz do projeto com as seguintes variáveis:
```env
API_KEY_GEMINI=sua_chave_da_api_gemini
TELEGRAM_BOT_TOKEN=seu_token_do_bot_telegram
PORT=8443  # Porta que o servidor irá rodar
```

3. Execute os comandos:
```bash
go mod tidy
go run main.go
```

### Frontend

1. Entre na pasta do frontend:
```bash
cd frontend
```

2. Configure o arquivo `public/config.json`:
```json
{
  "apiUrl": "http://localhost:8443"  # Deve corresponder à mesma porta configurada no backend
}
```

3. Instale as dependências e inicie o servidor de desenvolvimento:
```bash
npm install
npm run dev
```

## Configuração

### Variáveis de Ambiente (Backend)

| Variável | Descrição | Exemplo |
|----------|-----------|---------|
| `API_KEY_GEMINI` | Chave de API para o serviço Gemini | `xxxxxxxxxxxx` |
| `TELEGRAM_BOT_TOKEN` | Token do seu bot do Telegram | `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11` |
| `PORT` | Porta que o servidor irá rodar | `8443` |

### Arquivo de Configuração (Frontend)

O arquivo `frontend/public/config.json` contém as configurações do frontend:

```json
{
  "apiUrl": "http://localhost:8443"  # URL da API do backend
}
```

> **Importante**: Certifique-se de que a porta configurada no backend (variável `PORT`) corresponda à porta especificada no `apiUrl` do frontend.

## Licença

MIT