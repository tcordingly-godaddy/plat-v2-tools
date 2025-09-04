package datagen

import (
	"strings"
	"testing"
)

func TestNewBackupDataGen(t *testing.T) {
	tests := []struct {
		name               string
		rootDir            string
		maxFileCountPerDir int
		sizeChoice         string
		expectedRootDir    string
		expectedMaxFiles   int
		expectedSizeChoice string
	}{
		{
			name:               "medium size distribution",
			rootDir:            "./test-backup",
			maxFileCountPerDir: 25,
			sizeChoice:         "medium",
			expectedRootDir:    "./test-backup",
			expectedMaxFiles:   25,
			expectedSizeChoice: "medium",
		},
		{
			name:               "large size distribution",
			rootDir:            "/var/www/backup",
			maxFileCountPerDir: 50,
			sizeChoice:         "large",
			expectedRootDir:    "/var/www/backup",
			expectedMaxFiles:   50,
			expectedSizeChoice: "large",
		},
		{
			name:               "default size for invalid choice",
			rootDir:            "./backup",
			maxFileCountPerDir: 30,
			sizeChoice:         "invalid",
			expectedRootDir:    "./backup",
			expectedMaxFiles:   30,
			expectedSizeChoice: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewBackupDataGen(tt.rootDir, tt.maxFileCountPerDir, tt.sizeChoice)

			if gen.DataGenRootDir != tt.expectedRootDir {
				t.Errorf("Expected root dir %s, got %s", tt.expectedRootDir, gen.DataGenRootDir)
			}

			if gen.MaxFileCountPerDir != tt.expectedMaxFiles {
				t.Errorf("Expected max files %d, got %d", tt.expectedMaxFiles, gen.MaxFileCountPerDir)
			}

			if gen.SizeChoice != tt.expectedSizeChoice {
				t.Errorf("Expected size choice %s, got %s", tt.expectedSizeChoice, gen.SizeChoice)
			}
		})
	}
}

func TestGenerateCreateDirectoryCommand(t *testing.T) {
	tests := []struct {
		name        string
		rootDir     string
		expectedCmd string
	}{
		{
			name:        "simple directory path",
			rootDir:     "./backup",
			expectedCmd: "mkdir -p ./backup",
		},
		{
			name:        "absolute directory path",
			rootDir:     "/var/www/backup-data",
			expectedCmd: "mkdir -p /var/www/backup-data",
		},
		{
			name:        "nested directory path",
			rootDir:     "./wp-content/backup-gen/test",
			expectedCmd: "mkdir -p ./wp-content/backup-gen/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewBackupDataGen(tt.rootDir, 30, "medium")
			cmd := gen.GenerateCreateDirectoryCommand()

			if cmd != tt.expectedCmd {
				t.Errorf("Expected command %s, got %s", tt.expectedCmd, cmd)
			}
		})
	}
}

func TestGenerateCreateFileCommand(t *testing.T) {
	tests := []struct {
		name    string
		rootDir string
	}{
		{
			name:    "simple directory",
			rootDir: "./backup",
		},
		{
			name:    "absolute directory",
			rootDir: "/var/www/backup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewBackupDataGen(tt.rootDir, 30, "medium")
			dataGen := DefaultDataGen()
			cmd := gen.GenerateCreateFileCommand(dataGen)

			// Check that command contains required components
			if !strings.Contains(cmd, "mkdir -p") {
				t.Error("Command should contain 'mkdir -p'")
			}

			if !strings.Contains(cmd, tt.rootDir) {
				t.Errorf("Command should contain root directory %s", tt.rootDir)
			}

			if !strings.Contains(cmd, "head -c") {
				t.Error("Command should contain 'head -c'")
			}

			if !strings.Contains(cmd, "/dev/urandom") {
				t.Error("Command should contain '/dev/urandom'")
			}

			if !strings.Contains(cmd, ">") {
				t.Error("Command should contain output redirection")
			}
		})
	}
}

