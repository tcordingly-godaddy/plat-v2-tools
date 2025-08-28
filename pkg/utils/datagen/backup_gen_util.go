package datagen

import (
	"fmt"
	"math/rand"
	"strings"
)

const (
	MinNameLength = 15
	MaxNameLength = 25
)

// BackupDataGen generates a file or directory with random data for v2 workpress site
type BackupDataGen struct {
	DataGenRootDir     string
	MaxFileCountPerDir int
	SizeChoice         string
}

func NewBackupDataGen(rootDir string, maxFileCountPerDir int, sizeChoice string) *BackupDataGen {

	return &BackupDataGen{
		DataGenRootDir:     rootDir,
		MaxFileCountPerDir: maxFileCountPerDir,
		SizeChoice:         sizeChoice,
	}
}

// GenerateCreateDirectoryCommand creates a command to make a directory with random name
func (dg *BackupDataGen) GenerateCreateDirectoryCommand() string {
	return fmt.Sprintf("mkdir -p %s", dg.DataGenRootDir)
}

// GenerateCreateFileCommand creates a command to generate a file with random name and random data
func (dg *BackupDataGen) GenerateCreateFileCommand(dataGen *DataGen) string {
	// Create command to generate random data file
	// Uses head to create a file with random data from /dev/urandom
	return fmt.Sprintf("mkdir -p %s && head -c %d /dev/urandom > %s/%s",
		dg.DataGenRootDir, dataGen.GenerateRandomSize(), dg.DataGenRootDir, GenerateRandomName())
}

// GenerateMultipleFilesCommand creates a command to generate multiple files in the specified directory
func (dg *BackupDataGen) GenerateMultipleFilesCommand(dataGen *FileSizeTypeDataGen, path string, numFiles int) string {
	var commands []string

	// Determine the target directory
	targetDir := fmt.Sprintf("%s/%s", dg.DataGenRootDir, path)

	// Ensure base directory exists
	commands = append(commands, fmt.Sprintf("mkdir -p %s", targetDir))

	// Generate files directly in base directory
	for range numFiles {
		// Create a new filename for each iteration to avoid conflicts
		filePath := fmt.Sprintf("%s/%s", targetDir, GenerateRandomName())
		commands = append(commands,
			fmt.Sprintf("head -c %d /dev/urandom > %s",
				dataGen.GenerateRandomSize(), filePath))
		if dataGen.IsDone() {
			break
		}
	}

	return strings.Join(commands, " && ")
}

// GenerateFileSizeType generates the commands for creating a directories and files for a given size distribution type
func (dg *BackupDataGen) GenerateFileSizeType(sizeType *FileSizeTypeDataGen) []string {
	var cmds []string

	basePath := sizeType.Name
	for !sizeType.IsDone() {
		// If we have reached the max file count per directory, add a new directory
		if len(cmds)%dg.MaxFileCountPerDir == 0 {
			basePath = basePath + "/" + GenerateRandomName()
		}

		additionPath := basePath + "/" + GenerateRandomName()
		// Generate the files in the directory
		cmd := dg.GenerateMultipleFilesCommand(sizeType, additionPath, dg.MaxFileCountPerDir)
		cmds = append(cmds, cmd)
	}

	return cmds
}

// GenerateBackupDataOnApp generates all the commands to create the desired file distribution on the app
// No data is actually generated until the commands are executed
func (dg *BackupDataGen) GenerateBackupDataOnApp() []string {
	var backupDataGenCmds []string

	fileSizesTemplate := NewFileSizeDistribution(dg.SizeChoice)

	for _, sizeType := range fileSizesTemplate.SizeDistributions {

		cmds := dg.GenerateFileSizeType(sizeType)
		backupDataGenCmds = append(backupDataGenCmds, cmds...)
	}

	return backupDataGenCmds
}

func GenerateRandomName() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Random length between min and max
	length := rand.Intn(MaxNameLength-MinNameLength+1) + MinNameLength

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
