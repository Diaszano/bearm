<div align="center">

# 🐻 Bearm

**Remova com segurança, restaure com confiança.**

Substituto nativo, performático e resiliente a falhas para o `rm` do Unix, escrito em Go.<br>
Move arquivos, diretórios e links simbólicos para a lixeira nativa do seu sistema operacional em vez de destruí-los permanentemente.

[![Go Version](https://img.shields.io/github/go-mod/go-version/Diaszano/bearm?style=flat-square&logo=go)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Platform Support](https://img.shields.io/badge/plataforma-Linux%20%7C%20macOS-informational?style=flat-square)]()
[![Architecture](https://img.shields.io/badge/arch-amd64%20%7C%20arm64-informational?style=flat-square)]()
[![Pure Go](https://img.shields.io/badge/dependências-zero%20externas-brightgreen?style=flat-square)]()
[![PRs Welcome](https://img.shields.io/badge/PRs-bem--vindos-brightgreen.svg?style=flat-square)](CONTRIBUTING.md)

---

**Idiomas:** [🇺🇸 English](README.md) • 🇧🇷 **Português**

</div>

---

## 📖 Sumário

- [Por que o Bearm?](#-por-que-o-bearm)
- [Tabela Comparativa](#-tabela-comparativa)
- [Recursos Principais](#-recursos-principais)
- [Plataformas Suportadas](#-plataformas-suportadas)
- [Instalação](#-instalação)
  - [Método 1: via `go install` (Recomendado)](#método-1-via-go-install-recomendado)
  - [Método 2: Compilar a partir do Código-Fonte](#método-2-compilar-a-partir-do-código-fonte)
- [Primeiro Uso Seguro (Teste em Sandbox)](#-primeiro-uso-seguro-teste-em-sandbox)
- [Uso no Dia a Dia](#-uso-no-dia-a-dia)
  - [1. Usando o Bearm Diretamente](#1-usando-o-bearm-diretamente)
  - [2. Configurando o Alias para `rm`](#2-configurando-o-alias-para-rm)
  - [3. Ignorando o Alias Quando Necessário](#3-ignorando-o-alias-quando-necessário)
- [Recuperação e Subcomandos Nativos](#-recuperação-e-subcomandos-nativos)
  - [Listando Arquivos na Lixeira (`list`)](#listando-arquivos-na-lixeira-list)
  - [Restaurando Arquivos (`restore`)](#restaurando-arquivos-restore)
  - [Exclusão Permanente (`purge`)](#exclusão-permanente-purge)
  - [Diagnóstico e Integridade do Histórico (`doctor`)](#diagnóstico-e-integridade-do-histórico-doctor)
  - [Validação de Configuração (`config`)](#validação-de-configuração-config)
- [Garantias de Segurança](#-garantias-de-segurança)
- [Configuração](#-configuração)
- [Documentação Detalhada](#-documentação-detalhada)
- [Contribuindo](#-contribuindo)
- [Licença](#-licença)

---

## 💡 Por que o Bearm?

Todo desenvolvedor ou administrador de sistemas já sentiu o frio na espinha ao executar um `rm -rf` por engano. O utilitário padrão `rm(1)` do Unix remove diretamente os inodes no sistema de arquivos: uma vez confirmada a exclusão, os dados são perdidos sem chance de recuperação simples.

Outras ferramentas de lixeira costumam exigir interpretadores pesados (como Python), utilizam flags incompatíveis com o `rm`, falham entre diferentes sistemas de arquivos (mounts) ou carecem de testes diferenciais rigorosos.

**O Bearm resolve isso de ponta a ponta:**
1. **Compatibilidade total com flags do `rm`**: Mantém sua memória muscular intacta no terminal, suportando todas as opções GNU e BSD (`-r`, `-f`, `-i`, `-v`, `-d`, etc.).
2. **Integração com a Lixeira Nativa**: Segue o padrão FreeDesktop Trash no Linux e utiliza as pastas de lixeira visíveis do macOS.
3. **Movimentação atômica O(1) de diretórios**: Move pastas inteiras com um único `rename` no filesystem, sem precisar varrer recursivamente milhares de arquivos.
4. **Proteções rígidas de segurança**: Recusa categoricamente apagar caminhos críticos como `/`, `.`, `..`, pastas de lixeira ativas e arquivos de configuração.
5. **Restauração instantânea**: Apagou sem querer? Basta rodar `bearm restore --last` para recuperar tudo imediatamente.
6. **Desempenho puro e zero telemetria**: Binário único e estático em Go, sem processos em segundo plano (daemons), sem telemetria e sem chamadas de rede.

---

## 📊 Tabela Comparativa

| Recurso | `/bin/rm` Padrão | `trash-cli` (Python) | `rmtrash` (Shell) | **Bearm** (Go) |
| :--- | :---: | :---: | :---: | :---: |
| **Recuperação / Desfazer** | ❌ Não (Permanente) | ✅ Sim | ✅ Sim | 🛡️ **Sim (`restore --last`)** |
| **Compatibilidade de Flags `rm`** | ✅ Nativo | ❌ Flags incompatíveis | ⚠️ Parcial | 🎯 **Total (GNU & BSD)** |
| **Caminho Rápido para Pastas** | N/A | ❌ Varrredura recursiva | ❌ Varredura recursiva | ⚡ **Atômico O(1)** |
| **Proteção Rígida da Raiz** | ⚠️ `--preserve-root` | ⚠️ Inconsistente | ❌ Mínima | 🔒 **Estrita (`/`, `..`, lixeira, config)** |
| **Padrão de Lixeira do SO** | ❌ Não usa | ⚠️ Apenas Linux | ⚠️ Apenas macOS | 🍏🐧 **Nativo no Linux e macOS** |
| **Diário de Auditoria (Journal)** | ❌ Nenhum | ⚠️ Metadados básicos | ❌ Nenhum | 📜 **Append-only em JSONL** |
| **Consumo e Dependências** | Baixo (C) | Alto (Python) | Alto (Subshells) | ⚡ **Binário único estático** |
| **Telemetria e Rede** | Nenhuma | Nenhuma | Nenhuma | 🚫 **Zero telemetria e 100% offline** |

---

## ✨ Recursos Principais

- **Perfis de Compatibilidade**: Seleção automática de acordo com o sistema operacional (perfil GNU no Linux, perfil BSD no macOS, além de POSIX).
- **Movimentação Atômica no Mesmo Filesystem**: Move diretórios em tempo constante O(1) sem percorrer os itens filhos, exceto quando inspeção de padrões é solicitada.
- **Lixeira Nativa do SO**: Suporte a FreeDesktop Trash no Linux (`~/.local/share/Trash` e `.Trash/<uid>` por partição) e lixeira nativa no macOS (`~/.Trash`).
- **Reserva Atômica Anti-Colisão**: IDs únicos e criação exclusiva evitam que arquivos com o mesmo nome sobrescrevam itens existentes na lixeira.
- **Proteções Rígidas Invioláveis**: Bloqueia remoção de `/`, `.`, `..`, equivalentes léxicos e raízes de dados/configuração do Bearm.
- **Diário Append-Only (JSONL)**: Todas as ações (`trashed`, `restored`, `purged`) são registradas com precisão de nanossegundos e validação contra falhas de energia.
- **Configuração Declarativa em TOML**: Carregamento estrito sem execução de comandos de shell ou riscos de injeção.
- **Privacidade Absoluta**: Zero telemetria, zero requisições de rede e zero execução de subshells em tempo de execução.

---

## 💻 Plataformas Suportadas

| Sistema Operacional | Arquitetura | Requisito do Go | Status |
| :--- | :--- | :--- | :--- |
| **Linux** | `amd64` (x86_64) | Go 1.24+ | Suporte Total (FreeDesktop Trash) |
| **Linux** | `arm64` (AArch64) | Go 1.24+ | Suporte Total (FreeDesktop Trash) |
| **macOS** | `amd64` (Intel) | Go 1.24+ | Suporte Total (Lixeira do Sistema) |
| **macOS** | `arm64` (Apple Silicon) | Go 1.24+ | Suporte Total (Lixeira do Sistema) |

> [!NOTE]
> O Windows não é suportado. No macOS, metadados de **Colocar de Volta** do Finder não são garantidos; utilize o comando `bearm restore` como autoridade de recuperação.

---

## 📦 Instalação

O Bearm é distribuído diretamente pelas ferramentas oficiais da linguagem Go ou por compilação local a partir do código-fonte. Não há necessidade de gerenciadores de pacotes de terceiros (como Homebrew).

### Método 1: via `go install` (Recomendado)

Se você já possui o **Go 1.24 ou superior** instalado em sua máquina, instale a versão mais recente com um único comando:

```bash
go install github.com/Diaszano/bearm/cmd/bearm@latest
```

Certifique-se de que a pasta de binários do Go (geralmente `~/go/bin` ou `$(go env GOPATH)/bin`) está incluída no seu `$PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Valide a instalação:

```bash
bearm version
```

---

### Método 2: Compilar a partir do Código-Fonte

Compile o binário diretamente no seu computador para ter controle total:

```bash
# 1. Clone o repositório
git clone https://github.com/Diaszano/bearm.git
cd bearm

# 2. Compile o binário
make build
# Ou diretamente via go build (caso não utilize make):
# go build -trimpath -o bin/bearm ./cmd/bearm

# 3. Instale o executável no seu diretório local de binários
mkdir -p "$HOME/.local/bin"
install -m 0755 bin/bearm "$HOME/.local/bin/bearm"
```

Certifique-se de que `$HOME/.local/bin` está presente no seu `$PATH`.

Para mais detalhes sobre configuração do ambiente e resolução de caminhos, consulte o [Guia de Instalação (em inglês)](docs/installation.md).

---

## 🧪 Primeiro Uso Seguro (Teste em Sandbox)

Antes de utilizar o Bearm com seus arquivos de trabalho, você pode executar este teste seguro em um diretório descartável:

```bash
# 1. Cria um diretório de testes isolado
trial="$(mktemp -d)"
mkdir -p "$trial/home" "$trial/work"
echo "Conteúdo de teste seguro" > "$trial/work/exemplo.txt"

# 2. Executa o bearm rm apontando para a sandbox
HOME="$trial/home" \
BEARM_STATE_HOME="$trial/state" \
BEARM_CONFIG_HOME="$trial/config" \
BEARM_TRASH="$trial/trash" \
bearm rm "$trial/work/exemplo.txt"

# Confirma que o arquivo foi movido para a lixeira
ls -la "$trial/work/exemplo.txt" 2>/dev/null || echo "Arquivo enviado com segurança para a lixeira!"

# 3. Restaura o arquivo imediatamente
HOME="$trial/home" \
BEARM_STATE_HOME="$trial/state" \
BEARM_CONFIG_HOME="$trial/config" \
BEARM_TRASH="$trial/trash" \
bearm restore --last

# 4. Confirma que o conteúdo está de volta
cat "$trial/work/exemplo.txt"

# 5. Remove o diretório de testes
rm -rf "$trial"
```

---

## 🚀 Uso no Dia a Dia

### 1. Usando o Bearm Diretamente

Você pode invocar `bearm rm` utilizando exatamente a mesma sintaxe do `rm` tradicional:

```bash
# Remove um ou mais arquivos
bearm rm arquivo.txt notas.md

# Remove diretórios recursivamente
bearm rm -r build/
bearm rm -rf cache_temporario/

# Modo detalhado (mostra cada item movido)
bearm rm -v documento.pdf

# Remoção segura de arquivos que iniciam com hífen
bearm rm -- -arquivo-com-hifen.txt
```

### 2. Configurando o Alias para `rm`

Para proteger seu terminal de exclusões acidentais de forma transparente, configure um alias no seu shell:

#### Bash (`~/.bashrc` ou `~/.bash_profile`):
```bash
alias rm='bearm rm'
```

#### Zsh (`~/.zshrc`):
```zsh
alias rm='bearm rm'
```

#### Fish (`~/.config/fish/config.fish`):
```fish
alias rm='bearm rm'
```

> [!IMPORTANT]
> **Nunca remova ou sobrescreva `/bin/rm`**. Utilize sempre um alias de shell no nível do usuário. Scripts do sistema operacional e gerenciadores de pacotes dependem do comportamento destrutivo de `/bin/rm`.

### 3. Ignorando o Alias Quando Necessário

Se em algum momento específico você precisar da exclusão destrutiva original:

```bash
\rm arquivo_temporario.tmp       # Prefixo com barra invertida
command rm arquivo_temporario.tmp # Usando o comando interno do shell
/bin/rm arquivo_temporario.tmp    # Caminho absoluto do executável
```

---

## ♻️ Recuperação e Subcomandos Nativos

O Bearm inclui ferramentas nativas para gerenciar a lixeira e auditar o histórico de operações.

### Listando Arquivos na Lixeira (`list`)

Exibe os arquivos atualmente na lixeira que foram removidos pelo Bearm:

```bash
bearm list
bearm list --limit 20
bearm list --json
```

Exemplo de saída:
```text
ITEM_ID           DATA_DELECAO                CAMINHO_ORIGINAL
itm_01j7abc123    2026-09-14T21:30:00-03:00   /home/usuario/projeto/main.go
```

### Restaurando Arquivos (`restore`)

Restaura itens de volta ao caminho original em que estavam:

```bash
# Desfaz a operação de remoção mais recente
bearm restore --last

# Restaura um item específico pelo ID
bearm restore itm_01j7abc123

# Restaura todos os itens pertencentes a uma operação em lote
bearm restore --operation op_01j7abc999
```

> [!TIP]
> Se o arquivo original já tiver sido recriado no destino, a política padrão (`fail`) interrompe a restauração para evitar perda de dados. Você pode configurar `restore.collision_policy = "rename"` no seu `config.toml` para que o arquivo seja restaurado com o sufixo `.restored.N`.

### Exclusão Permanente (`purge`)

O comando `bearm purge` é a **única** forma de apagar permanentemente itens gerenciados pelo Bearm:

```bash
# Exibe confirmação antes da exclusão permanente
bearm purge itm_01j7abc123

# Pula a confirmação interativa com --yes
bearm purge itm_01j7abc123 --yes

# Limpa permanentemente todos os itens da última operação
bearm purge --last --yes
```

### Diagnóstico e Integridade do Histórico (`doctor`)

Verifica a integridade do diário append-only e valida se os registros correspondem aos arquivos físicos na lixeira:

```bash
bearm doctor
bearm doctor --json
```

### Validação de Configuração (`config`)

Inspeciona e valida o arquivo de configuração:

```bash
bearm config path   # Exibe o caminho do arquivo config.toml ativo
bearm config check  # Valida sintaxe e permissões de segurança
```

---

## 🛡️ Garantias de Segurança

O Bearm foi projetado com práticas defensivas em cada camada de execução:

- **Fronteiras Léxicas Rígidas**: Bloqueia tentativas de exclusão de `/`, `.`, `..`, caminhos relativos equivalentes e diretórios internos do Bearm.
- **Tratamento Seguro de Links Simbólicos**: Um symlink é movido *como link simbólico*; seu alvo de destino nunca é seguido ou modificado.
- **Reserva Atômica de Metadados**: Se a movimentação falhar no meio do caminho, as reservas são revertidas e os arquivos de origem permanecem intactos.
- **Imunidade a Injeção de Comandos**: As chamadas ao sistema de arquivos são feitas diretamente pelas APIs de baixo nível do Go, sem passar pelo shell.
- **Resiliência a Falhas de Energia**: O diário JSONL utiliza travas exclusivas de arquivo e chamadas `fsync`, garantindo que uma queda de energia não corrompa o histórico.

Para entender o modelo de ameaças completo, consulte o documento [Security Model (em inglês)](docs/security.md).

---

## ⚙️ Configuração

O Bearm funciona perfeitamente sem qualquer arquivo de configuração. Caso deseje personalizar o comportamento, utilize o arquivo declarativo TOML:

- **Linux**: `$XDG_CONFIG_HOME/bearm/config.toml` (ou `~/.config/bearm/config.toml`)
- **macOS**: `~/Library/Application Support/Bearm/config.toml`

### Exemplo de `config.toml`:

```toml
language = "pt-BR"            # "pt-BR" ou "en" para mensagens dos comandos nativos
compatibility_profile = "auto" # "auto", "gnu", "bsd" ou "posix"

[trash]
per_mount = true              # Linux: usa lixeira no mesmo filesystem quando disponível
custom_path = ""              # Caminho absoluto opcional para lixeira personalizada

[safety]
preserve_root = true          # Exige preservação rígida da raiz
allowed_roots = []            # Lista opcional de diretórios absolutos autorizados para remoção
protected_patterns = [        # Padrões glob que o Bearm se recusará estritamente a apagar
  "**/.git/**",
  "**/production.env"
]
inspect_descendants = false   # Inspeciona itens filhos em diretórios recursivos

[restore]
collision_policy = "fail"     # "fail", "rename" ou "overwrite"
```

### Principais Variáveis de Ambiente:

| Variável | Descrição |
| :--- | :--- |
| `BEARM_COMPAT` | Força o perfil de compatibilidade (`gnu`, `bsd`, `posix`) |
| `BEARM_LANG` | Idioma de saída dos comandos nativos (`pt-BR`, `en`) |
| `BEARM_CONFIG_HOME`| Caminho absoluto para o diretório de configurações |
| `BEARM_STATE_HOME` | Caminho absoluto para o diretório de estado e diário |
| `BEARM_DATA_HOME`  | Caminho absoluto para o diretório de dados |
| `BEARM_TRASH`      | Caminho absoluto para raiz personalizada de lixeira |

Consulte o [Guia de Configuração (em inglês)](docs/configuration.md) para todos os detalhes técnicos.

---

## 📚 Documentação Detalhada

| Documento | Descrição |
| :--- | :--- |
| [Guia de Instalação](docs/installation.md) | Instruções aprofundadas sobre PATH, binários e shells |
| [Referência da CLI](docs/cli.md) | Lista completa de comandos, flags, opções e códigos de saída |
| [Guia de Configuração](docs/configuration.md) | Especificação do TOML, escopos de segurança e variáveis |
| [Recuperação e Diário](docs/recovery.md) | Ciclo de vida dos itens, formato JSONL e procedimentos de desastre |
| [Compatibilidade GNU](docs/compatibility/gnu.md) | Análise de paridade com o `rm` do GNU coreutils |
| [Compatibilidade BSD](docs/compatibility/bsd.md) | Análise de paridade com o `rm` do BSD/macOS |
| [Visão Geral da Arquitetura](docs/architecture/overview.md) | Estrutura de pacotes, fluxo de dados e caminho rápido O(1) |
| [Modelo de Segurança](docs/security.md) | Limites de confiança, invariantes de segurança e mitigação de riscos |

---

## 🤝 Contribuindo

Contribuições são muito bem-vindas! Leia o arquivo [CONTRIBUTING.md](CONTRIBUTING.md) para entender os padrões de código, testes de compatibilidade diferencial e diretrizes para pull requests.

---

## 📄 Licença

O Bearm é distribuído sob os termos da licença [MIT](LICENSE).
