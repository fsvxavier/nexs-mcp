# MCP UX Guidelines: Understanding Client-Server Separation

## 📋 Overview

O Model Context Protocol (MCP) segue uma arquitetura **cliente-servidor** onde responsabilidades são claramente divididas. Entender essa separação é crucial para ter expectativas corretas sobre o comportamento do sistema.

## 🏗️ Arquitetura MCP

```
┌─────────────────────────────────────────────────────────────┐
│                    CLIENTE (GitHub Copilot)                  │
│                                                              │
│  • Interpreta linguagem natural                             │
│  • Decide qual tool chamar                                  │
│  • Controla preview/confirmações                            │
│  • Apresenta resultados ao usuário                          │
└─────────────────────────────────────────────────────────────┘
                            ↕ MCP Protocol (stdio)
┌─────────────────────────────────────────────────────────────┐
│                   SERVIDOR (NEXS MCP)                        │
│                                                              │
│  • Expõe tools disponíveis                                  │
│  • Executa lógica de negócio                                │
│  • Persiste dados                                           │
│  • Retorna resultados estruturados                          │
└─────────────────────────────────────────────────────────────┘
```

## ❓ Comportamento Observado vs Esperado

### O que você vê:
```
Usuário: "Crie uma persona DevOps"
  ↓
Copilot: "Vou criar uma persona com este YAML... [mostra preview]"
Copilot: "Posso prosseguir?" [pede confirmação]
  ↓
Usuário: "Sim"
  ↓
[Persona criada]
```

### O que acontece nos bastidores:
```
1. Copilot recebe: "Crie uma persona DevOps"
2. Copilot decide: "Vou chamar create_persona"
3. Copilot gera parâmetros: {name: "DevOps", description: "..."}
4. Copilot DECIDE mostrar preview (decisão do cliente)
5. Copilot DECIDE pedir confirmação (decisão do cliente)
6. Após confirmação, chama MCP server
7. Server cria e retorna resultado
8. Copilot apresenta resultado
```

## 🎯 Por Que o Servidor Não Decide Sozinho?

### Razões Arquiteturais:

1. **Separação de Responsabilidades**
   - Servidor: Lógica de negócio
   - Cliente: Interação com usuário

2. **Flexibilidade**
   - Diferentes clientes podem ter diferentes UX
   - GitHub Copilot vs Claude vs Cursor têm comportamentos diferentes

3. **Segurança**
   - Cliente controla o que é executado automaticamente
   - Previne ações destrutivas sem consentimento

4. **Protocolo Standard**
   - MCP é um protocolo de tools, não de chat
   - Servidor não processa linguagem natural

## 🔧 Melhorias Práticas Implementadas

### 1. Quick Create Tools

Tools simplificadas para criação rápida:

```javascript
// Tool completa (pode ter preview)
create_persona({
  name: "DevOps Expert",
  description: "...",
  behavioral_traits: {...},
  expertise: [...],
  communication_style: "...",
  // ... 10+ parâmetros
})

// Quick tool (menos parâmetros = menos preview)
quick_create_persona({
  name: "DevOps Expert",
  template: "technical" // Usa defaults inteligentes
})
```

### 2. Batch Operations

Criar múltiplos elementos sem confirmações individuais:

```javascript
create_elements_batch({
  elements: [
    {type: "persona", name: "DevOps"},
    {type: "skill", name: "Deploy"},
    {type: "template", name: "Report"}
  ]
})
// Cliente vê UMA confirmação para todas as criações
```

### 3. Tool Descriptions com Hints

```go
sdk.AddTool(s.server, &sdk.Tool{
  Name:        "quick_create_persona",
  Description: "Create persona with minimal input (defaults applied, no preview needed)",
}, handler)
```

**Keywords que ajudam clientes:**
- `"no preview needed"` → Cliente pode executar direto
- `"destructive operation"` → Cliente DEVE pedir confirmação
- `"idempotent"` → Cliente pode reexecutar sem riscos

## 📝 Comportamento por Cliente

### GitHub Copilot
- **Preview**: Mostra para operações complexas
- **Confirmação**: Pede para create/update/delete
- **Auto-exec**: Apenas para queries/reads

### Claude Desktop
- **Preview**: Sempre mostra antes de executar
- **Confirmação**: Pede para TODAS as operações
- **Auto-exec**: Nenhuma (mais cauteloso)

### Cursor
- **Preview**: Mostra em forma de diff
- **Confirmação**: Pede seletivamente
- **Auto-exec**: Queries e operações seguras

