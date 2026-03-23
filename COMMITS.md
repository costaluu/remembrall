# Guia de Commits — Pipeline de Versionamento Automático

> Este guia define como escrever commits que se integram corretamente ao
> pipeline de versionamento automático. Cada commit que chega à `master`
> **pode gerar uma nova release**, portanto seguir este padrão é obrigatório.

---

## Por que isso importa?

O pipeline lê as mensagens dos commits para decidir automaticamente qual
versão publicar. Se os commits forem escritos errado, a versão errada é
gerada — ou nenhuma versão é gerada.

```
Commit mal escrito  →  versão errada publicada  →  auto-update do CLI quebrado
Commit bem escrito  →  versão correta publicada  →  tudo funciona
```

---

## Anatomia de um commit

```
<tipo>(<escopo opcional>): <descrição curta>

<corpo opcional>

<rodapé opcional — onde vai BREAKING CHANGE>
```

### Campos

| Campo | Obrigatório | Descrição |
|---|---|---|
| `tipo` | ✅ | Categoria da mudança (`feat`, `fix`, `chore`, etc.) |
| `escopo` | ❌ | Parte do sistema afetada, entre parênteses |
| `descrição` | ✅ | Resumo imperativo, em minúsculas, sem ponto final |
| `corpo` | ❌ | Detalhes adicionais separados por linha em branco |
| `rodapé` | ❌ | `BREAKING CHANGE:` obrigatório quando há quebra de contrato |

---

## Tipos e seu impacto na versão

O pipeline usa exatamente estas regras de prioridade:

```
BREAKING CHANGE  →  MAJOR  (vX+1.0.0)
feat             →  MINOR  (vX.Y+1.0)
qualquer outro   →  PATCH  (vX.Y.Z+1)
```

### Tabela completa de tipos

| Tipo | Impacto na versão | Quando usar |
|---|---|---|
| `feat` | **MINOR** | Nova funcionalidade visível ao usuário |
| `fix` | PATCH | Correção de bug |
| `perf` | PATCH | Melhoria de desempenho sem nova feature |
| `refactor` | PATCH | Refatoração interna sem mudança de comportamento |
| `docs` | PATCH | Mudanças apenas em documentação |
| `test` | PATCH | Adição ou correção de testes |
| `chore` | PATCH | Tarefas de manutenção (deps, configs, CI) |
| `build` | PATCH | Mudanças no sistema de build ou dependências externas |
| `ci` | PATCH | Mudanças nos arquivos de CI/CD |
| `style` | PATCH | Formatação de código (sem mudança de lógica) |
| `revert` | PATCH | Reversão de um commit anterior |
| `BREAKING CHANGE` | **MAJOR** | Qualquer tipo com quebra de compatibilidade |

> **Regra de ouro:** se não for `feat` e não tiver `BREAKING CHANGE`,
> o pipeline vai incrementar o PATCH independente do tipo.

---

## Exemplos práticos

### PATCH — correção de bug

```
fix: corrige parsing de flags com valor vazio
```

```
fix(config): lê corretamente variáveis de ambiente com espaços
```

```
fix: remove panic ao receber argumento desconhecido

O CLI encerrava com panic quando recebia flags não registradas.
Agora retorna erro amigável com lista de flags válidas.
```

---

### MINOR — nova funcionalidade

```
feat: adiciona comando update
```

```
feat(auth): implementa login via token de API
```

```
feat: adiciona flag --output para exportar resultado em JSON

Permite que outros CLIs consumam a saída diretamente sem precisar
fazer parsing do texto formatado.
```

---

### MAJOR — breaking change

A `BREAKING CHANGE` pode aparecer de **duas formas**:

**Forma 1 — no rodapé (recomendada):**

```
feat: substitui flags --host e --port por --endpoint

BREAKING CHANGE: as flags --host e --port foram removidas.
Use --endpoint=http://host:porta no lugar.
```

**Forma 2 — com `!` após o tipo:**

```
feat!: substitui flags --host e --port por --endpoint
```

```
fix!: remove suporte ao formato de config legado .ini
```

> ⚠️ O pipeline detecta `BREAKING CHANGE` no texto do commit.
> O `!` sozinho **não é detectado** pelo pipeline atual — use o rodapé
> ou inclua `BREAKING CHANGE` na mensagem.

---

## Regras de escrita

### ✅ Faça

- Escreva a descrição no **imperativo**: *"adiciona"*, *"corrige"*, *"remove"*
- Use **letras minúsculas** na descrição
- Seja **específico**: diga o que mudou, não o que você fez
- Mantenha a **primeira linha com até 72 caracteres**
- Separe corpo do título com **linha em branco**

