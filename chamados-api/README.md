# API de chamados (Go + MongoDB) com login

Variáveis (veja `.env.example`). No PowerShell, na mesma janela do `go run .`:

    $env:MONGO_URI    = "sua-string-do-atlas"
    $env:JWT_SECRET   = "<segredo aleatório com 32+ caracteres>"
    $env:ADMIN_EMAIL  = "ti@suaempresa.com"
    $env:ADMIN_SENHA  = "uma-senha-forte"
    go mod tidy
    go run .

Gerar um JWT_SECRET:

    $b = New-Object byte[] 48; [Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($b); [Convert]::ToBase64String($b)

Na primeira subida, se ainda não existir ninguém da TI, a conta de `ADMIN_EMAIL` é criada com perfil `ti`.

## Rotas
| Método | Rota | Quem | O que faz |
|---|---|---|---|
| POST | /api/auth/registrar | público | cria conta de usuário comum `{nome, email, senha}` e devolve o token |
| POST | /api/auth/login | público | `{email, senha}` devolve `{token, usuario}` |
| GET | /api/auth/eu | logado | dados de quem está logado |
| GET | /api/chamados | logado | usuário vê só os seus; TI vê todos |
| POST | /api/chamados | logado | abre chamado |
| GET | /api/chamados/{id} | dono / TI | detalhe |
| POST | /api/chamados/{id}/mensagens | dono / TI | envia mensagem |
| PATCH | /api/chamados/{id} | só TI | altera status e responsável |
| GET | /api/equipe | só TI | nomes da equipe de TI |
| POST | /api/usuarios | só TI | cria conta `{nome, email, senha, perfil}` (perfil `usuario` ou `ti`) |

Rotas protegidas exigem o cabeçalho `Authorization: Bearer <token>`. O token vale 12 horas.

Chamados criados antes do login não têm dono (`solicitanteId`), então só a TI os enxerga.

## Configuração por arquivo .env
A API lê o arquivo `.env` na pasta onde roda (copie o `.env.example`). Variáveis já definidas no terminal ou no Render têm prioridade sobre o arquivo. O `.env` está no `.gitignore`: não o envie para o GitHub.
