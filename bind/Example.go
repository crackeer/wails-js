package bind

import (
	"context"
	"fmt"
)

// App struct
type Example struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewExample() *Example {
	return &Example{}
}

// Greet returns a greeting for the given name
//
//	@receiver a
//	@param name
//	@return string
func (a *Example) Greet(name string) string {
	fmt.Println(fmt.Sprintf("Hello %s, It's show time!", name))
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
