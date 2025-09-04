package datagen

import "math/rand"

const (
	TotalSize300MB = int64(1024 * 1024 * 300)       // 300MB
	TotalSize500MB = int64(1024 * 1024 * 500)       // 500MB
	TotalSize2GB   = int64(1024 * 1024 * 2000)      // 2GB
	TotalSize5GB   = int64(1024 * 1024 * 1024 * 5)  // 5GB
	TotalSize10GB  = int64(1024 * 1024 * 1024 * 10) // 10GB
)

// FileSizeDistribution defines how data should be distributed across different file size categories
type FileSizeDistribution struct {
	MinTotalSize int // minimum total size to generate in bytes
	MaxTotalSize int // maximum total size to generate in bytes

	SizeDistributions []*FileSizeTypeDataGen
}

// FileSizeTypeDataGen defines a data generation configuration for a specific file size category
type FileSizeTypeDataGen struct {
	Name         string
	DataGen      *DataGen
	MaxTotalSize int64 // maximum total size in bytes for this file size category
}

func NewFileSizeDataGen(minFileSize, maxFileSize, maxTotalSize int) *FileSizeTypeDataGen {
	return &FileSizeTypeDataGen{
		DataGen: &DataGen{
			MinSizeInBytes: minFileSize,
			MaxSizeInBytes: maxFileSize,
		},
		MaxTotalSize: int64(maxTotalSize),
	}
}

func (f *FileSizeTypeDataGen) GenerateRandomSize() int {
	return f.DataGen.GenerateRandomSize()
}

func (f *FileSizeTypeDataGen) IsDone() bool {
	return f.DataGen.GetBytesGenerated() >= f.MaxTotalSize
}

func MediumSiteSizeDistributionConfig() *FileSizeDistribution {
	// Generate a random size between MinMedTotalSize and MaxMedTotalSize
	size := rand.Intn(int(TotalSize2GB-TotalSize300MB+1)) + int(TotalSize300MB)

	return &FileSizeDistribution{
		SizeDistributions: []*FileSizeTypeDataGen{
			{
				Name: "large",
				DataGen: &DataGen{
					MinSizeInBytes: 1024 * 1024 * 1, // 1MB
					MaxSizeInBytes: 1024 * 1024 * 5, // 5MB
				},
				MaxTotalSize: int64((size * 10) / 100), // 10% large files
			},
			{
				Name: "medium",
				DataGen: &DataGen{
					MinSizeInBytes: 1024 * 400,  // 400KB
					MaxSizeInBytes: 1024 * 1024, // 1MB
				},
				MaxTotalSize: int64((size * 60) / 100), // 60% medium files
			},
			{
				Name: "small",
				DataGen: &DataGen{
					MinSizeInBytes: 1024 * 150, // 150KB
					MaxSizeInBytes: 1024 * 400, // 400KB
				},
				MaxTotalSize: int64((size * 30) / 100), // 30% small files
			},
		},
	}
}

func LargeSiteSizeDistributionConfig() *FileSizeDistribution {
	// Generate a random size between MinLargeTotalSize and MaxLargeTotalSize
	size := rand.Intn(int(TotalSize10GB-TotalSize5GB+1)) + int(TotalSize5GB)

	return &FileSizeDistribution{
		SizeDistributions: []*FileSizeTypeDataGen{
			{
				Name: "large",
				DataGen: &DataGen{
					MinSizeInBytes: 1024 * 1024 * 1,  // 1MB
					MaxSizeInBytes: 1024 * 1024 * 20, // 20MB
				},
				MaxTotalSize: int64((size * 45) / 100), // 45% large files
			},
			{
				Name: "medium",
				DataGen: &DataGen{
					MinSizeInBytes: 1024 * 400,  // 400KB
					MaxSizeInBytes: 1024 * 1024, // 1MB
				},
				MaxTotalSize: int64((size * 35) / 100), // 35% medium files
			},
			{
				Name: "small",
				DataGen: &DataGen{
					MinSizeInBytes: 1024 * 150, // 150KB
					MaxSizeInBytes: 1024 * 400, // 400KB
				},
				MaxTotalSize: int64((size * 20) / 100), // 20% small files
			},
		},
	}
}

