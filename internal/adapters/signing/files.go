// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package signing

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MaxPathLength defines a safe maximum path length limit to prevent resource exhaustion.
// Most filesystems have a limit of 4096, but we're being more conservative.
const MaxPathLength = 1024

// MaxSymlinkDepth defines the maximum depth of symlink resolution to prevent
// infinite loops from circular symlinks or excessive resource consumption.
const MaxSymlinkDepth = 10

// SafeWriteFile writes a file atomically with proper permissions and error handling.
// This prevents TOCTOU vulnerabilities by writing to a temporary file first and then
// atomically moving it to the destination. It sets file permissions before writing any data.
func SafeWriteFile(path string, data []byte, perm os.FileMode) error {
	// Get the directory portion of the destination path
	dir := filepath.Dir(path)

	// Ensure the directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create a temporary file in the same directory for atomic operations
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	tmpPath := tmpFile.Name()

	// Ensure temporary file is cleaned up on error
	successful := false
	defer func() {
		// Close the file if it's still open
		tmpFile.Close()

		// If the operation wasn't successful, remove the temp file
		if !successful {
			os.Remove(tmpPath)
		}
	}()

	// Set permissions before writing content (important for security)
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Write data to the temporary file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// Ensure data is written to disk
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file data: %w", err)
	}

	// Close the file explicitly before rename
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Atomically move the temporary file to the destination
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to move temporary file to destination: %w", err)
	}

	// Mark operation as successful so the deferred cleanup doesn't remove the file
	successful = true

	return nil
}

// FindFilesWithExtensions returns all files in a directory with the specified extensions.
// It does not recurse into subdirectories.
func FindFilesWithExtensions(dir string, extensions []string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Convert extensions slice to a map for O(1) lookups
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[ext] = true
	}

	// Pre-allocate slice with a reasonable capacity
	// A good estimate is to allocate for half the entries
	// since not all entries will match our extensions
	files := make([]string, 0, len(entries)/2)

	// Filter and map entries to full paths
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if extMap[ext] {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

// SanitizePath sanitizes and validates a directory path with robust security checks.
// It resolves symbolic links, ensures the path is a directory, and performs
// several security validations to prevent common vulnerabilities.
func SanitizePath(path string) (string, error) {
	// Normalize path to handle .. sequences properly
	cleanPath := filepath.Clean(path)

	// Get absolute path (applies further normalization)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Check if path has appropriate canonical form and resolve symlinks
	canonPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", fmt.Errorf("path evaluation failed: %w", err)
	}

	// Check if path exists and is a directory (after symlink resolution)
	fileInfo, err := os.Stat(canonPath)
	if err != nil {
		return "", fmt.Errorf("path error: %w", err)
	}

	if !fileInfo.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", canonPath)
	}

	return canonPath, nil
}

// IsWithinDirectory checks if a path is contained within a base directory.
// This function helps prevent directory traversal attacks by ensuring
// one path is contained within another. It resolves symlinks to their
// real paths before checking containment.
func IsWithinDirectory(path, baseDir string) (bool, error) {
	// Ensure both paths are absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("invalid path: %w", err)
	}

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return false, fmt.Errorf("invalid base directory: %w", err)
	}

	// Check if paths exist before trying to resolve symlinks
	_, err = os.Stat(absPath)
	if err != nil {
		// Allow checking paths that don't exist yet, but fail on other errors
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("path error: %w", err)
		}
	} else {
		// Path exists, resolve symlinks
		resolvedPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return false, fmt.Errorf("path evaluation failed: %w", err)
		}

		absPath = resolvedPath
	}

	_, err = os.Stat(absBaseDir)
	if err != nil {
		return false, fmt.Errorf("base directory error: %w", err)
	}

	// Always resolve symlinks in the base directory
	canonBaseDir, err := filepath.EvalSymlinks(absBaseDir)
	if err != nil {
		return false, fmt.Errorf("base directory evaluation failed: %w", err)
	}

	// Calculate a relative path to verify containment
	relPath, err := filepath.Rel(canonBaseDir, absPath)
	if err != nil {
		return false, err
	}

	// A path is within base dir if the relative path doesn't start with ".."
	return !strings.HasPrefix(relPath, ".."), nil
}

