# Alinhar documentação aos comportamentos suportados — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Alinhar documentação aos comportamentos suportados.

**Architecture:** Alterar os pontos existentes e preservar contratos públicos. Editar somente documentação; go.mod é referência, não alvo de mudança. Remover badge zero external dependencies e dizer que o binário não exige instalar bibliotecas Go no destino. Anunciar fail/rename nos exemplos; explicar overwrite reservado e recusado pela CLI. Manter traduções equivalentes e não ampliar promessas de crash consistency.

**Tech Stack:** Go 1.24.0, testing, filesystem local; Markdown/YAML nas tarefas documentais e CI.

**Spec:** [Design da revisão](../specs/2026-11-02-audit-remediation-design.md), critério 5.

## Global Constraints

- Go 1.24.0 como versão mínima.
- Linux e macOS continuam suportados.
- Nenhuma dependência nova de runtime.
- Preservar journal schema_version=1 e registros existentes.
- Não sobrescrever arquivos na restauração nem apagar conteúdo fora da lixeira validada.
- Não publicar releases, alterar credenciais ou executar purge em dados reais durante a validação.

## Review Focus

- Conferência manual: README e README.pt-BR não alegam zero bibliotecas externas.
- Conferência manual: exemplos anunciam fail/rename; overwrite aparece apenas explicado como reservado.
- Conferência manual: nomes das três dependências conferem com go.mod sem removê-las.
- Conferência manual: links relativos continuam válidos.
- Conferência manual: texto de recuperação não promete recuperar eventos nunca persistidos ou resistir a SIGKILL.

## Task 1: Alinhar documentação aos comportamentos suportados

**Files:**
- `README.md`
- `README.pt-BR.md`
- `docs/configuration.md`
- `docs/recovery.md`
- `go.mod`

**Interfaces:**
- Consumes: contratos atuais e critérios 5 do design; ler arquivos listados e todos os callers antes de editar.
- Produces: comportamento descrito em Architecture; sem novas flags ou dependências. Plano documental/CI não altera interfaces Go.

- [ ] **Step 1: Criar regressões ou checklist verificável.**
  - Conferência manual: README e README.pt-BR não alegam zero bibliotecas externas.
  - Conferência manual: exemplos anunciam fail/rename; overwrite aparece apenas explicado como reservado.
  - Conferência manual: nomes das três dependências conferem com go.mod sem removê-las.
  - Conferência manual: links relativos continuam válidos.
  - Conferência manual: texto de recuperação não promete recuperar eventos nunca persistidos ou resistir a SIGKILL.
- [ ] **Step 2: Confirmar estado anterior.** Executar `git diff --check`. Para regressões Go, os novos testes devem falhar pelo comportamento descrito; testes existentes devem passar. Para documentação/CI, registrar as ocorrências que violam o checklist, sem criar testes artificiais.
- [ ] **Step 3: Aplicar a mudança mínima descrita em Architecture.** Manter error handling e fixtures; atualizar callers afetados, inclusive testes que passavam nil no purge se o plano 3 já estiver aplicado. Não corrigir outros achados neste commit.
- [ ] **Step 4: Verificar.** Executar `git diff --check` e `rg -n "overwrite|depend|crash|Power loss" README.md README.pt-BR.md docs/configuration.md docs/recovery.md`. Testes devem passar, diff não deve ter whitespace inválido; no plano 5, revisar manualmente as ocorrências do rg. No plano 6, validar workflows no CI sem acionar release; se ferramentas não estiverem instaladas, registrar validação pendente, nunca declarar aprovação.
- [ ] **Step 5: Revisar e criar commit isolado.** `git diff --check`; revisar somente os arquivos da tarefa; adicionar os arquivos alterados nominalmente e executar `git commit -m "docs: align dependency and restore policy claims"`. Não incluir mudanças preexistentes do usuário.

## Aceitação e limites

Todos os cinco itens de Review Focus devem estar cobertos pela tarefa. Executar `go test ./...` após mudanças Go e `go test -race ./...` ao integrar os planos 1–4. Validar macOS em runner macOS; cross-compilation não substitui execução. Nenhuma implementação foi realizada ao escrever este plano.
