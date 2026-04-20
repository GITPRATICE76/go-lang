package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("my_secret_key")

// ================= REQUEST =================
type LoginRequest struct {
	Username string `json:"Username"`
	Password string `json:"password"`
}

// ================= ADS =================
type ADSRequest struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

type ADSResponse struct {
	Status  bool   `json:"Status"`
	Message string `json:"Message"`
}

// ================= ADS RESULT =================
type ADSResult struct {
	Success bool
	Message string
	Raw     string
}

// ================= ADS CALL =================
func ValidateWithADS(username, password string) (ADSResult, error) {

	url := "https://csplads.brnetsaas.com:6445/CSADServices/v1/BRNetConnect/ads_getuserdetailswithauthentication"

	payload := ADSRequest{
		Username: username,
		Password: password,
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return ADSResult{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}

	fmt.Println("👉 Calling ADS API...")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ ADS API CALL FAILED:", err)
		return ADSResult{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	raw := string(body)

	fmt.Println("📦 ADS RESPONSE:", raw)

	var adsResp ADSResponse
	json.Unmarshal(body, &adsResp)

	if adsResp.Status {
		return ADSResult{
			Success: true,
			Message: adsResp.Message,
			Raw:     raw,
		}, nil
	}

	return ADSResult{
		Success: false,
		Message: adsResp.Message,
		Raw:     raw,
	}, nil
}

// ================= LOGIN =================
func Login(c *gin.Context) {
	var req LoginRequest

	fmt.Println("====================================")
	fmt.Println("👉 LOGIN API HIT")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	fmt.Println("👉 Username:", req.Username)

	// ✅ STEP 1: ADS AUTH
	adsResult, err := ValidateWithADS(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "ADS call failed",
			"ads_error": err.Error(),
		})
		return
	}

	if !adsResult.Success {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":     "Invalid credentials",
			"ads_message": adsResult.Message,
			"ads_raw":     adsResult.Raw,
		})
		return
	}

	fmt.Println("✅ ADS VERIFIED")

	// ✅ STEP 2: DB CONNECT
	db, err := ConnectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "DB connection failed"})
		return
	}
	defer db.Close()

	// ✅ IMPORTANT: ADD DOMAIN
	emailPattern := req.Username + "@craftsilicon.com"

	query := `
		SELECT id, name, role, department, team
		FROM users
		WHERE email LIKE @Email
	`

	row := db.QueryRow(query, sql.Named("Email", emailPattern))

	var (
		id         int
		name       string
		role       string
		department string
		team       *string
	)

	err = row.Scan(&id, &name, &role, &department, &team)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ USER NOT FOUND")

			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "User not found in system",
				"ads_raw": adsResult.Raw,
			})
			return
		}

		fmt.Println("❌ DB ERROR:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}

	fmt.Println("✅ USER FOUND:", name)

	// ✅ STEP 3: JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         id,
		"name":       name,
		"role":       role,
		"department": department,
		"team":       team,
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Token generation failed"})
		return
	}

	// ✅ RESPONSE
	c.JSON(http.StatusOK, gin.H{
		"token":      tokenString,
		"id":         id,
		"name":       name,
		"role":       role,
		"department": department,
		"team":       team,
		"ads_raw":    adsResult.Raw,
	})
}