// SafeJoin safely joins path elements, preventing path traversal attacks.
// It rejects any element containing ".." sequences, guards against symlink attacks,
// and ensures the final path remains within the base directory using multiple
// layers of validation.
func SafeJoin(baseDir string, elements ...string) (string, error) {
	// Verify baseDir exists and is a directory using a file descriptor
	// to prevent TOCTOU race conditions
	dirFile, err := os.Open(baseDir)
	if err != nil {
		return "", fmt.Errorf("invalid base directory: %w", err)
	}
	defer dirFile.Close()

	// Get file info directly from the file descriptor
	fi, err := dirFile.Stat()
	if err != nil {
		return "", fmt.Errorf("cannot stat base directory: %w", err)
	}

	if !fi.IsDir() {
		return "", fmt.Errorf("base path is not a directory: %s", baseDir)
	}

	// Clean base path and get its absolute form
	absBaseDir, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return "", fmt.Errorf("cannot get absolute base directory: %w", err)
	}

	// Resolve base directory symlinks for security
	canonBase, err := filepath.EvalSymlinks(absBaseDir)
	if err != nil {
		return "", fmt.Errorf("cannot resolve base directory symlinks: %w", err)
	}

	// Check each element and build path incrementally
	// Start with the canonical base path
	result := canonBase

	// Additional security check - verify each component individually
	for elementIndex, element := range elements {
		// Reject empty elements
		if element == "" {
			continue
		}

		// First clean the element to normalize it
		cleanElement := filepath.Clean(element)

		// Comprehensive path traversal blocking
		// 1. Check for parent directory references (..)
		if strings.Contains(element, "..") || strings.Contains(cleanElement, "..") {
			return "", fmt.Errorf("path element contains illegal sequence: %s", element)
		}

		// 2. Check for other suspicious patterns
		if strings.Contains(cleanElement, "\\") ||
			strings.Contains(cleanElement, "%2e%2e") || // URL-encoded ..
			strings.Contains(cleanElement, "%2E%2E") {
			return "", fmt.Errorf("path element contains suspicious character sequence: %s", element)
		}

		// 3. Reject absolute paths in elements
		if filepath.IsAbs(cleanElement) {
			return "", fmt.Errorf("path element cannot be absolute: %s", element)
		}

		// Safely join the element
		nextResult := filepath.Join(result, cleanElement)

		// Perform thorough path containment verification
		// 4. Check path containment using relative path
		relPath, err := filepath.Rel(canonBase, nextResult)
		if err != nil {
			return "", fmt.Errorf("path calculation error: %w", err)
		}

		if strings.HasPrefix(relPath, "..") || strings.Contains(relPath, "/../") {
			return "", fmt.Errorf("path element causes directory traversal: %s", element)
		}

		// 5. Check path containment using string prefix (belt and suspenders approach)
		// Ensure resulting path has canonical base as prefix
		if !strings.HasPrefix(nextResult, canonBase+string(filepath.Separator)) && nextResult != canonBase {
			return "", fmt.Errorf("path escapes base directory: %s", element)
		}

		// 6. For all but the last element, verify each component if it exists
		// This prevents symlink trickery where intermediate components are symlinks
		if elementIndex < len(elements)-1 {
			if _, err := os.Lstat(nextResult); err == nil {
				// Path exists, check if it's a symlink
				realPath, err := filepath.EvalSymlinks(nextResult)
				if err != nil {
					return "", fmt.Errorf("failed to evaluate potential symlink: %w", err)
				}

				// Verify the resolved path is still within the base directory
				if !strings.HasPrefix(realPath, canonBase) {
					return "", fmt.Errorf("symlink in path escapes base directory: %s", element)
				}

				// Use the resolved path going forward if it exists
				nextResult = realPath
			}
		}

		// Update our result for the next iteration
		result = nextResult
	}

	return result, nil
}

