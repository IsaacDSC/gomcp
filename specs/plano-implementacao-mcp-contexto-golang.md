# Plano de Implementacao - MCP de Contexto para Projetos Golang

## 1. Objetivo

Implementar no MCP a capacidade de:

- identificar automaticamente quando o workspace e um projeto Go;
- extrair contexto tecnico relevante do projeto;
- usar esse contexto para melhorar a qualidade das respostas e tarefas solicitadas no Cursor.

## 2. Resultado esperado

Quando o usuario pedir uma tarefa de desenvolvimento em um projeto Go, o MCP deve:

1. detectar que o projeto e Go;
2. montar um resumo de contexto (modulo, versao, estrutura, convencoes);
3. retornar esse contexto de forma estruturada para apoiar a escrita/implementacao.

## 3. Escopo do MVP

### Inclui

- Nova tool: `golang_project_context`.
- Deteccao de projeto Go por evidencias de workspace.
- Extracao de contexto essencial:
  - `go.mod` (module + versao Go);
  - estrutura de pastas (`cmd`, `internal`, `pkg`, `test`);
  - indicadores de qualidade (`Makefile`, `Dockerfile`, testes);
  - convencoes aplicaveis (Effective Go + regras locais).
- Resposta em JSON padronizado.

### Nao inclui no MVP

- analise semantica profunda do codigo;
- refatoracao automatica;
- suporte multi-linguagem com o mesmo nivel de profundidade.

## 3.1 Principios obrigatorios de implementacao

Este plano deve seguir obrigatoriamente os principios definidos em `specs/plano-implementacao-mcp-go.md`, com enfase em:

- preferir `internal/` para evitar acoplamento e exposicao desnecessaria de API;
- interfaces pequenas e orientadas ao consumidor;
- erros com contexto (`fmt.Errorf("...: %w", err)`);
- uso de `context.Context` em operacoes com I/O;
- composicao em vez de abstracoes prematuras;
- pacotes por dominio com nomes curtos e idiomaticos;
- testes table-driven para regras de negocio.

Convencoes de codigo (Effective Go + regras locais) obrigatorias:

- `gofmt` em todo codigo alterado;
- comentarios de documentacao em declaracoes exportadas;
- sem getters com prefixo `Get`;
- sem `import .` (exceto casos muito especificos de testes);
- early return e evitar `else` apos `return`;
- sem `panic` em fluxo normal da aplicacao;
- interfaces no pacote consumidor quando possivel;
- consistencia de receiver por tipo (pointer/value sem mistura arbitraria);
- concorrencia com estrategia clara de cancelamento e ownership de canais.

Checklist detalhado obrigatorio (implementacao e code review):

- Comentarios de documentacao devem preceder declaracoes exportadas.
- Comentarios devem explicar intencao/decisao, nao o obvio.
- Pacotes com nomes curtos, minusculos e sem underscore.
- Nome do pacote deve refletir o diretorio base (ex.: `internal/mcp` -> pacote `mcp`).
- Evitar colisao e redundancia: usar nomes evocativos e breves para facilitar leitura no import.
- Interfaces pequenas; para 1 metodo, priorizar sufixo `-er` quando semantico (`Reader`, `Writer` etc.).
- Usar early return para reduzir aninhamento.
- Evitar `else` apos `return` quando desnecessario.
- Preferir assinaturas pequenas e focadas por responsabilidade.
- Evitar panic em fluxo normal de aplicacao; retornar erro.
- Erro e valor explicito de retorno, sem esconder falhas.
- Enriquecer erro com contexto usando `%w`.
- Tratar e logar erro no nivel adequado (evitar log duplicado em camadas).
- Definir erros sentinela apenas quando houver necessidade real de comparacao.
- Definir interface no pacote consumidor, nao no provedor (quando possivel).
- Usar ponteiro receiver quando houver mutacao de estado ou estrutura grande.
- Usar value receiver para tipos pequenos/imutaveis e quando nao houver necessidade de mutacao.
- Regra de chamada: metodos com receiver por valor podem ser chamados em valor e ponteiro; metodos com receiver por ponteiro exigem ponteiro (exceto quando o valor for addressable e o compilador puder tomar endereco automaticamente).
- Manter consistencia por tipo: evitar misturar value e pointer receiver sem motivo claro.
- Preferir composicao e embedding com criterio, sem "heranca disfarcada".

