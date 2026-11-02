# Correções da revisão de Bearm

Status: proposta para revisão; implementação não iniciada.

## Objetivo e escopo
Resolver os seis achados da revisão em entregas separadas: integridade do journal, cancelamento, limites do purge, metadados, documentação e versões de ferramentas.

## Restrições globais
- Go 1.24.0 como versão mínima.
- Linux e macOS continuam suportados.
- Nenhuma dependência nova de runtime.
- Preservar journal schema_version=1 e registros existentes.
- Não sobrescrever arquivos na restauração nem apagar conteúdo fora da lixeira validada.
- Não publicar releases, alterar credenciais ou executar purge em dados reais durante a validação.

## Abordagem proposta
Correções locais nos serviços existentes, testes de regressão e documentação atualizada. Alternativa: refazer o journal como transações persistidas antes de cada efeito; oferece recuperação mais forte, mas amplia muito o escopo. Não será implementada aqui. Para a cauda truncada, preferir rejeitar append com erro claro antes de modificar o arquivo; recuperação automática que descarta bytes exige uma política separada.

## Critérios por ponto
1. Append em journal truncado falha sem alterar bytes; entradas completas continuam legíveis. Documentar recuperação manual por cópia de segurança, sem descarte silencioso.
2. Cancelamento interrompe novos itens; efeitos concluídos ainda são registrados. Erros reais de persistência continuam visíveis. SIGKILL e falta de energia permanecem fora da garantia desta correção.
3. Purge só remove descendentes estritos de raízes verificadas para o backend, inclusive por montagem; lista vazia recusa tudo. Não derivar confiança apenas do caminho registrado. Validar caminhos absolutos e componentes intermediários sem symlinks; o item final pode ser symlink e deve ser removido sem seguir o alvo. Não prometer proteção contra troca concorrente de diretórios por processos com acesso à conta.
4. Após restore/purge bem-sucedido, limpar somente o sidecar correspondente ao backend. Metadado ausente é sucesso; erro de limpeza não desfaz efeito concluído nem impede registrar lifecycle. Não adicionar campos obrigatórios aos registros antigos.
5. READMEs em inglês e português distinguem dependências de compilação e runtime; somente fail e rename são anunciados como políticas utilizáveis, overwrite fica reservado.
6. Fixar versões iguais por ferramenta em CI, security, readiness e release. Manter matriz Go mínima/stable; não confundir essa matriz com versões fixas das ferramentas. Usar versões já comprovadas no repositório, sem atualizar dependências de aplicação.

## Ordem e limites
Ordem recomendada: 1, 2, 3, 4, 5, 6. Plano 4 deve preservar os limites estabelecidos no 3. Cada plano pode ser revisado isoladamente; não executar mudanças concorrentes nos arquivos compartilhados. Sem refatoração de arquitetura, nova CLI, recuperação automática ou promessa de durabilidade contra queda de energia.
