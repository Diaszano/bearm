# Simplificações da auditoria Ponytail — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Resolver os oito achados de complexidade da auditoria de 2026-09-14, preservando o comportamento público da CLI.

**Architecture:** Compartilhar apenas as duas implementações comprovadamente duplicadas; remover estado sem leitores e código inacessível pela CLI. Manter os backends, a política de segurança e os formatos persistidos existentes.

**Tech Stack:** Go 1.24+, biblioteca padrão e dependências já presentes em `go.mod`.

**Spec:** Relatório de auditoria da conversa de 2026-09-14, reproduzido nas oito tarefas abaixo. Este plano é independente do plano já concluído de 2026-09-13.

## Global Constraints

- Não adicionar dependências nem mudar formatos JSON, JSONL ou `.trashinfo`.
- Preservar argumentos, mensagens, códigos de saída, prompts e comportamento GNU/BSD/POSIX.
- Preservar as verificações de segurança, reserva atômica, rollback e tratamento de erros.
- Manter suporte Linux/macOS, amd64/arm64 e Go 1.24.
- Reutilizar testes existentes para refatorações; acrescentar somente verificações que cubram um risco real.
- Não usar redução de linhas como critério de correção. A estimativa da auditoria (~80 linhas de produção) não é uma meta obrigatória.
- Há alterações preexistentes em `.github/workflows/ci.yml`, `.golangci.yml`, `Makefile`, `internal/app/app_extra_test.go` e `internal/platform/dirs.go`. Preservá-las; não incluí-las incidentalmente em commits.
- Não executar este plano como parte de sua criação. Na execução, trabalhar sequencialmente e fazer commits por tarefa apenas conforme o fluxo autorizado, selecionando arquivos ou hunks explicitamente.

## Preparação

- [x] Registrar `git status --short` e `git diff --stat` antes de editar.
- [x] Executar `go test ./...` como referência. Na auditoria, esse comando passou no Linux, incluindo integração e compatibilidade.
- [x] Antes de cada remoção, confirmar os consumidores com `rg`; se surgirem leitores novos, ajustar o corte sem alterar comportamento.

## Mapa dos arquivos

| Tarefa | Arquivos | Responsabilidade |
| --- | --- | --- |
| 1 | `internal/platform/device_{linux,darwin}.go`; novo `internal/platform/mount_unix.go` | Compartilhar descoberta do ponto de montagem |
| 2 | `internal/trash/{custom,darwin}/backend.go`; novos `internal/trash/metadata.go` e `metadata_test.go` | Compartilhar codificação JSON |
| 3 | `internal/domain/removal.go`, `internal/planner/planner.go`, `internal/cli/invocation.go`, `internal/cli/render_compatibility.go` | Remover campos sem leitores |
| 4 | `internal/restore/doctor.go`, `doctor_test.go`, `internal/app/app.go` | Converter serviço sem estado próprio em função |
| 5 | `internal/restore/service.go`, `service_test.go` | Remover implementação de sobrescrita bloqueada pela CLI |
| 6 | `internal/planner/protected_walk.go` | Remover ordenação redundante |
| 7 | `internal/config/file.go` | Retornar diretamente a função existente |
| 8 | `internal/cli/errors.go` | Remover variante de erro sem produtor |

## Tarefa 1: Compartilhar `MountPoint`

**Interfaces:** Preservar `MountPoint(string) (string, error)` e `DeviceID(string) (uint64, error)`.

- [x] Conferir consumidores e testes:

```sh
rg -n 'MountPoint|DeviceID' internal
go test ./internal/platform ./internal/trash/linux
```

- [x] Criar `internal/platform/mount_unix.go` com este cabeçalho e mover para ele, sem alterar o corpo, a função `MountPoint` de `device_linux.go`:

```go
//go:build linux || darwin

package platform

import (
    "path/filepath"
    "golang.org/x/sys/unix"
)
```

- [x] Remover `MountPoint` e o import `path/filepath` dos dois arquivos específicos. Manter `DeviceID` em cada um: Darwin necessita da conversão de `stat.Dev` para `uint64`.
- [x] Rodar `go test ./internal/platform ./internal/trash/linux` no Linux e `go test ./internal/platform ./internal/trash/darwin` no macOS. Esperado: PASS, incluindo os testes existentes de arquivo, diretório e symlink.

## Tarefa 2: Compartilhar metadados JSON

**Interfaces:** Criar `trash.RenderMetadata(path string, deletedAt time.Time) ([]byte, error)`. Consumidores: somente custom e Darwin; Linux continua usando `RenderTrashInfo`.

