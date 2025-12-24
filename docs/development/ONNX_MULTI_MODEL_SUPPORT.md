# Suporte para Múltiplos Modelos ONNX

## Visão Geral

Este documento descreve as adaptações necessárias para suportar tanto modelos **cross-encoder** (rerankers) quanto **sentence transformers** (embeddings) no `ONNXScorer`.

## Modelos Testados

### ✅ MS MARCO MiniLM-L-6-v2 (Atual - FUNCIONA)
- **Tipo**: Cross-encoder reranker
- **Input**: Query + Passage concatenados
- **Output**: Score único (logits)
- **Arquitetura**: 3 inputs (input_ids, attention_mask, token_type_ids)
- **Output Name**: `logits`
- **Tamanho**: 87MB
- **Idiomas**: 9/11 (81.8%) - Falha em CJK (Japanese, Chinese)
- **Status**: ✅ PRODUÇÃO - 61 testes passando, fallback automático para CJK

### ❌ Paraphrase-Multilingual-MiniLM-L12-v2 (INCOMPATÍVEL)
- **Tipo**: Sentence transformer (bi-encoder)
- **Input**: Texto individual
- **Output**: Embeddings de 384 dimensões
- **Arquitetura**: 3 inputs esperados, mas output incompatível
- **Output Name**: `last_hidden_state` (código espera `logits`)
- **Tamanho**: 449MB
- **Idiomas**: 50+ incluindo CJK
- **Erro**: `Invalid output name: logits`
- **Status**: ❌ Requer refatoração para suportar embeddings

### ❌ distiluse-base-multilingual-cased-v1 (INCOMPATÍVEL)
- **Tipo**: Sentence transformer (bi-encoder)
- **Input**: Texto individual
- **Output**: Embeddings de 512 dimensões
- **Arquitetura**: 2 inputs (input_ids, attention_mask) - DistilBERT
- **Output Name**: Desconhecido
- **Tamanho**: 514MB
- **Idiomas**: 14 idiomas
- **Erro**: `Invalid input name: token_type_ids`
- **Status**: ❌ DistilBERT não usa token_type_ids

### ❌ distiluse-base-multilingual-cased-v2 (INCOMPATÍVEL)
- **Tipo**: Sentence transformer (bi-encoder)
- **Input**: Texto individual
- **Output**: Embeddings de 512 dimensões
- **Arquitetura**: 2 inputs (input_ids, attention_mask) - DistilBERT
- **Output Name**: Desconhecido
- **Tamanho**: 514MB
- **Idiomas**: 50 idiomas incluindo CJK
- **Erro**: `Invalid input name: token_type_ids`
- **Status**: ❌ DistilBERT não usa token_type_ids

## Resumo dos Testes

**Total de modelos testados**: 4  
**Modelos funcionando**: 1 (MS MARCO MiniLM-L-6-v2)  
**Taxa de sucesso**: 25%

### Problemas Identificados por Categoria

#### 1. Token Type IDs (2 modelos)
Modelos baseados em DistilBERT **não usam** `token_type_ids`:
- ❌ distiluse-v1 (DistilBERT)
- ❌ distiluse-v2 (DistilBERT)

#### 2. Output Name/Shape (1 modelo)
Modelo retorna embeddings em vez de scores:
- ❌ paraphrase-multilingual-MiniLM-L12-v2 (output: embeddings 384-dim)

## Problema Atual

O código atual em `internal/quality/onnx.go` está **hardcoded** para cross-encoders:

```go
// Linha 125-130: Hardcoded para cross-encoder
session, err := ort.NewAdvancedSession(
    s.modelPath,
    []string{"input_ids", "attention_mask", "token_type_ids"},
    []string{"logits"}, // ❌ Sentence transformers não têm "logits"
    []ort.ArbitraryTensor{inputTensor, attentionMask, tokenTypeIDs},
    []ort.ArbitraryTensor{outputTensor},
    nil,
)

// Linha 114-117: Output shape para score único
outputShape := ort.NewShape(1, 1) // ❌ Sentence transformers retornam (1, 384)
outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
```

**Erro ao usar paraphrase-multilingual**:
```
Error running network: Invalid output name: logits
```

