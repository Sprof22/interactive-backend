package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"interactive-path/initializers"
	"interactive-path/models"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"golang.org/x/oauth2"
)

const (
	AUTH0_AUDIENCE = "https://hello-world.example.com"
	AUTH0_DOMAIN   = "codefigma.us.auth0.com"
)

func ValidateToken(tokenStr string) (jwt.Token, error) {
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", AUTH0_DOMAIN)

	keySet, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(
		[]byte(tokenStr),
		jwt.WithKeySet(keySet),
		jwt.WithValidate(true),
	)

	if err != nil {
		return nil, err
	}

	// Check audience
	aud := token.Audience()
	if len(aud) == 0 || aud[0] != AUTH0_AUDIENCE {
		return nil, fmt.Errorf("invalid audience")
	}

	return token, nil
}

func CreateUser(c *gin.Context) {

	accessToken := c.GetHeader("Authorization")
	fmt.Println("here is the access Token: ", accessToken)
	if accessToken == "" {
		c.JSON(400, gin.H{
			"error": "Access token not provided",
		})
		return
	}

	// Assuming your JWT token is in the format "Bearer <token>"
	parts := strings.Split(accessToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(400, gin.H{
			"error": "Invalid token format",
		})
		return
	}

	token := parts[1]

	t, err := ValidateToken(token)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid token: " + err.Error(),
		})
		return
	}

	sub, _ := t.Get("sub")
	aud, _ := t.Get("aud")
	iat, _ := t.Get("iat")
	exp, _ := t.Get("exp")
	azp, _ := t.Get("azp")
	scope, _ := t.Get("scope")

	fmt.Println("SUB: ", sub)
	fmt.Println("Token: ", aud)
	fmt.Println("Token: ", iat)
	fmt.Println("Token: ", exp)
	fmt.Println("Token: ", azp)
	fmt.Println("Token: ", scope)

	// Check if a user with the same Sub already exists
	var existingUser models.User
	result := initializers.DB.Where("sub = ?", sub).First(&existingUser)

	if result.RowsAffected > 0 {
		c.JSON(400, gin.H{
			"error": "User with the same Sub already exists",
		})
		return
	}

	var body struct {
		Email    string
		Name     string
		Picture  string
		Nickname string
	}

	c.Bind(&body)
	user := models.User{Email: body.Email, Name: body.Name, Sub: sub.(string), Picture: body.Picture, Nickname: body.Nickname}

	// result := initializers.DB.Create(&user)
	result = initializers.DB.Create(&user)

	if result.Error != nil {
		c.JSON(400, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"user": user,
	})
}
func UpdateUser(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	fmt.Println("here is the access Token: ", accessToken)
	if accessToken == "" {
		c.JSON(400, gin.H{
			"error": "Access token not provided",
		})
		return
	}

	// Assuming your JWT token is in the format "Bearer <token>"
	parts := strings.Split(accessToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(400, gin.H{
			"error": "Invalid token format",
		})
		return
	}

	token := parts[1]

	t, err := ValidateToken(token)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid token: " + err.Error(),
		})
		return
	}

	sub, _ := t.Get("sub")

	// Bind the request body to a struct
	var body struct {
		GithubToken string
	}

	c.Bind(&body)

	// Update the user's GithubAccessToken
	var user models.User
	result := initializers.DB.Where("sub = ?", sub).First(&user)

	if result.Error != nil {
		c.JSON(400, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	user.GithubToken = body.GithubToken
	result = initializers.DB.Save(&user)

	if result.Error != nil {
		c.JSON(400, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"user": user,
	})
}
func CreateProject(c *gin.Context) {
	// Get the access token from the header
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(400, gin.H{
			"error": "Access token not provided",
		})
		return
	}

	// Assuming your JWT token is in the format "Bearer <token>"
	parts := strings.Split(accessToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(400, gin.H{
			"error": "Invalid token format",
		})
		return
	}

	token := parts[1]

	// Validate the token
	t, err := ValidateToken(token)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid token: " + err.Error(),
		})
		return
	}

	// Get the sub (user identifier) from the token
	sub, _ := t.Get("sub")
	var existingUser models.User
	initializers.DB.Where("sub = ?", sub).First(&existingUser)

	// Log the GitHub token
	fmt.Println("GitHub Token:", existingUser.GithubToken)

	// Bind the request body to a struct
	var body struct {
		Title       string
		Description string
	}

	c.Bind(&body)

	// Create a new project
	project := models.Project{
		Title:       body.Title,
		Description: body.Description,
		UserID:      sub.(string),
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: existingUser.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repo := &github.Repository{
		Name:    github.String(project.Title),
		Private: github.Bool(true),
	}
	rep, res, err := client.Repositories.Create(ctx, "", repo)

	// // Save the project to the database
	// result := initializers.DB.Save(&project)
	// if result.Error != nil {
	// 	c.JSON(400, gin.H{
	// 		"error": result.Error.Error(),
	// 	})
	// 	return
	// }
	if err != nil {
		fmt.Println("failed gere")
		fmt.Println("failed here", err)
	} else {
		fmt.Println(rep, "hey", res)

		// Save the project to the database
		result := initializers.DB.Save(&project)
		if result.Error != nil {
			c.JSON(400, gin.H{
				"error": result.Error.Error(),
			})
			return
		}

		// Return the project to the client
		c.JSON(200, project)
	}
}

