---
name: dev-golang
description: >-
  Desenvolve e revisa codigo Go usando contexto do MCP golang_project_context.
  Use ao implementar features, corrigir bugs, refatorar ou revisar codigo Go;
  quando o usuario mencionar Go, golang, go.mod, cmd/, internal/ ou este repositorio MCP.
---

# Dev Golang (MCP)

Skill para desenvolvimento Go orientado pelo MCP `golang_project_context`. O contexto vem da tool — nao assuma layout ou convencoes sem consulta-la.

## Fluxo obrigatorio

Antes de propor ou implementar mudancas em codigo Go:

1. Chame `golang_project_context` com argumentos padrao: `{}` ou omita campos opcionais.
2. **Nao** passe `workspace_path` com caminho absoluto do host (ex.: `/Users/...`) se o MCP roda em Docker ou ambiente isolado — causa erro `-32603`.
3. Use `workspace_path` apenas quando for um path **dentro** do runtime do MCP (mesmo mount/container/workdir).
4. Resuma o retorno relevante: modulo, versao Go, layout, sinais de qualidade, guidance.
5. Trate o retorno como restricoes para implementacao e review.

Se a tool falhar, reporte o erro e peca retry antes de continuar.

## Invocacao da tool

```json
{}
```

Com mais detalhe de layout (quando necessario):

```json
{
  "include_files": true,
  "max_depth": 4
}
```

| Parametro | Tipo | Default | Descricao |
|-----------|------|---------|-----------|
| `workspace_path` | string | base do MCP | Path opcional; omitir na duvida |
| `include_files` | boolean | false | Inclui `sample_files` no layout |
| `max_depth` | integer | 4 | Profundidade de scan (1–10) |

## Interpretar o retorno

| Campo | Uso |
|-------|-----|
| `is_golang_project` | false → confirmar antes de assumir projeto Go |
| `confidence` | high/medium/low — ajustar rigor conforme confianca |
| `module.name` | Import path e referencias de pacote |
| `module.go_version` | Versao minima de Go |
| `layout.entrypoints` | Onde comecar (`cmd/*/main.go`) |
| `layout.key_dirs` | Onde colocar codigo (`cmd`, `internal`, `test`) |
| `quality` | Makefile, Dockerfile, testes existentes |
| `guidance` | Convencoes obrigatorias — aplicar em todo codigo novo/alterado |

### Score de deteccao (referencia)

- `go.mod`: +60 | `*.go`: +20 | `cmd` ou `internal`: +10 | `go.sum`: +10
- `is_golang_project = true` quando score >= 40
- `confidence`: high >= 70, medium 40–69, low < 40

## Workflow de implementacao

```
Task Progress:
- [ ] Chamar golang_project_context
- [ ] Resumir contexto para o usuario
- [ ] Ler codigo existente nos pacotes afetados
- [ ] Implementar seguindo guidance do MCP
- [ ] Validar com make lint (ou fmt + vet + test)
```

### Onde colocar codigo

- **`cmd/`** — entrypoints (`main.go` fino; wiring apenas)
- **`internal/`** — logica de dominio, MCP, tools (nao exportar fora do modulo)
- **`test/`** — testes de integracao quando aplicavel
- Evitar `pkg/` salvo necessidade real de API publica

### Validacao local

```bash
make lint    # fmt + vet + test
make build   # bin/mcp-server
make run     # go run ./cmd/mcp-server
```

Testar a tool MCP diretamente:

```bash
go run ./example -tool golang_project_context -workspace . -max-depth 4 -include-files
```

## Checklist de review

Aplicar itens de `guidance` retornados pelo MCP. Prioridades:

- Erros com contexto: `fmt.Errorf("...: %w", err)`
- `context.Context` em operacoes com I/O
- Sem `panic` em fluxo normal; sem getters `Get`
- Early return; sem `else` apos `return`
- Interfaces pequenas no pacote **consumidor**
- Receivers consistentes por tipo (pointer vs value)
- Dependencias injetadas e validadas em `New...`
- Testes table-driven para regras de negocio
- `gofmt` em todo codigo alterado

Lista completa em [reference.md](reference.md).

## Estrutura deste repositorio (MCP Go)

```
cmd/mcp-server/     # entrypoint
internal/
  app/              # bootstrap
  mcp/              # protocolo MCP
  tools/            # echo, timestamp, golang_project_context
  config/
  observability/
test/integration/   # testes MCP end-to-end
specs/              # planos de implementacao
```

Tools MCP disponiveis: `echo`, `timestamp`, `golang_project_context`.

## Quando nao usar esta skill

- Tarefas puramente de documentacao ou git sem codigo Go
- Projetos confirmados como nao-Go (`is_golang_project: false`)
- Ambiente MCP indisponivel — informar o usuario em vez de inventar contexto

## Recursos

- Convencoes completas: [reference.md](reference.md)
- Plano da tool: `specs/plano-implementacao-mcp-contexto-golang.md`
- Principios do projeto: `specs/plano-implementacao-mcp-go.md`
