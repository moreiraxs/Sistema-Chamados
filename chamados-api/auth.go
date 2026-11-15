package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const (
	PerfilUsuario = "usuario"
	PerfilTI      = "ti"
)

var ErrEmailEmUso = errors.New("e-mail já cadastrado")

type Usuario struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Nome      string             `bson:"nome" json:"nome"`
	Email     string             `bson:"email" json:"email"`
	SenhaHash string             `bson:"senhaHash" json:"-"`
	Perfil    string             `bson:"perfil" json:"perfil"`
	CriadoEm  time.Time          `bson:"criadoEm" json:"-"`
}

// ---- acesso ao banco ----

func (s *Store) CriarUsuario(ctx context.Context, u *Usuario) error {
	u.CriadoEm = time.Now().UTC()
	res, err := s.usuarios.InsertOne(ctx, u)
	if mongo.IsDuplicateKeyError(err) {
		return ErrEmailEmUso
	}
	if err != nil {
		return err
	}
	u.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (s *Store) umUsuario(ctx context.Context, filtro bson.M) (*Usuario, error) {
	var u Usuario
	err := s.usuarios.FindOne(ctx, filtro).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) UsuarioPorEmail(ctx context.Context, email string) (*Usuario, error) {
	return s.umUsuario(ctx, bson.M{"email": email})
}

func (s *Store) UsuarioPorID(ctx context.Context, id string) (*Usuario, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil
	}
	return s.umUsuario(ctx, bson.M{"_id": oid})
}

func (s *Store) NomesTI(ctx context.Context) ([]string, error) {
	cur, err := s.usuarios.Find(ctx, bson.M{"perfil": PerfilTI}, options.Find().SetSort(bson.D{{Key: "nome", Value: 1}}))
	if err != nil {
		return nil, err
	}
	var us []Usuario
	if err := cur.All(ctx, &us); err != nil {
		return nil, err
	}
	nomes := make([]string, 0, len(us))
	for _, u := range us {
		nomes = append(nomes, u.Nome)
	}
	return nomes, nil
}

// ---- autenticação ----

type Auth struct {
	store   *Store
	segredo []byte
}

type claims struct {
	Nome   string `json:"nome"`
	Perfil string `json:"perfil"`
	jwt.RegisteredClaims
}

func (a *Auth) emitir(u *Usuario) (string, error) {
	agora := time.Now()
	c := claims{Nome: u.Nome, Perfil: u.Perfil, RegisteredClaims: jwt.RegisteredClaims{
		Subject:   u.ID.Hex(),
		IssuedAt:  jwt.NewNumericDate(agora),
		ExpiresAt: jwt.NewNumericDate(agora.Add(12 * time.Hour)),
	}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(a.segredo)
}

// exigir protege uma rota: só passa com um token Bearer válido.
func (a *Auth) exigir(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bruto, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok {
			erro(w, http.StatusUnauthorized, "faça login para continuar")
			return
		}
		var c claims
		tok, err := jwt.ParseWithClaims(bruto, &c, func(*jwt.Token) (any, error) { return a.segredo, nil },
			jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !tok.Valid {
			erro(w, http.StatusUnauthorized, "sessão inválida ou expirada")
			return
		}
		ident := Identidade{ID: c.Subject, Nome: c.Nome, EhTI: c.Perfil == PerfilTI}
		next(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, ident)))
	})
}

func (a *Auth) garantirAdmin(ctx context.Context, nome, email, senha string) error {
	if email == "" || senha == "" {
		return nil
	}
	n, err := a.store.usuarios.CountDocuments(ctx, bson.M{"perfil": PerfilTI})
	if err != nil || n > 0 {
		return err
	}
	_, err = a.novoUsuario(ctx, cadastro{Nome: nome, Email: email, Senha: senha}, PerfilTI)
	return err
}

// ---- handlers ----

type cadastro struct {
	Nome   string `json:"nome"`
	Email  string `json:"email"`
	Senha  string `json:"senha"`
	Perfil string `json:"perfil"`
}

type respostaAuth struct {
	Token   string   `json:"token"`
	Usuario *Usuario `json:"usuario"`
}

func normalizarEmail(e string) string { return strings.ToLower(strings.TrimSpace(e)) }

