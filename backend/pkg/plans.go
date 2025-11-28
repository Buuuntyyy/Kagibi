package pkg

// Plan constants
const (
	PlanFree       = "free"
	PlanBasic      = "basic"
	PlanPro        = "pro"
	PlanEnterprise = "enterprise"
)

// Storage limits in bytes
const (
	StorageFree       int64 = 15 * 1024 * 1024 * 1024       // 15 GB
	StorageBasic      int64 = 100 * 1024 * 1024 * 1024      // 100 GB
	StoragePro        int64 = 1 * 1024 * 1024 * 1024 * 1024 // 1 TB
	StorageEnterprise int64 = 5 * 1024 * 1024 * 1024 * 1024 // 5 TB
)

// GetStorageLimit returns the storage limit for a given plan
func GetStorageLimit(plan string) int64 {
	switch plan {
	case PlanBasic:
		return StorageBasic
	case PlanPro:
		return StoragePro
	case PlanEnterprise:
		return StorageEnterprise
	default:
		return StorageFree
	}
}
