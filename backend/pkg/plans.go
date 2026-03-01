package pkg

// Plan constants — Free / Pro / Business
const (
	PlanFree     = "free"
	PlanPro      = "pro"
	PlanBusiness = "business"
)

// Storage limits in bytes
const (
	StorageFree     int64 = 5 * 1024 * 1024 * 1024   // 5 GB
	StoragePro      int64 = 50 * 1024 * 1024 * 1024  // 50 GB
	StorageBusiness int64 = 200 * 1024 * 1024 * 1024 // 200 GB
)

// P2P share limits (max simultaneous active shares)
const (
	P2PLimitFree     = 5
	P2PLimitPro      = 50
	P2PLimitBusiness = 200
)

// Pricing in euro cents
const (
	PriceProMonthly      = 500  // 5,00 €/month
	PriceProYearly       = 5000 // 50,00 €/year  (~17% discount)
	PriceBusinessMonthly = 1500 // 15,00 €/month
	PriceBusinessYearly  = 15000 // 150,00 €/year (~17% discount)
)

// GetStorageLimit returns the storage limit in bytes for a given plan code.
func GetStorageLimit(plan string) int64 {
	switch plan {
	case PlanPro:
		return StoragePro
	case PlanBusiness:
		return StorageBusiness
	default:
		return StorageFree
	}
}

// GetP2PLimit returns the maximum number of active P2P shares for a plan.
func GetP2PLimit(plan string) int {
	switch plan {
	case PlanPro:
		return P2PLimitPro
	case PlanBusiness:
		return P2PLimitBusiness
	default:
		return P2PLimitFree
	}
}