// Generate over 1 million files
func P95FileCountSizeDistributionConfig() *FileSizeDistribution {
	// Generate a random size between MinMedTotalSize and MaxMedTotalSize
	return &FileSizeDistribution{
		SizeDistributions: []*FileSizeTypeDataGen{
			{
				Name: "p95",
				DataGen: &DataGen{
					MinSizeInBytes: 1024 * 2, // 2kB
					MaxSizeInBytes: 1024 * 5, // 5kB
				},
				MaxTotalSize: int64(rand.Intn(int(TotalSize5GB-TotalSize2GB+1)) + int(TotalSize2GB)),
			},
		},
	}
}

func P90FileCountSizeDistributionConfig() *FileSizeDistribution {
	// Generate a random size between MinMedTotalSize and MaxMedTotalSize
	return &FileSizeDistribution{
		SizeDistributions: []*FileSizeTypeDataGen{
			{
				Name: "p90",
				DataGen: &DataGen{
					MinSizeInBytes: 1024 * 20, // 20kB
					MaxSizeInBytes: 1024 * 50, // 50kB
				},
				MaxTotalSize: int64(rand.Intn(int(TotalSize5GB-TotalSize2GB+1)) + int(TotalSize2GB)),
			},
		},
	}
}

func P75FileCountSizeDistributionConfig() *FileSizeDistribution {
	// Generate a random size between MinMedTotalSize and MaxMedTotalSize
	return &FileSizeDistribution{
		SizeDistributions: []*FileSizeTypeDataGen{
			{
				Name: "p75",
				DataGen: &DataGen{
					MinSizeInBytes: 1024 * 100, // 100kB
					MaxSizeInBytes: 1024 * 250, // 250kB
				},
				MaxTotalSize: int64(rand.Intn(int(TotalSize5GB-TotalSize2GB+1)) + int(TotalSize2GB)),
			},
		},
	}
}

func P50FileCountSizeDistributionConfig() *FileSizeDistribution {
	// Generate a random size between MinMedTotalSize and MaxMedTotalSize
	return &FileSizeDistribution{
		SizeDistributions: []*FileSizeTypeDataGen{
			{
				Name: "p50",
				DataGen: &DataGen{
					MinSizeInBytes: 1024 * 175, // 175kB
					MaxSizeInBytes: 1024 * 450, // 450kB
				},
				MaxTotalSize: int64(rand.Intn(int(TotalSize5GB-TotalSize2GB+1)) + int(TotalSize2GB)),
			},
		},
	}
}

func FileCountSizeDistributionConfig() *FileSizeDistribution {
	// Generate a random size between MinMedTotalSize and MaxMedTotalSize
	return &FileSizeDistribution{
		SizeDistributions: []*FileSizeTypeDataGen{
			{
				Name: "fileCount",
				DataGen: &DataGen{
					MinSizeInBytes: 1024 * 5,  // 5	kB
					MaxSizeInBytes: 1024 * 50, // 50 kB
				},
				MaxTotalSize: int64(rand.Intn(int(TotalSize500MB-TotalSize300MB+1)) + int(TotalSize300MB)),
			},
		},
	}
}

func NewFileSizeDistribution(sizeChoice string) *FileSizeDistribution {
	switch sizeChoice {
	case "medium":
		return MediumSiteSizeDistributionConfig()
	case "large":
		return LargeSiteSizeDistributionConfig()
	case "p95":
		return P95FileCountSizeDistributionConfig()
	case "p90":
		return P90FileCountSizeDistributionConfig()
	case "p75":
		return P75FileCountSizeDistributionConfig()
	case "p50":
		return P50FileCountSizeDistributionConfig()
	case "fileCount":
		return FileCountSizeDistributionConfig()
	default:
		return MediumSiteSizeDistributionConfig()
	}
}
