#!/bin/bash
# Script para executar testes de integração e benchmarks dos modelos NLP ONNX

set -e

echo "╔═══════════════════════════════════════════════════════════════════════╗"
echo "║       TESTES DE INTEGRAÇÃO E BENCHMARKS - MODELOS NLP ONNX           ║"
echo "╚═══════════════════════════════════════════════════════════════════════╝"
echo

# Verificar se os modelos existem
ENTITY_MODEL="models/bert-base-ner/model.onnx"
SENTIMENT_MODEL="models/distilbert-sentiment/model.onnx"

if [ ! -f "$ENTITY_MODEL" ]; then
    echo "❌ Modelo BERT NER não encontrado em $ENTITY_MODEL"
    echo "   Execute: python3 scripts/download_nlp_models.py"
    exit 1
fi

if [ ! -f "$SENTIMENT_MODEL" ]; then
    echo "❌ Modelo DistilBERT Sentiment não encontrado em $SENTIMENT_MODEL"
    echo "   Execute: python3 scripts/download_nlp_models.py"
    exit 1
fi

echo "✅ Modelos NLP encontrados:"
echo "   • BERT NER: $(du -h $ENTITY_MODEL | cut -f1)"
echo "   • DistilBERT Sentiment: $(du -h $SENTIMENT_MODEL | cut -f1)"
echo

# Função para executar comandos com feedback
run_test() {
    local title=$1
    local cmd=$2

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📋 $title"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo

    if eval $cmd; then
        echo
        echo "✅ $title - PASSOU"
    else
        echo
        echo "❌ $title - FALHOU"
        return 1
    fi
    echo
}

# Opções de execução
MODE=${1:-all}

case $MODE in
    "test")
        echo "🧪 Executando apenas testes de integração..."
        echo

        run_test "Testes de Integração ONNX" \
            "cd internal/application && go test -tags integration -v -run TestONNXBERTProvider_Integration_RealModels -timeout 5m"

        run_test "Testes de Integração - Sentiment Analyzer" \
            "cd internal/application && go test -tags integration -v -run TestSentimentAnalyzer_Integration_RealModels -timeout 5m"

        run_test "Testes de Integração - Entity Extractor" \
            "cd internal/application && go test -tags integration -v -run TestEntityExtractor_Integration_RealModels -timeout 5m"
        ;;

    "bench")
        echo "⚡ Executando apenas benchmarks..."
        echo

        run_test "Benchmark - Entity Extraction" \
            "cd internal/application && go test -tags integration -bench=BenchmarkONNXBERTProvider_ExtractEntities$ -benchmem -benchtime=10x"

        run_test "Benchmark - Sentiment Analysis" \
            "cd internal/application && go test -tags integration -bench=BenchmarkONNXBERTProvider_AnalyzeSentiment$ -benchmem -benchtime=10x"

        run_test "Benchmark - Batch Operations" \
            "cd internal/application && go test -tags integration -bench=Batch -benchmem -benchtime=5x"
        ;;

    "bench-all")
        echo "⚡ Executando todos os benchmarks..."
        echo

        run_test "Todos os Benchmarks" \
            "cd internal/application && go test -tags integration -bench=BenchmarkONNX -benchmem -benchtime=5x"
        ;;

    "quick")
        echo "🚀 Execução rápida (1 teste + 1 benchmark)..."
        echo

        run_test "Teste Rápido - Sentiment Analysis" \
            "cd internal/application && go test -tags integration -v -run TestONNXBERTProvider_Integration_RealModels/AnalyzeSentiment_RealDistilBERT_Positive -timeout 2m"

        run_test "Benchmark Rápido - Sentiment" \
            "cd internal/application && go test -tags integration -bench=BenchmarkONNXBERTProvider_AnalyzeSentiment$ -benchmem -benchtime=3x"
        ;;

    "all"|*)
        echo "🔬 Execução completa (testes + benchmarks)..."
        echo

        # Testes
        echo "═══════════════════════════════════════════════════════════════════════"
        echo "  PARTE 1: TESTES DE INTEGRAÇÃO"
        echo "═══════════════════════════════════════════════════════════════════════"
        echo

        run_test "Testes ONNX Provider" \
            "cd internal/application && go test -tags integration -v -run TestONNXBERTProvider_Integration -timeout 5m"

        run_test "Testes Sentiment Analyzer" \
            "cd internal/application && go test -tags integration -v -run TestSentimentAnalyzer_Integration -timeout 5m"

        run_test "Testes Entity Extractor" \
            "cd internal/application && go test -tags integration -v -run TestEntityExtractor_Integration -timeout 5m"

        # Benchmarks
        echo "═══════════════════════════════════════════════════════════════════════"
        echo "  PARTE 2: BENCHMARKS DE PERFORMANCE"
        echo "═══════════════════════════════════════════════════════════════════════"
        echo

        run_test "Benchmark - Entity Extraction (curto)" \
            "cd internal/application && go test -tags integration -bench=BenchmarkONNXBERTProvider_ExtractEntities_Short -benchmem -benchtime=5x"

        run_test "Benchmark - Entity Extraction (médio)" \
            "cd internal/application && go test -tags integration -bench=BenchmarkONNXBERTProvider_ExtractEntities$ -benchmem -benchtime=5x"

        run_test "Benchmark - Sentiment Analysis (curto)" \
            "cd internal/application && go test -tags integration -bench=BenchmarkONNXBERTProvider_AnalyzeSentiment_Short -benchmem -benchtime=5x"

        run_test "Benchmark - Sentiment Analysis (médio)" \
            "cd internal/application && go test -tags integration -bench=BenchmarkONNXBERTProvider_AnalyzeSentiment$ -benchmem -benchtime=5x"

        run_test "Benchmark - Combined Operations" \
            "cd internal/application && go test -tags integration -bench=BenchmarkONNXBERTProvider_Combined -benchmem -benchtime=3x"
        ;;
esac

echo "╔═══════════════════════════════════════════════════════════════════════╗"
echo "║                    TESTES CONCLUÍDOS COM SUCESSO                      ║"
echo "╚═══════════════════════════════════════════════════════════════════════╝"
echo
echo "📊 Resumo dos Testes:"
echo
echo "✅ Testes de Integração:"
echo "   • ONNX Provider com modelos reais (BERT + DistilBERT)"
echo "   • Entity Extraction com relacionamentos"
echo "   • Sentiment Analysis (positivo, negativo, neutro)"
echo "   • Batch operations"
echo
echo "⚡ Benchmarks de Performance:"
echo "   • Entity extraction: curto, médio, longo"
echo "   • Sentiment analysis: curto, médio, longo"
echo "   • Operações batch"
echo "   • Operações combinadas"
echo
echo "💡 Dicas:"
echo "   • Executar apenas testes:     $0 test"
echo "   • Executar apenas benchmarks: $0 bench"
echo "   • Executar todos benchmarks:  $0 bench-all"
echo "   • Execução rápida:            $0 quick"
echo
echo "📚 Arquivos de teste:"
echo "   • internal/application/onnx_integration_test.go"
echo "   • internal/application/onnx_benchmark_test.go"
echo
