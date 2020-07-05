package old

import (
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("bash", "hello.sh")
	if output, er := cmd.Output(); er != nil {
		log.Print(er.Error())
	} else {
		log.Print(output)
	}
}
