# ONNX Models Benchmark Results

Data: 23 de dezembro de 2025
CPU: Intel(R) Core(TM) i7-10750H @ 2.60GHz (12 cores)

## 📊 RESUMO EXECUTIVO

### 🏆 MODELOS EM PRODUÇÃO

**MODELO PADRÃO:** MS MARCO MiniLM-L-6-v2
- **Velocidade:** 61.64ms por inferência (média em 9 idiomas)
- **Throughput:** ~16 inferências/segundo
- **Cobertura:** 9 idiomas (latinos/árabes/hindi)
- **Uso:** Padrão para aplicações de baixa latência

**MODELO CONFIGURÁVEL:** Paraphrase-Multilingual-MiniLM-L12-v2
- **Velocidade:** 109.41ms por inferência (média em 11 idiomas)
- **Throughput:** ~9 inferências/segundo
- **Cobertura:** 11 idiomas incluindo CJK (japonês/chinês)
- **Uso:** Opção para máxima cobertura multilíngue

---

## 📈 COMPARAÇÃO DETALHADA

### Efetividade (Score Médio)

| Rank | Modelo | Score Médio | Cobertura | Performance |
|------|--------|-------------|-----------|-------------|
| 🥇 | **Paraphrase-Multilingual** | **0.5904** | 11/11 (100%) | ⭐⭐⭐⭐⭐ EXCELENTE |
| 🥈 | MS MARCO | 0.3451 | 9/11 (81.8%) | ⭐⭐⭐ BOM (sem CJK) |

### Velocidade (Latência Média)

| Rank | Modelo | Latência | Throughput | Uso de Memória |
|------|--------|----------|------------|----------------|
| 🥇 | **MS MARCO** | **61.64ms** | ~16 inf/s | 13-15 KB |
| 🥈 | Paraphrase-Multilingual | 109.41ms | ~9 inf/s | 800 KB |

### Performance por Tamanho de Texto

*(Benchmarks detalhados por tamanho de texto disponíveis nos testes de performance)*

---

## 🌍 COBERTURA MULTILÍNGUE

### Paraphrase-Multilingual (11/11 idiomas - 100%)
```
✅ Portuguese: 0.5138  ✅ English: 0.6500     ✅ Spanish: 0.5653
✅ French: 0.5721      ✅ German: 0.6886      ✅ Italian: 0.6191
✅ Russian: 0.5008     ✅ Arabic: 0.6597      ✅ Hindi: 0.6804
✅ Japanese: 0.4569    ✅ Chinese: 0.5876
```
**Score Médio:** 0.5904 | **Latência Média:** 109.41ms

### MS MARCO (9/9 - Apenas idiomas não-CJK)
```
✅ Portuguese: 0.3212  ✅ English: 0.3332     ✅ Spanish: 0.3241
✅ French: 0.3249      ✅ German: 0.3171      ✅ Italian: 0.3661
✅ Russian: 0.3821     ✅ Arabic: 0.3743      ✅ Hindi: 0.3626
⊘ Japanese: SKIPPED   ⊘ Chinese: SKIPPED
```
*Idiomas CJK não testados: vocabulário limitado (modelo treinado apenas para inglês)*
**Score Médio (9 idiomas):** 0.3451 | **Latência Média:** 61.64ms

---

## 💡 RECOMENDAÇÕES

### ⚡ Modelo Padrão: MS MARCO MiniLM-L-6-v2
**Use quando:**
- ✅ Velocidade é prioridade (1.8x mais rápido)
- ✅ Conteúdo em idiomas latinos, árabe ou hindi
- ✅ Aplicações em tempo real
- ✅ Restrições de memória (13-15 KB vs 800 KB)

**Evite quando:**
- ⚠️ Precisa processar japonês ou chinês (CJK)
- ⚠️ Qualidade máxima é crítica

### 🌍 Modelo Configurável: Paraphrase-Multilingual-MiniLM-L12-v2
**Use quando:**
- ✅ Qualidade máxima é prioridade (71% mais efetivo)
- ✅ Precisa de cobertura CJK (japonês/chinês)
- ✅ 100% cobertura multilíngue é requisito
- ✅ Latência de ~110ms é aceitável

**Evite quando:**
- ⚠️ Latência abaixo de 100ms é crítica
- ⚠️ Restrições severas de memória

---

## 🔬 DETALHES TÉCNICOS

### Arquitetura dos Modelos

| Modelo | Tipo | Hidden Dim | Camadas | Parâmetros | Tokenizer |
|--------|------|------------|---------|------------|-----------|
| Paraphrase-Multilingual | MiniLM | 384 | 12 | 118M | bert-base-multi |
| MS MARCO | MiniLM | 384 | 6 | 22M | bert-base-uncased |

### Outputs ONNX

| Modelo | Output Name | Output Shape | Processing |
|--------|-------------|--------------|------------|
| Paraphrase-Multilingual | last_hidden_state | [1, 512, 384] | [CLS] extraction |
| MS MARCO | logits | [1, 1] | Direct score |

---

## 📊 ANÁLISE DE TRADE-OFFS

### Velocidade vs Efetividade

```
Efetividade (Score)
  0.60 │                    ●  Paraphrase-Multilingual
       │                       (Melhor qualidade + CJK)
  0.50 │
       │
  0.40 │
       │
  0.30 │  ●  MS MARCO
       │     (Mais rápido - sem CJK)
  0.20 │
       │
       └─────┴─────┴─────┴─────┴─────> Velocidade (ms)
         50   100   150   200   250
```

### Recomendação Final

**🎯 CONFIGURAÇÃO DE PRODUÇÃO:**

#### Modelo Padrão: MS MARCO MiniLM-L-6-v2
- **Perfil:** Velocidade máxima para idiomas não-CJK
- **Performance:** 61.64ms latência | Score 0.3451
- **Cobertura:** 9 idiomas (português, inglês, espanhol, francês, alemão, italiano, russo, árabe, hindi)
- **Uso:** Default para aplicações de baixa latência

#### Modelo Configurável: Paraphrase-Multilingual-MiniLM-L12-v2
- **Perfil:** Qualidade máxima com cobertura total
- **Performance:** 109.41ms latência | Score 0.5904
- **Cobertura:** 11 idiomas (inclui japonês e chinês)
- **Uso:** Configurável para aplicações que requerem CJK ou máxima qualidade

**Quando alternar entre modelos:**

| Cenário | Modelo Recomendado | Motivo |
|---------|-------------------|---------|
| API em tempo real (sem CJK) | **MS MARCO** | 1.8x mais rápido |
| Conteúdo japonês/chinês | **Paraphrase-Multilingual** | Única opção com CJK |
| Análise de qualidade | **Paraphrase-Multilingual** | 71% mais efetivo |
| Alta concorrência | **MS MARCO** | Menor uso de memória |