## Adaptações Necessárias

### 1. Detecção de Tipo de Modelo

Adicionar campo na `Config` e `ONNXScorer`:

```go
// Config em config.go
type Config struct {
    // ... campos existentes ...
    ONNXModelType  string // "reranker" ou "embedder"
    ONNXOutputName string // "logits", "last_hidden_state", etc.
    EmbeddingDim   int    // 384, 768, 1024, etc.
}

// ONNXScorer em onnx.go
type ONNXScorer struct {
    // ... campos existentes ...
    modelType    string
    outputName   string
    embeddingDim int
}
```

### 2. Criação Dinâmica de Tensores

Modificar `initialize()` para suportar diferentes shapes:

```go
func (s *ONNXScorer) initialize() error {
    // ... código existente ...
    
    var outputShape ort.Shape
    var outputTensor *ort.Tensor[float32]
    var outputNames []string
    
    switch s.modelType {
    case "reranker":
        // Cross-encoder: retorna score único
        outputShape = ort.NewShape(1, 1)
        outputNames = []string{"logits"}
        
    case "embedder":
        // Sentence transformer: retorna embeddings
        outputShape = ort.NewShape(1, s.embeddingDim) // Ex: (1, 384)
        outputNames = []string{"last_hidden_state"} // ou "sentence_embedding"
        
    default:
        return fmt.Errorf("unknown model type: %s", s.modelType)
    }
    
    outputTensor, err = ort.NewEmptyTensor[float32](outputShape)
    if err != nil {
        // cleanup...
        return fmt.Errorf("failed to create output tensor: %w", err)
    }
    
    // Create session with dynamic output names
    session, err := ort.NewAdvancedSession(
        s.modelPath,
        []string{"input_ids", "attention_mask", "token_type_ids"},
        outputNames, // ✅ Dinâmico agora
        []ort.ArbitraryTensor{inputTensor, attentionMask, tokenTypeIDs},
        []ort.ArbitraryTensor{outputTensor},
        nil,
    )
    
    // ... resto do código ...
}
```

### 3. Processamento de Inferência Dual

Modificar `runInference()` para processar ambos os tipos:

```go
func (s *ONNXScorer) runInference(tokenIDs []int64) (float64, float64, error) {
    // ... preparação de inputs (igual) ...
    
    // Run inference
    err := s.session.Run()
    if err != nil {
        return 0, 0, fmt.Errorf("inference execution failed: %w", err)
    }
    
    outputData := s.outputTensor.GetData()
    if len(outputData) == 0 {
        return 0, 0, fmt.Errorf("empty output data")
    }
    
    var qualityScore float64
    
    switch s.modelType {
    case "reranker":
        // Cross-encoder: output direto é o score
        rawScore := float64(outputData[0])
        qualityScore = 1.0 / (1.0 + math.Exp(-rawScore/10.0))
        
    case "embedder":
        // Sentence transformer: calcular norma L2 do embedding
        // Score baseado na magnitude do vetor
        var sumSquares float64
        for _, val := range outputData {
            sumSquares += float64(val) * float64(val)
        }
        magnitude := math.Sqrt(sumSquares)
        
        // Normalizar para 0-1 (assumindo magnitude típica 1-10)
        qualityScore = magnitude / 10.0
        if qualityScore > 1.0 {
            qualityScore = 1.0
        }
    }
    
    // Clamp to valid range
    if qualityScore < 0 {
        qualityScore = 0
    }
    if qualityScore > 1 {
        qualityScore = 1
    }
    
    confidence := 0.9
    if qualityScore < 0.1 || qualityScore > 0.9 {
        confidence = 0.95
    }
    
    return qualityScore, confidence, nil
}
```

### 4. Suporte para Similaridade (Opcional - Melhor Abordagem)

Para sentence transformers, a abordagem correta é:
1. Gerar embeddings para query e passage **separadamente**
2. Calcular similaridade (cosine ou dot product)

