package state

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/corynth/corynth/pkg/logging"
	corynthtypes "github.com/corynth/corynth/pkg/types"
	"github.com/corynth/corynth/pkg/workflow"
)

// S3StateManager manages workflow state using AWS S3 as backend
type S3StateManager struct {
	client     *s3.Client
	bucket     string
	keyPrefix  string
	region     string
	encryption bool
	locking    bool
	logger     *logging.Logger
}

// S3Config contains configuration for S3 state backend
type S3Config struct {
	Bucket     string `json:"bucket"`
	KeyPrefix  string `json:"key_prefix"`
	Region     string `json:"region"`
	AccessKey  string `json:"access_key,omitempty"`
	SecretKey  string `json:"secret_key,omitempty"`
	Endpoint   string `json:"endpoint,omitempty"`
	Encryption bool   `json:"encryption"`
	Locking    bool   `json:"locking"`
}

// NewS3StateManager creates a new S3 state manager
func NewS3StateManager(backendConfig map[string]string) (*S3StateManager, error) {
	// Parse configuration
	s3Config := &S3Config{
		KeyPrefix:  "corynth/state/",
		Region:     "us-east-1",
		Encryption: true,
		Locking:    true,
	}

	if bucket, ok := backendConfig["bucket"]; ok && bucket != "" {
		s3Config.Bucket = bucket
	} else {
		return nil, fmt.Errorf("S3 backend requires 'bucket' configuration")
	}

	if prefix, ok := backendConfig["key_prefix"]; ok {
		s3Config.KeyPrefix = prefix
	}

	if region, ok := backendConfig["region"]; ok && region != "" {
		s3Config.Region = region
	}

	if accessKey, ok := backendConfig["access_key"]; ok {
		s3Config.AccessKey = accessKey
	}

	if secretKey, ok := backendConfig["secret_key"]; ok {
		s3Config.SecretKey = secretKey
	}

	if endpoint, ok := backendConfig["endpoint"]; ok {
		s3Config.Endpoint = endpoint
	}

	if enc, ok := backendConfig["encryption"]; ok {
		s3Config.Encryption = enc == "true"
	}

	if lock, ok := backendConfig["locking"]; ok {
		s3Config.Locking = lock == "true"
	}

	// Create AWS config
	ctx := context.Background()
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(s3Config.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Override with custom credentials if provided
	if s3Config.AccessKey != "" && s3Config.SecretKey != "" {
		awsCfg.Credentials = credentials.NewStaticCredentialsProvider(
			s3Config.AccessKey,
			s3Config.SecretKey,
			"",
		)
	}

	// Create S3 client with custom endpoint if provided
	var s3Client *s3.Client
	if s3Config.Endpoint != "" {
		s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(s3Config.Endpoint)
			o.UsePathStyle = true // Required for MinIO and other S3-compatible services
		})
	} else {
		s3Client = s3.NewFromConfig(awsCfg)
	}

	// Verify bucket exists and is accessible
	_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s3Config.Bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("bucket '%s' not accessible: %w", s3Config.Bucket, err)
	}

	return &S3StateManager{
		client:     s3Client,
		bucket:     s3Config.Bucket,
		keyPrefix:  s3Config.KeyPrefix,
		region:     s3Config.Region,
		encryption: s3Config.Encryption,
		locking:    s3Config.Locking,
		logger:     logging.NewDefaultLogger("s3-state"),
	}, nil
}

// SaveState saves workflow execution state to S3
func (s *S3StateManager) SaveState(state *workflow.ExecutionState) error {
	ctx := context.Background()

	// Serialize state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Generate S3 key
	key := s.stateKey(state.ID)

	// Prepare put object input
	putInput := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
		Metadata: map[string]string{
			"workflow-name":   state.WorkflowName,
			"execution-id":    state.ID,
			"status":          string(state.Status),
			"corynth-version": "1.0.0",
		},
	}

	// Add server-side encryption if enabled
	if s.encryption {
		putInput.ServerSideEncryption = types.ServerSideEncryptionAes256
	}

	// Add content type
	putInput.ContentType = aws.String("application/json")

	// Acquire lock if locking is enabled
	if s.locking {
		lockKey := s.lockKey(state.ID)
		acquired, err := s.acquireLock(ctx, lockKey)
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		if !acquired {
			return fmt.Errorf("failed to acquire lock for state %s", state.ID)
		}
		defer s.releaseLock(ctx, lockKey)
	}

	// Upload to S3
	_, err = s.client.PutObject(ctx, putInput)
	if err != nil {
		return fmt.Errorf("failed to upload state to S3: %w", err)
	}

	return nil
}

// LoadState loads workflow execution state from S3
func (s *S3StateManager) LoadState(stateID string) (*workflow.ExecutionState, error) {
	ctx := context.Background()

	// Generate S3 key
	key := s.stateKey(stateID)

	// Get object from S3
	getInput := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, getInput)
	if err != nil {
		// Check if object doesn't exist
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("state not found: %s", stateID)
		}
		return nil, fmt.Errorf("failed to get state from S3: %w", err)
	}
	defer result.Body.Close()

	// Read and parse the state
	var state workflow.ExecutionState
	if err := json.NewDecoder(result.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode state: %w", err)
	}

	return &state, nil
}

