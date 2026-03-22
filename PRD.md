```markdown id="c7xk2p"
# 📄 PRD – Remembrall (CLI Reminder App)

## 🧭 Visão Geral

**Remembrall** é um aplicativo CLI desenvolvido em Go para gerenciamento de lembretes locais, com persistência em SQLite. O objetivo é ser um software **open source, cross-platform e production-ready**, com forte foco em automação de DevOps, distribuição simplificada e manutenção de longo prazo.

---

# 🔧 DevOps, Versionamento e Distribuição

---

## 1. 📦 Distribuição e Releases

- O aplicativo será distribuído via **GitHub Releases**
- Serão gerados binários para:
  - Windows
  - Linux
  - macOS

### 📌 Automação

- O processo de build e release será totalmente automatizado via CI/CD
- A cada merge na branch `main`, será gerada uma nova release

---

## 2. 🌿 Estratégia de Branches

- `main`
  - Contém código estável
  - Gera releases oficiais

- `dev`
  - Contém código em desenvolvimento
  - Versão identificada como `dev`
  - Não gera releases oficiais

---

## 3. 🏷️ Estratégia de Versionamento (Automatizada)

O projeto utiliza **Semantic Versioning (SemVer)** com geração automática de versões via CI/CD.

### 📌 Objetivo

Eliminar completamente a necessidade de controle manual de versão pelo desenvolvedor.

---

### 🔢 Formato

```

MAJOR.MINOR.PATCH[-PRERELEASE]

````

---

### 📌 Exemplos

- `1.0.0-beta.1`
- `1.2.0`
- `1.2.1`

---

### 🔄 Estratégia de geração

Baseado em **Conventional Commits**:

| Tipo de commit | Impacto na versão |
|----------------|------------------|
| `feat:`        | Incrementa MINOR |
| `fix:`         | Incrementa PATCH |
| `feat!:`       | Incrementa MAJOR |

---

### 🤖 Responsabilidade do CI

- Determinar próxima versão automaticamente
- Criar tag
- Gerar changelog
- Publicar release

---

## 4. 🔍 Verificação de Atualização

O CLI deve ser capaz de:

```bash
remembrall update check
````

### 📌 Comportamento

* Consultar GitHub Releases
* Comparar versão local com versão mais recente
* Informar ao usuário se há atualização disponível

---

## 5. 🔄 Atualização do Aplicativo

O CLI deve suportar:

```bash
remembrall update apply
```

---

### 📌 Comportamento

* Baixar novo binário
* Substituir o atual
* Executar migrações de banco automaticamente

---

## 6. 🗄️ Estratégia de Migração de Banco de Dados

O Remembrall utilizará SQLite com um sistema de migração **incremental e versionado**, baseado exclusivamente em **forward migrations (up migrations)**.

---

### 📌 Princípios

* Não haverá suporte a rollback de schema
* Todas as migrações são **imutáveis**
* Todas as migrações devem ser distribuídas junto com o aplicativo
* O sistema garante apenas evolução progressiva do banco de dados

---

### 📦 Distribuição de Migrations

Cada release do Remembrall deve incluir:

* Todas as migrations históricas desde a versão inicial
* As migrations devem estar disponíveis localmente no binário ou empacotadas junto ao aplicativo

👉 Isso garante que qualquer usuário, independente da versão atual, consiga atualizar corretamente.

---

### 📌 Estrutura de Migrations

```
migrations/
  0001_init.sql
  0002_add_reminder_table.sql
  0003_add_due_date.sql
