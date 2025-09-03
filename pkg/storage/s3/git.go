package s3

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/labbs/git-server-s3/internal/config"
	"github.com/rs/zerolog"
)

type S3Storage struct {
	Logger zerolog.Logger
	bucket string
	client *awss3.Client
}

func NewS3Storage(logger zerolog.Logger) *S3Storage {
	return &S3Storage{
		Logger: logger,
	}
}

func (s3s *S3Storage) Configure() error {
	s3s.Logger.Info().Msg("Configuring S3 storage")

	if config.Storage.S3.Bucket == "" {
		return errors.New("S3 bucket is not configured")
	}

	s3s.bucket = config.Storage.S3.Bucket

	// Initialize S3 client
	var s3Config S3Config
	s3Config.Logger = s3s.Logger
	err := s3Config.Configure()
	if err != nil {
		return fmt.Errorf("failed to configure S3 client: %w", err)
	}
	s3s.client = s3Config.Client

	// Test connection by listing objects (alternative to HeadBucket which requires fewer permissions)
	_, err = s3s.client.ListObjectsV2(context.TODO(), &awss3.ListObjectsV2Input{
		Bucket:  aws.String(s3s.bucket),
		MaxKeys: aws.Int32(1), // Only get 1 object to minimize overhead
	})
	if err != nil {
		s3s.Logger.Error().
			Err(err).
			Str("bucket", s3s.bucket).
			Str("endpoint", config.Storage.S3.Endpoint).
			Msg("ListObjects operation failed")
		return fmt.Errorf("failed to access S3 bucket %s: %w", s3s.bucket, err)
	}

	s3s.Logger.Info().Str("bucket", s3s.bucket).Msg("S3 storage configured successfully")
	return nil
}

func (s3s *S3Storage) GetStorer(repoPath string) (storer.Storer, error) {
	if !s3s.RepositoryExists(repoPath) {
		return nil, errors.New("repository does not exist")
	}

	return NewS3Storer(s3s.client, s3s.bucket, s3s.getRepoKey(repoPath), s3s.Logger), nil
}

