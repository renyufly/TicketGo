package pagination

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Page struct {
	Limit  int
	Offset int
}

func Parse(c *gin.Context) (Page, error) {
	limit, err := positiveInt(c.DefaultQuery("limit", "20"))
	if err != nil || limit > 100 {
		return Page{}, fmt.Errorf("limit must be between 1 and 100")
	}
	offset, err := nonNegativeInt(c.DefaultQuery("offset", "0"))
	if err != nil || offset > 10000 {
		return Page{}, fmt.Errorf("offset must be between 0 and 10000")
	}
	return Page{Limit: limit, Offset: offset}, nil
}

func positiveInt(v string) (int, error) {
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("positive integer required")
	}
	return n, nil
}
func nonNegativeInt(v string) (int, error) {
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("non-negative integer required")
	}
	return n, nil
}
