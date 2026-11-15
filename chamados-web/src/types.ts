export type Status = 'aberto' | 'andamento' | 'fechado'
export type Perfil = 'usuario' | 'ti'

export interface Usuario {
  id: string
  nome: string
  email: string
  perfil: Perfil
}

export interface Mensagem {
  autor?: string
  texto: string
  sistema?: boolean
  em: string
}

export interface Chamado {
  id: number
  titulo: string
  descricao: string
  categoria: string
  status: Status
  solicitante: string
  solicitanteId: string
  responsavel: string
  criadoEm: string
  atualizadoEm: string
  mensagens: Mensagem[]
}

export interface NovoChamadoDados {
  titulo: string
  descricao: string
  categoria: string
}

export interface Mudancas {
  status?: Status
  responsavel?: string
}