func TestGenerateMultipleFilesCommand(t *testing.T) {
	tests := []struct {
		name        string
		rootDir     string
		path        string
		numFiles    int
		minCommands int // minimum expected commands (mkdir + files)
	}{
		{
			name:        "single file",
			rootDir:     "./backup",
			path:        "test",
			numFiles:    1,
			minCommands: 2, // mkdir + 1 file
		},
		{
			name:        "multiple files",
			rootDir:     "/var/backup",
			path:        "data/images",
			numFiles:    5,
			minCommands: 6, // mkdir + 5 files
		},
		{
			name:        "no files",
			rootDir:     "./backup",
			path:        "empty",
			numFiles:    0,
			minCommands: 1, // just mkdir
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewBackupDataGen(tt.rootDir, 30, "medium")
			dataGen := &FileSizeTypeDataGen{
				Name: "test",
				DataGen: &DataGen{
					MinSizeInBytes: 1024,
					MaxSizeInBytes: 2048,
				},
				MaxTotalSize: 1024 * 1024,
			}

			cmd := gen.GenerateMultipleFilesCommand(dataGen, tt.path, tt.numFiles)

			// Check that command contains required components
			expectedTargetDir := tt.rootDir + "/" + tt.path
			if !strings.Contains(cmd, "mkdir -p "+expectedTargetDir) {
				t.Errorf("Command should contain 'mkdir -p %s'", expectedTargetDir)
			}

			// Count command parts separated by newlines
			commands := strings.Split(cmd, "\n")
			if len(commands) < tt.minCommands {
				t.Errorf("Expected at least %d commands, got %d", tt.minCommands, len(commands))
			}

			// For files > 0, check head commands
			if tt.numFiles > 0 {
				headCount := 0
				for _, command := range commands {
					if strings.Contains(command, "head -c") {
						headCount++
					}
				}

				if headCount != tt.numFiles {
					t.Errorf("Expected %d head commands, got %d", tt.numFiles, headCount)
				}
			}
		})
	}
}

func TestGenerateBackupDataOnApp(t *testing.T) {
	tests := []struct {
		name            string
		rootDir         string
		maxFiles        int
		sizeChoice      string
		expectedMinCmds int // minimum expected commands
	}{
		{
			name:            "medium distribution",
			rootDir:         "./backup",
			maxFiles:        30,
			sizeChoice:      "medium",
			expectedMinCmds: 1, // at least one size distribution should generate commands
		},
		{
			name:            "large distribution",
			rootDir:         "./backup",
			maxFiles:        50,
			sizeChoice:      "large",
			expectedMinCmds: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewBackupDataGen(tt.rootDir, tt.maxFiles, tt.sizeChoice)
			cmds := gen.GenerateBackupDataOnApp()

			if len(cmds) == 0 {
				t.Error("Expected generated commands, got empty string")
			}

			// Split commands by newlines and check each one
			cmdLines := strings.Split(cmds, "\n")
			nonEmptyLines := 0
			for _, line := range cmdLines {
				if strings.TrimSpace(line) != "" {
					nonEmptyLines++
				}
			}

			if nonEmptyLines < tt.expectedMinCmds {
				t.Errorf("Expected at least %d non-empty command lines, got %d", tt.expectedMinCmds, nonEmptyLines)
			}

			// Check that commands contain expected components
			if !strings.Contains(cmds, "mkdir -p") {
				t.Error("Commands should contain 'mkdir -p'")
			}

			if !strings.Contains(cmds, tt.rootDir) {
				t.Errorf("Commands should contain root directory %s", tt.rootDir)
			}
		})
	}
}

func TestGenerateRandomName(t *testing.T) {
	// Test multiple generations to ensure randomness and constraints
	names := make(map[string]bool)

	for i := 0; i < 100; i++ {
		name := GenerateRandomName()

		// Check length constraints
		if len(name) < MinNameLength || len(name) > MaxNameLength {
			t.Errorf("Name length %d is outside range [%d, %d]", len(name), MinNameLength, MaxNameLength)
		}

		// Check character set
		for _, char := range name {
			if !isValidChar(char) {
				t.Errorf("Invalid character %c in generated name %s", char, name)
			}
		}

		names[name] = true
	}

	// Check that we got some variety (not all the same name)
	if len(names) < 90 { // Allow for some small chance of duplicates
		t.Errorf("Expected high variety in names, got only %d unique names out of 100", len(names))
	}
}

func TestGenerateRandomNameCharacterSet(t *testing.T) {
	name := GenerateRandomName()

	// Test that generated name only contains valid characters
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for _, char := range name {
		found := false
		for _, validChar := range validChars {
			if char == validChar {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Generated name %s contains invalid character: %c", name, char)
		}
	}
}

func TestGenerateRandomNameLength(t *testing.T) {
	// Test length distribution over many generations
	lengths := make(map[int]int)

	for i := 0; i < 1000; i++ {
		name := GenerateRandomName()
		lengths[len(name)]++
	}

	// Check that all lengths in range are represented
	for length := MinNameLength; length <= MaxNameLength; length++ {
		if lengths[length] == 0 {
			t.Errorf("Length %d was never generated", length)
		}
	}

	// Check no lengths outside range
	for length := range lengths {
		if length < MinNameLength || length > MaxNameLength {
			t.Errorf("Invalid length %d generated", length)
		}
	}
}

// Helper function to check if character is valid
func isValidChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9')
}

// Benchmark tests
func BenchmarkGenerateRandomName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateRandomName()
	}
}

func BenchmarkGenerateCreateDirectoryCommand(b *testing.B) {
	gen := NewBackupDataGen("./benchmark", 30, "medium")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gen.GenerateCreateDirectoryCommand()
	}
}