// ListStates returns a list of all available workflow states from S3
func (s *S3StateManager) ListStates() ([]workflow.ExecutionState, error) {
	ctx := context.Background()

	// List objects with state prefix
	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(s.keyPrefix + "executions/"),
	}

	var states []workflow.ExecutionState
	paginator := s3.NewListObjectsV2Paginator(s.client, listInput)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, obj := range page.Contents {
			// Skip non-state files
			if !strings.HasSuffix(*obj.Key, ".json") {
				continue
			}

			// Extract state ID from key
			keyParts := strings.Split(*obj.Key, "/")
			if len(keyParts) < 2 {
				continue
			}

			stateFile := keyParts[len(keyParts)-1]
			stateID := strings.TrimSuffix(stateFile, ".json")

			// Load the state
			state, err := s.LoadState(stateID)
			if err != nil {
				// Skip states that can't be loaded
				continue
			}

			states = append(states, *state)
		}
	}

	return states, nil
}

// FindStatesByWorkflow finds all execution states for a specific workflow
func (s *S3StateManager) FindStatesByWorkflow(workflowName string) ([]workflow.ExecutionState, error) {
	states, err := s.ListStates()
	if err != nil {
		return nil, err
	}

	var matchingStates []workflow.ExecutionState
	for _, state := range states {
		if state.WorkflowName == workflowName {
			matchingStates = append(matchingStates, state)
		}
	}

	return matchingStates, nil
}

// GetLatestState returns the most recent execution state for a workflow
func (s *S3StateManager) GetLatestState(workflowName string) (*workflow.ExecutionState, error) {
	states, err := s.FindStatesByWorkflow(workflowName)
	if err != nil {
		return nil, err
	}

	if len(states) == 0 {
		return nil, fmt.Errorf("no states found for workflow: %s", workflowName)
	}

	// Find the most recent state
	var latest *workflow.ExecutionState
	for i := range states {
		if latest == nil || states[i].StartTime.After(latest.StartTime) {
			latest = &states[i]
		}
	}

	return latest, nil
}

// SaveWorkflowOutput saves workflow outputs for use by other workflows
func (s *S3StateManager) SaveWorkflowOutput(workflowName string, outputs map[string]interface{}) error {
	ctx := context.Background()

	outputData := corynthtypes.WorkflowOutput{
		WorkflowName: workflowName,
		Outputs:      outputs,
		Timestamp:    time.Now(),
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal outputs: %w", err)
	}

	// Generate S3 key
	key := s.outputKey(workflowName)

	// Upload to S3
	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
		Metadata: map[string]string{
			"workflow-name":   workflowName,
			"corynth-version": "1.0.0",
			"type":            "workflow-output",
		},
	}

	if s.encryption {
		putInput.ServerSideEncryption = types.ServerSideEncryptionAes256
	}

	_, err = s.client.PutObject(ctx, putInput)
	if err != nil {
		return fmt.Errorf("failed to upload outputs to S3: %w", err)
	}

	return nil
}

// LoadWorkflowOutput loads outputs from a previous workflow execution
func (s *S3StateManager) LoadWorkflowOutput(workflowName string) (*corynthtypes.WorkflowOutput, error) {
	ctx := context.Background()

	// Generate S3 key
	key := s.outputKey(workflowName)

	// Get object from S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("no outputs found for workflow: %s", workflowName)
		}
		return nil, fmt.Errorf("failed to get outputs from S3: %w", err)
	}
	defer result.Body.Close()

	var output corynthtypes.WorkflowOutput
	if err := json.NewDecoder(result.Body).Decode(&output); err != nil {
		return nil, fmt.Errorf("failed to decode outputs: %w", err)
	}

	return &output, nil
}

// CleanupOldStates removes state files older than the specified duration
func (s *S3StateManager) CleanupOldStates(maxAge time.Duration) error {
	ctx := context.Background()
	cutoff := time.Now().Add(-maxAge)

	// List all state objects
	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(s.keyPrefix),
	}

	paginator := s3.NewListObjectsV2Paginator(s.client, listInput)
	var objectsToDelete []types.ObjectIdentifier

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.LastModified.Before(cutoff) {
				objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
					Key: obj.Key,
				})
			}
		}
	}

	// Delete objects in batches
	if len(objectsToDelete) > 0 {
		// S3 allows up to 1000 objects per delete request
		batchSize := 1000
		for i := 0; i < len(objectsToDelete); i += batchSize {
			end := i + batchSize
			if end > len(objectsToDelete) {
				end = len(objectsToDelete)
			}

			batch := objectsToDelete[i:end]
			_, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(s.bucket),
				Delete: &types.Delete{
					Objects: batch,
				},
			})
			if err != nil {
				s.logger.WarnError(err, "failed to delete batch of old state objects")
			}
		}
	}

	return nil
}

// Helper functions

func (s *S3StateManager) stateKey(stateID string) string {
	return filepath.Join(s.keyPrefix, "executions", stateID+".json")
}

func (s *S3StateManager) outputKey(workflowName string) string {
	return filepath.Join(s.keyPrefix, "outputs", workflowName+".json")
}

func (s *S3StateManager) lockKey(stateID string) string {
	return filepath.Join(s.keyPrefix, "locks", stateID+".lock")
}

// Simple locking implementation using S3 object creation
func (s *S3StateManager) acquireLock(ctx context.Context, lockKey string) (bool, error) {
	// Try to create lock object
	lockData := map[string]interface{}{
		"acquired_at": time.Now().Format(time.RFC3339),
		"hostname":    getHostname(),
		"process_id":  os.Getpid(),
	}

	data, _ := json.Marshal(lockData)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(lockKey),
		Body:   bytes.NewReader(data),
		Metadata: map[string]string{
			"type": "lock",
		},
	})

	if err != nil {
		// Check if object already exists (lock held by someone else)
		if strings.Contains(err.Error(), "PreconditionFailed") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *S3StateManager) releaseLock(ctx context.Context, lockKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(lockKey),
	})
	return err
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}