# Sistema de Chamados

Sistema de chamados de TI com login, perfis de acesso e conversa entre usuário e equipe de TI. Criado para ajudar no dia a dia de uma empresa.

## Funcionalidades

- Cadastro e login com autenticação JWT
- Dois perfis: **Usuário** e **TI**
- Usuário abre chamados e acompanha apenas os próprios
- TI vê todos os chamados, altera status e atribui responsável
- Conversa dentro de cada chamado, com histórico de mudanças de status e responsável
- Filtro por status e busca por título ou número

## Tecnologias

| Camada | Tecnologias |
|---|---|
| API | Go, MongoDB, JWT, bcrypt |
| Front-end | React, TypeScript, Vite |

## Estrutura

```
Sistema-Chamados/
├── chamados-api/   # API REST em Go
└── chamados-web/   # Interface em React + TypeScript
```

## Como rodar

### Pré-requisitos

- [Go](https://go.dev/dl/) 1.22 ou superior
- [Node.js](https://nodejs.org/) 18 ou superior
- Um MongoDB: conta gratuita no [MongoDB Atlas](https://www.mongodb.com/atlas) **ou** MongoDB local (veja a opção com Docker abaixo)

### 1. API

```bash
cd chamados-api
cp .env.example .env     # no PowerShell: Copy-Item .env.example .env
```

Edite o `.env` com os seus valores (veja a tabela de variáveis abaixo).

Para gerar um `JWT_SECRET` seguro no PowerShell:

```powershell
$b = New-Object byte[] 48; [Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($b); [Convert]::ToBase64String($b)
```

Depois, inicie a API:

```bash
go mod tidy
go run .
```

A API sobe em `http://localhost:8080`. Na primeira execução, se ainda não existir ninguém da TI, a conta definida em `ADMIN_EMAIL` e `ADMIN_SENHA` é criada com perfil **TI**.

### 2. Front-end

Em outro terminal:

```bash
cd chamados-web
npm install
npm run dev
```

Acesse `http://localhost:5173`.

### MongoDB local com Docker (opcional)

Se preferir não usar o Atlas, o `docker-compose.yml` da API sobe apenas o MongoDB:

```bash
cd chamados-api
docker compose up -d
```

Neste caso, use `MONGO_URI=mongodb://localhost:27017` no `.env`.

## Variáveis de ambiente

### API (`chamados-api/.env`)

| Variável | Descrição | Padrão |
|---|---|---|
| `MONGO_URI` | String de conexão do MongoDB | `mongodb://localhost:27017` |
| `DB_NAME` | Nome do banco | `chamados` |
| `JWT_SECRET` | Segredo para assinar os tokens (**mínimo 32 caracteres**) | obrigatória |
| `ADMIN_EMAIL` | E-mail da conta de TI criada na primeira execução | opcional |
| `ADMIN_SENHA` | Senha dessa conta | opcional |
| `CORS_ORIGIN` | Origem do front-end autorizada | `http://localhost:5173` |
| `PORT` | Porta da API | `8080` |
| `TRUST_PROXY` | `true` somente se a API estiver atrás de um proxy (usa `X-Forwarded-For` no limite de tentativas) | `false` |

O arquivo `.env` está no `.gitignore` e **nunca deve ser enviado ao GitHub**. O `.env.example` contém apenas valores de exemplo.

### Front-end (`chamados-web/.env`)

| Variável | Descrição | Padrão |
|---|---|---|
| `VITE_API_URL` | URL da API | `http://localhost:8080` |

## Rotas da API

| Método | Rota | Acesso | Descrição |
|---|---|---|---|
| POST | `/api/auth/registrar` | público | Cria conta de usuário comum |
| POST | `/api/auth/login` | público | Devolve token e dados do usuário |
| GET | `/api/auth/eu` | logado | Dados de quem está logado |
| GET | `/api/chamados` | logado | Usuário vê os seus; TI vê todos |
| POST | `/api/chamados` | logado | Abre um chamado |
| GET | `/api/chamados/{id}` | dono ou TI | Detalhe do chamado |
| POST | `/api/chamados/{id}/mensagens` | dono ou TI | Envia mensagem |
| PATCH | `/api/chamados/{id}` | TI | Altera status e responsável |
| GET | `/api/equipe` | TI | Lista a equipe de TI |
| POST | `/api/usuarios` | TI | Cria usuário ou outro membro da TI |

As rotas protegidas exigem o cabeçalho `Authorization: Bearer <token>`. O token vale 12 horas.

## Segurança

- Senhas armazenadas com **bcrypt**
- **JWT** com algoritmo fixado (HS256), expiração e segredo mínimo de 32 caracteres
- **Autorização por dono do chamado**: usuário comum só acessa os próprios chamados; ações de TI são verificadas no servidor
- Cadastro público sempre cria perfil comum (não é possível se autopromover a TI)
- **Limite de tentativas por IP** nas rotas de login e cadastro, contra força bruta
- Resposta de login sem revelar se o e-mail existe, inclusive no tempo de resposta
- Consultas ao banco sem concatenação de dados do usuário (proteção contra injeção NoSQL)
- Validação e limite de tamanho nos campos e no corpo das requisições
- Timeouts configurados no servidor HTTP
- Segredos fora do código, via variáveis de ambiente

## Próximos passos

- Paginação na listagem de chamados
- Testes automatizados
- Deploy da API e do front-end