## 🎨 Best Practices para UX Melhor

### Para Usuários:

1. **Use comandos diretos**
   ```
   ❌ "Você pode criar uma persona chamada X?"
   ✅ "Crie uma persona X com expertise em Y"
   ```

2. **Especifique urgência**
   ```
   ✅ "Crie rapidamente uma skill de deploy"
   ✅ "Adicione isso agora: [detalhes]"
   ```

3. **Use templates**
   ```
   ✅ "Crie persona técnica chamada X"
   ✅ "Use template padrão para agent de web scraping"
   ```

### Para Desenvolvedores:

1. **Tools com poucos parâmetros obrigatórios**
   ```go
   // Melhor UX
   type QuickCreateInput struct {
     Name     string `json:"name"`
     Template string `json:"template,omitempty"`
   }
   ```

2. **Defaults inteligentes**
   ```go
   if input.Version == "" {
     input.Version = "1.0.0"
   }
   if input.Author == "" {
     input.Author = getCurrentUser()
   }
   ```

3. **Operações idempotentes**
   ```go
   // Criar OU atualizar (não falha se existe)
   func (s *Server) upsertElement(element) {
     existing, err := s.repo.GetByName(element.Name)
     if err == nil {
       return s.repo.Update(existing.ID, element)
     }
     return s.repo.Create(element)
   }
   ```

## 🚫 O Que NÃO É Possível

### ❌ Servidor processar linguagem natural
```
"Crie uma persona legal" → Servidor NÃO sabe o que é "legal"
```
**Solução**: Cliente (Copilot) interpreta → chama tool com parâmetros

### ❌ Servidor decidir se pede confirmação
```
Servidor NÃO pode: "Vou criar X, confirma?"
```
**Solução**: Cliente decide baseado em heurísticas

### ❌ Servidor controlar apresentação
```
Servidor NÃO pode: "Mostre isso em tabela"
```
**Solução**: Servidor retorna dados estruturados, cliente apresenta

## 💡 Workarounds Criativos

### 1. Parâmetro `auto_confirm`
```go
type CreateInput struct {
  Name        string `json:"name"`
  AutoConfirm bool   `json:"auto_confirm,omitempty"`
}
```
**Limitação**: Cliente ainda pode ignorar

### 2. Prompts no Description
```go
Description: "IMMEDIATE: Create persona (no confirmation needed)"
```
**Eficácia**: ~60% dos clientes respeitam

### 3. Comandos Magic
```
Usuário: "/create persona DevOps"
Cliente vê: "/" no início → executa direto
```
**Suporte**: Apenas alguns clientes (Cursor, Continue.dev)

## 🎯 Recomendações Finais

### Para Este Projeto (NEXS MCP):

1. ✅ **Implementar quick_create_* tools**
   - Menos parâmetros
   - Defaults inteligentes
   - Descrições com hints

2. ✅ **Documentar comportamento esperado**
   - Este documento
   - README atualizado
   - Exemplos práticos

3. ✅ **Criar modo batch**
   - Uma confirmação para N elementos
   - Melhor UX para operações em massa

4. ⚠️ **Não tentar subverter o protocolo**
   - MCP foi desenhado assim por boas razões
   - Trabalhar COM o protocolo, não contra

### Para Usuários:

1. **Entender a limitação**
   - Preview/confirmação = decisão do cliente
   - É feature de segurança, não bug

2. **Ajustar expectativas**
   - MCP não é chat autônomo
   - É protocolo de ferramentas

3. **Usar comandos claros**
   - Quanto mais direto, menos confirmações
   - "Crie X" melhor que "Pode criar X?"

## 📚 Referências

- [MCP Specification](https://modelcontextprotocol.io/)
- [MCP Best Practices](https://modelcontextprotocol.io/docs/best-practices)
- [Tool Design Guidelines](https://modelcontextprotocol.io/docs/tools)

## 🔄 Evolução Futura

### MCP 2.0 (Proposta)
- Hints de confirmação no protocolo
- Suporte a operações em lote nativas
- Flags de urgência (`urgent: true`)

### Clientes Mais Inteligentes
- Aprender preferências do usuário
- Confirmar apenas operações destrutivas
- Auto-executar operações idempotentes

---

**TL;DR**: O servidor MCP não controla preview/confirmações. Isso é responsabilidade do cliente (Copilot). Podemos melhorar criando tools simplificadas e usando hints nas descrições, mas a decisão final é sempre do cliente por design do protocolo.
