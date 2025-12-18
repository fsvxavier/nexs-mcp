# NEXS MCP - Tools Reference

Documentação completa de todas as ferramentas disponíveis no NEXS MCP Server.

## 📋 Índice

1. [list_elements](#list_elements)
2. [get_element](#get_element)
3. [create_element](#create_element)
4. [update_element](#update_element)
5. [delete_element](#delete_element)

---

## list_elements

Lista elementos com filtros opcionais.

### Parâmetros

| Nome | Tipo | Obrigatório | Descrição |
|------|------|-------------|-----------|
| `type` | string | Não | Filtrar por tipo (persona, skill, template, agent, memory, ensemble) |
| `is_active` | boolean | Não | Filtrar por status ativo |
| `tags` | array[string] | Não | Filtrar por tags (deve conter todas) |
| `limit` | integer | Não | Número máximo de resultados (padrão: 10, máx: 100) |
| `offset` | integer | Não | Número de resultados a pular (padrão: 0) |

### Exemplos

**Listar todos:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "list_elements",
    "arguments": {}
  },
  "id": 1
}
```

**Filtrar por tipo:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "list_elements",
    "arguments": {
      "type": "persona"
    }
  },
  "id": 1
}
```

**Com paginação:**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "list_elements",
    "arguments": {
      "limit": 5,
      "offset": 10
    }
  },
  "id": 1
}
```

### Resposta

```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"elements\":[...],\"count\":5}"
    }]
  },
  "id": 1
}
```

---

## get_element

Obtém um elemento específico por ID.

### Parâmetros

| Nome | Tipo | Obrigatório | Descrição |
|------|------|-------------|-----------|
| `id` | string | Sim | ID do elemento |

### Exemplo

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "get_element",
    "arguments": {
      "id": "persona_Senior_Engineer_20251218-123456"
    }
  },
  "id": 1
}
```

### Resposta

```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"element\":{\"id\":\"...\",\"type\":\"persona\",\"name\":\"...\"}}"
    }]
  },
  "id": 1
}
```

---

## create_element

Cria um novo elemento.

### Parâmetros

| Nome | Tipo | Obrigatório | Descrição |
|------|------|-------------|-----------|
| `type` | string | Sim | Tipo do elemento (persona, skill, template, agent, memory, ensemble) |
| `name` | string | Sim | Nome do elemento (3-100 caracteres) |
| `description` | string | Não | Descrição (máx 500 caracteres) |
| `version` | string | Sim | Versão semver (ex: 1.0.0) |
| `author` | string | Sim | Autor do elemento |
| `tags` | array[string] | Não | Tags para categorização |

### Exemplo

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "create_element",
    "arguments": {
      "type": "persona",
      "name": "Senior Software Engineer",
      "description": "Expert in Go and distributed systems",
      "version": "1.0.0",
      "author": "NEXS Team",
      "tags": ["engineering", "backend", "golang"]
    }
  },
  "id": 1
}
```

### Resposta

```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"id\":\"persona_Senior_Software_Engineer_20251218-123456\",\"element\":{...}}"
    }]
  },
  "id": 1
}
```

---

## update_element

Atualiza um elemento existente.

### Parâmetros

| Nome | Tipo | Obrigatório | Descrição |
|------|------|-------------|-----------|
| `id` | string | Sim | ID do elemento |
| `name` | string | Não | Novo nome |
| `description` | string | Não | Nova descrição |
| `tags` | array[string] | Não | Novas tags (substitui as existentes) |
| `is_active` | boolean | Não | Status ativo |

### Exemplo

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "update_element",
    "arguments": {
      "id": "persona_Senior_Engineer_20251218-123456",
      "description": "Updated description",
      "tags": ["engineering", "backend", "golang", "kubernetes"]
    }
  },
  "id": 1
}
```

### Resposta

```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"id\":\"...\",\"element\":{...}}"
    }]
  },
  "id": 1
}
```

---

## delete_element

Remove um elemento.

### Parâmetros

| Nome | Tipo | Obrigatório | Descrição |
|------|------|-------------|-----------|
| `id` | string | Sim | ID do elemento a remover |

### Exemplo

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "delete_element",
    "arguments": {
      "id": "persona_Senior_Engineer_20251218-123456"
    }
  },
  "id": 1
}
```

### Resposta

```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"id\":\"persona_Senior_Engineer_20251218-123456\",\"deleted\":true}"
    }]
  },
  "id": 1
}
```

---

## Tipos de Elementos

### persona
Representa uma persona de IA com características e comportamentos específicos.

### skill
Uma habilidade ou capacidade que pode ser associada a outros elementos.

### template
Template reutilizável para criação de prompts ou configurações.

### agent
Agente autônomo que combina persona, skills e templates.

### memory
Armazenamento de contexto e histórico.

### ensemble
Conjunto coordenado de múltiplos agentes.

---

## Códigos de Erro JSON-RPC

| Código | Mensagem | Descrição |
|--------|----------|-----------|
| -32700 | Parse error | JSON inválido |
| -32600 | Invalid Request | Requisição malformada |
| -32601 | Method not found | Método não existe |
| -32602 | Invalid params | Parâmetros inválidos |
| -32603 | Internal error | Erro interno do servidor |

---

## Estrutura de Dados

### ElementMetadata

```go
type ElementMetadata struct {
    ID          string                 `json:"id"`
    Type        ElementType            `json:"type"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Version     string                 `json:"version"`
    Author      string                 `json:"author"`
    Tags        []string               `json:"tags"`
    IsActive    bool                   `json:"is_active"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

---

## Exemplos de Scripts

Veja o diretório [examples/](../examples/) para scripts prontos de cada ferramenta.
