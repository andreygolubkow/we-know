# We Know

A Go-based tool for analyzing Git repositories to determine code ownership and contributions.

## Overview

We Know is a command-line application that helps you understand who owns and contributes to different parts of your codebase. It analyzes Git repositories to determine who has edited each file and how many lines they've contributed.

## Getting Started

### Prerequisites

- Go 1.16 or higher
- Git

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/we-know.git
   cd we-know
   ```

2. Install dependencies:
   ```
   go mod download
   ```

### Usage

The application analyzes a Git repository and displays information about who has edited each file:

```
go run pkg/cmd/console/main.go <repository-url> <branch>
```

Example:
```
go run pkg/cmd/console/main.go https://github.com/example/repo.git main
```

This will:
1. Clone the specified repository to a temporary directory
2. Analyze each file in the repository
3. Display information about who has edited each file and how many lines they've contributed

## Similar Projects

|  Project name   |                        Link                         |
|:---------------:|:---------------------------------------------------:|
|      cloc       |     [GitHub](https://github.com/AlDanial/cloc)      |
|    git-stats    | [GitHub](https://github.com/IonicaBizau/git-stats)  |
|  gitinspector   |   [GitHub](https://github.com/ejwa/gitinspector)    |
| git-quick-stats | [GitHub](https://github.com/arzzen/git-quick-stats) |

## License

This project is licensed under the MIT License - see the LICENSE file for details.
