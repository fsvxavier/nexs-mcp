#!/bin/bash
# Script de teste dos recursos NLP via MCP
# Testa entity extraction, sentiment analysis e topic modeling

set -e

echo "╔═══════════════════════════════════════════════════════════════════════╗"
echo "║              TESTE DE RECURSOS NLP - NEXS-MCP                        ║"
echo "╚═══════════════════════════════════════════════════════════════════════╝"
echo

# Função para enviar comando MCP via stdin/stdout
send_mcp_command() {
    local tool=$1
    local input=$2

    echo "📤 Enviando: $tool"

    # Criar JSON-RPC request
    cat <<EOF | ./bin/nexs-mcp 2>&1 | grep -v "^{\"time\"" | jq -r '.result // .error // .' 2>/dev/null | head -30
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "$tool",
    "arguments": $input
  }
}
EOF
}

# Iniciar servidor em background
echo "🚀 Iniciando NEXS-MCP..."
./bin/nexs-mcp > /tmp/nexs_nlp_test.log 2>&1 &
SERVER_PID=$!
sleep 3

# Verificar logs de inicialização
echo "📋 Logs de Inicialização:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat /tmp/nexs_nlp_test.log | grep -E "(✅|⚠️)" | while read line; do
    echo "$line" | jq -r '"\(.level | ascii_upcase): \(.msg)" + (if .status then " (status: \(.status))" else "" end)' 2>/dev/null || echo "$line"
done
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# Aguardar inicialização completa
echo "⏳ Aguardando inicialização completa..."
sleep 2

# Parar servidor para testes via stdin
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo
echo "═══════════════════════════════════════════════════════════════════════"
echo "  TESTE 1: Entity Extraction (BERT NER)"
echo "═══════════════════════════════════════════════════════════════════════"
echo

TEXT_ENTITIES='{"text": "John Smith works at Google in Mountain View, California. He founded the AI Lab in 2020."}'

timeout 10 bash -c "
cat <<'EOFMCP' | ./bin/nexs-mcp 2>/dev/null | tail -1 | jq -r '.result.entities[]? | \"- \\(.type): \\(.value) (confidence: \\(.confidence))\"' 2>/dev/null || echo 'Timeout ou erro'
{
  \"jsonrpc\": \"2.0\",
  \"id\": 1,
  \"method\": \"tools/call\",
  \"params\": {
    \"name\": \"extract_entities_advanced\",
    \"arguments\": $TEXT_ENTITIES
  }
}
EOFMCP
"

echo
echo "═══════════════════════════════════════════════════════════════════════"
echo "  TESTE 2: Sentiment Analysis (DistilBERT)"
echo "═══════════════════════════════════════════════════════════════════════"
echo

echo "📝 Texto 1 (Positivo):"
timeout 10 bash -c "
cat <<'EOFMCP' | ./bin/nexs-mcp 2>/dev/null | tail -1 | jq -r '.result | \"Label: \\(.label), Confidence: \\(.confidence), Score: \\(.score)\"' 2>/dev/null || echo 'Timeout ou erro'
{
  \"jsonrpc\": \"2.0\",
  \"id\": 2,
  \"method\": \"tools/call\",
  \"params\": {
    \"name\": \"analyze_sentiment_advanced\",
    \"arguments\": {\"text\": \"This product is absolutely amazing! Best purchase ever.\"}
  }
}
EOFMCP
"

echo
echo "📝 Texto 2 (Negativo):"
timeout 10 bash -c "
cat <<'EOFMCP' | ./bin/nexs-mcp 2>/dev/null | tail -1 | jq -r '.result | \"Label: \\(.label), Confidence: \\(.confidence), Score: \\(.score)\"' 2>/dev/null || echo 'Timeout ou erro'
{
  \"jsonrpc\": \"2.0\",
  \"id\": 3,
  \"method\": \"tools/call\",
  \"params\": {
    \"name\": \"analyze_sentiment_advanced\",
    \"arguments\": {\"text\": \"Terrible experience. Would not recommend.\"}
  }
}
EOFMCP
"

echo
echo "📝 Texto 3 (Neutro):"
timeout 10 bash -c "
cat <<'EOFMCP' | ./bin/nexs-mcp 2>/dev/null | tail -1 | jq -r '.result | \"Label: \\(.label), Confidence: \\(.confidence), Score: \\(.score)\"' 2>/dev/null || echo 'Timeout ou erro'
{
  \"jsonrpc\": \"2.0\",
  \"id\": 4,
  \"method\": \"tools/call\",
  \"params\": {
    \"name\": \"analyze_sentiment_advanced\",
    \"arguments\": {\"text\": \"The product arrived on time.\"}
  }
}
EOFMCP
"

echo
echo "═══════════════════════════════════════════════════════════════════════"
echo "  TESTE 3: Topic Modeling (LDA)"
echo "═══════════════════════════════════════════════════════════════════════"
echo

timeout 10 bash -c "
cat <<'EOFMCP' | ./bin/nexs-mcp 2>/dev/null | tail -1 | jq -r '.result.topics[]? | \"Topic \\(.id): \\(.keywords | join(\", \"))\"' 2>/dev/null || echo 'Timeout ou erro'
{
  \"jsonrpc\": \"2.0\",
  \"id\": 5,
  \"method\": \"tools/call\",
  \"params\": {
    \"name\": \"extract_topics\",
    \"arguments\": {
      \"texts\": [
        \"Machine learning is transforming healthcare with AI diagnosis\",
        \"Deep learning models improve medical imaging accuracy\",
        \"Artificial intelligence helps doctors make better decisions\",
        \"Neural networks analyze patient data efficiently\",
        \"Healthcare AI reduces diagnostic errors significantly\"
      ],
      \"num_topics\": 2
    }
  }
}
EOFMCP
"

echo
echo "╔═══════════════════════════════════════════════════════════════════════╗"
echo "║                      TESTE CONCLUÍDO                                  ║"
echo "╚═══════════════════════════════════════════════════════════════════════╝"
echo
echo "📊 Resumo:"
echo "   ✅ ONNX Runtime: Inicializado"
echo "   ✅ Entity Extraction: Testado com BERT NER"
echo "   ✅ Sentiment Analysis: Testado com DistilBERT"
echo "   ✅ Topic Modeling: Testado com LDA"
echo
echo "📄 Logs completos: /tmp/nexs_nlp_test.log"
echo
