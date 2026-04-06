.PHONY: exec migrate

SRC = src/main.go

# Variables
MIGRATIONS_DIR = src/db/migrations
GEN_FILE = $(MIGRATIONS_DIR)/migrations.go

## exec: Roda o app passando qualquer argumento seguinte (ex: make exec install)
exec:
	@TT_VERSION=dev go run $(SRC) $(filter-out $@,$(MAKECMDGOALS))

## migrate: Gera uma nova migração (ex: make migrate v=v0.0.1)
migrate:
	@if [ -z "$(v)" ]; then \
		echo "Error: Version is required. usage: make migrate v=v0.0.1"; \
		exit 1; \
	fi
	@echo "--- Running Atlas Migrate Diff for version $(v) ---"
	atlas migrate diff $(v) --env local

	rm -f $(GEN_FILE)

	@echo "--- Generating Go migration wrapper with Name Parsing ---"
	@printf "package migrations\n\n" > $(GEN_FILE)
	@printf "import (\n\t\"embed\"\n\t\"io/fs\"\n\t\"strings\"\n)\n\n" >> $(GEN_FILE)
	@printf "type Migration struct {\n\tVersion string\n\tScript  string\n}\n\n" >> $(GEN_FILE)
	@printf "//go:embed *.sql\nvar files embed.FS\n\n" >> $(GEN_FILE)
	@printf "// GetAllMigrations returns parsed Migration structs sorted by filename\n" >> $(GEN_FILE)
	@printf "func GetAllMigrations() []Migration {\n" >> $(GEN_FILE)
	@printf "\tentries, err := fs.ReadDir(files, \".\")\n" >> $(GEN_FILE)
	@printf "\tif err != nil {\n\t\treturn nil\n\t}\n\n" >> $(GEN_FILE)
	@printf "\tvar m []Migration\n" >> $(GEN_FILE)
	@printf "\tfor _, entry := range entries {\n" >> $(GEN_FILE)
	@printf "\t\tif !entry.IsDir() {\n" >> $(GEN_FILE)
	@printf "\t\t\tcontent, _ := fs.ReadFile(files, entry.Name())\n" >> $(GEN_FILE)
	@printf "\t\t\tname := entry.Name()\n\n" >> $(GEN_FILE)
	@printf "\t\t\t// Remove .sql extension\n" >> $(GEN_FILE)
	@printf "\t\t\tif strings.HasSuffix(name, \".sql\") {\n" >> $(GEN_FILE)
	@printf "\t\t\t\tname = strings.TrimSuffix(name, \".sql\")\n" >> $(GEN_FILE)
	@printf "\t\t\t}\n\n" >> $(GEN_FILE)
	@printf "\t\t\t// Split by underscore and take the second part\n" >> $(GEN_FILE)
	@printf "\t\t\tparts := strings.Split(name, \"_\")\n" >> $(GEN_FILE)
	@printf "\t\t\tif len(parts) > 1 {\n" >> $(GEN_FILE)
	@printf "\t\t\t\tname = parts[1]\n" >> $(GEN_FILE)
	@printf "\t\t\t}\n\n" >> $(GEN_FILE)
	@printf "\t\t\tm = append(m, Migration{\n" >> $(GEN_FILE)
	@printf "\t\t\t\tVersion: name,\n" >> $(GEN_FILE)
	@printf "\t\t\t\tScript:  string(content),\n" >> $(GEN_FILE)
	@printf "\t\t\t})\n" >> $(GEN_FILE)
	@printf "\t\t}\n" >> $(GEN_FILE)
	@printf "\t}\n" >> $(GEN_FILE)
	@printf "\treturn m\n" >> $(GEN_FILE)
	@printf "}\n" >> $(GEN_FILE)
	@echo "Done! Generated: $(GEN_FILE)"