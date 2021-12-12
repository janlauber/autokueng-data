package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
)

// SecretKey is the secret key used to sign the JWT token
var SecretKey string
var UploadSecret string

func main() {
	SecretKey = os.Getenv("JWT_SECRET_KEY")
	if SecretKey == "" {
		panic("JWT_SECRET_KEY is not set, must be the same as the one used in the backend api")
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	app.Static("/images", "./images")

	app.Post("/upload", handleImageUpload)

	app.Delete("/delete/:imageName", handleImageDelete)

	log.Fatal(app.Listen(":9000"))

}

func validateTokenData(token *jwt.Token) {
	claims := token.Claims.(jwt.MapClaims)

	data := claims["data"].(map[string]interface{})
	uploadSecret := data["uploadSecret"].(string)

	if uploadSecret != UploadSecret {
		token.Valid = false
	}
	token.Valid = true
}

func handleImageUpload(c *fiber.Ctx) error {

	cookie := c.Cookies("jwt-autokueng-api")
	if cookie == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Parse the token and validate it
	token, _ := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

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

	// error if file type is not png or jpg
	fileType := strings.Split(file.Filename, ".")
	if fileType[len(fileType)-1] != "png" && fileType[len(fileType)-1] != "jpg" {
		return c.Status(400).JSON(fiber.Map{
			"status":  400,
			"message": "file type is not ending on .png or .jpg",
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

	imageUrl := fmt.Sprintf("http://localhost:9000/images/%s", image)

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

	cookie := c.Cookies("jwt-autokueng-api")
	if cookie == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Parse the token
	token, _ := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	// Check if the token is valid
	validateTokenData(token)

	if !token.Valid {
		return c.Status(401).JSON(fiber.Map{
			"status":  401,
			"message": "invalid token",
		})
	}

	// Get IP address of the client
	ip := c.IP()

	imageName := c.Params("imageName")
	err := os.Remove(fmt.Sprintf("./images/%s", imageName))

	if err != nil {
		log.Println("image delete error: ", err)
		return c.Status(500).SendString("image delete error")
	}
	log.Printf("%s deleted %s", ip, imageName)
	return c.JSON(fiber.Map{"status": 200, "message": "image deleted successfully"})
}
