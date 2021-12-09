package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
)

func main() {
	JWT_SECRET_KEY := os.Getenv("SECRET_KEY")
	if JWT_SECRET_KEY == "" {
		panic("SECRET_KEY is not set")
	}

	app := fiber.New()
	app.Use(cors.New())
	app.Static("/images", "./images")

	app.Post("/upload", handleImageUpload)

	app.Delete("/delete/:imageName", handleImageDelete)

	log.Fatal(app.Listen(":9000"))

}

func handleImageUpload(c *fiber.Ctx) error {
	file, err := c.FormFile("image")

	if err != nil {
		log.Println("image upload error: ", err)
		return c.Status(500).SendString("image upload error")
	}

	uniqueId := uuid.New()
	filename := strings.Replace(uniqueId.String(), "-", "-", -1)
	fileExt := strings.Split(file.Filename, ".")[1]
	image := fmt.Sprintf("%s.%s", filename, fileExt)
	err = c.SaveFile(file, fmt.Sprintf("./images/%s", image))

	if err != nil {
		log.Println("image upload error: ", err)
		return c.Status(500).SendString("image upload error")
	}

	imageUrl := fmt.Sprintf("http://localhost:9000/images/%s", image)

	data := map[string]interface{}{
		"imageName": image,
		"imageUrl":  imageUrl,
		"header":    file.Header,
		"size":      file.Size,
	}

	return c.JSON(fiber.Map{"status": 201, "message": "image uploaded successfully", "data": data})
}

func handleImageDelete(c *fiber.Ctx) error {
	imageName := c.Params("imageName")
	err := os.Remove(fmt.Sprintf("./images/%s", imageName))

	if err != nil {
		log.Println("image delete error: ", err)
		return c.Status(500).SendString("image delete error")
	}

	return c.JSON(fiber.Map{"status": 200, "message": "image deleted successfully"})
}
