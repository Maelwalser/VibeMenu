package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ── Enum types ────────────────────────────────────────────────────────────────

type ArchPattern string

const (
	ArchMonolith        ArchPattern = "monolith"
	ArchModularMonolith ArchPattern = "modular-monolith"
	ArchMicroservices   ArchPattern = "microservices"
	ArchEventDriven     ArchPattern = "event-driven"
)

type CommProtocol string

const (
	ProtoREST       CommProtocol = "REST"
	ProtoGraphQL    CommProtocol = "GraphQL"
	ProtoGRPC       CommProtocol = "gRPC"
	ProtoWebSockets CommProtocol = "WebSockets"
	ProtoMixed      CommProtocol = "mixed"
)

type SerializationFmt string

const (
	SerialJSON        SerializationFmt = "JSON"
	SerialProtobuf    SerializationFmt = "Protobuf"
	SerialMessagePack SerializationFmt = "MessagePack"
	SerialMixed       SerializationFmt = "mixed"
)

type ComputeEnv string

const (
	ComputeServerless     ComputeEnv = "serverless"
	ComputeContainerized  ComputeEnv = "containerized"
	ComputeBareMetalVM    ComputeEnv = "bare-metal/VM"
)

type DatabaseType string

const (
	DBPostgres DatabaseType = "PostgreSQL"
	DBMySQL    DatabaseType = "MySQL"
	DBMongo    DatabaseType = "MongoDB"
	DBDynamo   DatabaseType = "DynamoDB"
	DBSQLite   DatabaseType = "SQLite"
	DBOther    DatabaseType = "other"
)

type CacheStore string

const (
	CacheRedis     CacheStore = "Redis"
	CacheMemcached CacheStore = "Memcached"
	CacheNone      CacheStore = "none"
)

type RenderingMode string

const (
	RenderSPA RenderingMode = "SPA"
	RenderSSR RenderingMode = "SSR"
	RenderSSG RenderingMode = "SSG"
	RenderISR RenderingMode = "ISR"
)


type E2EFramework string

const (
	E2EPlaywright E2EFramework = "Playwright"
	E2ECypress    E2EFramework = "Cypress"
	E2ENone       E2EFramework = "none"
)

type CIPlatform string

const (
	CIGitHubActions CIPlatform = "GitHub Actions"
	CIGitLabCI      CIPlatform = "GitLab CI"
	CICircleCI      CIPlatform = "CircleCI"
	CIJenkins       CIPlatform = "Jenkins"
	CINone          CIPlatform = "none"
)

type SecretsBackend string

const (
	SecretsVault    SecretsBackend = "HashiCorp Vault"
	SecretsAWS      SecretsBackend = "AWS Secrets Manager"
	SecretsGCP      SecretsBackend = "GCP Secret Manager"
	SecretsEnvFiles SecretsBackend = "env files"
	SecretsNone     SecretsBackend = "none"
)

type LogSolution string

const (
	LogELK        LogSolution = "ELK Stack"
	LogDatadog    LogSolution = "Datadog"
	LogSplunk     LogSolution = "Splunk"
	LogCloudWatch LogSolution = "CloudWatch"
	LogOther      LogSolution = "other"
)

// ── Database source definitions ───────────────────────────────────────────────

// DBSourceDef describes a named database or cache source used in the project.
type DBSourceDef struct {
	Alias     string       `json:"alias"`              // e.g. "primary", "cache", "analytics"
	Type      DatabaseType `json:"type"`
	Version   string       `json:"version,omitempty"`
	Namespace string       `json:"namespace,omitempty"` // schema / keyspace / database name
	IsCache   bool         `json:"is_cache"`
	Notes     string       `json:"notes,omitempty"`
}

// ── Column / Entity definitions ───────────────────────────────────────────────

// ColumnType enumerates the SQL/schema data types for a column.
type ColumnType string

