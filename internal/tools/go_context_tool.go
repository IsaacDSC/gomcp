package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultMaxDepth = 4
	minMaxDepth     = 1
	maxMaxDepth     = 10
)

type golangProjectContextTool struct {
	basePath string
}

// NewGolangProjectContextTool creates the tool that inspects Go workspace context.
func NewGolangProjectContextTool(basePath string) Tool {
	return golangProjectContextTool{basePath: basePath}
}

func (golangProjectContextTool) Name() string { return "golang_project_context" }

func (golangProjectContextTool) Description() string {
	return "Detects if workspace is a Go project and returns structured project context."
}

func (golangProjectContextTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"workspace_path": map[string]any{
				"type":        "string",
				"description": "Optional absolute or relative workspace path.",
			},
			"include_files": map[string]any{
				"type":        "boolean",
				"description": "Include file-level details in layout output.",
			},
			"max_depth": map[string]any{
				"type":        "integer",
				"description": "Maximum directory depth to scan (1-10).",
			},
		},
	}
}

func (t golangProjectContextTool) Call(_ context.Context, arguments json.RawMessage) (map[string]any, error) {
	input, err := parseContextToolInput(arguments)
	if err != nil {
		return nil, fmt.Errorf("parse golang_project_context input: %w", err)
	}

	workspacePath, err := t.resolveWorkspacePath(input.WorkspacePath)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace path: %w", err)
	}

	signals, files, err := collectSignals(workspacePath, input.MaxDepth)
	if err != nil {
		return nil, fmt.Errorf("collect workspace signals: %w", err)
	}

	score := calculateGoScore(signals)
	confidence := scoreToConfidence(score)
	isGoProject := score >= 40

	moduleName, goVersion, err := parseGoMod(filepath.Join(workspacePath, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("parse go.mod: %w", err)
	}

	entrypoints := findEntrypoints(workspacePath, input.MaxDepth)
	keyDirs := findKeyDirs(signals)
	testsPresent := signals.HasTestDir || signals.HasGoTests

	layout := map[string]any{
		"entrypoints": entrypoints,
		"key_dirs":    keyDirs,
	}
	if input.IncludeFiles {
		layout["sample_files"] = files
	}

	return map[string]any{
		"is_golang_project": isGoProject,
		"confidence":        confidence,
		"signals": map[string]any{
			"has_go_mod":       signals.HasGoMod,
			"has_go_sum":       signals.HasGoSum,
			"has_go_files":     signals.HasGoFiles,
			"has_cmd_dir":      signals.HasCmdDir,
			"has_internal_dir": signals.HasInternalDir,
		},
		"module": map[string]any{
			"name":       moduleName,
			"go_version": goVersion,
		},
		"layout": layout,
		"quality": map[string]any{
			"has_makefile":   signals.HasMakefile,
			"has_dockerfile": signals.HasDockerfile,
			"has_tests":      testsPresent,
		},
		"guidance": defaultGuidance(),
	}, nil
}

type contextToolInput struct {
	WorkspacePath string `json:"workspace_path"`
	IncludeFiles  bool   `json:"include_files"`
	MaxDepth      int    `json:"max_depth"`
}

func parseContextToolInput(arguments json.RawMessage) (contextToolInput, error) {
	input := contextToolInput{
		MaxDepth: defaultMaxDepth,
	}

	if len(arguments) == 0 {
		return input, nil
	}
	if err := json.Unmarshal(arguments, &input); err != nil {
		return contextToolInput{}, fmt.Errorf("invalid JSON: %w", err)
	}
	if input.MaxDepth == 0 {
		input.MaxDepth = defaultMaxDepth
	}
	if input.MaxDepth < minMaxDepth || input.MaxDepth > maxMaxDepth {
		return contextToolInput{}, fmt.Errorf("max_depth must be between %d and %d", minMaxDepth, maxMaxDepth)
	}
	return input, nil
}

func (t golangProjectContextTool) resolveWorkspacePath(workspacePath string) (string, error) {
	candidate := strings.TrimSpace(workspacePath)
	if candidate == "" {
		candidate = t.basePath
	}
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(t.basePath, candidate)
	}

	absolute, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("compute absolute path: %w", err)
	}

	info, err := os.Stat(absolute)
	if err != nil {
		return "", fmt.Errorf("stat workspace: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workspace_path is not a directory")
	}
	return absolute, nil
}

type workspaceSignals struct {
	HasGoMod       bool
	HasGoSum       bool
	HasGoFiles     bool
	HasCmdDir      bool
	HasInternalDir bool
	HasMakefile    bool
	HasDockerfile  bool
	HasTestDir     bool
	HasGoTests     bool
}

