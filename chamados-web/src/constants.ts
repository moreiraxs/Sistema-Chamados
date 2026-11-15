import type { Status } from './types'

export const STATUS: Record<Status, string> = { aberto: 'Aberto', andamento: 'Em andamento', fechado: 'Fechado' }
export const CATEGORIAS = ['Acesso e senhas', 'Hardware', 'Software', 'Rede e internet', 'Outros']

export const formatarData = (iso: string) =>
  new Date(iso).toLocaleString('pt-BR', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' })