```

---

### 📌 Controle de versão do schema

O sistema deve utilizar uma tabela interna:

```
schema_migrations
```

Responsável por armazenar:

* Versão atual aplicada no banco

---

### ⚙️ Execução de Migrations

Durante a execução do aplicativo (ou update):

1. Ler versão atual do banco
2. Identificar migrations pendentes
3. Executar migrations em ordem crescente
4. Atualizar versão na tabela `schema_migrations`

---

### 📌 Cenário suportado

Usuário está na versão:

```
1.0.0
```

Atualiza para:

```
1.3.0
```

O sistema deve executar automaticamente:

```
1.1 → 1.2 → 1.3
```

---

### 🛡️ Regras obrigatórias

* Nunca modificar migrations antigas
* Sempre criar novas migrations
* Migrations devem ser determinísticas
* Execução deve parar imediatamente em caso de erro

---

### 💾 Segurança operacional

Antes de executar migrations:

* Deve ser realizado backup automático do banco SQLite

---

### ⚠️ Limitações e responsabilidade

* Não há suporte a rollback de banco de dados
* O sistema não garante compatibilidade com versões antigas após atualização

---

### ❗ Política de versão

* O Remembrall é projetado para operar **sempre na versão mais recente**
* Usuários que optarem por permanecer em versões antigas:

  * podem enfrentar inconsistências
  * podem não conseguir atualizar corretamente
  * não terão garantia de suporte

👉 O projeto **não se responsabiliza por problemas decorrentes de uso de versões desatualizadas**

---

## 8. 📦 Distribuição via Gerenciadores de Pacote

O projeto será distribuído via:

* Homebrew
* Chocolatey
* Pacman (AUR)

---

### 🍺 Homebrew

#### Estratégia

* Criar tap customizado:

```
remembrall/homebrew-tap
```

#### Processo

1. CI gera release
2. CI atualiza fórmula
3. Commit automático no tap

#### Requisitos

* Binários via HTTPS
* SHA256 válido
* Licença definida

---

### 🍫 Chocolatey

#### Estratégia

* Criar pacote `.nuspec`

#### Processo

1. CI gera pacote
2. Publica no repositório Chocolatey

#### Requisitos

* URL estável
* Checksums obrigatórios
* Scripts seguros
* Aprovação inicial manual

---

### 🐧 Pacman (AUR)

#### Estratégia

* Publicação via AUR

#### Processo

1. CI atualiza `PKGBUILD`
2. Publica no AUR

#### Requisitos

* Source acessível
* Checksums válidos
* Script auditável

---

### 🎯 Objetivo

Permitir instalação simples:

```bash
brew install remembrall
choco install remembrall
yay -S remembrall
```

---

## 11. 📁 Convenção de Artefatos

### 📦 Padrão

```
<name>_<version>_<os>_<arch>.<ext>
```

---

### 📌 Exemplos

```
remembrall_1.2.0_linux_amd64.tar.gz
remembrall_1.2.0_windows_amd64.zip
remembrall_1.2.0_darwin_arm64.tar.gz
```

---

### 📌 Regras

* OS:

  * linux
  * windows
  * darwin

* Arquitetura:

  * amd64
  * arm64

---

### 📦 Conteúdo

* Binário
* README (opcional)
* Licença

---

## 12. 📜 Licença

Licença adotada:

MIT License

---

### 📌 Implicações

* Uso comercial permitido
* Modificação permitida
* Redistribuição permitida
* Obrigatoriedade de manter a licença

---

### 🎯 Objetivo

* Maximizar adoção
* Reduzir restrições legais
* Facilitar contribuições open source

---

# 🧩 Aplicação – Estrutura Inicial (CLI, Configuração e Updates)

Esta seção define os aspectos fundamentais da experiência do usuário no Remembrall, incluindo comportamento da CLI, configuração e estratégia de atualização. As definições aqui priorizam **simplicidade, previsibilidade e baixo custo operacional**.

---

## 🔄 Estratégia de Update

O Remembrall adota uma abordagem **opinativa e obrigatória** em relação a atualizações.

---

### 📌 Princípios

- O sistema sempre incentiva o uso da versão mais recente
- Não há controle do usuário sobre comportamento de atualização
- Não há configuração para update (auto update, canais, etc.)
- O sistema não suporta:
  - rollback
  - downgrade
  - escolha de versão

---

### 🔔 Notificação de Atualização

A verificação de atualização deve ocorrer **em todo comando executado pelo usuário**.

---

### 📌 Comportamento

1. O CLI realiza uma verificação silenciosa de nova versão
2. Caso exista uma versão mais recente:
   - O usuário é notificado imediatamente
3. O comando principal continua sendo executado normalmente

---

### 📌 Exemplo de saída

```bash
New version available: 1.3.0
Run: remembrall update apply