```go
// Nova função para sentence transformers
func (s *ONNXScorer) computeEmbedding(tokenIDs []int64) ([]float32, error) {
    // ... preparar inputs ...
    
    err := s.session.Run()
    if err != nil {
        return nil, fmt.Errorf("inference failed: %w", err)
    }
    
    outputData := s.outputTensor.GetData()
    embedding := make([]float32, len(outputData))
    copy(embedding, outputData)
    
    return embedding, nil
}

func (s *ONNXScorer) ScoreWithQuery(ctx context.Context, query, passage string) (*Score, error) {
    if s.modelType != "embedder" {
        return s.Score(ctx, passage) // Fallback para reranker
    }
    
    // Gerar embeddings
    queryTokens, _ := s.encodeContent(query)
    passageTokens, _ := s.encodeContent(passage)
    
    queryEmb, err := s.computeEmbedding(queryTokens)
    if err != nil {
        return nil, err
    }
    
    passageEmb, err := s.computeEmbedding(passageTokens)
    if err != nil {
        return nil, err
    }
    
    // Calcular similaridade cosine
    similarity := cosineSimilarity(queryEmb, passageEmb)
    
    // Normalizar para 0-1 (similaridade já é -1 a 1, converter para 0-1)
    score := (similarity + 1.0) / 2.0
    
    return &Score{
        Value:      score,
        Confidence: 0.9,
        Method:     "onnx-embedder",
        Timestamp:  time.Now(),
        Metadata: map[string]interface{}{
            "model":      "paraphrase-multilingual-MiniLM-L12-v2",
            "similarity": similarity,
        },
    }, nil
}

func cosineSimilarity(a, b []float32) float64 {
    if len(a) != len(b) {
        return 0
    }
    
    var dotProduct, normA, normB float64
    for i := range a {
        dotProduct += float64(a[i]) * float64(b[i])
        normA += float64(a[i]) * float64(a[i])
        normB += float64(b[i]) * float64(b[i])
    }
    
    if normA == 0 || normB == 0 {
        return 0
    }
    
    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

## Configuração

### Opção 1: Auto-detecção (Complexo)

Inspecionar modelo ONNX para determinar tipo automaticamente:
- Verificar output names ("logits" → reranker, "last_hidden_state" → embedder)
- Verificar output shape ((1,1) → reranker, (1, dim) → embedder)

### Opção 2: Configuração Manual (Recomendado)

```go
// Para MS MARCO (atual)
config := &Config{
    ONNXModelPath:  "models/ms-marco-MiniLM-L-6-v2.onnx",
    ONNXModelType:  "reranker",
    ONNXOutputName: "logits",
    EmbeddingDim:   1,
}

// Para Paraphrase Multilingual
config := &Config{
    ONNXModelPath:  "models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
    ONNXModelType:  "embedder",
    ONNXOutputName: "last_hidden_state",
    EmbeddingDim:   384,
}
```

## Estimativa de Esforço

### Problema 1: Suporte a Token Type IDs Opcional (CRÍTICO)

**Modelos afetados**: 2 de 4 (50%)

**Esforço**: 4-6 horas

**Mudanças necessárias**:

1. **Detectar se modelo precisa de token_type_ids** (1h)
```go
// config.go
type Config struct {
    // ... campos existentes ...
    RequiresTokenTypeIds bool // true para BERT, false para DistilBERT/RoBERTa
}
```

2. **Criar inputs condicionalmente** (2h)
```go
// onnx.go - initialize()
var inputNames []string
var inputTensors []ort.ArbitraryTensor

// Inputs obrigatórios
inputNames = append(inputNames, "input_ids", "attention_mask")
inputTensors = append(inputTensors, inputTensor, attentionMask)

// Token type IDs opcional
if s.config.RequiresTokenTypeIds {
    inputNames = append(inputNames, "token_type_ids")
    inputTensors = append(inputTensors, tokenTypeIDs)
}

