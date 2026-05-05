package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WordPressSiteSpec defines the desired state of a WordPress site
type WordPressSiteSpec struct {
	// Site title for WordPress installation
	// Changing this after creation has no effect
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=50
	SiteTitle string `json:"siteTitle,omitempty"`

	// Admin email
	// Changing this after creation has no effect
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
	AdminEmail string `json:"adminEmail,omitempty"`

	// Admin user secret key ref
	// needs to be the name of a secret,
	// secret needs a username and password field
	// +kubebuilder:validation:Required
	AdminUserSecretKeyRef string `json:"adminUserSecretKeyRef,omitempty"`

	// Database configuration
	// +kubebuilder:validation:Required
	Database DatabaseConfig `json:"database"`

	// WordPress configuration
	// +kubebuilder:validation:Required
	WordPress WordPressConfig `json:"wordpress"`

	// Ingress configuration
	// +kubebuilder:validation:Required
	Ingress *IngressConfig `json:"ingress,omitempty"`
}

// DatabaseConfig defines the MySQL database configuration
type DatabaseConfig struct {
	// Whether to create a new database or use existing one
	// If true, a new MySQL database will be created
	// If false, connection details need to be provided via the referenced secret
	// +kubebuilder:default=true
	CreateNew bool `json:"createNew,omitempty"`
}

// WordPressConfig defines WordPress deployment configuration
type WordPressConfig struct {
	// Image to use for the WordPress container
	// +kubebuilder:default="wordpress:latest"
	Image string `json:"image,omitempty"`

	// StorageSize for WordPress persistent volume
	// +kubebuilder:default="1Gi"
	StorageSize string `json:"storageSize,omitempty"`

	// StorageClaimName overrides the managed WordPress persistent volume claim name
	// +optional
	StorageClaimName string `json:"storageClaimName,omitempty"`

	// PHP configuration overrides
	// +optional
	PHPConfig map[string]string `json:"phpConfig,omitempty"`

	// MaxUploadLimit sets the maximum upload file size (e.g., "64M")
	// +kubebuilder:default="64M"
	// +optional
	MaxUploadLimit string `json:"maxUploadLimit,omitempty"`

	// Environment variables to pass to the WordPress container
	// +optional
	Env []EnvVar `json:"env,omitempty"`

	// EnvFrom sources to pass to the WordPress container
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`

	// StorageMounts overrides the default WordPress PVC mount.
	// When empty, the managed PVC is mounted at /var/www/html for legacy behavior.
	// When set, the managed PVC is mounted only at these paths.
	// +optional
	StorageMounts []WordPressStorageMount `json:"storageMounts,omitempty"`

	// Replicas is the number of WordPress instances to run
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas,omitempty"`

	// Resource requirements for the WordPress pod
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`

	// ImagePullSecret references an optional pull secret for private WordPress images
	// +optional
	ImagePullSecret string `json:"imagePullSecret,omitempty"`
}

// EnvVar represents an environment variable in a container
type EnvVar struct {
	// Name of the environment variable
	Name string `json:"name"`

	// Value of the environment variable
	Value string `json:"value"`
}

// WordPressStorageMount maps the managed WordPress PVC to a container path.
type WordPressStorageMount struct {
	// MountPath is the path inside the WordPress container
	// +kubebuilder:validation:Required
	MountPath string `json:"mountPath"`

	// SubPath is the optional path inside the managed WordPress PVC
	// +optional
	SubPath string `json:"subPath,omitempty"`

	// ReadOnly controls whether the mount is read-only
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`
}

// ResourceRequirements defines CPU/Memory limits and requests
type ResourceRequirements struct {
	// CPU limit
	// +optional
	CPULimit string `json:"cpuLimit,omitempty"`

	// Memory limit
	// +optional
	MemoryLimit string `json:"memoryLimit,omitempty"`

	// CPU request
	// +optional
	CPURequest string `json:"cpuRequest,omitempty"`

	// Memory request
	// +optional
	MemoryRequest string `json:"memoryRequest,omitempty"`
}

// IngressConfig defines ingress configuration for WordPress
// +kubebuilder:object:generate=true
// +kubebuilder:validation:Optional
// +kubebuilder:validation:Type=object
type IngressConfig struct {
	// Enable or disable ingress for the WordPress site
	// +kubebuilder:default=true
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled,omitempty"`

	// Hostname for the WordPress site
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// Enable TLS/HTTPS
	// +kubebuilder:default=false
	// +optional
	TLS bool `json:"tls,omitempty"`

	// TLS secret name override
	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty"`

	// Annotations to add to the ingress
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels to add to the ingress
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// IngressClassName override, immutable after creation
	// +optional
	IngressClassName string `json:"ingressClassName,omitempty"`
}

// WordPressSiteStatus defines the observed state of WordPressSite
type WordPressSiteStatus struct {
	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Ready indicates whether the WordPress site is operational
	// +optional
	Ready bool `json:"ready"`

	// DeploymentStatus tracks the WordPress deployment status
	DeploymentStatus string `json:"deploymentStatus,omitempty"`

	// LastReconcileTime is the last time the resources were reconciled
	// +optional
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`

	// MySQLVersion is the version of MySQL being used
	// +optional
	MySQLVersion string `json:"mysqlVersion,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url",description="Site URL"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Site status"
// +kubebuilder:printcolumn:name="Database",type="string",JSONPath=".status.databaseStatus",description="Database status"
// +kubebuilder:printcolumn:name="Deployment",type="string",JSONPath=".status.deploymentStatus",description="Deployment status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// WordPressSite is the Schema for the wordpresssites API
type WordPressSite struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WordPressSiteSpec   `json:"spec"`
	Status WordPressSiteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WordPressSiteList contains a list of WordPressSite
type WordPressSiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WordPressSite `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WordPressSite{}, &WordPressSiteList{})
}