- [x] Criar `internal/trash/metadata_test.go`, pacote `trash_test`, com imports `bytes`, `testing`, `time` e `github.com/Diaszano/bearm/internal/trash`, e o teste:

```go
func TestRenderMetadata(t *testing.T) {
    when := time.Date(2026, 9, 14, 9, 0, 0, 0, time.FixedZone("BRT", -3*60*60))
    got, err := trash.RenderMetadata("/tmp/ação\n.txt", when)
    want := "{\"schema_version\":1,\"original_path\":\"/tmp/ação\\n.txt\",\"deleted_at\":\"2026-09-14T12:00:00Z\"}\n"
    if err != nil || string(got) != want {
        t.Fatalf("metadata = %q, %v; want %q", got, err, want)
    }
    _, err = trash.RenderMetadata("/tmp/file", time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC))
    if err == nil {
        t.Fatal("expected time serialization error")
    }
    if bytes.Count(got, []byte{'\n'}) != 1 {
        t.Fatal("metadata must contain exactly one trailing newline")
    }
}
```

- [x] Rodar `go test ./internal/trash -run TestRenderMetadata`; antes da implementação, deve falhar pela função inexistente.
- [x] Criar `internal/trash/metadata.go`, pacote `trash`, com imports `encoding/json` e `time`:

```go
// RenderMetadata encodes private trash metadata as one JSON line.
func RenderMetadata(path string, deletedAt time.Time) ([]byte, error) {
    data, err := json.Marshal(struct {
        SchemaVersion int       `json:"schema_version"`
        OriginalPath  string    `json:"original_path"`
        DeletedAt     time.Time `json:"deleted_at"`
    }{1, path, deletedAt.UTC()})
    if err != nil {
        return nil, err
    }
    return append(data, '\n'), nil
}
```

- [x] Nos dois backends, substituir os blocos `json.Marshal` e `append` pela chamada abaixo, mantendo o tratamento de erro existente e removendo o import `encoding/json`:

```go
metadata, err := trash.RenderMetadata(target.AbsolutePath, deletedAt)
if err != nil {
    return domain.TrashRecord{}, err
}
```

- [x] Rodar `go test ./internal/trash ./internal/trash/custom`; no macOS, incluir `./internal/trash/darwin`. Esperado: PASS, preservando os testes de reserva e rollback.

## Tarefa 3: Remover três campos sem leitores

**Interfaces:** `RemovalPlan`, `Invocation` e `CompatibilityRenderer` mantêm seus campos utilizados e suas funções públicas existentes.

- [x] Confirmar referências:

```sh
rg -n 'CreatedAt|Program|language' internal/domain internal/planner internal/cli
```

- [x] Remover estas declarações e suas atribuições:

```go
// internal/domain/removal.go e internal/planner/planner.go
CreatedAt time.Time
CreatedAt: time.Now().UTC(),

// internal/cli/invocation.go
Program string
Program: program,
Program: "rm",

// internal/cli/render_compatibility.go
language i18n.Language
language: language,
```

- [x] Remover os imports `time` agora sem uso em `removal.go` e `planner.go`. Manter a variável local `program`, usada para escolher o modo, e o parâmetro `language`, usado para construir o catálogo.
- [x] Rodar `go test ./internal/domain ./internal/planner ./internal/cli ./internal/app`. Esperado: PASS sem mudança nas expectativas dos testes.

## Tarefa 4: Converter `Doctor` em função

**Interfaces:** Substituir `NewDoctor(repository).Check(ctx)` por `restore.Check(ctx, repository)`; manter `Finding` e os resultados atuais.

- [x] Remover a struct `Doctor` e seu construtor. Trocar o cabeçalho do método e as três referências `d.repository`:

```go
// Check returns diagnostics without mutating journal or trash.
func Check(ctx context.Context, repository *journal.Repository) ([]Finding, error) {
    // Corpo existente, usando repository em vez de d.repository.
}
```

O corpo permanece integralmente igual: `ReadAll`, diagnóstico de linha incompleta, `ActiveItems`, diagnóstico de arquivo ausente e retorno. Não refatorar as leituras do journal nesta tarefa.

- [x] Em `internal/app/app.go`, usar:

```go
findings, err := restore.Check(ctx, a.dependencies.Repository)
```

- [x] Nos dois testes em `doctor_test.go`, substituir construção e chamada por `restore.Check(context.Background(), repository)`; no teste de linha incompleta, passar `journal.New(path, time.Now)` diretamente.
- [x] Rodar `go test ./internal/restore ./internal/app`; confirmar ausência de consumidores com `rg -n 'NewDoctor|type Doctor' internal`. Esperado: testes passam e busca não encontra resultados.

## Tarefa 5: Remover implementação de sobrescrita

