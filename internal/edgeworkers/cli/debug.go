package cli

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	edgeworkersapi "github.com/akamai/edgeworkers-cli/internal/edgeworkers/api"
	"github.com/spf13/cobra"
)

// jwtClaims holds the debug token claims needed for display; the token itself is opaque to the CLI.
type jwtClaims struct {
	ACL []string `json:"acl"`
	Exp int64    `json:"exp"`
}

// decodeJWTClaims extracts claims from a JWT's payload segment without verifying its signature.
func decodeJWTClaims(token string) (jwtClaims, error) {
	var claims jwtClaims
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claims, fmt.Errorf("token is not a valid JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return claims, fmt.Errorf("decode token payload: %w", err)
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return claims, fmt.Errorf("parse token claims: %w", err)
	}
	return claims, nil
}

// authTokenResult is the CLI-facing view of a debug auth token, combining the token with claims decoded from it.
type authTokenResult struct {
	Hostnames []string `json:"hostnames,omitempty"`
	Expiry    string   `json:"expiry,omitempty"`
	Token     string   `json:"token"`
}

// generateSecret returns a random 32-byte hex string for use as a debug secret key.
func generateSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// newDebugCmd returns the debug command group.
func newDebugCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "debug", Short: "Security and debug tools"}
	cmd.AddCommand(newDebugSecretCmd(), newDebugAuthTokenCmd())
	return cmd
}

func newDebugSecretCmd() *cobra.Command {
	return &cobra.Command{
		Use: "secret", Short: "Generate a random secret key for PMUSER_EW_DEBUG_KEY",
		RunE: func(cmd *cobra.Command, args []string) error {
			secret := generateSecret()
			return outputData(map[string]string{"secret": secret}, jsonOutput(cmd), func() {
				fmt.Fprintln(os.Stdout, secret)
			})
		},
	}
}

func newDebugAuthTokenCmd() *cobra.Command {
	var hostnameList string
	var expiry int
	cmd := &cobra.Command{
		Use: "auth-token [hostname]", Short: "Generate an EdgeWorker debug auth token", Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var hostnames []string
			if len(args) > 0 {
				hostnames = append(hostnames, strings.Split(args[0], ",")...)
			}
			if hostnameList != "" {
				hostnames = append(hostnames, strings.Split(hostnameList, ",")...)
			}
			request := edgeworkersapi.SecureTokenRequest{}
			if len(hostnames) > 0 {
				request.Hostnames = &hostnames
			}
			if expiry > 0 {
				expiryValue := float64(expiry)
				request.Expiry = &expiryValue
			}
			productConfig, err := configStore.LoadEffectiveProduct(cmd, product)
			if err != nil {
				return handleErr("connect to Akamai API", err, "")
			}
			debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
			c, err := edgeworkersapi.NewClient(productConfig.EdgeRc, productConfig.Section, productConfig.AccountSwitchKey, time.Duration(productConfig.TimeoutSeconds)*time.Second, debug)
			if err != nil {
				return handleErr("connect to Akamai API", err, "")
			}
			raw, err := c.CreateSecureToken(cmd.Context(), request)
			if err != nil {

				return handleErr("generate debug auth token", err, "")

			}
			var result edgeworkersapi.AuthToken
			if err := decodeAPIResult(raw, &result); err != nil {
				return err
			}
			output := authTokenResult{Token: result.Token}
			if claims, err := decodeJWTClaims(result.Token); err == nil {
				output.Hostnames = claims.ACL
				output.Expiry = time.Unix(claims.Exp, 0).UTC().Format(time.RFC3339)
			}
			return outputData(output, jsonOutput(cmd), func() {
				fmt.Fprintf(os.Stdout, "Hostname: %s\nExpiry: %s\nAkamai-EW-Trace: %s\n",
					strings.Join(output.Hostnames, ","), output.Expiry, output.Token)
			})
		},
	}
	cmd.Flags().StringVar(&hostnameList, "acl", "", "additional comma-separated hostnames to include in the token; use \"/*\" to allow all hosts")
	cmd.Flags().IntVar(&expiry, "expiry", 0, "token expiry in minutes, from 1 to 720 (defaults to the API's 480-minute default)")
	return cmd
}
