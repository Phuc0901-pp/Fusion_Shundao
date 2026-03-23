package delivery

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"fusion/internal/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = func() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "shundao-solar-secret-2026"
	}
	return []byte(s)
}()

const (
	loginAllowedStartHour = 4
	loginAllowedEndHour   = 18
	loginAllowedEndMinute = 30
	tokenRefreshExpiry    = 15 * time.Minute
	maxFailedAttempts     = 5
)

// ── Exponential Backoff Brute-force Tracker ──────────────────────────────────
// Lockout durations: 5m → 10m → 30m → 90m → 270m → ...
// After the predefined steps, duration multiplies by 3 each time.
var lockoutSteps = []time.Duration{
	5 * time.Minute,
	10 * time.Minute,
	30 * time.Minute,
	90 * time.Minute,
}

type ipRecord struct {
	failCount    int       // failures within current window
	lockoutCount int       // how many times this IP has been locked
	lockedUntil  time.Time // zero means not locked
}

var (
	ipRecords   = make(map[string]*ipRecord)
	ipRecordsMu sync.Mutex
)

func getRecord(ip string) *ipRecord {
	// must be called with ipRecordsMu held
	rec, ok := ipRecords[ip]
	if !ok {
		rec = &ipRecord{}
		ipRecords[ip] = rec
	}
	return rec
}

// lockoutDuration calculates the lockout duration for the Nth lockout (1-indexed).
func lockoutDuration(n int) time.Duration {
	if n <= len(lockoutSteps) {
		return lockoutSteps[n-1]
	}
	// After predefined steps, multiply last step by 3^(extra)
	extra := n - len(lockoutSteps)
	d := lockoutSteps[len(lockoutSteps)-1]
	for i := 0; i < extra; i++ {
		d *= 3
	}
	return d
}

// checkAndLock checks if the IP is currently locked.
// Returns (locked bool, unlocksAt time.Time, remaining int).
func checkAndLock(ip string) (bool, time.Time) {
	ipRecordsMu.Lock()
	defer ipRecordsMu.Unlock()
	rec := getRecord(ip)
	if !rec.lockedUntil.IsZero() && time.Now().Before(rec.lockedUntil) {
		return true, rec.lockedUntil
	}
	// Auto-unlock: reset fail count every time lockout expires
	if !rec.lockedUntil.IsZero() && time.Now().After(rec.lockedUntil) {
		rec.failCount = 0
		rec.lockedUntil = time.Time{}
	}
	return false, time.Time{}
}

// recordFail increments failure and triggers lockout if threshold reached.
// Returns (triggered418 bool, unlocksAt time.Time, remaining int).
func recordFail(ip string) (bool, time.Time, int) {
	ipRecordsMu.Lock()
	defer ipRecordsMu.Unlock()
	rec := getRecord(ip)
	rec.failCount++
	if rec.failCount >= maxFailedAttempts {
		rec.lockoutCount++
		d := lockoutDuration(rec.lockoutCount)
		rec.lockedUntil = time.Now().Add(d)
		rec.failCount = 0
		return true, rec.lockedUntil, 0
	}
	remaining := maxFailedAttempts - rec.failCount
	return false, time.Time{}, remaining
}

func resetFail(ip string) {
	ipRecordsMu.Lock()
	defer ipRecordsMu.Unlock()
	if rec, ok := ipRecords[ip]; ok {
		rec.failCount = 0
		// keep lockoutCount so next lockout is still longer
	}
}

type jwtClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// parseAndValidateToken parses JWT from "Authorization: Bearer <token>"
func parseAndValidateToken(r *http.Request) (*jwtClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return nil, jwt.ErrSignatureInvalid
	}
	tokenStr := authHeader[7:]
	claims := &jwtClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	return claims, err
}