<ID>  <TEXT>       <DATE>
1     Buy milk     2026-03-22 18:00
````

---

### ⚙️ Requisitos

* A verificação deve ser:

  * rápida
  * não bloqueante ou com timeout curto
* Falhas na verificação:

  * não devem impactar o comando principal
  * devem ser silenciosas

---

### 📌 Atualização do Aplicativo

O processo de atualização deve ser realizado manualmente via comando:

```bash
remembrall update apply
```

---

### 📌 Fluxo de atualização

1. Verificar versão mais recente
2. Baixar binário correspondente ao sistema operacional
3. Validar integridade do binário
4. Substituir binário atual
5. Executar migrations de banco de dados
6. Finalizar atualização

---

## ⚙️ Configuração do Usuário

O Remembrall adota uma configuração **mínima e focada exclusivamente na experiência do usuário**.

---

### 📁 Localização

* Linux/macOS:

```bash
~/.config/remembrall/config.yaml
```

* Windows:

```bash
%APPDATA%\remembrall\config.yaml
```

---

### 📄 Estrutura

```yaml
date_format: "2006-01-02 15:04"
```

---

### 📌 Parâmetros suportados

#### 📅 `date_format`

* Define o formato de exibição de datas no CLI
* Não impacta armazenamento interno

---

### ❌ Configurações não suportadas

O sistema explicitamente **não permite configuração** para:

* timezone
* comportamento de update
* canais de distribuição

---

## 🌍 Timezone

---

### 📌 Regra

* O timezone é sempre obtido automaticamente do sistema operacional
* Não é configurável pelo usuário

---

### 📌 Comportamento

* Datas devem ser armazenadas preferencialmente em UTC
* Conversão para timezone local ocorre apenas na exibição

---

## 🧠 Filosofia da Aplicação

O Remembrall segue princípios claros:

---

### 1. 🧱 CLI opinativa

* Redução de opções desnecessárias
* Comportamento previsível
* Menor necessidade de suporte

---

### 2. 🔄 Atualização contínua

* Usuários são constantemente incentivados a atualizar
* Redução de fragmentação de versões
* Garantia de compatibilidade com migrations

---

### 3. ⚙️ Configuração mínima

* Apenas o essencial é configurável
* Evita complexidade desnecessária
* Foco na experiência do usuário

---

````markdown id="cli-setup-commands-001"
# 🧩 CLI – Comandos Iniciais e Setup

Esta seção define os comandos iniciais do Remembrall relacionados à inicialização do sistema, gerenciamento do daemon e atualização do aplicativo.

O foco é garantir uma experiência consistente, previsível e automatizada, com responsabilidade clara do usuário sobre o ambiente.

---

## 🖥️ Estrutura Geral da CLI

O Remembrall seguirá o padrão tradicional de aplicações CLI:

