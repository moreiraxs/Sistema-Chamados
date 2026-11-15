import { useMemo, useState } from 'react'
import { STATUS } from '../constants'
import type { Chamado, Status } from '../types'
import StatusBadge from './StatusBadge'

interface Props {
  chamados: Chamado[]
  selecionadoId: number | null
  mostrarContagem: boolean // só a TI vê quantos chamados existem
  onSelecionar: (id: number) => void
}

const STATUSES = Object.keys(STATUS) as Status[]

export default function ListaChamados({ chamados, selecionadoId, mostrarContagem, onSelecionar }: Props) {
  const [filtro, setFiltro] = useState<Status | 'todos'>('todos')
  const [busca, setBusca] = useState('')

  const contagem = useMemo(() => {
    const c: Record<Status, number> = { aberto: 0, andamento: 0, fechado: 0 }
    chamados.forEach((x) => { c[x.status]++ })
    return c
  }, [chamados])

  const exibidos = useMemo(
    () =>
      chamados.filter(
        (c) =>
          (filtro === 'todos' || c.status === filtro) &&
          `${c.id} ${c.titulo} ${c.solicitante}`.toLowerCase().includes(busca.toLowerCase())
      ),
    [chamados, filtro, busca]
  )

  return (
    <section className="lista">
      <div className="chips" role="group" aria-label="Filtrar por status">
        <button className="chip" aria-pressed={filtro === 'todos'} onClick={() => setFiltro('todos')}>
          Todos {mostrarContagem && <b>{chamados.length}</b>}
        </button>
        {STATUSES.map((s) => (
          <button key={s} className="chip" aria-pressed={filtro === s} onClick={() => setFiltro(s)}>
            {STATUS[s]} {mostrarContagem && <b>{contagem[s]}</b>}
          </button>
        ))}
      </div>
      <div className="busca">
        <input aria-label="Buscar" placeholder="Buscar por título ou número" value={busca} onChange={(e) => setBusca(e.target.value)} />
      </div>
      {exibidos.length === 0 && <p className="vazio">Nenhum chamado encontrado.</p>}
      {exibidos.map((c) => (
        <button key={c.id} className="linha" aria-current={selecionadoId === c.id} onClick={() => onSelecionar(c.id)}>
          <span className="id">#{c.id}</span>
          <span className="titulo">{c.titulo}</span>
          <span className="meta">
            <StatusBadge status={c.status} />
            <span>{c.responsavel || 'Sem responsável'}</span>
          </span>
        </button>
      ))}
    </section>
  )
}