session, err := ort.NewAdvancedSession(
    s.modelPath,
    inputNames,      // ✅ Dinâmico
    outputNames,
    inputTensors,    // ✅ Dinâmico
    []ort.ArbitraryTensor{outputTensor},
    nil,
)
```

3. **Atualizar runInference()** (1h)
```go
// onnx.go - runInference()
// Preencher token_type_ids apenas se necessário
if s.config.RequiresTokenTypeIds {
    typeIDsData := s.tokenTypeIDs.GetData()
    for i := range typeIDsData {
        typeIDsData[i] = 0
    }
}
```

4. **Testes** (2h)
   - Testar modelos BERT (MS MARCO) - deve continuar funcionando
   - Testar modelos DistilBERT - deve funcionar agora
   - Validar 11 idiomas para cada modelo

**Impacto**: Habilita 2 modelos DistilBERT multilíngues adicionais

### Problema 2: Suporte a Output Dinâmico (IMPORTANTE)

**Modelos afetados**: 3 de 4 (sentence transformers)

**Esforço**: 6-10 horas

**Mudanças necessárias**:

1. **Adicionar configuração de tipo e shape** (1h)
```go
// config.go
type Config struct {
    // ... campos existentes ...
    ONNXModelType    string // "reranker" ou "embedder"
    ONNXOutputName   string // "logits", "last_hidden_state", etc.
    ONNXOutputShape  []int64 // [1, 1] ou [1, 384] ou [1, 512] ou [1, 768]
}
```

2. **Criar output tensor dinâmico** (2h)
```go
// onnx.go - initialize()
var outputShape ort.Shape
if len(s.config.ONNXOutputShape) == 2 {
    outputShape = ort.NewShape(s.config.ONNXOutputShape[0], s.config.ONNXOutputShape[1])
} else {
    // Default: reranker com score único
    outputShape = ort.NewShape(1, 1)
}

outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
```

3. **Processar outputs diferentes** (3h)
```go
// onnx.go - runInference()
outputData := s.outputTensor.GetData()

var qualityScore float64

