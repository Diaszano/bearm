# Instalação e configuração

Este guia descreve como instalar, compilar e configurar o Bearm em plataformas
Unix com suporte.

## Plataformas e requisitos

- **Sistemas operacionais**: Linux (kernel 3.x ou superior), macOS 11 ou
  superior.
- **Arquiteturas**: `amd64` (x86_64), `arm64` (Apple Silicon e Linux ARM de
  64 bits).
- **Go**: versão **1.27.3 ou superior** para compilar ou instalar com
  `go install`.

> [!NOTE]
> Windows não é compatível. O Bearm depende de semânticas de sistemas de
> arquivos Unix, da especificação FreeDesktop Trash no Linux e das pastas da
> Lixeira do macOS.

## Métodos de instalação

O Bearm está disponível via `go install`, Homebrew pelo tap oficial do projeto
e compilação direta do código-fonte.

### Método 1: `go install` (recomendado)

```bash
go install github.com/Diaszano/bearm/cmd/bearm@latest
```

O Go instala o binário em `$(go env GOPATH)/bin` (geralmente `$HOME/go/bin`).
Inclua essa pasta no `PATH`.

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
bearm version
```

### Método 2: Homebrew

```bash
brew install Diaszano/tap/bearm
```

### Método 3: compilar do código-fonte

1. Clone o repositório:

   ```bash
   git clone https://github.com/Diaszano/bearm.git
   cd bearm
   ```

2. Compile com Make:

   ```bash
   make build
   ```

   Ou use o comando Go diretamente:

   ```bash
   go build -trimpath -o bin/bearm ./cmd/bearm
   ```

3. Instale no diretório local de binários:

   ```bash
   mkdir -p "$HOME/.local/bin"
   install -m 0755 bin/bearm "$HOME/.local/bin/bearm"
   ```

   Inclua `$HOME/.local/bin` no `PATH`. Para instalação global, use
   `sudo install -m 0755 bin/bearm /usr/local/bin/bearm`.

4. Confira a instalação:

   ```bash
   bearm version
   ```

## Conferir a instalação

`bearm version` exibe metadados do build. Você também pode consultar e validar
os caminhos de configuração:

```bash
bearm config path
bearm config check
```

## Integração com o shell: alias de `rm`

Para usar o Bearm como alternativa reversível a `rm`, adicione ao arquivo de
configuração do seu shell:

```bash
alias rm='bearm rm'
```

No Bash, edite `~/.bashrc`; no Zsh, `~/.zshrc`; no Fish, adicione a linha a
`~/.config/fish/config.fish`. Nunca substitua, sobrescreva ou remova
`/bin/rm`: scripts do sistema e gerenciadores de pacotes podem depender do
comportamento destrutivo original.

## Ignorar o alias

Quando precisar explicitamente do `rm` destrutivo do sistema:

```bash
\rm arquivo-temporario
command rm arquivo-temporario
/bin/rm arquivo-temporario
```

## Atualizar o Bearm

Instalação via Go:

```bash
go install github.com/Diaszano/bearm/cmd/bearm@latest
```

Instalação via Homebrew:

```bash
brew upgrade bearm
```

Instalação a partir do código-fonte:

```bash
cd bearm
git pull origin main
make build
install -m 0755 bin/bearm "$HOME/.local/bin/bearm"
```

## Desinstalar

1. Remova o binário (`command -v bearm` mostra sua localização):

   ```bash
   rm -f "$(command -v bearm)"
   ```

2. Remova o alias do arquivo de configuração do shell.
3. Opcionalmente, remova configuração e estado. No Linux, os locais seguem
   `$XDG_CONFIG_HOME`, `$XDG_STATE_HOME` e `$XDG_DATA_HOME`, com alternativas
   em `$HOME/.config`, `$HOME/.local/state` e `$HOME/.local/share`, todos sob
   a pasta `bearm`. No macOS, os dados ficam em
   `$HOME/Library/Application Support/Bearm`.
