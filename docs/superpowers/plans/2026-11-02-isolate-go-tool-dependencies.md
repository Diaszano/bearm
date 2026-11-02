# Isolar dependências de ferramentas Go em `tools/` — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Gerenciar as ferramentas Go de desenvolvimento e CI em um módulo `tools/` independente, sem adicionar suas dependências transitivas ao `go.mod` principal.

**Architecture:** Criar `tools/go.mod` e `tools/go.sum` com diretivas `tool` para `golangci-lint`, `govulncheck` e GoReleaser, pinando golangci-lint em 2.15.0 (o 2.3.0 falha ao ler export data do Go 1.27.3), govulncheck em 1.6.0 e GoReleaser em 2.17.0. Executar as ferramentas com `go tool -modfile=tools/go.mod ...`, mantendo o diretório raiz do repositório como diretório de trabalho, e atualizar Makefile, script de acceptance e GitHub Actions. O módulo principal continua contendo apenas dependências necessárias para compilar/testar o Bearm.

**Tech Stack:** Go 1.26.4+ para as ferramentas (GoReleaser v2.17.0 exige essa versão), módulo principal mantido em Go 1.24.0, módulos Go, diretivas `tool`, Makefile e GitHub Actions.

**Spec:** Escopo definido neste plano; versões de ferramentas devem coincidir com os pins ativos em `.github/workflows/` na implementação.

## Scope

Inclui ferramentas Go atualmente chamadas pelo projeto: `github.com/golangci/golangci-lint/v2/cmd/golangci-lint`, `golang.org/x/vuln/cmd/govulncheck` e `github.com/goreleaser/goreleaser/v2`. As versões ficam definidas uma vez em `tools/go.mod` e são consumidas localmente e no CI: golangci-lint 2.15.0, govulncheck 1.6.0 e GoReleaser 2.17.0.

Syft e Cosign continuam instalados por suas GitHub Actions fixadas por SHA: são binários externos, não dependências de módulo Go. Não mover dependências de runtime do produto para `tools/`, não atualizar pins sem motivo e não adicionar bibliotecas auxiliares.

## Global Constraints

- O módulo principal não ganha dependências usadas somente por lint, security ou release.
- `tools/go.mod` é um módulo separado, declara Go 1.26.4 e versiona suas dependências em `tools/go.sum`; o `go` directive de raiz permanece inalterado.
- Todos os comandos de ferramentas usados em Makefile, scripts e workflows devem obter a versão do módulo `tools` usando `go tool -modfile=tools/go.mod`; a execução permanece na raiz do projeto. O módulo usa Go 1.26.4 por requisito do GoReleaser 2.17.0.
- Preservar comandos e gates atuais: fmt, lint, govulncheck, GoReleaser check e snapshot.
- A instalação não pode depender de `@latest`, `HEAD` ou binários globais preexistentes.

## Review Focus

- `go.mod` principal permanece sem dependências das três ferramentas e sem mudanças transitivas sem relação.
- Todos os call sites usam `go tool -modfile=tools/go.mod`; lint e GoReleaser recebem caminhos/configuração a partir da raiz do projeto.
- Os comandos rodam com diretório raiz/configuração de GoReleaser corretos, apesar do módulo separado.
- O módulo `tools` usa Go 1.26.4, exigido pelo GoReleaser v2.17.0; o mínimo de Go do produto permanece 1.24.0. Jobs que fixam Go para executar ferramentas usam versão compatível.
- `scripts/acceptance.sh` invoca as três ferramentas Go pelo módulo `tools` sem binários globais; snapshots de acceptance ignoram assinatura e ignoram SBOM se Syft faltar.

---

## Task 1: Criar o módulo de ferramentas

**Files:**
- Create: `tools/go.mod`
- Create: `tools/go.sum`

**Interfaces:**
- Consumes: versão Go mínima declarada em `go.mod` do projeto e pins existentes nos workflows.
- Produces: diretivas `tool` para as três ferramentas; comandos invocáveis por `go tool -modfile=tools/go.mod <nome>`.

- [ ] **Step 1: Criar o módulo separado.** Definir `module github.com/Diaszano/bearm/tools`, usar `go 1.26.4` (exigido pelo GoReleaser pinado, sem alterar o mínimo do módulo principal) e adicionar as diretivas `tool` com versões fixas que correspondam aos workflows existentes.
- [ ] **Step 2: Gerar e validar o lockfile.** Executar `go mod tidy` em `tools/`; depois confirmar os pins com `go tool -modfile=tools/go.mod golangci-lint --version`, `go tool -modfile=tools/go.mod govulncheck -version` e `go tool -modfile=tools/go.mod goreleaser --version`.
- [ ] **Step 3: Confirmar isolamento.** Executar `git diff -- go.mod go.sum`; nenhum dos dois arquivos do módulo principal deve mudar por causa da instalação das ferramentas. Confirmar os módulos e diretivas com `go list -modfile=tools/go.mod -m all` e `go tool -modfile=tools/go.mod` para listar as diretivas disponíveis.

