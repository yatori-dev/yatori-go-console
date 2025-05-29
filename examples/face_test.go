package examples

import (
	"fmt"
	"github.com/yatori-dev/yatori-go-core/utils"
	"testing"
)

func TestFace(t *testing.T) {
	base64, err := utils.GetFaceBase64()
	if err != nil {
		panic(err)
	}
	fmt.Println(base64)
}
