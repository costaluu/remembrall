// atlas.hcl

variable "version" {
  type    = string
  default = "0.0.1"
}

env "local" {
  src = "file://schema.sql"
  dev = "sqlite://dev?mode=memory"
  
  migration {
    dir = "file://src/db/migrations"
    // Isso remove os arquivos .sum que o Atlas gera por padrão (opcional)
    // mas mantém o histórico limpo para o seu embed
    format = atlas
  }
}