// ValidateGitRepoPath performs extensive validation of repository paths with additional
// checks specific to git repositories. This function is a comprehensive, secure path
// validator that handles all edge cases and prevents security vulnerabilities.
// Returns the canonical path to the repository root.
func ValidateGitRepoPath(repoPath string) (string, error) {
	// Empty path validation
	if repoPath == "" {
		return "", errors.New("repository path cannot be empty")
	}

	// Normalize the path to handle .. sequences properly
	cleanPath := filepath.Clean(repoPath)

	// Reject suspicious path patterns
	if strings.Contains(cleanPath, "\\") ||
		strings.Contains(cleanPath, "%2e%2e") || // URL-encoded ..
		strings.Contains(cleanPath, "%2E%2E") {
		return "", fmt.Errorf("repository path contains suspicious character sequences: %s", repoPath)
	}

	// Convert to absolute path for consistent validation
	if !filepath.IsAbs(cleanPath) {
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return "", fmt.Errorf("could not resolve absolute path: %w", err)
		}

		cleanPath = absPath
	}

	// Check maximum path length to prevent resource exhaustion
	if len(cleanPath) > MaxPathLength {
		return "", fmt.Errorf("repository path exceeds maximum length (%d characters)", MaxPathLength)
	}

	// Resolve symlinks to get canonical path with custom depth limiting
	canonPath, err := safeResolveSymlinks(cleanPath, MaxSymlinkDepth)
	if err != nil {
		return "", fmt.Errorf("path evaluation failed: %w", err)
	}

	// Validate the path exists and has proper permissions using a file descriptor
	// to prevent TOCTOU race conditions
	dirFile, err := os.Open(canonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("repository path does not exist: %s", canonPath)
		}

		if os.IsPermission(err) {
			return "", fmt.Errorf("insufficient permissions to access repository path: %s", canonPath)
		}

		return "", fmt.Errorf("invalid repository path: %w", err)
	}
	defer dirFile.Close()

	// Get file info directly from the file descriptor
	fileInfo, err := dirFile.Stat()
	if err != nil {
		return "", fmt.Errorf("could not stat repository path: %w", err)
	}

	// Ensure it's a directory
	if !fileInfo.IsDir() {
		return "", fmt.Errorf("repository path is not a directory: %s", canonPath)
	}

	// Check for read/write permissions
	if fileInfo.Mode().Perm()&0600 != 0600 {
		return "", fmt.Errorf("insufficient permissions to read/write in repository path: %s", canonPath)
	}

	// Validate that the .git directory exists
	gitDirPath := filepath.Join(canonPath, ".git")

	gitPath, err := verifyGitDirectory(gitDirPath)
	if err != nil {
		return "", err
	}

	// Validate that the git directory is within the repository root
	isWithin, err := IsWithinDirectory(gitPath, canonPath)
	if err != nil {
		return "", fmt.Errorf("path validation error: %w", err)
	}

	if !isWithin {
		return "", errors.New("invalid git directory: path traversal detected")
	}

	return canonPath, nil
}

// safeResolveSymlinks resolves symlinks in a path with a maximum depth limit
// to prevent infinite loops and excessive resource consumption.
func safeResolveSymlinks(path string, maxDepth int) (string, error) {
	if maxDepth <= 0 {
		return "", errors.New("maximum symlink resolution depth exceeded")
	}

	// Get file info to check if it's a symlink
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}

	// If not a symlink, we're done with this component
	if info.Mode()&os.ModeSymlink == 0 {
		// If it's a directory, it's a final path component
		if info.IsDir() {
			return filepath.Abs(path)
		}

		// If it's a file, resolve its directory
		dir := filepath.Dir(path)

		resolvedDir, err := safeResolveSymlinks(dir, maxDepth)
		if err != nil {
			return "", err
		}

		return filepath.Join(resolvedDir, filepath.Base(path)), nil
	}

	// It's a symlink, read the link target
	linkTarget, err := os.Readlink(path)
	if err != nil {
		return "", err
	}

	// If relative link, join with parent directory
	if !filepath.IsAbs(linkTarget) {
		dir := filepath.Dir(path)
		linkTarget = filepath.Join(dir, linkTarget)
	}

	// Clean up the path
	linkTarget = filepath.Clean(linkTarget)

	// Recurse to handle multi-level symlinks, decrementing depth counter
	return safeResolveSymlinks(linkTarget, maxDepth-1)
}

