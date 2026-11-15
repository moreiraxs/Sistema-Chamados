import { useState, type ChangeEvent, type FormEvent } from 'react'
import { CATEGORIAS } from '../constants'
import type { NovoChamadoDados } from '../types'

interface Props {
  onSalvar: (dados: NovoChamadoDados) => Promise<void>
  onFechar: () => void
}

export default function NovoChamado({ onSalvar, onFechar }: Props) {
  const [form, setForm] = useState<NovoChamadoDados>({ titulo: '', descricao: '', categoria: CATEGORIAS[0] })
  const [erro, setErro] = useState('')
  const campo = (nome: keyof NovoChamadoDados) =>
    (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) =>
      setForm({ ...form, [nome]: e.target.value })

  const enviar = (e: FormEvent) => {
    e.preventDefault()
    if (!form.titulo.trim() || !form.descricao.trim()) {
      setErro('Preencha o título e a descrição para abrir o chamado.')
      return
    }
    void onSalvar({ ...form, titulo: form.titulo.trim(), descricao: form.descricao.trim() })
  }

  return (
    <div className="overlay" onClick={(e) => e.target === e.currentTarget && onFechar()}>
      <form className="modal" onSubmit={enviar}>
        <h2>Novo chamado</h2>
        <label>Título<input value={form.titulo} onChange={campo('titulo')} placeholder="Resuma o problema em uma frase" autoFocus /></label>
        <label>Descrição<textarea value={form.descricao} onChange={campo('descricao')} placeholder="O que aconteceu? Desde quando? Já tentou algo?" /></label>
        <label>Categoria
          <select value={form.categoria} onChange={campo('categoria')}>
            {CATEGORIAS.map((c) => <option key={c}>{c}</option>)}
          </select>
        </label>
        {erro && <p className="erro" role="alert">{erro}</p>}
        <div className="acoes">
          <button type="button" className="btn" onClick={onFechar}>Cancelar</button>
          <button type="submit" className="btn primario">Abrir chamado</button>
        </div>
      </form>
    </div>
  )
}
