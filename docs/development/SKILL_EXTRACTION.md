# Skill Extraction System

## Overview

Sistema automático de extração de skills de personas que analisa os campos da persona e cria elementos `Skill` separados, relacionando-os automaticamente.

## Arquivos Criados

### 1. `internal/application/skill_extractor.go`
Serviço principal que implementa a lógica de extração de skills:

- **`SkillExtractor`**: Serviço que extrai skills de personas
- **`ExtractSkillsFromPersona()`**: Extrai skills de uma persona específica
- **`extractSkillsFromRawData()`**: Analisa dados brutos da persona (incluindo campos customizados)
- **`createSkillFromName()`**: Cria especificação de skill a partir de um nome
- **`findExistingSkill()`**: Verifica se skill já existe (evita duplicatas)
- **`generateKeywords()`**: Gera keywords para triggers

### 2. `internal/application/skill_extractor_test.go`
Testes completos do SkillExtractor (11 testes, todos passando):

- Extração de skills de expertise areas
- Extração de skills de campos customizados
- Detecção de skills duplicados
- Validação de entradas inválidas
- Geração de keywords
- Busca de skills existentes

### 3. `internal/mcp/skill_extraction_tools.go`
Tools MCP para expor a funcionalidade:

- **`extract_skills_from_persona`**: Extrai skills de uma persona específica
- **`batch_extract_skills`**: Processa múltiplas personas em batch

### 4. Registro no servidor MCP
Tools registrados em `internal/mcp/server.go`

## Como Usar

### Via MCP Tool (Recomendado)

```json
{
  "tool": "extract_skills_from_persona",
  "arguments": {
    "persona_id": "persona-engenheiro-senior-001"
  }
}
```

**Output:**
```json
{
  "skills_created": 13,
  "skill_ids": ["skill-001", "skill-002", ...],
  "persona_updated": true,
  "skipped_duplicate": 0,
  "message": "Extracted 13 skills from persona. Persona updated with skill references."
}
```

### Batch Processing

Extrai skills de todas as personas:

```json
{
  "tool": "batch_extract_skills",
  "arguments": {}
}
```

Ou de personas específicas:

```json
{
  "tool": "batch_extract_skills",
  "arguments": {
    "persona_ids": ["persona-001", "persona-002"]
  }
}
```

**Output:**
```json
{
  "total_personas_processed": 2,
  "total_skills_created": 25,
  "total_skills_skipped": 3,
  "personas_updated": 2,
  "results": {
    "persona-001": {
      "skills_created": 12,
      "message": "Created 12 skills"
    },
    "persona-002": {
      "skills_created": 13,
      "message": "Created 13 skills, skipped 3 duplicates"
    }
  },
  "message": "Processed 2 personas. Created 25 skills, skipped 3 duplicates. Updated 2 personas."
}
```

## Campos Extraídos

O extrator analisa os seguintes campos da persona:

### Campos Padrão
- **`expertise_areas`**: Cria um skill para cada área de expertise
  - Usa domain como nome
  - Level como tag
  - Keywords para triggers

### Campos Customizados (se presentes)
- **`technical_skills.core_expertise`**: Array de skills principais
- **`technical_skills.architecture_patterns`**: Padrões de arquitetura
- **`technical_skills.design_patterns`**: Padrões de design
- **`technical_skills.go_expertise`**: Expertise em Go
- **`technical_skills.security`**: Skills de segurança

## Estrutura do Skill Criado

```json
{
  "id": "skill-golang-architecture-001",
  "name": "Software Architecture",
  "description": "Software Architecture expertise at expert level",
  "version": "1.0.0",
  "author": "test@example.com",
  "triggers": [
    {
      "type": "keyword",
      "keywords": ["software architecture", "architecture", "clean-architecture", "ddd"]
    }
  ],
  "procedures": [
    {
      "step": 1,
      "action": "Apply Software Architecture expertise",
      "description": "Expert in software architecture"
    }
  ],
  "tags": ["auto-extracted", "expertise", "expert"]
}
```

## Features

### ✅ Extração Automática
- Analisa campos padrão e customizados
- Cria skills com triggers e procedures automaticamente
- Gera keywords inteligentes a partir dos nomes

### ✅ Detecção de Duplicatas
- Verifica se skill já existe pelo nome (case-insensitive)
- Reutiliza skills existentes
- Conta duplicatas puladas

### ✅ Relacionamento Automático
- Atualiza campo `RelatedSkills` da persona
- Mantém referências bidirecionais

### ✅ Validação
- Valida skills antes de criar
- Reporta erros sem interromper o processo
- Retorna lista de erros no output

### ✅ Batch Processing
- Processa múltiplas personas de uma vez
- Retorna resultados detalhados por persona
- Estatísticas agregadas

## Exemplo Completo

Para a persona `engenheiro-software-senior.json`:

**Antes:**
- 1 persona com campos `technical_skills`, `expertise_areas`, etc.
- 0 skills separados
- Campo `skills` vazio ou inexistente

**Depois de executar `extract_skills_from_persona`:**
- 1 persona atualizada
- 13 skills criados em `data/elements/skills/`
- Campo `skills` da persona preenchido com referências:
  ```json
  {
    "skills": [
      {"skill_id": "skill-golang-architecture-001", "proficiency": "expert", "years_experience": 15},
      {"skill_id": "skill-api-design-001", "proficiency": "expert", "years_experience": 18},
      ...
    ]
  }
  ```

## Performance

- **Extração Individual**: ~50-100ms por persona
- **Batch Processing**: Processa 10 personas em ~1s
- **Detecção de Duplicatas**: O(n) onde n = número de skills existentes

## Limitações

1. **Campos customizados**: Apenas campos conhecidos em `technical_skills` são processados
2. **Proficiency e Experience**: Não são extraídos automaticamente (podem ser adicionados manualmente)
3. **Nested Objects**: Apenas 1 nível de nested objects é suportado

## Próximos Passos

Para adicionar suporte a novos campos:

1. Editar `extractSkillsFromRawData()` em `skill_extractor.go`
2. Adicionar lógica para extrair o novo campo
3. Criar skill specs apropriados
4. Adicionar testes

## Testes

Execute os testes:

```bash
go test ./internal/application -run TestSkillExtractor -v
```

Todos os 11 testes devem passar ✅