func BenchmarkGenerateCreateFileCommand(b *testing.B) {
	gen := NewBackupDataGen("./benchmark", 30, "medium")
	dataGen := DefaultDataGen()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gen.GenerateCreateFileCommand(dataGen)
	}
}

func TestGenerateFileSizeType(t *testing.T) {
	tests := []struct {
		name            string
		rootDir         string
		maxFiles        int
		sizeTypeName    string
		minSize         int
		maxSize         int
		maxTotalSize    int64
		expectedMinCmds int
	}{
		{
			name:            "small files generation",
			rootDir:         "./test-backup",
			maxFiles:        5,
			sizeTypeName:    "small",
			minSize:         100,
			maxSize:         200,
			maxTotalSize:    1000, // Will generate about 5-10 files
			expectedMinCmds: 1,
		},
		{
			name:            "larger batch generation",
			rootDir:         "./backup",
			maxFiles:        10,
			sizeTypeName:    "medium",
			minSize:         500,
			maxSize:         1000,
			maxTotalSize:    10000, // Will generate multiple batches
			expectedMinCmds: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewBackupDataGen(tt.rootDir, tt.maxFiles, "medium")

			// Create a FileSizeTypeDataGen for testing
			sizeType := &FileSizeTypeDataGen{
				Name: tt.sizeTypeName,
				DataGen: &DataGen{
					MinSizeInBytes: tt.minSize,
					MaxSizeInBytes: tt.maxSize,
				},
				MaxTotalSize: tt.maxTotalSize,
			}

			cmds := gen.GenerateFileSizeType(sizeType)

			// Check that we got some commands
			if len(cmds) == 0 {
				t.Error("Expected generated commands, got empty string")
			}

			// Split commands by newlines and count non-empty lines
			cmdLines := strings.Split(cmds, "\n")
			nonEmptyLines := 0
			for _, line := range cmdLines {
				if strings.TrimSpace(line) != "" {
					nonEmptyLines++
				}
			}

			if nonEmptyLines < tt.expectedMinCmds {
				t.Errorf("Expected at least %d non-empty command lines, got %d", tt.expectedMinCmds, nonEmptyLines)
			}

			// All commands should be valid and contain expected components
			if !strings.Contains(cmds, "mkdir -p") {
				t.Error("Commands should contain 'mkdir -p'")
			}

			if !strings.Contains(cmds, tt.rootDir) {
				t.Errorf("Commands should contain root directory %s", tt.rootDir)
			}

			if !strings.Contains(cmds, tt.sizeTypeName) {
				t.Errorf("Commands should contain size type name %s", tt.sizeTypeName)
			}

			// Check that the size type is done after generation
			if !sizeType.IsDone() {
				t.Error("Size type should be done after generating commands")
			}
		})
	}
}

func TestGenerateFileSizeTypeEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		maxTotalSize int64
		expectedCmds int
	}{
		{
			name:         "zero max total size",
			maxTotalSize: 0,
			expectedCmds: 0,
		},
		{
			name:         "very small max total size",
			maxTotalSize: 1,
			expectedCmds: 1, // Will generate one command batch, then be done
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewBackupDataGen("./test", 10, "medium")

			sizeType := &FileSizeTypeDataGen{
				Name: "test",
				DataGen: &DataGen{
					MinSizeInBytes: 100,
					MaxSizeInBytes: 200,
				},
				MaxTotalSize: tt.maxTotalSize,
			}

			cmds := gen.GenerateFileSizeType(sizeType)

			// Count non-empty command lines
			cmdLines := strings.Split(cmds, "\n")
			nonEmptyLines := 0
			for _, line := range cmdLines {
				if strings.TrimSpace(line) != "" {
					nonEmptyLines++
				}
			}

			if tt.expectedCmds == 0 {
				if len(strings.TrimSpace(cmds)) != 0 {
					t.Errorf("Expected empty commands string, got %d non-empty lines", nonEmptyLines)
				}
			} else if nonEmptyLines < tt.expectedCmds {
				t.Errorf("Expected at least %d non-empty command lines, got %d", tt.expectedCmds, nonEmptyLines)
			}
		})
	}
}

func BenchmarkGenerateFileSizeType(b *testing.B) {
	gen := NewBackupDataGen("./benchmark", 30, "medium")
	sizeType := &FileSizeTypeDataGen{
		Name: "benchmark",
		DataGen: &DataGen{
			MinSizeInBytes: 1024,
			MaxSizeInBytes: 2048,
		},
		MaxTotalSize: 1024 * 100, // 100KB total
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset the size type for each iteration
		sizeType.DataGen.ResetBytesGenerated()
		gen.GenerateFileSizeType(sizeType)
	}
}

func BenchmarkGenerateBackupDataOnApp(b *testing.B) {
	gen := NewBackupDataGen("./benchmark", 30, "medium")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gen.GenerateBackupDataOnApp()
	}
}
