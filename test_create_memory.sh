#!/bin/bash
# Script para testar create_memory e auto-save via MCP stdio

echo "=== Teste de create_memory e Auto-Save ==="
echo ""

# Build se necessário
if [ ! -f bin/nexs-mcp ]; then
    echo "Compilando nexs-mcp..."
    go build -o bin/nexs-mcp ./cmd/nexs-mcp
fi

# Criar arquivo de teste MCP
cat > /tmp/test_mcp_flow.json <<'EOF'
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0","capabilities":{}}}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"set_user_context","arguments":{"username":"test_user","metadata":{"session":"auto-save-test"}}}}
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"working_memory_add","arguments":{"session_id":"auto-save-test_user","content":"Esta é uma working memory de teste para o auto-save","priority":"high","tags":["teste","auto-save"]}}}
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"create_memory","arguments":{"name":"Teste Memory Manual","content":"Conteúdo de teste criado via create_memory","description":"Memory de teste para validar o fluxo","version":"1.0.0","author":"test_user","tags":["teste","manual"]}}}
{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"list_elements","arguments":{"type":"memory","limit":5}}}
EOF

echo "1. Iniciando servidor MCP com teste de ferramentas..."
echo ""

# Executar comandos MCP via stdin
timeout 10s bin/nexs-mcp < /tmp/test_mcp_flow.json > /tmp/test_mcp_output.txt 2>&1 || true

echo "2. Resultado da execução:"
echo ""

# Mostrar output relevante
if [ -f /tmp/test_mcp_output.txt ]; then
    echo "--- Logs do servidor ---"
    grep -E "(set_user_context|working_memory|create_memory|Auto-save)" /tmp/test_mcp_output.txt | tail -20 || echo "Nenhum log relevante encontrado"
    echo ""
fi

echo "3. Verificando elementos criados:"
echo ""

# Verificar memories criadas
if [ -d .nexs-mcp/elements/memory ]; then
    echo "Memories encontradas:"
    find .nexs-mcp/elements/memory -name "*.yaml" -type f -exec basename {} \; 2>/dev/null | head -5

    if [ $(find .nexs-mcp/elements/memory -name "*.yaml" -type f 2>/dev/null | wc -l) -eq 0 ]; then
        echo "  ⚠️  Nenhuma memory salva ainda"
        echo "  Motivo possível:"
        echo "    - Auto-save ainda não rodou (intervalo: $(grep NEXS_AUTO_SAVE_INTERVAL .env | cut -d= -f2))"
        echo "    - Working memories não foram criadas"
        echo "    - Usuário não foi definido"
    fi
else
    echo "  ⚠️  Pasta memory/ não existe ainda"
fi

echo ""
echo "4. Verificando working memories:"
echo ""

if [ -d .nexs-mcp/working_memory ]; then
    echo "Working memories encontradas:"
    ls -lh .nexs-mcp/working_memory/ 2>/dev/null | tail -5
else
    echo "  ⚠️  Pasta working_memory/ não existe"
fi

echo ""
echo "=== Como usar via cliente MCP real ==="
echo ""
echo "1. Configure Claude Desktop ou Cline com nexs-mcp"
echo ""
echo "2. Execute estas ferramentas em sequência:"
echo "   a) set_user_context:"
echo '      {"username": "seu_usuario", "metadata": {}}'
echo ""
echo "   b) working_memory_add:"
echo '      {"session_id": "auto-save-seu_usuario", "content": "Contexto importante"}'
echo ""
echo "   c) create_memory (cria memory permanente imediatamente):"
echo '      {"name": "Minha Memory", "content": "Conteúdo", "author": "seu_usuario"}'
echo ""
echo "3. Aguarde auto-save ($(grep NEXS_AUTO_SAVE_INTERVAL .env | cut -d= -f2)) ou use save_conversation_context"
echo ""

# Cleanup
rm -f /tmp/test_mcp_flow.json /tmp/test_mcp_output.txt
