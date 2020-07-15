package ci

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gdsoumya/better_ci/parsers"
	"github.com/gdsoumya/better_ci/types"
)

func (c *Config) Deploy(cmt *types.EventDetails) {
	dir, err := c.ClonePR(cmt)
	if err != nil {
		c.CleanUp(dir, map[string]string{}, cmt)
		return
	}
	if dir == "" {
		return
	}
	ciConfig, err := parsers.ConfigParser(dir)
	log.Print("CI-CONFIG: ", ciConfig)
	if err != nil {
		c.CommentPR(cmt, "**ERROR IN CI**")
		log.Print("ERROR PARSING CONFIG : ", err.Error())
	}
	err = c.ciCMD(dir, ciConfig.CMD, cmt)
	if err != nil {
		c.CleanUp(dir, map[string]string{}, cmt)
		return
	}
	imageMap, er := c.buildImage(dir, ciConfig.BUILD, cmt)
	if er != nil {
		c.CleanUp(dir, imageMap, cmt)
		return
	}
	if ciConfig.DOCKER != "" {
		c.dockerDeploy(dir, ciConfig.DOCKER, imageMap, cmt)
	} else if ciConfig.K8S != "" {
		c.k8sDeploy(dir, ciConfig.K8S, imageMap, cmt)
	}
	c.CleanUp(dir, imageMap, cmt)
}
func (c *Config) dockerDeploy(dir string, docker string, imageMap map[string]string, cmt *types.EventDetails) error {
	log.Print("EXECUTING DOCKER STAGE FOR: ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	err := parsers.DockerParser(dir+"/"+docker, imageMap)
	if err != nil {
		c.CommentPR(cmt, "**ERROR IN CI**")
		log.Print("XX FAILED to PARSE DOCKER-COMPOSE FOR : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		log.Print("ERROR : ", err.Error())
		return err
	}
	log.Print("STARTING DOCKER FOR : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	cmd := exec.Command("docker-compose", "-f", docker, "up", "-d")
	cmd.Dir = dir
	// run command
	if output, err := cmd.Output(); err != nil {
		c.CommentPR(cmt, "**ERROR IN CI**")
		log.Print("XX FAILED to Start Docker for ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		log.Print("ERROR : ", err.Error())
		return err
	} else {
		log.Printf("\n%s", output)
		log.Print("DEPLOYED : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	}
	cmtData := "\n**LINK EXPIRES IN : " + strconv.Itoa(cmt.Time) + "mins**\n"
	cmd = exec.Command("docker-compose", "ps")
	cmd.Dir=dir
	output, err := cmd.Output()
	if err != nil {
		c.CommentPR(cmt, "**ERROR IN CI**")
		log.Print("XX ERROR Fetching Docker-Compose PS for ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		log.Print("ERROR : ", err, output)
		return err
	}
	ports := strings.Split(strings.Trim(string(output), "'"), "\n")
	ports=ports[2:]
	portData:="Name | Address\n--- | ---\n"
	for _,value:=range ports{
		if value==""{
			continue
		}
		port := strings.Split(strings.TrimSpace(value)," ")
		if strings.Contains(port[len(port)-1],"->") {
			svName := strings.Split(port[0], "pr"+cmt.PR+"_")[1]
			addr := strings.Split(port[len(port)-1], "->")[0]
			addr=c.host+":"+strings.Split(addr,":")[1]
			portData += svName + " | "+addr+"\n"
		}
	}
	cmtData += portData
	c.CommentPR(cmt, cmtData)
	time.Sleep(time.Duration(cmt.Time) * time.Minute)
	log.Print("STOPPING DOCKER FOR : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	cmd = exec.Command("docker-compose", "-f", docker, "down", "-v", "--rmi", "all") //ADD --rmi all
	cmd.Dir = dir
	// run command
	if output, err := cmd.Output(); err != nil {
		c.CommentPR(cmt, "**ERROR IN CI**")
		log.Print("XX FAILED to Stop Docker for ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		log.Print("ERROR : ", err.Error())
		return err
	} else {
		log.Printf("\n%s", output)
		log.Print("STOPPED : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	}
	c.CommentPR(cmt, "\n**LINK EXPIRED**")
	return nil
}

func (c *Config) k8sDeploy(dir string, k8s string, imageMap map[string]string, cmt *types.EventDetails) error {
	log.Print("EXECUTING K8S STAGE FOR: ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	err := parsers.K8sParser(dir+"/"+k8s, imageMap)
	if err != nil {
		c.CommentPR(cmt, "**ERROR IN CI**")
		log.Print("XX FAILED to PARSE K8S-MANIFEST FOR : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		log.Print("ERROR : ", err.Error())
		return err
	}
	log.Print("CREATING K8S NAMESPACE FOR : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	cmd := exec.Command("kubectl", "create", "ns", cmt.Username+"-"+cmt.Repo+"-pr"+cmt.PR)
	if output, err := cmd.Output(); err != nil {
		c.CommentPR(cmt, "**ERROR IN CI**")
		log.Print("XX FAILED to Create K8s Namespace For ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		log.Print("ERROR : ", err.Error())
		return err
	} else {
		log.Printf("\n%s", output)
		log.Print("K8s Namespace Created For : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	}
	log.Print("STARTING K8s DEPLOYMENT FOR : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	cmd = exec.Command("kubectl", "apply", "-f", k8s, "-n", cmt.Username+"-"+cmt.Repo+"-pr"+cmt.PR)
	cmd.Dir = dir
	if output, err := cmd.Output(); err != nil {
		c.CommentPR(cmt, "**ERROR IN CI**")
		log.Print("XX FAILED to START K8s DEPLOYMENT FOR ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		log.Print("ERROR : ", err.Error())
		return err
	} else {
		log.Printf("\n%s", output)
		log.Print("DEPLOYED IN K8s : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	}
	cmd = exec.Command("kubectl", "get", "svc", "-n", cmt.Username+"-"+cmt.Repo+"-pr"+cmt.PR, "-o", `go-template='{{range .items}}{{range.spec.ports}}{{if .nodePort}}{{.name}}{{" "}}{{.nodePort}}{{" "}}{{.protocol}}{{"\n"}}{{end}}{{end}}{{end}}'`)
	output, err := cmd.Output()
	if err != nil {
		//c.CommentPR(cmt, "**ERROR IN CI**")
		log.Print("XX ERROR Fetching NodePorts for ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		log.Print("ERROR : ", err.Error())
		return err
	}
	cmtData := "\n**LINK EXPIRES IN : " + strconv.Itoa(cmt.Time) + "mins**\n"
	ports := strings.Split(strings.Trim(string(output), "'"), "\n")
	portData := "SVC Name | Address | Protocol\n--- | --- | ---\n"
	for _, value := range ports {
		if value == "" {
			continue
		}
		port := strings.Split(value, " ")
		portData += port[0] + " | " + c.host + ":" + port[1] + " | " + port[2] + "\n"
	}
	cmtData += portData
	c.CommentPR(cmt, cmtData)
	time.Sleep(time.Duration(cmt.Time) * time.Minute)
	log.Print("STOPPING K8s DEPLOYMENT FOR : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	cmd = exec.Command("kubectl", "delete", "ns", cmt.Username+"-"+cmt.Repo+"-pr"+cmt.PR)
	// run command
	if output, err := cmd.Output(); err != nil {
		c.CommentPR(cmt, "**ERROR IN CI**")
		log.Print("XX FAILED to Stop K8s Deployment for ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		log.Print("ERROR : ", err.Error())
		return err
	} else {
		log.Printf("\n%s", output)
		log.Print("STOPPED K8s DEPLOYMENT for: ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	}
	c.CommentPR(cmt, "\n**LINK EXPIRED**")
	return nil
}

func (c *Config) ciCMD(dir string, cmds []string, cmt *types.EventDetails) error {
	log.Print("EXECUTING CMD STEP FOR : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	for _, cmd := range cmds {
		log.Print("EXECUTING CMD : ", cmd)
		s := strings.Split(cmd, " ")
		if len(s) > 0 {
			name := s[0]
			var args []string
			if len(s) > 1 {
				args = s[1:]
			}
			cmd := exec.Command(name, args...)
			cmd.Dir = dir
			if output, err := cmd.Output(); err != nil {
				log.Print("ERROR EXECUTING CMD :", cmd)
				log.Print(err.Error())
				c.CommentPR(cmt, "**ERROR IN CI**")
				return err
			} else {
				log.Print(string(output))
			}
		}
	}
	return nil
}

func (c *Config) buildImage(dir string, images []parsers.Image, cmt *types.EventDetails) (map[string]string, error) {
	log.Print("EXECUTING BUILD STEP FOR : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	imageMap := map[string]string{}
	tag := cmt.Username + "." + cmt.Repo + ".pr" + cmt.PR
	for _, image := range images {
		log.Print("EXECUTING BUILD : ", image.NAME)
		imgName := c.dockeru + "/" + image.NAME + ":" + tag
		cmd := exec.Command("docker", "build", "-t", imgName, "-f", image.FILE, image.CONTEXT)
		cmd.Dir = dir
		if output, err := cmd.Output(); err != nil {
			log.Print("ERROR BUILDING IMAGE :", image.NAME)
			log.Print(err.Error())
			c.CommentPR(cmt, "**ERROR IN CI**")
			return imageMap, err
		} else {
			log.Print(string(output))
		}
		if image.PUSH {
			cmd = exec.Command("bash", "scripts/dockerPush.sh", c.dockeru, c.dockerp, imgName)
			if output, err := cmd.Output(); err != nil {
				log.Print("ERROR BUILDING IMAGE :", image.NAME)
				log.Print(err.Error())
				c.CommentPR(cmt, "**ERROR IN CI**")
				return imageMap, err
			} else {
				log.Print(string(output))
			}
		}
		imageMap[image.NAME] = imgName
	}
	return imageMap, nil
}