## 4. Contrato da tool (proposta)

### Nome

`golang_project_context`

### Input (JSON)

```json
{
  "workspace_path": "/abs/path/opcional",
  "include_files": true,
  "max_depth": 4
}
```

### Output (JSON)

```json
{
  "is_golang_project": true,
  "confidence": "high",
  "signals": {
    "has_go_mod": true,
    "has_go_sum": true,
    "has_cmd_dir": true,
    "has_internal_dir": true
  },
  "module": {
    "name": "github.com/org/repo",
    "go_version": "1.26.1"
  },
  "layout": {
    "entrypoints": ["cmd/mcp-server/main.go"],
    "key_dirs": ["internal/mcp", "internal/tools", "test/integration"]
  },
  "quality": {
    "has_makefile": true,
    "has_dockerfile": true,
    "has_tests": true
  },
  "guidance": [
    "Aplicar gofmt em todo codigo alterado",
    "Preferir erros com contexto usando %w",
    "Evitar getters com prefixo Get"
  ]
}
```

## 5. Regras de deteccao de projeto Go

Pontuacao sugerida:

- `go.mod` presente: +60
- `*.go` no workspace: +20
- diretorios `cmd` ou `internal`: +10
- `go.sum` presente: +10

Classificacao:

- `high`: score >= 70
- `medium`: score entre 40 e 69
- `low`: score < 40

`is_golang_project = true` quando score >= 40.

## 6. Arquitetura de implementacao

## Fase 1 - Contrato e dominio

- Criar tipos de request/response da nova tool em `internal/tools`.
- Definir validacoes de input (`max_depth`, path).
- Padronizar erros de validacao.

## Fase 2 - Detector e extrator de contexto

- Criar servico `GoProjectDetector`:
  - verifica sinais;
  - calcula score/confianca;
  - retorna decisao.
- Criar servico `GoContextExtractor`:
  - parser basico de `go.mod`;
  - mapeamento de estrutura principal;
  - indicadores de qualidade do repositorio.

## Fase 3 - Integracao com MCP

- Registrar tool em `tools/list`.
- Implementar `tools/call` para `golang_project_context`.
- Garantir resposta serializavel e consistente.

## Fase 4 - Testes

- Unitarios:
  - detector (scores e limiares);
  - parser de `go.mod`;
  - validacao de input.
- Integracao:
  - fluxo MCP com chamada da nova tool.
  - cenario projeto nao-Go.

## Fase 5 - Documentacao

- Atualizar `README.md` com:
  - descricao da tool;
  - exemplo de request/response;
  - como usar no Cursor durante tarefas Go.

## 7. Criterios de aceite

- `tools/list` inclui `golang_project_context`.
- tool retorna `is_golang_project=true` em repos Go validos.
- output inclui modulo, versao Go, layout e guidance.
- `go test ./...` passa com cenarios positivos e negativos.
- README documenta uso da tool fim a fim.
- implementacao respeita os principios obrigatorios descritos na secao 3.1.

## 8. Riscos e mitigacoes

- **Risco:** falso positivo em projeto nao-Go.
  - **Mitigacao:** score por sinais multiplos, nao apenas extensao de arquivo.
- **Risco:** output muito verboso e pouco util.
  - **Mitigacao:** limitar campos essenciais no MVP.
- **Risco:** acoplamento forte ao layout atual.
  - **Mitigacao:** heuristicas flexiveis com fallback.

## 9. Proximos passos apos MVP

- adicionar leitura de linters (`golangci-lint`) e cobertura;
- incluir sugestoes contextuais por tipo de tarefa (bugfix, feature, refactor);
- suportar monorepo com multiplos `go.mod`;
- incorporar ranking de relevancia por pacote.
