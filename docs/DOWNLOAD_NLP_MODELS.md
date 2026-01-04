# 📥 Guia de Download dos Modelos NLP para NEXS-MCP

**Version:** 1.4.0 (Sprint 18)
**Last Updated:** January 4, 2026
**Applies to:** NEXS-MCP v1.4.0+

## 🎯 Modelos Necessários

1. **BERT NER** (Entity Extraction)
   - Modelo: `protectai/bert-base-NER-onnx` (já em formato ONNX)
   - Base: `dslim/bert-base-NER`
   - Formato: ONNX
   - Tamanho: ~400 MB
   - Destino: `models/bert-base-ner/model.onnx`

2. **DistilBERT Sentiment** (Sentiment Analysis)
   - Modelo: `lxyuan/distilbert-base-multilingual-cased-sentiments-student`
   - Formato: ONNX (requer conversão)
   - Tamanho: ~500 MB
   - Idiomas: Multilingual (incluindo português)
   - Destino: `models/distilbert-sentiment/model.onnx`

## 📋 Método 1: Download de Hugging Face + Conversão (Recomendado)

### Pré-requisitos:
```bash
pip install torch transformers onnx onnxruntime optimum
```

### Script de Download e Conversão:

Crie o arquivo `download_nlp_models.py`:

```python
#!/usr/bin/env python3
"""
Script para baixar e converter modelos NLP para ONNX
"""
import os
from pathlib import Path
import torch
from transformers import AutoTokenizer
from optimum.onnxruntime import ORTModelForTokenClassification, ORTModelForSequenceClassification

def download_and_convert_bert_ner():
    """Baixa BERT NER já em formato ONNX"""
    print("📥 Baixando BERT NER...")
    # Modelo já convertido pela ProtectAI
    model_name = "protectai/bert-base-NER-onnx"
    output_dir = "models/bert-base-ner"

    # Criar diretório
    Path(output_dir).mkdir(parents=True, exist_ok=True)

    # Baixar tokenizer do modelo original
    tokenizer = AutoTokenizer.from_pretrained("dslim/bert-base-NER")
    tokenizer.save_pretrained(output_dir)

    # Converter para ONNX usando Optimum
    model = ORTModelForTokenClassification.from_pretrained(
        model_name,
        export=True
    )
    model.save_pretrained(output_dir)

    print(f"✅ BERT NER salvo em: {output_dir}/")
    print(f"   Arquivos: model.onnx, vocab.txt, tokenizer.json")

def download_and_convert_distilbert_sentiment():
    """Baixa e converte DistilBERT Sentiment para ONNX"""
    print("\n📥 Baixando DistilBERT Sentiment...")
    model_name = "distilbert-base-uncased-finetuned-sst-2-english"
    output_dir = "models/distilbert-sentiment"

    # Criar diretório
    Path(output_dir).mkdir(parents=True, exist_ok=True)

    # Baixar tokenizer
    tokenizer = AutoTokenizer.from_pretrained(model_name)
    tokenizer.save_pretrained(output_dir)

    # Converter para ONNX usando Optimum
    model = ORTModelForSequenceClassification.from_pretrained(
        model_name,
        export=True
    )
    model.save_pretrained(output_dir)

    print(f"✅ DistilBERT Sentiment salvo em: {output_dir}/")
    print(f"   Arquivos: model.onnx, vocab.txt, tokenizer.json")

def main():
    print("🚀 Iniciando download e conversão de modelos NLP\n")

    # BERT NER
    try:
        download_and_convert_bert_ner()
    except Exception as e:
        print(f"❌ Erro ao processar BERT NER: {e}")

    # DistilBERT Sentiment
    try:
        download_and_convert_distilbert_sentiment()
    except Exception as e:
        print(f"❌ Erro ao processar DistilBERT Sentiment: {e}")

    print("\n✅ Download e conversão concluídos!")
    print("\nPróximos passos:")
    print("1. Verifique os arquivos em models/bert-base-ner/ e models/distilbert-sentiment/")
    print("2. Configure o .env:")
    print("   NEXS_NLP_ENTITY_EXTRACTION_ENABLED=true")
    print("   NEXS_NLP_SENTIMENT_ENABLED=true")
    print("3. Reinicie o nexs-mcp")

if __name__ == "__main__":
    main()
```

Execute o script:
```bash
python3 download_nlp_models.py
```

## 📋 Método 2: Download Manual via wget

