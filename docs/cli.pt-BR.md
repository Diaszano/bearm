# Referência de comandos e opções da CLI

O Bearm tem dois modos de execução:

1. **Modo de compatibilidade**: ativado por `bearm rm ...` ou por uma chamada
   através de um executável/alias chamado `rm`. Imita opções e comportamentos
   do `rm(1)` Unix enquanto move os itens para a lixeira.
2. **Modo nativo**: ativado por subcomandos como `bearm list`, `restore`,
   `purge`, `doctor` e `config`.

## Modo de compatibilidade (`rm`)

```bash
bearm rm [OPÇÕES]... [ARQUIVO]...
```

### Opções suportadas por perfil

| Opção | Opção longa | Perfis | Descrição |
| :--- | :--- | :---: | :--- |
| `-f` | `--force` | GNU, BSD, POSIX | Ignora arquivos inexistentes e nunca pede confirmação. |
| `-i` | `--interactive=always` | GNU, BSD, POSIX | Pergunta antes de remover cada alvo. |
| `-I` | `--interactive=once` | GNU, BSD | Pergunta uma vez para mais de três arquivos ou remoção recursiva. |
| `-r`, `-R` | `--recursive` | GNU, BSD, POSIX | Permite remover diretórios e seu conteúdo. |
| `-d` | `--dir`, `--directory` | GNU, BSD | Permite remover diretórios vazios. |
| `-v` | `--verbose` | GNU, BSD | Informa o que está sendo movido para a lixeira. |
|  | `--one-file-system` | GNU | Não percorre diretórios em outro sistema de arquivos. |
|  | `--preserve-root` | GNU | Protege `/` (padrão e proteção incondicional do Bearm). |
|  | `--no-preserve-root` | GNU | Desativa a proteção GNU; a proteção rígida do Bearm continua ativa. |
|  | `--help` | GNU | Exibe a ajuda de compatibilidade. |
|  | `--version` | GNU | Exibe a versão e metadados de build do Bearm. |
| `--` |  | GNU, BSD, POSIX | Trata os argumentos seguintes como caminhos. |

### Diferenças entre perfis e ordem das opções

- **GNU** (padrão no Linux ou `BEARM_COMPAT=gnu`): aceita opções até `--`;
  aplica-as da esquerda para a direita, portanto `-f` posterior substitui
  `-i` anterior.
- **BSD** (padrão no macOS ou `BEARM_COMPAT=bsd`): para de interpretar opções
  no primeiro operando não opcional. Argumentos posteriores iniciados por `-`
  são nomes de arquivo, salvo se vierem após `--`.
- **POSIX** (`BEARM_COMPAT=posix`): segue as convenções estritas do POSIX.1.

### Garantias de segurança durante `rm`

- **Alvos protegidos**: rejeita `/`, `.`, `..`, equivalentes lexicais, pastas
  de lixeira ativas e arquivos de configuração/estado do Bearm.
- **Links simbólicos**: move apenas o link; nunca segue nem remove o destino.
- **Movimentação atômica de diretórios**: com remoção recursiva sem prompts, o
  diretório inteiro pode ser movido por um único `rename` atômico O(1).

## Subcomandos nativos

### `bearm list`

Lista itens ativos na lixeira removidos pelo Bearm.

```bash
bearm list [--limit <N>] [--operation <ID_OPERAÇÃO>] [--json]
```

- `--limit <N>`: máximo de registros (padrão: `50`).
- `--operation <ID_OPERAÇÃO>`: filtra itens de uma operação em lote.
- `--json`: imprime um array JSON legível por máquina.

### `bearm restore`

Restaura arquivos ou diretórios aos caminhos originais.

```bash
bearm restore ID_ITEM [ID_ITEM...]
bearm restore --operation ID_OPERAÇÃO
bearm restore --last
```

Se o destino original já existir, a política padrão `fail` preserva os dois
itens e interrompe. Com `restore.collision_policy = "rename"`, restaura com um
sufixo `.restored.N`. A política de sobrescrita é recusada pelo comando padrão,
a menos que seja confirmada explicitamente.

### `bearm purge`

Exclui permanentemente itens da lixeira. É o único comando Bearm que desvincula
arquivos de forma permanente.

```bash
bearm purge ID_ITEM [ID_ITEM...] [--yes]
bearm purge --operation ID_OPERAÇÃO [--yes]
bearm purge --last [--yes]
```

`--yes` dispensa a confirmação interativa. A purga é irreversível.

### `bearm doctor`

Audita a consistência do diário e dos backends da lixeira:

```bash
bearm doctor [--json]
```

Detecta itens órfãos na lixeira, registros que apontam para arquivos ausentes
e linhas finais incompletas no diário.

### `bearm config`

```bash
bearm config path   # mostra o caminho da configuração
bearm config check  # valida sintaxe e limites de segurança
```

### `bearm version`

Imprime versão, hash do commit, horário de build, runtime Go e arquitetura.

### Códigos de saída

| Código | Perfil/contexto | Significado |
| :---: | :--- | :--- |
| `0` | Todos | Sucesso; inclui arquivos ignorados com `-f`. |
| `1` | Todos | Falha operacional: item ausente, permissão, colisão ou erro de disco. |
| `2` | Nativo/App | Erro de parsing ou subcomando desconhecido. |
| `3` | App/Config | Configuração, diretório ou padrão protegido inválido. |
| `64` | BSD | Erro de uso (`EX_USAGE`) para opções inválidas ou não suportadas. |

### Variáveis de ambiente

| Variável | Padrão | Descrição |
| --- | --- | --- |
| `BEARM_COMPAT` | `auto` | Perfil: `auto`, `gnu`, `bsd` ou `posix`. |
| `BEARM_LANG` | `pt-BR` | Idioma das mensagens nativas: `pt-BR` ou `en`. |
| `BEARM_CONFIG_HOME` | Linux: `~/.config/bearm`; macOS: `~/Library/Application Support/Bearm` | Diretório da configuração; caminho absoluto. |
| `BEARM_STATE_HOME` | Linux: `~/.local/state/bearm`; macOS: `.../Bearm/state` | Diretório do diário/estado; caminho absoluto. |
| `BEARM_DATA_HOME` | Linux: `~/.local/share/bearm`; macOS: `.../Bearm/data` | Diretório de dados; caminho absoluto. |
| `BEARM_TRASH` | Padrão da plataforma | Caminho absoluto para a lixeira personalizada. |
| `BEARM_TRASH_PER_MOUNT` | `true` | No Linux, tenta usar lixeira no mesmo sistema de arquivos (`.Trash/<uid>` / `.Trash-<uid>`). |
