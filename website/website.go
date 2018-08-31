// website
package website

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Starting website server...")
	http.Handle("/", http.FileServer(http.Dir("")))
	http.ListenAndServeTLS(":80", "", "", nil)
}
