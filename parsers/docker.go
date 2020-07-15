package parsers

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gdsoumya/better_ci/utils"
)

func DockerParser(path string, imageMap map[string]string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	pat := regexp.MustCompile(`#{.*}`)
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "#{PORT}") {
			tempP, _ := utils.GetFreePort()
			temp := strconv.Itoa(tempP)
			line = strings.Replace(line, "#{PORT}", temp, -1)
		} else if pat.MatchString(line) {
			key := pat.FindString(line)
			s := strings.Replace(key, "#{", "", 1)
			s = strings.Replace(s, "}", "", 1)
			if value, ok := imageMap[s]; ok {
				line = strings.Replace(line, key, value, -1)
			}
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return  err
	}
	err = utils.PrintLines(path, lines)
	if err != nil {
		return  err
	}
	return nil
}
