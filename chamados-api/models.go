package main

import "time"

const (
	StatusAberto    = "aberto"
	StatusAndamento = "andamento"
	StatusFechado   = "fechado"
)

var rotuloStatus = map[string]string{
	StatusAberto:    "Aberto",
	StatusAndamento: "Em andamento",
	StatusFechado:   "Fechado",
}

func statusValido(s string) bool { _, ok := rotuloStatus[s]; return ok }

// Mensagem é uma mensagem da conversa ou um aviso do sistema (mudança de status/responsável).
type Mensagem struct {
	Autor   string    `bson:"autor,omitempty" json:"autor,omitempty"`
	Texto   string    `bson:"texto" json:"texto"`
	Sistema bool      `bson:"sistema,omitempty" json:"sistema,omitempty"`
	Em      time.Time `bson:"em" json:"em"`
}

// Chamado usa os mesmos nomes de campo JSON do front-end.
type Chamado struct {
	ID            int64      `bson:"_id" json:"id"`
	Titulo        string     `bson:"titulo" json:"titulo"`
	Descricao     string     `bson:"descricao" json:"descricao"`
	Categoria     string     `bson:"categoria" json:"categoria"`
	Status        string     `bson:"status" json:"status"`
	Solicitante   string     `bson:"solicitante" json:"solicitante"`
	SolicitanteID string     `bson:"solicitanteId" json:"solicitanteId"`
	Responsavel   string     `bson:"responsavel" json:"responsavel"`
	CriadoEm      time.Time  `bson:"criadoEm" json:"criadoEm"`
	AtualizadoEm  time.Time  `bson:"atualizadoEm" json:"atualizadoEm"`
	Mensagens     []Mensagem `bson:"mensagens" json:"mensagens"`
}
