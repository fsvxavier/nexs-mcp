# Claude Desktop Integration Guide

Este guia mostra como integrar o NEXS MCP Server com o Claude Desktop.

## 📋 Pré-requisitos

- Claude Desktop instalado
- NEXS MCP Server compilado (`make build`)

## 🔧 Configuração

### 1. Localizar o arquivo de configuração

O arquivo de configuração do Claude Desktop varia por sistema operacional:

**macOS:**
```bash
~/Library/Application Support/Claude/claude_desktop_config.json
```

**Linux:**
```bash
~/.config/Claude/claude_desktop_config.json
```

**Windows:**
```powershell
%APPDATA%\Claude\claude_desktop_config.json
```

### 2. Adicionar NEXS MCP ao config

Edite o arquivo `claude_desktop_config.json` e adicione:

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "/caminho/completo/para/nexs-mcp/bin/nexs-mcp",
      "args": ["-storage", "file"],
      "env": {
        "NEXS_DATA_DIR": "/caminho/para/dados"
      }
    }
  }
}
```

**Exemplo completo:**

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "/home/user/nexs-mcp/bin/nexs-mcp",
      "args": ["-storage", "file", "-data-dir", "/home/user/.nexs/data"],
      "env": {
        "NEXS_STORAGE_TYPE": "file"
      }
    }
  }
}
```

### 3. Reiniciar Claude Desktop

Feche completamente o Claude Desktop e reabra.

## ✅ Verificar Integração

No Claude Desktop, você pode testar se o servidor está funcionando:

```
Você pode listar os elementos disponíveis no NEXS MCP?
```

Claude deve responder usando a ferramenta `list_elements`.

## 🛠️ Ferramentas Disponíveis

Após a integração, Claude terá acesso a:

1. **list_elements** - Listar elementos com filtros
2. **get_element** - Obter elemento por ID
3. **create_element** - Criar novo elemento
4. **update_element** - Atualizar elemento
5. **delete_element** - Remover elemento

## 📝 Exemplos de Uso com Claude

### Criar uma Persona

```
Crie uma persona chamada "Data Scientist Expert" especializada em machine learning e análise de dados.
```

### Listar Personas

```
Liste todas as personas disponíveis.
```

### Atualizar um Elemento

```
Atualize a persona "Data Scientist Expert" adicionando a tag "python".
```

### Buscar por Tags

```
Mostre todos os elementos com a tag "engineer".
```

## 🐛 Troubleshooting

### Servidor não aparece no Claude

1. Verifique se o caminho do binário está correto
2. Certifique-se que o arquivo tem permissão de execução:
   ```bash
   chmod +x /caminho/para/nexs-mcp/bin/nexs-mcp
   ```
3. Teste o servidor manualmente:
   ```bash
   echo '{"jsonrpc":"2.0","method":"initialize","id":1}' | /caminho/para/nexs-mcp/bin/nexs-mcp
   ```

### Erros de permissão no diretório de dados

Certifique-se que o diretório de dados existe e tem permissão de escrita:

```bash
mkdir -p /caminho/para/dados
chmod 755 /caminho/para/dados
```

### Verificar logs

Em desenvolvimento, você pode redirecionar logs:

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "/caminho/para/nexs-mcp/bin/nexs-mcp",
      "args": ["-storage", "file"],
      "stderr": "/tmp/nexs-mcp.log"
    }
  }
}
```

## 🔒 Segurança

- Use caminhos absolutos para o binário
- Mantenha permissões adequadas no diretório de dados (755)
- Em produção, considere usar storage em arquivo para persistência

## 📚 Recursos Adicionais

- [MCP Specification](https://modelcontextprotocol.io/)
- [Claude Desktop Documentation](https://www.anthropic.com/claude)
- [NEXS MCP README](../../README.md)
