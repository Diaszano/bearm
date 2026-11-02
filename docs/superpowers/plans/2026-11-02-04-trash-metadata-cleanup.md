# Limpar metadados após restore e purge — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Limpar metadados após restore e purge.

**Architecture:** Alterar os pontos existentes e preservar contratos públicos. Adicionar trash.RemoveRecordMetadata(record domain.TrashRecord) error reutilizado nos dois serviços: linux-freedesktop usa <root>/info/<base>.trashinfo a partir de <root>/files/<base>; custom-trash usa info/<base>.json; darwin-system-trash usa <root>/.bearm-info/<base>.json. Validar layout conhecido e recusar intermediários symlink. Backend desconhecido retorna erro sem apagar caminho inferido. Limpar somente após efeito físico bem-sucedido; mesmo se limpeza falhar, persistir evento e acrescentar ItemFailed separado ao resultado. ENOENT é sucesso. Sem campo novo no journal.

**Tech Stack:** Go 1.24.0, testing, filesystem local; Markdown/YAML nas tarefas documentais e CI.

**Spec:** [Design da revisão](../specs/2026-11-02-audit-remediation-design.md), critério 4.

## Global Constraints

- Go 1.24.0 como versão mínima.
- Linux e macOS continuam suportados.
- Nenhuma dependência nova de runtime.
- Preservar journal schema_version=1 e registros existentes.
- Não sobrescrever arquivos na restauração nem apagar conteúdo fora da lixeira validada.
- Não publicar releases, alterar credenciais ou executar purge em dados reais durante a validação.

## Review Focus

- TestRemoveRecordMetadataLayouts: tabela dos três backends apaga sidecar certo e preserva vizinho.
- TestRemoveRecordMetadataMissingAndUnknown: ausente retorna nil; backend desconhecido e layout inválido retornam erro sem apagar.
- TestRestoreAndPurgeCleanMetadataOnSuccess: sidecar desaparece e evento correto é gravado.
- TestFailedOperationPreservesMetadata: colisão/falha física mantém o sidecar.
- TestCleanupFailureStillRecordsLifecycle: diretório não vazio no lugar do sidecar produz erro visível e evento; ancestral symlink não permite apagar metadado externo.

## Task 1: Limpar metadados após restore e purge

**Files:**
- `internal/trash/metadata.go`
- `internal/trash/metadata_test.go`
- `internal/restore/service.go`
- `internal/restore/service_test.go`
- `internal/restore/purge.go`
- `internal/restore/purge_test.go`

**Interfaces:**
- Consumes: contratos atuais e critérios 4 do design; ler arquivos listados e todos os callers antes de editar.
- Produces: comportamento descrito em Architecture; sem novas flags ou dependências. Plano documental/CI não altera interfaces Go.

- [ ] **Step 1: Criar regressões ou checklist verificável.**
  - TestRemoveRecordMetadataLayouts: tabela dos três backends apaga sidecar certo e preserva vizinho.
  - TestRemoveRecordMetadataMissingAndUnknown: ausente retorna nil; backend desconhecido e layout inválido retornam erro sem apagar.
  - TestRestoreAndPurgeCleanMetadataOnSuccess: sidecar desaparece e evento correto é gravado.
  - TestFailedOperationPreservesMetadata: colisão/falha física mantém o sidecar.
  - TestCleanupFailureStillRecordsLifecycle: diretório não vazio no lugar do sidecar produz erro visível e evento; ancestral symlink não permite apagar metadado externo.
- [ ] **Step 2: Confirmar estado anterior.** Executar `go test ./internal/trash ./internal/restore -count=1`. Para regressões Go, os novos testes devem falhar pelo comportamento descrito; testes existentes devem passar. Para documentação/CI, registrar as ocorrências que violam o checklist, sem criar testes artificiais.
- [ ] **Step 3: Aplicar a mudança mínima descrita em Architecture.** Manter error handling e fixtures; atualizar callers afetados, inclusive testes que passavam nil no purge se o plano 3 já estiver aplicado. Não corrigir outros achados neste commit.
- [ ] **Step 4: Verificar.** Executar `go test ./internal/trash ./internal/restore -count=1` e `go test -race ./internal/trash/... ./internal/restore`. Testes devem passar, diff não deve ter whitespace inválido; no plano 5, revisar manualmente as ocorrências do rg. No plano 6, validar workflows no CI sem acionar release; se ferramentas não estiverem instaladas, registrar validação pendente, nunca declarar aprovação.
- [ ] **Step 5: Revisar e criar commit isolado.** `git diff --check`; revisar somente os arquivos da tarefa; adicionar os arquivos alterados nominalmente e executar `git commit -m "fix: remove trash metadata after completed lifecycle actions"`. Não incluir mudanças preexistentes do usuário.

## Aceitação e limites

Todos os cinco itens de Review Focus devem estar cobertos pela tarefa. Executar `go test ./...` após mudanças Go e `go test -race ./...` ao integrar os planos 1–4. Validar macOS em runner macOS; cross-compilation não substitui execução. Nenhuma implementação foi realizada ao escrever este plano.