const (
	ColTypeText        ColumnType = "text"
	ColTypeVarchar     ColumnType = "varchar"
	ColTypeChar        ColumnType = "char"
	ColTypeInt         ColumnType = "int"
	ColTypeBigInt      ColumnType = "bigint"
	ColTypeSmallInt    ColumnType = "smallint"
	ColTypeSerial      ColumnType = "serial"
	ColTypeBigSerial   ColumnType = "bigserial"
	ColTypeBoolean     ColumnType = "boolean"
	ColTypeFloat       ColumnType = "float"
	ColTypeDouble      ColumnType = "double"
	ColTypeDecimal     ColumnType = "decimal"
	ColTypeJSON        ColumnType = "json"
	ColTypeJSONB       ColumnType = "jsonb"
	ColTypeUUID        ColumnType = "uuid"
	ColTypeTimestamp   ColumnType = "timestamp"
	ColTypeTimestampTZ ColumnType = "timestamptz"
	ColTypeDate        ColumnType = "date"
	ColTypeTime        ColumnType = "time"
	ColTypeBytea       ColumnType = "bytea"
	ColTypeEnum        ColumnType = "enum"
	ColTypeArray       ColumnType = "array"
	ColTypeOther       ColumnType = "other"
)

// CascadeAction defines the referential action on a foreign key.
type CascadeAction string

const (
	CascadeNoAction   CascadeAction = "NO ACTION"
	CascadeRestrict   CascadeAction = "RESTRICT"
	CascadeCascade    CascadeAction = "CASCADE"
	CascadeSetNull    CascadeAction = "SET NULL"
	CascadeSetDefault CascadeAction = "SET DEFAULT"
)

// IndexType enumerates supported index algorithms.
type IndexType string

const (
	IndexBTree IndexType = "btree"
	IndexHash  IndexType = "hash"
	IndexGIN   IndexType = "gin"
	IndexGIST  IndexType = "gist"
	IndexBRIN  IndexType = "brin"
)

// ForeignKey describes a column-level foreign key reference and its referential actions.
type ForeignKey struct {
	RefEntity string        `json:"ref_entity"`
	RefColumn string        `json:"ref_column"`
	OnDelete  CascadeAction `json:"on_delete"`
	OnUpdate  CascadeAction `json:"on_update"`
}

// ColumnDef fully specifies a single column within an entity.
type ColumnDef struct {
	Name       string      `json:"name"`
	Type       ColumnType  `json:"type"`
	Length     string      `json:"length,omitempty"`     // e.g. "255" for varchar(255)
	Nullable   bool        `json:"nullable"`
	PrimaryKey bool        `json:"primary_key"`
	Unique     bool        `json:"unique"`
	Default    string      `json:"default,omitempty"`
	Check      string      `json:"check,omitempty"`      // SQL CHECK expression
	ForeignKey *ForeignKey `json:"foreign_key,omitempty"`
	Index      bool        `json:"index"`
	IndexType  IndexType   `json:"index_type,omitempty"`
	Notes      string      `json:"notes,omitempty"`
}

// UniqueConstraint represents a composite unique constraint across multiple columns.
type UniqueConstraint struct {
	Name    string   `json:"name,omitempty"`
	Columns []string `json:"columns"`
}

// EntityDef defines a domain entity (table/collection) and all its columns.
type EntityDef struct {
	Name        string `json:"name"`
	Database    string `json:"database,omitempty"`    // alias ref to DBSourceDef
	Description string `json:"description,omitempty"`

	// Caching
	Cached     bool   `json:"cached"`
	CacheStore string `json:"cache_store,omitempty"` // alias of a cache DBSourceDef
	CacheTTL   string `json:"cache_ttl,omitempty"`   // e.g. "5m", "1h", "24h"

	Columns           []ColumnDef        `json:"columns"`
	UniqueConstraints []UniqueConstraint `json:"unique_constraints,omitempty"`
	Notes             string             `json:"notes,omitempty"`
}

// ── Phase 1: Universal Global Constants ──────────────────────────────────────

// DomainPillar captures entity schemas, RBAC, and compliance boundaries.
type DomainPillar struct {
	Entities   []EntityDef `json:"entities,omitempty"`
	RBACMatrix string      `json:"rbac_matrix"`
	Compliance string      `json:"compliance"` // GDPR, HIPAA, PCI-DSS, none
}

// TopologyPillar defines the structural model and inter-domain contracts.
type TopologyPillar struct {
	ArchPattern     ArchPattern      `json:"arch_pattern"`
	CommProtocol    CommProtocol     `json:"comm_protocol"`
	Serialization   SerializationFmt `json:"serialization"`
	DomainNotes     string           `json:"domain_notes,omitempty"`
}

