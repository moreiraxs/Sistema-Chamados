package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type API struct{ store *Store }

func escrever(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func erro(w http.ResponseWriter, status int, msg string) {
	escrever(w, status, map[string]string{"erro": msg})
}

func lerJSON(w http.ResponseWriter, r *http.Request, destino any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(destino); err != nil {
		erro(w, http.StatusBadRequest, "JSON inválido")
		return false
	}
	return true
}

func idDaRota(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		erro(w, http.StatusBadRequest, "id inválido")
		return 0, false
	}
	return id, true
}

func (a *API) falhou(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNaoEncontrado) {
		erro(w, http.StatusNotFound, "chamado não encontrado")
		return
	}
	log.Printf("erro: %v", err)
	erro(w, http.StatusInternalServerError, "erro interno")
}

// GET /api/chamados?status=aberto&busca=email
func (a *API) listar(w http.ResponseWriter, r *http.Request) {
	ident := identidadeDe(r)
	status := r.URL.Query().Get("status")
	if status != "" && !statusValido(status) {
		erro(w, http.StatusBadRequest, "status inválido")
		return
	}
	lista, err := a.store.Listar(r.Context(), Filtro{
		Solicitante: restricao(ident),
		Status:      status,
		Busca:       strings.TrimSpace(r.URL.Query().Get("busca")),
	})
	if err != nil {
		a.falhou(w, err)
		return
	}
	escrever(w, http.StatusOK, lista)
}

// POST /api/chamados
func (a *API) criar(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Titulo    string `json:"titulo"`
		Descricao string `json:"descricao"`
		Categoria string `json:"categoria"`
	}
	if !lerJSON(w, r, &in) {
		return
	}
	titulo, descricao, categoria := strings.TrimSpace(in.Titulo), strings.TrimSpace(in.Descricao), strings.TrimSpace(in.Categoria)
	if titulo == "" || descricao == "" {
		erro(w, http.StatusBadRequest, "título e descrição são obrigatórios")
		return
	}
	if utf8.RuneCountInString(titulo) > 200 || utf8.RuneCountInString(descricao) > 5000 {
		erro(w, http.StatusBadRequest, "título (máx. 200) ou descrição (máx. 5000) muito longos")
		return
	}
	if categoria == "" {
		categoria = "Outros"
	}
	agora := time.Now().UTC()
	c := Chamado{
		Titulo: titulo, Descricao: descricao, Categoria: categoria,
		Status: StatusAberto, Solicitante: identidadeDe(r).Nome, SolicitanteID: identidadeDe(r).ID,
		CriadoEm: agora, AtualizadoEm: agora, Mensagens: []Mensagem{},
	}
	if err := a.store.Criar(r.Context(), &c); err != nil {
		a.falhou(w, err)
		return
	}
	escrever(w, http.StatusCreated, c)
}

// GET /api/chamados/{id}
func (a *API) obter(w http.ResponseWriter, r *http.Request) {
	id, ok := idDaRota(w, r)
	if !ok {
		return
	}
	c, err := a.store.Buscar(r.Context(), id, restricao(identidadeDe(r)))
	if err != nil {
		a.falhou(w, err)
		return
	}
	escrever(w, http.StatusOK, c)
}

// POST /api/chamados/{id}/mensagens
func (a *API) enviarMensagem(w http.ResponseWriter, r *http.Request) {
	id, ok := idDaRota(w, r)
	if !ok {
		return
	}
	var in struct {
		Texto string `json:"texto"`
	}
	if !lerJSON(w, r, &in) {
		return
	}
	texto := strings.TrimSpace(in.Texto)
	if texto == "" || utf8.RuneCountInString(texto) > 5000 {
		erro(w, http.StatusBadRequest, "a mensagem deve ter de 1 a 5000 caracteres")
		return
	}
	ident := identidadeDe(r)
	msg := Mensagem{Autor: ident.Nome, Texto: texto, Em: time.Now().UTC()}
	if err := a.store.AdicionarMensagem(r.Context(), id, restricao(ident), msg); err != nil {
		a.falhou(w, err)
		return
	}
	c, err := a.store.Buscar(r.Context(), id, restricao(ident))
	if err != nil {
		a.falhou(w, err)
		return
	}
	escrever(w, http.StatusCreated, c)
}

// PATCH /api/chamados/{id}  { "status": "fechado", "responsavel": "Rafael Lima" }  (somente TI)
func (a *API) atualizar(w http.ResponseWriter, r *http.Request) {
	ident := identidadeDe(r)
	if !ident.EhTI {
		erro(w, http.StatusForbidden, "somente a TI altera status e responsável")
		return
	}
	id, ok := idDaRota(w, r)
	if !ok {
		return
	}
	var in struct {
		Status      *string `json:"status"`
		Responsavel *string `json:"responsavel"`
	}
	if !lerJSON(w, r, &in) {
		return
	}
	if in.Status == nil && in.Responsavel == nil {
		erro(w, http.StatusBadRequest, "informe status e/ou responsavel")
		return
	}
	if in.Status != nil && !statusValido(*in.Status) {
		erro(w, http.StatusBadRequest, "status inválido (use aberto, andamento ou fechado)")
		return
	}
	if in.Responsavel != nil {
		v := strings.TrimSpace(*in.Responsavel)
		in.Responsavel = &v
	}
	if err := a.store.Atualizar(r.Context(), id, in.Status, in.Responsavel, ident.Nome); err != nil {
		a.falhou(w, err)
		return
	}
	c, err := a.store.Buscar(r.Context(), id, "")
	if err != nil {
		a.falhou(w, err)
		return
	}
	escrever(w, http.StatusOK, c)
}
