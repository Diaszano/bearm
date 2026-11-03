# Compatibilidade com `rm` GNU

O perfil GNU é selecionado automaticamente no Linux ou explicitamente com
`BEARM_COMPAT=gnu`.

## Opções suportadas

| Opção | Comportamento |
| --- | --- |
| `-f`, `--force` | Ignora operandos ausentes e desativa interação anterior |
| `-i`, `--interactive=always` | Pergunta para cada alvo lógico |
| `-I`, `--interactive=once` | Pergunta uma vez para solicitações recursivas ou grandes |
| `--interactive=never` | Desativa perguntas |
| `-r`, `-R`, `--recursive` | Permite remover diretórios |
| `-d`, `--dir`, `--directory` | Permite remover diretórios vazios |
| `-v`, `--verbose` | Exibe remoções lógicas bem-sucedidas |
| `--one-file-system` | Mantém a travessia no sistema de arquivos do operando |
| `--preserve-root` | Ativa proteção de raiz compatível |
| `--preserve-root=all` | Protege operandos diretórios em outro dispositivo que não o do pai |
| `--no-preserve-root` | Desativa somente a proteção de raiz do GNU |
| `--help`, `--version` | Exibe ajuda ou informações de build do Bearm |
| `--` | Encerra o parsing de opções |

Opções curtas combinadas são aceitas. Opções GNU podem aparecer após
operandos até que `--` encerre o parsing.

## Ordem das opções

As opções são aplicadas da esquerda para a direita. Um `-f` ou
`--interactive=never` posterior desativa a interação anterior. Um `-i`, `-I`
ou `--interactive=...` posterior limpa `force` e define o novo modo de
interação.

```bash
bearm rm -i -f ausente       # force prevalece
bearm rm -f -i arquivo       # interactive always prevalece
bearm rm arquivo -v          # -v é interpretado como opção
```

## Diagnósticos e código de saída

O perfil GNU usa diagnósticos de uso e caminho no estilo GNU e retorna:

- `0` quando todas as operações têm sucesso ou são ignoradas com `force`;
- `1` para falhas de uso ou operacionais.

Operações bem-sucedidas não escrevem nada, exceto com `-v`. Prompts e
diagnósticos são escritos no erro padrão. A seleção de idioma verifica
`LC_ALL`, depois `LC_MESSAGES`, depois `LANG`, e usa inglês como alternativa.

## Comportamentos não suportados

Opções desconhecidas ou incompatíveis falham, em vez de serem ignoradas. O
Bearm não implementa `-P` e `-W` do BSD nem extensões destrutivas específicas
de implementações que não estejam documentadas no perfil.

## Proteções mais rígidas do Bearm

`--no-preserve-root` não desativa a proteção rígida de `/`. O Bearm também
recusa `.`, `..`, equivalentes lexicais, raízes ativas da lixeira e suas
próprias raízes de configuração, estado e dados. O modo de compatibilidade
não exclui permanentemente alvos dentro da lixeira gerenciada pelo Bearm.

## Host dos testes diferenciais

A CI Linux executa a suíte em `ubuntu-latest` hospedado pelo GitHub, com Go
1.27.3 e Go estável. Cada execução resolve o `rm` instalado e seleciona o
perfil GNU; os testes comparam códigos de saída, fluxos estáveis e resultados
normalizados na árvore de origem. A imagem do runner fornece a versão exata do
GNU coreutils, portanto a compatibilidade depende da versão registrada
naquela execução da CI, não de uma versão histórica fixa.
