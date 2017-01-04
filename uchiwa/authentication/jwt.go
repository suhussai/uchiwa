package authentication

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/mitchellh/mapstructure"
	"github.com/sensu/uchiwa/uchiwa/structs"
	log "github.com/Sirupsen/logrus"
)

// JWTToken constant
const JWTToken = "jwtToken"

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

// GetJWTFromContext retrieves the JWT Token from the request
func GetJWTFromContext(r *http.Request) *jwt.Token {
	if value := context.Get(r, JWTToken); value != nil {
		return value.(*jwt.Token)
	}
	return nil
}

// GetRoleFromToken ...
func GetRoleFromToken(token *jwt.Token) (*Role, error) {
	r, ok := token.Claims["Role"]
	if !ok {
		return &Role{}, errors.New("Could not retrieve the user Role from the JWT")
	}
	var role Role
	err := mapstructure.Decode(r, &role)
	if err != nil {
		return &Role{}, err
	}
	return &role, nil
}

// GetToken returns a string that contain the token
func GetToken(role *Role, username string) (string, error) {
	if username == "" {
		return "", errors.New("Could not generate a token for the user. Invalid username")
	}

	t := jwt.New(jwt.GetSigningMethod("RS256"))
	t.Claims["Role"] = role
	t.Claims["Username"] = username

	if privateKey == nil {
		return "", errors.New("Could not generate a token for the user. Invalid private key")
	}

	tokenString, err := t.SignedString(privateKey)
	return tokenString, err
}

// generateKeyPair generates an RSA keypair of 2048 bits using a random rand.Reader
func generateKeyPair() *rsa.PrivateKey {
	keypair, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Could not generate an RSA keypair.")
	}

	return keypair
}

// generateToken generates a private and public RSA keys
// in order to be used for the JWT signature
func generateToken() (*rsa.PrivateKey, *rsa.PublicKey) {
	log.Debug("Generating new temporary RSA keys")
	privateKey := generateKeyPair()
	// Precompute some calculations
	privateKey.Precompute()
	publicKey := &privateKey.PublicKey

	return privateKey, publicKey
}

// initToken initializes the token by weither loading the keys from the
// filesystem with the loadToken() function or by generating temporarily
// ones with the generateToken() function
func initToken(a structs.Auth) {
	var err error
	privateKey, publicKey, err = loadToken(a)
	if err != nil {
		// At this point we need to generate temporary RSA keys
		log.Debug(err)
		privateKey, publicKey = generateToken()
	}
}

// loadToken loads a private and public RSA keys from the filesystem
// in order to be used for the JWT signature
func loadToken(a structs.Auth) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	log.Debug("Attempting to load the RSA keys from the filesystem")

	if a.PrivateKey == "" || a.PublicKey == "" {
		return nil, nil, errors.New("The paths to the private and public RSA keys were not provided")
	}

	// Read the files from the filesystem
	prv, err := ioutil.ReadFile(a.PrivateKey)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Unable to open the private key file.")
	}
	pub, err := ioutil.ReadFile(a.PublicKey)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Unable to open the public key file.")
	}

	// Parse the RSA keys
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(prv)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Unable to parse the private key file.")
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pub)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Unable to parse the public key.")
	}

	log.Info("Provided RSA keys successfully loaded")
	return privateKey, publicKey, nil
}

// setJWTIntoContext injects the JWT Token into the request for later use
func setJWTInContext(r *http.Request, token *jwt.Token) {
	context.Set(r, JWTToken, token)
}

// verifyJWT extracts and verifies the validity of the JWT
func verifyJWT(r *http.Request) (*jwt.Token, error) {
	token, err := jwt.ParseFromRequest(r, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			log.WithFields(log.Fields{
				"method": t.Header["alg"],
			}).Debug("Unexpected signing method.")
			return nil, errors.New("")
		}
		return publicKey, nil
	})

	if token == nil || err != nil {
		return nil, errors.New("")
	}

	if !token.Valid {
		log.Debug("Invalid JWT")
		return nil, errors.New("")
	}

	return token, nil
}
