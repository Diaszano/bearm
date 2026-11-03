# Política de segurança

## Versões com suporte

Até o Bearm chegar à versão 1.0, as correções de segurança são feitas na
branch de desenvolvimento mais recente. Após o início das versões estáveis,
este documento listará explicitamente as linhas de versões com suporte.

## Relato privado

Relate vulnerabilidades pelo fluxo privado de relato do GitHub:

1. Abra a aba **Security** do repositório Bearm.
2. Selecione **Advisories**.
3. Selecione **Report a vulnerability**.

Se o relato privado não estiver disponível, entre em contato com a pessoa
responsável pelo repositório pelo canal privado exibido no perfil do GitHub.
Não abra uma issue pública sobre uma vulnerabilidade ainda sem correção.

Inclua:

- versão ou commit afetado do Bearm;
- sistema operacional e arquitetura;
- reprodução mínima usando caminhos descartáveis;
- comportamento de segurança esperado e observado;
- impacto e se há perda de dados ou limites de privilégio envolvidos.

Não envie credenciais, conteúdo de arquivos, caminhos pessoais ou arquivos
compactados com dados privados. A manutenção confirmará o recebimento,
investigará, coordenará uma correção e publicará um aviso quando apropriado.

## Escopo

Áreas sensíveis à segurança incluem:

- contorno de proteções rígidas ou configuradas de caminhos;
- seguir um destino de link simbólico;
- sobrescrever um item da lixeira ou seus metadados existentes;
- propriedade insegura ou tratamento incorreto de links na pasta da lixeira;
- injeção na configuração ou contorno das verificações de propriedade;
- corrupção do diário que cause restauração ou purga insegura;
- exclusão permanente sem confirmação explícita nativa.

Diferenças gerais de compatibilidade sem impacto de segurança devem ser
relatadas em uma issue normal, após consultar a documentação de compatibilidade
[GNU](docs/compatibility/gnu.pt-BR.md) e [BSD](docs/compatibility/bsd.pt-BR.md).