func FetchUserProjects(c *gin.Context) {
	// Get the user's sub from the token
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(400, gin.H{
			"error": "Access token not provided",
		})
		return
	}

	parts := strings.Split(accessToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(400, gin.H{
			"error": "Invalid token format",
		})
		return
	}

	token := parts[1]

	t, err := ValidateToken(token)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid token: " + err.Error(),
		})
		return
	}

	sub, _ := t.Get("sub")

	// Query the database for projects belonging to the user
	var projects []models.Project
	result := initializers.DB.Where("user_id = ?", sub).Find(&projects)

	if result.Error != nil {
		c.JSON(400, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"projects": projects,
	})
}

func FetchAllProjects(c *gin.Context) {
	var projects []models.Project

	// Query the database to get all projects
	result := initializers.DB.Find(&projects)

	if result.Error != nil {
		c.JSON(400, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"projects": projects,
	})
}

func DeleteProject(c *gin.Context) {
	// Get the project ID from the route parameter
	id := c.Param("id")

	// Delete the project from the database
	result := initializers.DB.Delete(&models.Project{}, "id = ?", id)

	if result.Error != nil {
		c.JSON(400, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	// Return a successful response
	c.JSON(204, nil)
}

func UpdateProject(c *gin.Context) {
	// Get the access token from the header
	accessToken := c.GetHeader("Authorization")
	if accessToken == "" {
		c.JSON(400, gin.H{
			"error": "Access token not provided",
		})
		return
	}

	// Assuming your JWT token is in the format "Bearer <token>"
	parts := strings.Split(accessToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(400, gin.H{
			"error": "Invalid token format",
		})
		return
	}

	token := parts[1]

	// Validate the token
	t, err := ValidateToken(token)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid token: " + err.Error(),
		})
		return
	}

	// Get the sub (user identifier) from the token
	sub, _ := t.Get("sub")

	// Bind the request body to a struct
	var body struct {
		Title       string
		Description string
	}

	c.Bind(&body)
	fmt.Println("Title:", body.Title)
	fmt.Println("Description:", body.Description)
	fmt.Println("Description:", body)
	// Create a new project with the provided information
	project := models.Project{
		Title:       body.Title,
		Description: body.Description,
		UserID:      sub.(string), // Use the sub as the user identifier
	}

	// Save the project to the database
	result := initializers.DB.Create(&project)

	if result.Error != nil {
		c.JSON(400, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"project": project,
	})
}

