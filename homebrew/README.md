# Homebrew Tap for NEXS-MCP

Este é o Homebrew Tap oficial para o NEXS-MCP Server.

## Instalação

```bash
# Adicionar o tap
brew tap fsvxavier/nexs-mcp

# Instalar o NEXS-MCP
brew install nexs-mcp
```

## Verificação

```bash
# Verificar versão instalada
nexs-mcp --version

# Ver help
nexs-mcp --help
```

## Atualização

```bash
# Atualizar lista de formulas
brew update

# Atualizar NEXS-MCP
brew upgrade nexs-mcp
```

## Desinstalação

```bash
# Remover NEXS-MCP
brew uninstall nexs-mcp

# Remover o tap (opcional)
brew untap fsvxavier/nexs-mcp
```

## Integração com Claude Desktop

Após a instalação, adicione ao seu `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "nexs-mcp": {
      "command": "/usr/local/bin/nexs-mcp"
    }
  }
}
```

Localizações do arquivo de configuração:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

## Diretórios

Após a instalação, os seguintes diretórios são criados:

- **Data**: `/usr/local/var/nexs-mcp/data/`
- **Config**: `/usr/local/etc/nexs-mcp/`
- **Auth**: `~/.nexs-mcp/auth/`

## Troubleshooting

### Permissões

Se encontrar problemas de permissão:

```bash
# Corrigir permissões do binário
chmod +x /usr/local/bin/nexs-mcp

# Criar diretórios se necessário
mkdir -p ~/.nexs-mcp/auth
chmod 700 ~/.nexs-mcp/auth
```

### Reinstalação Limpa

```bash
# Desinstalar completamente
brew uninstall nexs-mcp

# Limpar cache
brew cleanup

# Reinstalar
brew install nexs-mcp
```

### Logs

```bash
# Ver logs de instalação
brew info nexs-mcp

# Verificar fórmula
brew audit nexs-mcp
```

## Desenvolvimento

Para testar a fórmula localmente:

```bash
# Clone o tap
git clone https://github.com/fsvxavier/homebrew-nexs-mcp.git

# Teste a fórmula
cd homebrew-nexs-mcp
brew install --build-from-source Formula/nexs-mcp.rb

# Ou instale diretamente
brew install ./Formula/nexs-mcp.rb
```

## Suporte

- **Issues**: https://github.com/fsvxavier/nexs-mcp/issues
- **Discussions**: https://github.com/fsvxavier/nexs-mcp/discussions
- **Main Repository**: https://github.com/fsvxavier/nexs-mcp

## Licença

MIT License - veja [LICENSE](../LICENSE) no repositório principal.
