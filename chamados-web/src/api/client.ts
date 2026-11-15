const BASE: string = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

let token: string | null = null
let aoExpirar: (() => void) | null = null

export const definirToken = (t: string | null) => { token = t }
export const registrarAoExpirar = (fn: () => void) => { aoExpirar = fn }

export async function requisitar<T>(caminho: string, metodo = 'GET', corpo?: unknown): Promise<T> {
  const headers: Record<string, string> = {}
  if (token) headers.Authorization = `Bearer ${token}`
  if (corpo !== undefined) headers['Content-Type'] = 'application/json'

  let resposta: Response
  try {
    resposta = await fetch(`${BASE}${caminho}`, {
      method: metodo,
      headers,
      body: corpo === undefined ? undefined : JSON.stringify(corpo),
    })
  } catch {
    throw new Error('Não consegui falar com a API. Ela está rodando?')
  }
  if (!resposta.ok) {
    const dados = await resposta.json().catch(() => ({}))
    if (resposta.status === 401 && token) aoExpirar?.() // sessão expirada: volta para o login
    throw new Error(dados.erro ?? `Erro ${resposta.status}`)
  }
  return resposta.json() as Promise<T>
}
