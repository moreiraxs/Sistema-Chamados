import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import * as authApi from '../api/authApi'
import { definirToken, registrarAoExpirar } from '../api/client'
import type { Usuario } from '../types'

const CHAVE = 'chamados:token'

interface AuthValor {
  usuario: Usuario | null
  carregando: boolean
  entrar: (email: string, senha: string) => Promise<void>
  cadastrar: (nome: string, email: string, senha: string) => Promise<void>
  sair: () => void
}

const Contexto = createContext<AuthValor | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [usuario, setUsuario] = useState<Usuario | null>(null)
  const [carregando, setCarregando] = useState(true)

  const sair = useCallback(() => {
    localStorage.removeItem(CHAVE)
    definirToken(null)
    setUsuario(null)
  }, [])

  // Ao abrir o app: se já havia login salvo, confirma com a API quem é.
  useEffect(() => {
    registrarAoExpirar(sair)
    const salvo = localStorage.getItem(CHAVE)
    if (!salvo) {
      setCarregando(false)
      return
    }
    definirToken(salvo)
    authApi.eu().then(setUsuario).catch(sair).finally(() => setCarregando(false))
  }, [sair])

  const iniciar = (r: authApi.RespostaAuth) => {
    localStorage.setItem(CHAVE, r.token)
    definirToken(r.token)
    setUsuario(r.usuario)
  }

  const valor = useMemo<AuthValor>(
    () => ({
      usuario,
      carregando,
      entrar: async (email, senha) => iniciar(await authApi.login(email, senha)),
      cadastrar: async (nome, email, senha) => iniciar(await authApi.registrar(nome, email, senha)),
      sair,
    }),
    [usuario, carregando, sair]
  )
  return <Contexto.Provider value={valor}>{children}</Contexto.Provider>
}

export function useAuth(): AuthValor {
  const v = useContext(Contexto)
  if (!v) throw new Error('useAuth precisa estar dentro de <AuthProvider>')
  return v
}