func collectSignals(root string, maxDepth int) (workspaceSignals, []string, error) {
	signals := workspaceSignals{}
	collectedFiles := make([]string, 0, 8)

	if fileExists(filepath.Join(root, "go.mod")) {
		signals.HasGoMod = true
	}
	if fileExists(filepath.Join(root, "go.sum")) {
		signals.HasGoSum = true
	}
	if dirExists(filepath.Join(root, "cmd")) {
		signals.HasCmdDir = true
	}
	if dirExists(filepath.Join(root, "internal")) {
		signals.HasInternalDir = true
	}
	if fileExists(filepath.Join(root, "Makefile")) {
		signals.HasMakefile = true
	}
	if fileExists(filepath.Join(root, "Dockerfile")) {
		signals.HasDockerfile = true
	}
	if dirExists(filepath.Join(root, "test")) {
		signals.HasTestDir = true
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if depthFromRoot(rel) > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(d.Name(), ".go") {
			signals.HasGoFiles = true
			if len(collectedFiles) < 8 {
				collectedFiles = append(collectedFiles, filepath.ToSlash(rel))
			}
		}
		if strings.HasSuffix(d.Name(), "_test.go") {
			signals.HasGoTests = true
		}
		return nil
	})
	if err != nil {
		return workspaceSignals{}, nil, err
	}

	return signals, collectedFiles, nil
}

func parseGoMod(goModPath string) (string, string, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil
		}
		return "", "", err
	}
	defer file.Close()

	var moduleName string
	var goVersion string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
		if strings.HasPrefix(line, "go ") {
			goVersion = strings.TrimSpace(strings.TrimPrefix(line, "go "))
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", err
	}
	return moduleName, goVersion, nil
}

func calculateGoScore(signals workspaceSignals) int {
	score := 0
	if signals.HasGoMod {
		score += 60
	}
	if signals.HasGoFiles {
		score += 20
	}
	if signals.HasCmdDir || signals.HasInternalDir {
		score += 10
	}
	if signals.HasGoSum {
		score += 10
	}
	return score
}

func scoreToConfidence(score int) string {
	if score >= 70 {
		return "high"
	}
	if score >= 40 {
		return "medium"
	}
	return "low"
}

func findEntrypoints(root string, maxDepth int) []string {
	result := make([]string, 0, 4)
	cmdRoot := filepath.Join(root, "cmd")
	if !dirExists(cmdRoot) {
		return result
	}

	_ = filepath.WalkDir(cmdRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			if rel != "." && depthFromRoot(rel) > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "main.go" {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			result = append(result, filepath.ToSlash(rel))
		}
		return nil
	})

	return result
}

func findKeyDirs(signals workspaceSignals) []string {
	dirs := make([]string, 0, 4)
	if signals.HasCmdDir {
		dirs = append(dirs, "cmd")
	}
	if signals.HasInternalDir {
		dirs = append(dirs, "internal")
	}
	if signals.HasTestDir {
		dirs = append(dirs, "test")
	}
	return dirs
}

// defaultGuidance returns the mandatory coding guidelines derived from the
// project implementation plan (specs/plano-implementacao-mcp-contexto-golang.md).
func defaultGuidance() []string {
	return []string{
		// Principios arquiteturais
		"Preferir internal/ para evitar acoplamento e exposicao desnecessaria de API.",
		"Interfaces pequenas e orientadas ao consumidor.",
		"Erros com contexto: fmt.Errorf(\"...: %w\", err).",
		"Usar context.Context em operacoes com I/O.",
		"Composicao em vez de abstracoes prematuras.",
		"Pacotes por dominio com nomes curtos e idiomaticos.",
		"Testes table-driven para regras de negocio.",

		// Convencoes Effective Go + regras locais
		"Aplicar gofmt em todo codigo alterado.",
		"Comentarios de documentacao em declaracoes exportadas.",
		"Sem getters com prefixo Get.",
		"Sem import . (exceto casos muito especificos de testes).",
		"Early return e evitar else apos return.",
		"Sem panic em fluxo normal da aplicacao.",
		"Interfaces no pacote consumidor quando possivel.",
		"Consistencia de receiver por tipo (pointer/value sem mistura arbitraria).",
		"Concorrencia com estrategia clara de cancelamento e ownership de canais.",

		// Checklist de implementacao e code review
		"Comentarios devem explicar intencao/decisao, nao o obvio.",
		"Pacotes com nomes curtos, minusculos e sem underscore.",
		"Nome do pacote deve refletir o diretorio base (ex.: internal/mcp -> pacote mcp).",
		"Evitar colisao e redundancia: usar nomes evocativos e breves para facilitar leitura no import.",
		"Interface com 1 metodo: priorizar sufixo -er quando semantico (Reader, Writer etc.).",
		"Preferir assinaturas pequenas e focadas por responsabilidade.",
		"Erro e valor explicito de retorno, sem esconder falhas.",
		"Tratar e logar erro no nivel adequado (evitar log duplicado em camadas).",
		"Definir erros sentinela apenas quando houver necessidade real de comparacao.",
		"Definir interface no pacote consumidor, nao no provedor (quando possivel).",
		"Usar ponteiro receiver quando houver mutacao de estado ou estrutura grande.",
		"Usar value receiver para tipos pequenos/imutaveis sem necessidade de mutacao.",
		"Metodos com receiver por ponteiro exigem ponteiro (exceto quando addressable).",
		"Manter consistencia por tipo: evitar misturar value e pointer receiver sem motivo claro.",
		"Preferir composicao e embedding com criterio, sem heranca disfarcada.",
	}
}

func depthFromRoot(relPath string) int {
	if relPath == "." || relPath == "" {
		return 0
	}
	return strings.Count(filepath.ToSlash(relPath), "/") + 1
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
