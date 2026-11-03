# Configuração

O Bearm funciona sem arquivo de configuração. Carrega os padrões, lê o TOML
estrito e aplica substituições do ambiente. Uma configuração inválida
interrompe o processo antes de qualquer alteração no sistema de arquivos.

## Caminho da configuração

Veja o caminho ativo:

```bash
bearm config path
```

Locais padrão:

- Linux: `$XDG_CONFIG_HOME/bearm/config.toml`, ou
  `$HOME/.config/bearm/config.toml` como alternativa.
- macOS: `$HOME/Library/Application Support/Bearm/config.toml`.

`BEARM_CONFIG_HOME` substitui a pasta de configuração da plataforma e deve
ser um caminho absoluto.

## Esquema versão 1

```toml
language = "pt-BR"
compatibility_profile = "auto"

[trash]
per_mount = true
custom_path = ""

[safety]
preserve_root = true
allowed_roots = []
protected_patterns = []
inspect_descendants = false

[restore]
collision_policy = "fail"

[logging]
level = "error"
```

Campos desconhecidos são rejeitados.

### Valores de nível superior

- `language`: `pt-BR` ou `en` para comandos nativos do Bearm.
- `compatibility_profile`: `auto`, `gnu`, `bsd` ou `posix`.

### Lixeira

- `per_mount`: no Linux, tenta usar uma lixeira no mesmo sistema de arquivos.
- `custom_path`: raiz absoluta personalizada. Vazio seleciona o backend da
  plataforma. Movimentações entre sistemas de arquivos falham sem copiar ou
  excluir a origem.

No macOS, a lixeira do sistema no volume do usuário ou no volume de destino é
sempre selecionada; `per_mount` se aplica ao backend Linux.

### Segurança

- `preserve_root`: ativa a política de raiz do perfil de compatibilidade. A
  proteção rígida de `/` continua ativa independentemente disso.
- `allowed_roots`: raízes absolutas que limitam remoções no modo de
  compatibilidade. Lista vazia permite qualquer caminho não rejeitado por
  outra regra de segurança.
- `protected_patterns`: padrões de caminho absolutos compatíveis com
  gitignore.
- `inspect_descendants`: percorre alvos recursivos para preservar descendentes
  protegidos. O padrão `false` mantém o caminho rápido e verifica apenas os
  operandos explícitos.

Exemplo:

```toml
[safety]
allowed_roots = ["/home/alice/work"]
protected_patterns = [
  "/home/alice/work/important/**",
  "/home/alice/work/**/.git",
]
inspect_descendants = true
```

Padrões protegidos negados também devem ser absolutos, por exemplo:
`!/home/alice/work/important/cache/**`.

### Restauração

`collision_policy` aceita:

- `fail`: recusa um caminho original já existente;
- `rename`: restaura para um item irmão, como `file.restored.1`;
- `overwrite`: reservado para fluxos de sobrescrita explicitamente
  confirmados; o comando padrão atual o recusa.

### Registro

`level` aceita `error` ou `debug`. O Bearm escreve diagnósticos localmente no
erro padrão e nunca envia telemetria.

## Variáveis de ambiente

| Variável | Efeito |
| --- | --- |
| `BEARM_LANG` | Idioma nativo: `pt-BR` ou `en` |
| `BEARM_COMPAT` | Perfil de compatibilidade |
| `BEARM_TRASH` | Raiz absoluta da lixeira personalizada |
| `BEARM_TRASH_PER_MOUNT` | Roteamento Linux por montagem (booleano) |
| `BEARM_LOG` | `error` ou `debug` |
| `BEARM_CONFIG_HOME` | Raiz absoluta da configuração |
| `BEARM_STATE_HOME` | Raiz absoluta do diário/estado |
| `BEARM_DATA_HOME` | Raiz absoluta dos dados do Bearm |
| `XDG_CONFIG_HOME` | Base de configuração no Linux |
| `XDG_STATE_HOME` | Base de estado no Linux |
| `XDG_DATA_HOME` | Base de dados e lixeira do usuário no Linux |
| `LC_ALL`, `LC_MESSAGES`, `LANG` | Seleção de localidade para compatibilidade |

Substituições de diretórios do Bearm com caminhos relativos são rejeitadas.
Valores XDG relativos são ignorados em favor dos caminhos alternativos da
plataforma.

## Verificações de segurança

No Linux e macOS, o arquivo de configuração deve:

- ser um arquivo regular, não um link simbólico;
- pertencer ao usuário atual;
- não permitir escrita por grupo ou outros usuários.

Criação recomendada:

```bash
install -d -m 0700 "$(dirname "$(bearm config path)")"
install -m 0600 /dev/null "$(bearm config path)"
bearm config check
```

`bearm config check` retorna `0` para configuração válida e `3` para
configuração inválida.
