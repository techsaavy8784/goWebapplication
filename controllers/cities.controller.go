package controllers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"hyperpage/initializers"
	"hyperpage/models"
)

// GetCitiesResponse represents the response structure for the GetCities function.
type GetCitiesResponse struct {
	Status string        `json:"status"`
	Data   []models.City `json:"data"`
	Meta   GetCitiesMeta `json:"meta"`
}

// GetCitiesMeta represents the metadata for the paginated response.
type GetCitiesMeta struct {
	Limit string `json:"limit"`
	Skip  string `json:"skip"`
	Total int64  `json:"total"`
}

// GetCities gets a paginated list of cities with translations.
// @Summary Get a list of cities.
// @Description Retrieves a paginated list of city names along with their translations.
// @Tags Cities
// @Accept json
// @Produce json
// @Param limit query string false "Number of cities to retrieve (default: 10)"
// @Param skip query string false "Number of cities to skip (default: 0)"
// @Success 200 {object} GetCitiesResponse
// @Router /cities/all [get]
func GetCities(c *fiber.Ctx) error {
	// Get query parameters for pagination
	limit := c.Query("limit", "10")
	skip := c.Query("skip", "0") // Use skip directly from the query parameters

	limitNumber, err := strconv.Atoi(limit)
	if err != nil || limitNumber < 1 {
		limitNumber = 10
	}

	skipNumber, err := strconv.Atoi(skip)
	if err != nil || skipNumber < 0 {
		skipNumber = 0
	}

	// get count of all cities in the database
	var total int64
	if err := initializers.DB.Model(&models.City{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch cities count from the database",
		})
	}

	// get paginated city names from the database with translations
	var cities []models.City
	db := initializers.DB.
		Joins("JOIN city_translations ON cities.id = city_translations.city_id").
		Preload("Translations").
		Select("DISTINCT cities.id, cities.hex, cities.updated_at, cities.deleted_at").
		Offset(skipNumber).Limit(limitNumber).
		Order("cities.id").
		Find(&cities)

	if db.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch paginated cities from the database",
			"error":   db.Error.Error(),
		})
	}

	// fmt.Println(db.Statement.SQL.String()) // Log the generated SQL

	// return the paginated city names and metadata as a JSON response
	return c.JSON(GetCitiesResponse{
		Status: "success",
		Data:   cities,
		Meta: GetCitiesMeta{
			Limit: strconv.Itoa(limitNumber),
			Skip:  strconv.Itoa(skipNumber),
			Total: total,
		},
	})
}

func GetName(c *fiber.Ctx) error {
	// Get the name query parameter
	// Get the name and lang query parameters
	name := c.Query("name")
	lang := c.Query("lang")
	mode := c.Query("mode")

	// Get query parameters for pagination
	limit := c.Query("limit", "10")
	skip := c.Query("skip", "0")

	limitNumber, err := strconv.Atoi(limit)
	if err != nil || limitNumber < 1 {
		limitNumber = 10
	}

	skipNumber, err := strconv.Atoi(skip)
	if err != nil || skipNumber < 0 {
		skipNumber = 0
	}

	if mode == "translate" {
		var cityTranslation models.CityTranslation
		if err := initializers.DB.
			Where("name ILIKE ?", "%"+name+"%").
			First(&cityTranslation).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to fetch city translation from the database",
			})
		}

		var translatedCity models.CityTranslation
		if err := initializers.DB.
			Where("city_id = ? AND language = ?", cityTranslation.CityID, lang). // Replace "EN" with your target language code
			First(&translatedCity).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to fetch translated city name from the database",
			})
		}

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   translatedCity.Name,
		})

	}
	if name == "" || lang == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Both 'name' and 'lang' parameters are required",
		})
	}

	var cities []models.City
	if err := initializers.DB.
		Joins("JOIN city_translations ON cities.id = city_translations.city_id").
		Preload("Translations", "language = ?", lang).
		Where("city_translations.name ILIKE ? AND city_translations.language = ?", "%"+name+"%", lang).
		Offset(skipNumber).Limit(limitNumber).
		Find(&cities).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch cities from the database",
		})
	}

	// Get total count of matching cities
	var total int64
	if err := initializers.DB.
		Model(&models.City{}).
		Joins("JOIN city_translations ON cities.id = city_translations.city_id").
		Where("city_translations.name ILIKE ? AND city_translations.language = ?", "%"+name+"%", lang).
		Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch total count from the database",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   cities,
		"meta": fiber.Map{
			"limit": limitNumber,
			"skip":  skipNumber,
			"total": total,
		},
	})
}

