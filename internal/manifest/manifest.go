package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// TargetEnvironment describes where the app is deployed.
type TargetEnvironment string

const (
	EnvCloudNative TargetEnvironment = "cloud-native"
	EnvOnPremise   TargetEnvironment = "on-premise"
	EnvEdge        TargetEnvironment = "edge"
	EnvLocal       TargetEnvironment = "local"
)

// SystemTopology describes the architectural pattern.
type SystemTopology string

const (
	TopoMonolith    SystemTopology = "monolith"
	TopoMicro       SystemTopology = "microservices"
	TopoServerless  SystemTopology = "serverless"
	TopoEventDriven SystemTopology = "event-driven"
)

// ScalingStrategy describes how the system scales.
type ScalingStrategy string

const (
	ScaleHorizontal ScalingStrategy = "horizontal"
	ScaleVertical   ScalingStrategy = "vertical"
	ScaleBoth       ScalingStrategy = "both"
	ScaleNone       ScalingStrategy = "none"
)

// DatabaseType describes the primary database type.
type DatabaseType string

const (
	DBPostgres  DatabaseType = "PostgreSQL"
	DBMySQL     DatabaseType = "MySQL"
	DBMongo     DatabaseType = "MongoDB"
	DBDynamo    DatabaseType = "DynamoDB"
	DBRedis     DatabaseType = "Redis"
	DBSQLite    DatabaseType = "SQLite"
	DBOther     DatabaseType = "other"
)

// APIParadigm describes how data is exposed.
type APIParadigm string

const (
	APIREST    APIParadigm = "REST"
	APIGraphQL APIParadigm = "GraphQL"
	APIGrpc    APIParadigm = "gRPC"
	APITrpc    APIParadigm = "tRPC"
	APIMixed   APIParadigm = "mixed"
)

// --- Pillar structs ---

type ArchitecturePillar struct {
	TargetEnvironment TargetEnvironment `json:"target_environment"`
	CloudProvider     string            `json:"cloud_provider,omitempty"`
	Topology          SystemTopology    `json:"topology"`
	ScalingStrategy   ScalingStrategy   `json:"scaling_strategy"`
	ScalingNotes      string            `json:"scaling_notes,omitempty"`
}

type TechStackPillar struct {
	FrontendFramework  string `json:"frontend_framework"`
	FrontendVersion    string `json:"frontend_version"`
	StateManagement    string `json:"state_management"`
	StylingParadigm    string `json:"styling_paradigm"`
	BackendLanguage    string `json:"backend_language"`
	BackendFramework   string `json:"backend_framework"`
	RuntimeEnvironment string `json:"runtime_environment"`
	ThirdParty         string `json:"third_party_integrations"`
}

type DataArchPillar struct {
	DatabaseType     DatabaseType `json:"database_type"`
	SecondaryDB      string       `json:"secondary_db,omitempty"`
	CoreEntities     string       `json:"core_entities"`
	APIParadigm      APIParadigm  `json:"api_paradigm"`
	CachingStrategy  string       `json:"caching_strategy,omitempty"`
}

type FunctionalSpecPillar struct {
	UserRoles     string `json:"user_roles"`
	CoreJourneys  string `json:"core_journeys"`
	ErrorHandling string `json:"error_handling"`
}

type NFRPillar struct {
	Encryption      string `json:"encryption"`
	Sanitization    string `json:"input_sanitization"`
	RateLimiting    string `json:"rate_limiting"`
	LatencyTarget   string `json:"latency_target"`
	Compliance      string `json:"compliance"`
	Accessibility   string `json:"accessibility"`
}

type DevWorkflowPillar struct {
	ProjectStructure string `json:"project_structure"`
	TestFramework    string `json:"test_framework"`
	CoverageTarget   string `json:"coverage_target"`
	CIPlatform       string `json:"ci_platform"`
	Linting          string `json:"linting"`
	Formatting       string `json:"formatting"`
}

// Manifest is the root structure holding all six pillars.
type Manifest struct {
	CreatedAt    time.Time            `json:"created_at"`
	Architecture ArchitecturePillar   `json:"architecture"`
	TechStack    TechStackPillar      `json:"tech_stack"`
	DataArch     DataArchPillar       `json:"data_architecture"`
	Functional   FunctionalSpecPillar `json:"functional_specs"`
	NFR          NFRPillar            `json:"non_functional_requirements"`
	DevWorkflow  DevWorkflowPillar    `json:"dev_workflow"`
}

// Save writes the manifest to path as indented JSON.
func (m *Manifest) Save(path string) error {
	m.CreatedAt = time.Now()
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest to %s: %w", path, err)
	}
	return nil
}