switch s.config.ONNXModelType {
case "reranker":
    // Score único direto
    rawScore := float64(outputData[0])
    qualityScore = 1.0 / (1.0 + math.Exp(-rawScore/10.0))
    
case "embedder":
    // Calcular magnitude do embedding
    var sumSquares float64
    for _, val := range outputData {
        sumSquares += float64(val) * float64(val)
    }
    magnitude := math.Sqrt(sumSquares)
    qualityScore = math.Min(magnitude / 10.0, 1.0)
    
default:
    return 0, 0, fmt.Errorf("unknown model type: %s", s.config.ONNXModelType)
}
```

4. **Testes** (4h)
   - Testar cross-encoder (MS MARCO)
   - Testar sentence transformers (3 modelos)
   - Validar scores fazem sentido
   - Testar 11 idiomas

**Impacto**: Habilita uso de sentence transformers (embeddings)

### Problema 3: Similaridade Cosine para Embedders (OPCIONAL)

**Esforço**: 4-6 horas adicional

**Nota**: Abordagem correta para sentence transformers é calcular similaridade entre embeddings de query e passage, não magnitude do embedding. Implementação mais complexa mas resultados melhores.

### Esforço Total Estimado

**Abordagem Mínima** (Problema 1 apenas): 4-6 horas
- ✅ Habilita 2 modelos DistilBERT
- ❌ Sentence transformers ainda retornam scores incorretos

**Abordagem Intermediária** (Problemas 1 + 2): 10-16 horas  
- ✅ Habilita todos os 3 modelos multilíngues
- ✅ Sentence transformers funcionam (com magnitude)
- ❌ Scores de embedders podem ser subótimos

**Abordagem Completa** (Problemas 1 + 2 + 3): 14-22 horas
- ✅ Habilita todos os 3 modelos multilíngues
- ✅ Sentence transformers com similaridade cosine correta
- ✅ Melhor qualidade de scores
3. Adicionar auto-detecção de tipo de modelo (2h)
4. Testes extensivos com 11 idiomas (2-3h)
5. Documentação (1h)

## Arquivos a Modificar

1. **internal/quality/config.go**
   - Adicionar campos: `RequiresTokenTypeIds`, `ONNXModelType`, `ONNXOutputName`, `ONNXOutputShape`

2. **internal/quality/onnx.go**
   - Modificar struct `ONNXScorer`
   - Modificar `initialize()` para:
     * Inputs dinâmicos (com/sem token_type_ids)
     * Output shape dinâmico
     * Output names dinâmicos
   - Modificar `runInference()` para processar ambos tipos
   - Adicionar `computeEmbedding()` (opcional)
   - Adicionar `ScoreWithQuery()` (opcional)
   - Adicionar `cosineSimilarity()` (opcional)

3. **internal/quality/onnx_test.go**
   - Adicionar testes para modelos embedder
   - Testar modelos com/sem token_type_ids
   - Testar ambos os tipos de modelos

4. **internal/quality/multilingual_models_test.go**
   - Atualizar testes para usar configuração correta

## Configurações Recomendadas para Cada Modelo

### MS MARCO MiniLM-L-6-v2 (Atual)
```go
config := &Config{
    ONNXModelPath:        "models/ms-marco-MiniLM-L-6-v2/model.onnx",
    RequiresTokenTypeIds: true,
    ONNXModelType:        "reranker",
    ONNXOutputName:       "logits",
    ONNXOutputShape:      []int64{1, 1},
}
```

### distiluse-base-multilingual-cased-v1
```go
config := &Config{
    ONNXModelPath:        "models/distiluse-base-multilingual-cased-v1/model.onnx",
    RequiresTokenTypeIds: false, // DistilBERT não usa
    ONNXModelType:        "embedder",
    ONNXOutputName:       "last_hidden_state", // ou verificar modelo
    ONNXOutputShape:      []int64{1, 512},
}
```

### distiluse-base-multilingual-cased-v2
```go
config := &Config{
    ONNXModelPath:        "models/distiluse-base-multilingual-cased-v2/model.onnx",
    RequiresTokenTypeIds: false, // DistilBERT não usa
    ONNXModelType:        "embedder",
    ONNXOutputName:       "last_hidden_state",
    ONNXOutputShape:      []int64{1, 512},
}
```

### paraphrase-multilingual-MiniLM-L12-v2
```go
config := &Config{
    ONNXModelPath:        "models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
    RequiresTokenTypeIds: true, // BERT-based
    ONNXModelType:        "embedder",
    ONNXOutputName:       "last_hidden_state",
    ONNXOutputShape:      []int64{1, 384},
}
```

## Recomendação

**Opção 1: Manter MS MARCO atual** (RECOMENDADO)
- ✅ Funcional (9/11 idiomas = 81.8%)
- ✅ Testada (61 testes passando)
- ✅ Produção-ready
- ✅ Fallback automático para CJK
- ✅ Zero esforço adicional
- ❌ Não suporta CJK nativamente (mas fallback resolve)

**Opção 2: Implementar Problema 1 apenas** (SE CJK FOR CRÍTICO)
- ✅ Habilita 2 modelos DistilBERT multilíngues com CJK
- ✅ Esforço moderado (4-6 horas)
- ❌ Sentence transformers ainda com scores incorretos
- ❌ Requer testes extensivos

**Opção 3: Implementação completa** (APENAS SE MUITO NECESSÁRIO)
- ✅ Habilita todos os 3 modelos multilíngues
- ✅ Sentence transformers funcionam corretamente
- ❌ Esforço significativo (14-22 horas)
- ❌ Complexidade adicional no código
- ❌ Mais testes necessários

## Conclusão

Após testes com **4 modelos multilíngues**, apenas o **MS MARCO MiniLM-L-6-v2** funciona com o código atual. Todos os outros 3 modelos requerem refatoração significativa:

**Problemas encontrados**:
1. 50% (2/4) dos modelos não aceitam `token_type_ids` (DistilBERT)
2. 75% (3/4) dos modelos são sentence transformers (não cross-encoders)
3. Todos os 3 modelos multilíngues candidatos falharam por incompatibilidade

**Decisão recomendada**: Manter arquitetura atual até Sprint 9+, quando houver:
1. Necessidade **comprovada** de melhor suporte CJK (fallback atual funciona)
2. Tempo dedicado para refatoração e testes (14-22 horas)
3. Modelo multilingual **cross-encoder** com ONNX pronto e compatível

A abordagem atual (MS MARCO + fallback) oferece o melhor custo-benefício:
- ✅ 81.8% de cobertura nativa (9/11 idiomas)
- ✅ 100% de cobertura com fallback (11/11 idiomas)
- ✅ Produção estável
- ✅ Zero trabalho adicional
