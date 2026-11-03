# Compatibilidade com `rm` BSD/macOS

O perfil BSD é selecionado automaticamente no macOS ou explicitamente com
`BEARM_COMPAT=bsd`.

## Opções suportadas

| Opção | Comportamento |
| --- | --- |
| `-f`, `--force` | Ignora operandos ausentes e desativa a interação anterior |
| `-i`, `--interactive=always` | Pergunta para cada alvo lógico |
| `-I`, `--interactive=once` | Pergunta uma vez para solicitações recursivas ou grandes |
| `-r`, `-R`, `--recursive` | Permite remover diretórios |
| `-d`, `--dir`, `--directory` | Permite remover diretórios vazios |
| `-v`, `--verbose` | Exibe remoções lógicas bem-sucedidas |
| `--` | Encerra o parsing de opções |

Opções curtas combinadas são aceitas.

## Ordem das opções

No perfil BSD e POSIX, o parsing para no primeiro operando. Argumentos
posteriores iniciados por `-` são operandos, salvo se `--` tiver sido usado
antes do primeiro operando. Antes dele, opções são aplicadas da esquerda para
a direita: opções posteriores de força e interação prevalecem.

```bash
bearm rm -f arquivo -v       # -v é um operando
bearm rm -i -f ausente       # force prevalece
bearm rm -- -arquivo         # terminador explícito de opções
```

## Diagnósticos e código de saída

O perfil BSD usa diagnósticos de caminhos e de uso no estilo BSD. Erros de uso
retornam `64` (`EX_USAGE`) quando o perfil selecionado usa esse código; falhas
operacionais retornam `1` e o sucesso retorna `0`.

Operações bem-sucedidas não escrevem nada, exceto com `-v`. Prompts e
diagnósticos são escritos no erro padrão. A seleção de idioma verifica
`LC_ALL`, depois `LC_MESSAGES`, depois `LANG`, e usa inglês como alternativa.

## Comportamentos não suportados

O Bearm rejeita `-P` (sobrescrita segura) e `-W` (recuperação de exclusão) do
BSD, pois são incompatíveis com mover itens para a lixeira. Opções específicas
de perfil que não são suportadas geram diagnóstico de uso.

## Proteções mais rígidas do Bearm

O Bearm sempre recusa `/`, `.`, `..`, equivalentes lexicais, raízes ativas da
lixeira e suas próprias raízes de configuração, estado e dados. Um operando
que seja link simbólico é movido como link; seu destino nunca é removido.

Os metadados **Colocar de Volta** do Finder não são preservados. O diário e
`bearm restore` são a referência oficial para recuperação.

## Host dos testes diferenciais

A CI do macOS executa a suíte em `macos-latest` hospedado pelo GitHub, com Go
1.27.3 e Go estável. Cada execução resolve o `/bin/rm` instalado e seleciona o
perfil BSD; os testes comparam código de saída, fluxos estáveis e resultados
normalizados na árvore de origem. A imagem do runner fornece a versão exata do
`rm` BSD, portanto a compatibilidade depende da versão registrada naquela
execução da CI.