**Interfaces:** Manter a constante `CollisionOverwrite`, a aceitação do valor na configuração e a rejeição explícita já existente em `App.runRestore`. Remover somente o ramo destrutivo de colisão em `atomicRestore`.

- [x] Confirmar todos os consumidores com `rg -n 'CollisionOverwrite|atomicRestore|overwrite' internal docs/configuration.md docs/recovery.md`.
- [x] Renomear `TestRestoreCollisionOverwrite` para `TestRestoreCollisionOverwriteRejected`. Reutilizar a preparação e a chamada atuais; trocar as verificações finais por:

```go
if len(results) != 1 || results[0].Status != domain.ItemFailed {
    t.Fatalf("results = %#v", results)
}
for _, path := range []string{original, trashed} {
    if got, err := os.ReadFile(path); err != nil || string(got) != path {
        t.Fatalf("file changed: %s: %q, %v", path, got, err)
    }
}
```

- [x] Rodar `go test ./internal/restore -run TestRestoreCollisionOverwriteRejected`; esperado antes da remoção: FAIL.
- [x] Remover integralmente este ramo de `atomicRestore`; deixar a política cair no erro `unsupported restore collision policy` existente:

```go
case CollisionOverwrite:
    if err := os.RemoveAll(dst); err != nil {
        return "", err
    }
    if err := os.Rename(src, dst); err != nil {
        return "", err
    }
    return dst, nil
```

- [x] Atualizar o comentário da constante para `CollisionOverwrite is reserved and rejected by the CLI.` Não mudar o caminho sem colisão nem criar um fluxo de confirmação novo.
- [x] Rodar `go test ./internal/restore ./internal/app`, preservando `TestRunRestore_CollisionOverwritePolicyRejected`. Esperado: PASS e arquivos de origem/destino intactos na colisão rejeitada.

## Tarefa 6: Remover ordenação redundante

**Interfaces:** `ExpandTarget` mantém ordem determinística e filhos antes dos pais.

- [x] Em `protected_walk.go`, remover o import `sort` e este bloco; `os.ReadDir(path)` já entrega entradas ordenadas por nome:

```go
sort.Slice(children, func(i, j int) bool {
    return children[i].Name() < children[j].Name()
})
```

- [x] Rodar `go test ./internal/planner ./test/compatibility`. Esperado: PASS, sem alterar expectativas de travessia ou saída.

## Tarefa 7: Simplificar o retorno de `config.Load`

**Interfaces:** `Load(path string, getenv func(string) string) (Config, error)` permanece igual. `ApplyEnvironment` já retorna `Config{}` em caso de erro.

- [x] Substituir as cinco linhas finais de `Load` por:

```go
return ApplyEnvironment(value, getenv)
```

- [x] Rodar `go test ./internal/config`; esperado: PASS para configuração válida, inválida e overrides. Não alterar os caminhos de arquivo ausente nem as verificações de propriedade/permissões.

## Tarefa 8: Remover erro sem produtor

**Interfaces:** `UsageError` mantém `Kind`, `Option`, `Value` e as mensagens de variantes usadas.

- [x] Confirmar que `rg -n 'missing-value' internal cmd test` encontra apenas o ramo em `errors.go`.
- [x] Remover:

```go
case "missing-value":
    return fmt.Sprintf("missing value for %q", e.Option)
```

- [x] Rodar `go test ./internal/cli ./test/compatibility`; esperado: PASS sem ajustes nas mensagens esperadas.

## Verificação final

- [x] Formatar somente os arquivos Go alterados com `gofmt -w` e revisar `git diff --check` e `git diff --stat`.
- [x] Executar `make verify` com as ferramentas do projeto disponíveis. Ele inclui formatação, vet, lint, testes e race; não repetir a suíte completa após sucesso sem uma mudança ou falha que justifique.
- [x] Validar compilação cruzada em diretório temporário:

```sh
audit_build_dir="$(mktemp -d /tmp/bearm-audit-build.XXXXXX)"
for audit_os in linux darwin; do
    for audit_arch in amd64 arm64; do
        GOOS="$audit_os" GOARCH="$audit_arch" CGO_ENABLED=0 \
            go build -o "$audit_build_dir/bearm-$audit_os-$audit_arch" ./cmd/bearm
    done
done
```

- [x] Confirmar execução dos testes Darwin em host macOS/CI. Compilação cruzada não equivale à execução desses testes; registrar essa limitação se o host não estiver disponível.
- [x] Revisar o diff: oito achados cobertos, nenhuma dependência nova, nenhum formato persistido alterado e nenhuma alteração preexistente perdida.
- [x] Relatar cortes efetivamente obtidos e comandos executados, sem afirmar que checks não executados passaram.