### ❌ Não faça

```bash
# Sem tipo — pipeline trata como PATCH mas a intenção é perdida
corrige bug no login

# Tipo errado capitalizado
Fix: corrige bug no login

# Ponto final na descrição
fix: corrige bug no login.

# Vago demais
fix: ajustes

feat: melhorias

chore: várias coisas

# Múltiplas mudanças num commit só
feat: adiciona login e corrige parser e atualiza docs
```

---

## Commits que não geram nova versão

O pipeline só cria uma release se existirem **commits novos desde a última tag**.

Tipos que tecnicamente geram PATCH mas raramente justificam release manual:

```
docs: atualiza README com exemplos de uso
style: aplica gofmt em todos os arquivos
test: adiciona testes para o comando config
ci: ajusta timeout do job de build
```

> Esses commits são válidos e corretos — eles vão para a `master` e,
> quando combinados com um `feat` ou `fix` no mesmo ciclo, contribuem
> para a release normalmente.

---

## Múltiplos commits no mesmo ciclo

Quando vários commits chegam à `master` antes de uma release, o pipeline
analisa **todos eles juntos** e aplica a regra de **maior prioridade**:

```
fix: corrige timeout de conexão          → PATCH
feat: adiciona suporte a proxy HTTP      → MINOR  ← vence
docs: atualiza exemplos no README        → PATCH
```

Resultado: versão incrementa **MINOR** (o `feat` domina).

```
fix: corrige leitura de config           → PATCH
feat!: remove comando deprecated sync   → MAJOR  ← vence
feat: adiciona comando migrate           → MINOR
```

Resultado: versão incrementa **MAJOR** (o `BREAKING CHANGE` domina).

---

## Fluxo de trabalho recomendado

### Desenvolvimento em feature branch

```bash
# Crie a branch a partir da master atualizada
git checkout master && git pull
git checkout -b feat/comando-export

# Faça commits atômicos e descritivos durante o desenvolvimento
git commit -m "feat(export): adiciona estrutura básica do comando"
git commit -m "feat(export): implementa formato CSV"
git commit -m "feat(export): implementa formato JSON"
git commit -m "test(export): adiciona testes para ambos os formatos"
git commit -m "docs: documenta flags do comando export"
```

### Merge para master

```bash
# Opção 1: squash (um commit limpo na master)
git checkout master
git merge --squash feat/comando-export
git commit -m "feat(export): adiciona comando export com suporte a CSV e JSON"

# Opção 2: merge commit (preserva histórico)
git merge feat/comando-export

# Opção 3: rebase (histórico linear)
git rebase master feat/comando-export
git checkout master
git merge --ff-only feat/comando-export
```

> Após o push para `master`, o pipeline executa automaticamente.

---

## Checklist antes do push para master

```
[ ] A mensagem começa com um tipo válido (feat, fix, chore, etc.)
[ ] A descrição está em minúsculas e no imperativo
[ ] Não há ponto final na primeira linha
[ ] A primeira linha tem no máximo 72 caracteres
[ ] Se há quebra de compatibilidade, "BREAKING CHANGE" está no rodapé
[ ] O commit faz UMA coisa só (atômico)
[ ] O código está funcionando (não quebra o build)
```

---

## Referência rápida

```bash
# PATCH
git commit -m "fix: <o que foi corrigido>"
git commit -m "chore: <o que foi feito>"
git commit -m "docs: <o que foi documentado>"
git commit -m "test: <o que foi testado>"
git commit -m "refactor: <o que foi refatorado>"

# MINOR
git commit -m "feat: <o que foi adicionado>"
git commit -m "feat(<escopo>): <o que foi adicionado>"

# MAJOR
git commit -m "feat: <descrição>

BREAKING CHANGE: <o que quebrou e como migrar>"

git commit -m "fix: <descrição>

BREAKING CHANGE: <o que quebrou e como migrar>"
```

---

## Como verificar antes de commitar

```bash
# Visualize o que o pipeline vai capturar
git log $(git describe --tags --abbrev=0)..HEAD --pretty=format:"%s"

# Simule o tipo de bump que seria gerado
git log $(git describe --tags --abbrev=0)..HEAD --pretty=format:"%s" | \
  grep -qiE 'BREAKING CHANGE' && echo "MAJOR" || \
  (git log $(git describe --tags --abbrev=0)..HEAD --pretty=format:"%s" | \
  grep -qE '^feat' && echo "MINOR" || echo "PATCH")
```

---

*Qualquer commit que não siga o padrão de tipo será tratado como **PATCH**
pelo pipeline — a release será gerada, mas o versionamento pode não
refletir a real magnitude da mudança.*