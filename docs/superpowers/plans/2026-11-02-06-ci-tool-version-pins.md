# Padronizar versões das ferramentas de CI — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Padronizar versões das ferramentas de CI.

**Architecture:** Alterar os pontos existentes e preservar contratos públicos. Fixar GoReleaser em 2.17.0 e govulncheck em v1.6.0, versões já usadas no readiness. Para golangci-lint, usar versão explícita já referenciada em config/scripts/docs; se nenhuma existir, selecionar versão compatível com .golangci.yml e registrar a escolha antes da alteração. Substituir curl de HEAD por instalação oficial de versão fixa e action fixada por SHA, preservando fmt/run. Não criar workflow reutilizável nem alterar matriz Go. Inspecionar scripts e alinhar instalações duplicadas das mesmas ferramentas.

**Tech Stack:** Go 1.24.0, testing, filesystem local; Markdown/YAML nas tarefas documentais e CI.

**Spec:** [Design da revisão](../specs/2026-11-02-audit-remediation-design.md), critério 6.

## Global Constraints

- Go 1.24.0 como versão mínima.
- Linux e macOS continuam suportados.
- Nenhuma dependência nova de runtime.
- Preservar journal schema_version=1 e registros existentes.
- Não sobrescrever arquivos na restauração nem apagar conteúdo fora da lixeira validada.
- Não publicar releases, alterar credenciais ou executar purge em dados reais durante a validação.

## Review Focus

- Conferência: nenhum install de govulncheck usa @latest; todos usam v1.6.0.
- Conferência: todos os jobs GoReleaser usam 2.17.0, inclusive snapshot e readiness.
- Conferência: golangci-lint tem versão exata, sem instalador HEAD.
- Conferência: matriz 1.24.x/stable e comandos fmt/run permanecem.
- Validação de snapshot: goreleaser check passa; nenhuma execução publica release ou altera Homebrew.

## Task 1: Padronizar versões das ferramentas de CI

**Files:**
- `.github/workflows/ci.yml`
- `.github/workflows/security.yml`
- `.github/workflows/release-readiness.yml`
- `.github/workflows/release.yml`
- `scripts/closure-gates.sh`

**Interfaces:**
- Consumes: contratos atuais e critérios 6 do design; ler arquivos listados e todos os callers antes de editar.
- Produces: comportamento descrito em Architecture; sem novas flags ou dependências. Plano documental/CI não altera interfaces Go.

- [ ] **Step 1: Criar regressões ou checklist verificável.**
  - Conferência: nenhum install de govulncheck usa @latest; todos usam v1.6.0.
  - Conferência: todos os jobs GoReleaser usam 2.17.0, inclusive snapshot e readiness.
  - Conferência: golangci-lint tem versão exata, sem instalador HEAD.
  - Conferência: matriz 1.24.x/stable e comandos fmt/run permanecem.
  - Validação de snapshot: goreleaser check passa; nenhuma execução publica release ou altera Homebrew.
- [ ] **Step 2: Confirmar estado anterior.** Executar `git diff --check`. Para regressões Go, os novos testes devem falhar pelo comportamento descrito; testes existentes devem passar. Para documentação/CI, registrar as ocorrências que violam o checklist, sem criar testes artificiais.
- [ ] **Step 3: Aplicar a mudança mínima descrita em Architecture.** Manter error handling e fixtures; atualizar callers afetados, inclusive testes que passavam nil no purge se o plano 3 já estiver aplicado. Não corrigir outros achados neste commit.
- [ ] **Step 4: Verificar.** Executar `git diff --check` e `goreleaser check`. Testes devem passar, diff não deve ter whitespace inválido; no plano 5, revisar manualmente as ocorrências do rg. No plano 6, validar workflows no CI sem acionar release; se ferramentas não estiverem instaladas, registrar validação pendente, nunca declarar aprovação.
- [ ] **Step 5: Revisar e criar commit isolado.** `git diff --check`; revisar somente os arquivos da tarefa; adicionar os arquivos alterados nominalmente e executar `git commit -m "ci: pin consistent tool versions across workflows"`. Não incluir mudanças preexistentes do usuário.

## Aceitação e limites

Todos os cinco itens de Review Focus devem estar cobertos pela tarefa. Executar `go test ./...` após mudanças Go e `go test -race ./...` ao integrar os planos 1–4. Validar macOS em runner macOS; cross-compilation não substitui execução. Nenhuma implementação foi realizada ao escrever este plano.
