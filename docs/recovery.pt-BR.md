# Recuperação

O Bearm registra cada movimentação concluída para a lixeira como um evento
somente de acréscimo. Use os comandos nativos para inspecionar e recuperar os
itens.

## Listar itens ativos

```bash
bearm list
bearm list --limit 20
bearm list --json
```

A saída de texto mostra o ID do item, horário de remoção e caminho original. A
saída JSON contém o `TrashRecord` ativo completo, incluindo ID da operação e
backend.

## Restaurar

```bash
bearm restore ID_ITEM
bearm restore ID_ITEM_1 ID_ITEM_2
bearm restore --operation ID_OPERAÇÃO
bearm restore --last
```

A política de colisão padrão `fail` mantém intactos tanto o caminho original
existente quanto o item na lixeira. Defina
`restore.collision_policy = "rename"` para restaurar com um sufixo
`.restored.N`. O comando nativo atual recusa a política de sobrescrita sem
confirmação.

A restauração acrescenta um evento `restored`; os eventos `trashed` históricos
não são reescritos.

## Purgar

Remove permanentemente os itens ativos selecionados da lixeira:

```bash
bearm purge ID_ITEM
bearm purge ID_ITEM --yes
bearm purge --operation ID_OPERAÇÃO
bearm purge --last
```

Sem `--yes`, o Bearm pergunta pelo erro padrão. Recusar mantém todos os itens
inalterados. A purga só existe no modo nativo; a remoção em modo de
compatibilidade não pode apagar permanentemente um item gerenciado pelo Bearm.

A purga acrescenta um evento `purged` após a exclusão permanente. Trate
`--yes` como uma operação irreversível.

## Diagnosticar o histórico

```bash
bearm doctor
bearm doctor --json
```

O comando não altera dados. Ele reporta uma linha final incompleta no diário
ou um registro ativo cujo item não está na lixeira.

Uma linha JSONL final incompleta é ignorada ao reconstruir o estado ativo, mas
permanece disponível para diagnóstico. Novos eventos são recusados até que o
arquivo seja restaurado de um backup ou reparado manualmente; o Bearm nunca
descarta os bytes danificados automaticamente.

## Localização do diário

- Linux: `$XDG_STATE_HOME/bearm/journal.jsonl`, ou
  `$HOME/.local/state/bearm/journal.jsonl` como alternativa.
- macOS: `$HOME/Library/Application Support/Bearm/state/journal.jsonl`.

`BEARM_STATE_HOME` substitui a raiz de estado e deve ser absoluta.

As gravações usam um lock exclusivo, um objeto JSON por linha e `fsync`. Os
registros contêm caminhos e metadados do ciclo de vida, nunca o conteúdo dos
arquivos.

## Finder no macOS

O Bearm usa as pastas de Lixeira visíveis do sistema, mas não garante os
metadados **Colocar de Volta** do Finder. O Finder pode exibir o item sem
conhecer o caminho de origem. `bearm restore` e o diário são as referências
oficiais para recuperação.

## Procedimento manual para incidentes

Se `bearm doctor` indicar item ausente ou diário incompleto:

1. Pare de executar restauração ou purga no item afetado.
2. Copie o diário e os metadados relevantes da lixeira para uma pasta de
   diagnóstico segura.
3. Registre `bearm version`, o sistema operacional e a saída do diagnóstico.
4. Inspecione os caminhos sem editar ou truncar o diário original.
5. Use o processo privado de segurança se houver exposição de dados ou
   possibilidade de exclusão insegura.
