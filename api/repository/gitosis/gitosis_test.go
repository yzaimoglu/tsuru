package gitosis

import (
	ini "github.com/kless/goconfig/config"
	"github.com/timeredbull/tsuru/config"
	. "launchpad.net/gocheck"
	"os"
	"os/exec"
	"path"
)

func (s *S) TestAddGroup(c *C) {
	err := AddGroup("someGroup")
	c.Assert(err, IsNil)

	conf, err := ini.ReadDefault(path.Join(s.gitosisRepo, "gitosis.conf"))
	c.Assert(err, IsNil)
	//ensures that project have been added to gitosis.conf
	c.Assert(conf.HasSection("group someGroup"), Equals, true)

	//ensures that file is not overriden when a new project is added
	err = AddGroup("someOtherGroup")
	c.Assert(err, IsNil)
	// it should have both sections
	conf, err = ini.ReadDefault(path.Join(s.gitRoot, "gitosis-admin/gitosis.conf"))
	c.Assert(err, IsNil)
	c.Assert(conf.HasSection("group someGroup"), Equals, true)
	c.Assert(conf.HasSection("group someOtherGroup"), Equals, true)
}

func (s *S) TestAddGroupShouldReturnErrorWhenSectionAlreadyExists(c *C) {
	err := AddGroup("aGroup")
	c.Assert(err, IsNil)

	err = AddGroup("aGroup")
	c.Assert(err, NotNil)
}

func (s *S) TestAddGroupShouldCommitAndPushChangesToGitosisBare(c *C) {
	err := AddGroup("gandalf")
	c.Assert(err, IsNil)
	pwd := os.Getenv("PWD")
	os.Chdir(s.gitosisBare)
	bareOutput, err := exec.Command("git", "log", "-1", "--pretty=format:%H").CombinedOutput()
	c.Assert(err, IsNil)

	os.Chdir(s.gitosisRepo)
	repoOutput, err := exec.Command("git", "log", "-1", "--pretty=format:%H").CombinedOutput()
	c.Assert(err, IsNil)

	os.Chdir(pwd)

	c.Assert(string(repoOutput), Equals, string(bareOutput))
}

func (s *S) TestRemoveGroup(c *C) {
	err := AddGroup("someGroup")
	c.Assert(err, IsNil)

	conf, err := ini.ReadDefault(path.Join(s.gitosisRepo, "gitosis.conf"))
	c.Assert(err, IsNil)
	c.Assert(conf.HasSection("group someGroup"), Equals, true)

	err = RemoveGroup("someGroup")
	conf, err = ini.ReadDefault(path.Join(s.gitosisRepo, "gitosis.conf"))
	c.Assert(err, IsNil)
	c.Assert(conf.HasSection("group someGroup"), Equals, false)

	pwd := os.Getenv("PWD")
	os.Chdir(s.gitosisBare)
	bareOutput, err := exec.Command("git", "log", "-1", "--pretty=format:%s").CombinedOutput()
	c.Assert(err, IsNil)
	os.Chdir(pwd)

	expected := "Removing group someGroup from gitosis.conf"

	c.Assert(string(bareOutput), Equals, expected)
}

func (s *S) TestRemoveGroupCommitAndPushesChanges(c *C) {
	err := AddGroup("testGroup")
	c.Assert(err, IsNil)

	conf, err := ini.ReadDefault(path.Join(s.gitosisRepo, "gitosis.conf"))
	c.Assert(err, IsNil)
	c.Assert(conf.HasSection("group testGroup"), Equals, true)

	err = RemoveGroup("testGroup")
	conf, err = ini.ReadDefault(path.Join(s.gitosisRepo, "gitosis.conf"))
	c.Assert(err, IsNil)
	c.Assert(conf.HasSection("group testGroup"), Equals, false)
}

func (s *S) TestAddMemberToGroup(c *C) {
	err := AddGroup("take-over-the-world") // test also with a inexistent project
	c.Assert(err, IsNil)
	err = AddMember("take-over-the-world", "brain")
	c.Assert(err, IsNil)
	conf, err := ini.ReadDefault(path.Join(s.gitosisRepo, "gitosis.conf"))
	c.Assert(err, IsNil)
	c.Assert(conf.HasSection("group take-over-the-world"), Equals, true)
	c.Assert(conf.HasOption("group take-over-the-world", "members"), Equals, true)
	members, err := conf.String("group take-over-the-world", "members")
	c.Assert(err, IsNil)
	c.Assert(members, Equals, "brain")
}

func (s *S) TestAddMemberToGroupCommitsAndPush(c *C) {
	err := AddGroup("someTeam")
	c.Assert(err, IsNil)
	err = AddMember("someTeam", "brain")
	pwd := os.Getenv("PWD")
	os.Chdir(s.gitosisBare)
	bareOutput, err := exec.Command("git", "log", "-1", "--pretty=format:%s").CombinedOutput()
	c.Assert(err, IsNil)

	os.Chdir(pwd)

	commitMsg := "Adding member brain for group someTeam"

	c.Assert(string(bareOutput), Equals, commitMsg)
}

func (s *S) TestAddTwoMembersToGroup(c *C) {
	err := AddGroup("pink-floyd")
	c.Assert(err, IsNil)
	err = AddMember("pink-floyd", "one-of-these-days")
	c.Assert(err, IsNil)
	err = AddMember("pink-floyd", "comfortably-numb")
	c.Assert(err, IsNil)
	conf, err := ini.ReadDefault(path.Join(s.gitosisRepo, "gitosis.conf"))
	members, err := conf.String("group pink-floyd", "members")
	c.Assert(err, IsNil)
	c.Assert(members, Equals, "one-of-these-days comfortably-numb")
}

func (s *S) TestAddMemberToGroupReturnsErrorIfTheMemberIsAlreadyInTheGroup(c *C) {
	err := AddGroup("pink-floyd")
	c.Assert(err, IsNil)
	err = AddMember("pink-floyd", "time")
	c.Assert(err, IsNil)
	err = AddMember("pink-floyd", "time")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "^This user is already member of this group$")
}

func (s *S) TestAddAndCommit(c *C) {
	confPath := path.Join(s.gitosisRepo, "gitosis.conf")
	conf, err := ini.ReadDefault(confPath)
	c.Assert(err, IsNil)
	conf.AddSection("foo bar")
	PushToGitosis("Some commit message")

	pwd := os.Getenv("PWD")
	os.Chdir(s.gitosisBare)
	bareOutput, err := exec.Command("git", "log", "-1", "--pretty=format:%s").CombinedOutput()
	c.Assert(err, IsNil)
	os.Chdir(pwd)

	c.Assert(string(bareOutput), Equals, "Some commit message")
}

func (s *S) TestConfPathReturnsGitosisConfPath(c *C) {
	repoPath, err := config.GetString("git:gitosis-repo")
	expected := path.Join(repoPath, "gitosis.conf")
	obtained, err := ConfPath()
	c.Assert(err, IsNil)
	c.Assert(obtained, Equals, expected)
}
