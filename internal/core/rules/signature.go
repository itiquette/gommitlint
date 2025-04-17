// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/itiquette/gommitlint/internal/domain"
)

// SignatureRule enforces that commits are cryptographically signed using SSH or GPG keys.
//
// This rule helps ensure security and verifiability in the git history by requiring
// all commits to be properly signed. Signed commits establish a chain of trust and
// provide strong assurance about who authored each change, which is especially
// important for security-sensitive projects.
type SignatureRule struct {
	errors           []*domain.ValidationError
	requireSignature bool
	allowedSigTypes  []string
}

// SignatureOption is a function that modifies a SignatureRule.
type SignatureOption func(*SignatureRule)

// WithRequireSignature configures whether a signature is required.
func WithRequireSignature(require bool) SignatureOption {
	return func(rule *SignatureRule) {
		rule.requireSignature = require
	}
}

// WithAllowedSignatureTypes sets the allowed signature types.
// Valid types are "gpg" and "ssh".
func WithAllowedSignatureTypes(types []string) SignatureOption {
	return func(rule *SignatureRule) {
		rule.allowedSigTypes = types
	}
}

// NewSignatureRule creates a new SignatureRule with the specified options.
func NewSignatureRule(options ...SignatureOption) *SignatureRule {
	rule := &SignatureRule{
		errors:           []*domain.ValidationError{},
		requireSignature: true,                   // Default to requiring signature
		allowedSigTypes:  []string{"gpg", "ssh"}, // Default to allowing both types
	}

	// Apply provided options
	for _, option := range options {
		option(rule)
	}

	return rule
}

// Name returns the name of the rule.
func (r *SignatureRule) Name() string {
	return "Signature"
}

// Validate validates a commit against the rule.
func (r *SignatureRule) Validate(commit *domain.CommitInfo) []*domain.ValidationError {
	// Reset errors
	r.errors = []*domain.ValidationError{}

	// Get signature from commit
	signature := commit.Signature

	// Check for empty signature
	if signature == "" || len(strings.TrimSpace(signature)) == 0 {
		// If signature is not required, this is not an error
		if !r.requireSignature {
			return r.errors
		}

		r.addError(
			domain.ValidationErrorMissingSignature,
			"commit does not have a SSH/GPG signature",
			nil,
		)

		return r.errors
	}

	// Trim whitespace for validation
	signature = strings.TrimSpace(signature)

	// Check for GPG signature
	if strings.HasPrefix(signature, "-----BEGIN PGP SIGNATURE-----") {
		// Verify if GPG signatures are allowed
		if !r.isSignatureTypeAllowed("gpg") {
			r.addError(
				domain.ValidationErrorDisallowedSigType,
				"GPG signatures are not allowed with current configuration",
				map[string]string{
					"signature_type": "gpg",
					"allowed_types":  strings.Join(r.allowedSigTypes, ", "),
				},
			)

			return r.errors
		}

		if !strings.Contains(signature, "-----END PGP SIGNATURE-----") {
			r.addError(
				domain.ValidationErrorIncompleteGPGSig,
				"incomplete GPG signature (missing end marker)",
				map[string]string{
					"signature_type": "gpg",
				},
			)

			return r.errors
		}

		// Use ProtonMail's openpgp library to validate the format
		block, err := armor.Decode(strings.NewReader(signature))
		if err != nil {
			r.addError(
				domain.ValidationErrorInvalidGPGFormat,
				"invalid GPG signature format: "+err.Error(),
				map[string]string{
					"signature_type": "gpg",
					"error_details":  err.Error(),
				},
			)

			return r.errors
		}

		if block.Type != "PGP SIGNATURE" {
			r.addError(
				domain.ValidationErrorInvalidGPGFormat,
				"invalid GPG armor block type: "+block.Type,
				map[string]string{
					"signature_type": "gpg",
					"block_type":     block.Type,
				},
			)

			return r.errors
		}

		// We don't verify the signature cryptographically, just check it can be read
		reader := packet.NewReader(block.Body)

		_, err = reader.Next()
		if err != nil {
			r.addError(
				domain.ValidationErrorInvalidGPGFormat,
				"malformed GPG signature content: "+err.Error(),
				map[string]string{
					"signature_type": "gpg",
					"error_details":  err.Error(),
				},
			)

			return r.errors
		}

		return r.errors // Valid GPG signature format
	}

	// Check for SSH signature
	if strings.HasPrefix(signature, "-----BEGIN SSH SIGNATURE-----") {
		// Verify if SSH signatures are allowed
		if !r.isSignatureTypeAllowed("ssh") {
			r.addError(
				domain.ValidationErrorDisallowedSigType,
				"SSH signatures are not allowed with current configuration",
				map[string]string{
					"signature_type": "ssh",
					"allowed_types":  strings.Join(r.allowedSigTypes, ", "),
				},
			)

			return r.errors
		}

		// Use SSH-specific validation
		err := r.verifySSHSignatureFormat(signature)
		if err != nil {
			errorCode := domain.ValidationErrorInvalidSSHFormat
			if strings.Contains(err.Error(), "incomplete SSH signature") {
				errorCode = domain.ValidationErrorIncompleteSSHSig
			} else if strings.Contains(err.Error(), "malformed SSH signature") {
				errorCode = domain.ValidationErrorInvalidSSHFormat
			}

			r.addError(
				errorCode,
				err.Error(),
				map[string]string{
					"signature_type": "ssh",
					"error_details":  err.Error(),
				},
			)

			return r.errors
		}

		return r.errors // Valid SSH signature format
	}

	// Not a recognized signature format
	r.addError(
		domain.ValidationErrorUnknownSigFormat,
		"invalid signature format (must be GPG or SSH)",
		map[string]string{
			"signature_prefix": signature[:r.safeMin(len(signature), 20)], // Just capture the first bit for context
		},
	)

	return r.errors
}

