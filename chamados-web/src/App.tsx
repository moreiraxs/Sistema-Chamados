import { useEffect, useMemo, useState } from 'react'
import * as api from './api/chamadosApi'
import { AuthProvider, useAuth } from './auth/AuthContext'
import DetalheChamado from './components/DetalheChamado'
import ListaChamados from './components/ListaChamados'
import Login from './components/Login'
import NovoChamado from './components/NovoChamado'
import type { Chamado, Mudancas, NovoChamadoDados, Usuario } from './types'

function Painel({ usuario }: { usuario: Usuario }) {
  const { sair } = useAuth()
  const ehTI = usuario.perfil === 'ti'
  const [chamados, setChamados] = useState<Chamado[]>([])
  const [equipe, setEquipe] = useState<string[]>([])
  const [selecionadoId, setSelecionadoId] = useState<number | null>(null)
  const [criando, setCriando] = useState(false)
  const [erro, setErro] = useState('')

  const executar = async (acao: () => Promise<unknown>) => {
    try {
      await acao()
      setErro('')
    } catch (e) {
      setErro(e instanceof Error ? e.message : 'Algo deu errado')
    }
  }
  const recarregar = () => executar(async () => setChamados(await api.listar()))

  useEffect(() => {
    void recarregar()
    if (ehTI) void executar(async () => setEquipe(await api.equipe()))
  }, [])

  const atual = useMemo(() => chamados.find((c) => c.id === selecionadoId), [chamados, selecionadoId])
  const abertos = chamados.filter((c) => c.status === 'aberto').length

  const criar = (dados: NovoChamadoDados) =>
    executar(async () => {
      const novo = await api.criar(dados)
      await recarregar()
      setCriando(false)
      setSelecionadoId(novo.id)
    })
  const enviar = (id: number, texto: string) =>
    executar(async () => { await api.enviarMensagem(id, texto); await recarregar() })
  const atualizar = (id: number, mudancas: Mudancas) =>
    executar(async () => { await api.atualizar(id, mudancas); await recarregar() })

  return (
    <>
      <header>
        <div>
          <h1>{ehTI ? 'Chamados' : 'Meus chamados'}</h1>
          {/* Só a TI vê o total de chamados do sistema */}
          {ehTI && <p className="sub">{chamados.length} no sistema · {abertos} abertos</p>}
        </div>
        <div className="topo-acoes">
          <span className="quem">
            {usuario.nome}
            <span className={`etiqueta ${ehTI ? 'ti' : 'usuario'}`}>{ehTI ? 'TI' : 'Usuário'}</span>
          </span>
          {!ehTI && <button className="btn primario" onClick={() => setCriando(true)}>Novo chamado</button>}
          <button className="btn" onClick={sair}>Sair</button>
        </div>
      </header>

      {erro && <p className="erro" role="alert" style={{ margin: 0, padding: '10px 20px' }}>{erro}</p>}

      <main className={`layout${atual ? ' com-selecao' : ''}`}>
        <ListaChamados chamados={chamados} selecionadoId={selecionadoId} mostrarContagem={ehTI} onSelecionar={setSelecionadoId} />
        {atual ? (
          <DetalheChamado key={atual.id} chamado={atual} nomeAtual={usuario.nome} ehTI={ehTI} equipe={equipe}
            onEnviar={enviar} onAtualizar={atualizar} onVoltar={() => setSelecionadoId(null)} />
        ) : (
          <p className="vazio detalhe-vazio">Selecione um chamado na lista para ver os detalhes.</p>
        )}
      </main>

      {criando && <NovoChamado onSalvar={criar} onFechar={() => setCriando(false)} />}
    </>
  )
}

function Raiz() {
  const { usuario, carregando } = useAuth()
  if (carregando) return <p className="vazio">Carregando…</p>
  return usuario ? <Painel usuario={usuario} /> : <Login />
}

export default function App() {
  return (
    <AuthProvider>
      <Raiz />
    </AuthProvider>
  )
}
