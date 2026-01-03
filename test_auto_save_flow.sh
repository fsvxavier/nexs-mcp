#!/bin/bash
# Script para testar o fluxo completo de auto-save
# Demonstra como usar o auto-save corretamente

echo "=== Teste de Auto-Save Flow ==="
echo ""
echo "IMPORTANTE: O auto-save funciona da seguinte forma:"
echo ""
echo "1. Cliente MCP deve chamar 'set_user_context' para definir o usuário:"
echo "   Exemplo: set_user_context({username: 'john_doe', metadata: {}})"
echo ""
echo "2. Durante a conversa, working memories são criadas automaticamente"
echo "   quando você usa outras ferramentas (create_memory, etc)"
echo ""
echo "3. O worker auto-save roda a cada 5 minutos (NEXS_AUTO_SAVE_INTERVAL)"
echo "   e salva as working memories como Memory elements"
echo ""
echo "4. Configuração necessária no .env:"
echo "   NEXS_AUTO_SAVE_MEMORIES=true"
echo "   NEXS_AUTO_SAVE_INTERVAL=5m"
echo ""
echo "=== Verificando configuração atual ==="
echo ""

if [ -f .env ]; then
    echo "AUTO_SAVE_MEMORIES: $(grep NEXS_AUTO_SAVE_MEMORIES .env || echo 'não configurado')"
    echo "AUTO_SAVE_INTERVAL: $(grep NEXS_AUTO_SAVE_INTERVAL .env || echo 'não configurado')"
else
    echo "Arquivo .env não encontrado!"
fi

echo ""
echo "=== Estado da pasta memory ==="
ls -lh .nexs-mcp/elements/memory/ 2>/dev/null || echo "Pasta memory vazia ou não existe"

echo ""
echo "=== Para testar manualmente ==="
echo "1. Inicie o servidor: ./bin/nexs-mcp"
echo "2. Use um cliente MCP (Claude Desktop, Cline, etc)"
echo "3. Chame a ferramenta 'set_user_context' primeiro"
echo "4. Use outras ferramentas para criar contexto"
echo "5. Aguarde 5 minutos (ou o intervalo configurado)"
echo "6. Verifique .nexs-mcp/elements/memory/"
echo ""
echo "=== Alternativa: Usar save_conversation_context manualmente ==="
echo "Você pode chamar 'save_conversation_context' a qualquer momento"
echo "para salvar o contexto sem esperar o auto-save"
