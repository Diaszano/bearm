# Modelo de segurança

O Bearm reduz o risco de exclusões acidentais; não torna segura a manipulação
de um sistema de arquivos local não confiável. Use-o com os mesmos cuidados e
privilégios de conta que usaria com o `rm` do sistema.

## Ativos protegidos

- arquivos de origem e árvores de diretórios;
- identidade de links simbólicos e integridade de destinos;
- itens na lixeira e metadados contra colisão;
- propriedade e integridade da configuração;
- histórico de operações somente de acréscimo;
- intenção explícita do usuário para purga permanente.

## Ameaças e controles

### Injeção de shell e caminhos

O Bearm não executa shell e passa caminhos diretamente às APIs de sistema de
arquivos do Go. A configuração é TOML estrito, nunca executado como código.
Nomes de arquivo podem conter espaços, quebras de linha, metacaracteres,
hífens iniciais e texto não ASCII; use `--` antes de operandos ambíguos.

### Links simbólicos

O planejamento e as verificações de segurança usam `Lstat`. Um operando que
seja um link final é movido como link, sem seguir seu destino. As pastas da
lixeira devem ser diretórios reais, não links simbólicos.

### Proteções rígidas e configuradas

O modo de compatibilidade rejeita `/`, `.`, `..`, equivalentes lexicais,
lixeiras resolvidas do Bearm e as raízes de configuração, estado e dados do
Bearm. Raízes permitidas e padrões protegidos podem restringir ainda mais o
escopo. A segurança é verificada no planejamento e novamente imediatamente
antes da execução.

### Colisões e concorrência

A reserva de metadados usa criação exclusiva. O Bearm tenta o nome base e
depois sufixos numéricos somente em caso de colisão. Nunca sobrescreve item ou
reserva de metadados existente. Operações concorrentes serializam gravações do
diário por um lock consultivo.

### Confiança nas pastas da lixeira

No Linux, a seleção por montagem usa `<mount>/.Trash/<uid>` apenas quando
`.Trash` é um diretório real com sticky bit; caso contrário, usa
`<mount>/.Trash-<uid>`. Caminhos personalizados devem ser absolutos e suas
pastas são validadas.

### Confiança na configuração

No Linux e macOS, o arquivo de configuração deve ser regular, pertencer ao
usuário atual e não permitir escrita por grupo ou outros usuários. Campos
desconhecidos, raízes protegidas relativas, lixeira personalizada relativa e
valores de enumeração desconhecidos interrompem a inicialização antes de
alterar o sistema de arquivos.

### Exclusão permanente

Somente `bearm purge` no modo nativo exclui permanentemente itens gerenciados
na lixeira. Exige confirmação interativa ou `--yes`. O modo de compatibilidade
não transforma um alvo dentro da lixeira resolvida do Bearm em outra operação
de remoção.

### Privacidade

O Bearm não tem telemetria e não acessa a rede durante a execução. O diário
armazena caminhos original e da lixeira, IDs de operação e item, horários,
backend, ID do dispositivo e estado. Nunca armazena conteúdo dos arquivos.

## Limites em caso de falha

- Falha de reserva mantém a origem intacta.
- Falha de `rename` desfaz os metadados reservados.
- Movimentações para uma lixeira personalizada em outro sistema de arquivos
  falham sem cópia alternativa.
- Falhas em alvos independentes produzem resultado diferente de zero sem
  informar incorretamente que os alvos com falha foram movidos.
- Uma linha final incompleta do diário é reportada por `bearm doctor`.

## Limitações

O Bearm não protege contra uma conta ou processo capaz de alterar o binário,
suas dependências, diretório de estado ou pastas da lixeira por fora do Bearm.
Não apaga conteúdo de forma segura, não preserva os metadados **Colocar de
Volta** do Finder e não oferece garantia de backup. A lixeira facilita a
recuperação, mas não substitui um backup.

## Relato de vulnerabilidade

Siga o processo privado na [política de segurança](../SECURITY.pt-BR.md). Não
inclua em issues públicas caminhos reais sensíveis, conteúdo de arquivos,
credenciais ou detalhes de exploração.
