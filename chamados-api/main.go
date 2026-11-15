package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/time/rate"
)

type visitante struct {
	lim   *rate.Limiter
	visto time.Time
}

type limitador struct {
	mu     sync.Mutex
	v      map[string]*visitante
	taxa   rate.Limit
	rajada int
}

func novoLimitador(porMinuto, rajada int) *limitador {
	l := &limitador{
		v:      map[string]*visitante{},
		taxa:   rate.Every(time.Minute / time.Duration(porMinuto)),
		rajada: rajada,
	}
	go l.limpar()
	return l
}

func (l *limitador) permitir(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.v[ip]
	if !ok {
		v = &visitante{lim: rate.NewLimiter(l.taxa, l.rajada)}
		l.v[ip] = v
	}
	v.visto = time.Now()
	return v.lim.Allow()
}

func (l *limitador) limpar() {
	for range time.Tick(5 * time.Minute) {
		l.mu.Lock()
		for ip, v := range l.v {
			if time.Since(v.visto) > 10*time.Minute {
				delete(l.v, ip)
			}
		}
		l.mu.Unlock()
	}
}

func ipDe(r *http.Request) string {

	if os.Getenv("TRUST_PROXY") == "true" {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			return strings.TrimSpace(strings.Split(xff, ",")[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (l *limitador) limitar(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !l.permitir(ipDe(r)) {
			w.Header().Set("Retry-After", "60")
			erro(w, http.StatusTooManyRequests, "muitas tentativas, aguarde um minuto")
			return
		}
		next(w, r)
	}
}

func env(chave, padrao string) string {
	if v := os.Getenv(chave); v != "" {
		return v
	}
	return padrao
}

func main() {
	carregarEnv(".env")

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(env("MONGO_URI", "mongodb://localhost:27017")))
	if err != nil {
		log.Fatalf("mongo: %v", err)
	}
	defer client.Disconnect(context.Background())

	iniCtx, cancelar := context.WithTimeout(ctx, 10*time.Second)
	defer cancelar()
	if err := client.Ping(iniCtx, nil); err != nil {
		log.Fatalf("não consegui falar com o MongoDB: %v", err)
	}
	store, err := NewStore(iniCtx, client.Database(env("DB_NAME", "chamados")))
	if err != nil {
		log.Fatalf("índices: %v", err)
	}
	api := &API{store: store}

	segredo := []byte(os.Getenv("JWT_SECRET"))
	if len(segredo) < 32 {
		log.Fatal("defina JWT_SECRET com pelo menos 32 caracteres")
	}
	auth := &Auth{store: store, segredo: segredo}
	if err := auth.garantirAdmin(iniCtx, env("ADMIN_NOME", "Administrador TI"), os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_SENHA")); err != nil {
		log.Fatalf("admin inicial: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		escrever(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	rota := func(padrao string, h http.HandlerFunc) { mux.Handle(padrao, auth.exigir(h)) }
	rota("GET /api/chamados", api.listar)
	rota("POST /api/chamados", api.criar)
	rota("GET /api/chamados/{id}", api.obter)
	rota("PATCH /api/chamados/{id}", api.atualizar)
	rota("POST /api/chamados/{id}/mensagens", api.enviarMensagem)

	limite := novoLimitador(10, 5)
	mux.HandleFunc("POST /api/auth/registrar", limite.limitar(auth.registrar))
	mux.HandleFunc("POST /api/auth/login", limite.limitar(auth.login))
	rota("GET /api/auth/eu", auth.eu)
	rota("GET /api/equipe", auth.equipe)
	rota("POST /api/usuarios", auth.criarUsuario)

	srv := &http.Server{
		Addr:              ":" + env("PORT", "8080"),
		Handler:           cors(env("CORS_ORIGIN", "http://localhost:5173"), registrar(mux)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
	}
	go func() {
		log.Printf("API ouvindo em %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("servidor: %v", err)
		}
	}()

	<-ctx.Done()
	fim, cancelarFim := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelarFim()
	_ = srv.Shutdown(fim)
}
