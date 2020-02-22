package controllers

import (
	"net/http"
	"strconv"

	"github.com/akaishi-sandbox/sam-go/infrastructure"
	"github.com/akaishi-sandbox/sam-go/interfaces/database"
	"github.com/akaishi-sandbox/sam-go/usecase"
	"github.com/labstack/echo"
)

type ItemController struct {
	Interactor usecase.ItemInteractor
}

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

func (controller *ItemController) Search(c echo.Context) (err error) {
	// id, _ := strconv.Atoi(c.Param("id"))
	searchResult, err := controller.Interactor.Search(controller.queryStringParameters(c))
	// QueryStringParameters
	// user, err := controller.Interactor.UserById(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, searchResult)
	return
}

func (controller *ItemController) Recommend(c echo.Context) (err error) {
	id, _ := strconv.Atoi(c.Param("id"))
	// user, err := controller.Interactor.UserById(id)
	// if err != nil {
	// 	c.JSON(500, NewError(err))
	// 	return
	// }
	c.JSON(http.StatusOK, id)
	return
}

func (controller *ItemController) Classification(c echo.Context) (err error) {
	id, _ := strconv.Atoi(c.Param("id"))
	// user, err := controller.Interactor.UserById(id)
	// if err != nil {
	// 	c.JSON(500, NewError(err))
	// 	return
	// }
	c.JSON(http.StatusOK, id)
	return
}

func (controller *ItemController) Access(c echo.Context) (err error) {
	id, _ := strconv.Atoi(c.Param("id"))
	// user, err := controller.Interactor.UserById(id)
	// if err != nil {
	// 	c.JSON(500, NewError(err))
	// 	return
	// }
	c.JSON(http.StatusOK, id)
	return
}