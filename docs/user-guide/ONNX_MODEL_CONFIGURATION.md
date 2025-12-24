# Configuração de Modelos ONNX

**Data:** 23 de dezembro de 2025  
**Status:** Produção

---

## 📊 Visão Geral

O NEXS MCP Server utiliza modelos ONNX para avaliação de qualidade de conteúdo. Dois modelos estão disponíveis em produção, cada um otimizado para diferentes casos de uso.

### Modelos Disponíveis

| Modelo | Status | Velocidade | Qualidade | Idiomas | Uso Recomendado |
|--------|--------|------------|-----------|---------|-----------------|
| **MS MARCO** | Default | 61ms | ⭐⭐⭐ | 9 | APIs de baixa latência |
| **Paraphrase-Multilingual** | Configurável | 109ms | ⭐⭐⭐⭐⭐ | 11 | Máxima qualidade/CJK |

---

## 🚀 Uso Básico

### Modelo Padrão (MS MARCO)

**Sem configuração necessária** - O MS MARCO é usado automaticamente:

```go
import "github.com/fsvxavier/nexs-mcp/internal/quality"

// Usa MS MARCO automaticamente
scorer, err := quality.NewONNXScorer(nil)
if err != nil {
    log.Fatal(err)
}
defer scorer.Close()

score, err := scorer.Score(ctx, "Texto para avaliar")
```

**Características:**
- ✅ Velocidade máxima: 61.64ms por inferência
- ✅ Menor uso de memória: 13-15 KB
- ✅ Suporta 9 idiomas: português, inglês, espanhol, francês, alemão, italiano, russo, árabe, hindi
- ⚠️ Não suporta japonês e chinês (CJK)

---

## 🌍 Paraphrase-Multilingual (Configurável)

### Quando Usar

Use o modelo Paraphrase-Multilingual quando precisar de:
- ✅ Máxima qualidade (71% mais efetivo que MS MARCO)
- ✅ Suporte a japonês e chinês (CJK)
- ✅ Cobertura completa de 11 idiomas
- ✅ Latência de ~110ms é aceitável

### Configuração via Código

```go
import "github.com/fsvxavier/nexs-mcp/internal/quality"

config := &quality.Config{
    ONNXModelPath:        "models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
    RequiresTokenTypeIds: true,
    ONNXModelType:        "embedder",
    ONNXOutputName:       "last_hidden_state",
    ONNXOutputShape:      []int64{1, 512, 384},
}

scorer, err := quality.NewONNXScorer(config)
if err != nil {
    log.Fatal(err)
}
defer scorer.Close()
```

### Configuração via JSON (MCP Config)

```json
{
  "quality_config": {
    "onnx_model_path": "models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
    "requires_token_type_ids": true,
    "onnx_model_type": "embedder",
    "onnx_output_name": "last_hidden_state",
    "onnx_output_shape": [1, 512, 384]
  }
}
```

### Configuração via Variáveis de Ambiente

```bash
export NEXS_ONNX_MODEL_PATH="models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx"
export NEXS_ONNX_MODEL_TYPE="embedder"
export NEXS_REQUIRES_TOKEN_TYPE_IDS="true"
export NEXS_ONNX_OUTPUT_NAME="last_hidden_state"
```

---

## 📊 Comparação Detalhada

### Performance

| Métrica | MS MARCO | Paraphrase-Multilingual | Diferença |
|---------|----------|-------------------------|-----------|
| Latência Média | 61.64ms | 109.41ms | +77% |
| Throughput | ~16 inf/s | ~9 inf/s | -44% |
| Uso de Memória | 13-15 KB | 800 KB | +53x |
| Score Médio | 0.3451 | 0.5904 | +71% |

### Cobertura de Idiomas

| Idioma | MS MARCO | Paraphrase-Multilingual |
|--------|----------|-------------------------|
| Português | ✅ 0.3212 | ✅ 0.5138 |
| Inglês | ✅ 0.3332 | ✅ 0.6500 |
| Espanhol | ✅ 0.3241 | ✅ 0.5653 |
| Francês | ✅ 0.3249 | ✅ 0.5721 |
| Alemão | ✅ 0.3171 | ✅ 0.6886 |
| Italiano | ✅ 0.3661 | ✅ 0.6191 |
| Russo | ✅ 0.3821 | ✅ 0.5008 |
| Árabe | ✅ 0.3743 | ✅ 0.6597 |
| Hindi | ✅ 0.3626 | ✅ 0.6804 |
| Japonês | ❌ Não suportado | ✅ 0.4569 |
| Chinês | ❌ Não suportado | ✅ 0.5876 |

---

## 🎯 Matriz de Decisão

### Quando usar MS MARCO (Default)

```
✅ API em tempo real (latência < 70ms crítica)
✅ Alto volume de requisições
✅ Conteúdo em idiomas latinos/árabe/hindi
✅ Restrições de memória
✅ Qualidade "boa" é suficiente
```

### Quando usar Paraphrase-Multilingual

```
✅ Conteúdo japonês ou chinês
✅ Máxima qualidade é prioritária
✅ Análise de sentimento/moderação
✅ Latência de ~110ms é aceitável
✅ Memória não é limitação
```

---

## 🔧 Exemplos Práticos

### Exemplo 1: Avaliação de Conteúdo Multilíngue

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/fsvxavier/nexs-mcp/internal/quality"
)

