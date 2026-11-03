# Registrar efeitos concluídos após cancelamento — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Registrar efeitos concluídos após cancelamento.

**Architecture:** Alterar os pontos existentes e preservar contratos públicos. Preservar Execute, Restore e Purge e suas assinaturas. Usar context.WithoutCancel(ctx) somente na persistência dos resultados já concluídos; manter checks ctx.Err antes de novos efeitos. Não criar goroutines, retries ou abstração de contexto. Usar fake que cancela após a primeira mutação e journal que rejeita contexto cancelado.

**Tech Stack:** Go 1.24.0, testing, filesystem local; Markdown/YAML nas tarefas documentais e CI.

**Spec:** [Design da revisão](../specs/2026-11-02-audit-remediation-design.md), critério 2.

## Global Constraints

- Go 1.24.0 como versão mínima.
- Linux e macOS continuam suportados.
- Nenhuma dependência nova de runtime.
- Preservar journal schema_version=1 e registros existentes.
- Não sobrescrever arquivos na restauração nem apagar conteúdo fora da lixeira validada.
- Não publicar releases, alterar credenciais ou executar purge em dados reais durante a validação.

## Review Focus

- TestExecutePersistsCompletedItemsAfterCancellation: primeiro Move cancela contexto; um evento é persistido e segundo item fica intacto.
- TestRestorePersistsCompletedItemsAfterCancellation: cancelar após primeiro rename via journal/fake ou hook mínimo de teste; registro concluído é persistido.
- TestPurgePersistsCompletedItemsAfterCancellation: cancelar entre itens de modo determinístico; item apagado ganha evento e próximo fica intacto.
- TestAlreadyCancelledDoesNotMutate: contexto inicialmente cancelado não move nem remove.
- TestFinalizationJournalErrorIsReported: erro real de append produz falha sem ocultar sucesso físico do primeiro item.

## Task 1: Registrar efeitos concluídos após cancelamento

**Files:**
- `internal/removal/executor.go`
- `internal/removal/executor_test.go`
- `internal/restore/service.go`
- `internal/restore/service_test.go`
- `internal/restore/purge.go`
- `internal/restore/purge_test.go`
- `docs/recovery.md`

**Interfaces:**
- Consumes: contratos atuais e critérios 2 do design; ler arquivos listados e todos os callers antes de editar.
- Produces: comportamento descrito em Architecture; sem novas flags ou dependências. Plano documental/CI não altera interfaces Go.

- [ ] **Step 1: Criar regressões ou checklist verificável.**
  - TestExecutePersistsCompletedItemsAfterCancellation: primeiro Move cancela contexto; um evento é persistido e segundo item fica intacto.
  - TestRestorePersistsCompletedItemsAfterCancellation: cancelar após primeiro rename via journal/fake ou hook mínimo de teste; registro concluído é persistido.
  - TestPurgePersistsCompletedItemsAfterCancellation: cancelar entre itens de modo determinístico; item apagado ganha evento e próximo fica intacto.
  - TestAlreadyCancelledDoesNotMutate: contexto inicialmente cancelado não move nem remove.
  - TestFinalizationJournalErrorIsReported: erro real de append produz falha sem ocultar sucesso físico do primeiro item.
- [ ] **Step 2: Confirmar estado anterior.** Executar `go test ./internal/removal ./internal/restore -count=1`. Para regressões Go, os novos testes devem falhar pelo comportamento descrito; testes existentes devem passar. Para documentação/CI, registrar as ocorrências que violam o checklist, sem criar testes artificiais.
- [ ] **Step 3: Aplicar a mudança mínima descrita em Architecture.** Manter error handling e fixtures; atualizar callers afetados, inclusive testes que passavam nil no purge se o plano 3 já estiver aplicado. Não corrigir outros achados neste commit.
- [ ] **Step 4: Verificar.** Executar `go test ./internal/removal ./internal/restore -count=1` e `go test -race ./internal/removal ./internal/restore ./test/integration`. Testes devem passar, diff não deve ter whitespace inválido; no plano 5, revisar manualmente as ocorrências do rg. No plano 6, validar workflows no CI sem acionar release; se ferramentas não estiverem instaladas, registrar validação pendente, nunca declarar aprovação.
- [ ] **Step 5: Revisar e criar commit isolado.** `git diff --check`; revisar somente os arquivos da tarefa; adicionar os arquivos alterados nominalmente e executar `git commit -m "fix: persist completed lifecycle events after cancellation"`. Não incluir mudanças preexistentes do usuário.

## Aceitação e limites

Todos os cinco itens de Review Focus devem estar cobertos pela tarefa. Executar `go test ./...` após mudanças Go e `go test -race ./...` ao integrar os planos 1–4. Validar macOS em runner macOS; cross-compilation não substitui execução. Nenhuma implementação foi realizada ao escrever este plano.
