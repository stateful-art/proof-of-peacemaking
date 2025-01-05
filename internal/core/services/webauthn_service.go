package services

import (
	"context"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

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
	wconfig := &webauthn.Config{
		RPDisplayName: "Proof of Peacemaking",
		RPID:          "localhost", // Change this for production
		RPOrigins:     []string{"http://localhost:3003"},
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
		credentials = append(credentials, webauthn.Credential{
			ID:              cred.CredentialID,
			PublicKey:       cred.PublicKey,
			AttestationType: "",                                  // Not storing attestation type
			Transport:       []protocol.AuthenticatorTransport{}, // Not storing transport
			Flags:           webauthn.CredentialFlags{},          // Not storing flags
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

	// Configure registration options to prefer cross-platform authenticators
	options, session, err := s.webauthn.BeginRegistration(
		webAuthnUser,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.CrossPlatform,         // Prefer cross-platform authenticators (like password managers)
			ResidentKey:             protocol.ResidentKeyPreferred,  // Prefer resident keys but don't require them
			UserVerification:        protocol.VerificationPreferred, // Prefer user verification
		}),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation), // Don't need attestation
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

	options, session, err := s.webauthn.BeginLogin(webAuthnUser)
	if err != nil {
		return nil, webauthn.SessionData{}, fmt.Errorf("failed to begin authentication: %w", err)
	}

	return options, *session, nil
}

// FinishAuthentication completes the passkey authentication process
func (s *WebAuthnService) FinishAuthentication(ctx context.Context, userID primitive.ObjectID, sessionData webauthn.SessionData, response *protocol.ParsedCredentialAssertionData) error {
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

	credential, err := s.webauthn.ValidateLogin(webAuthnUser, sessionData, response)
	if err != nil {
		return fmt.Errorf("failed to finish authentication: %w", err)
	}

	// Update the credential's sign count
	existingCred, err := s.passkeyRepository.GetCredentialByCredentialID(ctx, credential.ID)
	if err != nil {
		return fmt.Errorf("failed to get credential: %w", err)
	}
	if existingCred == nil {
		return fmt.Errorf("credential not found")
	}

	if err := s.passkeyRepository.UpdateCredentialSignCount(ctx, existingCred.ID, credential.Authenticator.SignCount); err != nil {
		return fmt.Errorf("failed to update sign count: %w", err)
	}

	// Find the user-passkey relationship and update last used
	for _, up := range userPasskeys {
		if up.CredentialID == existingCred.ID {
			deviceInfo := "" // Get device info from the client
			if err := s.passkeyRepository.UpdateUserPasskeyLastUsed(ctx, up.ID, deviceInfo); err != nil {
				return fmt.Errorf("failed to update last used: %w", err)
			}
			break
		}
	}

	return nil
}
