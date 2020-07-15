package ci

import (
	"errors"
	"log"
	"os"
	"os/exec"

	"github.com/gdsoumya/better_ci/types"
	"github.com/gdsoumya/better_ci/utils"
	"github.com/google/go-github/v32/github"
)

func (c *Config) ClonePR(cmt *types.EventDetails) (string, error) {
	dir := cmt.Username + "_" + cmt.Repo + "_pr" + cmt.PR
	if utils.DirPresent(dir) {
		c.CommentPR(cmt, "** WAIT FOR PREVIOUS BUILD TO EXPIRE **")
		log.Print("BUILD IN PROGRESS : " + cmt.Username + "/" + cmt.Repo + ":PR#" + cmt.PR)
		return "", nil
	}
	errDir := os.MkdirAll(dir, 0755)
	if errDir != nil {
		c.CommentPR(cmt, "** ERROR IN CI **")
		log.Print("DIR CREATION ERROR : " + cmt.Username + "/" + cmt.Repo + ":PR#" + cmt.PR)
		return dir, errors.New("ERROR")
	}
	cmd := exec.Command("git", "clone", "https://github.com/"+cmt.Username+"/"+cmt.Repo+".git", ".")
	cmd.Dir = dir
	// run command
	if _, err := cmd.Output(); err != nil {
		c.CommentPR(cmt, "** ERROR IN CI **")
		log.Print("XX FAILED to Cloned Repo for ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		return dir, err
	}
	log.Print("Cloned Repo for ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	cmd = exec.Command("git", "fetch", "origin", "refs/pull/"+cmt.PR+"/head:pr_"+cmt.PR)
	cmd.Dir = dir
	// run command
	if _, err := cmd.Output(); err != nil {
		c.CommentPR(cmt, "** ERROR IN CI **")
		log.Print("XX FAILED to Setup PR for ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		return dir, err
	}
	cmd = exec.Command("git", "checkout", "pr_"+cmt.PR)
	cmd.Dir = dir
	// run command
	if _, err := cmd.Output(); err != nil {
		c.CommentPR(cmt, "** ERROR IN CI **")
		log.Print("XX FAILED to Setup PR for ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
		return dir, err
	}
	log.Print("Setup PR for", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	return dir, nil

}

func (c *Config) CommentPR(cmt *types.EventDetails, cmtData string) {
	var data github.IssueComment
	cmt.Body += "\n" + cmtData
	data.Body = &cmt.Body
	_, _, err := c.client.Issues.EditComment(c.ctx, cmt.Username, cmt.Repo, cmt.ID, &data)
	if err != nil {
		log.Print("ERROR CommentPR: ", err.Error())
	}
}

func (c *Config) CleanUp(dir string, imageMap map[string]string, cmt *types.EventDetails) error {
	if !utils.DirPresent(dir) {
		return nil
	}
	err := os.RemoveAll(dir)
	if err != nil {
		c.CommentPR(cmt, "** ERROR IN CI **")
		log.Print("XX FAILED to Clean Up for : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR, "\n", err.Error())
		return err
	}
	for _, value := range imageMap {
		log.Print("DELETING IMAGE : ", value)
		cmd := exec.Command("docker", "rmi", value)
		// run command
		if _, err := cmd.Output(); err != nil {
			//c.CommentPR(cmt, "** ERROR IN CI **")
			log.Print("XX FAILED to Remove Images for :", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR, "\n", err.Error())
			return err
		}
	}
	log.Print("CLEANUP Complete for : ", cmt.Username+"/"+cmt.Repo+":PR#"+cmt.PR)
	return nil
}