// Result returns a concise validation result as a human-readable string.
func (r *SignatureRule) Result() string {
	if len(r.errors) != 0 {
		// Return a concise error message
		return "Missing or invalid signature"
	}

	return "SSH/GPG signature found"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r *SignatureRule) VerboseResult() string {
	if len(r.errors) != 0 {
		// Return a more detailed error message in verbose mode
		errorMsg := r.errors[0].Message
		errorCode := domain.ValidationErrorCode(r.errors[0].Code)
		ctx := r.errors[0].Context
		sigType, hasSigType := ctx["signature_type"]

		// We're deliberately not handling all possible validation error codes here,
		// just the ones that can be generated by this specific rule.
		//nolint:exhaustive
		switch errorCode {
		case domain.ValidationErrorMissingSignature:
			return "Commit does not have a SSH/GPG signature. Sign your commit using 'git commit -S -m \"Message\"'"
		case domain.ValidationErrorIncompleteGPGSig:
			return "Incomplete GPG signature (missing end marker). Try signing again with 'git commit -S -m \"Message\"'"
		case domain.ValidationErrorIncompleteSSHSig:
			return "Incomplete SSH signature (missing end marker). Try signing again with 'git commit -S -m \"Message\"'"
		case domain.ValidationErrorInvalidGPGFormat:
			return "Invalid GPG signature format. Try signing again with 'git commit -S -m \"Message\"'"
		case domain.ValidationErrorInvalidSSHFormat:
			return "Invalid SSH signature format. Try signing again with 'git commit --ssh-sign -m \"Message\"'"
		case domain.ValidationErrorDisallowedSigType:
			var allowedTypes string
			if val, ok := ctx["allowed_types"]; ok {
				allowedTypes = val
			}

			return fmt.Sprintf("Disallowed signature type: %s. Allowed types: %s", sigType, allowedTypes)
		case domain.ValidationErrorUnknownSigFormat:
			return "Invalid signature format (must be GPG or SSH). Check your git configuration."
		default:
			// Fallback to string matching if the error code doesn't match
			if strings.Contains(errorMsg, "does not have a SSH/GPG signature") {
				return "Commit does not have a SSH/GPG signature. Sign your commit using 'git commit -S -m \"Message\"'"
			} else if strings.Contains(errorMsg, "incomplete GPG signature") {
				return "Incomplete GPG signature (missing end marker). Try signing again with 'git commit -S -m \"Message\"'"
			} else if strings.Contains(errorMsg, "incomplete SSH signature") {
				return "Incomplete SSH signature (missing end marker). Try signing again with 'git commit -S -m \"Message\"'"
			} else if strings.Contains(errorMsg, "invalid GPG") || (hasSigType && sigType == "gpg") {
				return "Invalid GPG signature format. Try signing again with 'git commit -S -m \"Message\"'"
			} else if strings.Contains(errorMsg, "invalid SSH") || (hasSigType && sigType == "ssh") {
				return "Invalid SSH signature format. Try signing again with 'git commit --ssh-sign -m \"Message\"'"
			} else if strings.Contains(errorMsg, "not allowed") {
				var allowedTypes string
				if val, ok := ctx["allowed_types"]; ok {
					allowedTypes = val
				}

				if hasSigType {
					return fmt.Sprintf("Disallowed signature type: %s. Allowed types: %s", sigType, allowedTypes)
				}
			} else if strings.Contains(errorMsg, "unknown") || strings.Contains(errorMsg, "invalid signature format") {
				return "Invalid signature format (must be GPG or SSH). Check your git configuration."
			}
		}

		// Default case
		return r.errors[0].Error()
	}

	return "SSH/GPG signature found (format verified only, not cryptographically validated). Run with '--help signature' for more info."
}

