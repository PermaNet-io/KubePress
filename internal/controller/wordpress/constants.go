package wordpress

import crmv1 "hostzero.de/m/v2/api/v1"

// GetResourceName returns the prefixed resource name
func GetResourceName(wpName string) string {
	return wpName
}

// GetConfigMapName returns the name for the WordPress config map
func GetConfigMapName(wpName string) string {
	if len(wpName) > 63-5 { // 5 is for the suffix "--php"
		wpName = wpName[:63-5]
	}

	return GetResourceName(wpName) + "--php"
}

func GetPVCName(wpName string) string {
	return wpName
}

func GetStoragePVCName(wp *crmv1.WordPressSite) string {
	if wp.Spec.WordPress.StorageClaimName != "" {
		return wp.Spec.WordPress.StorageClaimName
	}
	return GetPVCName(wp.Name)
}

func GetSFTPServiceName(wpName string) string {
	return wpName + "--sftp"
}

// GetTLSSecretName returns the TLS secret name for the ingress
func GetTLSSecretName(wpName string) string {
	if len(wpName) > 63-5 { // 5 is for the suffix "--tls"
		wpName = wpName[:63-5]
	}

	return GetResourceName(wpName) + "--tls"
}

// GetDatabaseSecretName returns the name for the shared WordPress/database secret
func GetDatabaseSecretName(wpName string) string {
	return GetResourceName(wpName)
}
