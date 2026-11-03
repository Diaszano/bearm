# Como contribuir com o Bearm

## Requisitos

- Go 1.27.3 ou superior;
- Git;
- `make`;
- ambiente de teste Linux ou macOS.

As ferramentas Go são fixadas em `tools/go.mod` e executadas pelo Makefile e
pelo script de aceitação. Não é necessário instalá-las globalmente. O comando
`go` seleciona automaticamente a toolchain necessária.

## Fluxo de trabalho

1. Use nomes de branch convencionais em inglês, como
   `feat/linux-trash-backend`, `fix/restore-collision` ou `docs/configuration`.
2. Para cada mudança de comportamento, escreva um teste que falhe e confirme
   que a falha decorre do comportamento ausente.
3. Implemente a menor mudança que faça o teste passar.
4. Execute:

   ```bash
   make verify
   ```

5. Mantenha cada tarefa em um commit Conventional Commit focado.

### Especificação Conventional Commits

Mensagens de commit e títulos de Pull Request devem seguir a especificação
[Conventional Commits v1.0.0](https://www.conventionalcommits.org/pt-br/):

```text
<tipo>(<escopo opcional>): <descrição>
```

Tipos aceitos:

- `feat`: funcionalidade;
- `fix`: correção de defeito;
- `docs`: apenas documentação;
- `style`: formatação sem mudança de lógica;
- `refactor`: reorganização sem corrigir defeito ou adicionar funcionalidade;
- `perf`: melhoria de desempenho;
- `test`: inclusão ou correção de testes;
- `build`: sistema de build ou dependências externas;
- `ci`: configuração de CI e scripts;
- `chore`: outras mudanças que não alteram código-fonte ou testes;
- `revert`: reversão de commit anterior.

Para mudanças incompatíveis, acrescente `!` antes dos dois-pontos ou inclua
`BREAKING CHANGE:` no rodapé.

Não inclua trailers de atribuição a IA. Código-fonte, identificadores e
comentários são escritos em inglês. O inglês é a fonte da documentação
pública; mantenha uma versão correspondente em português brasileiro
(`.pt-BR.md`) ao atualizar um documento.

## Mudanças de compatibilidade

Uma mudança de comportamento compatível deve incluir um fixture que execute o
`rm` do sistema e o Bearm de teste em diretório temporário isolado. Compare os
aspectos observáveis estáveis para o caso:

- código de saída;
- saída padrão;
- erro padrão ou prompt;
- resultado normalizado na árvore de origem;
- ordem de travessia, quando relevante.

Não enfraqueça as proteções rígidas do Bearm para imitar o `rm` do sistema.

## Testes

Antes de enviar uma mudança:

```bash
test -z "$(gofmt -l .)"
go mod verify
go vet ./...
go test ./...
go test -race ./...
```

Mudanças em backends devem ser executadas no sistema operacional de destino.
Os testes só podem operar em diretórios temporários isolados, nunca em dados
do usuário.

## APIs Go públicas

Documente cada declaração Go exportada com um comentário Go padrão. Prefira
arquivos e interfaces focados, mantendo parsing, políticas, mutações do
sistema de arquivos, diário e renderização separados.

## Segurança

Não abra uma issue pública para uma vulnerabilidade possivelmente ainda sem
correção. Siga a [política de segurança](SECURITY.pt-BR.md).
