# Referencia — Guidance Golang (MCP)

Convencoes retornadas por `golang_project_context`. Fonte: `internal/tools/go_context_tool.go` e `specs/plano-implementacao-mcp-contexto-golang.md`.

## Principios arquiteturais

- Preferir `internal/` para evitar acoplamento e exposicao desnecessaria de API.
- Interfaces pequenas e orientadas ao consumidor.
- Erros com contexto: `fmt.Errorf("...: %w", err)`.
- Usar `context.Context` em operacoes com I/O.
- Composicao em vez de abstracoes prematuras.
- Pacotes por dominio com nomes curtos e idiomaticos.
- Testes table-driven para regras de negocio.

## Convencoes Effective Go + regras locais

- Aplicar `gofmt` em todo codigo alterado.
- Comentarios de documentacao em declaracoes exportadas.
- Sem getters com prefixo `Get`.
- Sem `import .` (exceto casos muito especificos de testes).
- Early return e evitar `else` apos `return`.
- Sem `panic` em fluxo normal da aplicacao.
- Interfaces no pacote consumidor quando possivel.
- Consistencia de receiver por tipo (pointer/value sem mistura arbitraria).
- Concorrencia com estrategia clara de cancelamento e ownership de canais.

## Checklist de implementacao e code review

- Design explicito em vez de decisoes implicitas.
- Comentarios devem explicar intencao/decisao, nao o obvio.
- Codigo navegavel: organizacao e naming.
- Pacotes com nomes curtos, minusculos e sem underscore.
- Nome do pacote deve refletir o diretorio base (ex.: `internal/mcp` -> pacote `mcp`).
- Evitar colisao e redundancia: usar nomes evocativos e breves para facilitar leitura no import.
- Interface com 1 metodo: priorizar sufixo `-er` quando semantico (`Reader`, `Writer` etc.).
- Preferir assinaturas pequenas e focadas por responsabilidade.
- Se o metodo for utilizado em um unico ponto, mantenha apenas ele como responsavel pela orquestracao, evitando multiplos niveis desnecessarios.
- Erro e valor explicito de retorno, sem esconder falhas.
- Tratar e logar erro no nivel adequado (evitar log duplicado em camadas).
- Definir erros sentinela apenas quando houver necessidade real de comparacao.
- Definir interface no pacote consumidor, nao no provedor (quando possivel).
- Usar ponteiro receiver quando houver mutacao de estado ou estrutura grande.
- Usar value receiver para tipos pequenos/imutaveis sem necessidade de mutacao.
- Metodos com receiver por ponteiro exigem ponteiro (exceto quando addressable).
- Dependencias obrigatorias devem ser injetadas e validadas na func `New...`, evitando checagens de ponteiro nil dentro de metodos de negocio.
- Manter consistencia por tipo: evitar misturar value e pointer receiver sem motivo claro.
- Preferir composicao e embedding com criterio, sem heranca disfarcada.

## Contrato JSON da tool

### Input

```json
{
  "workspace_path": "/abs/path/opcional",
  "include_files": true,
  "max_depth": 4
}
```

### Output (exemplo)

```json
{
  "is_golang_project": true,
  "confidence": "high",
  "signals": {
    "has_go_mod": true,
    "has_go_sum": true,
    "has_go_files": true,
    "has_cmd_dir": true,
    "has_internal_dir": true
  },
  "module": {
    "name": "github.com/IsaacDSC/kvs",
    "go_version": "1.26.1"
  },
  "layout": {
    "entrypoints": ["cmd/node/main.go"],
    "key_dirs": ["cmd", "internal"],
    "sample_files": ["cmd/node/main.go", "internal/api/bulk_put_handle_http.go"]
  },
  "quality": {
    "has_makefile": true,
    "has_dockerfile": true,
    "has_tests": true
  },
  "guidance": ["..."]
}
```

## Erros comuns com MCP em Docker

| Problema | Causa | Solucao |
|----------|-------|---------|
| `-32603` ao chamar tool | `workspace_path` aponta para FS do host | Omitir parametro ou usar path do container |
| `is_golang_project: false` | Workspace errado ou sem `go.mod` | Verificar mount e diretorio de trabalho |
| Layout vazio | `max_depth` muito baixo | Aumentar `max_depth` (max 10) |