// verifyGitDirectory checks that the .git directory is valid.
// It handles both regular git repositories and git submodules.
func verifyGitDirectory(gitDirPath string) (string, error) {
	// First check if .git exists
	fileInfo, err := os.Lstat(gitDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("not a git repository (no .git directory found)")
		}

		return "", fmt.Errorf("error accessing .git directory: %w", err)
	}

	// Check if it's a directory (normal git repository)
	if fileInfo.IsDir() {
		// Verify essential git files exist to confirm it's a real git repository
		configPath := filepath.Join(gitDirPath, "config")
		if _, err := os.Stat(configPath); err != nil {
			return "", errors.New("invalid git repository (missing config file)")
		}

		// Looks like a valid git directory
		return gitDirPath, nil
	}

	// Not a directory - could be a git submodule with a gitfile link
	if fileInfo.Mode().IsRegular() {
		// Try to read it as a gitfile (used in submodules)
		content, err := os.ReadFile(gitDirPath)
		if err != nil {
			return "", fmt.Errorf("error reading git file: %w", err)
		}

		// Parse the gitfile content
		gitfileContent := strings.TrimSpace(string(content))
		if strings.HasPrefix(gitfileContent, "gitdir: ") {
			// Extract the actual git directory path
			actualGitDir := strings.TrimPrefix(gitfileContent, "gitdir: ")

			// Resolve the path if relative
			if !filepath.IsAbs(actualGitDir) {
				repoRoot := filepath.Dir(gitDirPath)
				actualGitDir = filepath.Join(repoRoot, actualGitDir)
			}

			// Confirm this is a real git directory
			if _, err := os.Stat(filepath.Join(actualGitDir, "config")); err != nil {
				return "", errors.New("invalid git submodule reference")
			}

			// Validate that this path isn't trying to escape
			cleanedPath := filepath.Clean(actualGitDir)
			if strings.Contains(cleanedPath, "..") {
				return "", errors.New("invalid git submodule path (contains path traversal)")
			}

			return cleanedPath, nil
		}

		return "", errors.New("invalid git file format")
	}

	// Neither a directory nor a regular file
	return "", errors.New("invalid git repository structure")
}

// IsValidGitHookType verifies if a hook type is valid to prevent command injection.
// Git hooks need to have specific, well-known names.
func IsValidGitHookType(hookType string) bool {
	validHooks := map[string]bool{
		"commit-msg":            true,
		"pre-commit":            true,
		"pre-push":              true,
		"pre-receive":           true,
		"update":                true,
		"post-update":           true,
		"post-receive":          true,
		"post-merge":            true,
		"pre-rebase":            true,
		"pre-applypatch":        true,
		"post-applypatch":       true,
		"post-checkout":         true,
		"post-commit":           true,
		"pre-auto-gc":           true,
		"post-rewrite":          true,
		"sendemail-validate":    true,
		"fsmonitor-watchman":    true,
		"p4-pre-submit":         true,
		"p4-prepare-changelist": true,
		"p4-changelist":         true,
		"p4-post-changelist":    true,
		"prepare-commit-msg":    true,
	}

	// Check if hookType is in the valid hooks map
	_, exists := validHooks[hookType]

	return exists
}

// GetGitHookPath safely computes the path to a git hook.
// Given a repository path and hook type, returns the absolute path to the hook.
// This function handles path validation and security checks.
func GetGitHookPath(repoPath, hookType string) (string, error) {
	// Validate repository path
	canonPath, err := ValidateGitRepoPath(repoPath)
	if err != nil {
		return "", fmt.Errorf("invalid repository path: %w", err)
	}

	// Validate hook type
	if !IsValidGitHookType(hookType) {
		return "", fmt.Errorf("invalid hook type: %s", hookType)
	}

	// Get git directory path - this handles submodules correctly
	gitDir := filepath.Join(canonPath, ".git")

	gitDirPath, err := verifyGitDirectory(gitDir)
	if err != nil {
		return "", fmt.Errorf("invalid git directory: %w", err)
	}

	// Use SafeJoin to safely get hooks directory path
	hooksDir, err := SafeJoin(gitDirPath, "hooks")
	if err != nil {
		return "", fmt.Errorf("invalid hooks directory: %w", err)
	}

	// Use SafeJoin to safely get hook path
	hookPath, err := SafeJoin(hooksDir, hookType)
	if err != nil {
		return "", fmt.Errorf("invalid hook path: %w", err)
	}

	// Additional verification that hook path is within git directory
	isWithin, err := IsWithinDirectory(hookPath, canonPath)
	if err != nil {
		return "", fmt.Errorf("path validation error: %w", err)
	}

	if !isWithin {
		return "", errors.New("invalid hook path: path traversal detected")
	}

	return hookPath, nil
}