## Task 2: Migrar comandos locais e acceptance

**Files:**
- Modify: `Makefile`
- Modify: `scripts/acceptance.sh`
- Modify if needed: `docs/installation.md` ou documentação de contribuição que instrui instalação manual de ferramentas

**Interfaces:**
- Consumes: módulo `tools` e comandos `go tool -modfile=tools/go.mod ...` da Task 1.
- Produces: alvos locais resolvendo as ferramentas pelas versões do módulo separado, executadas a partir da raiz.

- [ ] **Step 1: Atualizar os alvos Makefile.** Usar `go tool -modfile=tools/go.mod golangci-lint` para `fmt`, `fmt-check`, `lint` e `lint-fix`; usar GoReleaser do módulo de tools no alvo `snapshot`, mantendo `--snapshot --clean`; a invocação a partir da raiz deve permitir resolver o arquivo de configuração e os hooks existentes.
- [ ] **Step 2: Atualizar `scripts/acceptance.sh`.** Remover detecção de binários globais como requisito; invocar as ferramentas Go via módulo `tools`, mantendo a raiz como diretório atual. Ignorar assinatura em snapshots e ignorar SBOM quando Syft não estiver disponível, como já faz o script.
- [ ] **Step 3: Validar comportamento local.** Rodar `make fmt-check`, `make lint`, `make snapshot` e `./scripts/acceptance.sh` sem depender de instalações globais; cada alvo deve usar os pins do módulo tools.

## Task 3: Migrar workflows GitHub Actions

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/security.yml`
- Modify: `.github/workflows/release.yml`
- Modify: `.github/workflows/release-readiness.yml`

**Interfaces:**
- Consumes: módulo `tools` com versões fixas.
- Produces: jobs de CI/release usando ferramentas declaradas em `tools/go.mod`.

- [ ] **Step 1: Remover instalações ad hoc.** Remover `go install ...@version`, downloads pelo instalador de golangci-lint e invocações GoReleaser action que escolhem uma versão fora do módulo; conservar pins de SHA das actions externas Syft e Cosign.
- [ ] **Step 2: Executar as ferramentas pelo módulo.** Adicionar `go tool -modfile=tools/go.mod ...` aos jobs lint, security, release, snapshot e readiness, passando paths/configuração relativos à raiz do repositório. Manter o Go setup necessário antes das invocações.
- [ ] **Step 3: Preservar gates e permissões.** Não remover `go vet`, testes, race tests, fuzz, acceptance, `govulncheck`, validações de assinatura ou checks de release; não ampliar permissões de workflows.
- [ ] **Step 4: Validar referências.** Buscar `go install`, `@latest`, `golangci-lint` e `goreleaser` em workflows/scripts/Makefile. As únicas versões Go dessas ferramentas devem estar no `tools/go.mod`; actions externas permanecem fixadas por SHA.

## Task 4: Verificação e entrega

**Files:**
- Review: `tools/`, `Makefile`, `scripts/acceptance.sh`, `.github/workflows/`, docs tocadas

- [ ] **Step 1: Validar módulo principal.** Executar `go mod tidy -diff`, `go mod verify`, `go test ./...`, `go test -race ./...` e `go vet ./...` na raiz; todos devem passar e não modificar os arquivos do módulo principal.
- [ ] **Step 2: Validar ferramentas.** Executar lint, security scan e snapshot via módulo `tools`; confirmar versões com `go list -modfile=tools/go.mod -m all` e `go tool -modfile=tools/go.mod <tool> --version` conforme suporte de cada comando.
- [ ] **Step 3: Validar GitHub Actions.** Confirmar que todos workflows Linux/macOS relevantes passam e que release/readiness executam snapshot sem publicar artifacts, tags, releases ou cask.
- [ ] **Step 4: Revisar documentação e diff.** `git diff --check`; explicar a invocação local via Makefile e onde editar pins de ferramentas. Confirmar ausência de dependências de ferramentas no `go.mod` principal.
- [ ] **Step 5: Commit isolado** com mensagem `build: isolate Go tool dependencies in tools module`.

## Aceitação e limites

O plano está completo quando ferramentas Go são declaradas uma única vez no módulo `tools` (golangci-lint 2.15.0, govulncheck 1.6.0, GoReleaser 2.17.0), todos os call sites locais e CI as executam usando esse módulo a partir da raiz, `go.mod`/`go.sum` do produto ficam sem dependências de tooling, os gates preservados passam e não ocorre publicação durante a validação. Syft/Cosign permanecem como actions externas fixadas por SHA. Nenhuma implementação foi feita ao escrever este plano.
