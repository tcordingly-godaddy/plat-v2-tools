package datagen

import "math/rand"

const (
	MaxMedTotalSize   = int64(1024 * 1024 * 2000)      // 2GB
	MinMedTotalSize   = int64(1024 * 1024 * 300)       // 300MB
	MaxLargeTotalSize = int64(1024 * 1024 * 1024 * 10) // 10GB
	MinLargeTotalSize = int64(1024 * 1024 * 1024 * 5)  // 5GB
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
	size := rand.Intn(int(MaxMedTotalSize-MinMedTotalSize+1)) + int(MinMedTotalSize)

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
	size := rand.Intn(int(MaxLargeTotalSize-MinLargeTotalSize+1)) + int(MinLargeTotalSize)

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

func NewFileSizeDistribution(sizeChoice string) *FileSizeDistribution {
	switch sizeChoice {
	case "medium":
		return MediumSiteSizeDistributionConfig()
	case "large":
		return LargeSiteSizeDistributionConfig()
	default:
		return MediumSiteSizeDistributionConfig()
	}
}
