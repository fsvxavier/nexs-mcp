# Troubleshooting Guide

Guia para resolução de problemas comuns do NEXS MCP Server.

## 📋 Índice

1. [Problemas de Instalação](#problemas-de-instalação)
2. [Problemas de Execução](#problemas-de-execução)
3. [Integração com Claude Desktop](#integração-com-claude-desktop)
4. [Problemas de Storage](#problemas-de-storage)
5. [Performance](#performance)
6. [Logs e Debug](#logs-e-debug)

---

## Problemas de Instalação

### Go version incompatível

**Sintoma:**
```
go: module requires Go 1.25 or later
```

**Solução:**
```bash
# Verificar versão atual
go version

# Atualizar Go para 1.25+
# Linux/macOS:
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
```

### Build falha

**Sintoma:**
```
# github.com/fsvxavier/nexs-mcp/internal/...
undefined: ...
```

**Solução:**
```bash
# Limpar cache e rebuild
go clean -cache -modcache
go mod download
go mod tidy
make build
```

---

## Problemas de Execução

### Server não inicia

**Sintoma:**
```
panic: runtime error
```

**Diagnóstico:**
```bash
# Verificar se o binário está corrompido
file ./bin/nexs-mcp

# Recompilar
make clean build

# Testar com verbose
./bin/nexs-mcp -h
```

### Permissões negadas

**Sintoma:**
```
permission denied: ./bin/nexs-mcp
```

**Solução:**
```bash
# Adicionar permissão de execução
chmod +x ./bin/nexs-mcp

# Verificar
ls -l ./bin/nexs-mcp
```

### Porta já em uso

**Sintoma:**
```
bind: address already in use
```

**Solução:**
```bash
# Encontrar processo usando a porta
lsof -i :PORT_NUMBER
# ou
netstat -tulpn | grep PORT_NUMBER

# Matar processo
kill -9 PID
```

---

## Integração com Claude Desktop

### Server não aparece no Claude

**Diagnóstico:**

1. **Verificar configuração:**
```bash
# macOS
cat ~/Library/Application\ Support/Claude/claude_desktop_config.json

# Linux
cat ~/.config/Claude/claude_desktop_config.json
```

2. **Validar JSON:**
```bash
# Usar um validador JSON online ou:
python3 -m json.tool < claude_desktop_config.json
```

3. **Verificar caminho do binário:**
```bash
# Testar se o caminho está correto
/caminho/completo/para/nexs-mcp/bin/nexs-mcp -h
```

**Soluções comuns:**

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "/absolute/path/to/nexs-mcp/bin/nexs-mcp",
      "args": ["-storage", "file"],
      "env": {
        "NEXS_DATA_DIR": "/absolute/path/to/data"
      }
    }
  }
}
```

### Ferramentas não aparecem

**Diagnóstico:**
```bash
# Testar servidor manualmente
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./bin/nexs-mcp
```

**Resposta esperada:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "tools": [
      {"name": "list_elements", ...},
      {"name": "get_element", ...},
      ...
    ]
  },
  "id": 1
}
```

### Erro ao chamar ferramenta

**Sintomas:**
- Claude retorna erro ao usar ferramenta
- Timeout na execução

**Solução:**
```bash
# 1. Verificar logs
tail -f /tmp/nexs-mcp.log

# 2. Testar ferramenta diretamente
echo '{"jsonrpc":"2.0","method":"tools/call","params":{"name":"list_elements","arguments":{}},"id":1}' | ./bin/nexs-mcp

# 3. Verificar permissões do diretório de dados
ls -la /path/to/data
chmod 755 /path/to/data
```

---

## Problemas de Storage

### Falha ao criar diretório de dados

**Sintoma:**
```
failed to create base directory: permission denied
```

**Solução:**
```bash
# Criar diretório manualmente com permissões corretas
mkdir -p /path/to/data
chmod 755 /path/to/data

# Ou usar diretório no home do usuário
./bin/nexs-mcp -data-dir ~/nexs-data
```

### Dados não persistem

**Diagnóstico:**
```bash
# 1. Verificar se está usando file storage
./bin/nexs-mcp -storage file -data-dir /path/to/data

# 2. Verificar se arquivos são criados
ls -R /path/to/data

# 3. Verificar estrutura esperada
# data/
#   YYYY-MM-DD/
#     persona/
#       *.yaml
```

**Solução:**
```bash
# Se usar storage memory, dados não persistem
# Mudar para file storage:
./bin/nexs-mcp -storage file
```

### Arquivos YAML corrompidos

**Sintoma:**
```
failed to unmarshal file: yaml: unmarshal errors
```

**Solução:**
```bash
# 1. Validar arquivo YAML
cat /path/to/file.yaml

# 2. Se corrompido, remover
rm /path/to/corrupted.yaml

# 3. Recriar elemento via ferramenta create_element
```

---

## Performance

### Server lento

**Diagnóstico:**
```bash
# 1. Verificar número de elementos
find /path/to/data -name "*.yaml" | wc -l

# 2. Monitorar recursos
top -p $(pgrep nexs-mcp)
```

**Otimizações:**

1. **Usar storage em memória para testes:**
```bash
./bin/nexs-mcp -storage memory
```

2. **Limpar dados antigos:**
```bash
# Remover elementos inativos
find /path/to/data -name "*.yaml" -mtime +30 -delete
```

3. **Aumentar limite de arquivos abertos:**
```bash
ulimit -n 4096
```

### Alto uso de memória

**Diagnóstico:**
```bash
# Verificar uso de memória
ps aux | grep nexs-mcp

# Profile de memória
go tool pprof http://localhost:6060/debug/pprof/heap
```

**Solução:**
- Reiniciar servidor periodicamente
- Usar paginação ao listar elementos
- Limitar número de elementos no storage

---

## Logs e Debug

### Habilitar logs detalhados

**Desenvolvimento:**
```bash
# Redirecionar stderr para arquivo
./bin/nexs-mcp 2> debug.log

# Com Claude Desktop
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "/path/to/nexs-mcp",
      "stderr": "/tmp/nexs-mcp-debug.log"
    }
  }
}
```

### Analisar logs

```bash
# Monitorar em tempo real
tail -f /tmp/nexs-mcp-debug.log

# Buscar erros
grep -i "error" /tmp/nexs-mcp-debug.log

# Últimas 100 linhas
tail -n 100 /tmp/nexs-mcp-debug.log
```

### Debug com delve

```bash
# Instalar delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debugar
dlv debug ./cmd/nexs-mcp
```

---

## Problemas Conhecidos

### Issue #1: YAML com caracteres especiais

**Problema:** Elementos com caracteres Unicode podem não serializar corretamente.

**Workaround:** Usar somente ASCII nos nomes até fix.

**Status:** Planejado para v0.2.0

### Issue #2: Limite de tags

**Problema:** Muitas tags (>100) podem causar lentidão na busca.

**Workaround:** Limitar a 10-20 tags por elemento.

**Status:** Otimização planejada.

---

## Reportar Problemas

Se o problema persistir:

1. Coletar informações:
```bash
# Versão
./bin/nexs-mcp -version

# Sistema operacional
uname -a

# Go version
go version

# Logs relevantes
tail -n 100 /tmp/nexs-mcp-debug.log
```

2. Criar issue no GitHub:
- https://github.com/fsvxavier/nexs-mcp/issues
- Incluir informações acima
- Descrever passos para reproduzir

---

## Recursos Adicionais

- [FAQ](FAQ.md)
- [Tools Reference](TOOLS.md)
- [Architecture](plano/ARCHITECTURE.md)
- [GitHub Issues](https://github.com/fsvxavier/nexs-mcp/issues)
