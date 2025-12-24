# ONNX Runtime - Configuração Permanente de Variáveis de Ambiente

## Problema

Ao compilar nexs-mcp com suporte ONNX (`make build ONNX=1`), o build falha porque o linker não encontra a biblioteca `libonnxruntime.so`, mesmo que ela esteja instalada em `/usr/local/lib/`.

Isso ocorre porque:
1. O sistema não está procurando bibliotecas em `/usr/local/lib/` por padrão
2. As variáveis de ambiente `CGO_CFLAGS` e `CGO_LDFLAGS` não estão configuradas permanentemente
3. O cache do linker dinâmico (`ldconfig`) pode não incluir `/usr/local/lib/`

## Solução Permanente

### 1. Configurar ldconfig (Recomendado)

Configure o sistema para incluir `/usr/local/lib/` no cache do linker dinâmico:

```bash
# Criar arquivo de configuração para /usr/local/lib
echo "/usr/local/lib" | sudo tee /etc/ld.so.conf.d/usr-local-lib.conf

# Atualizar cache do linker
sudo ldconfig

# Verificar que ONNX Runtime está no cache
ldconfig -p | grep onnxruntime
```

**Resultado esperado:**
```
libonnxruntime.so.1 (libc6,x86-64) => /usr/local/lib/libonnxruntime.so.1
libonnxruntime.so (libc6,x86-64) => /usr/local/lib/libonnxruntime.so
```

### 2. Configurar Variáveis de Ambiente Permanentes (Opcional)

Se o método acima não funcionar, configure as variáveis no seu shell:

#### Para Bash (~/.bashrc)

```bash
# Adicionar ao final do arquivo ~/.bashrc
cat >> ~/.bashrc << 'EOF'

# ONNX Runtime Configuration
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib -lonnxruntime"
export LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"
EOF

# Recarregar configuração
source ~/.bashrc
```

#### Para Zsh (~/.zshrc)

```bash
# Adicionar ao final do arquivo ~/.zshrc
cat >> ~/.zshrc << 'EOF'

# ONNX Runtime Configuration
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib -lonnxruntime"
export LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"
EOF

# Recarregar configuração
source ~/.zshrc
```

#### Para Fish (~/.config/fish/config.fish)

```bash
# Adicionar ao final do arquivo ~/.config/fish/config.fish
cat >> ~/.config/fish/config.fish << 'EOF'

# ONNX Runtime Configuration
set -x CGO_CFLAGS "-I/usr/local/include"
set -x CGO_LDFLAGS "-L/usr/local/lib -lonnxruntime"
set -x LD_LIBRARY_PATH "/usr/local/lib:$LD_LIBRARY_PATH"
EOF

# Recarregar configuração
source ~/.config/fish/config.fish
```

### 3. Configuração do Sistema (Alternativa)

Para configurar para todos os usuários do sistema:

```bash
# Criar arquivo de variáveis de ambiente do sistema
sudo tee /etc/profile.d/onnxruntime.sh << 'EOF'
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib -lonnxruntime"
export LD_LIBRARY_PATH="/usr/local/lib:${LD_LIBRARY_PATH}"
EOF

# Tornar executável
sudo chmod +x /etc/profile.d/onnxruntime.sh

# Aplicar para a sessão atual
source /etc/profile.d/onnxruntime.sh
```

## Verificação da Configuração

### 1. Verificar Instalação do ONNX Runtime

```bash
# Verificar arquivos instalados
ls -la /usr/local/lib/libonnxruntime*
ls -la /usr/local/include/onnxruntime/

# Verificar cache do linker
ldconfig -p | grep onnxruntime
```

### 2. Verificar Variáveis de Ambiente

```bash
echo "CGO_CFLAGS: $CGO_CFLAGS"
echo "CGO_LDFLAGS: $CGO_LDFLAGS"
echo "LD_LIBRARY_PATH: $LD_LIBRARY_PATH"
```

### 3. Testar Build com ONNX

```bash
# Limpar builds anteriores
make clean

# Tentar build com ONNX
make build ONNX=1

# Verificar que o binário foi linkado com ONNX
ldd bin/nexs-mcp | grep onnxruntime
```

**Resultado esperado:**
```
libonnxruntime.so.1 => /usr/local/lib/libonnxruntime.so.1 (0x...)
```

### 4. Testar Execução

