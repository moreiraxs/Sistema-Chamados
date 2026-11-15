import { useState, type FormEvent } from 'react'
import { useAuth } from '../auth/AuthContext'

export default function Login() {
  const { entrar, cadastrar } = useAuth()
  const [criandoConta, setCriandoConta] = useState(false)
  const [nome, setNome] = useState('')
  const [email, setEmail] = useState('')
  const [senha, setSenha] = useState('')
  const [erro, setErro] = useState('')
  const [enviando, setEnviando] = useState(false)

  const enviar = async (e: FormEvent) => {
    e.preventDefault()
    setErro('')
    setEnviando(true)
    try {
      if (criandoConta) await cadastrar(nome, email, senha)
      else await entrar(email, senha)
    } catch (err) {
      setErro(err instanceof Error ? err.message : 'Algo deu errado')
      setEnviando(false)
    }
  }

  return (
    <form className="login" onSubmit={enviar}>
      <h1>{criandoConta ? 'Criar conta' : 'Entrar'}</h1>
      <p className="sub">Sistema de chamados</p>
      {criandoConta && (
        <label>Nome<input value={nome} onChange={(e) => setNome(e.target.value)} autoComplete="name" required /></label>
      )}
      <label>E-mail<input type="email" value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="email" required /></label>
      <label>Senha
        <input type="password" value={senha} onChange={(e) => setSenha(e.target.value)} minLength={criandoConta ? 8 : undefined}
          autoComplete={criandoConta ? 'new-password' : 'current-password'} required />
      </label>
      {criandoConta && <p className="sub">Use pelo menos 8 caracteres.</p>}
      {erro && <p className="erro" role="alert">{erro}</p>}
      <button type="submit" className="btn primario" disabled={enviando}>
        {enviando ? 'Aguarde…' : criandoConta ? 'Criar conta' : 'Entrar'}
      </button>
      <button type="button" className="troca" onClick={() => { setCriandoConta(!criandoConta); setErro('') }}>
        {criandoConta ? 'Já tenho conta' : 'Não tenho conta, quero criar'}
      </button>
    </form>
  )
}
