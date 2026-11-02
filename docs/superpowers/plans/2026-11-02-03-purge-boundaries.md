# Validar raízes antes de purge — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Validar raízes antes de purge.

**Architecture:** Alterar os pontos existentes e preservar contratos públicos. Preservar NewPurger(journal domain.EventAppender, clock func() time.Time, trashRoots []string) *Purger; vazio deve recusar tudo. No app, resolver destino pelo Backend.Resolve(ctx, domain.PlannedTarget) com AbsolutePath=record.OriginalPath e validar correspondência backend/layout sem mover conteúdo. Usar raízes de conteúdo (FilesDir), não info/root abrangente. Resolver pode preparar diretórios: documentar e testar ausência de remoção no erro. Reusar resolvers existentes para home/custom/per-mount; nunca autorizar uma raiz derivada só de TrashedPath. Recusar raiz igual, caminhos relativos e symlinks intermediários; permitir symlink final.

**Tech Stack:** Go 1.24.0, testing, filesystem local; Markdown/YAML nas tarefas documentais e CI.

**Spec:** [Design da revisão](../specs/2026-11-02-audit-remediation-design.md), critério 3.

## Global Constraints

- Go 1.24.0 como versão mínima.
- Linux e macOS continuam suportados.
- Nenhuma dependência nova de runtime.
- Preservar journal schema_version=1 e registros existentes.
- Não sobrescrever arquivos na restauração nem apagar conteúdo fora da lixeira validada.
- Não publicar releases, alterar credenciais ou executar purge em dados reais durante a validação.

## Review Focus

- TestPurgeRejectsEmptyRoots: arquivo externo permanece intacto quando raízes são nil.
- TestPurgeRejectsRootAndSiblingPrefix: raiz e trash-other são recusadas; filho legítimo é apagado.
- TestPurgeRejectsSymlinkParent: trash/link/item não apaga alvo externo; symlink final é removido sem seguir seu alvo.
- TestAppPurgeValidatesResolvedDestination: journal apontando fora do destino resolvido é recusado inclusive com --yes.
- TestAppPurgeSupportsBackendRoots: custom/home/per-mount legítimos funcionam em fixtures isoladas; backend desconhecido ou resolução com erro recusa sem apagar.

## Task 1: Validar raízes antes de purge

**Files:**
- `internal/restore/purge.go`
- `internal/restore/purge_internal_test.go`
- `internal/restore/purge_test.go`
- `internal/app/app.go`
- `internal/app/dependencies.go`
- `internal/app/backend_linux.go`
- `internal/app/backend_darwin.go`
- `internal/app/app_test.go`
- `docs/security.md`

**Interfaces:**
- Consumes: contratos atuais e critérios 3 do design; ler arquivos listados e todos os callers antes de editar.
- Produces: comportamento descrito em Architecture; sem novas flags ou dependências. Plano documental/CI não altera interfaces Go.

- [ ] **Step 1: Criar regressões ou checklist verificável.**
  - TestPurgeRejectsEmptyRoots: arquivo externo permanece intacto quando raízes são nil.
  - TestPurgeRejectsRootAndSiblingPrefix: raiz e trash-other são recusadas; filho legítimo é apagado.
  - TestPurgeRejectsSymlinkParent: trash/link/item não apaga alvo externo; symlink final é removido sem seguir seu alvo.
  - TestAppPurgeValidatesResolvedDestination: journal apontando fora do destino resolvido é recusado inclusive com --yes.
  - TestAppPurgeSupportsBackendRoots: custom/home/per-mount legítimos funcionam em fixtures isoladas; backend desconhecido ou resolução com erro recusa sem apagar.
- [ ] **Step 2: Confirmar estado anterior.** Executar `go test ./internal/restore ./internal/app -count=1`. Para regressões Go, os novos testes devem falhar pelo comportamento descrito; testes existentes devem passar. Para documentação/CI, registrar as ocorrências que violam o checklist, sem criar testes artificiais.
- [ ] **Step 3: Aplicar a mudança mínima descrita em Architecture.** Manter error handling e fixtures; atualizar callers afetados, inclusive testes que passavam nil no purge se o plano 3 já estiver aplicado. Não corrigir outros achados neste commit.
- [ ] **Step 4: Verificar.** Executar `go test ./internal/restore ./internal/app -count=1` e `go test -race ./internal/restore ./internal/app ./test/integration`. Testes devem passar, diff não deve ter whitespace inválido; no plano 5, revisar manualmente as ocorrências do rg. No plano 6, validar workflows no CI sem acionar release; se ferramentas não estiverem instaladas, registrar validação pendente, nunca declarar aprovação.
- [ ] **Step 5: Revisar e criar commit isolado.** `git diff --check`; revisar somente os arquivos da tarefa; adicionar os arquivos alterados nominalmente e executar `git commit -m "fix: enforce validated trash boundaries for purge"`. Não incluir mudanças preexistentes do usuário.

## Aceitação e limites

Todos os cinco itens de Review Focus devem estar cobertos pela tarefa. Executar `go test ./...` após mudanças Go e `go test -race ./...` ao integrar os planos 1–4. Validar macOS em runner macOS; cross-compilation não substitui execução. Nenhuma implementação foi realizada ao escrever este plano.