func main() {
    // Configurar para máxima qualidade multilíngue
    config := &quality.Config{
        ONNXModelPath:        "models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
        RequiresTokenTypeIds: true,
        ONNXModelType:        "embedder",
        ONNXOutputName:       "last_hidden_state",
        ONNXOutputShape:      []int64{1, 512, 384},
    }
    
    scorer, err := quality.NewONNXScorer(config)
    if err != nil {
        log.Fatal(err)
    }
    defer scorer.Close()
    
    // Testar com múltiplos idiomas
    texts := []string{
        "Este é um excelente exemplo de texto em português",
        "This is a great example of English text",
        "これは日本語のテキストの素晴らしい例です", // Japonês
        "这是一个很好的中文文本示例", // Chinês
    }
    
    ctx := context.Background()
    scores, err := scorer.ScoreBatch(ctx, texts)
    if err != nil {
        log.Fatal(err)
    }
    
    for i, score := range scores {
        fmt.Printf("Texto %d: Score = %.4f (Confiança: %.2f%%)\n", 
            i+1, score.Value, score.Confidence*100)
    }
}
```

### Exemplo 2: API de Alta Performance

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/fsvxavier/nexs-mcp/internal/quality"
)

func main() {
    // Usar modelo padrão para máxima velocidade
    scorer, err := quality.NewONNXScorer(nil) // MS MARCO automático
    if err != nil {
        log.Fatal(err)
    }
    defer scorer.Close()
    
    ctx := context.Background()
    content := "Conteúdo para avaliação rápida de qualidade"
    
    score, err := scorer.Score(ctx, content)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Score: %.4f (Latência típica: ~60ms)\n", score.Value)
}
```

### Exemplo 3: Sistema de Fallback

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/fsvxavier/nexs-mcp/internal/quality"
)

func main() {
    // Tentar MS MARCO primeiro, fallback para Paraphrase se CJK
    config := &quality.Config{
        DefaultScorer:  "onnx",
        EnableFallback: true,
        FallbackChain:  []string{"onnx", "implicit"},
    }
    
    fallbackScorer, err := quality.NewFallbackScorer(config)
    if err != nil {
        log.Fatal(err)
    }
    defer fallbackScorer.Close()
    
    ctx := context.Background()
    
    // MS MARCO será usado aqui (idioma suportado)
    score1, _ := fallbackScorer.Score(ctx, "English text")
    fmt.Printf("English: %.4f (método: %s)\n", score1.Value, score1.Method)
    
    // Fallback para implicit se CJK detectado
    score2, _ := fallbackScorer.Score(ctx, "日本語テキスト")
    fmt.Printf("Japanese: %.4f (método: %s)\n", score2.Value, score2.Method)
}
```

---

## 📥 Download de Modelos

### MS MARCO (23 MB)

```bash
# Linux/macOS
wget -O models/ms-marco-MiniLM-L-6-v2/model.onnx \
  https://huggingface.co/sentence-transformers/ms-marco-MiniLM-L-6-v2/resolve/main/onnx/model.onnx

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://huggingface.co/sentence-transformers/ms-marco-MiniLM-L-6-v2/resolve/main/onnx/model.onnx" `
  -OutFile "models/ms-marco-MiniLM-L-6-v2/model.onnx"
```

### Paraphrase-Multilingual (470 MB)

```bash
# Linux/macOS
wget -O models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx \
  https://huggingface.co/sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2/resolve/main/onnx/model.onnx

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://huggingface.co/sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2/resolve/main/onnx/model.onnx" `
  -OutFile "models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx"
```

---

## 🔍 Troubleshooting

### Erro: "ONNX model not found"

**Solução:** Baixe o modelo usando os comandos acima.

```bash
mkdir -p models/ms-marco-MiniLM-L-6-v2
# Download command aqui
```

### Erro: "Token out of vocabulary" (CJK)

**Problema:** MS MARCO não suporta japonês/chinês.

**Solução:** Alternar para Paraphrase-Multilingual:

```go
config := &quality.Config{
    ONNXModelPath: "models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
    ONNXModelType: "embedder",
    // ... outras configurações
}
```

### Performance abaixo do esperado

**Diagnóstico:**
```go
import "time"

start := time.Now()
score, err := scorer.Score(ctx, text)
latency := time.Since(start)

fmt.Printf("Latência: %v (esperado: MS MARCO ~60ms, Paraphrase ~110ms)\n", latency)
```

**Possíveis causas:**
- Modelo errado carregado
- Textos muito longos (> 512 tokens)
- CPU sobrecarregada
- Primeiro run (cold start)

---

## 📚 Recursos Adicionais

- [BENCHMARK_RESULTS.md](../../BENCHMARK_RESULTS.md) - Resultados completos dos benchmarks
- [ONNX_QUALITY_AUDIT.md](../../ONNX_QUALITY_AUDIT.md) - Auditoria técnica completa
- [ONNX Runtime Go](https://github.com/yalue/onnxruntime_go) - Biblioteca utilizada
- [Hugging Face Models](https://huggingface.co/sentence-transformers) - Repositório de modelos

---

## 🔄 Histórico de Modelos

### Modelos Descontinuados

Os seguintes modelos foram testados mas **não estão em produção**:

- ❌ **distiluse-base-multilingual-cased-v1** (768-dim)
  - Motivo: Desempenho inferior (0.2270 score, 180ms latência)
  - Descontinuado: 23/12/2025

- ❌ **distiluse-base-multilingual-cased-v2** (768-dim)
  - Motivo: Desempenho inferior (0.2303 score, 172ms latência)
  - Descontinuado: 23/12/2025

Estes modelos ainda aparecem em alguns testes legados marcados com `t.Skip()`.

---

**Última atualização:** 23 de dezembro de 2025  
**Versão:** 1.0.0  
**Status:** Produção
