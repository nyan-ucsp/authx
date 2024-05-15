package authx

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/nyan-ucsp/authx/models"
	"github.com/nyan-ucsp/authx/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var jwtSecret []byte

type AuthxServer struct {
	Host      string `envconfig:"SERVER_HOST"`
	Port      int    `envconfig:"SERVER_PORT"`
	JwtSecret string `envconfig:"SERVER_JWT_SECRET"`
}

type AuthxDatabase struct {
	Host     string `envconfig:"DB_HOST"`
	Port     int    `envconfig:"DB_PORT"`
	Name     string `envconfig:"DB_NAME"`
	Username string `envconfig:"DB_USERNAME"`
	Password string `envconfig:"DB_PASSWORD"`
}

type AuthxConfig struct {
	Server   AuthxServer
	Database AuthxDatabase
}

type OptionFunc func(*AuthxConfig)

// Default returns an Engine instance.
// Server: Host: "localhost", Port: 8080, JwtSecret: "devxmm@2023Secret
// Database: Host:"localhost",Port: 5432,Name: "postgres",Username: "postgres",Password: "postgres",
func New(opts ...OptionFunc) *AuthxConfig {
	engine := &AuthxConfig{
		Server: AuthxServer{Host: "localhost", Port: 8080, JwtSecret: "devxmm@2023Secret"},
		Database: AuthxDatabase{
			Host:     "localhost",
			Port:     5432,
			Name:     "postgres",
			Username: "postgres",
			Password: "postgres",
		},
	}
	return engine.With(opts...)
}

// With returns a new Engine instance with the provided options.
func (authConfigs *AuthxConfig) With(opts ...OptionFunc) *AuthxConfig {
	for _, opt := range opts {
		opt(authConfigs)
	}
	dsn := "host=" + authConfigs.Database.Host + " user=" + authConfigs.Database.Username + " password=" + authConfigs.Database.Password + " dbname=" + authConfigs.Database.Name + " port=" + strconv.Itoa(authConfigs.Database.Port) + " sslmode=disable"
	jwtSecret = []byte(authConfigs.Server.JwtSecret)
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	err = db.AutoMigrate(&models.User{}, &models.Session{})
	if err != nil {
		panic("failed to migrate database")
	}
	return authConfigs
}

func GetPostgresDB() *gorm.DB {
	return db
}
func GetJWTSecret() []byte {
	return jwtSecret
}

func GenerateJWT(userID uint, duration time.Duration) (string, time.Time, error) {
	// Define claims
	expiredTime := time.Now().UTC().Add(duration)
	claims := jwt.MapClaims{
		"sub": userID,                  // Subject (user identifier)
		"exp": expiredTime.Unix(),      // Expiration time (set based on duration)
		"iat": time.Now().UTC().Unix(), // Issued at
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", expiredTime, fmt.Errorf("error signing token: %w", err)
	}
	return tokenString, expiredTime, nil
}

// ------------Business Logic-------------------
// !------------[User]-----------------------
func Register(id uint, email *string, phone *string, password string) *models.User {
	pwd, err := utils.HashPassword(password)
	if err != nil {
		println(err.Error())
		return nil
	}
	newUser := models.User{
		ID:        id,
		UserName:  utils.GenerateUsername(3),
		Email:     email,
		Phone:     phone,
		Password:  string(pwd),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	result := db.Create(&newUser)
	if result.Error != nil {
		return nil
	}
	return &newUser
}

func EmailLogin(email string, password string) bool {
	var user *models.User
	db.Find(&models.User{Email: &email}).First(&user)
	if user == nil {
		return false
	}
	if utils.CheckPassword([]byte(user.Password), password) {
		return true
	} else {
		return false
	}

}

func PhoneLogin(phone string, password string) bool {
	var user *models.User
	db.Find(&models.User{Phone: &phone}).First(&user)
	if user == nil {
		return false
	}
	if utils.CheckPassword([]byte(user.Password), password) {
		return true
	} else {
		return false
	}
}

// !------------[Session]-----------------------
// Claims defines the expected structure of the JWT payload
type Claims struct {
	jwt.StandardClaims
	// Add any custom claims you use in your JWT here
}

func IsValidToken(tokenString string) (bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return false, errors.New("invalid token signature")
		}
		return false, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return false, errors.New("invalid token claims")
	}

	// Check for expiration
	if time.Now().UTC().Unix() > claims.ExpiresAt {
		return false, errors.New("token expired")
	}

	return true, nil
}

func AddSession(userId uint, refreshToken string, expiredAt time.Time) *models.Session {
	newSession := models.Session{
		UserId:       userId,
		RefreshToken: refreshToken,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		ExpiredAt:    expiredAt.UTC(),
	}
	result := db.Create(&newSession)
	if result.Error != nil {
		return nil
	}
	return &newSession
}

func RefreshSession(refreshToken string, duration time.Duration) *models.Session {
	var session *models.Session
	db.Find(&models.Session{RefreshToken: refreshToken}).First(&session)
	if session == nil {
		return nil
	}
	if session.ExpiredAt.Compare(time.Now().UTC()) == 1 {
		return nil
	}
	newRefreshToken, expiredTime, err := GenerateJWT(session.UserId, duration)
	if err != nil {
		return nil
	}
	session.RefreshToken = newRefreshToken
	session.ExpiredAt = expiredTime
	session.UpdatedAt = time.Now().UTC()
	result := db.Updates(&session)
	if result.Error != nil {
		return nil
	}
	return session
}

func Logout(refreshToken string) bool {
	result := db.Delete(&models.Session{RefreshToken: refreshToken})
	return result.Error == nil
}
