package desktopserver

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rdrgzma/robo-juridico-desktop/internal/bridge"
	"github.com/rdrgzma/robo-juridico-desktop/internal/calculator"
	"github.com/rdrgzma/robo-juridico-desktop/internal/prazos"
	_ "modernc.org/sqlite"
)

type DesktopServer struct {
	core    *bridge.CoreBridge
	dbLocal *sql.DB
	staticFS fs.FS
	router  *chi.Mux
}

type ItemHistorico struct {
	ID        int64     `json:"id"`
	Tipo      string    `json:"tipo"`       // "LIQUIDACAO", "REVISIONAL", "PRAZO", "CONSULTA"
	Titulo    string    `json:"titulo"`
	Descricao string    `json:"descricao"`
	Payload   string    `json:"payload"`
	Resultado string    `json:"resultado"`
	CriadoEm  time.Time `json:"criado_em"`
}

func NovoDesktopServer(core *bridge.CoreBridge, staticFS fs.FS, dbPath string) (*DesktopServer, error) {
	// Garante pasta do banco de dados local
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(10000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir banco local: %w", err)
	}

	// Cria tabela de histórico local do cliente
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS historico_desktop (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tipo TEXT NOT NULL,
			titulo TEXT NOT NULL,
			descricao TEXT,
			payload TEXT NOT NULL,
			resultado TEXT NOT NULL,
			criado_em DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("erro ao migrar historico: %w", err)
	}

	s := &DesktopServer{
		core:     core,
		dbLocal:  db,
		staticFS: staticFS,
		router:   chi.NewRouter(),
	}

	s.configurarRotas()
	return s, nil
}

func (s *DesktopServer) Router() http.Handler {
	return s.router
}

func (s *DesktopServer) Fechar() error {
	if s.dbLocal != nil {
		return s.dbLocal.Close()
	}
	return nil
}

func (s *DesktopServer) configurarRotas() {
	r := s.router

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS restrito para uso desktop / loopback
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if req.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, req)
		})
	})

	// Endpoints da API Desktop
	r.Route("/api", func(api chi.Router) {
		api.Get("/sistema/info", s.handleSistemaInfo)
		api.Post("/licenca/ativar", s.handleAtivarLicenca)

		api.Post("/processo/validar-cnj", s.handleValidarCNJ)
		api.Post("/processo/consultar", s.handleConsultarProcesso)

		api.Post("/calculo/liquidacao", s.handleCalcularLiquidacao)
		api.Post("/calculo/revisional", s.handleCalcularRevisional)
		api.Post("/calculo/prazos", s.handleCalcularPrazos)
		api.Post("/calculo/exportar-csv", s.handleExportarCSV)

		api.Get("/indices", s.handleListarIndices)
		api.Post("/indices/sincronizar", s.handleSincronizarIndices)

		api.Get("/historico", s.handleListarHistorico)
		api.Post("/historico/salvar", s.handleSalvarHistorico)
		api.Delete("/historico/{id}", s.handleExcluirHistorico)
	})

	// Arquivos Estáticos da UI Embutida
	if s.staticFS != nil {
		fileServer := http.FileServer(http.FS(s.staticFS))
		r.Handle("/*", fileServer)
	}
}

// Handlers da API

func (s *DesktopServer) handleSistemaInfo(w http.ResponseWriter, r *http.Request) {
	info := s.core.ObterInfoSistema()
	jsonResp(w, http.StatusOK, info)
}

