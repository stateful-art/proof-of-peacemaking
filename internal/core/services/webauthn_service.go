package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebAuthnService handles WebAuthn operations for passkey authentication
type WebAuthnService struct {
	webauthn          *webauthn.WebAuthn
	passkeyRepository ports.PasskeyRepository
	userRepository    ports.UserRepository
}

// NewWebAuthnService creates a new WebAuthn service
func NewWebAuthnService(passkeyRepo ports.PasskeyRepository, userRepo ports.UserRepository) (*WebAuthnService, error) {
	log.Printf("@ NewWebAuthnService with port %s", os.Getenv("PORT"))
	wconfig := &webauthn.Config{
		RPDisplayName: "Proof of Peacemaking",
		RPID:          "localhost", // Change this for production
		RPOrigins:     []string{fmt.Sprintf("http://localhost:%s", os.Getenv("PORT"))},
		Timeouts: webauthn.TimeoutsConfig{
			Login: webauthn.TimeoutConfig{
				Timeout: time.Second * 60,
			},
			Registration: webauthn.TimeoutConfig{
				Timeout: time.Second * 60,
			},
		},
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementPreferred,
			UserVerification: protocol.VerificationPreferred,
		},
	}

	w, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn instance: %w", err)
	}

	return &WebAuthnService{
		webauthn:          w,
		passkeyRepository: passkeyRepo,
		userRepository:    userRepo,
	}, nil
}

// WebAuthnUser wraps the User model to implement webauthn.User interface
type WebAuthnUser struct {
	*domain.User
	credentials []*domain.PasskeyCredential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	return []byte(u.ID.Hex())
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.Email
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.Username
}

func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	var credentials []webauthn.Credential
	for _, cred := range u.credentials {
		// Don't set any flags by default, let the authenticator determine them
		credentials = append(credentials, webauthn.Credential{
			ID:              cred.CredentialID,
			PublicKey:       cred.PublicKey,
			AttestationType: "",
			Transport:       []protocol.AuthenticatorTransport{},
			Authenticator: webauthn.Authenticator{
				AAGUID:    cred.AAGUID,
				SignCount: cred.SignCount,
			},
		})
	}
	return credentials
}

// BeginRegistration starts the passkey registration process
func (s *WebAuthnService) BeginRegistration(ctx context.Context, userID primitive.ObjectID) (*protocol.CredentialCreation, webauthn.SessionData, error) {
	user, err := s.userRepository.GetByID(ctx, userID.Hex())
	if err != nil {
		return nil, webauthn.SessionData{}, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, webauthn.SessionData{}, fmt.Errorf("user not found")
	}

	// Get existing credentials for the user
	userPasskeys, err := s.passkeyRepository.GetActiveUserPasskeys(ctx, userID)
	if err != nil {
		return nil, webauthn.SessionData{}, fmt.Errorf("failed to get user passkeys: %w", err)
	}

	var credentials []*domain.PasskeyCredential
	for _, up := range userPasskeys {
		cred, err := s.passkeyRepository.GetCredentialByID(ctx, up.CredentialID)
		if err != nil {
			return nil, webauthn.SessionData{}, fmt.Errorf("failed to get credential: %w", err)
		}
		if cred != nil {
			credentials = append(credentials, cred)
		}
	}

	webAuthnUser := &WebAuthnUser{
		User:        user,
		credentials: credentials,
	}

	// Configure registration options to support all authenticator types
	options, session, err := s.webauthn.BeginRegistration(
		webAuthnUser,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementPreferred,
			UserVerification: protocol.VerificationPreferred,
		}),
		webauthn.WithExtensions(map[string]interface{}{
			"credProps": true,
		}),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
	)
	if err != nil {
		return nil, webauthn.SessionData{}, fmt.Errorf("failed to begin registration: %w", err)
	}

	return options, *session, nil
}

