#!/usr/bin/env python3
"""
Script para baixar e converter modelos NLP para ONNX
Usado pelo NEXS-MCP para entity extraction e sentiment analysis
"""
import os
import sys
from pathlib import Path

def check_dependencies():
    """Verifica se as dependências estão instaladas"""
    required = ['torch', 'transformers', 'optimum']
    missing = []

    for package in required:
        try:
            __import__(package)
        except ImportError:
            missing.append(package)

    if missing:
        print(f"❌ Pacotes faltando: {', '.join(missing)}")
        print(f"\nInstale com:")
        print(f"  pip install {' '.join(missing)} onnxruntime")
        sys.exit(1)

def download_and_convert_bert_ner():
    """Baixa BERT NER já em formato ONNX"""
    from transformers import AutoTokenizer
    from optimum.onnxruntime import ORTModelForTokenClassification

    print("📥 Baixando BERT NER (ONNX)...")
    # Modelo já convertido para ONNX pela ProtectAI
    model_name = "protectai/bert-base-NER-onnx"
    output_dir = "models/bert-base-ner"

    # Criar diretório
    Path(output_dir).mkdir(parents=True, exist_ok=True)

    # Baixar tokenizer do modelo original
    print("   Baixando tokenizer...")
    tokenizer = AutoTokenizer.from_pretrained("dslim/bert-base-NER")
    tokenizer.save_pretrained(output_dir)

    # Baixar modelo ONNX (já convertido)
    print("   Baixando modelo ONNX (já convertido)...")
    model = ORTModelForTokenClassification.from_pretrained(
        model_name
    )
    model.save_pretrained(output_dir)

    # Verificar arquivos
    onnx_file = Path(output_dir) / "model.onnx"
    vocab_file = Path(output_dir) / "vocab.txt"

    if onnx_file.exists():
        size_mb = onnx_file.stat().st_size / (1024 * 1024)
        print(f"✅ BERT NER salvo em: {output_dir}/")
        print(f"   Arquivo ONNX: {size_mb:.1f} MB")
        print(f"   Vocabulário: {'✅' if vocab_file.exists() else '❌'}")
        return True
    else:
        print(f"❌ Erro: arquivo model.onnx não foi criado")
        return False

def download_and_convert_distilbert_sentiment():
    """Baixa DistilBERT Sentiment Multilingual"""
    from transformers import AutoTokenizer
    from optimum.onnxruntime import ORTModelForSequenceClassification

    print("\n📥 Baixando DistilBERT Sentiment Multilingual...")
    # Modelo multilingual que suporta múltiplos idiomas incluindo português
    model_name = "lxyuan/distilbert-base-multilingual-cased-sentiments-student"
    output_dir = "models/distilbert-sentiment"

    # Criar diretório
    Path(output_dir).mkdir(parents=True, exist_ok=True)

    # Baixar tokenizer
    print("   Baixando tokenizer...")
    tokenizer = AutoTokenizer.from_pretrained(model_name)
    tokenizer.save_pretrained(output_dir)

    # Converter para ONNX usando Optimum
    print("   Convertendo para ONNX (pode demorar alguns minutos)...")
    model = ORTModelForSequenceClassification.from_pretrained(
        model_name,
        export=True
    )
    model.save_pretrained(output_dir)

    # Verificar arquivos
    onnx_file = Path(output_dir) / "model.onnx"
    vocab_file = Path(output_dir) / "vocab.txt"

    if onnx_file.exists():
        size_mb = onnx_file.stat().st_size / (1024 * 1024)
        print(f"✅ DistilBERT Sentiment salvo em: {output_dir}/")
        print(f"   Arquivo ONNX: {size_mb:.1f} MB")
        print(f"   Vocabulário: {'✅' if vocab_file.exists() else '❌'}")
        return True
    else:
        print(f"❌ Erro: arquivo model.onnx não foi criado")
        return False

def main():
    """Função principal"""
    print("=" * 70)
    print("🚀 Download e Conversão de Modelos NLP para NEXS-MCP")
    print("=" * 70)
    print()

    # Verificar dependências
    print("🔍 Verificando dependências...")
    check_dependencies()
    print("✅ Todas as dependências instaladas\n")

    # Download dos modelos
    results = {}

    # BERT NER
    try:
        results['bert_ner'] = download_and_convert_bert_ner()
    except Exception as e:
        print(f"❌ Erro ao processar BERT NER: {e}")
        results['bert_ner'] = False

    # DistilBERT Sentiment
    try:
        results['distilbert'] = download_and_convert_distilbert_sentiment()
    except Exception as e:
        print(f"❌ Erro ao processar DistilBERT Sentiment: {e}")
        results['distilbert'] = False

    # Resumo
    print("\n" + "=" * 70)
    print("📊 RESUMO")
    print("=" * 70)
    print(f"BERT NER:              {'✅ Sucesso' if results['bert_ner'] else '❌ Falhou'}")
    print(f"DistilBERT Sentiment:  {'✅ Sucesso' if results['distilbert'] else '❌ Falhou'}")
    print()

    if all(results.values()):
        print("✅ Download e conversão concluídos com sucesso!")
        print()
        print("📋 Próximos passos:")
        print()
        print("1. Verifique os modelos:")
        print("   ls -lh models/bert-base-ner/")
        print("   ls -lh models/distilbert-sentiment/")
        print()
        print("2. Configure o .env:")
        print("   NEXS_NLP_ENTITY_EXTRACTION_ENABLED=true")
        print("   NEXS_NLP_SENTIMENT_ENABLED=true")
        print()
        print("3. Compile com suporte ONNX:")
        print("   make build-onnx")
        print()
        print("4. Execute o NEXS-MCP:")
        print("   ./bin/nexs-mcp")
        print()
    else:
        print("⚠️  Alguns modelos falharam no download.")
        print("   Consulte os logs acima para detalhes.")
        print()
        print("💡 Dica: Tente executar novamente ou consulte:")
        print("   docs/DOWNLOAD_NLP_MODELS.md")
        sys.exit(1)

if __name__ == "__main__":
    main()