func (c *cadastro) validar() string {
	c.Nome = strings.TrimSpace(c.Nome)
	c.Email = normalizarEmail(c.Email)
	switch {
	case c.Nome == "" || utf8.RuneCountInString(c.Nome) > 100:
		return "informe um nome de até 100 caracteres"
	case !strings.Contains(c.Email, "@") || len(c.Email) > 200:
		return "e-mail inválido"
	case len(c.Senha) < 8 || len(c.Senha) > 72:
		return "a senha deve ter de 8 a 72 caracteres"
	}
	return ""
}

func (a *Auth) novoUsuario(ctx context.Context, c cadastro, perfil string) (*Usuario, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(c.Senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &Usuario{Nome: strings.TrimSpace(c.Nome), Email: normalizarEmail(c.Email), SenhaHash: string(hash), Perfil: perfil}
	if err := a.store.CriarUsuario(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (a *Auth) responderToken(w http.ResponseWriter, u *Usuario, status int) {
	tok, err := a.emitir(u)
	if err != nil {
		log.Printf("jwt: %v", err)
		erro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	escrever(w, status, respostaAuth{Token: tok, Usuario: u})
}

func falhaCadastro(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrEmailEmUso) {
		erro(w, http.StatusConflict, err.Error())
		return
	}
	log.Printf("cadastro: %v", err)
	erro(w, http.StatusInternalServerError, "erro interno")
}

// POST /api/auth/registrar  (público: sempre cria um usuário comum)
func (a *Auth) registrar(w http.ResponseWriter, r *http.Request) {
	var in cadastro
	if !lerJSON(w, r, &in) {
		return
	}
	if msg := in.validar(); msg != "" {
		erro(w, http.StatusBadRequest, msg)
		return
	}
	u, err := a.novoUsuario(r.Context(), in, PerfilUsuario)
	if err != nil {
		falhaCadastro(w, err)
		return
	}
	a.responderToken(w, u, http.StatusCreated)
}

var hashFalso, _ = bcrypt.GenerateFromPassword([]byte("senha-falsa"), bcrypt.DefaultCost)

// POST /api/auth/login
func (a *Auth) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}
	if !lerJSON(w, r, &in) {
		return
	}
	u, err := a.store.UsuarioPorEmail(r.Context(), normalizarEmail(in.Email))
	if err != nil {
		log.Printf("login: %v", err)
		erro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	hash := hashFalso
	if u != nil {
		hash = []byte(u.SenhaHash)
	}
	errSenha := bcrypt.CompareHashAndPassword(hash, []byte(in.Senha))
	if u == nil || errSenha != nil {
		erro(w, http.StatusUnauthorized, "e-mail ou senha incorretos")
		return
	}
	a.responderToken(w, u, http.StatusOK)
}

// GET /api/auth/eu
func (a *Auth) eu(w http.ResponseWriter, r *http.Request) {
	u, err := a.store.UsuarioPorID(r.Context(), identidadeDe(r).ID)
	if err != nil || u == nil {
		erro(w, http.StatusUnauthorized, "sessão inválida ou expirada")
		return
	}
	escrever(w, http.StatusOK, u)
}

// GET /api/equipe  (somente TI): nomes da equipe, para escolher o responsável
func (a *Auth) equipe(w http.ResponseWriter, r *http.Request) {
	if !identidadeDe(r).EhTI {
		erro(w, http.StatusForbidden, "somente a TI")
		return
	}
	nomes, err := a.store.NomesTI(r.Context())
	if err != nil {
		log.Printf("equipe: %v", err)
		erro(w, http.StatusInternalServerError, "erro interno")
		return
	}
	escrever(w, http.StatusOK, nomes)
}

// POST /api/usuarios  (somente TI): cria usuário comum ou outro membro da TI
func (a *Auth) criarUsuario(w http.ResponseWriter, r *http.Request) {
	if !identidadeDe(r).EhTI {
		erro(w, http.StatusForbidden, "somente a TI cria contas")
		return
	}
	var in cadastro
	if !lerJSON(w, r, &in) {
		return
	}
	if in.Perfil == "" {
		in.Perfil = PerfilUsuario
	}
	if in.Perfil != PerfilUsuario && in.Perfil != PerfilTI {
		erro(w, http.StatusBadRequest, "perfil inválido (use usuario ou ti)")
		return
	}
	if msg := in.validar(); msg != "" {
		erro(w, http.StatusBadRequest, msg)
		return
	}
	u, err := a.novoUsuario(r.Context(), in, in.Perfil)
	if err != nil {
		falhaCadastro(w, err)
		return
	}
	escrever(w, http.StatusCreated, u)
}