// computeExpiry returns the sooner of (now+15min) or (18:30:00 today)
func computeExpiry() time.Time {
	return time.Now().Add(tokenRefreshExpiry)
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// isWithinOperatingHours checks if now is in [04:00, 18:30)
func isWithinOperatingHours() bool {
	return true // Disabled outside-working-hours restriction
}

// ── Request / Response types ─────────────────────────────────────────────────
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
	FullName string `json:"full_name"`
	Message  string `json:"message"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ── HandleAuthLogin – POST /api/auth/login ────────────────────────────────────
func HandleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// Time-Lock Check
	if !isWithinOperatingHours() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Hệ thống chỉ hoạt động từ 04:00 đến 18:30. Vui lòng quay lại trong giờ làm việc.",
		})
		return
	}

	ip := getClientIP(r)

	// Brute-Force / Exponential Lockout Check
	if locked, unlocksAt := checkAndLock(ip); locked {
		w.WriteHeader(http.StatusTeapot) // 418
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":     "Truy cập tạm thời bị khóa do nhập sai quá nhiều lần.",
			"hacked":    true,
			"unlock_at": unlocksAt.Unix(), // Unix timestamp for frontend countdown
		})
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}
	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Tên đăng nhập và mật khẩu không được để trống"})
		return
	}

	// DB Lookup
	var account database.Account
	result := database.DB.Where("username = ?", req.Username).First(&account)
	if result.Error != nil {
		triggered, unlocksAt, remaining := recordFail(ip)
		if triggered {
			w.WriteHeader(http.StatusTeapot)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":     "Quá nhiều lần thất bại. Tài khoản tạm khóa.",
				"hacked":    true,
				"unlock_at": unlocksAt.Unix(),
			})
			return
		}
		if result.Error == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":     "Tên đăng nhập hoặc mật khẩu không đúng",
				"remaining": remaining,
			})
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Password Verify
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)); err != nil {
		triggered, unlocksAt, remaining := recordFail(ip)
		if triggered {
			w.WriteHeader(http.StatusTeapot)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":     "Quá nhiều lần thất bại. Tài khoản tạm khóa.",
				"hacked":    true,
				"unlock_at": unlocksAt.Unix(),
			})
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":     "Tên đăng nhập hoặc mật khẩu không đúng",
			"remaining": remaining,
		})
		return
	}

	// Reset fail counter on success
	resetFail(ip)

	// Generate JWT (expire at 18:30 absolute / 15 min, whichever is sooner)
	expiry := computeExpiry()
	claims := jwtClaims{
		UserID:   account.ID,
		Username: account.Username,
		Role:     account.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "shundao-solar",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	database.DB.Model(&account).Update("last_login_at", now)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginResponse{
		Token:    tokenString,
		Username: account.Username,
		Role:     account.Role,
		FullName: account.FullName,
		Message:  "Đăng nhập thành công",
	})
}

// ── HandleAuthRefresh – POST /api/auth/refresh ───────────────────────────────
func HandleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// Bot-check header guard
	if r.Header.Get("X-Shundao-Bot-Check") == "" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Bot check failed"})
		return
	}

	claims, err := parseAndValidateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Token không hợp lệ hoặc đã hết hạn"})
		return
	}
	if !isWithinOperatingHours() {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ngoài giờ hoạt động."})
		return
	}

	expiry := computeExpiry()
	newClaims := jwtClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "shundao-solar",
		},
	}
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenStr, err := newToken.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to refresh token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": newTokenStr})
}

// ── HandleAuthChangePassword – POST /api/auth/change-password ────────────────
func HandleAuthChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	claims, err := parseAndValidateToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Chưa đăng nhập hoặc token hết hạn"})
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OldPassword == "" || req.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Dữ liệu không hợp lệ"})
		return
	}

	var account database.Account
	if result := database.DB.Where("username = ?", claims.Username).First(&account); result.Error != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Tài khoản không tồn tại"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.OldPassword)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Mật khẩu cũ không đúng"})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Hash error", http.StatusInternalServerError)
		return
	}

	database.DB.Model(&account).Update("password_hash", string(newHash))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Đổi mật khẩu thành công"})
}
