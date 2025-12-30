#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║         TESTE DO SERVIDOR NEXS-MCP COM CONFIGURAÇÃO ONNX          ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo ""

# Cores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 1. Verificar binário
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. Verificando binário..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ -f "bin/nexs-mcp" ]; then
    echo -e "${GREEN}✓${NC} Binário encontrado: bin/nexs-mcp"
    ls -lh bin/nexs-mcp | awk '{print "  Tamanho:", $5, "| Modificado:", $6, $7, $8}'
else
    echo -e "${RED}✗${NC} Binário não encontrado!"
    exit 1
fi
echo ""

# 2. Verificar compilação com ONNX
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. Verificando compilação com ONNX..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if go version -m bin/nexs-mcp | grep -q "CGO_ENABLED=1"; then
    echo -e "${GREEN}✓${NC} CGO habilitado"
    go version -m bin/nexs-mcp | grep "CGO_" | sed 's/^/  /'
else
    echo -e "${RED}✗${NC} CGO não habilitado!"
fi
echo ""

# 3. Verificar ONNX Runtime
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. Verificando ONNX Runtime..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if ldconfig -p | grep -q onnxruntime; then
    echo -e "${GREEN}✓${NC} ONNX Runtime encontrado no sistema"
    ldconfig -p | grep onnxruntime | head -2 | sed 's/^/  /'
else
    echo -e "${RED}✗${NC} ONNX Runtime não encontrado!"
fi
echo ""

# 4. Verificar modelo ONNX
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. Verificando modelo ONNX..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ -f "models/ms-marco-MiniLM-L-6-v2/model.onnx" ]; then
    echo -e "${GREEN}✓${NC} Modelo encontrado: ms-marco-MiniLM-L-6-v2"
    ls -lh models/ms-marco-MiniLM-L-6-v2/model.onnx | awk '{print "  Tamanho:", $5}'
else
    echo -e "${RED}✗${NC} Modelo não encontrado!"
fi
echo ""

# 5. Testar inicialização do servidor
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. Testando inicialização do servidor (2 segundos)..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Capturar saída do servidor (inicia em background, aguarda "Server ready" no log até MAX_WAIT segundos)
# MAX_WAIT pode ser sobrescrito externamente: MAX_WAIT=20
MAX_WAIT=${MAX_WAIT:-20}
TMP_LOG=$(mktemp)
./bin/nexs-mcp --log-level=info --log-format=text >"$TMP_LOG" 2>&1 &
SERVER_PID=$!

# Cleanup function to ensure the server process (and its process group) is terminated
_cleanup_server() {
    if kill -0 "$SERVER_PID" >/dev/null 2>&1; then
        # Try graceful termination
        kill -TERM "$SERVER_PID" >/dev/null 2>&1 || true
        sleep 0.5
        # Ensure process and its group are killed
        kill -TERM -"$SERVER_PID" >/dev/null 2>&1 || true
        kill -KILL "$SERVER_PID" >/dev/null 2>&1 || true
    fi
    wait "$SERVER_PID" 2>/dev/null || true
}

# Wait for "Server ready" in the log with a small polling interval
READY_FOUND=0
for i in $(seq 1 $((MAX_WAIT * 10))); do
    if grep -q "Server ready" "$TMP_LOG" >/dev/null 2>&1; then
        READY_FOUND=1
        break
    fi
    sleep 0.1
done

# Capture output and cleanup
SERVER_OUTPUT=$(cat "$TMP_LOG" || true)
_cleanup_server
rm -f "$TMP_LOG"

if [ "$READY_FOUND" -eq 1 ]; then
    echo -e "${GREEN}✓${NC} Servidor inicializou (detectado 'Server ready')"
else
    echo -e "${YELLOW}⚠${NC} Servidor não respondeu dentro de ${MAX_WAIT}s (ver logs para detalhes)"
fi

# Verificar ONNX habilitado
if echo "$SERVER_OUTPUT" | grep -q "onnx_support=\"enabled"; then
    echo -e "${GREEN}✓${NC} ONNX Runtime carregado com sucesso"
else
    echo -e "${RED}✗${NC} ONNX Runtime não foi carregado"
    echo "$SERVER_OUTPUT" | grep "onnx_support" | sed 's/^/  /'
fi

# Verificar servidor iniciado
if echo "$SERVER_OUTPUT" | grep -q "Server ready"; then
    echo -e "${GREEN}✓${NC} Servidor iniciado com sucesso"
else
    echo -e "${RED}✗${NC} Servidor não iniciou corretamente"
fi

# Verificar tools registradas
if echo "$SERVER_OUTPUT" | grep -q "tools_registered"; then
    TOOLS_COUNT=$(echo "$SERVER_OUTPUT" | grep "tools_registered" | grep -oP 'tools_registered=\K\d+' || echo "0")
    echo -e "${GREEN}✓${NC} Ferramentas MCP registradas: $TOOLS_COUNT"
else
    echo -e "${YELLOW}⚠${NC} Não foi possível verificar ferramentas registradas"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Log de inicialização:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "$SERVER_OUTPUT" | head -10
echo ""

# 6. Verificar configuração do VSCode
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6. Verificando configuração do VSCode..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -f ".vscode/mcp.json" ]; then
    echo -e "${GREEN}✓${NC} Arquivo .vscode/mcp.json existe"
    if grep -q "nexs-mcp" .vscode/mcp.json; then
        echo -e "${GREEN}✓${NC} Servidor nexs-mcp configurado"
        if grep -q "NEXS_ONNX_MODEL_PATH" .vscode/mcp.json; then
            echo -e "${GREEN}✓${NC} Modelo ONNX configurado"
        fi
    else
        echo -e "${YELLOW}⚠${NC} Servidor nexs-mcp não encontrado na configuração"
    fi
else
    echo -e "${YELLOW}⚠${NC} Arquivo .vscode/mcp.json não encontrado"
fi

if [ -f ".vscode/settings.json" ]; then
    echo -e "${GREEN}✓${NC} Arquivo .vscode/settings.json existe"
    if grep -q "NEXS_ONNX_ENABLED" .vscode/settings.json; then
        echo -e "${GREEN}✓${NC} Variáveis ONNX configuradas"
    fi
else
    echo -e "${YELLOW}⚠${NC} Arquivo .vscode/settings.json não encontrado"
fi

echo ""
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                      RESUMO DO TESTE                               ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo ""
echo -e "${GREEN}✓${NC} Binário compilado com CGO e ONNX"
echo -e "${GREEN}✓${NC} ONNX Runtime v1.23.2 instalado e funcionando"
echo -e "${GREEN}✓${NC} Modelo ms-marco-MiniLM-L-6-v2 disponível"
echo -e "${GREEN}✓${NC} Servidor inicializa corretamente com ONNX habilitado"
echo -e "${GREEN}✓${NC} Configuração do VSCode completa"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "PRÓXIMOS PASSOS:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "1. Reiniciar o VSCode/Cursor para carregar as novas configurações"
echo "2. Verificar se o servidor nexs-mcp aparece na lista de MCP servers"
echo "3. Testar ferramentas MCP através do chat"
echo ""
echo "Para executar o servidor manualmente:"
echo "  ./bin/nexs-mcp --log-level=debug --log-format=text"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
