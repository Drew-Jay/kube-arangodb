//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package v1alpha

import (
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

const (
	defaultImage = "arangodb/arangodb:latest"
)

// DeploymentMode specifies the type of ArangoDB deployment to create.
type DeploymentMode string

const (
	// DeploymentModeSingle yields a single server
	DeploymentModeSingle DeploymentMode = "single"
	// DeploymentModeResilientSingle yields an agency and a resilient-single server pair
	DeploymentModeResilientSingle DeploymentMode = "resilientsingle"
	// DeploymentModeCluster yields an full cluster (agency, dbservers & coordinators)
	DeploymentModeCluster DeploymentMode = "cluster"
)

// Validate the mode.
// Return errors when validation fails, nil on success.
func (m DeploymentMode) Validate() error {
	switch m {
	case DeploymentModeSingle, DeploymentModeResilientSingle, DeploymentModeCluster:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown deployment mode: '%s'", string(m)))
	}
}

// HasSingleServers returns true when the given mode is "single" or "resilientsingle".
func (m DeploymentMode) HasSingleServers() bool {
	return m == DeploymentModeSingle || m == DeploymentModeResilientSingle
}

// HasAgents returns true when the given mode is "resilientsingle" or "cluster".
func (m DeploymentMode) HasAgents() bool {
	return m == DeploymentModeResilientSingle || m == DeploymentModeCluster
}

// HasDBServers returns true when the given mode is "cluster".
func (m DeploymentMode) HasDBServers() bool {
	return m == DeploymentModeCluster
}

// HasCoordinators returns true when the given mode is "cluster".
func (m DeploymentMode) HasCoordinators() bool {
	return m == DeploymentModeCluster
}

// SupportsSync returns true when the given mode supports dc2dc replication.
func (m DeploymentMode) SupportsSync() bool {
	return m == DeploymentModeCluster
}

// Environment in which to run the cluster
type Environment string

const (
	// EnvironmentDevelopment yields a cluster optimized for development
	EnvironmentDevelopment Environment = "development"
	// EnvironmentProduction yields a cluster optimized for production
	EnvironmentProduction Environment = "production"
)

// Validate the environment.
// Return errors when validation fails, nil on success.
func (e Environment) Validate() error {
	switch e {
	case EnvironmentDevelopment, EnvironmentProduction:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown environment: '%s'", string(e)))
	}
}

// StorageEngine specifies the type of storage engine used by the cluster
type StorageEngine string

const (
	// StorageEngineMMFiles yields a cluster using the mmfiles storage engine
	StorageEngineMMFiles StorageEngine = "mmfiles"
	// StorageEngineRocksDB yields a cluster using the rocksdb storage engine
	StorageEngineRocksDB StorageEngine = "rocksdb"
)

// Validate the storage engine.
// Return errors when validation fails, nil on success.
func (se StorageEngine) Validate() error {
	switch se {
	case StorageEngineMMFiles, StorageEngineRocksDB:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown storage engine: '%s'", string(se)))
	}
}

// validatePullPolicy the image pull policy.
// Return errors when validation fails, nil on success.
func validatePullPolicy(v v1.PullPolicy) error {
	switch v {
	case "", v1.PullAlways, v1.PullNever, v1.PullIfNotPresent:
		return nil
	default:
		return maskAny(errors.Wrapf(ValidationError, "Unknown pull policy: '%s'", string(v)))
	}
}

// RocksDBSpec holds rocksdb specific configuration settings
type RocksDBSpec struct {
	Encryption struct {
		KeySecretName string `json:"keySecretName,omitempty"`
	} `json:"encryption"`
}

// Validate the given spec
func (s RocksDBSpec) Validate() error {
	if err := k8sutil.ValidateOptionalResourceName(s.Encryption.KeySecretName); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *RocksDBSpec) SetDefaults() {
	// Nothing needed
}

// AuthenticationSpec holds authentication specific configuration settings
type AuthenticationSpec struct {
	JWTSecretName string `json:"jwtSecretName,omitempty"`
}

// Validate the given spec
func (s AuthenticationSpec) Validate() error {
	if err := k8sutil.ValidateOptionalResourceName(s.JWTSecretName); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *AuthenticationSpec) SetDefaults() {
	// Nothing needed
}

// SSLSpec holds SSL specific configuration settings
type SSLSpec struct {
	KeySecretName    string `json:"keySecretName,omitempty"`
	OrganizationName string `json:"organizationName,omitempty"`
	ServerName       string `json:"serverName,omitempty"`
}

// Validate the given spec
func (s SSLSpec) Validate() error {
	if err := k8sutil.ValidateOptionalResourceName(s.KeySecretName); err != nil {
		return maskAny(err)
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *SSLSpec) SetDefaults() {
	if s.OrganizationName == "" {
		s.OrganizationName = "ArangoDB"
	}
}

// SyncSpec holds dc2dc replication specific configuration settings
type SyncSpec struct {
	Enabled bool `json:"enabled,omitempty"`
}

// Validate the given spec
func (s SyncSpec) Validate(mode DeploymentMode) error {
	if s.Enabled && !mode.SupportsSync() {
		return maskAny(errors.Wrapf(ValidationError, "Cannot enable sync with mode: '%s'", mode))
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *SyncSpec) SetDefaults() {
}

type ServerGroup int

const (
	ServerGroupSingle       = 1
	ServerGroupAgents       = 2
	ServerGroupDBServers    = 3
	ServerGroupCoordinators = 4
	ServerGroupSyncMasters  = 5
	ServerGroupSyncWorkers  = 6
)

// AsRole returns the "role" value for the given group.
func (g ServerGroup) AsRole() string {
	switch g {
	case ServerGroupSingle:
		return "single"
	case ServerGroupAgents:
		return "agent"
	case ServerGroupDBServers:
		return "dbserver"
	case ServerGroupCoordinators:
		return "coordinator"
	case ServerGroupSyncMasters:
		return "syncmaster"
	case ServerGroupSyncWorkers:
		return "syncworker"
	default:
		return "?"
	}
}

// ServerGroupSpec contains the specification for all servers in a specific group (e.g. all agents)
type ServerGroupSpec struct {
	// Count holds the requested number of servers
	Count int `json:"count,omitempty"`
	// Args holds additional commandline arguments
	Args []string `json:"args,omitempty"`
	// StorageClassName specifies the classname for storage of the servers.
	StorageClassName string `json:"storageClassName,omitempty"`
	// Resources holds resource requests & limits
	Resources v1.ResourceRequirements `json:"resource,omitempty"`
}

// Validate the given group spec
func (s ServerGroupSpec) Validate(group ServerGroup, used bool) error {
	if used {
		if s.Count < 1 {
			return maskAny(errors.Wrapf(ValidationError, "Invalid count value %d. Expected >= 1", s.Count))
		}
	} else if s.Count != 0 {
		return maskAny(errors.Wrapf(ValidationError, "Invalid count value %d for un-used group. Expected 0", s.Count))
	}
	return nil
}

// SetDefaults fills in missing defaults
func (s *ServerGroupSpec) SetDefaults(group ServerGroup, used bool) {
	if s.Count == 0 && used {
		switch group {
		case ServerGroupSingle:
			s.Count = 1
		default:
			s.Count = 3
		}
	}
	if _, found := s.Resources.Requests[v1.ResourceStorage]; !found {
		switch group {
		case ServerGroupSingle, ServerGroupAgents, ServerGroupDBServers:
			if s.Resources.Requests == nil {
				s.Resources.Requests = make(map[v1.ResourceName]resource.Quantity)
			}
			s.Resources.Requests[v1.ResourceStorage] = resource.MustParse("8Gi")
		}
	}
}

// DeploymentSpec contains the spec part of a ArangoDeployment resource.
type DeploymentSpec struct {
	Mode            DeploymentMode `json:"mode,omitempty"`
	Environment     Environment    `json:"environment,omitempty"`
	StorageEngine   StorageEngine  `json:"storageEngine,omitempty"`
	Image           string         `json:"image,omitempty"`
	ImagePullPolicy v1.PullPolicy  `json:"imagePullPolicy,omitempty"`

	RocksDB        RocksDBSpec        `json:"rocksdb"`
	Authentication AuthenticationSpec `json:"auth"`
	SSL            SSLSpec            `json:"ssl"`
	Sync           SyncSpec           `json:"sync"`

	Single       ServerGroupSpec `json:"single"`
	Agents       ServerGroupSpec `json:"agents"`
	DBServers    ServerGroupSpec `json:"dbservers"`
	Coordinators ServerGroupSpec `json:"coordinators"`
	SyncMasters  ServerGroupSpec `json:"syncmasters"`
	SyncWorkers  ServerGroupSpec `json:"syncworkers"`
}

// SetDefaults fills in default values when a field is not specified.
func (s *DeploymentSpec) SetDefaults() {
	if s.Mode == "" {
		s.Mode = DeploymentModeCluster
	}
	if s.Environment == "" {
		s.Environment = EnvironmentDevelopment
	}
	if s.StorageEngine == "" {
		s.StorageEngine = StorageEngineMMFiles
	}
	if s.Image == "" && s.IsDevelopment() {
		s.Image = defaultImage
	}
	s.RocksDB.SetDefaults()
	s.Authentication.SetDefaults()
	s.SSL.SetDefaults()
	s.Sync.SetDefaults()
	s.Single.SetDefaults(ServerGroupSingle, s.Mode.HasSingleServers())
	s.Agents.SetDefaults(ServerGroupAgents, s.Mode.HasAgents())
	s.DBServers.SetDefaults(ServerGroupDBServers, s.Mode.HasDBServers())
	s.Coordinators.SetDefaults(ServerGroupCoordinators, s.Mode.HasCoordinators())
	s.SyncMasters.SetDefaults(ServerGroupSyncMasters, s.Sync.Enabled)
	s.SyncWorkers.SetDefaults(ServerGroupSyncWorkers, s.Sync.Enabled)
}

// Validate the specification.
// Return errors when validation fails, nil on success.
func (s *DeploymentSpec) Validate() error {
	if err := s.Mode.Validate(); err != nil {
		return maskAny(err)
	}
	if err := s.Environment.Validate(); err != nil {
		return maskAny(err)
	}
	if err := s.StorageEngine.Validate(); err != nil {
		return maskAny(err)
	}
	if err := validatePullPolicy(s.ImagePullPolicy); err != nil {
		return maskAny(err)
	}
	if s.Image == "" {
		return maskAny(errors.Wrapf(ValidationError, "image must be set"))
	}
	if err := s.RocksDB.Validate(); err != nil {
		return maskAny(err)
	}
	if err := s.Authentication.Validate(); err != nil {
		return maskAny(err)
	}
	if err := s.SSL.Validate(); err != nil {
		return maskAny(err)
	}
	if err := s.Sync.Validate(s.Mode); err != nil {
		return maskAny(err)
	}
	if err := s.Single.Validate(ServerGroupSingle, s.Mode.HasSingleServers()); err != nil {
		return maskAny(err)
	}
	if err := s.Agents.Validate(ServerGroupAgents, s.Mode.HasAgents()); err != nil {
		return maskAny(err)
	}
	if err := s.DBServers.Validate(ServerGroupDBServers, s.Mode.HasDBServers()); err != nil {
		return maskAny(err)
	}
	if err := s.Coordinators.Validate(ServerGroupCoordinators, s.Mode.HasCoordinators()); err != nil {
		return maskAny(err)
	}
	if err := s.SyncMasters.Validate(ServerGroupSyncMasters, s.Sync.Enabled); err != nil {
		return maskAny(err)
	}
	if err := s.SyncWorkers.Validate(ServerGroupSyncWorkers, s.Sync.Enabled); err != nil {
		return maskAny(err)
	}
	return nil
}

// IsDevelopment returns true when the spec contains a Development environment.
func (s DeploymentSpec) IsDevelopment() bool {
	return s.Environment == EnvironmentDevelopment
}