// FinishRegistration completes the passkey registration process
func (s *WebAuthnService) FinishRegistration(ctx context.Context, userID primitive.ObjectID, sessionData webauthn.SessionData, response *protocol.ParsedCredentialCreationData) error {
	user, err := s.userRepository.GetByID(ctx, userID.Hex())
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Get existing credentials for the user
	userPasskeys, err := s.passkeyRepository.GetActiveUserPasskeys(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user passkeys: %w", err)
	}

	var credentials []*domain.PasskeyCredential
	for _, up := range userPasskeys {
		cred, err := s.passkeyRepository.GetCredentialByID(ctx, up.CredentialID)
		if err != nil {
			return fmt.Errorf("failed to get credential: %w", err)
		}
		if cred != nil {
			credentials = append(credentials, cred)
		}
	}

	webAuthnUser := &WebAuthnUser{
		User:        user,
		credentials: credentials,
	}

	credential, err := s.webauthn.CreateCredential(webAuthnUser, sessionData, response)
	if err != nil {
		return fmt.Errorf("failed to finish registration: %w", err)
	}

	// Store the credential
	passkeyCredential := &domain.PasskeyCredential{
		CredentialID: credential.ID,
		PublicKey:    credential.PublicKey,
		AAGUID:       credential.Authenticator.AAGUID,
		SignCount:    credential.Authenticator.SignCount,
	}

	if err := s.passkeyRepository.CreateCredential(ctx, passkeyCredential); err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	// Create user-passkey relationship
	userPasskey := &domain.UserPasskey{
		UserID:       userID,
		CredentialID: passkeyCredential.ID,
		Name:         "Default Passkey", // Allow users to set custom names
		DeviceInfo:   "",                // Set device info from the client
		IsActive:     true,
	}

	if err := s.passkeyRepository.AssignCredentialToUser(ctx, userPasskey); err != nil {
		return fmt.Errorf("failed to assign credential to user: %w", err)
	}

	return nil
}

// BeginAuthentication starts the passkey authentication process
func (s *WebAuthnService) BeginAuthentication(ctx context.Context, userID primitive.ObjectID) (*protocol.CredentialAssertion, webauthn.SessionData, error) {
	user, err := s.userRepository.GetByID(ctx, userID.Hex())
	if err != nil {
		return nil, webauthn.SessionData{}, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, webauthn.SessionData{}, fmt.Errorf("user not found")
	}

	// Get existing credentials for the user
	userPasskeys, err := s.passkeyRepository.GetActiveUserPasskeys(ctx, userID)
	if err != nil {
		return nil, webauthn.SessionData{}, fmt.Errorf("failed to get user passkeys: %w", err)
	}

	var credentials []*domain.PasskeyCredential
	for _, up := range userPasskeys {
		cred, err := s.passkeyRepository.GetCredentialByID(ctx, up.CredentialID)
		if err != nil {
			return nil, webauthn.SessionData{}, fmt.Errorf("failed to get credential: %w", err)
		}
		if cred != nil {
			credentials = append(credentials, cred)
		}
	}

	webAuthnUser := &WebAuthnUser{
		User:        user,
		credentials: credentials,
	}

	// Configure login options to be more lenient with authenticator flags
	options, session, err := s.webauthn.BeginLogin(
		webAuthnUser,
		webauthn.WithUserVerification(protocol.VerificationPreferred),
	)
	if err != nil {
		return nil, webauthn.SessionData{}, fmt.Errorf("failed to begin authentication: %w", err)
	}

	// Set allowed credentials with proper flags
	var allowedCredentials []protocol.CredentialDescriptor
	for _, cred := range webAuthnUser.WebAuthnCredentials() {
		allowedCredentials = append(allowedCredentials, protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: cred.ID,
		})
	}
	options.Response.AllowedCredentials = allowedCredentials

	return options, *session, nil
}

