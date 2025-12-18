# NEXS MCP - Examples

Este diretório contém exemplos práticos de uso do NEXS MCP Server.

## 📁 Estrutos

- `basic/` - Exemplos básicos de uso das ferramentas
- `integration/` - Exemplos de integração com Claude Desktop
- `scripts/` - Scripts utilitários para testes

## 🚀 Quick Start

### 1. Executar o servidor

```bash
# Com file storage (padrão)
./bin/nexs-mcp

# Com storage em memória
./bin/nexs-mcp -storage memory

# Com diretório customizado
./bin/nexs-mcp -data-dir /caminho/para/dados
```

### 2. Testar com stdio

```bash
# Enviar comando initialize
echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":true}},"id":1}' | ./bin/nexs-mcp

# Listar ferramentas disponíveis
echo '{"jsonrpc":"2.0","method":"tools/list","id":2}' | ./bin/nexs-mcp
```

## 📝 Exemplos por Categoria

### Gerenciamento de Elementos

- [create_element.sh](basic/create_element.sh) - Criar elementos
- [list_elements.sh](basic/list_elements.sh) - Listar e filtrar elementos
- [update_element.sh](basic/update_element.sh) - Atualizar elementos
- [delete_element.sh](basic/delete_element.sh) - Remover elementos

### Integração

- [claude_desktop_setup.md](integration/claude_desktop_setup.md) - Configurar Claude Desktop
- [test_integration.sh](integration/test_integration.sh) - Testar integração

## 🔧 Variáveis de Ambiente

```bash
# Tipo de storage (memory ou file)
export NEXS_STORAGE_TYPE=file

# Diretório de dados
export NEXS_DATA_DIR=./data/elements

# Nome do servidor
export NEXS_SERVER_NAME=nexs-mcp
```

## 📚 Documentação Adicional

- [Tools Reference](../docs/TOOLS.md) - Referência completa das ferramentas
- [API Examples](../docs/API_EXAMPLES.md) - Exemplos de chamadas API
- [Troubleshooting](../docs/TROUBLESHOOTING.md) - Resolução de problemas
