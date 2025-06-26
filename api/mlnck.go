package api

import (
	"mlnck/pkg/helloasso"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetFormsHandler(c *gin.Context) {
	forms, err := helloasso.GetForms()

	// order forms by startDate
	for i := range forms {
		for j := i + 1; j < len(forms); j++ {
			if forms[i].StartDate > forms[j].StartDate {
				forms[i], forms[j] = forms[j], forms[i]
			}
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch forms"})
		return
	}
	c.JSON(http.StatusOK, forms)
}