// GlobalNFRPillar holds SLOs and disaster recovery parameters.
type GlobalNFRPillar struct {
	UptimeSLO      string `json:"uptime_slo"`       // e.g. "99.9%"
	ConcurrentConn string `json:"concurrent_conn"`  // e.g. "5000"
	RTO            string `json:"rto"`              // Recovery Time Objective
	RPO            string `json:"rpo"`              // Recovery Point Objective
	NFRNotes       string `json:"nfr_notes,omitempty"`
}

// ── Phase 2: Domain-Specific Execution Paths ─────────────────────────────────

// ServiceDef represents one backend module or microservice.
type ServiceDef struct {
	Name           string `json:"name"`
	Responsibility string `json:"responsibility"`
	Language       string `json:"language"`
	Framework      string `json:"framework"`
}

// BackendPillar covers compute environment, architecture pattern, and service definitions.
type BackendPillar struct {
	ArchPattern   ArchPattern  `json:"arch_pattern"`
	ComputeEnv    ComputeEnv   `json:"compute_env"`
	CloudProvider string       `json:"cloud_provider,omitempty"`
	// Monolith: single app language and framework.
	Language  string `json:"language,omitempty"`
	Framework string `json:"framework,omitempty"`
	// Microservices / Modular-monolith: per-service definitions.
	Services []ServiceDef `json:"services,omitempty"`
}

// FrontendPillar covers web rendering, framework, state management, styling, and browser support.
type FrontendPillar struct {
	Rendering     RenderingMode `json:"rendering"`
	Framework     string        `json:"framework"`      // e.g. "React 18", "Next.js 14"
	ServerState   string        `json:"server_state"`   // e.g. "React Query", "Apollo"
	ClientState   string        `json:"client_state"`   // e.g. "Zustand", "Redux"
	Styling       string        `json:"styling"`        // Tailwind, CSS-in-JS, SASS
	BrowserMatrix string        `json:"browser_matrix"` // e.g. "Chromium>100, Safari>15"
}


// ── Phase 3: Lifecycle Operations & Tooling ───────────────────────────────────

// TestingPillar defines coverage targets per test taxonomy.
type TestingPillar struct {
	UnitCoverage    string       `json:"unit_coverage"`    // e.g. "80%"
	IntegCoverage   string       `json:"integ_coverage"`
	E2EFramework    E2EFramework `json:"e2e_framework"`
	E2ECoverage     string       `json:"e2e_coverage"`
	TestingStrategy string       `json:"testing_strategy,omitempty"` // additional notes
}

// CICDPillar defines pipeline gates, environment strategy, and secrets management.
type CICDPillar struct {
	CIPlatform    CIPlatform     `json:"ci_platform"`
	PipelineGates string         `json:"pipeline_gates"` // blocking checks: lint, tests, vuln scan
	EnvStrategy   string         `json:"env_strategy"`   // dev/staging/prod definitions
	SecretsMgmt   SecretsBackend `json:"secrets_mgmt"`
}

// TelemetryPillar defines logging, metrics, tracing, and alerting strategy.
type TelemetryPillar struct {
	LogSolution LogSolution `json:"log_solution"`
	LogFormat   string      `json:"log_format"`  // JSON structured, plaintext
	Metrics     string      `json:"metrics"`     // Prometheus, Datadog, CloudWatch, none
	Tracing     string      `json:"tracing"`     // Jaeger, Zipkin, OpenTelemetry, none
	Alerting    string      `json:"alerting,omitempty"`
}

// ── Root manifest ─────────────────────────────────────────────────────────────

// Manifest is the root document holding all three phases.
type Manifest struct {
	CreatedAt time.Time `json:"created_at"`

	// Named database / cache sources and entity definitions
	Databases []DBSourceDef `json:"databases,omitempty"`
	Entities  []EntityDef   `json:"entities,omitempty"`

	// Phase 2 – Domain-Specific Execution Paths
	Backend  BackendPillar  `json:"backend"`
	Frontend FrontendPillar `json:"frontend"`

	// Phase 3 – Lifecycle Operations & Tooling
	Testing   TestingPillar   `json:"testing"`
	CICD      CICDPillar      `json:"cicd"`
	Telemetry TelemetryPillar `json:"telemetry"`
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
