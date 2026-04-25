# MCP Server em Go

Servidor MCP (Model Context Protocol) implementado em Go, com tools `echo`, `timestamp` e `golang_project_context`, com suporte de execucao local ou via Docker.

## Pre-requisitos

- Go 1.26.1+
- Docker e Docker Compose (opcional, para execucao em container)

## Setup local

Via compose:

```bash
docker compose up --build --force-recreate
```

## Healthcheck

One-shot healthcheck (local):

```bash
go run ./cmd/mcp-server --healthcheck
```

Healthcheck no container:

```bash
docker run --rm mcp-go-server:local --healthcheck
```

## Validacao rapida MCP

Com o processo rodando execute o comando abaixo

```sh 
go run ./example -tool golang_project_context -workspace . -max-depth 4 -include-files
```

## Configuracao no Cursor

No Cursor, adicione um servidor MCP em configuracao JSON usando o formato abaixo.

### Opcao 1: executar local com Go

```json
{
  "mcpServers": {
    "mcp-go-local": {
      "command": "go",
      "args": ["run", "./cmd/mcp-server"],
      "cwd": "/Users/isaacdsc/Projects/mcp",
      "env": {
        "LOG_LEVEL": "INFO"
      }
    }
  }
}
```

### Opcao 2: executar via Docker

```json
{
  "mcpServers": {
    "mcp-go-docker": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "-v",
        "/Users/isaacdsc/Projects/mcp:/workspace:ro",
        "-w",
        "/workspace",
        "mcp-go-server:local"
      ],
      "cwd": "/Users/isaacdsc/Projects/mcp",
      "env": {
        "LOG_LEVEL": "INFO"
      }
    }
  }
}
```

Com essa configuracao, a tool `golang_project_context` pode usar `workspace_path` como `/workspace` (ou omitir o campo para usar o diretório atual no container).

### Create cursor rules

*Baixe as rules com esse comando*

```bash
mkdir -p .cursor/rules && cd .cursor/rules && curl -fsSL "https://raw.githubusercontent.com/IsaacDSC/gomcp/main/.cursor/rules/mcp-golang-context.mdc" -o mcp-golang-context.mdc
```