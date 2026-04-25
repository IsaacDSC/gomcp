# Plano de Implementacao - Projeto MCP em Go com Docker

## 1. Objetivo do projeto

Construir um servidor MCP (Model Context Protocol) em Go, com foco em:

- arquitetura simples e extensivel;
- boas praticas de engenharia em Go;
- observabilidade basica;
- empacotamento e execucao via Docker;
- facilidade de manutencao e evolucao.

## 2. Escopo inicial (MVP)

### Funcionalidades do MVP

- Expor um servidor MCP funcional com pelo menos:
  - `tools/list`;
  - `tools/call` para uma ou duas ferramentas de exemplo;
  - endpoint/fluxo de `healthcheck`.
- Estruturar projeto para crescimento modular.
- Adicionar logs estruturados.
- Adicionar testes unitarios essenciais.
- Gerar imagem Docker pronta para execucao.

### Fora do escopo inicial

- persistencia complexa (banco relacional/NoSQL);
- autenticacao/autorizacao avancada;
- autoscaling e operacao em Kubernetes.

## 3. Requisitos nao funcionais

- **Confiabilidade**: tratar erros com contexto e mensagens claras.
- **Observabilidade**: logs estruturados e metricas basicas.
- **Portabilidade**: build reproduzivel local e em CI.
- **Seguranca**: executar container com usuario nao-root.
- **Qualidade**: cobertura de testes no nucleo do dominio.

## 4. Stack sugerida

- **Linguagem**: Go (versao estavel atual suportada pelo time).
- **Roteamento/servidor**: `net/http` (preferencia por simplicidade no MVP).
- **Logging**: `slog` (biblioteca padrao) ou logger estruturado equivalente.
- **Config**: variaveis de ambiente + arquivo `.env` apenas para desenvolvimento local.
- **Testes**: `testing` + table-driven tests.
- **Lint/format**: `gofmt`, `go vet`, `golangci-lint`.
- **Container**: Docker com build multi-stage.

## 5. Estrutura de pastas recomendada

