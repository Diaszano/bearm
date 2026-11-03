# Visão geral da arquitetura

O Bearm é um único binário Go, montado por uma camada fina em `cmd/bearm`.
Os pacotes em `internal` separam parsing, políticas, planejamento do sistema de
arquivos, operações da lixeira, histórico e recuperação.

## Direção das dependências

- `internal/domain` contém tipos de requisição, plano, resultado, lixeira e
  diário independentes da infraestrutura.
- `internal/cli` converte argumentos em requisições de domínio e renderiza
  resultados tipados e diagnósticos de compatibilidade.
- `internal/config`, `internal/i18n` e `internal/platform` fornecem
  configuração tipada, catálogos e caminhos do sistema operacional.
- `internal/safety` e `internal/planner` inspecionam requisições e o sistema de
  arquivos sem gravar na lixeira.
- `internal/removal` executa um plano concluído por interfaces de domínio.
- `internal/trash/linux`, `internal/trash/darwin` e `internal/trash/custom`
  implementam destinos por plataforma.
- `internal/journal` persiste eventos do ciclo de vida; não toma decisões de
  segurança.
- `internal/restore` implementa os casos de uso de listagem, recuperação,
  purga e diagnóstico sem interpretar a linha de comando.
- `internal/app` monta essas partes e coordena os casos de uso.

Os pacotes de domínio e política não importam backends de plataforma. Os
pacotes de lixeira não interpretam argumentos nem escrevem mensagens ao
usuário.

## Fluxo de remoção

```text
argv
  |
  v
resolver invocação -> interpretar -> validar -> planejar -> confirmar -> executar -> diário
                           |          |          |           |
                           |          |          |           +-- renderizar resultado
                           |          |          +-- usuário pode recusar
                           |          +-- segurança e inspeção do sistema de arquivos
                           +-- requisição GNU, BSD, POSIX ou nativa tipada
```

O parser mantém o perfil e a ordem das opções. O planejamento usa `Lstat`,
aplica políticas rígidas e configuradas, remove cobertura duplicada e cria
alvos imutáveis. O executor verifica novamente a política imediatamente antes
de cada movimentação.

## Caminho rápido sem travessia

Uma remoção recursiva comum pode mover um diretório com um único `rename` se
não houver confirmação por entrada, inspeção de descendentes protegidos ou
filtro de limites de montagem. O planejador registra o diretório como um alvo
e não visita os descendentes. Assim, o custo em relação à quantidade de itens
internos é praticamente constante para diretórios no mesmo sistema de
arquivos.

A travessia é explícita quando interação ou políticas exigem decisões por
entrada. As caminhadas são determinísticas, usam `Lstat` e não seguem links
simbólicos.

## Limites dos backends

No Linux, o Bearm usa a lixeira da conta conforme FreeDesktop ou uma lixeira
segura por montagem. Cria `files/` e `info/` e grava metadados `.trashinfo`.

No macOS, itens no volume do usuário vão para `$HOME/.Trash`; itens em outros
volumes vão para `<mount>/.Trashes/<uid>`. Os metadados do Bearm ficam em
`.bearm-info`, uma pasta privada dentro da lixeira.

A configuração de lixeira personalizada usa a reserva compartilhada, mas
aceita apenas `rename` atômico. Uma falha `EXDEV` não inicia cópia recursiva.

## Reserva atômica e limites de falha

O pacote compartilhado de lixeira reserva metadados com criação exclusiva,
tentando o nome base e depois sufixos numéricos apenas em caso de colisão. Se
a reserva falhar, a origem não muda. Se o `rename` falhar, a reserva é
desfeita. Se funcionar, a reserva é confirmada e gera um `TrashRecord`.

O diário serializa acréscimos com um lock consultivo e grava um evento JSON
completo por linha. Uma linha final incompleta é ignorada ao ler o estado
ativo e reportada por `bearm doctor`.

## Ciclo de recuperação e purga

```text
evento trashed -> item ativo -> evento restored
                              \-> evento purged
```

A restauração renomeia o item para o caminho original e acrescenta um evento
`restored`. A purga remove permanentemente apenas os itens ativos selecionados,
após confirmação, e acrescenta um evento `purged`. Linhas históricas nunca são
reescritas para representar uma mudança de estado.
