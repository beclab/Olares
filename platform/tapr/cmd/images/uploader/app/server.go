package app

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"bytetrade.io/web3os/tapr/pkg/app/middleware"
	"bytetrade.io/web3os/tapr/pkg/constants"
	"bytetrade.io/web3os/tapr/pkg/images"
	"bytetrade.io/web3os/tapr/pkg/kubesphere"
	"bytetrade.io/web3os/tapr/pkg/terminus"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type Server struct {
	KubeConfig *rest.Config
}

func (s *Server) ServerRun() {
	// create new fiber instance  and use across whole app
	app := fiber.New()

	// middleware to allow all clients to communicate using http and allow cors
	app.Use(cors.New())

	app.Post("/images/upload/v1", middleware.RequireAuth(s.KubeConfig, s.handleFileupload))

	// delete uploaded image by providing unique image name
	// app.Delete("/:imageName", handleDeleteImage)

	klog.Fatal(app.Listen(":8080"))
}

func (s *Server) handleFileupload(c *fiber.Ctx) error {

	// parse incomming image file
	file, err := c.FormFile("image")
	policy := c.FormValue("policy", "private")

	if err != nil {
		klog.Error("image upload error --> ", err)
		return c.JSON(fiber.Map{"code": http.StatusInternalServerError, "message": "Upload image error", "data": nil})

	}

	// get image writer
	imageWriter := func() images.ImageWriter {
		fileImageWriter := func(file *multipart.FileHeader) (string, error) {
			// generate new uuid for image name
			uniqueId := uuid.New()
			filename := strings.Replace(uniqueId.String(), "-", "", -1)

			filenameToken := strings.Split(file.Filename, ".")
			if len(filenameToken) < 2 {
				return "", errors.New("just uploading image file is allowed")
			}

			fileExt := filenameToken[1]
			if !images.IsImage(fileExt) {
				return "", errors.New("just uploading image file is allowed")
			}

			// generate image from filename and extension
			image := fmt.Sprintf("%s.%s", filename, fileExt)

			// save image to upload dir
			uploadDir, err := ensureUploadPath()
			if err != nil {
				return "", err
			}
			err = c.SaveFile(file, fmt.Sprintf("%s/%s", uploadDir, image))

			if err != nil {
				return "", err
			}

			token := c.Context().UserValueBytes([]byte(constants.AuthorizationTokenKey))
			user := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey))
			userZone, err := kubesphere.GetUserZone(context.TODO(), s.KubeConfig, user.(string))
			if err != nil {
				return "", err
			}

			uri := fmt.Sprintf("/resources/Home/Pictures/Upload/%s", image)
			domain := fmt.Sprintf("files.%s", userZone)
			imageUrl := fmt.Sprintf("https://%s%s", domain, uri)

			if policy == "public" {
				err := terminus.UpdatePolicy(context.TODO(), uri, "olares-app", "files", token.(string), policy)
				if err != nil {
					return "", err
				}
			}

			return imageUrl, nil
		}

		return images.ImageWriterFunc(fileImageWriter)
	}()

	imageUrl, err := imageWriter.Write(file)
	if err != nil {
		klog.Error("image save error --> ", err)
		return c.JSON(fiber.Map{"code": http.StatusInternalServerError, "message": "Save image error", "data": nil})
	}

	data := map[string]interface{}{
		"imageUrl": imageUrl,
		"size":     file.Size,
	}

	klog.Info("success to upload a new image, ", imageUrl)

	return c.JSON(fiber.Map{"code": http.StatusOK, "message": "Image uploaded successfully", "data": data})
}

// func handleDeleteImage(c *fiber.Ctx) error {
// 	// extract image name from params
// 	imageName := c.Params("imageName")

// 	// delete image from ./images
// 	err := os.Remove(fmt.Sprintf("./images/%s", imageName))
// 	if err != nil {
// 		log.Println(err)
// 		return c.JSON(fiber.Map{"code": 500, "message": "Server Error", "data": nil})
// 	}

// 	return c.JSON(fiber.Map{"code": 201, "message": "Image deleted successfully", "data": nil})
// }

func ensureUploadPath() (string, error) {
	rootPath := "/data"
	uploadPath := rootPath + "/Upload"

	mkdir := func() (string, error) {
		err := os.Mkdir(uploadPath, 0755)
		if err != nil {
			return "", err
		}

		return uploadPath, nil
	}

	dir, err := os.Stat(uploadPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mkdir()
		}

		return "", err
	}

	if !dir.IsDir() {
		err := os.Remove(uploadPath)
		if err != nil {
			return "", err
		}

		return mkdir()
	}

	return uploadPath, nil
}
