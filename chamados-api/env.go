package main

import (
	"bufio"
	"os"
	"strings"
)

// carregarEnv lê o arquivo .env (se existir) e define as variáveis que ainda não estão no ambiente.
// Variáveis já definidas no terminal (ou no Render) têm prioridade.
func carregarEnv(caminho string) {
	f, err := os.Open(caminho)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		linha := strings.TrimSpace(strings.TrimPrefix(sc.Text(), "\ufeff"))
		if linha == "" || strings.HasPrefix(linha, "#") {
			continue
		}
		chave, valor, ok := strings.Cut(linha, "=")
		if !ok {
			continue
		}
		chave, valor = strings.TrimSpace(chave), strings.TrimSpace(valor)
		if n := len(valor); n >= 2 && (valor[0] == '"' && valor[n-1] == '"' || valor[0] == '\'' && valor[n-1] == '\'') {
			valor = valor[1 : n-1]
		}
		if _, existe := os.LookupEnv(chave); !existe {
			os.Setenv(chave, valor)
		}
	}
}