### BERT NER:
```bash
# Criar diretório
mkdir -p models/bert-base-ner

# Download de modelos ONNX pré-convertidos
# Nota: Alguns modelos podem não ter versão ONNX pública
# Neste caso, use o Método 1 ou Método 3

# Alternativa: Baixar do Hugging Face Hub
huggingface-cli download dslim/bert-base-NER \
  --include "*.onnx" "vocab.txt" "tokenizer.json" \
  --local-dir models/bert-base-ner
```

### DistilBERT Sentiment:
```bash
# Criar diretório
mkdir -p models/distilbert-sentiment

# Download via Hugging Face CLI
huggingface-cli download distilbert-base-uncased-finetuned-sst-2-english \
  --include "*.onnx" "vocab.txt" "tokenizer.json" \
  --local-dir models/distilbert-sentiment
```

**Instalação do Hugging Face CLI:**
```bash
pip install huggingface_hub[cli]
```

## 📋 Método 3: Conversão Manual com PyTorch

Se os modelos ONNX não estiverem disponíveis, converta manualmente:

```python
#!/usr/bin/env python3
import torch
from transformers import AutoTokenizer, AutoModelForTokenClassification

# Baixar modelo PyTorch
model_name = "dslim/bert-base-NER"
model = AutoModelForTokenClassification.from_pretrained(model_name)
tokenizer = AutoTokenizer.from_pretrained(model_name)

# Salvar tokenizer
tokenizer.save_pretrained("models/bert-base-ner/")

# Preparar input dummy
dummy_input = {
    "input_ids": torch.randint(0, 28996, (1, 512)),
    "attention_mask": torch.ones(1, 512, dtype=torch.long),
    "token_type_ids": torch.zeros(1, 512, dtype=torch.long)
}

# Exportar para ONNX
torch.onnx.export(
    model,
    (dummy_input["input_ids"], dummy_input["attention_mask"], dummy_input["token_type_ids"]),
    "models/bert-base-ner/model.onnx",
    input_names=["input_ids", "attention_mask", "token_type_ids"],
    output_names=["logits"],
    dynamic_axes={
        "input_ids": {0: "batch", 1: "sequence"},
        "attention_mask": {0: "batch", 1: "sequence"},
        "token_type_ids": {0: "batch", 1: "sequence"},
        "logits": {0: "batch", 1: "sequence"}
    },
    opset_version=14
)

print("✅ Modelo BERT NER exportado para ONNX")
```

Repita o processo para DistilBERT Sentiment.

## 🔍 Verificação dos Modelos

Após o download, verifique se os arquivos estão corretos:

```bash
# Verificar estrutura
ls -lh models/bert-base-ner/
ls -lh models/distilbert-sentiment/

# Arquivos esperados:
# - model.onnx (arquivo principal)
# - vocab.txt (vocabulário)
# - tokenizer.json (configuração do tokenizer)
# - config.json (opcional, configuração do modelo)

# Verificar tamanho dos arquivos
du -sh models/bert-base-ner/
du -sh models/distilbert-sentiment/
```

### Teste de Carregamento:

```python
import onnxruntime as ort
import numpy as np

# Testar BERT NER
session1 = ort.InferenceSession('models/bert-base-ner/model.onnx')
print('✅ BERT NER carregado:', session1.get_inputs()[0].name)

# Testar DistilBERT Sentiment
session2 = ort.InferenceSession('models/distilbert-sentiment/model.onnx')
print('✅ DistilBERT Sentiment carregado:', session2.get_inputs()[0].name)
```

## ⚙️ Configuração Final

Depois de baixar os modelos, atualize o `.env`:

```bash
# Habilitar recursos NLP
NEXS_NLP_ENTITY_EXTRACTION_ENABLED=true
NEXS_NLP_SENTIMENT_ENABLED=true

# Paths dos modelos (já configurados por padrão)
NEXS_NLP_ENTITY_MODEL=models/bert-base-ner/model.onnx
NEXS_NLP_SENTIMENT_MODEL=models/distilbert-sentiment/model.onnx

# Configurações opcionais
NEXS_NLP_ENTITY_CONFIDENCE_MIN=0.7
NEXS_NLP_SENTIMENT_THRESHOLD=0.6
NEXS_NLP_USE_GPU=false
NEXS_NLP_BATCH_SIZE=16
NEXS_NLP_MAX_LENGTH=512
```

## 🚀 Teste Rápido

```bash
# Compilar com suporte ONNX
make build-onnx

# Executar
./bin/nexs-mcp

# Verificar logs
# Você deve ver:
# "ONNX BERT Provider initialized successfully"
# "Entity extraction enabled: true"
# "Sentiment analysis enabled: true"
```

