# Recusar append em journal truncado — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Recusar append em journal truncado.

**Architecture:** Alterar os pontos existentes e preservar contratos públicos. Manter Repository.AppendEvents(ctx context.Context, events []domain.JournalEvent) error. Sob lock exclusivo, verificar se arquivo não vazio termina em newline antes de anexar; abrir com leitura além de append. Retornar erro descritivo sem modificar bytes se houver cauda incompleta; não truncar nem tentar corrigir JSON completo inválido.

**Tech Stack:** Go 1.24.0, testing, filesystem local; Markdown/YAML nas tarefas documentais e CI.

**Spec:** [Design da revisão](../specs/2026-11-02-audit-remediation-design.md), critério 1.

## Global Constraints

- Go 1.24.0 como versão mínima.
- Linux e macOS continuam suportados.
- Nenhuma dependência nova de runtime.
- Preservar journal schema_version=1 e registros existentes.
- Não sobrescrever arquivos na restauração nem apagar conteúdo fora da lixeira validada.
- Não publicar releases, alterar credenciais ou executar purge em dados reais durante a validação.

## Review Focus

- TestAppendRejectsIncompleteTail: escrever uma entrada válida e uma cauda truncada; AppendEvents retorna erro; bytes antes/depois são iguais; ReadAll preserva a entrada válida e IncompleteTrailingLine=true.
- TestAppendAfterCompleteTail: append em arquivo vazio e depois em arquivo com newline mantém dois eventos legíveis.
- TestAppendConcurrentCompleteEvents: escritores simultâneos preservam todas as entradas completas.
- TestAppendRejectsTailWithoutNewline: JSON válido sem newline também é recusado sem alteração.
- TestAppendInvalidEventDoesNotWrite: evento inválido não modifica o arquivo.

## Task 1: Recusar append em journal truncado

**Files:**
- `internal/journal/repository.go`
- `internal/journal/recovery_test.go`
- `docs/recovery.md`

**Interfaces:**
- Consumes: contratos atuais e critérios 1 do design; ler arquivos listados e todos os callers antes de editar.
- Produces: comportamento descrito em Architecture; sem novas flags ou dependências. Plano documental/CI não altera interfaces Go.

- [ ] **Step 1: Criar regressões ou checklist verificável.**
  - TestAppendRejectsIncompleteTail: escrever uma entrada válida e uma cauda truncada; AppendEvents retorna erro; bytes antes/depois são iguais; ReadAll preserva a entrada válida e IncompleteTrailingLine=true.
  - TestAppendAfterCompleteTail: append em arquivo vazio e depois em arquivo com newline mantém dois eventos legíveis.
  - TestAppendConcurrentCompleteEvents: escritores simultâneos preservam todas as entradas completas.
  - TestAppendRejectsTailWithoutNewline: JSON válido sem newline também é recusado sem alteração.
  - TestAppendInvalidEventDoesNotWrite: evento inválido não modifica o arquivo.
- [ ] **Step 2: Confirmar estado anterior.** Executar `go test ./internal/journal -count=1`. Para regressões Go, os novos testes devem falhar pelo comportamento descrito; testes existentes devem passar. Para documentação/CI, registrar as ocorrências que violam o checklist, sem criar testes artificiais.
- [ ] **Step 3: Aplicar a mudança mínima descrita em Architecture.** Manter error handling e fixtures; atualizar callers afetados, inclusive testes que passavam nil no purge se o plano 3 já estiver aplicado. Não corrigir outros achados neste commit.
- [ ] **Step 4: Verificar.** Executar `go test ./internal/journal -count=1` e `go test -race ./internal/journal`. Testes devem passar, diff não deve ter whitespace inválido; no plano 5, revisar manualmente as ocorrências do rg. No plano 6, validar workflows no CI sem acionar release; se ferramentas não estiverem instaladas, registrar validação pendente, nunca declarar aprovação.
- [ ] **Step 5: Revisar e criar commit isolado.** `git diff --check`; revisar somente os arquivos da tarefa; adicionar os arquivos alterados nominalmente e executar `git commit -m "fix: reject journal appends after incomplete trailing lines"`. Não incluir mudanças preexistentes do usuário.

## Aceitação e limites

Todos os cinco itens de Review Focus devem estar cobertos pela tarefa. Executar `go test ./...` após mudanças Go e `go test -race ./...` ao integrar os planos 1–4. Validar macOS em runner macOS; cross-compilation não substitui execução. Nenhuma implementação foi realizada ao escrever este plano.
