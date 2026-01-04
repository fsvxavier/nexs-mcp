# Auto-Save de Memórias - Guia de Uso

## ⚠️ Problema Comum: Pasta `memory/` Vazia

Se você configurou `NEXS_AUTO_SAVE_MEMORIES=true` mas a pasta `.nexs-mcp/elements/memory/` continua vazia, aqui está o porquê:

## Como o Auto-Save Funciona

O auto-save **NÃO** salva automaticamente toda a conversa. Ele funciona em 3 etapas:

### 1️⃣ Definir Contexto do Usuário

**PRIMEIRO PASSO OBRIGATÓRIO**: O cliente MCP deve chamar a ferramenta `set_user_context`:

```json
{
  "tool": "set_user_context",
  "arguments": {
    "username": "seu_usuario",
    "metadata": {
      "project": "meu_projeto",
      "session": "2026-01-02"
    }
  }
}
```

**Sem isso, o auto-save não funciona!** Você verá nos logs:
```
Auto-save skipped: No user context set. Use 'set_user_context' tool to define user.
```

### 2️⃣ Criar Working Memories

Durante a sessão, working memories precisam ser criadas. Isso acontece:

**Opção A - Automaticamente** quando você usa ferramentas MCP:
- `create_memory`
- `create_persona`
- `create_skill`
- Outras ferramentas que manipulam elementos

**Opção B - Manualmente** chamando `working_memory_add`:

```json
{
  "tool": "working_memory_add",
  "arguments": {
    "session_id": "auto-save-seu_usuario",
    "content": "Contexto importante da conversa",
    "priority": "high",
    "tags": ["importante", "projeto-x"]
  }
}
```

### 3️⃣ Worker Auto-Save

O worker roda automaticamente a cada intervalo configurado:

```bash
# .env
NEXS_AUTO_SAVE_MEMORIES=true
NEXS_AUTO_SAVE_INTERVAL=5m  # Padrão: 5 minutos
```

O worker:
1. ✅ Verifica se há um usuário definido
2. ✅ Lista working memories da sessão `auto-save-{usuario}`
3. ✅ Aplica otimizações (deduplication, compression, summarization)
4. ✅ Salva como Memory element em `.nexs-mcp/elements/memory/{data}/`
5. ✅ Limpa working memories após salvar

## Logs Informativos

### ✅ Funcionando Corretamente
```json
{"level":"INFO","msg":"Auto-save worker started","interval":"5m"}
{"level":"INFO","msg":"User context updated","user":"john_doe"}
{"level":"INFO","msg":"Performing auto-save of conversation context","user":"john_doe","memory_count":3}
{"level":"INFO","msg":"Successfully auto-saved conversation context","memory_id":"mem_123","user":"john_doe"}
```

### ❌ Problemas Comuns

**Sem usuário definido:**
```json
{"level":"DEBUG","msg":"Auto-save skipped: No user context set. Use 'set_user_context' tool to define user."}
```
**Solução**: Chame `set_user_context` primeiro

**Sem working memories:**
```json
{"level":"DEBUG","msg":"Auto-save skipped: No working memories to save","user":"john","tip":"Working memories are created when you use MCP tools"}
```
**Solução**: Use ferramentas MCP ou adicione memórias manualmente

## Alternativa: Save Manual

Se você não quer esperar o auto-save, pode salvar manualmente:

```json
{
  "tool": "save_conversation_context",
  "arguments": {
    "context": "Contexto completo da conversa até agora...",
    "summary": "Resumo da conversa",
    "tags": ["manual-save", "importante"],
    "importance": "high"
  }
}
```

## Fluxo Completo de Exemplo

### Via Cliente MCP (Claude Desktop, Cline, etc)

```javascript
// 1. Definir usuário
await mcp.callTool("set_user_context", {
  username: "john_doe",
  metadata: {project: "nexs-integration"}
});

// 2. Usar ferramentas normalmente (cria working memories automaticamente)
await mcp.callTool("create_persona", {
  name: "Expert Developer",
  description: "Senior developer persona"
});

// 3. Ou adicionar working memory manualmente
await mcp.callTool("working_memory_add", {
  session_id: "auto-save-john_doe",
  content: "Important conversation context",
  priority: "high"
});

// 4. Aguardar auto-save (5 min) ou salvar manualmente
await mcp.callTool("save_conversation_context", {
  context: "Full conversation...",
  summary: "Session summary"
});

// 5. Verificar resultado
await mcp.callTool("list_elements", {
  type: "memory",
  limit: 10
});
```

## Verificar Se Está Funcionando

```bash
# Ver logs do servidor
tail -f logs/nexs-mcp.log

# Ver working memories
ls -lh .nexs-mcp/working_memory/

# Ver memories salvas
ls -lh .nexs-mcp/elements/memory/$(date +%Y-%m-%d)/

# Verificar configuração
grep AUTO_SAVE .env
```

## Arquitetura

```
Cliente MCP
    ↓
set_user_context → globalUserSession
    ↓
Uso de ferramentas → WorkingMemory.Add()
    ↓
Auto-Save Worker (5min) → performAutoSave()
    ↓
    ├─ Deduplicate memories
    ├─ Optimize context window
    ├─ Summarize old content
    ├─ Compress prompts
    └─ Save as Memory element
    ↓
.nexs-mcp/elements/memory/YYYY-MM-DD/
```

## Otimizações Aplicadas

Durante o auto-save, estas otimizações são aplicadas automaticamente:

1. **Semantic Deduplication** (92%+ similaridade)
   - Remove memórias duplicadas
   - Economiza tokens e espaço

2. **Context Window Management**
   - Trunca contexto grande (estratégia híbrida)
   - Preserva conteúdo mais relevante

3. **TF-IDF Summarization**
   - Sumariza conteúdo antigo (>1000 chars)
   - Redução de ~70% no tamanho

4. **Prompt Compression**
   - Remove redundâncias sintáticas
   - Compressão de ~35%

## Referências

- Código: `internal/mcp/server.go` (linha 934-1208)
- Testes: `test_auto_save_integration.sh`
- Documento de análise: `ANALISE_AUTO_SAVE_E_OTIMIZACAO.md`