```bash
# Executar binário
./bin/nexs-mcp --help

# Verificar que ONNX está disponível (deve aparecer nos logs)
./bin/nexs-mcp 2>&1 | grep -i onnx
```

## Troubleshooting

### Erro: "cannot find -lonnxruntime"

**Causa:** O linker não encontra a biblioteca ONNX Runtime.

**Solução:**
```bash
# 1. Verificar instalação
ls -la /usr/local/lib/libonnxruntime*

# 2. Atualizar cache do linker
sudo ldconfig

# 3. Se ainda não funcionar, configurar LD_LIBRARY_PATH
export LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"
```

### Erro: "libonnxruntime.so.1: cannot open shared object file"

**Causa:** O binário foi compilado com sucesso, mas o sistema não encontra a biblioteca em runtime.

**Solução:**
```bash
# 1. Configurar LD_LIBRARY_PATH permanentemente
echo 'export LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"' >> ~/.bashrc
source ~/.bashrc

# 2. OU adicionar ao ldconfig
echo "/usr/local/lib" | sudo tee /etc/ld.so.conf.d/usr-local-lib.conf
sudo ldconfig
```

### Erro: "onnxruntime_c_api.h: No such file or directory"

**Causa:** Headers do ONNX Runtime não estão instalados ou CGO_CFLAGS não está configurado.

**Solução:**
```bash
# 1. Verificar headers
ls -la /usr/local/include/onnxruntime/

# 2. Configurar CGO_CFLAGS
export CGO_CFLAGS="-I/usr/local/include"

# 3. Se headers não existirem, reinstalar ONNX Runtime
make install-onnx
```

### Variáveis Não Persistem entre Sessões

**Causa:** As variáveis foram definidas apenas na sessão atual.

**Solução:**
- Adicionar as variáveis ao arquivo de configuração do seu shell (~/.bashrc, ~/.zshrc, etc.)
- OU usar a configuração do sistema (/etc/profile.d/onnxruntime.sh)
- Sempre executar `source` no arquivo após modificação

## Plataformas

### Linux

- ✅ Suportado
- Instalação: `/usr/local/lib/`
- Cache: `ldconfig`
- Variável: `LD_LIBRARY_PATH`

### macOS

- ✅ Suportado
- Instalação: `/usr/local/lib/`
- Cache: `update_dyld_shared_cache`
- Variável: `DYLD_LIBRARY_PATH` (use com cuidado, pode ser desabilitado pelo SIP)

### Windows

- ⚠️ Requer configuração manual
- Instalação: `C:\Program Files\onnxruntime\`
- Variável: `PATH`

## Comandos Úteis

```bash
# Verificar onde o sistema procura bibliotecas
ldconfig -v 2>/dev/null | grep /usr/local/lib

# Verificar bibliotecas linkadas em um binário
ldd bin/nexs-mcp

# Verificar símbolos ONNX no binário
nm bin/nexs-mcp | grep -i onnx

# Verificar tamanho do binário (com ONNX é ~1MB maior)
ls -lh bin/nexs-mcp

# Limpar e rebuildar
make clean && make build ONNX=1

# Build portátil (sem ONNX)
make clean && make build
```

## Recomendações

### Para Desenvolvimento

Use a **Solução 1 (ldconfig)** - É a mais limpa e não polui suas variáveis de ambiente.

### Para CI/CD

Use variáveis de ambiente diretas no comando:
```bash
CGO_ENABLED=1 \
CGO_CFLAGS="-I/usr/local/include" \
CGO_LDFLAGS="-L/usr/local/lib -lonnxruntime" \
go build -o nexs-mcp ./cmd/nexs-mcp
```

### Para Distribuição

**Preferir builds portáteis** (sem ONNX) para facilitar distribuição:
```bash
make build        # Build portátil (padrão)
make build-all    # Multi-plataforma portátil
```

Para versões ONNX:
```bash
make build ONNX=1      # Build local com ONNX
make build-all ONNX=1  # Multi-plataforma com ONNX
```

## Referências

- [ONNX Runtime Releases](https://github.com/microsoft/onnxruntime/releases)
- [Linux Shared Libraries](https://tldp.org/HOWTO/Program-Library-HOWTO/shared-libraries.html)
- [ldconfig Manual](https://man7.org/linux/man-pages/man8/ldconfig.8.html)
- [CGO Environment Variables](https://pkg.go.dev/cmd/cgo)
