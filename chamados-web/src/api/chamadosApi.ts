import { requisitar } from './client'
import type { Chamado, Mudancas, NovoChamadoDados } from '../types'

export const listar = () => requisitar<Chamado[]>('/api/chamados')
export const criar = (dados: NovoChamadoDados) => requisitar<Chamado>('/api/chamados', 'POST', dados)
export const enviarMensagem = (id: number, texto: string) =>
  requisitar<Chamado>(`/api/chamados/${id}/mensagens`, 'POST', { texto })
export const atualizar = (id: number, mudancas: Mudancas) =>
  requisitar<Chamado>(`/api/chamados/${id}`, 'PATCH', mudancas)
export const equipe = () => requisitar<string[]>('/api/equipe')