func ExtractAccessToken(responseBody string) (string, error) {
	// Check if the response body contains the access token.
	if !strings.Contains(responseBody, "access_token=") {
		return "", errors.New("response body does not contain access token")
	}

	// Decode the response body into a map.
	var bodyMap map[string]string
	err := json.NewDecoder(strings.NewReader(responseBody)).Decode(&bodyMap)
	if err != nil {
		return "", err
	}

	// Get the access token from the map.
	accessToken := bodyMap["access_token"]

	// Return the access token.
	return accessToken, nil
}

func GetGithubAccessTokenGin(c *gin.Context) {
	// Get the client ID and client secret from the environment variables
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURL := os.Getenv("GITHUB_REDIRECT_URI")

	// Get the code from the request body
	code := c.PostForm("code")

	// Prepare the request body
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURL)

	// Encode the request body as application/x-www-form-urlencoded
	encodedBody := data.Encode()

	log.Println("Encoded Body:", encodedBody)

	// Create a new POST request
	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(encodedBody))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Close the response body
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	responseBody := string(body)
	fmt.Println("Response Body:", responseBody)

	accessToken, err := ExtractAccessToken(strings.TrimSpace(responseBody))
	fmt.Println("accessToken: ", accessToken)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Return the access token to the client
	c.JSON(200, gin.H{"access_token": accessToken})
}

// func ExchangeToken(c *gin.Context) {

// 	// Get the access token from the header
// 	accessToken := c.GetHeader("Authorization")
// 	if accessToken == "" {
// 		c.JSON(400, gin.H{
// 			"error": "Access token not provided",
// 		})
// 		return
// 	}

// 	// Assuming your JWT token is in the format "Bearer <token>"
// 	parts := strings.Split(accessToken, " ")
// 	if len(parts) != 2 || parts[0] != "Bearer" {
// 		c.JSON(400, gin.H{
// 			"error": "Invalid token format",
// 		})
// 		return
// 	}

// 	token := parts[1]

// 	// Validate the token
// 	_, err := ValidateToken(token)
// 	if err != nil {
// 		c.JSON(400, gin.H{
// 			"error": "Invalid token: " + err.Error(),
// 		})
// 		return
// 	}

// 	// Get the sub (user identifier) from the token
// 	// sub, _ := t.Get("sub")

// 	var body struct {
// 		Code string
// 	}

// 	c.Bind(&body)

// 	// Create a new HTTP request.
// 	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Set the request headers.
// 	req.Header.Set("Content-Type", "application/json")

// 	// Encode the request body as JSON.
// 	requestBody, err := json.Marshal(map[string]string{
// 		"redirect_uri":  os.Getenv("GITHUB_REDIRECT_URI"),
// 		"code":          body.Code,
// 		"client_id":     os.Getenv("GITHUB_CLIENT_ID"),
// 		"client_secret": os.Getenv("GITHUB_SECRET_KEY"),
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode request body: " + err.Error()})
// 		return
// 	}

// 	// func (r *bytes.Reader) Close() error {
// 	// 	return nil
// 	// }

// 	// Create a new bytes.Reader struct.
// 	reader := bytes.NewReader(requestBody)

// 	// Set the request body.
// 	req.Body = reader

// 	// Send the request and get the response.
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	defer resp.Body.Close()

// 	// Check the response status code.
// 	if resp.StatusCode != http.StatusOK {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange GitHub code: " + resp.Status})
// 		return
// 	}

// 	// Decode the response body.
// 	var responseBody struct {
// 		AccessToken string `json:"access_token"`
// 	}

// 	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode GitHub response: " + err.Error()})
// 		return
// 	}

// 	// Return the access token to the client.
// 	c.JSON(http.StatusOK, gin.H{
// 		"access_token": responseBody.AccessToken,
// 	})
// }
