package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
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

	resp, err := client.Do(req)
	if err != nil {
		return ADSResult{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var adsResp ADSResponse
	json.Unmarshal(body, &adsResp)

	if adsResp.Status {
		return ADSResult{Success: true, Message: adsResp.Message}, nil
	}

	return ADSResult{Success: false, Message: adsResp.Message}, nil
}

// ================= LOGIN =================
func Login(c *gin.Context) {

	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	input := strings.TrimSpace(req.Username)

	// ✅ STEP 1: ADS AUTH
	adsResult, err := ValidateWithADS(input, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "ADS call failed"})
		return
	}

	if !adsResult.Success {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		return
	}


	db, err := ConnectDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "DB connection failed"})
		return
	}
	defer db.Close()

	// ✅ STEP 3: FLEXIBLE MATCHING (FIXED)
	query := `
		SELECT TOP 1 id, name, role, department, team
		FROM users
		WHERE 
			LOWER(email) = LOWER(@input)
			OR LOWER(LEFT(email, CHARINDEX('@', email) - 1)) LIKE LOWER(@input + '%')
	`

	row := db.QueryRow(query, sql.Named("input", input))

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
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "User not found in system",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error",
		})
		return
	}

	
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


	c.JSON(http.StatusOK, gin.H{
		"token":      tokenString,
		"id":         id,
		"name":       name,
		"role":       role,
		"department": department,
		"team":       team,
	})
}