// Help returns a description of how to fix the rule violation.
func (r *SignatureRule) Help() string {
	if len(r.errors) == 0 {
		return `No errors to fix
Note: This rule only checks that a signature exists and has valid formatting.
It does NOT verify the cryptographic validity of the signature or that it was 
created by a trusted key. For full security, additional verification is required.`
	}

	// Check error code for targeted help
	if len(r.errors) > 0 {
		errorCode := domain.ValidationErrorCode(r.errors[0].Code)
		errorMsg := r.errors[0].Message

		// We're deliberately not handling all possible validation error codes here,
		// just the ones that can be generated by this specific rule.
		//nolint:exhaustive
		switch errorCode {
		case domain.ValidationErrorMissingSignature:
			return `Sign your commit using either GPG or SSH to verify your identity.

For GPG signing:
- Generate a GPG key: 'gpg --gen-key'
- Configure Git: 'git config --global user.signingkey YOUR_KEY_ID'
- Sign commits: 'git commit -S -m "Message"'

For SSH signing:
- Configure Git: 'git config --global gpg.format ssh'
- Set signing key: 'git config --global user.signingkey ~/.ssh/id_ed25519.pub'
- Sign commits: 'git commit -S -m "Message"'

For convenience, you can enable automatic signing:
'git config --global commit.gpgsign true'`
		case domain.ValidationErrorDisallowedSigType:
			allowedTypes := strings.Join(r.allowedSigTypes, ", ")

			return fmt.Sprintf(`Your signature type is not allowed with the current configuration.
Allowed signature types: %s

To use a different signature type, configure Git accordingly:

For GPG signing:
- 'git config --global gpg.format gpg'
- 'git config --global user.signingkey YOUR_GPG_KEY_ID'

For SSH signing:
- 'git config --global gpg.format ssh'
- 'git config --global user.signingkey PATH_TO_YOUR_SSH_KEY'`, allowedTypes)
		case domain.ValidationErrorIncompleteGPGSig:
			return `Sign your commit correctly. Your GPG signature is incomplete. It's missing the end marker line.
A complete GPG signature must have both:
- A beginning line: "-----BEGIN PGP SIGNATURE-----"
- An ending line: "-----END PGP SIGNATURE-----"

Try signing your commit again with 'git commit -S -m "Message"'`
		case domain.ValidationErrorIncompleteSSHSig:
			return `Sign your commit correctly. Your SSH signature is incomplete. It's missing the end marker line.
A complete SSH signature must have both:
- A beginning line: "-----BEGIN SSH SIGNATURE-----"
- An ending line: "-----END SSH SIGNATURE-----"

Try signing your commit again with 'git commit -S -m "Message"'`
		case domain.ValidationErrorInvalidGPGFormat:
			return `Sign your commit correctly. Your GPG signature has an invalid format.
Make sure your GPG key is properly configured:
- Check that your key is valid: 'gpg --list-secret-keys'
- Ensure Git is configured correctly: 'git config --global user.signingkey YOUR_KEY_ID'
- Try signing a new commit: 'git commit -S -m "Message"'`
		case domain.ValidationErrorInvalidSSHFormat:
			return `Sign your commit correctly. Your SSH signature has an invalid format.
Make sure your SSH key is properly configured:
- Check that your key exists: 'ls ~/.ssh'
- Ensure Git is configured correctly: 'git config --global gpg.format ssh' and 
  'git config --global user.signingkey PATH_TO_YOUR_PUBLIC_KEY'
- Try signing a new commit: 'git commit -S -m "Message"'`
		case domain.ValidationErrorUnknownSigFormat:
			return `Sign your commit correctly. Your commit signature uses an unknown format.
Git supports two signature formats:
- GPG signatures: Starting with "-----BEGIN PGP SIGNATURE-----"
- SSH signatures: Starting with "-----BEGIN SSH SIGNATURE-----"

Make sure you're using one of these formats and that your signature tools are properly configured.`
		default:
			// Fallback to string matching if the error code doesn't match
			if strings.Contains(errorMsg, "does not have a SSH/GPG signature") {
				return `Sign your commit using either GPG or SSH to verify your identity.

For GPG signing:
- Generate a GPG key: 'gpg --gen-key'
- Configure Git: 'git config --global user.signingkey YOUR_KEY_ID'
- Sign commits: 'git commit -S -m "Message"'

For SSH signing:
- Configure Git: 'git config --global gpg.format ssh'
- Set signing key: 'git config --global user.signingkey ~/.ssh/id_ed25519.pub'
- Sign commits: 'git commit -S -m "Message"'

For convenience, you can enable automatic signing:
'git config --global commit.gpgsign true'`
			} else if strings.Contains(errorMsg, "incomplete GPG signature") {
				return `Sign your commit correctly. Your GPG signature is incomplete. It's missing the end marker line.
A complete GPG signature must have both:
- A beginning line: "-----BEGIN PGP SIGNATURE-----"
- An ending line: "-----END PGP SIGNATURE-----"

Try signing your commit again with 'git commit -S -m "Message"'`
			} else if strings.Contains(errorMsg, "invalid GPG") {
				return `Sign your commit correctly. Your GPG signature has an invalid format.
Make sure your GPG key is properly configured:
- Check that your key is valid: 'gpg --list-secret-keys'
- Ensure Git is configured correctly: 'git config --global user.signingkey YOUR_KEY_ID'
- Try signing a new commit: 'git commit -S -m "Message"'`
			} else if strings.Contains(errorMsg, "incomplete SSH signature") {
				return `Sign your commit correctly. Your SSH signature is incomplete. It's missing the end marker line.
A complete SSH signature must have both:
- A beginning line: "-----BEGIN SSH SIGNATURE-----"
- An ending line: "-----END SSH SIGNATURE-----"

Try signing your commit again with 'git commit -S -m "Message"'`
			} else if strings.Contains(errorMsg, "invalid SSH") {
				return `Sign your commit correctly. Your SSH signature has an invalid format.
Make sure your SSH key is properly configured:
- Check that your key exists: 'ls ~/.ssh'
- Ensure Git is configured correctly: 'git config --global gpg.format ssh' and 
  'git config --global user.signingkey PATH_TO_YOUR_PUBLIC_KEY'
- Try signing a new commit: 'git commit -S -m "Message"'`
			} else if strings.Contains(errorMsg, "invalid signature format") {
				return `Sign your commit correctly. Your commit signature uses an unknown format.
Git supports two signature formats:
- GPG signatures: Starting with "-----BEGIN PGP SIGNATURE-----"
- SSH signatures: Starting with "-----BEGIN SSH SIGNATURE-----"

Make sure you're using one of these formats and that your signature tools are properly configured.`
			}
		}
	}

	// Default help
	return `Sign your commit using SSH or GPG to verify your identity.
You can do this by:
1. Setting up GPG signing:
   - Generate a GPG key if you don't have one: 'gpg --gen-key'
   - Configure Git to use your key: 'git config --global user.signingkey YOUR_KEY_ID'
   - Sign commits with: 'git commit -S' or set automatic signing: 'git config --global commit.gpgsign true'

2. Setting up SSH signing:
   - Configure Git to use your SSH key: 'git config --global gpg.format ssh'
   - Set your signing key: 'git config --global user.signingkey ~/.ssh/id_ed25519.pub'
   - Sign commits with: 'git commit -S'

Note: This rule only checks for the existence and basic format of a signature.
For full security, Git or your Git platform should verify signatures cryptographically.

For more information, see:
- GPG signing: https://git-scm.com/book/en/v2/Git-Tools-Signing-Your-Work
- SSH signing: https://docs.github.com/en/authentication/managing-commit-signature-verification/about-commit-signature-verification`
}