func (s *DesktopServer) handleAtivarLicenca(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErro(w, http.StatusBadRequest, "Payload JSON invalido")
		return
	}

	dados, err := s.core.AtivarLicenca(body.Token)
	if err != nil {
		jsonErro(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	jsonResp(w, http.StatusOK, map[string]interface{}{
		"sucesso": true,
		"licenca": dados,
	})
}

func (s *DesktopServer) handleValidarCNJ(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Numero string `json:"numero"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErro(w, http.StatusBadRequest, "Informe o numero do processo")
		return
	}

	res := s.core.ValidarCNJ(body.Numero)
	jsonResp(w, http.StatusOK, res)
}

func (s *DesktopServer) handleConsultarProcesso(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Numero string `json:"numero"`
		Modo   string `json:"modo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErro(w, http.StatusBadRequest, "Payload invalido")
		return
	}

	proc, avisos, err := s.core.ConsultarProcesso(r.Context(), body.Numero, body.Modo)
	if err != nil {
		jsonErro(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResp(w, http.StatusOK, map[string]interface{}{
		"sucesso":  true,
		"processo": proc,
		"avisos":   avisos,
	})
}

func (s *DesktopServer) handleCalcularLiquidacao(w http.ResponseWriter, r *http.Request) {
	var req calculator.CalculoJudicialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErro(w, http.StatusBadRequest, "Formato do calculo invalido: "+err.Error())
		return
	}

	res, err := s.core.CalcularLiquidacao(req)
	if err != nil {
		jsonErro(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	jsonResp(w, http.StatusOK, res)
}

func (s *DesktopServer) handleCalcularRevisional(w http.ResponseWriter, r *http.Request) {
	var req calculator.CalculoFinanciamentoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErro(w, http.StatusBadRequest, "Formato da requisicao revisional invalido: "+err.Error())
		return
	}

	res, err := s.core.CalcularRevisional(req)
	if err != nil {
		jsonErro(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	jsonResp(w, http.StatusOK, res)
}

func (s *DesktopServer) handleCalcularPrazos(w http.ResponseWriter, r *http.Request) {
	var req prazos.CalculoPrazoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErro(w, http.StatusBadRequest, "Parametros de prazo invalidos: "+err.Error())
		return
	}

	res, err := s.core.CalcularPrazo(req)
	if err != nil {
		jsonErro(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResp(w, http.StatusOK, res)
}

func (s *DesktopServer) handleExportarCSV(w http.ResponseWriter, r *http.Request) {
	var req calculator.CalculoJudicialResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErro(w, http.StatusBadRequest, "Objeto de calculo invalido")
		return
	}

	csvStr, err := s.core.ExportarCalculoCSV(&req)
	if err != nil {
		jsonErro(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"memoria_calculo_judicial.csv\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(csvStr))
}

func (s *DesktopServer) handleListarIndices(w http.ResponseWriter, r *http.Request) {
	indices := s.core.ListarIndices()
	jsonResp(w, http.StatusOK, indices)
}

func (s *DesktopServer) handleSincronizarIndices(w http.ResponseWriter, r *http.Request) {
	pontos, err := s.core.SincronizarIndices(r.Context())
	if err != nil {
		jsonErro(w, http.StatusInternalServerError, "Erro ao sincronizar: "+err.Error())
		return
	}

	jsonResp(w, http.StatusOK, map[string]interface{}{
		"sucesso":           true,
		"pontos_atualizados": pontos,
		"mensagem":          fmt.Sprintf("Sincronizacao concluida. %d pontos atualizados.", pontos),
	})
}

func (s *DesktopServer) handleListarHistorico(w http.ResponseWriter, r *http.Request) {
	rows, err := s.dbLocal.Query(`
		SELECT id, tipo, titulo, descricao, payload, resultado, criado_em 
		FROM historico_desktop 
		ORDER BY id DESC LIMIT 100
	`)
	if err != nil {
		jsonErro(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var lista []ItemHistorico
	for rows.Next() {
		var item ItemHistorico
		var desc sql.NullString
		if err := rows.Scan(&item.ID, &item.Tipo, &item.Titulo, &desc, &item.Payload, &item.Resultado, &item.CriadoEm); err == nil {
			if desc.Valid {
				item.Descricao = desc.String
			}
			lista = append(lista, item)
		}
	}

	if lista == nil {
		lista = []ItemHistorico{}
	}

	jsonResp(w, http.StatusOK, lista)
}

func (s *DesktopServer) handleSalvarHistorico(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Tipo      string `json:"tipo"`
		Titulo    string `json:"titulo"`
		Descricao string `json:"descricao"`
		Payload   string `json:"payload"`
		Resultado string `json:"resultado"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErro(w, http.StatusBadRequest, "Dados do historico invalidos")
		return
	}

	res, err := s.dbLocal.Exec(`
		INSERT INTO historico_desktop (tipo, titulo, descricao, payload, resultado)
		VALUES (?, ?, ?, ?, ?)
	`, body.Tipo, body.Titulo, body.Descricao, body.Payload, body.Resultado)
	if err != nil {
		jsonErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	novoID, _ := res.LastInsertId()
	jsonResp(w, http.StatusCreated, map[string]interface{}{
		"sucesso": true,
		"id":      novoID,
	})
}

func (s *DesktopServer) handleExcluirHistorico(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := s.dbLocal.Exec(`DELETE FROM historico_desktop WHERE id = ?`, id)
	if err != nil {
		jsonErro(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResp(w, http.StatusOK, map[string]bool{"sucesso": true})
}

// Helpers

func jsonResp(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func jsonErro(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso": false,
		"erro":    msg,
	})
}
