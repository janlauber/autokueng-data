package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// SecretKey is the secret key used to sign the JWT token
var SecretKey string
var UploadSecret string
var URL string
var corsString string

func init() {
	corsString = os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsString == "" {
		corsString = "*"
		// log
		log.Default().Println("CORS_ALLOWED_ORIGINS is not set, allowing all origins")
	}
}

func main() {
	SecretKey = os.Getenv("JWT_SECRET_KEY")
	if SecretKey == "" {
		panic("JWT_SECRET_KEY is not set, must be the same as the one used in the backend api")
	}

	URL = os.Getenv("URL")
	if URL == "" {
		URL = "http://localhost:9000"
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     corsString,
	}))

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Static("/images", "./images")

	app.Post("/upload", handleImageUpload)

	app.Delete("/images/:imageName", handleImageDelete)

	log.Fatal(app.Listen(":9000"))

}

func handleImageUpload(c *fiber.Ctx) error {

	// Validate the token

	token, err := CheckAuth(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	if !token.Valid {
		return c.Status(401).JSON(fiber.Map{
			"status":  401,
			"message": "invalid token",
		})
	}

	// Get IP address of the client
	ip := c.IP()
	file, err := c.FormFile("image")

	if err != nil {
		log.Println("image upload error: ", err)
		return c.Status(500).SendString("image upload error")
	}

	// convert size to MB and round to 2 decimal places
	fileSizeMB := float64(file.Size) / 1024 / 1024

	// error if file size is greater than 2MB
	if fileSizeMB > 2 {
		return c.Status(400).JSON(fiber.Map{
			"status":  400,
			"message": "file size is greater than 2MB",
		})
	}

	// error if file type is not png or jpg or jpeg or gif
	fileType := strings.Split(file.Filename, ".")
	if fileType[len(fileType)-1] != "png" && fileType[len(fileType)-1] != "jpg" && fileType[len(fileType)-1] != "jpeg" && fileType[len(fileType)-1] != "gif" {
		return c.Status(400).JSON(fiber.Map{
			"status":  400,
			"message": "file type is not ending on .png or .jpg or .jpeg",
		})
	}

	uniqueId := uuid.New()
	filename := strings.Replace(uniqueId.String(), "-", "-", -1)
	fileExt := strings.Split(file.Filename, ".")[len(strings.Split(file.Filename, "."))-1]
	image := fmt.Sprintf("%s.%s", filename, fileExt)
	err = c.SaveFile(file, fmt.Sprintf("./images/%s", image))

	if err != nil {
		log.Println("image upload error: ", err)
		return c.Status(500).SendString("image upload error")
	}

	imageUrl := fmt.Sprintf("%s/images/%s", URL, image)

	imageData := map[string]interface{}{
		"imageName": image,
		"imageUrl":  imageUrl,
		"header":    file.Header,
		"size":      file.Size,
	}

	if fileSizeMB > 1 {
		log.Printf("%s uploaded %s with %.2f MB", ip, image, fileSizeMB)
	} else {
		log.Printf("%s uploaded %s with %.2f KB", ip, image, fileSizeMB*1024)
	}

	return c.Status(201).JSON(fiber.Map{"data": imageData})
}

func handleImageDelete(c *fiber.Ctx) error {
	// Validate the token

	token, err := CheckAuth(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	if !token.Valid {
		return c.Status(401).JSON(fiber.Map{
			"status":  401,
			"message": "invalid token",
		})
	}

	// Get IP address of the client
	ip := c.IP()

	imageName := c.Params("imageName")
	err = os.Remove(fmt.Sprintf("./images/%s", imageName))

	if err != nil {
		log.Println("image delete error: ", err)
		return c.Status(500).SendString("image delete error")
	}
	log.Printf("%s deleted %s", ip, imageName)
	return c.JSON(fiber.Map{"status": 200, "message": "image deleted successfully"})
}

func CheckAuth(c *fiber.Ctx) (*jwt.Token, error) {
	var token *jwt.Token
	var tokenString string
	var err error

	bearerToken := c.Get("Authorization")

	bearerTokenSplit := strings.Split(bearerToken, " ")
	if len(bearerTokenSplit) != 2 {
		return token, errors.New("Invalid token")
	} else {
		tokenString = bearerTokenSplit[1]
	}

	if len(tokenString) == 0 {
		return token, errors.New("Invalid bearer token")
	}

	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

	if err != nil {
		return token, errors.New("Invalid token")
	}

	if token == nil || !token.Valid {
		return token, errors.New("Invalid token")
	}

	return token, nil
}
