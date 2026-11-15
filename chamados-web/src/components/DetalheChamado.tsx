import { useState } from 'react'
import { STATUS, formatarData } from '../constants'
import type { Chamado, Mudancas, Status } from '../types'
import StatusBadge from './StatusBadge'

interface Props {
  chamado: Chamado
  nomeAtual: string
  ehTI: boolean
  equipe: string[]
  onEnviar: (id: number, texto: string) => Promise<void>
  onAtualizar: (id: number, mudancas: Mudancas) => Promise<void>
  onVoltar: () => void
}

export default function DetalheChamado({ chamado, nomeAtual, ehTI, equipe, onEnviar, onAtualizar, onVoltar }: Props) {
  const [texto, setTexto] = useState('')

  const enviar = async () => {
    if (!texto.trim()) return
    await onEnviar(chamado.id, texto.trim())
    setTexto('')
  }

  return (
    <article className="detalhe">
      <button className="btn voltar" onClick={onVoltar}>Voltar para a lista</button>
      <p className="id">Chamado #{chamado.id} · aberto em {formatarData(chamado.criadoEm)} por {chamado.solicitante}</p>
      <h2>{chamado.titulo}</h2>
      <p className="meta"><StatusBadge status={chamado.status} /><span className="id">{chamado.categoria}</span></p>

      <div className="props">
        <label>Status
          <select disabled={!ehTI} value={chamado.status} onChange={(e) => void onAtualizar(chamado.id, { status: e.target.value as Status })}>
            {(Object.keys(STATUS) as Status[]).map((s) => <option key={s} value={s}>{STATUS[s]}</option>)}
          </select>
        </label>
        <label>Responsável
          <select disabled={!ehTI} value={chamado.responsavel} onChange={(e) => void onAtualizar(chamado.id, { responsavel: e.target.value })}>
            <option value="">Sem responsável</option>
            {chamado.responsavel && !equipe.includes(chamado.responsavel) && <option>{chamado.responsavel}</option>}
            {equipe.map((n) => <option key={n}>{n}</option>)}
          </select>
        </label>
      </div>

      <p className="descricao">{chamado.descricao}</p>

      <h3>Conversa com a TI</h3>
      <div className="conversa">
        {chamado.mensagens.length === 0 && <p className="aviso">Nenhuma mensagem ainda.</p>}
        {chamado.mensagens.map((m, i) =>
          m.sistema ? (
            <p key={i} className="aviso">{m.texto} · {formatarData(m.em)}</p>
          ) : (
            <div key={i} className={`balao${m.autor === nomeAtual ? ' meu' : ''}`}>
              <div className="autor">{m.autor} · {formatarData(m.em)}</div>
              <div className="texto">{m.texto}</div>
            </div>
          )
        )}
      </div>

      <h3>Nova mensagem</h3>
      <textarea className="resposta" aria-label="Mensagem" value={texto} onChange={(e) => setTexto(e.target.value)}
        placeholder={ehTI ? 'Responda ao usuário' : 'Escreva sua mensagem para a TI'} />
      <button className="btn primario" onClick={() => void enviar()}>Enviar mensagem</button>
    </article>
  )
}