func (s3s *S3Storage) CreateRepository(repoPath string) error {
	if s3s.RepositoryExists(repoPath) {
		return errors.New("repository already exists")
	}

	// Create a minimal bare repository structure in S3
	repoKey := s3s.getRepoKey(repoPath)

	// Create basic config
	configContent := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = true
`
	_, err := s3s.client.PutObject(context.TODO(), &awss3.PutObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(repoKey + "/config"),
		Body:   strings.NewReader(configContent),
	})
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	// Create empty objects and refs directories by creating marker files
	_, err = s3s.client.PutObject(context.TODO(), &awss3.PutObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(repoKey + "/objects/.gitkeep"),
		Body:   strings.NewReader(""),
	})
	if err != nil {
		return fmt.Errorf("failed to create objects directory: %w", err)
	}

	// Create initial commit and main branch
	if err := s3s.createInitialCommit(repoKey); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	s3s.Logger.Info().Str("repo", repoPath).Msg("Repository created in S3 with initial commit")
	return nil
}

// createInitialCommit creates an initial commit with README.md and main branch
func (s3s *S3Storage) createInitialCommit(repoKey string) error {
	// Create a storer for this repository
	storer := NewS3Storer(s3s.client, s3s.bucket, repoKey, s3s.Logger)

	// Create README.md content
	readmeContent := []byte(`# Repository
	
This is a new Git repository hosted on S3.

## Getting Started

Clone this repository:
` + "```bash" + `
git clone <repository-url>
` + "```" + `

Start adding your files and make your first commit!
`)

	// Create blob object for README.md
	readmeBlob := &plumbing.MemoryObject{}
	readmeBlob.SetType(plumbing.BlobObject)
	readmeBlob.SetSize(int64(len(readmeContent)))
	readmeBlob.Write(readmeContent)

	// Store the blob
	readmeHash, err := storer.SetEncodedObject(readmeBlob)
	if err != nil {
		return fmt.Errorf("failed to store README.md blob: %w", err)
	}

	// Create tree with README.md
	tree := &object.Tree{
		Entries: []object.TreeEntry{
			{
				Name: "README.md",
				Mode: 0o100644, // Regular file
				Hash: readmeHash,
			},
		},
	}

	// Encode the tree
	treeObj := &plumbing.MemoryObject{}
	if err := tree.Encode(treeObj); err != nil {
		return fmt.Errorf("failed to encode tree: %w", err)
	}

	// Store the tree object
	treeHash, err := storer.SetEncodedObject(treeObj)
	if err != nil {
		return fmt.Errorf("failed to store tree: %w", err)
	}

	// Create initial commit
	commit := &object.Commit{
		Author: object.Signature{
			Name:  "Git Server",
			Email: "git-server@example.com",
			When:  time.Now(),
		},
		Committer: object.Signature{
			Name:  "Git Server",
			Email: "git-server@example.com",
			When:  time.Now(),
		},
		Message:      "Initial commit\n\nCreated repository with README.md",
		TreeHash:     treeHash,
		ParentHashes: []plumbing.Hash{}, // No parents for initial commit
	}

	// Encode the commit
	commitObj := &plumbing.MemoryObject{}
	if err := commit.Encode(commitObj); err != nil {
		return fmt.Errorf("failed to encode commit: %w", err)
	}

	// Store the commit object
	commitHash, err := storer.SetEncodedObject(commitObj)
	if err != nil {
		return fmt.Errorf("failed to store commit: %w", err)
	}

	// Create main branch pointing to the commit
	mainRef := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/main"), commitHash)
	if err := storer.SetReference(mainRef); err != nil {
		return fmt.Errorf("failed to create main branch: %w", err)
	}

	// Create HEAD pointing to main branch (symbolic reference)
	// This ensures that clone will checkout main branch by default
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, "refs/heads/main")
	if err := storer.SetReference(headRef); err != nil {
		return fmt.Errorf("failed to create HEAD: %w", err)
	}

	s3s.Logger.Debug().
		Str("treeHash", treeHash.String()).
		Str("commitHash", commitHash.String()).
		Str("readmeHash", readmeHash.String()).
		Msg("Created initial commit with README.md")

	return nil
}

func (s3s *S3Storage) RepositoryExists(repoPath string) bool {
	repoKey := s3s.getRepoKey(repoPath)

	// Check if HEAD exists to determine if repository exists
	_, err := s3s.client.HeadObject(context.TODO(), &awss3.HeadObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(repoKey + "/HEAD"),
	})

	return err == nil
}

func (s3s *S3Storage) DeleteRepository(repoPath string) error {
	if !s3s.RepositoryExists(repoPath) {
		return errors.New("repository does not exist")
	}

	repoKey := s3s.getRepoKey(repoPath)

	// List all objects with the repository prefix
	paginator := awss3.NewListObjectsV2Paginator(s3s.client, &awss3.ListObjectsV2Input{
		Bucket: aws.String(s3s.bucket),
		Prefix: aws.String(repoKey + "/"),
	})

	// Delete all objects in batches
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to list repository objects: %w", err)
		}

		if len(page.Contents) == 0 {
			break
		}

		// Prepare objects for deletion
		var objects []types.ObjectIdentifier
		for _, obj := range page.Contents {
			objects = append(objects, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		// Delete objects
		_, err = s3s.client.DeleteObjects(context.TODO(), &awss3.DeleteObjectsInput{
			Bucket: aws.String(s3s.bucket),
			Delete: &types.Delete{
				Objects: objects,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete repository objects: %w", err)
		}
	}

	s3s.Logger.Info().Str("repo", repoPath).Msg("Repository deleted from S3")
	return nil
}

func (s3s *S3Storage) ListRepositories() ([]string, error) {
	var repos []string
	prefix := "repositories/"

	paginator := awss3.NewListObjectsV2Paginator(s3s.client, &awss3.ListObjectsV2Input{
		Bucket:    aws.String(s3s.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		// Look for repository directories (should end with .git/)
		for _, commonPrefix := range page.CommonPrefixes {
			repoDir := aws.ToString(commonPrefix.Prefix)
			if strings.HasSuffix(repoDir, ".git/") {
				// Extract repository name
				repoName := strings.TrimPrefix(repoDir, prefix)
				repoName = strings.TrimSuffix(repoName, "/")
				repos = append(repos, repoName)
			}
		}
	}

	return repos, nil
}

// getRepoKey normalizes the repository path and returns the S3 key prefix
func (s3s *S3Storage) getRepoKey(repoPath string) string {
	// Clean the repo path and ensure it ends with .git
	cleanPath := strings.Trim(repoPath, "/")
	if !strings.HasSuffix(cleanPath, ".git") {
		cleanPath += ".git"
	}

	// Prefix with repositories/
	return "repositories/" + cleanPath
}