// Errors returns any violations of the rule.
func (r *SignatureRule) Errors() []*domain.ValidationError {
	return r.errors
}

// addError adds a structured validation error.
func (r *SignatureRule) addError(code domain.ValidationErrorCode, message string, context map[string]string) {
	err := domain.NewStandardValidationError(r.Name(), code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	r.errors = append(r.errors, err)
}

// isSignatureTypeAllowed checks if a signature type is allowed in the configuration.
func (r *SignatureRule) isSignatureTypeAllowed(sigType string) bool {
	// If no types are specified, all types are allowed
	if len(r.allowedSigTypes) == 0 {
		return true
	}

	for _, allowed := range r.allowedSigTypes {
		if allowed == sigType {
			return true
		}
	}

	return false
}

// safeMin returns the minimum of two integers (utility function for safety).
func (r *SignatureRule) safeMin(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// verifySSHSignatureFormat performs basic validation of SSH signature format
// without attempting to parse the detailed binary structure.
func (r *SignatureRule) verifySSHSignatureFormat(signature string) error {
	if !strings.HasPrefix(signature, "-----BEGIN SSH SIGNATURE-----") {
		return errors.New("missing SSH signature begin marker")
	}

	if !strings.Contains(signature, "-----END SSH SIGNATURE-----") {
		return errors.New("incomplete SSH signature (missing end marker)")
	}

	// Extract the base64-encoded content
	content := signature
	content = strings.TrimPrefix(content, "-----BEGIN SSH SIGNATURE-----")
	content = strings.TrimSuffix(content, "-----END SSH SIGNATURE-----")
	content = strings.TrimSpace(content)

	// Validate the base64 encoding
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return fmt.Errorf("invalid SSH signature encoding: %w", err)
	}

	// Basic structure checks
	// Check for the SSHSIG prefix
	if len(data) < 6 || !bytes.Equal(data[:6], []byte("SSHSIG")) {
		return errors.New("malformed SSH signature: missing SSHSIG prefix")
	}

	// A valid SSH signature should have reasonable length
	if len(data) < 50 {
		return errors.New("malformed SSH signature: too short")
	}

	// We'll avoid detailed binary parsing which can be fragile
	// and just check the overall structure is reasonable
	return nil
}
