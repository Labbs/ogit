# Makefile pour Git Server S3

.PHONY: help test test-unit test-integration test-coverage build run clean lint fmt vet deps

# Variables
BINARY_NAME=git-server-s3
BUILD_DIR=tmp
CONFIG_FILE=config.yaml

# Help
help: ## Afficher cette aide
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Installation des dépendances
deps: ## Installer les dépendances
	go mod download
	go mod tidy

# Tests unitaires
test-unit: ## Exécuter les tests unitaires
	go test -v -race -timeout 30s ./pkg/... ./internal/...

# Tests avec coverage
test-coverage: ## Exécuter les tests avec coverage
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./pkg/... ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Tests d'intégration (nécessite S3_TEST_*)
test-integration: ## Exécuter les tests d'intégration (nécessite variables S3_TEST_*)
	@if [ -z "$(S3_TEST_BUCKET)" ]; then \
		echo "❌ Variables S3_TEST_* non définies. Exemple:"; \
		echo "export S3_TEST_BUCKET=your-test-bucket"; \
		echo "export S3_TEST_REGION=us-east-1"; \
		echo "export S3_TEST_ENDPOINT=https://s3.us-east-1.amazonaws.com"; \
		echo "make test-integration"; \
		exit 1; \
	fi
	go test -v -tags=integration -timeout 5m ./...

# Tous les tests
test: test-unit ## Exécuter tous les tests (unitaires seulement par défaut)

# Tests avec plus de verbosité
test-verbose: ## Exécuter les tests avec plus de détails
	go test -v -race -coverprofile=coverage.out ./pkg/... ./internal/... -args -test.v

# Formatage du code
fmt: ## Formater le code
	go fmt ./...

# Vérifications statiques
vet: ## Vérifier le code avec go vet
	go vet ./...

# Linting (nécessite golangci-lint)
lint: ## Linter le code (nécessite golangci-lint)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint non installé. Installation:"; \
		echo "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Build
build: ## Compiler le serveur
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go

# Build avec informations de version
build-release: ## Compiler pour la release
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-X main.version=$(shell git describe --tags --always --dirty)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go

# Exécuter le serveur
run: build ## Compiler et exécuter le serveur
	./$(BUILD_DIR)/$(BINARY_NAME) server --config $(CONFIG_FILE)

# Exécuter en mode développement avec rechargement automatique (nécessite air)
dev: ## Exécuter en mode développement (nécessite air)
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "⚠️  air non installé. Installation:"; \
		echo "go install github.com/cosmtrek/air@latest"; \
		echo "Ou utilisez: make run"; \
	fi

# Nettoyage
clean: ## Nettoyer les fichiers générés
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean -cache -testcache

# Tests de performance
bench: ## Exécuter les benchmarks
	go test -bench=. -benchmem -run=^$$ ./pkg/... ./internal/...

# Tests de race conditions
race: ## Exécuter les tests avec détection de race conditions
	go test -race ./pkg/... ./internal/...

# Vérification de sécurité (nécessite gosec)
security: ## Vérifier la sécurité du code (nécessite gosec)
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec non installé. Installation:"; \
		echo "go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Mise à jour des dépendances
update-deps: ## Mettre à jour les dépendances
	go get -u ./...
	go mod tidy

# Vérification complète (CI)
ci: fmt vet test-unit ## Vérifications pour CI/CD

# Tests de mémoire avec Valgrind (Linux seulement)
memcheck: ## Tests de mémoire (Linux seulement)
	@if command -v valgrind >/dev/null 2>&1; then \
		go test -c ./pkg/storage/s3/ && valgrind --leak-check=full ./s3.test; \
	else \
		echo "⚠️  valgrind non disponible (Linux seulement)"; \
	fi

# Installation des outils de développement
install-tools: ## Installer les outils de développement
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Exemple de configuration de test
setup-test-env: ## Afficher un exemple de configuration pour les tests
	@echo "Configuration pour les tests d'intégration:"
	@echo ""
	@echo "# MinIO local (pour tests)"
	@echo "export S3_TEST_BUCKET=git-server-test"
	@echo "export S3_TEST_REGION=us-east-1"
	@echo "export S3_TEST_ENDPOINT=http://localhost:9000"
	@echo "export AWS_ACCESS_KEY_ID=minio"
	@echo "export AWS_SECRET_ACCESS_KEY=minio123"
	@echo ""
	@echo "# AWS S3 réel (pour tests en cloud)"
	@echo "export S3_TEST_BUCKET=your-test-bucket"
	@echo "export S3_TEST_REGION=us-east-1"
	@echo "export S3_TEST_ENDPOINT=https://s3.us-east-1.amazonaws.com"
	@echo "# + vos credentials AWS via AWS CLI ou variables d'environnement"

# Tests avec différents niveaux de log
test-debug: ## Exécuter les tests avec logs debug
	ZEROLOG_LEVEL=debug go test -v ./pkg/... ./internal/...

# Tests sur des fichiers spécifiques
test-s3: ## Tester seulement le package S3
	go test -v ./pkg/storage/s3/...

test-api: ## Tester seulement l'API
	go test -v ./internal/api/...

# Stats des tests
test-stats: ## Statistiques des tests
	@echo "📊 Statistiques des tests:"
	@find . -name "*_test.go" -not -path "./vendor/*" | wc -l | xargs echo "Fichiers de test:"
	@grep -r "func Test" --include="*_test.go" . | wc -l | xargs echo "Fonctions de test:"
	@grep -r "func Benchmark" --include="*_test.go" . | wc -l | xargs echo "Benchmarks:"
