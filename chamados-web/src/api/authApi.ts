import { requisitar } from './client'
import type { Usuario } from '../types'

export interface RespostaAuth {
  token: string
  usuario: Usuario
}

export const login = (email: string, senha: string) =>
  requisitar<RespostaAuth>('/api/auth/login', 'POST', { email, senha })

export const registrar = (nome: string, email: string, senha: string) =>
  requisitar<RespostaAuth>('/api/auth/registrar', 'POST', { nome, email, senha })

export const eu = () => requisitar<Usuario>('/api/auth/eu')