```bash
remembrall <command> [subcommand] [flags]
````

---

### 📌 Comando base

```bash
remembrall
```

* Exibe:

  * lista de comandos disponíveis
  * descrição breve de cada comando
  * instruções de uso

---

## ⚙️ Comando de Setup

```bash
remembrall setup
```

O comando `setup` é responsável por inicializar completamente o ambiente do Remembrall.

---

### 📌 Responsabilidades do setup

O setup deve:

1. Criar estrutura de diretórios do usuário
2. Criar ou configurar o banco de dados
3. Executar migrations iniciais
4. Criar arquivo de configuração
5. Configurar execução do daemon
6. Validar funcionamento geral do sistema

---

### 📁 Estrutura de diretórios

O setup deve garantir a existência do diretório:

* Linux/macOS:

```bash
~/.config/remembrall/
```

* Windows:

```bash
%APPDATA%\remembrall\
```

---

### 🗄️ Configuração do banco de dados

Durante o setup, o usuário deve informar:

* Caminho para banco SQLite local
  **ou**
* URL de conexão (ex: banco remoto)

---

### 📌 Comportamento

* Se o banco não existir:

  * deve ser criado automaticamente
* Deve ser validada a conexão antes de prosseguir
* Em caso de erro:

  * o setup deve ser interrompido imediatamente
  * erro deve ser exibido ao usuário

---

### 🔄 Migrations iniciais

Após validação do banco:

* Executar todas as migrations disponíveis
* Criar tabela `schema_migrations`
* Garantir estado consistente inicial

---

### ⚠️ Política de erro

* Em caso de falha:

  * o setup deve ser interrompido
  * o usuário deve corrigir o problema
  * o setup deve ser executado novamente do zero

---

### 📄 Arquivo de configuração

O setup deve gerar:

```bash
config.yaml
```

---

#### 📌 Conteúdo mínimo:

```yaml
database_url: "<connection_string>"
date_format: "2006-01-02 15:04"
```

---

### 🧠 PATH e execução global

O setup deve:

* orientar ou configurar o ambiente para permitir execução global do comando:

```bash
remembrall
```

* também deve permitir uso do alias:

```bash
rb
```

---

### 🔄 Configuração do daemon

Durante o setup, o usuário deve ser questionado:

```text
Do you want the daemon to start automatically on system startup? (y/n)
```

---

### 📌 Comportamento

* Se sim:

  * configurar inicialização automática conforme sistema operacional
* Se não:

  * daemon será iniciado manualmente

---

### ▶️ Inicialização do daemon

Ao final do setup:

* verificar se o daemon já está em execução
* se estiver:

  * parar o daemon
* iniciar o daemon novamente

---

## 🧠 Health Check Final

Ao final do setup, o sistema deve executar uma verificação completa.

---

### 📌 Deve validar:

* conexão com banco de dados
* migrations aplicadas corretamente
* daemon em execução

---

### 📌 Exemplo de saída

```text
Setup completed successfully

✔ Database connected
✔ Migrations applied
✔ Daemon running
```

---

## ⚙️ Comandos do Daemon

---

### ▶️ Iniciar daemon

```bash
remembrall daemon start
```

* Inicia o daemon manualmente
* Deve verificar se já está rodando
* Não deve iniciar múltiplas instâncias locais

---

### ⏹️ Parar daemon

```bash
remembrall daemon stop
```

* Interrompe execução do daemon

---

## 🔄 Comandos de Update

---

### 🔍 Verificar atualização

```bash
remembrall update check
```

* Consulta versão mais recente disponível
* Compara com versão local

---

### ⚡ Aplicar atualização

```bash
remembrall update apply
```

---

### 📌 Comportamento

* Baixar nova versão
* Substituir binário atual
* Executar migrations automaticamente

---

## ⚠️ Política de Concorrência

O Remembrall não implementa controle de concorrência entre múltiplas instâncias.

---

### 📌 Definição

* Múltiplas instâncias podem operar sobre o mesmo banco
* O sistema não garante exclusividade de execução do daemon
* O banco de dados é a única fonte de consistência

---

### 📌 Responsabilidade

* O usuário é responsável pelo ambiente
* O sistema não se responsabiliza por:

  * concorrência entre múltiplas máquinas
  * uso simultâneo do mesmo banco

---

### 📌 Requisito técnico

O daemon deve operar de forma **idempotente**, garantindo que ações não sejam executadas múltiplas vezes em cenários concorrentes.