func CreateCity(c *fiber.Ctx) error {
	var newCity models.City
	if err := c.BodyParser(&newCity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	newCity.UpdatedAt = time.Now()

	if err := initializers.DB.Create(&newCity).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to add the new city",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "New city added successfully",
		"data":    newCity,
	})
}

func DeleteCity(c *fiber.Ctx) error {
	cityID := c.Params("id")

	if cityID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "City ID is required",
		})
	}

	var city models.City
	if err := initializers.DB.First(&city, cityID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "City not found",
		})
	}

	// Find and delete all city translations associated with the city
	if err := initializers.DB.Where("city_id = ?", city.ID).Delete(&models.CityTranslation{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete city translations",
		})
	}

	if err := initializers.DB.Delete(&city).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete the city",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "City and associated translations deleted successfully",
		"data":    nil,
	})
}

func UpdateCity(c *fiber.Ctx) error {

	cityID := c.Params("id")

	if cityID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "City ID is required",
		})
	}

	var city models.City
	if err := initializers.DB.First(&city, cityID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "City not found",
		})
	}

	var updatedCity models.City
	if err := c.BodyParser(&updatedCity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	if err := initializers.DB.Model(&city).Updates(&updatedCity).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update the city",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "City updated successfully",
		"data":    city,
	})
}

func CreateCityTranslation(c *fiber.Ctx) error {
	var newTranslation models.CityTranslation

	if err := c.BodyParser(&newTranslation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	// Check if CityID is provided
	if newTranslation.CityID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "City ID is required",
		})
	}

	// You may want to add additional validation or checks here

	// Use newTranslation.CityID to retrieve the city from the database
	var city models.City
	if err := initializers.DB.First(&city, newTranslation.CityID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "City not found",
		})
	}

	// Assign the retrieved City ID to the translation
	newTranslation.CityID = city.ID

	// Add the new translation to the database
	if err := initializers.DB.Create(&newTranslation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to add the new translation",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "New translation added successfully",
		"data":    newTranslation,
	})
}

func GetCityTranslation(c *fiber.Ctx) error {
	translationID := c.Query("translationID")

	if translationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Translation ID is required",
		})
	}

	var translation models.CityTranslation
	if err := initializers.DB.First(&translation, translationID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Translation not found",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   translation,
	})
}

func UpdateCityTranslation(c *fiber.Ctx) error {
	translationID := c.Query("translationID")

	if translationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Translation ID is required",
		})
	}

	var translation models.CityTranslation
	if err := initializers.DB.First(&translation, translationID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Translation not found",
		})
	}

	var updatedTranslation models.CityTranslation
	if err := c.BodyParser(&updatedTranslation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request data",
		})
	}

	if err := initializers.DB.Model(&translation).Updates(&updatedTranslation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update the translation",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Translation updated successfully",
		"data":    translation,
	})
}

func DeleteCityTranslation(c *fiber.Ctx) error {
	translationID := c.Query("translationID")

	if translationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Translation ID is required",
		})
	}

	var translation models.CityTranslation
	if err := initializers.DB.First(&translation, translationID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Translation not found",
		})
	}

	if err := initializers.DB.Delete(&translation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete the translation",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Translation deleted successfully",
		"data":    nil,
	})
}
