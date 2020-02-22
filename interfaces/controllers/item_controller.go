package controllers

import (
	"net/http"

	"github.com/akaishi-sandbox/sam-go/infrastructure"
	"github.com/akaishi-sandbox/sam-go/interfaces/database"
	"github.com/akaishi-sandbox/sam-go/usecase"
	"github.com/labstack/echo"
)

// ItemController struct
type ItemController struct {
	Interactor usecase.ItemInteractor
}

// NewItemController instance
func NewItemController(elasticHandler *infrastructure.ElasticHandler) *ItemController {
	return &ItemController{
		Interactor: usecase.ItemInteractor{
			ItemRepository: &database.ItemRepository{
				ElasticHandler: elasticHandler,
			},
		},
	}
}

func (controller *ItemController) queryStringParameters(c echo.Context) map[string]string {
	parameters := make(map[string]string, len(c.ParamNames()))

	for _, name := range c.ParamNames() {
		parameters[name] = c.Param(name)
	}

	return parameters
}

// Search function
func (controller *ItemController) Search(c echo.Context) (err error) {
	searchResult, err := controller.Interactor.Search(controller.queryStringParameters(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, searchResult)
	return
}

// Recommend function
func (controller *ItemController) Recommend(c echo.Context) (err error) {
	searchResult, err := controller.Interactor.Recommend(controller.queryStringParameters(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, searchResult)
	return
}

// Classification function
func (controller *ItemController) Classification(c echo.Context) (err error) {
	searchResult, err := controller.Interactor.Classification(controller.queryStringParameters(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, searchResult)
	return
}

// Access function
func (controller *ItemController) Access(c echo.Context) (err error) {
	searchResult, err := controller.Interactor.AccessInfo(controller.queryStringParameters(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, searchResult)
	return
}
