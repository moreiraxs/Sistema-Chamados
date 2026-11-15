package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrNaoEncontrado = errors.New("chamado não encontrado")

type Store struct {
	chamados   *mongo.Collection
	contadores *mongo.Collection
	usuarios   *mongo.Collection
}

type Filtro struct {
	Solicitante string // vazio = todos (visão da TI)
	Status      string
	Busca       string
}

func NewStore(ctx context.Context, db *mongo.Database) (*Store, error) {
	s := &Store{chamados: db.Collection("chamados"), contadores: db.Collection("contadores"), usuarios: db.Collection("usuarios")}
	_, err := s.chamados.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "solicitanteId", Value: 1}, {Key: "criadoEm", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "criadoEm", Value: -1}}},
	})
	if err != nil {
		return nil, err
	}
	_, err = s.usuarios.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return s, err
}

// proximoID gera o número sequencial do chamado (1, 2, 3...).
func (s *Store) proximoID(ctx context.Context) (int64, error) {
	var out struct {
		Seq int64 `bson:"seq"`
	}
	err := s.contadores.FindOneAndUpdate(ctx,
		bson.M{"_id": "chamados"},
		bson.M{"$inc": bson.M{"seq": 1}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&out)
	return out.Seq, err
}

func (s *Store) Criar(ctx context.Context, c *Chamado) error {
	id, err := s.proximoID(ctx)
	if err != nil {
		return err
	}
	c.ID = id
	_, err = s.chamados.InsertOne(ctx, c)
	return err
}

func (s *Store) Listar(ctx context.Context, f Filtro) ([]Chamado, error) {
	q := bson.M{}
	if f.Solicitante != "" {
		q["solicitanteId"] = f.Solicitante
	}
	if f.Status != "" {
		q["status"] = f.Status
	}
	if f.Busca != "" {
		rx := bson.M{"$regex": regexp.QuoteMeta(f.Busca), "$options": "i"}
		ou := []bson.M{{"titulo": rx}, {"solicitante": rx}}
		if n, err := strconv.ParseInt(f.Busca, 10, 64); err == nil {
			ou = append(ou, bson.M{"_id": n})
		}
		q["$or"] = ou
	}
	cur, err := s.chamados.Find(ctx, q, options.Find().SetSort(bson.D{{Key: "criadoEm", Value: -1}}))
	if err != nil {
		return nil, err
	}
	lista := []Chamado{}
	if err := cur.All(ctx, &lista); err != nil {
		return nil, err
	}
	for i := range lista {
		if lista[i].Mensagens == nil {
			lista[i].Mensagens = []Mensagem{}
		}
	}
	return lista, nil
}

// Buscar retorna o chamado; se solicitante != "", só se ele pertencer a essa pessoa.
func (s *Store) Buscar(ctx context.Context, id int64, solicitante string) (*Chamado, error) {
	filtro := bson.M{"_id": id}
	if solicitante != "" {
		filtro["solicitanteId"] = solicitante
	}
	var c Chamado
	if err := s.chamados.FindOne(ctx, filtro).Decode(&c); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	if c.Mensagens == nil {
		c.Mensagens = []Mensagem{}
	}
	return &c, nil
}

func (s *Store) AdicionarMensagem(ctx context.Context, id int64, solicitante string, m Mensagem) error {
	filtro := bson.M{"_id": id}
	if solicitante != "" {
		filtro["solicitanteId"] = solicitante
	}
	res, err := s.chamados.UpdateOne(ctx, filtro, bson.M{
		"$push": bson.M{"mensagens": m},
		"$set":  bson.M{"atualizadoEm": m.Em},
	})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

// Atualizar muda status e/ou responsável e registra um aviso no histórico para cada mudança.
func (s *Store) Atualizar(ctx context.Context, id int64, status, responsavel *string, autor string) error {
	atual, err := s.Buscar(ctx, id, "")
	if err != nil {
		return err
	}
	agora := time.Now().UTC()
	set := bson.M{"atualizadoEm": agora}
	var avisos []Mensagem
	if status != nil && *status != atual.Status {
		set["status"] = *status
		avisos = append(avisos, Mensagem{Sistema: true, Em: agora,
			Texto: fmt.Sprintf("%s alterou o status para “%s”", autor, rotuloStatus[*status])})
	}
	if responsavel != nil && *responsavel != atual.Responsavel {
		set["responsavel"] = *responsavel
		texto := autor + " removeu o responsável"
		if *responsavel != "" {
			texto = fmt.Sprintf("%s atribuiu o chamado a %s", autor, *responsavel)
		}
		avisos = append(avisos, Mensagem{Sistema: true, Em: agora, Texto: texto})
	}
	update := bson.M{"$set": set}
	if len(avisos) > 0 {
		update["$push"] = bson.M{"mensagens": bson.M{"$each": avisos}}
	}
	_, err = s.chamados.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}
