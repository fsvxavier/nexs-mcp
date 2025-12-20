# NEXS MCP Infrastructure Layer

**Version:** 1.0.0  
**Last Updated:** December 20, 2025  
**Status:** Production

---

## Table of Contents

- [Introduction](#introduction)
- [Infrastructure Layer Purpose](#infrastructure-layer-purpose)
- [Repository Implementations](#repository-implementations)
- [File Storage](#file-storage)
- [GitHub Integration](#github-integration)
- [OAuth Authentication](#oauth-authentication)
- [Cryptography](#cryptography)
- [Sync System](#sync-system)
- [Conflict Detection](#conflict-detection)
- [PR Tracking](#pr-tracking)
- [HTTP Clients](#http-clients)
- [Performance Optimizations](#performance-optimizations)
- [Error Handling](#error-handling)
- [Testing Infrastructure](#testing-infrastructure)
- [Best Practices](#best-practices)

---

## Introduction

The **Infrastructure Layer** implements external integrations and storage mechanisms. It's where the rubber meets the road - connecting domain logic to real-world systems like file storage, GitHub API, and encryption services.

### Infrastructure Layer Location

```
internal/infrastructure/
├── repository.go                  # In-memory repository
├── file_repository.go             # YAML file storage
├── enhanced_file_repository.go    # Cached file storage
├── element_data.go                # Element serialization
├── github_client.go               # GitHub API wrapper
├── github_oauth.go                # OAuth device flow
├── github_publisher.go            # Collection publishing
├── pr_tracker.go                  # Pull request tracking
├── crypto.go                      # AES-256-GCM encryption
├── sync_metadata.go               # Sync state tracking
├── sync_incremental.go            # Delta synchronization
├── sync_conflict_detector.go      # Conflict resolution
└── *_test.go                      # Unit tests (68.1% coverage)
```

### Dependencies

```
Infrastructure Layer
      │
      ├─→ Domain Layer (implements interfaces)
      │
      ├─→ External Libraries (GitHub API, crypto, yaml)
      │
      └─→ File System / Network / OS
```

---

## Infrastructure Layer Purpose

### What Infrastructure Does

✅ **Implement Domain Interfaces**
- ElementRepository implementations
- Storage abstractions

✅ **External Service Integration**
- GitHub API client
- OAuth authentication
- HTTP communication

✅ **Data Persistence**
- File system operations
- YAML/JSON serialization
- Caching strategies

✅ **Security**
- Token encryption (AES-256-GCM)
- Secure storage
- Key derivation

### What Infrastructure Does NOT Do

❌ **Business Logic**
- Validation rules
- Domain invariants
- Business decisions

❌ **Orchestration**
- Use case implementation
- Workflow coordination
- Result aggregation

---

## Repository Implementations

### In-Memory Repository

Fast, ephemeral storage for development and testing:

```go
type InMemoryElementRepository struct {
    mu       sync.RWMutex
    elements map[string]domain.Element
}

func NewInMemoryElementRepository() *InMemoryElementRepository {
    return &InMemoryElementRepository{
        elements: make(map[string]domain.Element),
    }
}

// Implements domain.ElementRepository
func (r *InMemoryElementRepository) Create(element domain.Element) error {
    if element == nil {
        return fmt.Errorf("element cannot be nil")
    }
    
    r.mu.Lock()
    defer r.mu.Unlock()
    
    id := element.GetID()
    if _, exists := r.elements[id]; exists {
        return fmt.Errorf("element with ID %s already exists", id)
    }
    
    r.elements[id] = element
    return nil
}

func (r *InMemoryElementRepository) GetByID(id string) (domain.Element, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    element, exists := r.elements[id]
    if !exists {
        return nil, domain.ErrElementNotFound
    }
    
    return element, nil
}

func (r *InMemoryElementRepository) List(filter domain.ElementFilter) ([]domain.Element, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    elements := make([]domain.Element, 0)
    
    for _, elem := range r.elements {
        // Apply type filter
        if filter.Type != nil && elem.GetType() != *filter.Type {
            continue
        }
        
        // Apply active status filter
        if filter.IsActive != nil && elem.IsActive() != *filter.IsActive {
            continue
        }
        
        // Apply tags filter
        if len(filter.Tags) > 0 {
            meta := elem.GetMetadata()
            if !hasAllTags(meta.Tags, filter.Tags) {
                continue
            }
        }
        
        elements = append(elements, elem)
    }
    
    // Apply pagination
    return paginate(elements, filter.Offset, filter.Limit), nil
}
```

**Characteristics:**
- **Speed**: ~0.1ms per operation
- **Capacity**: Limited by RAM (~2KB per element)
- **Persistence**: None (data lost on restart)
- **Concurrency**: Thread-safe with RWMutex

---

## File Storage

### File Repository

Persistent YAML-based storage with intelligent file organization:

```go
type FileElementRepository struct {
    mu      sync.RWMutex
    baseDir string
    cache   map[string]*StoredElement // In-memory cache
}

func NewFileElementRepository(baseDir string) (*FileElementRepository, error) {
    if baseDir == "" {
        baseDir = "data/elements"
    }
    
    // Create base directory
    if err := os.MkdirAll(baseDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create base directory: %w", err)
    }
    
    repo := &FileElementRepository{
        baseDir: baseDir,
        cache:   make(map[string]*StoredElement),
    }
    
    // Load existing elements into cache
    if err := repo.loadCache(); err != nil {
        return nil, fmt.Errorf("failed to load cache: %w", err)
    }
    
    return repo, nil
}
```

### File Organization

Elements are organized by type and date:

```
data/elements/
├── personas/
│   ├── 2025-12-20/
│   │   ├── persona-technical-expert-1703088000.yaml
│   │   └── persona-creative-writer-1703088100.yaml
│   └── 2025-12-21/
│       └── persona-analyst-1703174400.yaml
├── skills/
│   └── 2025-12-20/
│       ├── skill-code-review-1703088000.yaml
│       └── skill-data-analysis-1703088200.yaml
├── templates/
├── agents/
├── memories/
└── ensembles/
```

### File Path Generation

```go
// Structure: baseDir/type/YYYY-MM-DD/id.yaml
func (r *FileElementRepository) getFilePath(metadata domain.ElementMetadata) string {
    typeDir := string(metadata.Type)
    dateDir := metadata.CreatedAt.Format("2006-01-02")
    filename := fmt.Sprintf("%s.yaml", metadata.ID)
    return filepath.Join(r.baseDir, typeDir, dateDir, filename)
}

// Example: data/elements/persona/2025-12-20/persona-technical-expert-1703088000.yaml
```

### Serialization Format

Elements are stored as YAML with structured data:

```go
type StoredElement struct {
    Metadata domain.ElementMetadata `yaml:"metadata"`
    Data     map[string]interface{} `yaml:"data,omitempty"`
}

// Serialize element to YAML
func (r *FileElementRepository) serializeElement(element domain.Element) (*StoredElement, error) {
    stored := &StoredElement{
        Metadata: element.GetMetadata(),
        Data:     make(map[string]interface{}),
    }
    
    // Convert element to map using reflection
    switch e := element.(type) {
    case *domain.Persona:
        stored.Data = personaToMap(e)
    case *domain.Skill:
        stored.Data = skillToMap(e)
    case *domain.Template:
        stored.Data = templateToMap(e)
    // ... other types
    }
    
    return stored, nil
}
```

### Example YAML File

```yaml
# persona-technical-expert-1703088000.yaml
metadata:
  id: "persona-technical-expert-1703088000"
  type: "persona"
  name: "Technical Expert"
  description: "Expert in software architecture and clean code"
  version: "1.0.0"
  author: "team@company.com"
  tags: ["architecture", "expert"]
  is_active: true
  created_at: "2025-12-20T10:00:00Z"
  updated_at: "2025-12-20T10:00:00Z"

data:
  behavioral_traits:
    - name: "Analytical"
      description: "Systematic problem-solving"
      intensity: 9
  expertise_areas:
    - domain: "Software Architecture"
      level: "expert"
      keywords: ["clean architecture", "DDD"]
  response_style:
    tone: "Professional"
    formality: "neutral"
    verbosity: "balanced"
  system_prompt: "You are a Senior Software Architect..."
  privacy_level: "public"
  hot_swappable: true
```

### CRUD Operations

```go
// Create element
func (r *FileElementRepository) Create(element domain.Element) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    metadata := element.GetMetadata()
    
    // Check if already exists
    if _, exists := r.cache[metadata.ID]; exists {
        return fmt.Errorf("element with ID %s already exists", metadata.ID)
    }
    
    // Serialize to stored format
    stored, err := r.serializeElement(element)
    if err != nil {
        return fmt.Errorf("failed to serialize element: %w", err)
    }
    
    // Get file path
    filePath := r.getFilePath(metadata)
    
    // Create directory if needed
    dir := filepath.Dir(filePath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }
    
    // Marshal to YAML
    data, err := yaml.Marshal(stored)
    if err != nil {
        return fmt.Errorf("failed to marshal YAML: %w", err)
    }
    
    // Write to file
    if err := os.WriteFile(filePath, data, 0644); err != nil {
        return fmt.Errorf("failed to write file: %w", err)
    }
    
    // Update cache
    r.cache[metadata.ID] = stored
    
    return nil
}

// Read element
func (r *FileElementRepository) GetByID(id string) (domain.Element, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    // Check cache first
    stored, exists := r.cache[id]
    if !exists {
        return nil, domain.ErrElementNotFound
    }
    
    // Deserialize to domain entity
    element, err := r.deserializeElement(stored)
    if err != nil {
        return nil, fmt.Errorf("failed to deserialize element: %w", err)
    }
    
    return element, nil
}

// Update element
func (r *FileElementRepository) Update(element domain.Element) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    metadata := element.GetMetadata()
    
    // Check if exists
    if _, exists := r.cache[metadata.ID]; !exists {
        return domain.ErrElementNotFound
    }
    
    // Update timestamp
    metadata.UpdatedAt = time.Now()
    
    // Serialize and write (same as Create)
    return r.writeElement(element)
}

// Delete element
func (r *FileElementRepository) Delete(id string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    stored, exists := r.cache[id]
    if !exists {
        return domain.ErrElementNotFound
    }
    
    // Get file path and delete
    filePath := r.getFilePath(stored.Metadata)
    if err := os.Remove(filePath); err != nil {
        return fmt.Errorf("failed to delete file: %w", err)
    }
    
    // Remove from cache
    delete(r.cache, id)
    
    return nil
}
```

### Cache Strategy

The file repository maintains an in-memory cache for performance:

```go
func (r *FileElementRepository) loadCache() error {
    return filepath.Walk(r.baseDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        // Skip directories and non-YAML files
        if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
            return nil
        }
        
        // Read file
        data, err := os.ReadFile(path)
        if err != nil {
            return fmt.Errorf("failed to read file %s: %w", path, err)
        }
        
        // Unmarshal YAML
        var stored StoredElement
        if err := yaml.Unmarshal(data, &stored); err != nil {
            return fmt.Errorf("failed to unmarshal file %s: %w", path, err)
        }
        
        // Add to cache
        r.cache[stored.Metadata.ID] = &stored
        
        return nil
    })
}
```

**Cache Benefits:**
- **Fast Reads**: ~0.1ms from cache vs ~2ms from disk
- **Memory Efficient**: Only metadata cached, ~1KB per element
- **Auto-Reload**: Cache rebuilt on startup
- **Consistency**: Write-through cache pattern

---

## GitHub Integration

### GitHub Client

Wrapper around GitHub API with high-level operations:

```go
type GitHubClient struct {
    client      *github.Client
    oauthClient *GitHubOAuthClient
}

func NewGitHubClient(oauthClient *GitHubOAuthClient) *GitHubClient {
    return &GitHubClient{
        oauthClient: oauthClient,
    }
}

func (c *GitHubClient) ensureAuthenticated(ctx context.Context) (*github.Client, error) {
    if c.client != nil {
        return c.client, nil
    }
    
    // Get token from OAuth client
    token, err := c.oauthClient.GetToken(ctx)
    if err != nil {
        return nil, fmt.Errorf("not authenticated: %w", err)
    }
    
    // Create authenticated client
    ts := oauth2.StaticTokenSource(token)
    tc := oauth2.NewClient(ctx, ts)
    c.client = github.NewClient(tc)
    
    return c.client, nil
}
```

### Repository Operations

```go
// List repositories
func (c *GitHubClient) ListRepositories(ctx context.Context) ([]*Repository, error) {
    client, err := c.ensureAuthenticated(ctx)
    if err != nil {
        return nil, err
    }
    
    var allRepos []*Repository
    opt := &github.RepositoryListByAuthenticatedUserOptions{
        ListOptions: github.ListOptions{PerPage: 100},
    }
    
    for {
        repos, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)
        if err != nil {
            return nil, fmt.Errorf("failed to list repositories: %w", err)
        }
        
        for _, repo := range repos {
            allRepos = append(allRepos, &Repository{
                Owner:         repo.Owner.GetLogin(),
                Name:          repo.GetName(),
                FullName:      repo.GetFullName(),
                Description:   repo.GetDescription(),
                Private:       repo.GetPrivate(),
                URL:           repo.GetHTMLURL(),
                DefaultBranch: repo.GetDefaultBranch(),
            })
        }
        
        if resp.NextPage == 0 {
            break
        }
        opt.Page = resp.NextPage
    }
    
    return allRepos, nil
}

// Get user info
func (c *GitHubClient) GetUser(ctx context.Context) (string, error) {
    client, err := c.ensureAuthenticated(ctx)
    if err != nil {
        return "", err
    }
    
    user, _, err := client.Users.Get(ctx, "")
    if err != nil {
        return "", fmt.Errorf("failed to get user: %w", err)
    }
    
    return user.GetLogin(), nil
}
```

### File Operations

```go
// Get file content
func (c *GitHubClient) GetFile(
    ctx context.Context,
    owner, repo, path, branch string,
) (*FileContent, error) {
    client, err := c.ensureAuthenticated(ctx)
    if err != nil {
        return nil, err
    }
    
    opts := &github.RepositoryContentGetOptions{
        Ref: branch,
    }
    
    fileContent, _, _, err := client.Repositories.GetContents(
        ctx, owner, repo, path, opts,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to get file: %w", err)
    }
    
    content, err := fileContent.GetContent()
    if err != nil {
        return nil, fmt.Errorf("failed to decode content: %w", err)
    }
    
    return &FileContent{
        Path:    path,
        Content: content,
        SHA:     fileContent.GetSHA(),
        Size:    fileContent.GetSize(),
    }, nil
}

// Create file
func (c *GitHubClient) CreateFile(
    ctx context.Context,
    owner, repo, path, message, content, branch string,
) (*CommitInfo, error) {
    client, err := c.ensureAuthenticated(ctx)
    if err != nil {
        return nil, err
    }
    
    opts := &github.RepositoryContentFileOptions{
        Message: github.String(message),
        Content: []byte(content),
        Branch:  github.String(branch),
    }
    
    resp, _, err := client.Repositories.CreateFile(ctx, owner, repo, path, opts)
    if err != nil {
        return nil, fmt.Errorf("failed to create file: %w", err)
    }
    
    return &CommitInfo{
        SHA:     resp.GetSHA(),
        Message: message,
        Author:  resp.GetAuthor().GetName(),
        Date:    resp.GetAuthor().GetDate().String(),
    }, nil
}

// Update file
func (c *GitHubClient) UpdateFile(
    ctx context.Context,
    owner, repo, path, message, content, sha, branch string,
) (*CommitInfo, error) {
    client, err := c.ensureAuthenticated(ctx)
    if err != nil {
        return nil, err
    }
    
    opts := &github.RepositoryContentFileOptions{
        Message: github.String(message),
        Content: []byte(content),
        SHA:     github.String(sha),
        Branch:  github.String(branch),
    }
    
    resp, _, err := client.Repositories.UpdateFile(ctx, owner, repo, path, opts)
    if err != nil {
        return nil, fmt.Errorf("failed to update file: %w", err)
    }
    
    return &CommitInfo{
        SHA:     resp.GetSHA(),
        Message: message,
    }, nil
}

// Delete file
func (c *GitHubClient) DeleteFile(
    ctx context.Context,
    owner, repo, path, message, sha, branch string,
) error {
    client, err := c.ensureAuthenticated(ctx)
    if err != nil {
        return err
    }
    
    opts := &github.RepositoryContentFileOptions{
        Message: github.String(message),
        SHA:     github.String(sha),
        Branch:  github.String(branch),
    }
    
    _, _, err = client.Repositories.DeleteFile(ctx, owner, repo, path, opts)
    return err
}
```

---

## OAuth Authentication

### Device Flow

GitHub OAuth using device flow (no browser required):

```go
type GitHubOAuthClient struct {
    clientID     string
    config       *oauth2.Config
    tokenPath    string
    currentToken *oauth2.Token
    encryptor    *TokenEncryptor
}

func NewGitHubOAuthClient(tokenPath string) (*GitHubOAuthClient, error) {
    clientID := os.Getenv("GITHUB_CLIENT_ID")
    if clientID == "" {
        clientID = DefaultClientID
    }
    
    config := &oauth2.Config{
        ClientID: clientID,
        Endpoint: github.Endpoint,
        Scopes:   []string{"repo", "user"},
    }
    
    encryptor, err := NewTokenEncryptor()
    if err != nil {
        return nil, fmt.Errorf("failed to initialize encryptor: %w", err)
    }
    
    return &GitHubOAuthClient{
        clientID:  clientID,
        config:    config,
        tokenPath: tokenPath,
        encryptor: encryptor,
    }, nil
}
```

### Authentication Flow

```go
// Step 1: Start device flow
func (c *GitHubOAuthClient) StartDeviceFlow(ctx context.Context) (*DeviceFlowResponse, error) {
    deviceAuth, err := c.config.DeviceAuth(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to initiate device flow: %w", err)
    }
    
    return &DeviceFlowResponse{
        DeviceCode:      deviceAuth.DeviceCode,
        UserCode:        deviceAuth.UserCode,
        VerificationURI: deviceAuth.VerificationURI,
        ExpiresIn:       int(deviceAuth.Interval * 60),
        Interval:        int(deviceAuth.Interval),
    }, nil
}

// Step 2: Poll for token
func (c *GitHubOAuthClient) PollForToken(
    ctx context.Context,
    deviceCode string,
    interval int,
) (*oauth2.Token, error) {
    ticker := time.NewTicker(time.Duration(interval) * time.Second)
    defer ticker.Stop()
    
    timeout := time.After(10 * time.Minute)
    
    for {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-timeout:
            return nil, fmt.Errorf("device flow timeout")
        case <-ticker.C:
            token, err := c.config.DeviceAccessToken(ctx, &oauth2.DeviceAuthResponse{
                DeviceCode: deviceCode,
            })
            if err != nil {
                if err.Error() == "authorization_pending" {
                    continue // Keep polling
                }
                return nil, err
            }
            
            c.currentToken = token
            return token, nil
        }
    }
}
```

**Flow Diagram:**

```
1. Client requests device code
   ├─→ GitHub returns: device_code, user_code, verification_uri
   │
2. User visits verification_uri
   ├─→ Enters user_code
   ├─→ Authorizes application
   │
3. Client polls for token
   ├─→ Receives "authorization_pending" until user authorizes
   ├─→ Receives token once authorized
   │
4. Token encrypted and saved
   └─→ Token reused for subsequent requests
```

### Token Management

```go
// Save token (encrypted)
func (c *GitHubOAuthClient) SaveToken(token *oauth2.Token) error {
    // Create directory
    dir := filepath.Dir(c.tokenPath)
    if err := os.MkdirAll(dir, 0700); err != nil {
        return err
    }
    
    // Marshal to JSON
    data, err := json.Marshal(token)
    if err != nil {
        return err
    }
    
    // Encrypt
    encrypted, err := c.encryptor.Encrypt(data)
    if err != nil {
        return err
    }
    
    // Write with restricted permissions
    if err := os.WriteFile(c.tokenPath, []byte(encrypted), 0600); err != nil {
        return err
    }
    
    c.currentToken = token
    return nil
}

// Load token (decrypt)
func (c *GitHubOAuthClient) LoadToken() (*oauth2.Token, error) {
    encryptedData, err := os.ReadFile(c.tokenPath)
    if err != nil {
        return nil, err
    }
    
    // Decrypt
    decrypted, err := c.encryptor.Decrypt(string(encryptedData))
    if err != nil {
        return nil, fmt.Errorf("failed to decrypt token: %w", err)
    }
    
    // Unmarshal
    var token oauth2.Token
    if err := json.Unmarshal(decrypted, &token); err != nil {
        return nil, err
    }
    
    c.currentToken = &token
    return &token, nil
}

// Get current token
func (c *GitHubOAuthClient) GetToken(ctx context.Context) (*oauth2.Token, error) {
    // Try current token first
    if c.currentToken != nil && c.currentToken.Valid() {
        return c.currentToken, nil
    }
    
    // Try loading from disk
    token, err := c.LoadToken()
    if err == nil && token.Valid() {
        return token, nil
    }
    
    return nil, fmt.Errorf("no valid token found: user needs to authenticate")
}
```

---

## Cryptography

### Token Encryption

AES-256-GCM encryption for secure token storage:

```go
type TokenEncryptor struct {
    key []byte // 256-bit key
}

func NewTokenEncryptor() (*TokenEncryptor, error) {
    // Get machine identifier
    machineID, err := getMachineID()
    if err != nil {
        return nil, err
    }
    
    // Get or create salt
    homeDir, _ := os.UserHomeDir()
    saltPath := filepath.Join(homeDir, ".nexs-mcp", ".salt")
    salt, err := getOrCreateSalt(saltPath)
    if err != nil {
        return nil, err
    }
    
    // Derive key using PBKDF2
    key := pbkdf2.Key(
        []byte(machineID),
        salt,
        100000,  // 100,000 iterations
        32,      // 256-bit key
        sha256.New,
    )
    
    return &TokenEncryptor{key: key}, nil
}
```

### Encryption Process

```go
func (e *TokenEncryptor) Encrypt(plaintext []byte) (string, error) {
    // Create AES cipher
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    // Create GCM mode
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    // Generate random nonce
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    // Encrypt and authenticate
    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    
    // Encode to base64
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *TokenEncryptor) Decrypt(ciphertext string) ([]byte, error) {
    // Decode from base64
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return nil, err
    }
    
    // Create cipher and GCM
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    // Extract nonce
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return nil, fmt.Errorf("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    
    // Decrypt and verify
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }
    
    return plaintext, nil
}
```

**Security Features:**
- **AES-256-GCM**: Industry-standard encryption with authentication
- **PBKDF2 Key Derivation**: 100,000 iterations with SHA-256
- **Random Nonce**: Unique nonce per encryption
- **Authenticated Encryption**: Prevents tampering
- **Machine-Specific**: Key derived from machine ID

---

## Sync System

### Sync Metadata

Track sync state to enable incremental synchronization:

```go
type SyncMetadata struct {
    LastSync       time.Time                `json:"last_sync"`
    RemoteSHA      map[string]string        `json:"remote_sha"`      // file path → SHA
    LocalSHA       map[string]string        `json:"local_sha"`       // element ID → content SHA
    ConflictPolicy string                   `json:"conflict_policy"` // remote_wins, local_wins, manual
    SyncHistory    []SyncHistoryEntry       `json:"sync_history"`
}

type SyncHistoryEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Action    string    `json:"action"` // push, pull, conflict
    Files     []string  `json:"files"`
    Success   bool      `json:"success"`
    Error     string    `json:"error,omitempty"`
}
```

### Incremental Sync

Only sync changed files:

```go
func (s *IncrementalSync) DetectChanges(
    ctx context.Context,
    localElements []domain.Element,
    remoteFiles []RemoteFile,
) (*SyncPlan, error) {
    plan := &SyncPlan{
        ToPush:      make([]domain.Element, 0),
        ToPull:      make([]RemoteFile, 0),
        Conflicts:   make([]Conflict, 0),
        ToDelete:    make([]string, 0),
    }
    
    // Build maps for efficient lookup
    localMap := make(map[string]domain.Element)
    for _, elem := range localElements {
        localMap[elem.GetID()] = elem
    }
    
    remoteMap := make(map[string]RemoteFile)
    for _, file := range remoteFiles {
        remoteMap[file.Path] = file
    }
    
    // Check local elements
    for _, elem := range localElements {
        elemPath := s.getElementPath(elem)
        remoteSHA := s.metadata.RemoteSHA[elemPath]
        localSHA := s.computeElementSHA(elem)
        
        remoteFile, existsRemote := remoteMap[elemPath]
        
        if !existsRemote {
            // New local element - push to remote
            plan.ToPush = append(plan.ToPush, elem)
        } else if remoteSHA != remoteFile.SHA {
            // Remote changed
            if localSHA != s.metadata.LocalSHA[elem.GetID()] {
                // Both changed - conflict!
                plan.Conflicts = append(plan.Conflicts, Conflict{
                    ElementID:   elem.GetID(),
                    LocalElement:  elem,
                    RemoteFile:   remoteFile,
                    LocalSHA:    localSHA,
                    RemoteSHA:   remoteFile.SHA,
                })
            } else {
                // Only remote changed - pull
                plan.ToPull = append(plan.ToPull, remoteFile)
            }
        } else if localSHA != s.metadata.LocalSHA[elem.GetID()] {
            // Only local changed - push
            plan.ToPush = append(plan.ToPush, elem)
        }
        // else: No changes on either side
    }
    
    // Check for remote files not in local
    for _, remoteFile := range remoteFiles {
        if _, existsLocal := localMap[remoteFile.ElementID]; !existsLocal {
            // New remote file - pull
            plan.ToPull = append(plan.ToPull, remoteFile)
        }
    }
    
    return plan, nil
}
```

---

## Conflict Detection

### Conflict Types

```go
type ConflictType string

const (
    ConflictBothModified ConflictType = "both_modified"
    ConflictDeletedLocal ConflictType = "deleted_locally"
    ConflictDeletedRemote ConflictType = "deleted_remotely"
)

type Conflict struct {
    ElementID     string
    Type          ConflictType
    LocalElement  domain.Element
    RemoteFile    RemoteFile
    LocalSHA      string
    RemoteSHA     string
    LocalModTime  time.Time
    RemoteModTime time.Time
}
```

### Conflict Resolution

```go
func (cd *ConflictDetector) ResolveConflict(
    conflict Conflict,
    strategy ConflictStrategy,
) (*Resolution, error) {
    switch strategy {
    case StrategyRemoteWins:
        return &Resolution{
            Action: ActionPull,
            Element: nil, // Will pull from remote
        }, nil
        
    case StrategyLocalWins:
        return &Resolution{
            Action: ActionPush,
            Element: conflict.LocalElement,
        }, nil
        
    case StrategyNewest:
        if conflict.LocalModTime.After(conflict.RemoteModTime) {
            return &Resolution{Action: ActionPush, Element: conflict.LocalElement}, nil
        }
        return &Resolution{Action: ActionPull}, nil
        
    case StrategyManual:
        // Return conflict for manual resolution
        return nil, fmt.Errorf("manual resolution required")
        
    default:
        return nil, fmt.Errorf("unknown strategy: %s", strategy)
    }
}
```

---

## PR Tracking

### PR Submission Tracking

```go
type PRSubmission struct {
    ID              string                 `json:"id"`
    ElementID       string                 `json:"element_id"`
    ElementType     domain.ElementType     `json:"element_type"`
    ElementName     string                 `json:"element_name"`
    RepositoryOwner string                 `json:"repository_owner"`
    RepositoryName  string                 `json:"repository_name"`
    PRNumber        int                    `json:"pr_number"`
    PRURL           string                 `json:"pr_url"`
    Status          PRStatus               `json:"status"` // pending, merged, rejected
    SubmittedAt     time.Time              `json:"submitted_at"`
    UpdatedAt       time.Time              `json:"updated_at"`
    MergedAt        *time.Time             `json:"merged_at,omitempty"`
}

type PRTracker struct {
    historyFile string
}

func (t *PRTracker) RecordSubmission(submission PRSubmission) error {
    history, err := t.LoadHistory()
    if err != nil {
        return err
    }
    
    history.Submissions[submission.ID] = submission
    history.Stats.TotalSubmissions++
    
    switch submission.Status {
    case PRStatusPending:
        history.Stats.Pending++
    case PRStatusMerged:
        history.Stats.Merged++
    case PRStatusRejected:
        history.Stats.Rejected++
    }
    
    return t.SaveHistory(history)
}
```

---

## HTTP Clients

### Rate Limiting

```go
type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
    }
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
    return rl.limiter.Wait(ctx)
}
```

### Retry Logic

```go
func (c *GitHubClient) withRetry(
    ctx context.Context,
    operation func() error,
) error {
    maxRetries := 3
    backoff := 100 * time.Millisecond
    
    for i := 0; i < maxRetries; i++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        // Check if retryable
        if !isRetryable(err) {
            return err
        }
        
        // Wait with exponential backoff
        time.Sleep(backoff)
        backoff *= 2
    }
    
    return fmt.Errorf("operation failed after %d retries", maxRetries)
}
```

---

## Performance Optimizations

### Connection Pooling

```go
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}
```

### Batch Operations

```go
func (r *FileElementRepository) BatchCreate(elements []domain.Element) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    for _, elem := range elements {
        if err := r.createNoLock(elem); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Lazy Loading

```go
func (r *FileElementRepository) GetByID(id string) (domain.Element, error) {
    // Check cache first
    if cached, exists := r.cache[id]; exists {
        return r.deserialize(cached), nil
    }
    
    // Load from disk on cache miss
    return r.loadFromDisk(id)
}
```

---

## Error Handling

### Error Types

```go
var (
    ErrNotAuthenticated = errors.New("not authenticated")
    ErrNetworkError     = errors.New("network error")
    ErrRateLimited      = errors.New("rate limited")
    ErrNotFound         = errors.New("not found")
)

func (c *GitHubClient) GetFile(...) (*FileContent, error) {
    resp, err := c.client.Get(url)
    if err != nil {
        return nil, fmt.Errorf("%w: %v", ErrNetworkError, err)
    }
    
    if resp.StatusCode == 404 {
        return nil, ErrNotFound
    }
    
    // ...
}
```

---

## Testing Infrastructure

### Mock Repository

```go
type MockRepository struct {
    elements map[string]domain.Element
    mu       sync.Mutex
}

func (m *MockRepository) Create(elem domain.Element) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.elements[elem.GetID()] = elem
    return nil
}
```

---

## Best Practices

### 1. Implement Interfaces Completely

```go
// ✅ Good: Implements all methods
type FileElementRepository struct { ... }

func (r *FileElementRepository) Create(element domain.Element) error
func (r *FileElementRepository) GetByID(id string) (domain.Element, error)
func (r *FileElementRepository) Update(element domain.Element) error
func (r *FileElementRepository) Delete(id string) error
func (r *FileElementRepository) List(filter domain.ElementFilter) ([]domain.Element, error)
func (r *FileElementRepository) Exists(id string) (bool, error)
```

### 2. Handle Errors Gracefully

```go
// ✅ Good: Wrap errors with context
if err := os.MkdirAll(dir, 0755); err != nil {
    return fmt.Errorf("failed to create directory %s: %w", dir, err)
}
```

### 3. Use Context for Cancellation

```go
// ✅ Good: Respect context
func (c *GitHubClient) ListRepos(ctx context.Context) ([]*Repo, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    // ...
}
```

### 4. Thread Safety

```go
// ✅ Good: Use mutex for shared state
func (r *FileElementRepository) Create(elem domain.Element) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    // Safe concurrent access
}
```

---

## Conclusion

The Infrastructure Layer bridges the gap between pure business logic and real-world systems. Through well-designed implementations of repositories, GitHub integration, and cryptography, it provides a solid foundation for the application while maintaining clean separation from domain concerns.

**Key Takeaways:**
- Implement domain interfaces
- Handle external failures gracefully
- Optimize for common cases
- Test with mocks
- Secure sensitive data

---

**Document Version:** 1.0.0  
**Total Lines:** 1146  
**Last Updated:** December 20, 2025
