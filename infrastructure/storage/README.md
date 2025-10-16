# Git Repository Storage Interface

Cette interface fournit une abstraction pour le stockage de dépôts Git, permettant de supporter différents backends (système de fichiers local et S3).

## Architecture

L'interface `GitRepositoryStorage` définit les méthodes nécessaires pour gérer des dépôts Git avec différents backends de stockage :

### Interface principale

```go
type GitRepositoryStorage interface {
    GetStorer(repoPath string) (storer.Storer, error)
    CreateRepository(repoPath string) error
    RepositoryExists(repoPath string) bool
    DeleteRepository(repoPath string) error
    ListRepositories() ([]string, error)
    Configure() error
}
```

### Backends supportés

#### 1. Local Storage (Implémenté)
- **Package**: `pkg/storage/local`
- **Type**: `LocalStorage`
- **Description**: Stocke les dépôts Git sur le système de fichiers local

#### 2. S3 Storage (Structure créée)
- **Package**: `pkg/storage/s3`
- **Type**: `S3Storage`
- **Description**: Stockera les dépôts Git sur Amazon S3 (ou compatible)
- **Status**: Structure de base créée, implémentation à compléter

## Utilisation

### Configuration

La configuration se fait via le fichier `config.yaml` :

```yaml
storage:
  type: "local"  # ou "s3"
  local:
    path: "./repositories"
  s3:
    bucket: "my-git-repos"
    endpoint: "https://s3.amazonaws.com"
    region: "us-east-1"
    access_key: "your-access-key"
    secret_key: "your-secret-key"
```

### Initialisation

```go
// Créer une instance de stockage basée sur la configuration
gitStorage, err := storage.NewGitRepositoryStorage(logger)
if err != nil {
    log.Fatal(err)
}

// Configurer le stockage
if err := gitStorage.Configure(); err != nil {
    log.Fatal(err)
}
```

### Opérations de base

```go
// Créer un nouveau dépôt
err := gitStorage.CreateRepository("my-repo.git")

// Vérifier si un dépôt existe
exists := gitStorage.RepositoryExists("my-repo.git")

// Obtenir un storer go-git pour un dépôt
storer, err := gitStorage.GetStorer("my-repo.git")

// Lister tous les dépôts
repos, err := gitStorage.ListRepositories()

// Supprimer un dépôt
err := gitStorage.DeleteRepository("my-repo.git")
```

### Intégration avec go-git server

```go
// Créer un loader pour un dépôt spécifique
loader := storage.NewGitServerLoader(gitStorage, "my-repo.git")

// Créer un serveur de transport go-git
srv := server.NewServer(loader)
ep := &transport.Endpoint{Path: "/my-repo.git"}

// Utiliser le serveur pour les sessions upload-pack/receive-pack
session, err := srv.NewUploadPackSession(ep, nil)
```

## Exemple complet

Voir `examples/git_server.go` pour un exemple complet d'implémentation d'un serveur Git HTTP utilisant cette interface.

L'exemple montre :
- Configuration automatique du backend de stockage
- API REST pour la gestion des dépôts
- Endpoints Git Smart HTTP Protocol
- Intégration avec Fiber pour le serveur web

## Extension pour S3

Pour compléter l'implémentation S3, il faudra :

1. **Implémenter un storer S3 personnalisé** qui implémente `storer.Storer`
2. **Créer les méthodes de gestion des dépôts** pour S3 (création, existence, suppression)
3. **Gérer la structure des objets S3** pour représenter un dépôt Git bare
4. **Optimiser les performances** avec des mécanismes de cache si nécessaire

## Avantages

- **Abstraction complète** : Le code serveur ne dépend pas du backend de stockage
- **Flexibilité** : Changement de backend via configuration
- **Extensibilité** : Ajout facile de nouveaux backends
- **Compatibilité go-git** : Intégration native avec l'écosystème go-git
- **Interface unifiée** : Mêmes opérations pour tous les backends