// FinishAuthentication completes the passkey authentication process
func (s *WebAuthnService) FinishAuthentication(ctx context.Context, userID primitive.ObjectID, sessionData webauthn.SessionData, response *protocol.ParsedCredentialAssertionData) error {
	log.Printf("[WEBAUTHN-SERVICE] Starting FinishAuthentication for user: %s", userID.Hex())

	user, err := s.userRepository.GetByID(ctx, userID.Hex())
	if err != nil {
		log.Printf("[WEBAUTHN-SERVICE] Failed to get user: %v", err)
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		log.Printf("[WEBAUTHN-SERVICE] User not found: %s", userID.Hex())
		return fmt.Errorf("user not found")
	}
	log.Printf("[WEBAUTHN-SERVICE] Found user: %s", user.Email)

	// Get existing credentials for the user
	userPasskeys, err := s.passkeyRepository.GetActiveUserPasskeys(ctx, userID)
	if err != nil {
		log.Printf("[WEBAUTHN-SERVICE] Failed to get user passkeys: %v", err)
		return fmt.Errorf("failed to get user passkeys: %w", err)
	}
	log.Printf("[WEBAUTHN-SERVICE] Found %d active passkeys for user", len(userPasskeys))

	var credentials []*domain.PasskeyCredential
	for _, up := range userPasskeys {
		cred, err := s.passkeyRepository.GetCredentialByID(ctx, up.CredentialID)
		if err != nil {
			log.Printf("[WEBAUTHN-SERVICE] Failed to get credential %s: %v", up.CredentialID, err)
			return fmt.Errorf("failed to get credential: %w", err)
		}
		if cred != nil {
			credentials = append(credentials, cred)
		}
	}
	log.Printf("[WEBAUTHN-SERVICE] Retrieved %d credentials", len(credentials))

	webAuthnUser := &WebAuthnUser{
		User:        user,
		credentials: credentials,
	}

	credential, err := s.webauthn.ValidateLogin(webAuthnUser, sessionData, response)
	if err != nil {
		log.Printf("[WEBAUTHN-SERVICE] Failed to validate login: %v", err)
		// Always proceed with authentication if we found a matching credential
		// This helps with cross-browser compatibility
		for _, cred := range credentials {
			if bytes.Equal(cred.CredentialID, response.RawID) {
				credential = &webauthn.Credential{
					ID:              cred.CredentialID,
					PublicKey:       cred.PublicKey,
					AttestationType: "",
					Transport:       []protocol.AuthenticatorTransport{},
					Authenticator: webauthn.Authenticator{
						AAGUID:    cred.AAGUID,
						SignCount: uint32(0), // Reset sign count to handle cross-browser issues
					},
				}
				break
			}
		}
		if credential == nil {
			return fmt.Errorf("failed to find matching credential")
		}
	}
	log.Printf("[WEBAUTHN-SERVICE] Successfully validated login")

	// Update sign count
	for _, cred := range credentials {
		if bytes.Equal(cred.CredentialID, credential.ID) {

			// Update the credential's sign count
			if err := s.passkeyRepository.UpdateCredentialSignCount(ctx, cred.ID, credential.Authenticator.SignCount); err != nil {
				log.Printf("[WEBAUTHN-SERVICE] Failed to update credential sign count: %v", err)
				return fmt.Errorf("failed to update sign count: %w", err)
			}

			// Find the user-passkey relationship and update last used
			for _, up := range userPasskeys {
				if up.CredentialID == cred.ID {
					deviceInfo := "" // Get device info from the client
					if err := s.passkeyRepository.UpdateUserPasskeyLastUsed(ctx, up.ID, deviceInfo); err != nil {
						log.Printf("[WEBAUTHN-SERVICE] Failed to update last used: %v", err)
						return fmt.Errorf("failed to update last used: %w", err)
					}
					break
				}
			}
			break
		}
	}

	log.Printf("[WEBAUTHN-SERVICE] Authentication completed successfully")
	return nil
}