```text
.
├── cmd/
│   └── mcp-server/
│       └── main.go
├── internal/
│   ├── app/              # bootstrap e wiring
│   ├── mcp/              # handlers/protocolo MCP
│   ├── tools/            # implementacao das tools
│   ├── config/           # leitura e validacao de config
│   └── observability/    # logger, metricas, tracing (futuro)
├── pkg/                  # componentes reutilizaveis publicos (se necessario)
├── specs/
│   └── plano-implementacao-mcp-go.md
├── test/
│   └── integration/
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

> Regra pratica: preferir `internal/` para evitar acoplamento e exposicao desnecessaria de API.

## 6. Principios de implementacao em Go

- Manter interfaces pequenas e orientadas ao consumidor.
- Retornar erros com contexto (`fmt.Errorf("...: %w", err)`).
- Evitar estado global mutavel.
- Usar `context.Context` em operacoes com I/O.
- Preferir composicao a heranca/abstracoes prematuras.
- Nomear pacotes por dominio (ex.: `tools`, `config`, `mcp`).
- Escrever testes table-driven para regras de negocio.

### 6.1 Convencoes obrigatorias (alinhadas ao Effective Go)

Referencia: [Effective Go](https://go.dev/doc/effective_go)

> Observacao: o documento oficial nao e atualizado com alta frequencia, entao estas convencoes devem ser combinadas com as praticas atuais de modulos Go e notas de release.

#### Formatacao e comentarios

- Todo codigo deve passar por `gofmt` (sem excecao).
- Comentarios de documentacao devem preceder declaracoes exportadas.
- Comentarios devem explicar intencao/decisao, nao o obvio.

#### Nomenclatura

- Pacotes com nomes curtos, minusculos e sem underscore.
- Nome do pacote deve refletir o diretorio base (ex.: `internal/mcp` -> pacote `mcp`).
- Evitar colisao e redundancia: usar nomes evocativos e breves para facilitar leitura no import.
- Evitar repeticao no nome (ex.: em `tools`, preferir `Runner` a `ToolRunner` quando fizer sentido).
- Nomes exportados devem considerar o prefixo do pacote (`mcp.Server`, nao `mcp.MCPServer` sem necessidade).
- Getters sem prefixo `Get` (usar `Owner()`, por exemplo).
- Interfaces pequenas; para 1 metodo, priorizar sufixo `-er` quando semantico (`Reader`, `Writer` etc.).
- Evitar `import .` fora de casos muito especificos de testes.

#### Fluxo de controle e funcoes

- Usar early return para reduzir aninhamento.
- Evitar `else` apos `return` quando desnecessario.
- Preferir assinaturas pequenas e focadas por responsabilidade.
- Evitar panic em fluxo normal de aplicacao; retornar erro.

#### Erros

- Erro e valor explicito de retorno, sem esconder falhas.
- Enriquecer erro com contexto usando `%w`.
- Tratar e logar erro no nivel adequado (evitar log duplicado em camadas).
- Definir erros sentinela apenas quando houver necessidade real de comparacao.

#### Metodos, interfaces e composicao

- Definir interface no pacote consumidor, nao no provedor (quando possivel).
- Usar ponteiro receiver quando houver mutacao de estado ou estrutura grande.
- Usar value receiver para tipos pequenos/imutaveis e quando nao houver necessidade de mutacao.
- Regra de chamada: metodos com receiver por valor podem ser chamados em valor e ponteiro; metodos com receiver por ponteiro exigem ponteiro (exceto quando o valor for addressable e o compilador puder tomar endereco automaticamente).
- Manter consistencia por tipo: evitar misturar value e pointer receiver sem motivo claro.
- Preferir composicao e embedding com criterio, sem "heranca disfarcada".

#### Concorrencia

- So introduzir goroutines quando houver ganho claro.
- Coordenar cancelamento com `context.Context`.
- Evitar goroutine sem estrategia de encerramento.
- Canais devem ter ownership claro (quem cria, escreve e fecha).

## 7. Plano por fases

## Fase 0 - Bootstrap do repositorio

- Inicializar modulo Go (`go mod init`).
- Criar estrutura base de diretorios.
- Configurar `.gitignore` para Go e arquivos locais.
- Criar `README.md` com instrucoes de build e run.

**Entrega:** projeto compila com `go build ./...`.

## Fase 1 - Servidor MCP basico

- Subir servidor HTTP com endpoint de health.
- Implementar handlers iniciais do protocolo MCP:
  - listagem de tools;
  - execucao de tool simples (ex.: echo, timestamp).
- Padronizar respostas e erros.

**Entrega:** fluxo MCP basico funcional localmente.

## Fase 2 - Qualidade e boas praticas

- Adicionar testes unitarios dos componentes principais.
- Criar pipeline local de qualidade:
  - `gofmt -w`;
  - `go vet ./...`;
  - `go test ./...`.
- Adicionar `Makefile` com alvos (`fmt`, `lint`, `test`, `run`).

**Entrega:** rotina minima de qualidade automatizada.

## Fase 3 - Dockerizacao

- Criar `Dockerfile` multi-stage:
  - etapa builder em imagem Go;
  - etapa final slim/distroless.
- Copiar apenas binario e artefatos necessarios.
- Rodar como usuario nao-root.
- Definir `HEALTHCHECK` basico no container.

**Entrega:** imagem executavel com `docker run`.

## Fase 4 - Operacao local e documentacao

- Criar `docker-compose.yml` para execucao local.
- Documentar variaveis de ambiente.
- Documentar exemplos de chamada MCP.

**Entrega:** onboarding de dev em poucos minutos.

## 8. Docker - recomendacoes de boas praticas

- Fixar versoes base (evitar `latest`).
- Usar build cache de modulos (`go mod download`) em camada separada.
- Reduzir superficie de ataque (imagem final minima).
- Definir `USER` nao-root.
- Expor apenas porta necessaria.
- Incluir `HEALTHCHECK`.

## 9. Estrategia de testes

- **Unitarios**: regras de negocio em `internal/tools` e `internal/mcp`.
- **Integracao**: subir servidor e validar fluxo MCP essencial.
- **Smoke em container**: validar `docker run` + `healthcheck`.

Exemplos de criterios de aceite:

- `go test ./...` sem falhas;
- servidor responde health com sucesso;
- `tools/list` retorna ferramentas registradas;
- `tools/call` executa ferramenta de exemplo corretamente.

## 10. CI/CD (sugestao inicial)

Pipeline minima em cada push/PR:

1. `go mod tidy` (verificacao)
2. `gofmt` (checagem)
3. `go vet ./...`
4. `go test ./...`
5. `docker build`

Opcional em fase posterior:

- publish de imagem em registry;
- scan de vulnerabilidades da imagem;
- assinaturas de artefato.

## 11. Riscos e mitigacoes

- **Risco:** acoplamento precoce ao protocolo.
  - **Mitigacao:** camada `internal/mcp` isolada do dominio de tools.
- **Risco:** crescimento sem padrao.
  - **Mitigacao:** conventions no `README` e revisao de PR orientada por checklists.
- **Risco:** imagem Docker grande.
  - **Mitigacao:** multi-stage + imagem final minima.

## 12. Checklist inicial de implementacao

- [ ] Inicializar modulo Go e estrutura de pastas
- [ ] Implementar servidor MCP minimo
- [ ] Implementar duas tools de exemplo
- [ ] Adicionar testes unitarios essenciais
- [ ] Criar `Makefile` com comandos padrao
- [ ] Dockerfile multi-stage com usuario nao-root
- [ ] `docker-compose.yml` para ambiente local
- [ ] Atualizar `README.md` com instrucoes de uso
- [ ] Garantir `gofmt` e convencoes de nomenclatura do Effective Go
- [ ] Revisar tratamento de erros com `%w` nas camadas principais

## 13. Proximos passos apos MVP

- adicionar autenticacao por token/chave;
- incluir metricas Prometheus;
- adicionar tracing distribuido;
- implementar versionamento de tools;
- adicionar testes de contrato MCP.

