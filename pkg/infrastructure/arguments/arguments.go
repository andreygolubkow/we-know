package arguments

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"we-know/pkg/infrastructure/report"
)

type Arguments struct {
	Path       string
	reportType report.ReportType
}

// ReadArguments parses command line arguments and returns an Arguments struct
func ReadArguments() (*Arguments, error) {
	args := &Arguments{
		reportType: report.ReportByFileUsers,
	}

	if len(os.Args) < 2 {
		return nil, fmt.Errorf("path is required")
	}

	args.Path = os.Args[1]

	// Parse remaining arguments
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]

		// Handle optional arguments that start with --
		if strings.HasPrefix(arg, "--") {
			if i+1 >= len(os.Args) {
				return nil, fmt.Errorf("value missing for argument %s", arg)
			}

			switch arg {
			case "--type":
				typeStr := os.Args[i+1]
				if typeStr == "user" {
					args.reportType = report.ReportByFileUsers
				} else if typeStr == "team" {
					args.reportType = report.ReportByFileTeams
				} else {
					return nil, fmt.Errorf("unknown argument value: %s - %s. Possible values: user, team", arg, typeStr)
				}
				i++
			default:
				return nil, fmt.Errorf("unknown argument: %s", arg)
			}
			continue
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
