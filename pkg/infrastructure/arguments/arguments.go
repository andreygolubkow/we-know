package arguments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	WrongArgumentsException = "Nothing to clone"
)

type Arguments struct {
	RepositoryURL string
	Branch        string
	Path          string
	// Add new fields here for future arguments
}

// ReadArguments parses command line arguments and returns an Arguments struct
func ReadArguments() (*Arguments, error) {
	args := &Arguments{}

	// Need at least repository URL
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("repository URL is required")
	}

	args.RepositoryURL = os.Args[1]

	// Parse remaining arguments
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]

		// Handle optional arguments that start with --
		if strings.HasPrefix(arg, "--") {
			if i+1 >= len(os.Args) {
				return nil, fmt.Errorf("value missing for argument %s", arg)
			}

			switch arg {
			case "--path":
				args.Path = os.Args[i+1]
				i++ // Skip the next argument since we consumed it
			// Add new argument cases here
			default:
				return nil, fmt.Errorf("unknown argument: %s", arg)
			}
			continue
		}

		// If not an optional argument, treat as branch name
		// (only if branch is not set yet)
		if args.Branch == "" {
			args.Branch = arg
		}
	}

	// Set defaults if not provided
	if args.Path == "" {
		workingDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %v", err)
		}
		args.Path = filepath.Join(workingDir, "tmp")
	}

	// Create directory if it doesn't exist
	if info, err := os.Stat(args.Path); os.IsNotExist(err) || !info.IsDir() {
		if err := os.MkdirAll(args.Path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %v", args.Path, err)
		}
	}

	return args, nil
}