## 📊 Modelos Alternativos

Se preferir modelos diferentes:

### Entity Extraction:
- `dbmdz/bert-large-cased-finetuned-conll03-english` (maior, mais preciso)
- `Babelscape/wikineural-multilingual-ner` (suporte multilingual)
- `xlm-roberta-large-finetuned-conll03-english` (multilingual)

### Sentiment Analysis:
- `cardiffnlp/twitter-roberta-base-sentiment-latest` (Twitter/social media)
- `nlptown/bert-base-multilingual-uncased-sentiment` (multilingual, 5 classes)
- `ProsusAI/finbert` (análise de sentimento financeiro)

Ajuste o script de download substituindo `model_name` pelo desejado.

## 🐳 Docker com Modelos Pré-configurados

Para ambiente Docker com modelos incluídos:

```dockerfile
FROM nexs-mcp:latest

# Instalar dependências Python
RUN apt-get update && apt-get install -y python3-pip && \
    pip3 install --no-cache-dir optimum onnxruntime transformers

# Download modelos durante build
WORKDIR /app
RUN python3 -c "
from optimum.onnxruntime import ORTModelForTokenClassification, ORTModelForSequenceClassification
from transformers import AutoTokenizer

# BERT NER
model1 = ORTModelForTokenClassification.from_pretrained('dslim/bert-base-NER', export=True)
model1.save_pretrained('/app/models/bert-base-ner')
tok1 = AutoTokenizer.from_pretrained('dslim/bert-base-NER')
tok1.save_pretrained('/app/models/bert-base-ner')

# DistilBERT Sentiment
model2 = ORTModelForSequenceClassification.from_pretrained('distilbert-base-uncased-finetuned-sst-2-english', export=True)
model2.save_pretrained('/app/models/distilbert-sentiment')
tok2 = AutoTokenizer.from_pretrained('distilbert-base-uncased-finetuned-sst-2-english')
tok2.save_pretrained('/app/models/distilbert-sentiment')
"

# Habilitar NLP no .env
RUN echo "NEXS_NLP_ENTITY_EXTRACTION_ENABLED=true" >> .env && \
    echo "NEXS_NLP_SENTIMENT_ENABLED=true" >> .env

CMD ["./nexs-mcp"]
```

Build e execute:
```bash
docker build -t nexs-mcp-nlp .
docker run -p 3000:3000 nexs-mcp-nlp
```

## ❓ Solução de Problemas

### Erro: "model.onnx not found"
- Verifique se os arquivos foram baixados corretamente
- Confirme os paths no `.env`
- Use paths absolutos se necessário

### Erro: "Failed to load ONNX model"
- Verifique se o ONNX Runtime está instalado: `ldconfig -p | grep onnx`
- Teste manualmente com Python: `import onnxruntime`
- Reinstale: `pip install onnxruntime` ou compile com suporte ONNX

### Erro: "Out of memory"
- Reduza `NEXS_NLP_BATCH_SIZE` (padrão: 16 → 8 ou 4)
- Use modelos menores (DistilBERT em vez de BERT)
- Reduza `NEXS_NLP_MAX_LENGTH` (padrão: 512 → 256)

### Performance ruim:
- Habilite GPU se disponível: `NEXS_NLP_USE_GPU=true`
- Verifique CUDA/ROCm instalado
- Use modelos otimizados (quantizados)
- Aumente `NEXS_NLP_BATCH_SIZE` se tiver RAM suficiente

### Erro: "Invalid model format"
- Verifique a versão do opset ONNX (deve ser 14+)
- Reconverta o modelo com `opset_version=14` ou superior
- Atualize o ONNX Runtime: `pip install --upgrade onnxruntime`

## 📚 Referências

- [Hugging Face Hub](https://huggingface.co/models)
- [ONNX Model Zoo](https://github.com/onnx/models)
- [Optimum Documentation](https://huggingface.co/docs/optimum)
- [ONNX Runtime](https://onnxruntime.ai/)
- [docs/NLP_FEATURES.md](./NLP_FEATURES.md) - Documentação completa dos recursos NLP

## 📝 Notas

- Os modelos são baixados do Hugging Face Hub e podem requerer aceitação de termos de uso
- Modelos multilingual suportam português, espanhol, francês, alemão, etc.
- Para produção, considere quantização (INT8) para reduzir tamanho e melhorar performance
- Backup dos modelos é recomendado para evitar redownload
