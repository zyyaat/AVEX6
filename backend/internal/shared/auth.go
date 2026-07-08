package shared

import (
	"context"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims - shared JWT claims for all 4 user types
type Claims struct {
	UserID       string `json:"user_id"`
	Phone        string `json:"phone"`
	Name         string `json:"name"`
	Admin        bool   `json:"admin"`
	IsDriver     bool   `json:"is_driver"`
	DriverID     string `json:"driver_id,omitempty"`
	IsMerchant   bool   `json:"is_merchant"`
	MerchantID   string `json:"merchant_id,omitempty"`
	RestaurantID string `json:"restaurant_id,omitempty"`
	IsAgent      bool   `json:"is_agent"`
	AgentID      string `json:"agent_id,omitempty"`
	jwt.RegisteredClaims
}

func HashPassword(p string) (string, error) {
	b, e := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), e
}

func CheckPassword(p, h string) bool {
	return bcrypt.CompareHashAndPassword([]byte(h), []byte(p)) == nil
}

func jwtSecret() string {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		log.Fatal("JWT_SECRET environment variable is required but not set")
	}
	return s
}

func GenerateJWT(uid, phone, name string, admin bool) (string, error) {
	claims := &Claims{
		UserID: uid, Phone: phone, Name: name, Admin: admin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSecret()))
}

func GenerateDriverJWT(driverID, phone, name string) (string, error) {
	claims := &Claims{
		UserID: driverID, Phone: phone, Name: name, IsDriver: true, DriverID: driverID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSecret()))
}

func GenerateMerchantJWT(merchantID, restaurantID, phone, name string) (string, error) {
	claims := &Claims{
		UserID: merchantID, Phone: phone, Name: name,
		IsMerchant: true, MerchantID: merchantID, RestaurantID: restaurantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSecret()))
}

func GenerateAgentJWT(agentID, phone, name string) (string, error) {
	claims := &Claims{
		UserID: agentID, Phone: phone, Name: name,
		IsAgent: true, AgentID: agentID, Admin: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSecret()))
}

func VerifyJWT(ts string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(ts, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret()), nil
	})
	return claims, err
}

// Phone helpers
func CleanPhone(p string) string {
	v := regexp.MustCompile(`[^\d+]`).ReplaceAllString(p, "")
	if strings.HasPrefix(v, "+20") {
		return "0" + v[3:]
	}
	if len(v) == 13 && strings.HasPrefix(v, "20") && v[2] == '1' {
		return "0" + v[2:]
	}
	if strings.HasPrefix(v, "+") {
		v = strings.ReplaceAll(v, "+", "")
	}
	return regexp.MustCompile(`[^0-9]`).ReplaceAllString(v, "")
}

func ValidPhone(p string) bool {
	matched, _ := regexp.MatchString(`^01[0125][0-9]{8}$`, p)
	return matched
}

// Context helpers
func ContextWithUser(r *http.Request, c *Claims) context.Context {
	return context.WithValue(r.Context(), "user", c)
}

func GetUser(r *http.Request) *Claims {
	if c, ok := r.Context().Value("user").(*Claims); ok {
		return c
	}
	return nil
}

// Convenience aliases
func GetDriver(r *http.Request) *Claims   { return GetUser(r) }
func GetMerchant(r *http.Request) *Claims { return GetUser(r) }
func GetAgent(r *http.Request) *Claims    { return GetUser(r) }
